// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package matrixwork

import (
	"errors"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MatrixAINetwork/go-matrix/accounts/abi"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/common/hexutil"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/core/vm"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/baseinterface"
)

type ChainReader interface {
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
	GetBlockByHash(hash common.Hash) *types.Block
	Engine(version []byte) consensus.Engine
	GetHeader(common.Hash, uint64) *types.Header
	Processor(version []byte) core.Processor
}
type txPoolReader interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Address]types.SelfTransactions, error)
}

var packagename string = "matrixwork"
var (
	depositDef = ` [{"constant": true,"inputs": [],"name": "getDepositList","outputs": [{"name": "","type": "address[]"}],"payable": false,"stateMutability": "view","type": "function"},
			{"constant": true,"inputs": [{"name": "addr","type": "address"}],"name": "getDepositInfo","outputs": [{"name": "","type": "uint256"},{"name": "","type": "bytes"},{"name": "","type": "uint256"}],"payable": false,"stateMutability": "view","type": "function"},
    		{"constant": false,"inputs": [{"name": "nodeID","type": "bytes"}],"name": "valiDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [{"name": "nodeID","type": "bytes"}],"name": "minerDeposit","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
    		{"constant": false,"inputs": [],"name": "withdraw","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
    		{"constant": false,"inputs": [],"name": "refund","outputs": [],"payable": false,"stateMutability": "nonpayable","type": "function"},
			{"constant": false,"inputs": [{"name": "addr","type": "address"}],"name": "interestAdd","outputs": [],"payable": true,"stateMutability": "payable","type": "function"},
			{"constant": false,"inputs": [{"name": "addr","type": "address"}],"name": "getinterest","outputs": [],"payable": false,"stateMutability": "payable","type": "function"}]`

	depositAbi, Abierr = abi.JSON(strings.NewReader(depositDef))
)

// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config *params.ChainConfig
	signer types.Signer

	State *state.StateDBManage // apply state changes here
	//ancestors *set.Set       // ancestor set (used for checking uncle parent validity)
	//family    *set.Set       // family set (used for checking uncle invalidity)
	//uncles    *set.Set       // uncle set
	tcount  int           // tx count in cycle
	gasPool *core.GasPool // available gas used to pack transactions

	Block *types.Block // the new block

	header *types.Header
	bc     ChainReader

	random   *baseinterface.Random
	txs      []types.CoinSelfTransaction
	Receipts []types.CoinReceipts

	transer      []types.SelfTransaction
	recpts       []*types.Receipt

	createdAt time.Time
}
type coingasUse struct {
	mapcoin  map[string]*big.Int
	mapprice map[string]*big.Int
	mu       sync.RWMutex
}

var mapcoingasUse coingasUse = coingasUse{mapcoin: make(map[string]*big.Int), mapprice: make(map[string]*big.Int)}

func (cu *coingasUse) setCoinGasUse(txer types.SelfTransaction, gasuse uint64) {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	coin := txer.GetTxCurrency()
	coin = params.MAN_COIN
	gasAll := new(big.Int).SetUint64(gasuse)
	priceAll := txer.GasPrice()
	if gas, ok := cu.mapcoin[coin]; ok {
		gasAll = new(big.Int).Add(gasAll, gas)
	}
	cu.mapcoin[coin] = gasAll
	if _, ok := cu.mapprice[coin]; !ok {
		cu.mapprice[coin] = priceAll
	}
}
func (cu *coingasUse) getCoinGasPrice(typ string) *big.Int {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	price, ok := cu.mapprice[typ]
	if !ok {
		price = new(big.Int).SetUint64(0)
	}
	return price
}

func (cu *coingasUse) getCoinGasUse(typ string) *big.Int {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	gas, ok := cu.mapcoin[typ]
	if !ok {
		gas = new(big.Int).SetUint64(0)
	}
	return gas
}
func (cu *coingasUse) clearmap() {
	cu.mu.Lock()
	defer cu.mu.Unlock()
	cu.mapcoin = make(map[string]*big.Int)
	cu.mapprice = make(map[string]*big.Int)
}
func NewWork(config *params.ChainConfig, bc ChainReader, gasPool *core.GasPool, header *types.Header) (*Work, error) {

	Work := &Work{
		config:  config,
		signer:  types.NewEIP155Signer(config.ChainId),
		gasPool: gasPool,
		header:  header,
		bc:      bc,
	}
	var err error

	Work.State, err = bc.StateAt(bc.GetBlockByHash(header.ParentHash).Root())

	if err != nil {
		return nil, err
	}
	return Work, nil
}

