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

	err = self.HandleNative(data.BlockNum) //处理初选列表更新
	if err != nil {
		log.ERROR(Module, "處理初選列表更新失敗 err", err)
		return err
	}

	log.INFO(Module, "roleUpdateProcess end height", data.BlockNum)
	return nil

}
func (self *ReElection) HandleNative(height uint64) error {
	if true == NeedReadTopoFromDB(height) { //300 600 900 重取缓存
		log.INFO(Module, "需要从db中读取native 高度", height)
		return self.GetNativeFromDB(height)
	}
	self.checkUpdateStatus(height - 1)
	allNative, err := self.readNativeData(height - 1) //
	if err != nil {
		log.Error(Module, "readNativeData failed height", height-1)
	}

	log.INFO(Module, "self,allNative", allNative)

	err = self.UpdateNative(height, allNative)
	log.INFO(Module, "更新初选列表结束 高度 ", height, "错误信息", err, "self,allNative", allNative)
	return err
}
func (self *ReElection) HandleTopGen(hash common.Hash) error {
	var err error

	if self.IsMinerTopGenTiming(hash) { //矿工生成时间 240
		log.INFO(Module, "是礦工拓撲生成時間點 height", hash.String())
		err = self.ToGenMinerTop(hash)
		if err != nil {
			log.ERROR(Module, "礦工拓撲生成時間點錯誤 err", err)
		}
	}

	if self.IsValidatorTopGenTiming(hash) { //验证者生成时间 260
		log.INFO(Module, "是驗證者拓撲生成時間點 height", hash)
		err = self.ToGenValidatorTop(hash)
		if err != nil {
			log.ERROR(Module, "驗證者拓撲生成時間點錯誤 err", err)
		}
	}

	return err
}
func (self *ReElection) UpdateNative(height uint64, allNative support.AllNative) error {

	allNative, err := self.ToNativeValidatorStateUpdate(height, allNative)
	if err != nil {
		log.INFO(Module, "ToNativeMinerStateUpdate validator err", err)
		return nil
	}

	err = self.writeNativeData(height, allNative)

	log.ERROR(Module, "更新初选列表状态后-写入数据库状态 err", err, "高度", height)

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
