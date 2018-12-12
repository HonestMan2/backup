package matrixstate

import (
	"encoding/json"

	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

func (self *keyManager) initCodec() {
	self.codecMap[mc.MSKeyTopologyGraph] = new(TopologyGraphCodec)
	self.codecMap[mc.MSKeyElectGraph] = new(ElectGraphCodec)
	self.codecMap[mc.MSKeyElectOnlineState] = new(ElectOnlineStateCodec)
	self.codecMap[mc.MSKeyBroadcastInterval] = new(BroadcastIntervalCodec)
	self.codecMap[mc.MSKeyElectGenTime] = new(ElectGenTimeCodec)
	self.codecMap[mc.MSKeyMatrixAccount] = new(MatrixNodeCodec)
	self.codecMap[mc.MSKeyElectConfigInfo] = new(ElectConfigInfoCodec)
	self.codecMap[mc.MSKeyVIPConfig] = new(MSPVIPConfigCodec)
	self.codecMap[mc.MSKeyPreBroadcastRoot] = new(MSPreBroadcastStateDBCodec)
	self.codecMap[mc.MSKeyMinHash] = new(MSKeyMinHashCodec)
	self.codecMap[mc.MSKeyPerAllTop] = new(MSKeyPerAllTopCodec)
	self.codecMap[mc.MSKeyPreMiner] = new(MSKeyPreMinerCodec)
	self.codecMap[mc.MSKeyLeaderConfig] = new(MSKeyLeaderConfigCodec)
	self.codecMap[mc.MSKeyBlkRewardCfg] = new(MSPRewardRateCfgCodec)
	self.codecMap[mc.MSKeyTxsRewardCfg] = new(MSPTxsRewardCfgCodec)
	self.codecMap[mc.MSKeyInterestCfg] = new(MSPInterestCfgCodec)
	self.codecMap[mc.MSKeyLotteryCfg] = new(MSPLotteryCfgCodec)
	self.codecMap[mc.MSKeySlashCfg] = new(MSPSlashCfgCodec)
	self.codecMap[mc.MSKeyMultiCoin] = new(MSPRewardRateCfgCodec)
	self.codecMap[mc.MSKeySuperBlockCfg] = new(MSPSuperBlkCfgCodec)
}

type codec interface {
	encodeFn(msg interface{}) ([]byte, error)
	decodeFn(data []byte) (interface{}, error)
}

////////////////////////////////////////////////////////////////////////
// key = MSPBroadcastInterval
type BroadcastIntervalCodec struct {
}

func (BroadcastIntervalCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("broadcast interval json.Marshal failed: %s", err)
	}
	return data, nil
}

func (BroadcastIntervalCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.BCIntervalInfo)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("broadcast interval json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("broadcast interval msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPTopologyGraph
type TopologyGraphCodec struct {
}

func (TopologyGraphCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (TopologyGraphCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.TopologyGraph)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPElectGraph
type ElectGraphCodec struct {
}

func (ElectGraphCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (ElectGraphCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.ElectGraph)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPElectOnlineState
type ElectOnlineStateCodec struct {
}

func (ElectOnlineStateCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (ElectOnlineStateCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.ElectOnlineStatus)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type ElectGenTimeCodec struct {
}

func (ElectGenTimeCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (ElectGenTimeCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.ElectGenTimeStruct)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPMatrixNode
type MatrixNodeCodec struct {
}

func (MatrixNodeCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MatrixNodeCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.MatrixSpecialAccounts)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = ElectConfigInfoCodec
type ElectConfigInfoCodec struct {
}

func (ElectConfigInfoCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (ElectConfigInfoCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.ElectConfigInfo)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPVIPConfigCodec
type MSPVIPConfigCodec struct {
}

func (MSPVIPConfigCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPVIPConfigCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new([]mc.VIPConfig)
	//msg:=[]mc.VIPConfig{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPreBroadcastStateDBCodec
type MSPreBroadcastStateDBCodec struct {
}

func (MSPreBroadcastStateDBCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPreBroadcastStateDBCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.PreBroadStateRoot)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSKeyLeaderConfig
type MSKeyLeaderConfigCodec struct {
}

func (MSKeyLeaderConfigCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSKeyLeaderConfigCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.LeaderConfig)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

////////////////////////////////////////////////////////////////////////
// key = MSPRewardRateCfgCodec
type MSPRewardRateCfgCodec struct {
}

func (MSPRewardRateCfgCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPRewardRateCfgCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.BlkRewardCfg)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSPTxsRewardCfgCodec struct {
}

func (MSPTxsRewardCfgCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPTxsRewardCfgCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.TxsRewardCfgStruct)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSPLotteryInfoCodec struct {
	PrizeLevel uint8  //奖励级别
	PrizeNum   uint64 //奖励名额
	PrizeMoney uint64 //奖励金额 单位man
}

func (MSPLotteryInfoCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPLotteryInfoCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.LotteryInfo)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSPLotteryCfgCodec struct {
}

func (MSPLotteryCfgCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPLotteryCfgCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.LotteryCfgStruct)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSPInterestCfgCodec struct {
}

func (MSPInterestCfgCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPInterestCfgCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.InterestCfgStruct)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSKeyMinHashCodec struct {
}

func (MSKeyMinHashCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSKeyMinHashCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.MinHashStruct)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSKeyPerAllTopCodec struct {
}

func (MSKeyPerAllTopCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSKeyPerAllTopCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.PreAllTopStruct)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSPSlashCfgCodec struct {
}

func (MSPSlashCfgCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPSlashCfgCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.SlashCfgStruct)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSPSuperBlkCfgCodec struct {
}

func (MSPSuperBlkCfgCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSPSuperBlkCfgCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.SuperBlkCfg)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}

type MSKeyPreMinerCodec struct {
}

func (MSKeyPreMinerCodec) encodeFn(msg interface{}) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("json.Marshal failed: %s", err)
	}
	return data, nil
}

func (MSKeyPreMinerCodec) decodeFn(data []byte) (interface{}, error) {
	msg := new(mc.PreMinerStruct)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	return msg, nil
}
