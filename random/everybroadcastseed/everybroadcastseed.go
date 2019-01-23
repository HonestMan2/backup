// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package everybroadcastseed

import (
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/params/manparams"
)

var (
	ModuleEveryBroadcastSeed   = "广播区块种子"
	mapEveryBroadcastSeedPlugs = make(map[string]preBroadcastSeedPlug)
)

func init() {
	baseinterface.RegRandom(manparams.EveryBroadcastSeed, newSubService)
}

type preBroadcastSeedPlug interface {
	CalcSeed(data common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error)
	Prepare(uint64) error
}

func newSubService(plug string, support baseinterface.RandomChainSupport) (baseinterface.RandomSubService, error) {
	everyBroadcastSeed := &preBroadcastSeed{
		plug:    plug,
		support: support,
	}
	return everyBroadcastSeed, nil
}

type preBroadcastSeed struct {
	plug    string
	support baseinterface.RandomChainSupport
}

func (self *preBroadcastSeed) SetValue(plug string, support baseinterface.RandomChainSupport) error {
	self.plug = plug
	self.support = support
	return nil
}

func RegisterEveryBlockSeedPlugs(name string, plug preBroadcastSeedPlug) {
	mapEveryBroadcastSeedPlugs[name] = plug
}

func (self *preBroadcastSeed) Prepare(height uint64) error {
	err := mapEveryBroadcastSeedPlugs[self.plug].Prepare(height)
	return err
}

func (self *preBroadcastSeed) CalcData(calcData common.Hash) (*big.Int, error) {
	return mapEveryBroadcastSeedPlugs[self.plug].CalcSeed(calcData, self.support)
}
