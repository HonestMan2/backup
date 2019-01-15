// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package man

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/matrix/go-matrix/core/matrixstate"

	"github.com/matrix/go-matrix/base58"

	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/reward/interest"
	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/reward/blkreward"

	"github.com/matrix/go-matrix/internal/manapi"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/reward/selectedreward"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/math"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/bloombits"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/txinterface"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/man/downloader"
	"github.com/matrix/go-matrix/man/gasprice"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/rpc"
	"github.com/pkg/errors"
)

// ManAPIBackend implements manapi.Backend for full nodes
type ManAPIBackend struct {
	man *Matrix
	gpo *gasprice.Oracle
}

func (b *ManAPIBackend) ChainConfig() *params.ChainConfig {
	return b.man.chainConfig
}

func (b *ManAPIBackend) NetRPCService() *manapi.PublicNetAPI {
	return b.man.netRPCService
}

func (b *ManAPIBackend) Config() *Config {
	return b.man.config
}

func (b *ManAPIBackend) CurrentBlock() *types.Block {
	return b.man.blockchain.CurrentBlock()
}

func (b *ManAPIBackend) Genesis() *types.Block {
	return b.man.blockchain.Genesis()
}

func (b *ManAPIBackend) SetHead(number uint64) {
	b.man.protocolManager.downloader.Cancel()
	b.man.blockchain.SetHead(number)
}

func (b *ManAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.man.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.man.blockchain.CurrentBlock().Header(), nil
	}
	return b.man.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *ManAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.man.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.man.blockchain.CurrentBlock(), nil
	}
	return b.man.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *ManAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.man.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.man.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *ManAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.man.blockchain.GetBlockByHash(hash), nil
}
func (b *ManAPIBackend) GetState() (*state.StateDB, error) {
	return b.man.BlockChain().State()
}
func (b *ManAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.man.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.man.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *ManAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.man.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.man.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *ManAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.man.blockchain.GetTdByHash(blockHash)
}

func (b *ManAPIBackend) GetEVM(ctx context.Context, msg txinterface.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(common.MainAccount, msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg.From(), msg.GasPrice(), header, b.man.BlockChain(), nil)
	return vm.NewEVM(context, state, b.man.chainConfig, vmCfg), vmError, nil
}

func (b *ManAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.man.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *ManAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.man.BlockChain().SubscribeChainEvent(ch)
}

func (b *ManAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.man.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *ManAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.man.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *ManAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.man.BlockChain().SubscribeLogsEvent(ch)
}

func (b *ManAPIBackend) ImportSuperBlock(ctx context.Context, filePath string) (common.Hash, error) {
	log.Info("ManAPIBackend", "收到超级区块插入", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("ManAPIBackend", "超级区块插入", "读取配置文件异常", "err", err)
		return common.Hash{}, errors.Errorf("reader config file from \"%s\" err (%v)", filePath, err)
	}

	matrixGenesis := new(core.Genesis1)
	if err := json.NewDecoder(file).Decode(matrixGenesis); err != nil {
		log.Error("ManAPIBackend", "超级区块插入", "文件数据解码错误", err)
		file.Close()
		return common.Hash{}, errors.Errorf("decode config file from \"%s\" err (%v)", filePath, err)
	}
	file.Close()

	superGen := new(core.Genesis)
	core.ManGenesisToEthGensis(matrixGenesis, superGen)

	superBlock, err := b.man.BlockChain().InsertSuperBlock(superGen, true)
	if err != nil {
		return common.Hash{}, err
	}
	for i := 0; i < 3; i++ {
		b.man.protocolManager.AllBroadcastBlock(superBlock, true)
		time.Sleep(100 * time.Millisecond)
	}

	return superBlock.Hash(), nil
}

//TODO 调用该方法的时候应该返回错误的切片
func (b *ManAPIBackend) SendTx(ctx context.Context, signedTx types.SelfTransaction) error {
	return b.man.txPool.AddRemote(signedTx)
}

func (b *ManAPIBackend) GetPoolTransactions() (types.SelfTransactions, error) {
	pending, err := b.man.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.SelfTransactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *ManAPIBackend) GetPoolTransaction(hash common.Hash) types.SelfTransaction {
	npooler, nerr := b.man.TxPool().GetTxPoolByType(types.NormalTxIndex)
	if nerr == nil {
		npool, ok := npooler.(*core.NormalTxPool)
		if ok {
			tx := npool.Get(hash)
			if tx == nil {
				return nil
			}
			return tx
		} else {
			return nil
		}
	}
	return nil
}

func (b *ManAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	npooler, nerr := b.man.TxPool().GetTxPoolByType(types.NormalTxIndex)
	if nerr == nil {
		npool, ok := npooler.(*core.NormalTxPool)
		if ok {
			return npool.State().GetNonce(addr), nil
		} else {
			return 0, errors.New("GetPoolNonce() unknown txpool")
		}
	}
	return 0, nerr
}

