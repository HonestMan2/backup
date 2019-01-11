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

	"github.com/matrix/go-matrix/accounts/abi"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/reward/blkreward"
	"github.com/matrix/go-matrix/reward/interest"
	"github.com/matrix/go-matrix/reward/lottery"
	"github.com/matrix/go-matrix/reward/slash"
	"github.com/matrix/go-matrix/reward/txsreward"
)

type ChainReader interface {
	StateAt(root []common.CoinRoot) (*state.StateDBManage, error)
	GetBlockByHash(hash common.Hash) *types.Block
	GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error)
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
	priceAll := new(big.Int).SetUint64(params.TxGasPrice) //txer.GasPrice()
	if gas, ok := cu.mapcoin[coin]; ok {
		gasAll = new(big.Int).Add(gasAll, gas)
	}
	cu.mapcoin[coin] = gasAll

	if _, ok := cu.mapprice[coin]; !ok {
		if priceAll.Cmp(new(big.Int).SetUint64(params.TxGasPrice)) >= 0 {
			cu.mapprice[coin] = priceAll
		}
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
func NewWork(config *params.ChainConfig, bc ChainReader, gasPool *core.GasPool, header *types.Header, random *baseinterface.Random) (*Work, error) {

	Work := &Work{
		config:  config,
		signer:  types.NewEIP155Signer(config.ChainId),
		gasPool: gasPool,
		header:  header,
		random:  random,
	}
	var err error

	Work.State, err = bc.StateAt(bc.GetBlockByHash(header.ParentHash).Root())

	if err != nil {
		return nil, err
	}
	return Work, nil
}

//func (env *Work) commitTransactions(mux *event.TypeMux, txs *types.TransactionsByPriceAndNonce, bc *core.BlockChain, coinbase common.Address) (listN []uint32, retTxs []types.SelfTransaction) {
func (env *Work) commitTransactions(mux *event.TypeMux, txser types.SelfTransactions, bc *core.BlockChain, coinbase common.Address) (listret []*common.RetCallTxN, retTxs []types.SelfTransaction) {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}

	var coalescedLogs []types.CoinLogs
	tmpRetmap := make(map[byte][]uint32)
	txs := types.GetCoinTX(txser)
	for _, txers := range txs {
		for _,txer := range txers.Txser{
			// If we don't have enough gas for any further transactions then we're done
			if env.gasPool.Gas() < params.TxGas {
				log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
				break
			}
			if txer.GetTxNLen() == 0 {
				log.Info("file work func commitTransactions err: tx.N is nil")
				continue
			}
			// We use the eip155 signer regardless of the current hf.
			from, _ := txer.GetTxFrom()

			// Start executing the transaction
			env.State.Prepare(txer.Hash(), common.Hash{}, env.tcount)
			err, logs := env.commitTransaction(txer, bc, coinbase, env.gasPool)
			switch err {
			case core.ErrGasLimitReached:
				// Pop the current out-of-gas transaction without shifting in the next from the account
				log.Trace("Gas limit exceeded for current block", "sender", from)
			case core.ErrNonceTooLow:
				// New head notification data race between the transaction pool and miner, shift
				log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", txer.Nonce())
			case core.ErrNonceTooHigh:
				// Reorg notification data race between the transaction pool and miner, skip account =
				log.Trace("Skipping account with hight nonce", "sender", from, "nonce", txer.Nonce())
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

func (env *Work) commitTransaction(tx types.SelfTransaction, bc *core.BlockChain, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	snap := env.State.Snapshot(tx.GetTxCurrency())
	var snap1 map[byte]int
	if tx.GetTxCurrency()!=params.MAN_COIN {
		snap1 = env.State.Snapshot(params.MAN_COIN)
	}
	receipt, _, _, err := core.ApplyTransaction(env.config, bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil {
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
func (env *Work) s_commitTransaction(tx types.SelfTransaction, bc *core.BlockChain, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	env.State.Prepare(tx.Hash(), common.Hash{}, env.tcount)
	snap := env.State.Snapshot(tx.GetTxCurrency())
	receipt, _, _, err := core.ApplyTransaction(env.config, bc, &coinbase, gp, env.State, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil {
		log.Info("file work", "func s_commitTransaction", err)
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

func (env *Work) Reverse(s []common.RewarTx) []common.RewarTx {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func (env *Work) ProcessTransactions(mux *event.TypeMux, tp *core.TxPoolManager, bc *core.BlockChain, upTime map[common.Address]uint64) (listret []*common.RetCallTxN, originalTxs []types.SelfTransaction, finalTxs []types.SelfTransaction) {
	pending, err := tp.Pending()
	if err != nil {
		log.Error("Failed to fetch pending transactions", "err", err)
		return nil, nil, nil
	}
	mapcoingasUse.clearmap()
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	listTx := make(types.SelfTransactions, 0)
	for _, txser := range pending {
		listTx = append(listTx, txser...)
	}
	listret, originalTxs = env.commitTransactions(mux, listTx, bc, common.Address{})
	finalTxs = append(finalTxs, originalTxs...)
	tmps := make([]types.SelfTransaction, 0)
	from := make([]common.Address, 0)
	for _, tx := range originalTxs {
		from = append(from, tx.From())
	}
	rewart := env.CalcRewardAndSlash(bc, upTime, from)

	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		//fmt.Printf("验证者%s\n",env.State.Dump(tx.GetTxCurrency(),tx.From()))
		err, _ := env.s_commitTransaction(tx, bc, common.Address{}, new(core.GasPool).AddGas(0))
		if err != nil {
			log.Error("file work", "func ProcessTransactions:::reward Tx call Error", err)
			continue
		}
		tmptxs := make([]types.SelfTransaction, 0)
		tmptxs = append(tmptxs, tx)
		tmptxs = append(tmptxs, tmps...)
		tmps = tmptxs
		//fmt.Printf("验证者%s\n",env.State.Dump(tx.GetTxCurrency(),tx.From()))
	}
	tmps = append(tmps, finalTxs...)
	finalTxs = tmps
	env.txs,env.Receipts =types.GetCoinTXRS(env.transer,env.recpts)
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
				if rewart.RewardTyp == common.RewardInerestType {
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
			if rewart.RewardTyp == common.RewardInerestType {
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
		tx := types.NewTransactions(env.State.GetNonce(rewart.CoinType, rewart.Fromaddr), to, value, 0, new(big.Int), databytes, extra, 0, common.ExtraUnGasTxType, 0)
		tx.SetFromLoad(rewart.Fromaddr)
		tx.SetTxS(big.NewInt(1))
		tx.SetTxV(big.NewInt(1))
		tx.SetTxR(big.NewInt(1))
		tx.SetTxCurrency(rewart.CoinType)
		txers = append(txers, tx)
	}
	return
}

//Broadcast
func (env *Work) ProcessBroadcastTransactions(mux *event.TypeMux, txs []types.CoinSelfTransaction, bc *core.BlockChain) {
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	mapcoingasUse.clearmap()
	for _, tx := range txs {
		for _,t:=range  tx.Txser{
		env.commitTransaction(t, bc, common.Address{}, nil)
		}
	}
	rewart := env.CalcRewardAndSlash(bc, nil, nil)
	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		err, _ := env.s_commitTransaction(tx, bc, common.Address{}, new(core.GasPool).AddGas(0))
		if err != nil {
			log.Error("file work", "func ProcessTransactions:::reward Tx call Error", err)
		}
	}
	env.txs,env.Receipts =types.GetCoinTXRS(env.transer,env.recpts)		
	return
}

func (env *Work) ConsensusTransactions(mux *event.TypeMux, txs []types.CoinSelfTransaction, bc *core.BlockChain, upTime map[common.Address]uint64) error {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}
	mapcoingasUse.clearmap()
	var coalescedLogs []types.CoinLogs
	tim := env.header.Time.Uint64()
	env.State.UpdateTxForBtree(uint32(tim))
	env.State.UpdateTxForBtreeBytime(uint32(tim))
	from := make([]common.Address, 0)
	for _, tx := range txs {
		// If we don't have enough gas for any further transactions then we're done
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			return errors.New("Not enough gas for further transactions")
		}
		// Start executing the transaction
		for _, t := range tx.Txser {
			env.State.Prepare(t.Hash(), common.Hash{}, env.tcount)
			err, logs := env.commitTransaction(t, bc, common.Address{}, env.gasPool)
			if err == nil {
				env.tcount++
				coalescedLogs = append(coalescedLogs,types.CoinLogs{t.GetTxCurrency(),logs})
			} else {
				return err
			}
			from = append(from, t.From())
		}
	}

	rewart := env.CalcRewardAndSlash(bc, upTime, from)
	txers := env.makeTransaction(rewart)
	for _, tx := range txers {
		err, _ := env.s_commitTransaction(tx, bc, common.Address{}, new(core.GasPool).AddGas(0))
		if err != nil {
			return err
		}
	}
	env.txs,env.Receipts =types.GetCoinTXRS(env.transer,env.recpts)
	if len(coalescedLogs) > 0 || env.tcount > 0 {
		// make a copy, the state caches the logs and these logs get "upgraded" from pending to mined
		// logs by filling in the block hash when the block was mined by the local miner. This can
		// cause a race condition if a log was "upgraded" before the PendingLogsEvent is processed.
		//cpy := make([]*types.Log, len(coalescedLogs))
		//for i, l := range coalescedLogs {
		//	cpy[i] = new(types.Log)
		//	*cpy[i] = *l
		//}
		go func(logs []types.CoinLogs, tcount int) {
			if len(logs) > 0 {
				mux.Post(core.PendingLogsEvent{Logs: logs})
			}
			if tcount > 0 {
				mux.Post(core.PendingStateEvent{})
			}
		}(coalescedLogs, env.tcount)
	}

	return nil
}
func (env *Work) GetTxs() []types.CoinSelfTransaction {
	return env.txs
}

func (env *Work) CalcRewardAndSlash(bc *core.BlockChain, upTime map[common.Address]uint64, account []common.Address) []common.RewarTx {
	bcInterval, err := manparams.NewBCIntervalByHash(env.header.ParentHash)
	if err != nil {
		log.Error("work", "获取广播周期失败", err)
		return nil
	}
	if bcInterval.IsBroadcastNumber(env.header.Number.Uint64()) {
		return nil
	}
	blkReward := blkreward.New(bc, env.State)
	rewardList := make([]common.RewarTx, 0)
	if nil != blkReward {
		//todo: read half number from state
		minersRewardMap := blkReward.CalcMinerRewards(env.header.Number.Uint64(), env.header.ParentHash)
		if 0 != len(minersRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.BlkMinerRewardAddress, To_Amont: minersRewardMap})
		}

		validatorsRewardMap := blkReward.CalcValidatorRewards(env.header.Leader, env.header.Number.Uint64())
		if 0 != len(validatorsRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.BlkValidatorRewardAddress, To_Amont: validatorsRewardMap})
		}
	}

	allGas := env.getGas()
	txsReward := txsreward.New(bc, env.State)
	if nil != txsReward {
		txsRewardMap := txsReward.CalcNodesRewards(allGas, env.header.Leader, env.header.Number.Uint64(), env.header.ParentHash)
		if 0 != len(txsRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.TxGasRewardAddress, To_Amont: txsRewardMap})
		}
	}
	lottery := lottery.New(bc, env.State, env.random)
	if nil != lottery {
		lotteryRewardMap := lottery.LotteryCalc(env.header.ParentHash, env.header.Number.Uint64())
		if 0 != len(lotteryRewardMap) {
			rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.LotteryRewardAddress, To_Amont: lotteryRewardMap})
		}
		lottery.LotterySaveAccount(account, env.header.VrfValue)
	}

	////todo 利息
	interestReward := interest.New(env.State)
	if nil == interestReward {
		return env.Reverse(rewardList)
	}
	interestCalcMap, interestPayMap := interestReward.InterestCalc(env.State, env.header.Number.Uint64())
	if 0 != len(interestPayMap) {
		rewardList = append(rewardList, common.RewarTx{CoinType: params.MAN_COIN, Fromaddr: common.InterestRewardAddress, To_Amont: interestPayMap, RewardTyp: common.RewardInerestType})
	}

	slash := slash.New(bc, env.State)
	if nil != slash {
		slash.CalcSlash(env.State, env.header.Number.Uint64(), upTime, interestCalcMap)
	}
	return env.Reverse(rewardList)
}

func (env *Work) getGas() *big.Int {

	price := mapcoingasUse.getCoinGasPrice(params.MAN_COIN)
	gas := mapcoingasUse.getCoinGasUse(params.MAN_COIN)
	allGas := new(big.Int).Mul(gas, price)
	log.INFO("奖励", "交易费奖励总额", allGas.String())
	balance := env.State.GetBalance(params.MAN_COIN, common.TxGasRewardAddress)

	if len(balance) == 0 {
		log.WARN("奖励", "交易费奖励账户余额不合法", "")
		return big.NewInt(0)
	}

	if balance[common.MainAccount].Balance.Cmp(big.NewInt(0)) <= 0 || balance[common.MainAccount].Balance.Cmp(allGas) <= 0 {
		log.WARN("奖励", "交易费奖励账户余额不合法，余额", balance)
		return big.NewInt(0)
	}
	return allGas
}
