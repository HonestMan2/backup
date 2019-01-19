package core

import (
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"sync"
)

type PreStateReadFn func(key string) (interface{}, error)
type ProduceMatrixStateDataFn func(block *types.Block, readFn PreStateReadFn) (interface{}, error)

type MatrixProcessor struct {
	mu          sync.RWMutex
	producerMap map[string]ProduceMatrixStateDataFn
}

func NewMatrixProcessor() *MatrixProcessor {
	return &MatrixProcessor{
		producerMap: make(map[string]ProduceMatrixStateDataFn),
	}
}

func (mp *MatrixProcessor) RegisterProducer(key string, producer ProduceMatrixStateDataFn) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if _, exist := mp.producerMap[key]; exist {
		log.Warn("MatrixProcessor", "已存在的key重复注册Producer", key)
	}
	mp.producerMap[key] = producer
}

func (mp *MatrixProcessor) ProcessStateVersion(version []byte, state *state.StateDBManage) error {
	if len(version) == 0 || state == nil {
		return errors.New("param is nil")
	}

	curVersion := matrixstate.GetVersionInfo(state)
	newVersion := string(version)
	if curVersion != newVersion {
		log.Info("MatrixProcessor", "版本号更新", "开始", "旧版本", curVersion, "新版本", newVersion)
		curVersion = newVersion
		if err := matrixstate.SetVersionInfo(state, curVersion); err != nil {
			log.Error("MatrixProcessor", "版本号更新失败", err)
			return err
		}
	}
	return nil
}

func (mp *MatrixProcessor) ProcessMatrixState(block *types.Block, state *state.StateDBManage) error {
	if block == nil || state == nil {
		return errors.New("param is nil")
	}

	// 获取matrix状态树管理类
	version := matrixstate.GetVersionInfo(state)
	mgr := matrixstate.GetManager(version)
	if mgr == nil {
		return matrixstate.ErrFindManager
	}

	readFn := func(key string) (interface{}, error) {
		if key == mc.MSKeyVersionInfo {
			return version, nil
		}
		opt, err := mgr.FindOperator(key)
		if err != nil {
			return nil, err
		}
		return opt.GetValue(state)
	}

	mp.mu.RLock()
	defer mp.mu.RUnlock()

	dataMap := make(map[string]interface{})
	for key := range mp.producerMap {
		data, err := mp.producerMap[key](block, readFn)
		if err != nil {
			return errors.Errorf("key(%s) produce matrix state data err(%v)", key, err)
		}
		if nil == data {
			continue
		}

		dataMap[key] = data
	}

	for key := range dataMap {
		opt, err := mgr.FindOperator(key)
		if err != nil {
			return errors.Errorf("key(%s) find operator err: %v", key, err)
		}
		if err := opt.SetValue(state, dataMap[key]); err != nil {
			return errors.Errorf("key(%s) set value err: %v", key, err)
		}
	}

	return nil
}
