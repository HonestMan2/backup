// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

//身份变更消息带来
/*
func (self *ReElection) roleUpdateProcess(data *mc.RoleUpdatedMsg) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.currentID = data.Role

	if common.RoleValidator != self.currentID { //不是验证者，不处理
		log.ERROR(Module, "當前不是驗證者，不處理", self.currentID)
		return nil
	}

	err := self.HandleTopGen(data.BlockHash) //处理拓扑生成
	if err != nil {
		log.ERROR(Module, "處理拓撲生成失敗 err", err)
		return err
	}

	err = self.HandleNative(data.BlockHash) //处理初选列表更新
	if err != nil {
		log.ERROR(Module, "處理初選列表更新失敗 err", err)
		return err
	}

	log.INFO(Module, "roleUpdateProcess end height", data.BlockNum)
	return nil

}
*/
/*
func (self *ReElection) HandleNative(hash common.Hash) error {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.Error(Module, "HandleNative阶段 err", err)
		return err
	}

	if true == NeedReadTopoFromDB(height) { //300 600 900 重取缓存
		log.INFO(Module, "需要从db中读取native 高度", height)
		return self.GetNativeFromDB(hash)
	}

	lastHash, err := self.GetHeaderHashByNumber(hash, height-1)
	if err != nil {
		log.Error(Module, "HandleNative err", err)
		return err
	}
	self.checkUpdateStatus(lastHash)
	allNative, err := self.readNativeData(lastHash) //
	if err != nil {
		log.Error(Module, "readNativeData failed height", height-1)
	}

	log.INFO(Module, "self,allNative", allNative)

	err = self.UpdateNative(hash, allNative)
	log.INFO(Module, "更新初选列表结束 高度 ", height, "错误信息", err, "self,allNative", allNative)
	return err
}
*/
type TopGenStatus struct {
	//V_State bool
	MastV []mc.ElectNodeInfo
	BackV []mc.ElectNodeInfo
	CandV []mc.ElectNodeInfo

	//M_State bool
	MastM []mc.ElectNodeInfo
	BackM []mc.ElectNodeInfo
	CandM []mc.ElectNodeInfo
}

func (self *ReElection) HandleTopGen(hash common.Hash) (TopGenStatus, error) {
	//var err error
	topGenStatus := TopGenStatus{}

	if self.IsMinerTopGenTiming(hash) { //矿工生成时间 240
		log.INFO(Module, "是矿工生成时间点 hash", hash.String())
		MastM, BackM, CandM, err := self.ToGenMinerTop(hash)
		if err != nil {
			log.ERROR(Module, "矿工拓扑生成错误 err", err)
			return topGenStatus, err
		}
		topGenStatus.MastM = append(topGenStatus.MastM, MastM...)
		topGenStatus.BackM = append(topGenStatus.BackM, BackM...)
		topGenStatus.CandM = append(topGenStatus.CandM, CandM...)
		//	topGenStatus.M_State = true

	}

	if self.IsValidatorTopGenTiming(hash) { //验证者生成时间 260
		log.INFO(Module, "是验证者生成时间点 height", hash)
		MastV, BackV, CandV, err := self.ToGenValidatorTop(hash)
		if err != nil {
			log.ERROR(Module, "验证者拓扑生成错误 err", err)
			return topGenStatus, err
		}
		topGenStatus.MastV = append(topGenStatus.MastV, MastV...)
		topGenStatus.BackV = append(topGenStatus.BackV, BackV...)
		topGenStatus.CandV = append(topGenStatus.CandV, CandV...)
		//topGenStatus.V_State = true
	}
	return topGenStatus, nil

}
func (self *ReElection) UpdateNative(hash common.Hash, allNative support.AllNative) error {

	allNative, err := self.ToNativeValidatorStateUpdate(hash, allNative)
	if err != nil {
		log.INFO(Module, "ToNativeMinerStateUpdate validator err", err)
		return nil
	}

	err = self.writeNativeData(hash, allNative)

	log.ERROR(Module, "更新初选列表状态后-写入数据库状态 err", err, "高度对应的hash", hash.String())

	return err

}

//是不是矿工拓扑生成时间段
func (self *ReElection) IsMinerTopGenTiming(hash common.Hash) bool {

	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return false
	}
	now := height % common.GetReElectionInterval()
	if now == MinerTopGenTiming {
		return true
	}
	return false
}

//是不是验证者拓扑生成时间段
func (self *ReElection) IsValidatorTopGenTiming(hash common.Hash) bool {

	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return false
	}

	now := height % common.GetReElectionInterval()
	if now == ValidatorTopGenTiming {
		return true
	}
	return false
}
func NeedReadTopoFromDB(height uint64) bool {
	if (height)%common.GetReElectionInterval() == 0 || height == 0 {
		return true
	}
	return false
}
