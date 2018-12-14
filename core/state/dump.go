// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package state

import (
	"encoding/json"
	"fmt"

	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/rlp"
	"github.com/matrix/go-matrix/trie"
)

type DumpAccount struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Root     string            `json:"root"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code"`
	Storage  map[string]string `json:"storage"`
}

type Dump struct {
	Root     string                 `json:"root"`
	Accounts map[string]DumpAccount `json:"accounts"`
}

func (self *StateDB) RawDump() Dump {
	dump := Dump{
		Root:     fmt.Sprintf("%x", self.trie.Hash()),
		Accounts: make(map[string]DumpAccount),
	}

	it := trie.NewIterator(self.trie.NodeIterator(nil))
	for it.Next() {
		addr := self.trie.GetKey(it.Key)
		var data []Account
		if err := rlp.DecodeBytes(it.Value, &data); err != nil {
			panic(err)
		}

		tBalance := new(big.Int)
		for _, tAccount := range data {
			if tAccount.Balance[0].AccountType == common.MainAccount {
				tBalance = tAccount.Balance[0].Balance
				break
			}
		}
		obj := newObject(nil, common.BytesToAddress(addr), data[0])
		account := DumpAccount{
			//Balance:  data.Balance.String(),
			Balance:  tBalance.String(),
			Nonce:    data[0].Nonce,
			Root:     common.Bytes2Hex(data[0].Root[:]),
			CodeHash: common.Bytes2Hex(data[0].CodeHash),
			Code:     common.Bytes2Hex(obj.Code(self.db)),
			Storage:  make(map[string]string),
		}
		storageIt := trie.NewIterator(obj.getTrie(self.db).NodeIterator(nil))
		for storageIt.Next() {
			account.Storage[common.Bytes2Hex(self.trie.GetKey(storageIt.Key))] = common.Bytes2Hex(storageIt.Value)
		}
		dump.Accounts[common.Bytes2Hex(addr)] = account
	}
	return dump
}

func (self *StateDB) Dump() []byte {
	json, err := json.MarshalIndent(self.RawDump(), "", "    ")
	if err != nil {
		fmt.Println("dump err", err)
	}

	return json
}
