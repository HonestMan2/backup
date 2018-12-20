package util

import (
	"math/big"
	"sort"

	"github.com/matrix/go-matrix/core/matrixstate"

	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/log"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/common"
)

const (
	PackageName = "奖励util"
)
const (
	RewardFullRate = uint64(10000)
	Stop           = "0"
)

var (
	//ValidatorBlockReward  *big.Int = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), big.NewInt(0)) // Block reward in wei for successfully mining a block
	MultilCoinBlockReward *big.Int = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), big.NewInt(0)) // Block reward in wei for successfully mining a block upward from Byzantium
	//分母10000
	ByzantiumTxsRewardDen *big.Int = big.NewInt(1000000000) // Block reward in wei for successfully mining a block upward from Byzantium
	ValidatorsBlockReward *big.Int = big.NewInt(5e+18)
	MinersBlockReward     *big.Int = big.NewInt(5e+18)

	ManPrice *big.Int = big.NewInt(1e18)
)

type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	GetBlockByNumber(number uint64) *types.Block

	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
	State() (*state.StateDB, error)
	GetMatrixStateData(key string) (interface{}, error)
	GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error)
	GetSuperBlockNum() (uint64, error)
	GetGraphByState(state matrixstate.StateDB) (*mc.TopologyGraph, *mc.ElectGraph, error)
}

type StateDB interface {
	GetBalance(typ string,addr common.Address) common.BalanceType
	GetMatrixData(hash common.Hash) (val []byte)
	SetMatrixData(hash common.Hash, val []byte)
}

func SetAccountRewards(rewards map[common.Address]*big.Int, account common.Address, reward *big.Int) {

	if 0 == reward.Cmp(big.NewInt(0)) {
		return
	}
	if nil == rewards {
		return
	}
	if account.Equal(common.Address{}) {
		log.ERROR(PackageName, "奖励的地址非法", account.Hex())
		return
	}
	if _, ok := rewards[account]; ok {
		rewards[account] = rewards[account].Add(rewards[account], reward)
	} else {
		rewards[account] = reward
	}
}

func CalcRateReward(rewardAmount *big.Int, rate uint64) *big.Int {
	temp := new(big.Int).Mul(rewardAmount, new(big.Int).SetUint64(rate))
	return new(big.Int).Div(temp, new(big.Int).SetUint64(RewardFullRate))
}

func CalcDepositRate(reward *big.Int, depositNodes map[common.Address]*big.Int) map[common.Address]*big.Int {

	if 0 == len(depositNodes) {
		log.ERROR(PackageName, "抵押列表为空", "")
		return nil
	}
	totalDeposit := new(big.Int)

	depositNodesFix := make(map[common.Address]*big.Int)

	for k, v := range depositNodes {
		depositTemp := new(big.Int).Div(v, big.NewInt(1e18))
		if depositTemp.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "定点化的抵押值错误", depositTemp)
			return nil
		}
		depositNodesFix[k] = depositTemp
		totalDeposit.Add(totalDeposit, depositTemp)
	}
	if totalDeposit.Cmp(big.NewInt(0)) <= 0 {
		log.ERROR(PackageName, "定点化抵押值为非法", totalDeposit)
		return nil
	}
	log.INFO(PackageName, "计算抵押总额,账户总抵押", totalDeposit, "定点化抵押", totalDeposit)

	rewardFixed := new(big.Int).Div(reward, big.NewInt(1e8))

	if 0 == rewardFixed.Cmp(big.NewInt(0)) {
		log.ERROR(PackageName, "定点化奖励金额为0", "")
		return nil
	}
	sortedKeys := make([]string, 0)

	for k := range depositNodesFix {
		sortedKeys = append(sortedKeys, k.String())
	}
	sort.Strings(sortedKeys)
	rewards := make(map[common.Address]*big.Int)
	for _, k := range sortedKeys {
		rateTemp := new(big.Int).Mul(depositNodesFix[common.HexToAddress(k)], big.NewInt(1e10))
		rate := new(big.Int).Div(rateTemp, totalDeposit)
		if rate.Cmp(big.NewInt(0)) < 0 {
			log.ERROR(PackageName, "定点化比例非法", rate)
			continue
		}
		log.INFO(PackageName, "计算比例,账户", k, "定点化比例", rate)

		rewardTemp := new(big.Int).Mul(rewardFixed, rate)
		rewardTemp1 := new(big.Int).Div(rewardTemp, big.NewInt(1e10))
		oneNodeReward := new(big.Int).Mul(rewardTemp1, big.NewInt(1e8))
		rewards[common.HexToAddress(k)] = oneNodeReward
		log.INFO(PackageName, "计算奖励金额,账户", k, "定点化金额", rewards[common.HexToAddress(k)])
		log.INFO(PackageName, "", "")
	}
	return rewards
}

func MergeReward(dst map[common.Address]*big.Int, src map[common.Address]*big.Int) {
	if 0 == len(src) {
		return
	}
	if nil == dst {
		log.ERROR(PackageName, "dst is nil", dst)
		return
	}
	for account, reward := range src {

		SetAccountRewards(dst, account, reward)
	}

}
