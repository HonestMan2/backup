package rewardexec

import (
	"math/big"

	"github.com/matrix/go-matrix/reward/cfg"
	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

const (
	PackageName = "奖励"
)

type BlockReward struct {
	chain           util.ChainReader
	st              util.StateDB
	rewardCfg       *cfg.RewardCfg
	specialAccounts *mc.MatrixSpecialAccounts
	bcInterval      *manparams.BCInterval
}

func New(chain util.ChainReader, rewardCfg *cfg.RewardCfg, st util.StateDB) *BlockReward {
	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.MinerOutRate+rewardCfg.RewardMount.RewardRate.ElectedMinerRate+rewardCfg.RewardMount.RewardRate.FoundationMinerRate {
		log.ERROR(PackageName, "矿工固定区块奖励比例配置错误", "")
		return nil
	}
	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.LeaderRate+rewardCfg.RewardMount.RewardRate.ElectedValidatorsRate+rewardCfg.RewardMount.RewardRate.FoundationValidatorRate {
		log.ERROR(PackageName, "验证者固定区块奖励比例配置错误", "")
		return nil
	}

	if util.RewardFullRate != rewardCfg.RewardMount.RewardRate.OriginElectOfflineRate+rewardCfg.RewardMount.RewardRate.BackupRewardRate {
		log.ERROR(PackageName, "替补固定区块奖励比例配置错误", "")
		return nil
	}

	interval, err := matrixstate.GetDataByState(mc.MSKeyBroadcastInterval, st)
	if err != nil {
		log.ERROR(PackageName, "获取广播周期失败", err)
		return nil
	}

	data, err := matrixstate.GetDataByState(mc.MSKeyMatrixAccount, st)
	if err != nil {
		log.ERROR(PackageName, "获取特殊账户消息失败", err)
		return nil
	}

	accounts, OK := data.(*mc.MatrixSpecialAccounts)
	if OK == false || accounts == nil {
		log.ERROR(PackageName, "获取特殊账户消息失败", "结构反射失败")
		return nil
	}

	br := &BlockReward{
		chain:           chain,
		rewardCfg:       rewardCfg,
		st:              st,
		specialAccounts: accounts,
	}
	br.bcInterval, err = manparams.NewBCIntervalWithInterval(interval)
	if nil != err {
		log.ERROR(PackageName, "获取广播周期失败", "")
		return nil
	}
	return br
}
func (br *BlockReward) calcValidatorRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int) {

	leaderBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.LeaderRate)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.ElectedValidatorsRate)
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.FoundationValidatorRate)
	return leaderBlkReward, electedReward, FoundationsBlkReward
}

func (br *BlockReward) calcMinerRateMount(blockReward *big.Int) (*big.Int, *big.Int, *big.Int) {

	minerOutReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.MinerOutRate)
	electedReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.ElectedMinerRate)
	FoundationsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.RewardMount.RewardRate.FoundationMinerRate)
	return minerOutReward, electedReward, FoundationsBlkReward
}

func (br *BlockReward) CalcValidatorRewards(Leader common.Address, num uint64) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(br.rewardCfg.RewardMount.ValidatorMount), util.ManPrice)
	halfNum := br.rewardCfg.RewardMount.ValidatorHalf
	blockReward := br.calcRewardMountByNumber(RewardMan, num-1, halfNum, common.BlkValidatorRewardAddress)
	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放验证者奖励", "")
		return nil
	}

	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}

	if br.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return nil
	}

	return br.getValidatorRewards(blockReward, Leader, num)
}