func (b *ManAPIBackend) Stats() (pending int, queued int) {
	bpooler, err := b.man.TxPool().GetTxPoolByType(types.BroadCastTxIndex)
	if err == nil {
		_, ok := bpooler.(*core.BroadCastTxPool)
		if ok {
			//_,btxs = bpool.Content()
		} else {
			queued = 0
		}
	}
	npooler, nerr := b.man.TxPool().GetTxPoolByType(types.NormalTxIndex)
	if nerr == nil {
		npool, ok := npooler.(*core.NormalTxPool)
		if ok {
			pending, _ = npool.Stats()
		} else {
			pending = 0
		}
	}
	return pending, queued
}

//TODO 应该将返回值加入切片中否则以后多一种交易就要添加一个返回值
func (b *ManAPIBackend) TxPoolContent() (ntxs map[common.Address]types.SelfTransactions, btxs map[common.Address]types.SelfTransactions) {
	ntxs = make(map[common.Address]types.SelfTransactions)
	btxs = make(map[common.Address]types.SelfTransactions)
	bpooler, err := b.man.TxPool().GetTxPoolByType(types.BroadCastTxIndex)
	if err == nil {
		_, ok := bpooler.(*core.BroadCastTxPool)
		if ok {
			//_,btxs = bpool.Content()
		} else {
			btxs = nil
		}
	}
	npooler, nerr := b.man.TxPool().GetTxPoolByType(types.NormalTxIndex)
	if nerr == nil {
		npool, ok := npooler.(*core.NormalTxPool)
		if ok {
			txlist := npool.Content()
			for k, vlist := range txlist {
				txser := make([]types.SelfTransaction, 0)
				for _, v := range vlist {
					txser = append(txser, v)
				}
				if vs, ok := ntxs[k]; !ok {
					txser = append(txser, vs...)
				}
				ntxs[k] = txser
			}
		} else {
			ntxs = nil
		}
	}
	return ntxs, btxs
}

func (b *ManAPIBackend) SubscribeNewTxsEvent(ch chan core.NewTxsEvent) event.Subscription {
	return b.man.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *ManAPIBackend) Downloader() *downloader.Downloader {
	return b.man.Downloader()
}

func (b *ManAPIBackend) ProtocolVersion() int {
	return b.man.ManVersion()
}

func (b *ManAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *ManAPIBackend) ChainDb() mandb.Database {
	return b.man.ChainDb()
}

func (b *ManAPIBackend) EventMux() *event.TypeMux {
	return b.man.EventMux()
}

func (b *ManAPIBackend) AccountManager() *accounts.Manager {
	return b.man.AccountManager()
}

func (b *ManAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.man.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *ManAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.man.bloomRequests)
	}
}

//
func (b *ManAPIBackend) SignTx(signedTx types.SelfTransaction, chainID *big.Int, blkHash common.Hash, signHeight uint64, usingEntrust bool) (types.SelfTransaction, error) {
	return b.man.signHelper.SignTx(signedTx, chainID, blkHash, signHeight, usingEntrust)
}

//
func (b *ManAPIBackend) SendBroadTx(ctx context.Context, signedTx types.SelfTransaction, bType bool) error {
	return b.man.txPool.AddBroadTx(signedTx, bType)
}

//
func (b *ManAPIBackend) FetcherNotify(hash common.Hash, number uint64) {

	/*
		2018-09-29 因为改到其他地方实现，所以此方法没有被调用。废弃
	*/
	return
	ids := ca.GetRolesByGroup(common.RoleValidator)
	for _, id := range ids {
		peer := b.man.protocolManager.Peers.Peer(id.String())
		b.man.protocolManager.fetcher.Notify(id.String(), hash, number, time.Now(), peer.RequestOneHeader, peer.RequestBodies)
	}
}

func (b *ManAPIBackend) GetDepositAccount(signAccount common.Address, blockHash common.Hash) (common.Address, error) {
	depositAccount, _, err := b.man.blockchain.GetA0AccountFromAnyAccount(signAccount, blockHash)
	return depositAccount, err
}

type TimeZone struct {
	Start uint64
	Stop  uint64
}

type RewardMount struct {
	Account  string
	Reward   *big.Int
	VipLevel common.VIPRoleType
	Stock    uint16
}
type InterestReward struct {
	Account  string
	Reward   *big.Int
	VipLevel common.VIPRoleType
	Stock    uint16
	Deposit  *big.Int
}

type AllReward struct {
	Time      TimeZone
	Miner     []RewardMount
	Validator []RewardMount
	Interest  []InterestReward
}

