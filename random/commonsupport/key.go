// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package commonsupport

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/common/readstatedb"
	"fmt"
)

type VoteData struct {
	PrivateData []byte
	PublicData []byte
}
func GetCommonMap(private map[common.Address][]byte,public map[common.Address][]byte)map[common.Address]VoteData{
	commonMap:=make(map[common.Address]VoteData)
	for address,privateData:=range private{
		publicData, ok := public[address]
		if ok==false{
			continue
		}
		commonMap[address]=VoteData{
			PrivateData:privateData,
			PublicData:publicData,
		}
	}
	return commonMap
}

func GetValidPrivateSum(commonMap map[common.Address]VoteData)*big.Int{
	PrivateSum:=big.NewInt(0)
	for _,v:=range commonMap{
		if CheckVoteDataIsCompare(v.PrivateData,v.PublicData){
			PrivateSum.Add(PrivateSum, common.BytesToHash(v.PrivateData).Big())
		}
	}
	return PrivateSum
}

func CheckVoteDataIsCompare(private []byte, public []byte) bool {
	curve := btcec.S256()
	pk1, err := btcec.ParsePubKey(public, curve)
	if err != nil {
		log.Warn(ModeleRandomCommon,"比对公私钥数据阶段 转换公钥失败 err",err)
		return false
	}
	if pk1==nil{
		log.Warn(ModeleRandomCommon,"比对公私钥数据阶段 转换的公钥为空 err",err)
		return false
	}

	pk1_1 := (*ecdsa.PublicKey)(pk1)
	if pk1_1==nil{
		log.Warn(ModeleRandomCommon,"比对公私钥数据阶段 公钥强转为*ecdsa.PublicKey失败","转换的为空")
		return false
	}

	xx, yy := pk1_1.Curve.ScalarBaseMult(private)
	if xx.Cmp(pk1_1.X) != 0 {
		log.Warn(ModeleRandomCommon,"对比公私钥数据阶段,X值不匹配","")
		return false
	}
	if yy.Cmp(pk1_1.Y) != 0 {
		log.Warn(ModeleRandomCommon,"对比公私要数据阶段,Y值不匹配","")
		return false
	}
	return true
}

func GetVoteData() (*big.Int, []byte, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Error(ModeleRandomCommon,"生成投票数据失败 err",err)
		return nil, nil, err
	}
	if key.D==nil{
		log.Error(ModeleRandomCommon,"生成投票数据失败","D为空")
	}
	return key.D, keystore.ECDSAPKCompression(&key.PublicKey), err
}

func GetNumberByHash(hash common.Hash, support baseinterface.RandomChainSupport) (uint64, error) {
	tHeader := support.BlockChain().GetHeaderByHash(hash)
	if tHeader == nil {
		log.Error(ModeleRandomCommon, "根据hash算header失败 hash", hash.String())
		return 0, errors.New("根据hash算header失败")
	}
	if tHeader.Number == nil {
		log.Error(ModeleRandomCommon, "header内的高度获取失败", hash.String())
		return 0, errors.New("header 内的高度获取失败")
	}
	return tHeader.Number.Uint64(), nil
}

func GetAncestorHash(hash common.Hash, height uint64, support baseinterface.RandomChainSupport) (common.Hash, error) {
	aimHash, err := support.BlockChain().GetAncestorHash(hash, height)
	if err != nil {
		log.Error(ModeleRandomCommon, "获取祖先hash失败 hash", hash.String(), "height", height, "err", err)
		return common.Hash{}, err
	}
	return aimHash, nil
}

func getKeyTransInfo(root  []common.CoinRoot,types string, support baseinterface.RandomChainSupport) map[common.Address][]byte {
	ans, err := core.GetBroadcastTxMap(support,root, types)
	if err != nil {
		log.Error(ModeleRandomCommon, "获取特殊交易失败 root", root, "types", types)
	}
	return ans
}

func GetValidVoteSum(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {
	height, err := GetNumberByHash(hash, support)
	if err != nil {
		log.Error(ModeleRandomCommon, "计算种子失败 err", err, "hash", hash.String())
		return nil, errors.New("计算hash高度失败")
	}

	preBroadcastRoot,err:=readstatedb.GetPreBroadcastRoot(support.BlockChain(),height)
	if err!=nil{
		log.Error(ModeleRandomCommon,"计算种子阶段,获取前2个广播区块的root值失败 err",err)
		return nil,fmt.Errorf("从状态树获取前2个广播区块root失败")
	}


	PrivateMap := getKeyTransInfo(preBroadcastRoot.LastStateRoot,mc.Privatekey, support)
	PublicMap := getKeyTransInfo(preBroadcastRoot.BeforeLastStateRoot ,mc.Publickey, support)

	commonMap:=GetCommonMap(PrivateMap,PublicMap)
	MapAns:=GetValidPrivateSum(commonMap)
	return MapAns, nil
}

