// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package olconsensus

import (
	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/msgsend"
	"github.com/matrix/go-matrix/p2p"
)

type NodeOnLineInfo struct {
	Address     common.Address
	Role        common.RoleType
	OnlineState []uint8
}

type TopNodeStateInterface interface {
	GetTopNodeOnlineState() []NodeOnLineInfo
}

type ValidatorAccountInterface interface {
	SignWithValidate(hash []byte, validate bool) (sig common.Signature, err error)
	IsSelfAddress(addr common.Address) bool
}

type MessageSendInterface interface {
	SendNodeMsg(subCode mc.EventCode, msg interface{}, Roles common.RoleType, address []common.Address)
}

type MessageCenterInterface interface {
	SubscribeEvent(aim mc.EventCode, ch interface{}) (event.Subscription, error)
	PublishEvent(aim mc.EventCode, data interface{}) error
}

type StateReaderInterface interface {
	GetMatrixStateDataByHash(key string, hash common.Hash) (interface{}, error)
}

////////////////////////////////////////////////////////////////////
type TopNodeInstance struct {
	signHelper *signhelper.SignHelper
	hd         *msgsend.HD
}

func NewTopNodeInstance(sh *signhelper.SignHelper, hd *msgsend.HD) *TopNodeInstance {
	return &TopNodeInstance{
		signHelper: sh,
		hd:         hd,
	}
}

func (self *TopNodeInstance) GetTopNodeOnlineState() []NodeOnLineInfo {
	onlineStat := make([]NodeOnLineInfo, 0)
	//调用p2p的接口获取节点在线状态
	result := p2p.GetTopNodeAliveInfo(common.RoleValidator | common.RoleBackupValidator)
	for _, value := range result {
		state := NodeOnLineInfo{
			Address:     value.Account,
			Role:        value.Type,
			OnlineState: value.Heartbeats,
		}
		onlineStat = append(onlineStat, state)
		log.Debug("TopNodeOnline", "获取在线状态, node", value.Account, "心跳", value.Heartbeats)
	}

	return onlineStat
}

func (self *TopNodeInstance) SignWithValidate(hash []byte, validate bool) (sig common.Signature, err error) {
	return self.signHelper.SignHashWithValidate(hash, validate)
}

func (self *TopNodeInstance) IsSelfAddress(addr common.Address) bool {
	return ca.GetAddress() == addr
}

func (self *TopNodeInstance) SendNodeMsg(subCode mc.EventCode, msg interface{}, Roles common.RoleType, address []common.Address) {
	self.hd.SendNodeMsg(subCode, msg, Roles, address)
}

func (self *TopNodeInstance) SubscribeEvent(aim mc.EventCode, ch interface{}) (event.Subscription, error) {
	return mc.SubscribeEvent(aim, ch)
}

func (self *TopNodeInstance) PublishEvent(aim mc.EventCode, data interface{}) error {
	return mc.PublishEvent(aim, data)
}