func (b *ManAPIBackend) GetFutureRewards(ctx context.Context, number rpc.BlockNumber) (interface{}, error) {

	state, _, err := b.StateAndHeaderByNumber(ctx, number)
	if state == nil || err != nil {
		return nil, err
	}
	bcInterval, err := manparams.GetBCIntervalInfoByNumber(uint64(number))
	if nil != err {
		return nil, err
	}
	latestElectNum := bcInterval.GetLastReElectionNumber()
	var allReward AllReward
	allReward.Time.Start = latestElectNum
	allReward.Time.Stop = latestElectNum + bcInterval.GetReElectionInterval()
	depositNodes, err := ca.GetElectedByHeight(new(big.Int).SetUint64(latestElectNum))
	if nil != err {
		return nil, err
	}
	if 0 == len(depositNodes) {
		return nil, err
	}
	originElectNodes, err := matrixstate.GetElectGraph(state)
	if err != nil {
		return nil, err
	}

	if originElectNodes == nil {
		errors.New("获取初选拓扑图结构为nil")
		return nil, err
	}
	if 0 == len(originElectNodes.ElectList) {
		errors.New("get获取初选列表为空")
		return nil, err
	}

	RewardMap, err := b.calcFutureBlkReward(state, latestElectNum+1, bcInterval, common.RoleMiner)
	if nil != err {
		return nil, err
	}
	minerRewardList := make([]RewardMount, 0)
	for k, v := range RewardMap {
		obj := RewardMount{Account: base58.Base58EncodeToString("MAN", k), Reward: v}
		for _, d := range originElectNodes.ElectList {
			if d.Account.Equal(k) {
				obj.VipLevel = d.VIPLevel
				obj.Stock = d.Stock
			}
		}
		minerRewardList = append(minerRewardList, obj)
	}
	allReward.Miner = minerRewardList
	validatorMap, err := b.calcFutureBlkReward(state, latestElectNum+1, bcInterval, common.RoleValidator)
	if nil != err {
		return nil, err
	}
	ValidatorRewardList := make([]RewardMount, 0)
	for k, v := range validatorMap {
		obj := RewardMount{Account: base58.Base58EncodeToString("MAN", k), Reward: v}
		for _, d := range originElectNodes.ElectList {
			if d.Account.Equal(k) {
				obj.VipLevel = d.VIPLevel
				obj.Stock = d.Stock
			}
		}
		ValidatorRewardList = append(ValidatorRewardList, obj)
	}
	allReward.Validator = ValidatorRewardList
	interestReward := interest.New(state)
	if nil == interestReward {
		return nil, err
	}
	interestCalcMap := interestReward.GetInterest(state, latestElectNum)
	interestNum := bcInterval.GetReElectionInterval() / interestReward.CalcInterval
	interestRewardList := make([]InterestReward, 0)
	for k, v := range interestCalcMap {
		allInterest := new(big.Int).Mul(v, new(big.Int).SetUint64(interestNum))
		obj := InterestReward{Account: base58.Base58EncodeToString("MAN", k), Reward: allInterest}

		for _, d := range depositNodes {
			if d.Address.Equal(k) {
				obj.Deposit = d.Deposit
			}
		}
		for _, d := range originElectNodes.ElectList {
			if d.Account.Equal(k) {
				obj.VipLevel = d.VIPLevel
				obj.Stock = d.Stock
			}
		}
		interestRewardList = append(interestRewardList, obj)
	}
	allReward.Interest = interestRewardList
	return allReward, nil
}

func (b *ManAPIBackend) calcFutureBlkReward(state *state.StateDB, latestElectNum uint64, bcInterval *mc.BCIntervalInfo, roleType common.RoleType) (map[common.Address]*big.Int, error) {
	selected := selectedreward.SelectedReward{}
	currentTop, originElectNodes, err := selected.GetTopAndDeposit(b.man.BlockChain(), state, latestElectNum, roleType)
	if nil != err {
		return nil, err
	}
	br := blkreward.New(b.man.BlockChain(), state)
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(br.GetRewardCfg().RewardMount.MinerMount), util.ManPrice)
	halfNum := br.GetRewardCfg().RewardMount.MinerHalf
	RewardMap := make(map[common.Address]*big.Int)

	var rewardAddr common.Address
	if roleType == common.RoleMiner {
		rewardAddr = common.BlkMinerRewardAddress
	} else {
		rewardAddr = common.BlkValidatorRewardAddress
	}
	for num := latestElectNum; num < bcInterval.GetNextReElectionNumber(latestElectNum); num++ {

		if bcInterval.IsBroadcastNumber(num) {
			continue
		}
		validatorReward := br.CalcRewardMountByNumber(RewardMan, uint64(num), halfNum, rewardAddr)
		minerOutAmount, electedMount, _ := br.CalcMinerRateMount(validatorReward)

		selectedNodesDeposit := selected.CaclSelectedDeposit(currentTop, originElectNodes, 0)
		if 0 == len(selectedNodesDeposit) {
			return nil, errors.New("获取参与的抵押列表错误")
		}

		electRewards := util.CalcStockRate(electedMount, selectedNodesDeposit)
		if 0 == len(electRewards) {
			return nil, errors.New("计算的参与奖励为nil")
		}
		for k := range electRewards {
			tmp := new(big.Int).SetUint64(uint64(len(electRewards)))
			util.SetAccountRewards(electRewards, k, new(big.Int).Div(minerOutAmount, tmp))
		}
		util.MergeReward(RewardMap, electRewards)

	}

	return RewardMap, nil
}
