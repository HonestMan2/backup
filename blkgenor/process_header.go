// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"github.com/matrix/go-matrix/consensus/manblk"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/txpoolCache"
	"github.com/pkg/errors"
)

func (p *Process) processBcHeaderGen() error {
	log.INFO(p.logExtraInfo(), "processHeaderGen", "start")
	defer log.INFO(p.logExtraInfo(), "processHeaderGen", "end")
	if p.bcInterval == nil {
		log.ERROR(p.logExtraInfo(), "区块生成阶段", "广播周期信息为空")
		return errors.New("广播周期信息为空")
	}
	originHeader, _, err := p.pm.manblk.Prepare(manblk.CommonBlk, manblk.AVERSION, p.number, p.bcInterval, p.preBlockHash)
	if err != nil {
		log.Error(p.logExtraInfo(), "准备去看失败", err)
		return err
	}

	_, stateDB, receipts, _, finalTxs, _, err := p.pm.manblk.ProcessState(manblk.CommonBlk, manblk.AVERSION, originHeader, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行交易和状态树失败", err)
		return err
	}

	//运行完matrix状态树后，生成root
	block, _, err := p.pm.manblk.Finalize(manblk.CommonBlk, manblk.AVERSION, originHeader, stateDB, finalTxs, nil, receipts, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "Finalize失败", err)
		return err
	}
	finalHeader := block.Header()
	err = p.setSignatures(finalHeader)
	if err != nil {
		return err
	}
	p.sendBroadcastMiningReq(finalHeader, finalTxs)
	return nil
}

func (p *Process) processHeaderGen() error {
	log.INFO(p.logExtraInfo(), "processHeaderGen", "start")
	defer log.INFO(p.logExtraInfo(), "processHeaderGen", "end")
	if p.bcInterval == nil {
		log.ERROR(p.logExtraInfo(), "区块生成阶段", "广播周期信息为空")
		return errors.New("广播周期信息为空")
	}
	originHeader, extraData, err := p.pm.manblk.Prepare(manblk.CommonBlk, manblk.AVERSION, p.number, p.bcInterval, p.preBlockHash)
	if err != nil {
		log.Error(p.logExtraInfo(), "准备阶段失败", err)
		return err
	}

	onlineConsensusResults, ok := extraData.([]*mc.HD_OnlineConsensusVoteResultMsg)

	if !ok {
		log.Error(p.logExtraInfo(), "反射在线状态失败", "")
		return errors.New("反射在线状态失败")
	}

	txsCode, stateDB, receipts, originalTxs, finalTxs, _, err := p.pm.manblk.ProcessState(manblk.CommonBlk, manblk.AVERSION, originHeader, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行交易和状态树失败", err)
		return err
	}

	//运行完matrix状态树后，生成root
	block, _, err := p.pm.manblk.Finalize(manblk.CommonBlk, manblk.AVERSION, originHeader, stateDB, finalTxs, nil, receipts, nil)
	if err != nil {
		log.Error(p.logExtraInfo(), "Finalize失败", err)
		return err
	}
	p.sendHeaderVerifyReq(block.Header(), txsCode, onlineConsensusResults, originalTxs, finalTxs, receipts, stateDB)
	return nil
}

func (p *Process) sendHeaderVerifyReq(header *types.Header, txsCode []*common.RetCallTxN, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg, originalTxs []types.SelfTransaction, finalTxs []types.SelfTransaction, receipts []*types.Receipt, stateDB *state.StateDB) {
	p2pBlock := &mc.HD_BlkConsensusReqMsg{
		Header:                 header,
		TxsCode:                txsCode,
		ConsensusTurn:          p.consensusTurn,
		OnlineConsensusResults: onlineConsensusResults,
		From: ca.GetAddress()}
	//send to local block verify module
	localBlock := &mc.LocalBlockVerifyConsensusReq{BlkVerifyConsensusReq: p2pBlock, OriginalTxs: originalTxs, FinalTxs: finalTxs, Receipts: receipts, State: stateDB}
	if len(originalTxs) > 0 {
		txpoolCache.MakeStruck(originalTxs, header.HashNoSignsAndNonce(), p.number)
	}
	log.INFO(p.logExtraInfo(), "本地发送区块验证请求, root", p2pBlock.Header.Root.TerminalString(), "高度", p.number)
	mc.PublishEvent(mc.BlockGenor_HeaderVerifyReq, localBlock)
	p.startConsensusReqSender(p2pBlock)
}

func (p *Process) sendBroadcastMiningReq(header *types.Header, finalTxs []types.SelfTransaction) {
	sendMsg := &mc.BlockData{Header: header, Txs: finalTxs}
	log.INFO(p.logExtraInfo(), "广播挖矿请求(本地), number", sendMsg.Header.Number, "root", header.Root.TerminalString(), "tx数量", sendMsg.Txs.Len())
	mc.PublishEvent(mc.HD_BroadcastMiningReq, &mc.BlockGenor_BroadcastMiningReqMsg{sendMsg})
}

func (p *Process) setSignatures(header *types.Header) error {

	signHash := header.HashNoSignsAndNonce()
	sign, err := p.signHelper().SignHashWithValidate(signHash.Bytes(), true, p.preBlockHash)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "广播区块生成，签名错误", err)
		return err
	}

	header.Signatures = make([]common.Signature, 0, 1)
	header.Signatures = append(header.Signatures, sign)

	return nil
}

func (p *Process) startConsensusReqSender(req *mc.HD_BlkConsensusReqMsg) {
	p.closeConsensusReqSender()
	sender, err := common.NewResendMsgCtrl(req, p.sendConsensusReqFunc, manparams.BlkPosReqSendInterval, manparams.BlkPosReqSendTimes)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "创建POS完成的req发送器", "失败", "err", err)
		return
	}
	p.consensusReqSender = sender
}

func (p *Process) closeConsensusReqSender() {
	if p.consensusReqSender == nil {
		return
	}
	p.consensusReqSender.Close()
	p.consensusReqSender = nil
}

func (p *Process) sendConsensusReqFunc(data interface{}, times uint32) {
	req, OK := data.(*mc.HD_BlkConsensusReqMsg)
	if !OK {
		log.ERROR(p.logExtraInfo(), "发出区块共识req", "反射消息失败", "次数", times)
		return
	}
	log.INFO(p.logExtraInfo(), "!!!!网络发送区块验证请求, hash", req.Header.HashNoSignsAndNonce(), "tx数量", len(req.TxsCode), "次数", times)
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusReq, req, common.RoleValidator, nil)
}
