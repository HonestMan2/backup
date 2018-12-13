// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package ca

import (
	"math/big"
	"sync"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

type TopologyGraphReader interface {
	GetCurrentHash() common.Hash
	GetHashByNumber(number uint64) common.Hash
	GetTopologyGraphByHash(blockHash common.Hash) (*mc.TopologyGraph, error)
	GetOriginalElectByHash(blockHash common.Hash) ([]common.Elect, error)
	GetNextElectByHash(blockHash common.Hash) ([]common.Elect, error)
	GetSpecialAccounts(blockHash common.Hash) (*mc.MatrixSpecialAccounts, error)
}

// Identity stand for node's identity.
type Identity struct {
	// self nodeId
	self discover.NodeID
	addr common.Address

	// if in elected duration
	duration      bool
	currentHeight *big.Int
	hash          common.Hash

	trChan          chan TopologyGraphReader
	topologyReader  TopologyGraphReader
	topology        *mc.TopologyGraph
	specialAccounts *mc.MatrixSpecialAccounts
	prevElect       []common.Elect
	currentNodes    []discover.NodeID
	frontNodes      []discover.NodeID

	// self previous, current and next role type
	currentRole common.RoleType

	// chan to listen block coming and quit message
	blockChan chan *types.Block
	quit      chan struct{}

	// lock and once to sync
	lock sync.RWMutex
	once sync.Once

	// sub to unsubscribe block channel
	sub event.Subscription

	// logger
	log log.Logger

	// deposit in current height
	deposit []vm.DepositDetail

	gapValidator      []common.Address
	gapValidatorCache []common.Address

	// addrByGroup
	addrByGroup map[common.RoleType][]common.Address
}

var ide = newIde()

func newIde() *Identity {
	return &Identity{
		quit:        make(chan struct{}),
		currentRole: common.RoleNil,
		duration:    false,
		trChan:      make(chan TopologyGraphReader, 1),
		topology:    new(mc.TopologyGraph),
		prevElect:   make([]common.Elect, 0),
	}
}

// init to do something before run.
func (ide *Identity) init(id discover.NodeID, path string) {
	ide.once.Do(func() {
		// check bootNode and set identity
		ide.self = id
		ide.log = log.New()
	})
}

// Run this Identity.
func Start(id discover.NodeID, path string) {
	ide.init(id, path)

	defer func() {
		ide.sub.Unsubscribe()

		close(ide.quit)
		close(ide.blockChan)
	}()

	select {
	case tr := <-ide.trChan:
		ide.topologyReader = tr
	case <-ide.quit:
		return
	}

	ide.blockChan = make(chan *types.Block)
	ide.sub, _ = mc.SubscribeEvent(mc.NewBlockMessage, ide.blockChan)
	log.INFO("CA", "订阅区块事件", "完成")
	mc.PublishEvent(mc.CA_ReqCurrentBlock, struct{}{})

	for {
		select {
		case block := <-ide.blockChan:
			header := block.Header()
			hash := block.Hash()
			ide.currentHeight = header.Number
			ide.hash = block.Hash()

			log.INFO("CA", "leader", header.Leader, "height", header.Number.Uint64(), "block hash", hash)

			// init current height deposit
			ide.deposit, _ = GetElectedByHeightWithdraw(header.Number)
			// get self address from deposit
			ide.addr = GetAddress()

			// get broadcast interval
			bcInterval, err := manparams.GetBCIntervalInfoByHash(hash)
			if err != nil {
				ide.log.Error("get broadcast interval", "error", err)
				continue
			}
			log.INFO("CA", "broadcast interval", bcInterval)

			// do topology
			tg, err := ide.topologyReader.GetTopologyGraphByHash(hash)
			if err != nil {
				ide.log.Error("get topology", "error", err)
				continue
			}
			ide.topology = tg

			// get special accounts
			accounts, err := ide.topologyReader.GetSpecialAccounts(hash)
			if err != nil {
				log.Error("ca", "get special accounts err", err)
				return
			}
			ide.specialAccounts = accounts
			log.INFO("CA", "special Accounts", bcInterval)

			// get elect
			elect, err := ide.topologyReader.GetNextElectByHash(hash)
			if err != nil {
				ide.log.Error("get next elect", "error", err)
				continue
			}
			ide.prevElect = elect

			// init topology
			initCurrentTopology()
			initNowTopologyResult()

			// get nodes in buckets
			nodesInBuckets := getNodesInBuckets(header.Number)

			// send role message to elect
			mc.PublishEvent(mc.CA_RoleUpdated, &mc.RoleUpdatedMsg{Role: ide.currentRole, BlockNum: header.Number.Uint64(), BlockHash: hash, Leader: header.Leader, IsSuperBlock: header.IsSuperHeader()})
			log.Info("ca publish identity", "data", mc.RoleUpdatedMsg{Role: ide.currentRole, BlockNum: header.Number.Uint64(), Leader: header.Leader})
			// get nodes in buckets and send to buckets
			mc.PublishEvent(mc.BlockToBuckets, mc.BlockToBucket{Ms: nodesInBuckets, Height: block.Header().Number, Role: ide.currentRole})
			// send identity to linker
			mc.PublishEvent(mc.BlockToLinkers, mc.BlockToLinker{Height: header.Number, BroadCastInterval: bcInterval, Role: ide.currentRole})
			mc.PublishEvent(mc.SendSyncRole, mc.SyncIdEvent{Role: ide.currentRole}) //lb
			mc.PublishEvent(mc.TxPoolManager, ide.currentRole)
		case <-ide.quit:
			return
		}
	}
}

// Stop this Identity.
func Stop() {
	ide.log.Info("identity stop")

	ide.lock.Lock()
	ide.quit <- struct{}{}
	ide.lock.Unlock()
}

// InitCurrentTopology init current topology.
func initCurrentTopology() {
	ide.lock.Lock()
	// change default role
	ide.currentRole = common.RoleDefault

	for _, t := range ide.topology.NodeList {
		if t.Account == ide.addr {
			log.INFO("initCurrentTopology", "accont", t.Account.String(), "type", t.Type)
			ide.currentRole = t.Type
			break
		}
	}

	if ide.specialAccounts.BroadcastAccount.NodeID == ide.self {
		ide.currentRole = common.RoleBroadcast
	}

	for _, im := range ide.specialAccounts.InnerMinerAccounts {
		if im.NodeID == ide.self {
			ide.currentRole = common.RoleInnerMiner
			break
		}
	}
	ide.lock.Unlock()
	log.Info("current topology", "info:", ide.topology)
}

// initNowTopologyResult
func initNowTopologyResult() {
	ide.lock.Lock()
	ide.addrByGroup = make(map[common.RoleType][]common.Address)
	for _, node := range ide.topology.NodeList {
		ide.addrByGroup[node.Type] = append(ide.addrByGroup[node.Type], node.Account)
	}

	ide.addrByGroup[common.RoleBroadcast] = append(ide.addrByGroup[common.RoleBroadcast], ide.specialAccounts.BroadcastAccount.Address)

	for _, im := range ide.specialAccounts.InnerMinerAccounts {
		ide.addrByGroup[common.RoleInnerMiner] = append(ide.addrByGroup[common.RoleInnerMiner], im.Address)
	}
	ide.lock.Unlock()
}

// SetTopologyReader
func SetTopologyReader(topologyReader TopologyGraphReader) {
	ide.trChan <- topologyReader
}

// GetRolesByGroup
func GetRolesByGroup(roleType common.RoleType) (result []discover.NodeID) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	for k, v := range ide.addrByGroup {
		if (k & roleType) != 0 {
			for _, addr := range v {
				id, err := ConvertAddressToNodeId(addr)
				if err != nil {
					ide.log.Error("convert error", "ca", err)
					continue
				}
				result = append(result, id)
			}
		}
	}
	return
}

