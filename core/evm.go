// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"math/big"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
)

// ChainContext supports retrieving headers and consensus parameters from the
// current blockchain to be used during transaction processing.
type ChainContext interface {
	// Engine retrieves the chain's consensus engine.
	Engine(version []byte) consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(common.Hash, uint64) *types.Header
}

//Y =========================begin============================
// NewEVMContext creates a new context for use in the EVM.
//func NewEVMContext(msg Message, header *types.Header, chain ChainContext, author *common.Address) vm.Context {
//	// If we don't have an explicit author (i.e. not mining), extract from the header
//	var beneficiary common.Address
//	if author == nil {
//		beneficiary, _ = chain.Engine().Author(header) // Ignore error, we're past header validation
//	} else {
//		beneficiary = *author
//	}
//	return vm.Context{
//		CanTransfer: CanTransfer,
//		Transfer:    Transfer,
//		GetHash:     GetHashFn(header, chain),
//		Origin:      msg.From(),
//		Coinbase:    beneficiary,
//		BlockNumber: new(big.Int).Set(header.Number),
//		Time:        new(big.Int).Set(header.Time),
//		Difficulty:  new(big.Int).Set(header.Difficulty),
//		GasLimit:    header.GasLimit,
//		GasPrice:    new(big.Int).Set(msg.GasPrice()),
//	}
//}

func NewEVMContext(sender common.Address, gasprice *big.Int, header *types.Header, chain ChainContext, author *common.Address) vm.Context {
	// If we don't have an explicit author (i.e. not mining), extract from the header
	var beneficiary common.Address
	if author == nil {
		beneficiary, _ = chain.Engine(header.Version).Author(header) // Ignore error, we're past header validation
	} else {
		beneficiary = *author
	}
	return vm.Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(header, chain),
		Origin:      sender,
		Coinbase:    beneficiary,
		BlockNumber: new(big.Int).Set(header.Number),
		Time:        new(big.Int).Set(header.Time),
		Difficulty:  new(big.Int).Set(header.Difficulty),
		GasLimit:    header.GasLimit,
		GasPrice:    new(big.Int).Set(gasprice),
	}
}

//Y ====================================end================================
// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *types.Header, chain ChainContext) func(n uint64) common.Hash {
	var cache map[uint64]common.Hash

	return func(n uint64) common.Hash {
		// If there's no hash cache yet, make one
		if cache == nil {
			cache = map[uint64]common.Hash{
				ref.Number.Uint64() - 1: ref.ParentHash,
			}
		}
		// Try to fulfill the request from the cache
		if hash, ok := cache[n]; ok {
			return hash
		}
		// Not cached, iterate the blocks and cache the hashes
		for header := chain.GetHeader(ref.ParentHash, ref.Number.Uint64()-1); header != nil; header = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1) {
			cache[header.Number.Uint64()-1] = header.ParentHash
			if n == header.Number.Uint64()-1 {
				return header.ParentHash
			}
		}
		return common.Hash{}
	}
}

// CanTransfer checks wether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	for _, tAccount := range db.GetBalance(addr) {
		if tAccount.AccountType == common.MainAccount {
			return tAccount.Balance.Cmp(amount) >= 0
		}
	}
	return false
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(common.MainAccount, sender, amount)
	db.AddBalance(common.MainAccount, recipient, amount)
}
