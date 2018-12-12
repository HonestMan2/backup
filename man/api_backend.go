// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package man

import (
	"context"
	"encoding/json"
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
	"math/big"
	"os"
	"time"
)

// ManAPIBackend implements manapi.Backend for full nodes
type ManAPIBackend struct {
	man *Matrix
	gpo *gasprice.Oracle
}

func (b *ManAPIBackend) ChainConfig() *params.ChainConfig {
	return b.man.chainConfig
}

func (b *ManAPIBackend) CurrentBlock() *types.Block {
	return b.man.blockchain.CurrentBlock()
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

	superGen := new(core.Genesis)
	if err := json.NewDecoder(file).Decode(superGen); err != nil {
		log.Error("ManAPIBackend", "超级区块插入", "文件数据解码错误", err)
		file.Close()
		return common.Hash{}, errors.Errorf("decode config file from \"%s\" err (%v)", filePath, err)
	}
	file.Close()

	superBlock, err := b.man.BlockChain().InsertSuperBlock(superGen)
	if err != nil {
		return common.Hash{}, err
	}

	b.man.EventMux().Post(core.NewMinedBlockEvent{Block: superBlock})
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

//YY
func (b *ManAPIBackend) SignTx(signedTx types.SelfTransaction, chainID *big.Int) (types.SelfTransaction, error) {
	return b.man.signHelper.SignTx(signedTx, chainID)
}

//YY
func (b *ManAPIBackend) SendBroadTx(ctx context.Context, signedTx types.SelfTransaction, bType bool) error {
	return b.man.txPool.AddBroadTx(signedTx, bType)
}

//YY
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