//func (env *Work) commitTransactions(mux *event.TypeMux, txs *types.TransactionsByPriceAndNonce, bc *core.BlockChain, coinbase common.Address) (listN []uint32, retTxs []types.SelfTransaction) {
func (env *Work) commitTransactions(mux *event.TypeMux, txser map[common.Address]types.SelfTransactions, coinbase common.Address) (listret []*common.RetCallTxN, retTxs []types.SelfTransaction) {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}

	var coalescedLogs []types.CoinLogs
	tmpRetmap := make(map[byte][]uint32)
	for _, txers := range txser {
		//txs := types.GetCoinTX(txers)
		for _,txer := range txers{
			// If we don't have enough gas for any further transactions then we're done
			if env.gasPool.Gas() < params.TxGas {
				log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
				break
			}
			if txer.GetTxNLen() == 0 {
				log.Info("work.go commitTransactions err: tx.N is nil")
				continue
			}
			// We use the eip155 signer regardless of the current hf.
			from, _ := txer.GetTxFrom()

			// Start executing the transaction
			env.State.Prepare(txer.Hash(), common.Hash{}, env.tcount)
			err, logs := env.commitTransaction(txer, env.bc, coinbase, env.gasPool)
			isSkipFrom := false
			switch err {
			case core.ErrGasLimitReached:
				// Pop the current out-of-gas transaction without shifting in the next from the account
				log.Trace("Gas limit exceeded for current block", "sender", from)
				isSkipFrom = true
			case core.ErrNonceTooLow:
				// New head notification data race between the transaction pool and miner, shift
				log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", txer.Nonce())
			case core.ErrNonceTooHigh:
				// Reorg notification data race between the transaction pool and miner, skip account =
				log.Trace("Skipping account with hight nonce", "sender", from, "nonce", txer.Nonce())
				isSkipFrom = true
			case nil:
				// Everything ok, collect the logs and shift in the next transaction from the same account
				if txer.GetTxNLen() != 0 {
					n := txer.GetTxN(0)
					if listN, ok := tmpRetmap[txer.TxType()]; ok {
						listN = append(listN, n)
						tmpRetmap[txer.TxType()] = listN
					} else {
						listN := make([]uint32, 0)
						listN = append(listN, n)
						tmpRetmap[txer.TxType()] = listN
					}
					retTxs = append(retTxs, txer)
				}
				coalescedLogs = append(coalescedLogs, types.CoinLogs{txer.GetTxCurrency(),logs})
				env.tcount++
			default:
				// Strange error, discard the transaction and get the next in line (note, the
				// nonce-too-high clause will prevent us from executing in vain).
				log.Debug("Transaction failed, account skipped", "hash", txer.Hash(), "err", err)
			}
			if isSkipFrom{
				break
			}
		}
	}
	for t, n := range tmpRetmap {
		ts := common.RetCallTxN{t, n}
		listret = append(listret, &ts)
	}
	if len(coalescedLogs) > 0 || env.tcount > 0 {
		// make a copy, the state caches the logs and these logs get "upgraded" from pending to mined
		// logs by filling in the block hash when the block was mined by the local miner. This can
		// cause a race condition if a log was "upgraded" before the PendingLogsEvent is processed.
		cpy := make([]types.CoinLogs, len(coalescedLogs))
		for i, l := range coalescedLogs {
			cpy[i] = *new(types.CoinLogs)
			cpy[i]=l
		}
		go func(logs []types.CoinLogs, tcount int) {
			if len(logs) > 0 {
				mux.Post(core.PendingLogsEvent{Logs: logs})
			}
			if tcount > 0 {
				mux.Post(core.PendingStateEvent{})
			}
		}(cpy, env.tcount)
	}
	return listret, retTxs
}