// GetRolesByGroupWithBackup
func GetRolesByGroupWithNextElect(roleType common.RoleType) (result []discover.NodeID) {
	result = GetRolesByGroup(roleType)
	for _, elect := range ide.prevElect {
		temp := true
		role := elect.Type.Transfer2CommonRole()
		if (role & roleType) != 0 {
			id, err := ConvertAddressToNodeId(elect.Account)
			if err != nil {
				ide.log.Error("convert error", "ca", err)
				continue
			}
			for _, r := range result {
				if r == id {
					temp = false
				}
			}
			if temp {
				result = append(result, id)
			}
		}
	}
	return
}

// GetRolesByGroupOnlyBackup
func GetRolesByGroupOnlyNextElect(roleType common.RoleType) (result []discover.NodeID) {
	for _, elect := range ide.prevElect {
		role := elect.Type.Transfer2CommonRole()
		if (role & roleType) != 0 {
			id, err := ConvertAddressToNodeId(elect.Account)
			if err != nil {
				ide.log.Error("convert error", "ca", err)
				continue
			}
			result = append(result, id)
		}
	}
	return
}

// Get self identity.
func GetRole() (role common.RoleType) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	return ide.currentRole
}

func GetHeight() *big.Int {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	return ide.currentHeight
}

// InDuration
func InDuration() bool {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	return ide.duration
}