func (br *BlockReward) getValidatorRewards(blockReward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	rewards := make(map[common.Address]*big.Int, 0)
	leaderBlkMount, electedMount, FoundationsMount := br.calcValidatorRateMount(blockReward)
	leaderReward := br.rewardCfg.SetReward.SetLeaderRewards(leaderBlkMount, Leader, num)
	electReward := br.rewardCfg.SetReward.GetSelectedRewards(electedMount, br.st, br.chain, common.RoleValidator|common.RoleBackupValidator, num, br.rewardCfg.RewardMount.RewardRate.BackupRewardRate)
	foundationReward := br.calcFoundationRewards(FoundationsMount, num)
	util.MergeReward(rewards, leaderReward)
	util.MergeReward(rewards, electReward)
	util.MergeReward(rewards, foundationReward)
	return rewards
}

func (br *BlockReward) getMinerRewards(blockReward *big.Int, num uint64) map[common.Address]*big.Int {
	rewards := make(map[common.Address]*big.Int, 0)

	minerOutAmount, electedMount, FoundationsMount := br.calcMinerRateMount(blockReward)
	minerOutReward := br.rewardCfg.SetReward.SetMinerOutRewards(minerOutAmount, br.st, num)
	electReward := br.rewardCfg.SetReward.GetSelectedRewards(electedMount, br.st, br.chain, common.RoleMiner|common.RoleBackupMiner, num, br.rewardCfg.RewardMount.RewardRate.BackupRewardRate)
	foundationReward := br.calcFoundationRewards(FoundationsMount, num)
	util.MergeReward(rewards, minerOutReward)
	util.MergeReward(rewards, electReward)
	util.MergeReward(rewards, foundationReward)
	return rewards
}

func (br *BlockReward) CalcMinerRewards(num uint64) map[common.Address]*big.Int {
	//广播区块不给矿工发钱
	RewardMan := new(big.Int).Mul(new(big.Int).SetUint64(br.rewardCfg.RewardMount.MinerMount), util.ManPrice)
	halfNum := br.rewardCfg.RewardMount.MinerHalf
	blockReward := br.calcRewardMountByNumber(RewardMan, num-1, halfNum, common.BlkMinerRewardAddress)
	if blockReward.Uint64() == 0 {
		log.Error(PackageName, "账户余额为0，不发放矿工奖励", "")
		return nil
	}
	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}

	if br.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return nil
	}
	return br.getMinerRewards(blockReward, num)
}
func (br *BlockReward) canCalcFoundationRewards(blockReward *big.Int, num uint64) bool {
	if br.bcInterval.IsBroadcastNumber(num) {
		return false
	}

	if blockReward.Cmp(big.NewInt(0)) <= 0 {
		log.ERROR(PackageName, "奖励金额错误", blockReward)
		return false
	}
	return true

}
func (br *BlockReward) calcFoundationRewards(blockReward *big.Int, num uint64) map[common.Address]*big.Int {

	if false == br.canCalcFoundationRewards(blockReward, num) {
		return nil
	}
	accountRewards := make(map[common.Address]*big.Int)
	accountRewards[br.specialAccounts.FoundationAccount.Address] = blockReward
	log.Info(PackageName, "基金会奖励,账户", br.specialAccounts.FoundationAccount.Address.Hex(), "金额", blockReward)
	return accountRewards
}

func (br *BlockReward) CalcNodesRewards(blockReward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int {

	if blockReward.Cmp(big.NewInt(0)) <= 0 {
		log.Error(PackageName, "账户余额非法，不发放奖励", blockReward)
		return nil
	}
	log.INFO(PackageName, "奖励金额", blockReward)

	if nil == br.rewardCfg {
		log.Error(PackageName, "奖励配置为空", "")
		return nil
	}

	if br.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return nil
	}

	rewards := make(map[common.Address]*big.Int, 0)

	validatorsBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.ValidatorsRate)
	log.INFO(PackageName, "验证者奖励总额", validatorsBlkReward)
	validatorReward := br.getValidatorRewards(validatorsBlkReward, Leader, num)
	minersBlkReward := util.CalcRateReward(blockReward, br.rewardCfg.MinersRate)
	log.INFO(PackageName, "矿工奖励总额", validatorsBlkReward)
	minerRewards := br.getMinerRewards(minersBlkReward, num)

	util.MergeReward(rewards, validatorReward)
	util.MergeReward(rewards, minerRewards)
	return rewards
}