func (env *Work) commitTransaction(tx types.SelfTransaction, bc ChainReader, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	snap := env.State.Snapshot(tx.GetTxCurrency())
	var snap1 map[byte]int
	if tx.GetTxCurrency()!=params.MAN_COIN {
		snap1 = env.State.Snapshot(params.MAN_COIN)
	}
	receipt, _, _, err := core.ApplyTransaction(env.config, bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil{
		env.State.RevertToSnapshot(tx.GetTxCurrency(), snap)
		if tx.GetTxCurrency()!=params.MAN_COIN {
			env.State.RevertToSnapshot(params.MAN_COIN, snap1)
		}
		return err, nil
	}
	env.transer = append(env.transer, tx)
	env.recpts = append(env.recpts, receipt)
	mapcoingasUse.setCoinGasUse(tx, receipt.GasUsed)
	return nil, receipt.Logs
}
func (env *Work) s_commitTransaction(tx types.SelfTransaction, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	env.State.Prepare(tx.Hash(), common.Hash{}, env.tcount)
	snap := env.State.Snapshot(tx.GetTxCurrency())
	receipt, _, _, err := core.ApplyTransaction(env.config, env.bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil {
		log.Error("s_commitTransaction commit err. ","err", err)
		env.State.RevertToSnapshot(tx.GetTxCurrency(), snap)
		return err, nil
	}
	tmps := make([]types.SelfTransaction, 0)
	tmps = append(tmps, tx)
	tmps = append(tmps, env.transer...)
	env.transer = tmps

	tmpr := make([]*types.Receipt, 0)
	tmpr = append(tmpr, receipt)
	tmpr = append(tmpr, env.recpts...)
	env.recpts = tmpr
	env.tcount++
	return nil, receipt.Logs
}

//Leader
var lostCnt int = 0

type retStruct struct {
	no  []uint32
	txs []*types.Transaction
}

func (env *Work) ProcessTransactions(mux *event.TypeMux, tp txPoolReader, upTime map[common.Address]uint64) (listret []*common.RetCallTxN, originalTxs []types.SelfTransaction, finalTxs []types.SelfTransaction) {
	pending, err := tp.Pending()
	if err != nil {
		log.Error("Failed to fetch pending transactions", "err", err)
		return nil, nil, nil
	}
	mapcoingasUse.clearmap()
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	log.Info("work", "关键时间点", "开始执行交易", "time", time.Now(), "块高", env.header.Number)
	listret, originalTxs = env.commitTransactions(mux, pending, common.Address{})
	finalTxs = append(finalTxs, originalTxs...)
	tmps := make([]types.SelfTransaction, 0)
	from := make([]common.Address, 0)
	for _, tx := range originalTxs {
		from = append(from, tx.From())
	}
	log.Info("work", "关键时间点", "执行交易完成，开始执行奖励", "time", time.Now(), "块高", env.header.Number,"tx num ",len(originalTxs))
	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, upTime, from, mapcoingasUse.getCoinGasUse(params.MAN_COIN).Uint64())
	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		err, _ := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
		if err != nil {
			log.Error("work.go", "ProcessTransactions:::reward Tx call Error", err)
			continue
		}
		tmptxs := make([]types.SelfTransaction, 0)
		tmptxs = append(tmptxs, tx)
		tmptxs = append(tmptxs, tmps...)
		tmps = tmptxs
	}
	tmps = append(tmps, finalTxs...)
	finalTxs = tmps
	env.txs,env.Receipts =types.GetCoinTXRS(env.transer,env.recpts)
	log.Info("work", "关键时间点", "奖励执行完成", "time", time.Now(), "块高", env.header.Number)
	return
}