// GetElectedByHeightAndRole get elected node, miner or validator by block height and type.
func GetElectedByHeightAndRole(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
	return depoistInfo.GetDepositList(height, roleType)
}

// GetElectedByHeight get all elected node by height.
func GetElectedByHeight(height *big.Int) ([]vm.DepositDetail, error) {
	return depoistInfo.GetAllDeposit(height)
}

// GetElectedByHeightWithdraw get all info in deposit.
func GetElectedByHeightWithdraw(height *big.Int) ([]vm.DepositDetail, error) {
	return depoistInfo.GetDepositAndWithDrawList(height)
}

// GetNodeNumber
func GetNodeNumber() (uint32, error) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	for _, n := range ide.topology.NodeList {
		if n.Account == ide.addr {
			return uint32(n.NodeNumber), nil
		}
	}
	return 0, errors.New("No current node number. ")
}

// GetGapValidator
func GetGapValidator() (rlt []discover.NodeID) {
	ori, err := ide.topologyReader.GetOriginalElectByHash(ide.hash)
	if err != nil {
		ide.log.Error("ca", "GetOriginalElect, error:", err)
		return
	}

	for _, or := range ori {
		if or.Type >= common.ElectRoleValidator {
			id, err := ConvertAddressToNodeId(or.Account)
			if err != nil {
				ide.log.Error("ca", "GetGapValidator, error:", err)
				continue
			}
			rlt = append(rlt, id)
		}
	}
	return
}

// getNodesInBuckets get miner nodes that should be in buckets.
func getNodesInBuckets(height *big.Int) (result []discover.NodeID) {
	electedMiners, _ := GetElectedByHeightAndRole(height, common.RoleMiner)

	msMap := make(map[common.Address]discover.NodeID)
	for _, m := range electedMiners {
		msMap[m.Address] = m.NodeID
	}
	for _, node := range ide.topology.NodeList {
		for key := range msMap {
			if key == node.Account {
				delete(msMap, key)
				break
			}
		}
	}
	for key, val := range msMap {
		if ide.addr == key {
			ide.currentRole = common.RoleBucket
		}
		result = append(result, val)
	}
	return
}

// GetTopologyInLinker
func GetTopologyInLinker() (result map[common.RoleType][]discover.NodeID) {
	ide.lock.RLock()
	defer ide.lock.RUnlock()

	ide.frontNodes = make([]discover.NodeID, 0)
	ide.frontNodes = ide.currentNodes
	ide.currentNodes = make([]discover.NodeID, 0)

	result = make(map[common.RoleType][]discover.NodeID)
	ide.lock.RLock()
	for k, v := range ide.addrByGroup {
		for _, addr := range v {
			id, err := ConvertAddressToNodeId(addr)
			if err != nil {
				ide.log.Error("convert error", "ca", err)
				continue
			}
			ide.currentNodes = append(ide.currentNodes, id)
			result[k] = append(result[k], id)
		}
	}
	ide.lock.RUnlock()
	for _, elect := range ide.prevElect {
		id, err := ConvertAddressToNodeId(elect.Account)
		if err != nil {
			ide.log.Error("convert error", "ca", err)
			continue
		}

		temp := true
		role := elect.Type.Transfer2CommonRole()
		for _, i := range result[role] {
			if i == id {
				temp = false
			}
		}
		if temp {
			ide.currentNodes = append(ide.currentNodes, id)
			result[role] = append(result[role], id)
		}
	}
	return
}

// GetDropNode
func GetDropNode() (result []discover.NodeID) {
	for _, fn := range ide.frontNodes {
		temp := false
		for _, cn := range ide.currentNodes {
			if cn == fn {
				temp = true
				break
			}
		}
		if !temp {
			result = append(result, fn)
		}
	}
	return
}

// GetFrontNodes
func GetFrontNodes() []discover.NodeID {
	ide.lock.RLock()
	defer ide.lock.RUnlock()
	return ide.frontNodes
}

// GetAddress
func GetAddress() common.Address {
	addr, err := ConvertNodeIdToAddress(ide.self)
	if err != nil {
		log.Error("ca get self address", "error", err)
	}
	return addr
}

