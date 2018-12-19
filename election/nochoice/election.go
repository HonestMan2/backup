// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package nochoice

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

const (
	DefauleStock = 1
)

type nochoice struct {
}

func init() {
	baseinterface.RegElectPlug("nochoice", RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &nochoice{}
}

func (self *nochoice) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.INFO("直接选举方案", "矿工拓扑生成", len(mmrerm.MinerList))
	nodeElect := support.NewElelection(nil, mmrerm.MinerList, mmrerm.ElectConfig, mmrerm.RandSeed, mmrerm.SeqNum)
	nodeElect.Disorder()
	nodeElect.Sort()
	nodeElect.ProcessBlackNode()
	nodeElect.ProcessWhiteNode()

	lastNode := nodeElect.GetLastNode()

	MinerTopGenAns := mc.MasterMinerReElectionRsp{
		SeqNum: nodeElect.SeqNum,
	}
	MinerTopGenAns.MasterMiner = support.TransElectNodeInfo(nodeElect.WhiteNodeInfo, common.RoleValidator)
	for _, v := range lastNode {
		MinerTopGenAns.MasterMiner = append(MinerTopGenAns.MasterMiner, support.MakeElectNode(v.Address, len(MinerTopGenAns.MasterMiner), DefauleStock, common.RoleMiner))
		if len(MinerTopGenAns.MasterMiner) >= int(nodeElect.EleCfg.MinerNum) {
			break
		}
	}
	return &MinerTopGenAns

}

func (self *nochoice) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("直接选举方案", "验证者拓扑生成", len(mvrerm.ValidatorList))
	nodeElect := support.NewElelection(nil, mvrerm.ValidatorList, mvrerm.ElectConfig, mvrerm.RandSeed, mvrerm.SeqNum)
	nodeElect.Disorder()
	nodeElect.Sort()
	nodeElect.ProcessBlackNode()
	nodeElect.ProcessWhiteNode()
	lastNode := nodeElect.GetLastNode()
	ValidatorTop := mc.MasterValidatorReElectionRsq{
		SeqNum: nodeElect.SeqNum,
	}
	ValidatorTop.MasterValidator = support.TransElectNodeInfo(nodeElect.WhiteNodeInfo, common.RoleValidator)

	for index, v := range lastNode {
		if len(ValidatorTop.MasterValidator) < int(mvrerm.ElectConfig.ValidatorNum) {
			ValidatorTop.MasterValidator = append(ValidatorTop.MasterValidator, support.MakeElectNode(v.Address, len(ValidatorTop.MasterValidator), DefauleStock, common.RoleValidator))
			continue
		}
		if len(ValidatorTop.BackUpValidator) < int(mvrerm.ElectConfig.BackValidator) {
			ValidatorTop.BackUpValidator = append(ValidatorTop.BackUpValidator, support.MakeElectNode(v.Address, len(ValidatorTop.BackUpValidator), DefauleStock, common.RoleBackupValidator))
			continue
		}
		ValidatorTop.CandidateValidator = append(ValidatorTop.CandidateValidator, support.MakeElectNode(v.Address, index, DefauleStock, common.RoleCandidateValidator))

	}
	return &ValidatorTop
}

func (self *nochoice) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(allNative, topoG)
}

func (self *nochoice) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