func (env *Work) makeTransaction(rewarts []common.RewarTx) (txers []types.SelfTransaction) {
	for _, rewart := range rewarts {
		sorted_keys := make([]string, 0)
		for k, _ := range rewart.To_Amont {
			sorted_keys = append(sorted_keys, k.String())
		}
		sort.Strings(sorted_keys)
		extra := make([]*types.ExtraTo_tr, 0)
		var to common.Address
		var value *big.Int
		databytes := make([]byte, 0)
		isfirst := true
		for _, addr := range sorted_keys {
			k := common.HexToAddress(addr)
			v := rewart.To_Amont[k]
			if isfirst {
				if rewart.RewardTyp == common.RewardInterestType {
					if k != common.ContractAddress {
						databytes = append(databytes, depositAbi.Methods["interestAdd"].Id()...)
						tmpbytes, _ := depositAbi.Methods["interestAdd"].Inputs.Pack(k)
						databytes = append(databytes, tmpbytes...)
						to = common.ContractAddress
						value = v
					} else {
						continue
					}
				} else {
					to = k
					value = v
				}
				isfirst = false
				continue
			}
			tmp := new(types.ExtraTo_tr)
			vv := new(big.Int).Set(v)
			var kk common.Address = k
			tmp.To_tr = &kk
			tmp.Value_tr = (*hexutil.Big)(vv)
			if rewart.RewardTyp == common.RewardInterestType {
				if kk != common.ContractAddress {
					bytes := make([]byte, 0)
					bytes = append(bytes, depositAbi.Methods["interestAdd"].Id()...)
					tmpbytes, _ := depositAbi.Methods["interestAdd"].Inputs.Pack(k)
					bytes = append(bytes, tmpbytes...)
					b := hexutil.Bytes(bytes)
					tmp.Input_tr = &b
					tmp.To_tr = &common.ContractAddress
				} else {
					continue
				}
			}
			extra = append(extra, tmp)
		}
		tx := types.NewTransactions(env.State.GetNonce(rewart.CoinType,rewart.Fromaddr), to, value, 0, new(big.Int), databytes,nil,nil,nil, extra, 0, env.rewardTypetransformation(rewart.RewardTyp), 0,rewart.CoinType,0)
		tx.SetFromLoad(rewart.Fromaddr)
		txers = append(txers, tx)
	}
	return
}
func (env *Work) rewardTypetransformation(inputType byte) byte {
	switch inputType {
	case common.RewardMinerType:
		return common.ExtraUnGasMinerTxType
	case common.RewardValidatorType:
		return common.ExtraUnGasValidatorTxType
	case common.RewardInterestType:
		return common.ExtraUnGasInterestTxType
	case common.RewardTxsType:
		return common.ExtraUnGasTxsType
	case common.RewardLotteryType:
		return common.ExtraUnGasLotteryTxType
	default:
		log.Error("work.go","rewardTypetransformation:Unknown reward type.",inputType)
		panic("rewardTypetransformation:Unknown reward type.")
		return common.ExtraUnGasMinerTxType
	}
}
//Broadcast
func (env *Work) ProcessBroadcastTransactions(mux *event.TypeMux, txs []types.CoinSelfTransaction) {
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	mapcoingasUse.clearmap()
	for _, tx := range txs {
		for _,t:=range  tx.Txser{
		env.commitTransaction(t, env.bc, common.Address{}, nil)
		}
	}

	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, nil, nil, mapcoingasUse.getCoinGasUse(params.MAN_COIN).Uint64())
	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		err, _ := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
		if err != nil {
			log.Error("work.go", "ProcessTransactions:::reward Tx call Error", err)
		}
	}
	env.txs,env.Receipts =types.GetCoinTXRS(env.transer,env.recpts)		
	return
}

func (env *Work) ConsensusTransactions(mux *event.TypeMux, txs []types.CoinSelfTransaction, upTime map[common.Address]uint64) error {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}
	mapcoingasUse.clearmap()
	var coalescedLogs []types.CoinLogs
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	from := make([]common.Address, 0)
	log.Info("work", "关键时间点", "开始执行交易", "time", time.Now(), "块高", env.header.Number)
	for _, tx := range txs {
		// If we don't have enough gas for any further transactions then we're done
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			return errors.New("Not enough gas for further transactions")
		}
		// Start executing the transaction
		for _, t := range tx.Txser {
			env.State.Prepare(t.Hash(), common.Hash{}, env.tcount)
			err, logs := env.commitTransaction(t, env.bc, common.Address{}, env.gasPool)
			if err == nil {
				env.tcount++
				coalescedLogs = append(coalescedLogs,types.CoinLogs{t.GetTxCurrency(),logs})
			} else {
				return err
			}
			from = append(from,t.From())
		}
	}
	log.Info("work", "关键时间点", "执行交易完成，开始执行奖励", "time", time.Now(), "块高", env.header.Number)
	rewart := env.bc.Processor(env.header.Version).ProcessReward(env.State, env.header, upTime, from, mapcoingasUse.getCoinGasUse(params.MAN_COIN).Uint64())
	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		err, _ := env.s_commitTransaction(tx, common.Address{}, new(core.GasPool).AddGas(0))
		if err != nil {
			return err
		}
	}
	env.txs,env.Receipts =types.GetCoinTXRS(env.transer,env.recpts)
	if len(coalescedLogs) > 0 || env.tcount > 0 {
		go func(logs []types.CoinLogs, tcount int) {
			if len(logs) > 0 {
				mux.Post(core.PendingLogsEvent{Logs: logs})
			}
			if tcount > 0 {
				mux.Post(core.PendingStateEvent{})
			}
		}(coalescedLogs, env.tcount)
	}
	log.Info("work", "关键时间点", "奖励执行完成", "time", time.Now(), "块高", env.header.Number)
	return nil
}
func (env *Work) GetTxs() []types.CoinSelfTransaction {
	return env.txs
}