// GetSelfLevel
func GetSelfLevel() int {
	switch {
	case ide.currentRole > common.RoleBucket:
		return TopNode
	case ide.currentRole == common.RoleBucket:
		m := big.Int{}
		return int(m.Mod(ide.addr.Hash().Big(), big.NewInt(4)).Int64()) + 1
	case ide.currentRole <= common.RoleDefault:
		return DefaultNode
	default:
		return ErrNode
	}
}

// GetTopologyByNumber
func GetTopologyByNumber(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
	hash := ide.topologyReader.GetHashByNumber(number)
	if (hash == common.Hash{}) {
		return nil, errors.Errorf("get hash by number(%d) err!", number)
	}
	return GetTopologyByHash(reqTypes, hash)
}

func GetTopologyByHash(reqTypes common.RoleType, hash common.Hash) (*mc.TopologyGraph, error) {
	tg, err := ide.topologyReader.GetTopologyGraphByHash(hash)
	if err != nil {
		log.Error("GetAccountTopologyInfo", "error", err, "hash", hash.TerminalString())
		return nil, err
	}

	rlt := &mc.TopologyGraph{
		CurNodeNumber: tg.CurNodeNumber,
	}
	for _, node := range tg.NodeList {
		if node.Type&reqTypes != 0 {
			rlt.NodeList = append(rlt.NodeList, node)
		}
	}

	return rlt, nil
}

// GetAccountTopologyInfo
func GetAccountTopologyInfo(account common.Address, number uint64) (*mc.TopologyNodeInfo, error) {
	hash := ide.topologyReader.GetHashByNumber(number)
	if (hash == common.Hash{}) {
		return nil, errors.Errorf("get hash by number(%d) err!", number)
	}

	tg, err := ide.topologyReader.GetTopologyGraphByHash(hash)
	if err != nil {
		ide.log.Error("GetAccountTopologyInfo", "error", err)
		return nil, err
	}
	for _, node := range tg.NodeList {
		if node.Account == account {
			return &node, nil
		}
	}
	return nil, errors.New("not found")
}

// GetAccountOriginalRole
func GetAccountOriginalRole(account common.Address, hash common.Hash) (common.RoleType, error) {
	accounts, err := ide.topologyReader.GetSpecialAccounts(hash)
	if err == nil {
		if accounts.BroadcastAccount.Address == account {
			return common.RoleBroadcast, nil
		}

		for _, im := range accounts.InnerMinerAccounts {
			if im.Address == account {
				return common.RoleInnerMiner, nil
			}
		}
	}

	ori, err := ide.topologyReader.GetOriginalElectByHash(hash)
	if err != nil {
		ide.log.Error("get original elect", "error", err)
		return common.RoleNil, err
	}

	for _, elect := range ori {
		if elect.Account == account {
			return elect.Type.Transfer2CommonRole(), nil
		}
	}
	return common.RoleNil, errors.New("not found")
}

// ConvertNodeIdToAddress
func ConvertNodeIdToAddress(id discover.NodeID) (addr common.Address, err error) {
	for _, node := range ide.deposit {
		if node.NodeID == id {
			return node.Address, nil
		}
	}
	hash := ide.topologyReader.GetCurrentHash()
	if (hash == common.Hash{}) {
		return common.Address{0}, errors.New("get current hash err")
	}

	account, err := ide.topologyReader.GetSpecialAccounts(hash)
	if nil != err {
		return common.Address{0}, errors.New("get special account err")
	}

	ide.specialAccounts = account

	if ide.specialAccounts.BroadcastAccount.NodeID == id {
		return ide.specialAccounts.BroadcastAccount.Address, nil
	}

	for _, im := range ide.specialAccounts.InnerMinerAccounts {
		if im.NodeID == id {
			return im.Address, nil
		}
	}

	return addr, errors.New("not found")
}

// ConvertAddressToNodeId
func ConvertAddressToNodeId(address common.Address) (id discover.NodeID, err error) {
	for _, node := range ide.deposit {
		if node.Address == address {
			return node.NodeID, nil
		}
	}

	hash := ide.topologyReader.GetCurrentHash()
	if (hash == common.Hash{}) {
		return discover.NodeID{0}, errors.New("get current hash err")
	}

	account, err := ide.topologyReader.GetSpecialAccounts(hash)
	if nil != err {
		return discover.NodeID{0}, errors.New("get special account err")
	}

	ide.specialAccounts = account
	if ide.specialAccounts.BroadcastAccount.Address == address {
		return ide.specialAccounts.BroadcastAccount.NodeID, nil
	}

	for _, im := range ide.specialAccounts.InnerMinerAccounts {
		if im.Address == address {
			return im.NodeID, nil
		}
	}

	return id, errors.New("not found")
}
