// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package state

import (
	"sync"

	"github.com/matrix/go-matrix/common"
)

type account struct {
	stateObject *stateObject
	nstart      uint64
	nonces      []bool
}

type ManagedState struct {
	*StateDBManage

	mu sync.RWMutex

	accounts map[common.Address]*account
}

// ManagedState returns a new managed state with the statedb as it's backing layer
func ManageState(statedb *StateDBManage) *ManagedState {
	return &ManagedState{
		StateDBManage:  statedb.Copy(),
		accounts: make(map[common.Address]*account),
	}
}

// SetState sets the backing layer of the managed state
func (ms *ManagedState) SetState(statedb *StateDBManage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.StateDBManage = statedb
}

// RemoveNonce removed the nonce from the managed state and all future pending nonces
func (ms *ManagedState) RemoveNonce(cointype string ,addr common.Address, n uint64) {
	if ms.hasAccount(addr) {
		ms.mu.Lock()
		defer ms.mu.Unlock()

		account := ms.getAccount(cointype,addr)
		if n-account.nstart <= uint64(len(account.nonces)) {
			reslice := make([]bool, n-account.nstart)
			copy(reslice, account.nonces[:n-account.nstart])
			account.nonces = reslice
		}
	}
}

// NewNonce returns the new canonical nonce for the managed account
func (ms *ManagedState) NewNonce(cointype string ,addr common.Address) uint64 {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	account := ms.getAccount(cointype,addr)
	for i, nonce := range account.nonces {
		if !nonce {
			return account.nstart + uint64(i)
		}
	}
	account.nonces = append(account.nonces, true)

	return uint64(len(account.nonces)-1) + account.nstart
}

// GetNonce returns the canonical nonce for the managed or unmanaged account.
//
// Because GetNonce mutates the DB, we must take a write lock.
func (ms *ManagedState) GetNonce(cointype string ,addr common.Address) uint64 {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.hasAccount(addr) {
		account := ms.getAccount(cointype,addr)
		return uint64(len(account.nonces)) + account.nstart
	} else {
		return ms.StateDBManage.GetNonce(cointype,addr)
	}
}

// SetNonce sets the new canonical nonce for the managed state
func (ms *ManagedState) SetNonce(cointype string,addr common.Address, nonce uint64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	so := ms.GetOrNewStateObject(cointype,addr)
	so.SetNonce(nonce)

	ms.accounts[addr] = newAccount(so)
}

// HasAccount returns whether the given address is managed or not
func (ms *ManagedState) HasAccount(addr common.Address) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.hasAccount(addr)
}

func (ms *ManagedState) hasAccount(addr common.Address) bool {
	_, ok := ms.accounts[addr]
	return ok
}

// populate the managed state
func (ms *ManagedState) getAccount(cointype string,addr common.Address) *account {
	if account, ok := ms.accounts[addr]; !ok {
		so := ms.GetOrNewStateObject(cointype,addr)
		ms.accounts[addr] = newAccount(so)
	} else {
		// Always make sure the state account nonce isn't actually higher
		// than the tracked one.
		so := ms.StateDBManage.getStateObject(cointype,addr)
		if so != nil && uint64(len(account.nonces))+account.nstart < so.Nonce() {
			ms.accounts[addr] = newAccount(so)
		}

	}

	return ms.accounts[addr]
}

func newAccount(so *stateObject) *account {
	return &account{so, so.Nonce(), nil}
}
