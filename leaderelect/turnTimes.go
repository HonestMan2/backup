// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"time"
)

type turnTimes struct {
	beginTimes            map[uint32]int64
	parentMiningTime      int64 // 预留父区块挖矿时间
	posOutTime            int64 // 区块POS共识超时时间
	reelectOutTime        int64 // 重选超时时间
	reelectHandleInterval int64 // 重选处理间隔时间
}

func newTurnTimes() *turnTimes {
	tt := &turnTimes{
		beginTimes:            make(map[uint32]int64),
		parentMiningTime:      0,
		posOutTime:            0,
		reelectOutTime:        0,
		reelectHandleInterval: 0,
	}

	return tt
}

func (tt *turnTimes) SetTimeConfig(config *mc.LeaderConfig) error {
	if config == nil {
		return ErrParamsIsNil
	}

	tt.parentMiningTime = config.ParentMiningTime
	tt.posOutTime = config.PosOutTime
	tt.reelectOutTime = config.ReelectOutTime
	tt.reelectHandleInterval = config.ReelectHandleInterval
	return nil
}

func (tt *turnTimes) SetBeginTime(consensusTurn uint32, time int64) bool {
	if oldTime, exist := tt.beginTimes[consensusTurn]; exist {
		if time <= oldTime {
			return false
		}
	}
	tt.beginTimes[consensusTurn] = time
	return true
}

func (tt *turnTimes) GetBeginTime(consensusTurn uint32) int64 {
	if beginTime, exist := tt.beginTimes[consensusTurn]; exist {
		return beginTime
	} else {
		return 0
	}
}

func (tt *turnTimes) GetPosEndTime(consensusTurn uint32) int64 {
	_, endTime := tt.CalTurnTime(consensusTurn, 0)
	return endTime

	posTime := tt.posOutTime
	if consensusTurn == 0 {
		posTime += tt.parentMiningTime
	}

	return tt.GetBeginTime(consensusTurn) + posTime
}

func (tt *turnTimes) CalState(consensusTurn uint32, time int64) (st stateDef, remainTime int64, reelectTurn uint32) {
	if tt.reelectOutTime == 0 {
		log.Error("critical", "turnTimes", "reelectOutTime == 0")
		return stReelect, 0, 0
	}
	posTime := tt.posOutTime
	if consensusTurn == 0 {
		posTime += tt.parentMiningTime
	}

	passTime := time - tt.GetBeginTime(consensusTurn)
	if passTime < posTime {
		return stPos, posTime - passTime, 0
	}

	st = stReelect
	reelectTurn = uint32((passTime-posTime)/tt.reelectOutTime) + 1
	_, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	remainTime = endTime - time
	return
}

func (tt *turnTimes) CalRemainTime(consensusTurn uint32, reelectTurn uint32, time int64) int64 {
	_, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	return endTime - time
}

func (tt *turnTimes) CalTurnTime(consensusTurn uint32, reelectTurn uint32) (beginTime int64, endTime int64) {
	posTime := tt.posOutTime
	if consensusTurn == 0 {
		posTime += tt.parentMiningTime
	}

	if reelectTurn == 0 {
		beginTime = tt.GetBeginTime(consensusTurn)
		endTime = beginTime + posTime
	} else {
		beginTime = tt.GetBeginTime(consensusTurn) + posTime + int64(reelectTurn-1)*tt.reelectOutTime
		endTime = beginTime + tt.reelectOutTime
	}
	return
}

func (tt *turnTimes) CheckTimeLegal(consensusTurn uint32, reelectTurn uint32, checkTime int64) error {
	beginTime, endTime := tt.CalTurnTime(consensusTurn, reelectTurn)
	if checkTime <= beginTime || checkTime >= endTime {
		return errors.Errorf("时间(%s)非法,轮次起止时间(%s - %s)",
			time.Unix(checkTime, 0).String(), time.Unix(beginTime, 0).String(), time.Unix(endTime, 0).String())
	}
	return nil
}
