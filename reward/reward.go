package reward

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
)

//todo:多个币种reward，超级节点
// gas支持两种，自付和 代付，代付的时候转的时候所有费用到基金会账户0x800001账户，然后再由0x80001代付，委托交易
// gas分段计价 第二笔gas，0x80001垫付，写入创世配置文件，初始金额，网络组判断  ，多币种和子链需要考虑，配置超级节点上链。
type Reward interface {
	CalcNodesRewards(blockReward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int
	CalcValidatorRewards(Leader common.Address, num uint64) map[common.Address]*big.Int
	CalcMinerRewards(num uint64) map[common.Address]*big.Int
}

type Lottery interface {
	LotteryCalc(num uint64) map[string]map[common.Address]*big.Int
}

type Slash interface {
	CalcSlash(state *state.StateDB, num uint64) map[common.Address]*big.Int
}
