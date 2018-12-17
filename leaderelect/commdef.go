// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"errors"
	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/msgsend"
)

var (
	ErrMsgAccountIsNull     = errors.New("不合法的账户：空账户")
	ErrValidatorsIsNil      = errors.New("验证者列表为空")
	ErrSepcialsIsNil        = errors.New("特殊账户为空")
	ErrValidatorNotFound    = errors.New("验证者未找到")
	ErrMsgExistInCache      = errors.New("缓存中已存在消息")
	ErrNoMsgInCache         = errors.New("缓存中没有目标消息")
	ErrParamsIsNil          = errors.New("参数为nil")
	ErrSelfReqIsNil         = errors.New("self请求不在缓存中")
	ErrPOSResultIsNil       = errors.New("POS结果为nil/header为nil")
	ErrLeaderResultIsNil    = errors.New("leader共识结果为nil")
	ErrCDCOrSignHelperisNil = errors.New("cdc or signHelper is nil")
)

type Matrix interface {
	BlockChain() *core.BlockChain
	SignHelper() *signhelper.SignHelper
	DPOSEngine() consensus.DPOSEngine
	Engine() consensus.Engine
	HD() *msgsend.HD
	FetcherNotify(hash common.Hash, number uint64)
}

const defaultBeginTime = int64(0)

const mangerCacheMax = 2

type stateDef uint8

const (
	stIdle stateDef = iota
	stPos
	stReelect
	stMining
)

func (s stateDef) String() string {
	switch s {
	case stIdle:
		return "未运行阶段"
	case stPos:
		return "POS阶段"
	case stReelect:
		return "重选阶段"
	case stMining:
		return "挖矿结果等待阶段"
	default:
		return "未知状态"
	}
}

type leaderData struct {
	leader     common.Address
	nextLeader common.Address
}

func (self *leaderData) copyData() *leaderData {
	newData := &leaderData{
		leader:     common.Address{},
		nextLeader: common.Address{},
	}

	newData.leader.Set(self.leader)
	newData.nextLeader.Set(self.nextLeader)
	return newData
}

type startControllerMsg struct {
	parentIsSupper bool
	parentHeader   *types.Header
	parentStateDB  *state.StateDB
}

func isFirstConsensusTurn(turnInfo *mc.ConsensusTurnInfo) bool {
	if turnInfo == nil {
		return false
	}
	return turnInfo.PreConsensusTurn == 0 && turnInfo.UsedReelectTurn == 0
}

func calcNextConsensusTurn(curConsensusTurn mc.ConsensusTurnInfo, curReelectTurn uint32) mc.ConsensusTurnInfo {
	return mc.ConsensusTurnInfo{
		PreConsensusTurn: curConsensusTurn.TotalTurns(),
		UsedReelectTurn:  curReelectTurn,
	}
}