//func (br *BlockReward) CalcRewardMountByBalance(state *state.StateDB, blockReward *big.Int, address common.Address) *big.Int {
//	//todo:后续从状态树读取对应币种减半金额,现在每个100个区块余额减半，如果减半值为0则不减半
//	halfBalance := new(big.Int).Exp(big.NewInt(10), big.NewInt(21), big.NewInt(0))
//	balance := state.GetBalance(address)
//	genesisState, _ := br.chain.StateAt(br.chain.Genesis().Root())
//	genesisBalance := genesisState.GetBalance(address)
//	log.INFO(PackageName, "计算区块奖励参数 衰减金额:", halfBalance.String(),
//		"初始账户", address.String(), "初始金额", genesisBalance[common.MainAccount].Balance.String(), "当前金额", balance[common.MainAccount].Balance.String())
//	var reward *big.Int
//	if balance[common.MainAccount].Balance.Cmp(genesisBalance[common.MainAccount].Balance) >= 0 {
//		reward = blockReward
//	}
//
//	subBalance := new(big.Int).Sub(genesisBalance[common.MainAccount].Balance, balance[common.MainAccount].Balance)
//	n := int64(0)
//	if 0 != halfBalance.Int64() {
//		n = new(big.Int).Div(subBalance, halfBalance).Int64()
//	}
//
//	if 0 == n {
//		reward = blockReward
//	} else {
//		reward = new(big.Int).Div(blockReward, new(big.Int).Exp(big.NewInt(2), big.NewInt(n), big.NewInt(0)))
//	}
//	log.INFO(PackageName, "计算区块奖励金额:", reward.String())
//	if balance[common.MainAccount].Balance.Cmp(reward) < 0 {
//		log.ERROR(PackageName, "账户余额不足，余额为", balance[common.MainAccount].Balance.String())
//		return big.NewInt(0)
//	} else {
//		return reward
//	}
//
//}

func (br *BlockReward) calcRewardMountByNumber(blockReward *big.Int, num uint64, halfNum uint64, address common.Address) *big.Int {
	//todo:后续从状态树读取对应币种减半金额,现在每个100个区块余额减半，如果减半值为0则不减半

	if blockReward.Cmp(big.NewInt(0)) < 0 {
		log.WARN(PackageName, "折半计算的奖励金额不合法", blockReward)
		return big.NewInt(0)
	}
	if nil == br.st {
		log.ERROR(PackageName, "状态树是空", "")
		return big.NewInt(0)
	}
	balance := br.st.GetBalance(address)
	if len(balance) == 0 {
		log.ERROR(PackageName, "账户余额获取不到", "")
		return nil
	}
	if balance[common.MainAccount].Balance.Cmp(big.NewInt(0)) < 0 {
		log.WARN(PackageName, "发送账户余额不合法，地址", address.Hex(), "余额", balance[common.MainAccount].Balance)
		return big.NewInt(0)
	}

	log.INFO(PackageName, "计算区块奖励参数 当前高度:", num, "半衰高度:", halfNum,
		"初始账户", address.String(), "当前金额", balance[common.MainAccount].Balance.String())
	var reward *big.Int

	n := uint64(0)
	if 0 != halfNum {
		n = num / halfNum
	}

	if 0 == n {
		reward = blockReward
	} else {
		reward = new(big.Int).Div(blockReward, new(big.Int).Exp(big.NewInt(2), new(big.Int).SetUint64(n), big.NewInt(0)))
	}
	log.INFO(PackageName, "计算区块奖励金额:", reward.String())
	if balance[common.MainAccount].Balance.Cmp(reward) < 0 {
		log.ERROR(PackageName, "账户余额不足，余额为", balance[common.MainAccount].Balance.String())
		return big.NewInt(0)
	} else {
		return reward
	}

}
