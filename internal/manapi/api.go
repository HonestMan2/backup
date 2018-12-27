// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package manapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"encoding/base64"
	"encoding/json"
	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/base58"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/hexutil"
	"github.com/matrix/go-matrix/common/math"
	"github.com/matrix/go-matrix/consensus/manash"
	"github.com/matrix/go-matrix/console"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/rawdb"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/crc8"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/crypto/aes"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/rlp"
	"github.com/matrix/go-matrix/rpc"
	"io/ioutil"
	"os"
)

const (
	defaultGasPrice = 50 * params.Shannon
)

// PublicMatrixAPI provides an API to access Matrix related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicMatrixAPI struct {
	b Backend
}

// NewPublicMatrixAPI creates a new Matrix protocol API.
func NewPublicMatrixAPI(b Backend) *PublicMatrixAPI {
	return &PublicMatrixAPI{b}
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicMatrixAPI) GasPrice(ctx context.Context) (*big.Int, error) {
	return s.b.SuggestPrice(ctx)
}

// ProtocolVersion returns the current Matrix protocol version this node supports
func (s *PublicMatrixAPI) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(s.b.ProtocolVersion())
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicMatrixAPI) Syncing() (interface{}, error) {
	progress := s.b.Downloader().Progress()

	// Return not syncing if the synchronisation already completed
	if progress.CurrentBlock >= progress.HighestBlock {
		return false, nil
	}
	// Otherwise gather the block sync stats
	return map[string]interface{}{
		"startingBlock": hexutil.Uint64(progress.StartingBlock),
		"currentBlock":  hexutil.Uint64(progress.CurrentBlock),
		"highestBlock":  hexutil.Uint64(progress.HighestBlock),
		"pulledStates":  hexutil.Uint64(progress.PulledStates),
		"knownStates":   hexutil.Uint64(progress.KnownStates),
	}, nil
}

// PublicTxPoolAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PublicTxPoolAPI struct {
	b Backend
}

// NewPublicTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicTxPoolAPI(b Backend) *PublicTxPoolAPI {
	return &PublicTxPoolAPI{b}
}

// Content returns the transactions contained within the transaction pool.
func (s *PublicTxPoolAPI) Content() map[string]map[string]map[string]*RPCTransaction {
	content := map[string]map[string]map[string]*RPCTransaction{
		"pending": make(map[string]map[string]*RPCTransaction),
		"queued":  make(map[string]map[string]*RPCTransaction),
	}
	pending, queue := s.b.TxPoolContent()

	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}

// Status returns the number of pending and queued transaction in the pool.
func (s *PublicTxPoolAPI) Status() map[string]hexutil.Uint {
	pending, queue := s.b.Stats()
	return map[string]hexutil.Uint{
		"pending": hexutil.Uint(pending),
		"queued":  hexutil.Uint(queue),
	}
}

// Inspect retrieves the content of the transaction pool and flattens it into an
// easily inspectable list.
func (s *PublicTxPoolAPI) Inspect() map[string]map[string]map[string]string {
	content := map[string]map[string]map[string]string{
		"pending": make(map[string]map[string]string),
		"queued":  make(map[string]map[string]string),
	}
	pending, queue := s.b.TxPoolContent()

	// Define a formatter to flatten a transaction into a string
	var format = func(tx types.SelfTransaction) string {
		if to := tx.To(); to != nil {
			return fmt.Sprintf("%s: %v wei + %v gas × %v wei", tx.To().Hex(), tx.Value(), tx.Gas(), tx.GasPrice())
		}
		return fmt.Sprintf("contract creation: %v wei + %v gas × %v wei", tx.Value(), tx.Gas(), tx.GasPrice())
	}
	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}

// PublicAccountAPI provides an API to access accounts managed by this node.
// It offers only methods that can retrieve accounts.
type PublicAccountAPI struct {
	am *accounts.Manager
}

// NewPublicAccountAPI creates a new PublicAccountAPI.
func NewPublicAccountAPI(am *accounts.Manager) *PublicAccountAPI {
	return &PublicAccountAPI{am: am}
}

// Accounts returns the collection of accounts this node manages
func (s *PublicAccountAPI) Accounts() []string {
	//addresses := make([]common.Address, 0) // return [] instead of nil if empty
	strAddrList := make([]string, 0)
	var tmpstr string
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			strAddr := base58.Base58EncodeToString("MAN", account.Address)
			if tmpstr == strAddr {
				continue
			}
			tmpstr = strAddr
			strAddrList = append(strAddrList, strAddr)
		}
	}

	return strAddrList
}

// PrivateAccountAPI provides an API to access accounts managed by this node.
// It offers methods to create, (un)lock en list accounts. Some methods accept
// passwords and are therefore considered private by default.
type PrivateAccountAPI struct {
	am        *accounts.Manager
	nonceLock *AddrLocker
	b         Backend
}

// NewPrivateAccountAPI create a new PrivateAccountAPI.
func NewPrivateAccountAPI(b Backend, nonceLock *AddrLocker) *PrivateAccountAPI {
	return &PrivateAccountAPI{
		am:        b.AccountManager(),
		nonceLock: nonceLock,
		b:         b,
	}
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (s *PrivateAccountAPI) ListAccounts() []common.Address {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// rawWallet is a JSON representation of an accounts.Wallet interface, with its
// data contents extracted into plain fields.
type rawWallet struct {
	URL      string             `json:"url"`
	Status   string             `json:"status"`
	Failure  string             `json:"failure,omitempty"`
	Accounts []accounts.Account `json:"accounts,omitempty"`
}

// ListWallets will return a list of wallets this node manages.
func (s *PrivateAccountAPI) ListWallets() []rawWallet {
	wallets := make([]rawWallet, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		status, failure := wallet.Status()

		raw := rawWallet{
			URL:      wallet.URL().String(),
			Status:   status,
			Accounts: wallet.Accounts(),
		}
		if failure != nil {
			raw.Failure = failure.Error()
		}
		wallets = append(wallets, raw)
	}
	return wallets
}

// OpenWallet initiates a hardware wallet opening procedure, establishing a USB
// connection and attempting to authenticate via the provided passphrase. Note,
// the method may return an extra challenge requiring a second open (e.g. the
// Trezor PIN matrix challenge).
func (s *PrivateAccountAPI) OpenWallet(url string, passphrase *string) error {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return err
	}
	pass := ""
	if passphrase != nil {
		pass = *passphrase
	}
	return wallet.Open(pass)
}

// DeriveAccount requests a HD wallet to derive a new account, optionally pinning
// it for later reuse.
func (s *PrivateAccountAPI) DeriveAccount(url string, path string, pin *bool) (accounts.Account, error) {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return accounts.Account{}, err
	}
	derivPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return accounts.Account{}, err
	}
	if pin == nil {
		pin = new(bool)
	}
	return wallet.Derive(derivPath, *pin)
}

// NewAccount will create a new account and returns the address for the new account.
//func (s *PrivateAccountAPI) NewAccount(password string) (common.Address, error) {
//	acc, err := fetchKeystore(s.am).NewAccount(password)
//	if err == nil {
//		return acc.Address, nil
//	}
//	return common.Address{}, err
//}
func (s *PrivateAccountAPI) NewAccount(password string) (string, error) {
	acc, err := fetchKeystore(s.am).NewAccount(password)
	if err == nil {
		return acc.ManAddress, nil
	}
	return "", err
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

// ImportRawKey stores the given hex encoded ECDSA key into the key directory,
// encrypting it with the passphrase.
func (s *PrivateAccountAPI) ImportRawKey(privkey string, password string) (common.Address, error) {
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}
	acc, err := fetchKeystore(s.am).ImportECDSA(key, password)
	return acc.Address, err
}
func GetPassword() (string, error) {
	password, err := console.Stdin.PromptPassword("Passphrase: ")
	if err != nil {
		return "", fmt.Errorf("Failed to read passphrase: %v", err)
	}
	confirm, err := console.Stdin.PromptPassword("Repeat passphrase: ")
	if err != nil {
		return "", fmt.Errorf("Failed to read passphrase confirmation: %v", err)
	}
	if password != confirm {
		return "", fmt.Errorf("Passphrases do not match")
	}
	return password, nil
}

func (s *PrivateAccountAPI) SetEntrustSignAccount(path string, password string, times int64) bool {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("文件失败", err, "path", path)
		return false
	}

	b, err := ioutil.ReadAll(f)
	bytesPass, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		fmt.Println("解密失败", err)
		return false
	}
	tpass, err := aes.AesDecrypt(bytesPass, []byte(password))
	if err != nil {
		fmt.Println("AedDecrypt失败", bytesPass, password)
		return false
	}

	var anss []mc.EntrustInfo
	err = json.Unmarshal(tpass, &anss)
	if err != nil {
		fmt.Println("加密文件解码失败 密码不正确")
		return false
	}
	entrustValue := make(map[common.Address]string, 0)

	for _, v := range anss {
		entrustValue[base58.Base58DecodeToAddress(v.Address)] = v.Password
	}
	manparams.EntrustAccountValue.SetEntrustValue(entrustValue)
	go manparams.SetTimer(times)
	return true
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
func (s *PrivateAccountAPI) UnlockAccount(strAddr string, password string, duration *uint64) (bool, error) {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return false, errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	addr := base58.Base58DecodeToAddress(strAddr)
	err := fetchKeystore(s.am).TimedUnlock(accounts.Account{Address: addr}, password, d)
	return err == nil, err
}

// LockAccount will lock the account associated with the given address when it's unlocked.
func (s *PrivateAccountAPI) LockAccount(strAddr string) bool {
	addr := base58.Base58DecodeToAddress(strAddr)
	return fetchKeystore(s.am).Lock(addr) == nil
}

// signTransactions sets defaults and signs the given transaction
// NOTE: the caller needs to ensure that the nonceLock is held, if applicable,
// and release it after the transaction has been submitted to the tx pool
func (s *PrivateAccountAPI) signTransaction(ctx context.Context, args SendTxArgs, passwd string) (types.SelfTransaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}
	wallet, err := s.am.Find(account)
	if err != nil {
		return nil, err
	}
	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()
	tx.Currency = args.Currency

	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainId
	}
	return wallet.SignTxWithPassphrase(account, passwd, tx, chainID)
}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given passwd isn't
// able to decrypt the key it fails.
func (s *PrivateAccountAPI) SendTransaction(ctx context.Context, args1 SendTxArgs1, passwd string) (common.Hash, error) {
	var args SendTxArgs
	args, err := StrArgsToByteArgs(args1)
	if err != nil {
		return common.Hash{}, err
	}
	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}
	signed, err := s.signTransaction(ctx, args, passwd)
	if err != nil {
		return common.Hash{}, err
	}
	Currency := args.Currency //币种
	signed.SetTxCurrency(Currency)
	return submitTransaction(ctx, s.b, signed)
}

// SignTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given passwd isn't
// able to decrypt the key it fails. The transaction is returned in RLP-form, not broadcast
// to other nodes
func (s *PrivateAccountAPI) SignTransaction(ctx context.Context, args SendTxArgs, passwd string) (*SignTransactionResult, error) {
	// No need to obtain the noncelock mutex, since we won't be sending this
	// tx into the transaction pool, but right back to the user
	if args.Gas == nil {
		return nil, fmt.Errorf("gas not specified")
	}
	if args.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	if args.Nonce == nil {
		return nil, fmt.Errorf("nonce not specified")
	}
	signed, err := s.signTransaction(ctx, args, passwd)
	if err != nil {
		return nil, err
	}
	data, err := rlp.EncodeToBytes(signed)
	if err != nil {
		return nil, err
	}
	return &SignTransactionResult{data, signed}, nil
}

// signHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calulcated as
//   keccak256("\x19Matrix Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Matrix Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

// Sign calculates an Matrix ECDSA signature for:
// keccack256("\x19Matrix Signed Message:\n" + len(message) + message))
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The key used to calculate the signature is decrypted with the given password.
//
// https://github.com/matrix/go-matrix/wiki/Management-APIs#personal_sign
func (s *PrivateAccountAPI) Sign(ctx context.Context, data hexutil.Bytes, strAddr string, passwd string) (hexutil.Bytes, error) {
	// Look up the wallet containing the requested signer
	addr := base58.Base58DecodeToAddress(strAddr)
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Assemble sign the data with the wallet
	signature, err := wallet.SignHashWithPassphrase(account, passwd, signHash(data))
	if err != nil {
		return nil, err
	}
	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// EcRecover returns the address for the account that was used to create the signature.
// Note, this function is compatible with man_sign and personal_sign. As such it recovers
// the address of:
// hash = keccak256("\x19Matrix Signed Message:\n"${message length}${message})
// addr = ecrecover(hash, signature)
//
// Note, the signature must conform to the secp256k1 curve R, S and V values, where
// the V value must be be 27 or 28 for legacy reasons.
//
// https://github.com/matrix/go-matrix/wiki/Management-APIs#personal_ecRecover
func (s *PrivateAccountAPI) EcRecover(ctx context.Context, data, sig hexutil.Bytes) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return common.Address{}, fmt.Errorf("invalid Matrix signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.Ecrecover(signHash(data), sig)
	if err != nil {
		return common.Address{}, err
	}
	pubKey := crypto.ToECDSAPub(rpk)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr, nil
}

// SignAndSendTransaction was renamed to SendTransaction. This method is deprecated
// and will be removed in the future. It primary goal is to give clients time to update.
func (s *PrivateAccountAPI) SignAndSendTransaction(ctx context.Context, args SendTxArgs1, passwd string) (common.Hash, error) {
	return s.SendTransaction(ctx, args, passwd)
}

// PublicBlockChainAPI provides an API to access the Matrix blockchain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicBlockChainAPI struct {
	b Backend
}

// NewPublicBlockChainAPI creates a new Matrix blockchain API.
func NewPublicBlockChainAPI(b Backend) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{b}
}

// BlockNumber returns the block number of the chain head.
func (s *PublicBlockChainAPI) BlockNumber() *big.Int {
	header, _ := s.b.HeaderByNumber(context.Background(), rpc.LatestBlockNumber) // latest header should always be available
	return header.Number
}

type RPCBalanceType struct {
	AccountType uint32       `json:"accountType"`
	Balance     *hexutil.Big `json:"balance"`
}

// GetBalance returns the amount of wei for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicBlockChainAPI) GetBalance(ctx context.Context, strAddress string, blockNr rpc.BlockNumber) ([]RPCBalanceType, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	address := base58.Base58DecodeToAddress(strAddress)
	var balance []RPCBalanceType
	b := state.GetBalance(address)
	if b == nil {
		tmp := new(RPCBalanceType)
		var i uint32
		for i = 0; i <= common.LastAccount; i++ {
			tmp.AccountType = i
			tmp.Balance = new(hexutil.Big)
			balance = append(balance, *tmp)
		}
	} else {
		for i := 0; i < len(b); i++ {
			balance = append(balance, RPCBalanceType{b[i].AccountType, (*hexutil.Big)(b[i].Balance)})
		}
	}

	//log.Info("GetBalance","余额:",balance)
	return balance, state.Error()
}

//钱包调用
func (s *PublicBlockChainAPI) GetEntrustList(strAuthFrom string) []common.EntrustType {
	state, err := s.b.GetState()
	if state == nil || err != nil {
		return nil
	}
	authFrom := base58.Base58DecodeToAddress(strAuthFrom)
	return state.GetAllEntrustList(authFrom)
}

func (s *PublicBlockChainAPI) GetAuthFrom(strEntrustFrom string, height uint64) string {
	state, err := s.b.GetState()
	if state == nil || err != nil {
		return ""
	}
	entrustFrom := base58.Base58DecodeToAddress(strEntrustFrom)
	addr := state.GetAuthFrom(entrustFrom, height)
	if addr.Equal(common.Address{}) {
		return ""
	}
	return base58.Base58EncodeToString("MAN", addr)
}
func (s *PublicBlockChainAPI) GetEntrustFrom(strAuthFrom string, height uint64) []string {
	state, err := s.b.GetState()
	if state == nil || err != nil {
		return nil
	}
	entrustFrom := base58.Base58DecodeToAddress(strAuthFrom)
	addrList := state.GetEntrustFrom(entrustFrom, height)
	var strAddrList []string
	for _, addr := range addrList {
		if !addr.Equal(common.Address{}) {
			strAddr := base58.Base58EncodeToString("MAN", addr)
			strAddrList = append(strAddrList, strAddr)
		}
	}
	return strAddrList
}
func (s *PublicBlockChainAPI) GetAuthFromByTime(strEntrustFrom string, time uint64) string {
	state, err := s.b.GetState()
	if state == nil || err != nil {
		return ""
	}
	entrustFrom := base58.Base58DecodeToAddress(strEntrustFrom)
	addr := state.GetGasAuthFromByTime(entrustFrom, time)
	if addr.Equal(common.Address{}) {
		return ""
	}
	return base58.Base58EncodeToString("MAN", addr)
}
func (s *PublicBlockChainAPI) GetEntrustFromByTime(strAuthFrom string, time uint64) []string {
	state, err := s.b.GetState()
	if state == nil || err != nil {
		return nil
	}
	entrustFrom := base58.Base58DecodeToAddress(strAuthFrom)
	addrList := state.GetEntrustFromByTime(entrustFrom, time)
	var strAddrList []string
	for _, addr := range addrList {
		if !addr.Equal(common.Address{}) {
			strAddr := base58.Base58EncodeToString("MAN", addr)
			strAddrList = append(strAddrList, strAddr)
		}
	}
	return strAddrList
}

// GetBlockByNumber returns the requested block. When blockNr is -1 the chain head is returned. When fullTx is true all
// transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		//response, err := s.rpcOutputBlock(block, true, fullTx)
		response, err := s.rpcOutputBlock1(block, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}

// GetBlockByHash returns the requested block. When fullTx is true all transactions in the block are returned in full
// detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		//return s.rpcOutputBlock(block, true, fullTx)
		return s.rpcOutputBlock1(block, true, fullTx)
	}
	return nil, err
}

// GetUncleByBlockNumberAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		uncles := block.Uncles()
		if index >= hexutil.Uint(len(uncles)) {
			log.Debug("Requested uncle not found", "number", blockNr, "hash", block.Hash(), "index", index)
			return nil, nil
		}
		block = types.NewBlockWithHeader(uncles[index])
		return s.rpcOutputBlock(block, false, false)
	}
	return nil, err
}

// GetUncleByBlockHashAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		uncles := block.Uncles()
		if index >= hexutil.Uint(len(uncles)) {
			log.Debug("Requested uncle not found", "number", block.Number(), "hash", blockHash, "index", index)
			return nil, nil
		}
		block = types.NewBlockWithHeader(uncles[index])
		return s.rpcOutputBlock(block, false, false)
	}
	return nil, err
}

// GetUncleCountByBlockNumber returns number of uncles in the block for the given block number
func (s *PublicBlockChainAPI) GetUncleCountByBlockNumber(ctx context.Context, blockNr rpc.BlockNumber) *hexutil.Uint {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		n := hexutil.Uint(len(block.Uncles()))
		return &n
	}
	return nil
}

// GetUncleCountByBlockHash returns number of uncles in the block for the given block hash
func (s *PublicBlockChainAPI) GetUncleCountByBlockHash(ctx context.Context, blockHash common.Hash) *hexutil.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hexutil.Uint(len(block.Uncles()))
		return &n
	}
	return nil
}

// GetCode returns the code stored at the given address in the state for the given block number.
func (s *PublicBlockChainAPI) GetCode(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	code := state.GetCode(address)
	return code, state.Error()
}

// GetStorageAt returns the storage from the state at the given address, key and
// block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta block
// numbers are also allowed.
func (s *PublicBlockChainAPI) GetStorageAt(ctx context.Context, address common.Address, key string, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	res := state.GetState(address, common.HexToHash(key))
	return res[:], state.Error()
}

// CallArgs represents the arguments for a call.
type CallArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      hexutil.Uint64  `json:"gas"`
	GasPrice hexutil.Big     `json:"gasPrice"`
	Value    hexutil.Big     `json:"value"`
	Data     hexutil.Bytes   `json:"data"`
}
type ManCallArgs struct {
	From     string         `json:"from"`
	To       *string        `json:"to"`
	Gas      hexutil.Uint64 `json:"gas"`
	GasPrice hexutil.Big    `json:"gasPrice"`
	Value    hexutil.Big    `json:"value"`
	Data     hexutil.Bytes  `json:"data"`
}

func (s *PublicBlockChainAPI) doCall(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber, vmCfg vm.Config, timeout time.Duration) ([]byte, uint64, bool, error) {
	defer func(start time.Time) { log.Debug("Executing EVM call finished", "runtime", time.Since(start)) }(time.Now())

	state, header, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, 0, false, err
	}
	// Set sender address or use a default if none specified
	addr := args.From
	if addr == (common.Address{}) {
		if wallets := s.b.AccountManager().Wallets(); len(wallets) > 0 {
			if accounts := wallets[0].Accounts(); len(accounts) > 0 {
				addr = accounts[0].Address
			}
		}
	}
	// Set default gas & gas price if none were set
	gas, gasPrice := uint64(args.Gas), args.GasPrice.ToInt()
	if gas == 0 {
		gas = math.MaxUint64 / 2
	}
	if gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice)
	}

	// Create new call message
	//msg := new(types.Transaction) //types.NewMessage(addr, args.To, 0, args.Value.ToInt(), gas, gasPrice, args.Data, false)
	msg := &types.TransactionCall{types.NewTransaction(params.NonceAddOne, *args.To, args.Value.ToInt(), gas, gasPrice, args.Data, 0, 0)}
	msg.SetFromLoad(addr)
	// Setup context so it may be cancelled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	// Make sure the context is cancelled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	// Get a new instance of the EVM.
	evm, vmError, err := s.b.GetEVM(ctx, msg, state, header, vmCfg)
	if err != nil {
		return nil, 0, false, err
	}
	// Wait for the context to be done and cancel the evm. Even if the
	// EVM has finished, cancelling may be done (repeatedly)
	go func() {
		<-ctx.Done()
		evm.Cancel()
	}()

	// Setup the gas pool (also for unmetered requests)
	// and apply the message.
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	res, gas, failed, err := core.ApplyMessage(evm, msg, gp)
	if err := vmError(); err != nil {
		return nil, 0, false, err
	}
	return res, gas, failed, err
}

func ManArgsToCallArgs(manargs ManCallArgs) (args CallArgs) {
	args.From = base58.Base58DecodeToAddress(manargs.From)
	args.To = new(common.Address)
	*args.To = base58.Base58DecodeToAddress(*manargs.To)
	args.GasPrice = manargs.GasPrice
	args.Gas = manargs.Gas
	args.Value = manargs.Value
	args.Data = manargs.Data
	return
}

// Call executes the given transaction on the state for the given block number.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func (s *PublicBlockChainAPI) Call(ctx context.Context, manargs ManCallArgs, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	args := ManArgsToCallArgs(manargs)
	result, _, _, err := s.doCall(ctx, args, blockNr, vm.Config{}, 5*time.Second)
	return (hexutil.Bytes)(result), err
}

// EstimateGas returns an estimate of the amount of gas needed to execute the
// given transaction against the current pending block.
func (s *PublicBlockChainAPI) EstimateGas(ctx context.Context, args CallArgs) (hexutil.Uint64, error) {
	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  uint64 = params.TxGas - 1
		hi  uint64
		cap uint64
	)
	if uint64(args.Gas) >= params.TxGas {
		hi = uint64(args.Gas)
	} else {
		// Retrieve the current pending block to act as the gas ceiling
		block, err := s.b.BlockByNumber(ctx, rpc.PendingBlockNumber)
		if err != nil {
			return 0, err
		}
		hi = block.GasLimit()
	}
	cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) bool {
		args.Gas = hexutil.Uint64(gas)

		_, _, failed, err := s.doCall(ctx, args, rpc.PendingBlockNumber, vm.Config{}, 0)
		if err != nil || failed {
			return false
		}
		return true
	}
	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		if !executable(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == cap {
		if !executable(hi) {
			return 0, fmt.Errorf("gas required exceeds allowance or always failing transaction")
		}
	}
	return hexutil.Uint64(hi), nil
}

// GetTopology get topology from ca by block number.
func (s *PublicBlockChainAPI) GetTopology(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
	return ca.GetTopologyByNumber(reqTypes, number)
}

// GetSelfLevel get self level from ca, including top node, buckets number and default.
func (s *PublicBlockChainAPI) GetSelfLevel() int {
	return ca.GetSelfLevel()
}

// GetSignAccounts get sign accounts form current block.
func (s *PublicBlockChainAPI) getSignAccountsByNumber1(ctx context.Context, blockNr rpc.BlockNumber) ([]common.VerifiedSign, error) {
	header, err := s.b.HeaderByNumber(ctx, blockNr)
	if header != nil {
		return header.SignAccounts(), nil
	}
	return nil, err
}

func (s *PublicBlockChainAPI) GetSignAccountsByNumber(ctx context.Context, blockNr rpc.BlockNumber) ([]common.VerifiedSign1, error) {
	verSignList, err := s.getSignAccountsByNumber1(ctx, blockNr)
	if err != nil {
		return nil, err
	}

	accounts := make([]common.VerifiedSign1, 0)
	for _, tmpverSign := range verSignList {
		accounts = append(accounts, common.VerifiedSign1{
			Sign:     tmpverSign.Sign,
			Account:  base58.Base58EncodeToString("MAN", tmpverSign.Account),
			Validate: tmpverSign.Validate,
			Stock:    tmpverSign.Stock,
		})
	}
	return accounts, nil
}

func (s *PublicBlockChainAPI) getSignAccountsByHash1(ctx context.Context, hash common.Hash) ([]common.VerifiedSign, error) {
	block, err := s.b.GetBlock(ctx, hash)
	if block != nil {
		return block.SignAccounts(), nil
	}
	return nil, err
}
func (s *PublicBlockChainAPI) GetSignAccountsByHash(ctx context.Context, hash common.Hash) ([]common.VerifiedSign1, error) {
	verSignList, err := s.getSignAccountsByHash1(ctx, hash)
	if err != nil {
		return nil, err
	}
	accounts := make([]common.VerifiedSign1, 0)
	for _, tmpverSign := range verSignList {
		accounts = append(accounts, common.VerifiedSign1{
			Sign:     tmpverSign.Sign,
			Account:  base58.Base58EncodeToString("MAN", tmpverSign.Account),
			Validate: tmpverSign.Validate,
			Stock:    tmpverSign.Stock,
		})
	}
	return accounts, nil
}
func (s *PublicBlockChainAPI) ImportSuperBlock(ctx context.Context, filePath string) (common.Hash, error) {
	return s.b.ImportSuperBlock(ctx, filePath)
}

// ExecutionResult groups all structured logs emitted by the EVM
// while replaying a transaction in debug mode as well as transaction
// execution status, the amount of gas used and the return value
type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

// StructLogRes stores a structured log emitted by the EVM while replaying a
// transaction in debug mode
type StructLogRes struct {
	Pc      uint64             `json:"pc"`
	Op      string             `json:"op"`
	Gas     uint64             `json:"gas"`
	GasCost uint64             `json:"gasCost"`
	Depth   int                `json:"depth"`
	Error   error              `json:"error,omitempty"`
	Stack   *[]string          `json:"stack,omitempty"`
	Memory  *[]string          `json:"memory,omitempty"`
	Storage *map[string]string `json:"storage,omitempty"`
}

// formatLogs formats EVM returned structured logs for json output
func FormatLogs(logs []vm.StructLog) []StructLogRes {
	formatted := make([]StructLogRes, len(logs))
	for index, trace := range logs {
		formatted[index] = StructLogRes{
			Pc:      trace.Pc,
			Op:      trace.Op.String(),
			Gas:     trace.Gas,
			GasCost: trace.GasCost,
			Depth:   trace.Depth,
			Error:   trace.Err,
		}
		if trace.Stack != nil {
			stack := make([]string, len(trace.Stack))
			for i, stackValue := range trace.Stack {
				stack[i] = fmt.Sprintf("%x", math.PaddedBigBytes(stackValue, 32))
			}
			formatted[index].Stack = &stack
		}
		if trace.Memory != nil {
			memory := make([]string, 0, (len(trace.Memory)+31)/32)
			for i := 0; i+32 <= len(trace.Memory); i += 32 {
				memory = append(memory, fmt.Sprintf("%x", trace.Memory[i:i+32]))
			}
			formatted[index].Memory = &memory
		}
		if trace.Storage != nil {
			storage := make(map[string]string)
			for i, storageValue := range trace.Storage {
				storage[fmt.Sprintf("%x", i)] = fmt.Sprintf("%x", storageValue)
			}
			formatted[index].Storage = &storage
		}
	}
	return formatted
}

// rpcOutputBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func (s *PublicBlockChainAPI) rpcOutputBlock(b *types.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	head := b.Header() // copies the header once
	fields := map[string]interface{}{
		"number":            (*hexutil.Big)(head.Number),
		"hash":              b.Hash(),
		"parentHash":        head.ParentHash,
		"nonce":             head.Nonce,
		"mixHash":           head.MixDigest,
		"sha3Uncles":        head.UncleHash,
		"logsBloom":         head.Bloom,
		"stateRoot":         head.Root,
		"miner":             head.Coinbase,
		"difficulty":        (*hexutil.Big)(head.Difficulty),
		"totalDifficulty":   (*hexutil.Big)(s.b.GetTd(b.Hash())),
		"extraData":         hexutil.Bytes(head.Extra),
		"size":              hexutil.Uint64(b.Size()),
		"gasLimit":          hexutil.Uint64(head.GasLimit),
		"gasUsed":           hexutil.Uint64(head.GasUsed),
		"timestamp":         (*hexutil.Big)(head.Time),
		"transactionsRoot":  head.TxHash,
		"receiptsRoot":      head.ReceiptHash,
		"leader":            head.Leader,
		"elect":             head.Elect,
		"nettopology":       head.NetTopology,
		"signatures":        head.Signatures,
		"version":           string(head.Version),
		"versionSignatures": head.VersionSignatures,
		"vrfvalue":          hexutil.Bytes(head.VrfValue),
	}

	if inclTx {
		formatTx := func(tx types.SelfTransaction) (interface{}, error) {
			return tx.Hash(), nil
		}

		if fullTx {
			formatTx = func(tx types.SelfTransaction) (interface{}, error) {
				return newRPCTransactionFromBlockHash(b, tx.Hash()), nil
			}
		}

		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		var err error
		for i, tx := range b.Transactions() {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}

	uncles := b.Uncles()
	uncleHashes := make([]common.Hash, len(uncles))
	for i, uncle := range uncles {
		uncleHashes[i] = uncle.Hash()
	}
	fields["uncles"] = uncleHashes

	return fields, nil
}

/************************************************************/
func (s *PublicBlockChainAPI) rpcOutputBlock1(b *types.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	head := b.Header() // copies the header once
	Coinbase1 := base58.Base58EncodeToString("MAN", head.Coinbase)
	Leader1 := base58.Base58EncodeToString("MAN", head.Leader)
	//head.NetTopology
	NetTopology1 := new(common.NetTopology1)
	listNetTopolog := make([]common.NetTopologyData1, 0)
	for _, addr := range head.NetTopology.NetTopologyData {
		tmpstruct := new(common.NetTopologyData1)
		tmpstruct.Account = base58.Base58EncodeToString("MAN", addr.Account)
		tmpstruct.Position = addr.Position
		listNetTopolog = append(listNetTopolog, *tmpstruct)
	}
	NetTopology1.Type = head.NetTopology.Type
	NetTopology1.NetTopologyData = append(NetTopology1.NetTopologyData, listNetTopolog...)

	//head.Elect
	listElect1 := make([]common.Elect1, 0)
	for _, elect := range head.Elect {
		tmpElect1 := new(common.Elect1)
		tmpElect1.Type = elect.Type
		tmpElect1.Account = base58.Base58EncodeToString("MAN", elect.Account)
		tmpElect1.Stock = elect.Stock
		listElect1 = append(listElect1, *tmpElect1)
	}

	fields := map[string]interface{}{
		"number":           (*hexutil.Big)(head.Number),
		"hash":             b.Hash(),
		"parentHash":       head.ParentHash,
		"nonce":            head.Nonce,
		"mixHash":          head.MixDigest,
		"sha3Uncles":       head.UncleHash,
		"logsBloom":        head.Bloom,
		"stateRoot":        head.Root,
		"miner":            Coinbase1,
		"difficulty":       (*hexutil.Big)(head.Difficulty),
		"totalDifficulty":  (*hexutil.Big)(s.b.GetTd(b.Hash())),
		"extraData":        hexutil.Bytes(head.Extra),
		"size":             hexutil.Uint64(b.Size()),
		"gasLimit":         hexutil.Uint64(head.GasLimit),
		"gasUsed":          hexutil.Uint64(head.GasUsed),
		"timestamp":        (*hexutil.Big)(head.Time),
		"transactionsRoot": head.TxHash,
		"receiptsRoot":     head.ReceiptHash,
		"leader":           Leader1,
		"elect":            listElect1,
		"nettopology":      NetTopology1,
		"signatures":       head.Signatures,
		"version":          hexutil.Bytes(head.Version),
	}

	if inclTx {
		formatTx := func(tx types.SelfTransaction) (interface{}, error) {
			return tx.Hash(), nil
		}

		if fullTx {
			formatTx = func(tx types.SelfTransaction) (interface{}, error) {
				return newRPCTransactionFromBlockHash(b, tx.Hash()), nil
			}
		}

		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		var err error
		for i, tx := range b.Transactions() {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}

	uncles := b.Uncles()
	uncleHashes := make([]common.Hash, len(uncles))
	for i, uncle := range uncles {
		uncleHashes[i] = uncle.Hash()
	}
	fields["uncles"] = uncleHashes

	return fields, nil
}

//hezi
type RPCTransaction1 struct {
	BlockHash        common.Hash    `json:"blockHash"`
	BlockNumber      *hexutil.Big   `json:"blockNumber"`
	From             string         `json:"from"`
	Gas              hexutil.Uint64 `json:"gas"`
	GasPrice         *hexutil.Big   `json:"gasPrice"`
	Hash             common.Hash    `json:"hash"`
	Input            hexutil.Bytes  `json:"input"`
	Nonce            hexutil.Uint64 `json:"nonce"`
	To               *string        `json:"to"`
	TransactionIndex hexutil.Uint   `json:"transactionIndex"`
	Value            *hexutil.Big   `json:"value"`
	V                *hexutil.Big   `json:"v"`
	R                *hexutil.Big   `json:"r"`
	S                *hexutil.Big   `json:"s"`
	TxEnterType      byte           `json:"TxEnterType"`
	IsEntrustTx      bool           `json:"IsEntrustTx"`
	Currency         string         `json:"Currency"`
	CommitTime       hexutil.Uint64 `json:"CommitTime"`
	MatrixType       byte           `json:"matrixType"`
	ExtraTo          []*ExtraTo_Mx1 `json:"extra_to"`
}

func RPCTransactionToString(data *RPCTransaction) *RPCTransaction1 {
	result := &RPCTransaction1{
		BlockHash:        data.BlockHash,
		BlockNumber:      data.BlockNumber,
		Gas:              data.Gas,
		GasPrice:         data.GasPrice,
		Hash:             data.Hash,
		Input:            data.Input,
		Nonce:            data.Nonce,
		TransactionIndex: data.TransactionIndex,
		Value:            data.Value,
		V:                data.V,
		R:                data.R,
		S:                data.S,
		TxEnterType:      data.TxEnterType,
		IsEntrustTx:      data.IsEntrustTx,
		Currency:         data.Currency,
		MatrixType:       data.MatrixType,
		CommitTime:       data.CommitTime,
	}
	//内部发送的交易没有币种，默认为MAN
	if data.Currency == "" {
		data.Currency = "MAN"
	}
	result.From = base58.Base58EncodeToString(data.Currency, data.From)
	result.To = new(string)
	*result.To = base58.Base58EncodeToString(data.Currency, *data.To)

	if len(data.ExtraTo) > 0 {
		extra := make([]*ExtraTo_Mx1, 0)
		for _, ar := range data.ExtraTo {
			if ar.To2 != nil {
				tmExtra := new(ExtraTo_Mx1)
				tmExtra.To2 = new(string)
				*tmExtra.To2 = base58.Base58EncodeToString(data.Currency, *ar.To2)
				tmExtra.Input2 = ar.Input2
				tmExtra.Value2 = ar.Value2
				extra = append(extra, tmExtra)
			}
		}
		result.ExtraTo = extra
	}

	return result
}

/************************************************************/
// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex hexutil.Uint    `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
	TxEnterType      byte            `json:"TxEnterType"`
	IsEntrustTx      bool            `json:"IsEntrustTx"`
	Currency         string          `json:"Currency"`
	CommitTime       hexutil.Uint64  `json:"CommitTime"`
	MatrixType       byte            `json:"matrixType"`
	ExtraTo          []*ExtraTo_Mx   `json:"extra_to"`
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx types.SelfTransaction, blockHash common.Hash, blockNumber uint64, index uint64) *RPCTransaction {
	var signer types.Signer //= types.FrontierSigner{}
	//if tx.Protected() {
	signer = types.NewEIP155Signer(tx.ChainId())
	//}

	var from common.Address

	if tx.GetMatrixType() == common.ExtraUnGasTxType {
		from = tx.From()
	} else {
		from, _ = types.Sender(signer, tx)
	}
	v, r, s := tx.RawSignatureValues()

	result := &RPCTransaction{
		From:        from,
		Gas:         hexutil.Uint64(tx.Gas()),
		GasPrice:    (*hexutil.Big)(tx.GasPrice()),
		Hash:        tx.Hash(),
		Input:       hexutil.Bytes(tx.Data()),
		Nonce:       hexutil.Uint64(tx.Nonce()),
		To:          tx.To(),
		Value:       (*hexutil.Big)(tx.Value()),
		V:           (*hexutil.Big)(v),
		R:           (*hexutil.Big)(r),
		S:           (*hexutil.Big)(s),
		TxEnterType: tx.TxType(),
		IsEntrustTx: tx.IsEntrustTx(),
		Currency:    tx.GetTxCurrency(),
		MatrixType:  tx.GetMatrixType(),
		CommitTime:  hexutil.Uint64(tx.GetCreateTime()),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = hexutil.Uint(index)
	}

	extra := tx.GetMatrix_EX()
	for _, ext := range extra {
		for _, e := range ext.ExtraTo {
			b := hexutil.Bytes(e.Payload)
			result.ExtraTo = append(result.ExtraTo, &ExtraTo_Mx{
				To2:    e.Recipient,
				Input2: &b,
				Value2: (*hexutil.Big)(e.Amount),
			})
		}
	}
	return result
}

// newRPCPendingTransaction returns a pending transaction that will serialize to the RPC representation
func newRPCPendingTransaction(tx types.SelfTransaction) *RPCTransaction {
	return newRPCTransaction(tx, common.Hash{}, 0, 0)
}

// newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockIndex(b *types.Block, index uint64) *RPCTransaction1 {
	rpcTrans := newRPCTransactionFromBlockIndex1(b, index)
	if rpcTrans != nil {
		return RPCTransactionToString(rpcTrans)
	}
	return nil
}

func newRPCTransactionFromBlockIndex1(b *types.Block, index uint64) *RPCTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), index)
}

// newRPCRawTransactionFromBlockIndex returns the bytes of a transaction given a block and a transaction index.
func newRPCRawTransactionFromBlockIndex(b *types.Block, index uint64) hexutil.Bytes {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	blob, _ := rlp.EncodeToBytes(txs[index])
	return blob
}

// newRPCTransactionFromBlockHash returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockHash(b *types.Block, hash common.Hash) *RPCTransaction1 {
	for idx, tx := range b.Transactions() {
		if tx.Hash() == hash {
			return newRPCTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}

// PublicTransactionPoolAPI exposes methods for the RPC interface
type PublicTransactionPoolAPI struct {
	b         Backend
	nonceLock *AddrLocker
}

// NewPublicTransactionPoolAPI creates a new RPC service with methods specific for the transaction pool.
func NewPublicTransactionPoolAPI(b Backend, nonceLock *AddrLocker) *PublicTransactionPoolAPI {
	return &PublicTransactionPoolAPI{b, nonceLock}
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block with the given block number.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByNumber(ctx context.Context, blockNr rpc.BlockNumber) *hexutil.Uint {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block with the given hash.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByHash(ctx context.Context, blockHash common.Hash) *hexutil.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetTransactionByBlockNumberAndIndex returns the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) *RPCTransaction1 {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) *RPCTransaction1 {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockNumberAndIndex returns the bytes of the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockHashAndIndex returns the bytes of the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionCount returns the number of transactions the given address has sent for the given block number
func (s *PublicTransactionPoolAPI) GetTransactionCount(ctx context.Context, strAddress string, blockNr rpc.BlockNumber) (*hexutil.Uint64, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	address := base58.Base58DecodeToAddress(strAddress)
	nonce := state.GetNonce(address)
	return (*hexutil.Uint64)(&nonce), state.Error()
}

// GetTransactionByHash returns the transaction for the given hash
func (s *PublicTransactionPoolAPI) getTransactionByHash1(ctx context.Context, hash common.Hash) *RPCTransaction {
	// Try to return an already finalized transaction
	if tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash); tx != nil {
		return newRPCTransaction(tx, blockHash, blockNumber, index)
	}
	// No finalized transaction, try to retrieve it from the pool
	if tx := s.b.GetPoolTransaction(hash); tx != nil {
		return newRPCPendingTransaction(tx)
	}
	// Transaction unknown, return as such
	return nil
}

//hezi
func (s *PublicTransactionPoolAPI) GetTransactionByHash(ctx context.Context, hash common.Hash) *RPCTransaction1 {
	rpcTrans := s.getTransactionByHash1(ctx, hash)
	if rpcTrans != nil {
		return RPCTransactionToString(rpcTrans)
	}
	return nil
}

// GetRawTransactionByHash returns the bytes of the transaction for the given hash.
func (s *PublicTransactionPoolAPI) GetRawTransactionByHash(ctx context.Context, hash common.Hash) (hexutil.Bytes, error) {
	var tx types.SelfTransaction

	// Retrieve a finalized transaction, or a pooled otherwise
	if tx, _, _, _ = rawdb.ReadTransaction(s.b.ChainDb(), hash); tx == nil {
		if tx = s.b.GetPoolTransaction(hash); tx == nil {
			// Transaction not found anywhere, abort
			return nil, nil
		}
	}
	// Serialize to RLP and return
	return rlp.EncodeToBytes(tx)
}

// GetTransactionReceipt returns the transaction receipt for the given transaction hash.
func (s *PublicTransactionPoolAPI) GetTransactionReceipt(ctx context.Context, hash common.Hash) (map[string]interface{}, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash)
	if tx == nil {
		return nil, nil
	}
	receipts, err := s.b.GetReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	if len(receipts) <= int(index) {
		return nil, nil
	}
	receipt := receipts[index]

	var signer types.Signer //= types.FrontierSigner{}
	//if tx.Protected() {
	signer = types.NewEIP155Signer(tx.ChainId())
	//}
	from, _ := types.Sender(signer, tx)

	fields := map[string]interface{}{
		"blockHash":         blockHash,
		"blockNumber":       hexutil.Uint64(blockNumber),
		"transactionHash":   hash,
		"transactionIndex":  hexutil.Uint64(index),
		"from":              from,
		"to":                tx.To(),
		"gasUsed":           hexutil.Uint64(receipt.GasUsed),
		"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
		"contractAddress":   nil,
		"logs":              receipt.Logs,
		"logsBloom":         receipt.Bloom,
	}

	// Assign receipt status or post state.
	if len(receipt.PostState) > 0 {
		fields["root"] = hexutil.Bytes(receipt.PostState)
	} else {
		fields["status"] = hexutil.Uint(receipt.Status)
	}
	if receipt.Logs == nil {
		fields["logs"] = [][]*types.Log{}
	}
	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if receipt.ContractAddress != (common.Address{}) {
		fields["contractAddress"] = receipt.ContractAddress
	}
	return fields, nil
}

// sign is a helper function that signs a transaction with the private key of the given address.
func (s *PublicTransactionPoolAPI) sign(strAddr string, tx types.SelfTransaction) (types.SelfTransaction, error) {
	addr := base58.Base58DecodeToAddress(strAddr)

	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Request the wallet to sign the transaction
	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainId
	}
	return wallet.SignTx(account, tx, chainID)
}

//YY
type ExtraTo_Mx struct {
	To2    *common.Address `json:"to"`
	Value2 *hexutil.Big    `json:"value"`
	Input2 *hexutil.Bytes  `json:"input"`
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	Currency string          `json:"currency"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Nonce    *hexutil.Uint64 `json:"nonce"`
	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred by clients.
	Data        *hexutil.Bytes `json:"data"`
	Input       *hexutil.Bytes `json:"input"`
	TxType      byte           `json:"txType"`     //YY
	LockHeight  uint64         `json:"lockHeight"` //YY
	IsEntrustTx byte           `json:"isEntrustTx"`
	ExtraTo     []*ExtraTo_Mx  `json:"extra_to"` //YY
}

type ExtraTo_Mx1 struct {
	To2    *string        `json:"to"`
	Value2 *hexutil.Big   `json:"value"`
	Input2 *hexutil.Bytes `json:"input"`
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs1 struct {
	From     string          `json:"from"`
	To       *string         `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Nonce    *hexutil.Uint64 `json:"nonce"`
	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred by clients.
	Data        *hexutil.Bytes `json:"data"`
	Input       *hexutil.Bytes `json:"input"`
	TxType      byte           `json:"txType"`     //YY
	LockHeight  uint64         `json:"lockHeight"` //YY
	IsEntrustTx byte           `json:"isEntrustTx"`
	ExtraTo     []*ExtraTo_Mx1 `json:"extra_to"` //YY
}

// setDefaults is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(ctx context.Context, b Backend) error {
	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		//YY
		if len(args.ExtraTo) > 0 && args.LockHeight > 0 && args.TxType > 0 {
			*(*uint64)(args.Gas) = 21000 * uint64(len(args.ExtraTo))
		} else {
			*(*uint64)(args.Gas) = 21000
		}
	}
	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		if price.Cmp(new(big.Int).SetUint64(params.TxGasPrice)) < 0 {
			price.Set(new(big.Int).SetUint64(params.TxGasPrice))
		}
		args.GasPrice = (*hexutil.Big)(price)
	}
	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Nonce == nil {
		nonce, err := b.GetPoolNonce(ctx, args.From)
		if err != nil {
			return err
		}
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}
	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return errors.New(`Both "data" and "input" are set and not equal. Please use "input" to pass transaction call data.`)
	}
	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}
		if len(input) == 0 {
			return errors.New(`contract creation without any data provided`)
		}
	}
	return nil
}

func (args *SendTxArgs) toTransaction() *types.Transaction {
	var input []byte
	if args.Data != nil {
		input = *args.Data
	} else if args.Input != nil {
		input = *args.Input
	}
	if args.To == nil {
		return types.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input, 0, args.IsEntrustTx)
	}
	if args.TxType == 0 && args.LockHeight == 0 && args.ExtraTo == nil { //YY
		return types.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input, 0, args.IsEntrustTx)
	}
	//YY
	txtr := make([]*types.ExtraTo_tr, 0)
	if len(args.ExtraTo) > 0 {
		for _, extra := range args.ExtraTo {
			tmp := new(types.ExtraTo_tr)
			va := extra.Value2
			if va == nil {
				va = (*hexutil.Big)(big.NewInt(0))
			}
			tmp.To_tr = extra.To2
			tmp.Value_tr = va
			tmp.Input_tr = extra.Input2
			txtr = append(txtr, tmp)
		}
	}
	return types.NewTransactions(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input, txtr, args.LockHeight, args.TxType, args.IsEntrustTx)

}

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx types.SelfTransaction) (common.Hash, error) {
	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	if tx.To() == nil {
		signer := types.MakeSigner(b.ChainConfig(), b.CurrentBlock().Number())
		from, err := types.Sender(signer, tx)
		if err != nil {
			return common.Hash{}, err
		}
		addr := crypto.CreateAddress(from, tx.Nonce())
		log.Info("Submitted contract creation", "fullhash", tx.Hash().Hex(), "contract", addr.Hex())
	} else {
		//log.Info("Submitted transaction", "fullhash", tx.Hash().Hex(), "recipient", tx.To())
	}
	//log.Info("file api","func submitTransaction",tx.Hash().String())
	return tx.Hash(), nil
}

func CheckCrc8(strData string) bool {
	Crc := strData[len(strData)-1 : len(strData)]
	reCrc := crc8.CalCRC8([]byte(strData[0 : len(strData)-1]))
	ModCrc := reCrc % 58
	ret := base58.EncodeInt(ModCrc)
	if Crc != ret {
		return false
	}
	return true
}
func CheckCurrency(strData string) bool {
	currency := strings.Split(strData, ".")[0]
	if len(currency) < 2 || len(currency) > 8 {
		return false
	}
	return true
}
func CheckFormat(strData string) bool {
	if !strings.Contains(strData, ".") {
		return false
	}
	return true
}
func CheckParams(strData string) error {
	if !CheckFormat(strData) {
		return errors.New("format error")
	}
	if !CheckCrc8(strData) {
		return errors.New("CRC error")
	}
	if !CheckCurrency(strData) {
		return errors.New("currency error")
	}
	return nil
}
func StrArgsToByteArgs(args1 SendTxArgs1) (args SendTxArgs, err error) {
	from := args1.From
	err = CheckParams(from)
	if err != nil {
		return SendTxArgs{}, err
	}
	args.Currency = strings.Split(args1.From, ".")[0]
	args.From = base58.Base58DecodeToAddress(from)
	if args1.To != nil {
		to := *args1.To
		err = CheckParams(to)
		if err != nil {
			return SendTxArgs{}, err
		}
		args.To = new(common.Address)
		*args.To = base58.Base58DecodeToAddress(to)
	}
	args.Gas = args1.Gas
	args.GasPrice = args1.GasPrice
	args.Value = args1.Value
	args.Nonce = args1.Nonce
	args.Data = args1.Data
	args.Input = args1.Input
	args.TxType = args1.TxType
	args.LockHeight = args1.LockHeight
	args.IsEntrustTx = args1.IsEntrustTx
	if len(args1.ExtraTo) > 0 { //扩展交易中的to属性不填写则删掉这个扩展交易
		extra := make([]*ExtraTo_Mx, 0)
		for _, ar := range args1.ExtraTo {
			if ar.To2 != nil {
				//extra = append(extra, ar)
				tmp := *ar.To2
				err = CheckParams(tmp)
				if err != nil {
					return SendTxArgs{}, err
				}
				tmExtra := new(ExtraTo_Mx)
				tmExtra.To2 = new(common.Address)
				*tmExtra.To2 = base58.Base58DecodeToAddress(tmp)
				tmExtra.Input2 = ar.Input2
				tmExtra.Value2 = ar.Value2
				extra = append(extra, tmExtra)
			}
		}
		args.ExtraTo = extra
	}
	return args, nil
}

// SendTransaction creates a transaction for the given argument, sign it and submit it to the
// transaction pool.
func (s *PublicTransactionPoolAPI) SendTransaction(ctx context.Context, args1 SendTxArgs1, passwd string) (common.Hash, error) {
	//from字段格式: 2-8长度币种（大写）+ “.”+ 以太坊地址的base58编码 + crc8/58
	var args SendTxArgs
	args, err := StrArgsToByteArgs(args1)
	if err != nil {
		return common.Hash{}, err
	}
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}
	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return common.Hash{}, err
	}
	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	} else { //YY add else
		nc1 := params.NonceAddOne
		nc := uint64(*args.Nonce)
		if nc < nc1 {
			err = errors.New("Nonce Wrongful")
			return common.Hash{}, err
		}
	}
	//YY
	if len(args.ExtraTo) > 0 { //扩展交易中的to和input属性不填写则删掉这个扩展交易
		extra := make([]*ExtraTo_Mx, 0)
		for _, ar := range args.ExtraTo {
			if ar.To2 != nil || ar.Input2 != nil {
				extra = append(extra, ar)
			}
		}
		args.ExtraTo = extra
	}
	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()
	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainId
	}
	tx.Currency = args.Currency
	signed, err := wallet.SignTxWithPassphrase(account, passwd, tx, chainID)
	if err != nil {
		return common.Hash{}, err
	}
	Currency := args.Currency //币种
	signed.SetTxCurrency(Currency)
	return submitTransaction(ctx, s.b, signed)
}

// SendRawTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *PublicTransactionPoolAPI) SendRawTransaction(ctx context.Context, encodedTx hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	tx.Mtype = true
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, tx)
}

// Sign calculates an ECDSA signature for:
// keccack256("\x19Matrix Signed Message:\n" + len(message) + message).
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The account associated with addr must be unlocked.
//
// https://github.com/matrix/wiki/wiki/JSON-RPC#man_sign
func (s *PublicTransactionPoolAPI) Sign(strAddr string, data hexutil.Bytes) (hexutil.Bytes, error) {
	addr := base58.Base58DecodeToAddress(strAddr)
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Sign the requested hash with the wallet
	signature, err := wallet.SignHash(account, signHash(data))
	if err == nil {
		signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	}
	return signature, err
}

// SignTransactionResult represents a RLP encoded signed transaction.
type SignTransactionResult struct {
	Raw hexutil.Bytes         `json:"raw"`
	Tx  types.SelfTransaction `json:"tx"`
}

// SignTransaction will sign the given transaction with the from account.
// The node needs to have the private key of the account corresponding with
// the given from address and it needs to be unlocked.
func (s *PublicTransactionPoolAPI) SignTransaction(ctx context.Context, args1 SendTxArgs1) (*SignTransactionResult, error) {
	var args SendTxArgs
	args, err := StrArgsToByteArgs(args1)
	if err != nil {
		return nil, err
	}
	if args.Gas == nil {
		return nil, fmt.Errorf("gas not specified")
	}
	if args.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	if args.Nonce == nil {
		return nil, fmt.Errorf("nonce not specified")
	}
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	tx, err := s.sign(args1.From, args.toTransaction())
	if err != nil {
		return nil, err
	}
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	return &SignTransactionResult{data, tx}, nil
}

// PendingTransactions returns the transactions that are in the transaction pool and have a from address that is one of
// the accounts this node manages.
func (s *PublicTransactionPoolAPI) PendingTransactions() ([]*RPCTransaction, error) {
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return nil, err
	}

	transactions := make([]*RPCTransaction, 0, len(pending))
	for _, tx := range pending {
		var signer types.Signer //= types.HomesteadSigner{}
		//if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
		//}
		from, _ := types.Sender(signer, tx)
		if _, err := s.b.AccountManager().Find(accounts.Account{Address: from}); err == nil {
			transactions = append(transactions, newRPCPendingTransaction(tx))
		}
	}
	return transactions, nil
}

// Resend accepts an existing transaction and a new gas price and limit. It will remove
// the given transaction from the pool and reinsert it with the new gas price and limit.
func (s *PublicTransactionPoolAPI) Resend(ctx context.Context, sendArgs1 SendTxArgs1, gasPrice *hexutil.Big, gasLimit *hexutil.Uint64) (common.Hash, error) {
	var sendArgs SendTxArgs
	sendArgs, err := StrArgsToByteArgs(sendArgs1)
	if err != nil {
		return common.Hash{}, err
	}
	if sendArgs.Nonce == nil {
		return common.Hash{}, fmt.Errorf("missing transaction nonce in transaction spec")
	}
	if err := sendArgs.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	matchTx := sendArgs.toTransaction()
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return common.Hash{}, err
	}

	for _, p := range pending {
		var signer types.Signer //= types.HomesteadSigner{}
		//if p.Protected() {
		signer = types.NewEIP155Signer(p.ChainId())
		//}
		wantSigHash := signer.Hash(matchTx)

		if pFrom, err := types.Sender(signer, p); err == nil && pFrom == sendArgs.From && signer.Hash(p) == wantSigHash {
			// Match. Re-sign and send the transaction.
			if gasPrice != nil && (*big.Int)(gasPrice).Sign() != 0 {
				sendArgs.GasPrice = gasPrice
			}
			if gasLimit != nil && *gasLimit != 0 {
				sendArgs.Gas = gasLimit
			}
			Currency := strings.Split(sendArgs1.From, ".")[0] //币种
			strFrom := base58.Base58EncodeToString(Currency, sendArgs.From)
			signedTx, err := s.sign(strFrom, sendArgs.toTransaction())
			if err != nil {
				return common.Hash{}, err
			}
			if err = s.b.SendTx(ctx, signedTx); err != nil {
				return common.Hash{}, err
			}
			return signedTx.Hash(), nil
		}
	}

	return common.Hash{}, fmt.Errorf("Transaction %#x not found", matchTx.Hash())
}

// PublicDebugAPI is the collection of Matrix APIs exposed over the public
// debugging endpoint.
type PublicDebugAPI struct {
	b Backend
}

// NewPublicDebugAPI creates a new API definition for the public debug methods
// of the Matrix service.
func NewPublicDebugAPI(b Backend) *PublicDebugAPI {
	return &PublicDebugAPI{b: b}
}

// GetBlockRlp retrieves the RLP encoded for of a single block.
func (api *PublicDebugAPI) GetBlockRlp(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	encoded, err := rlp.EncodeToBytes(block)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", encoded), nil
}

// PrintBlock retrieves a block and returns its pretty printed form.
func (api *PublicDebugAPI) PrintBlock(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	return spew.Sdump(block), nil
}

// SeedHash retrieves the seed hash of a block.
func (api *PublicDebugAPI) SeedHash(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	return fmt.Sprintf("0x%x", manash.SeedHash(number)), nil
}

// PrivateDebugAPI is the collection of Matrix APIs exposed over the private
// debugging endpoint.
type PrivateDebugAPI struct {
	b Backend
}

// NewPrivateDebugAPI creates a new API definition for the private debug methods
// of the Matrix service.
func NewPrivateDebugAPI(b Backend) *PrivateDebugAPI {
	return &PrivateDebugAPI{b: b}
}

// ChaindbProperty returns leveldb properties of the chain database.
func (api *PrivateDebugAPI) ChaindbProperty(property string) (string, error) {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return "", fmt.Errorf("chaindbProperty does not work for memory databases")
	}
	if property == "" {
		property = "leveldb.stats"
	} else if !strings.HasPrefix(property, "leveldb.") {
		property = "leveldb." + property
	}
	return ldb.LDB().GetProperty(property)
}

func (api *PrivateDebugAPI) ChaindbCompact() error {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return fmt.Errorf("chaindbCompact does not work for memory databases")
	}
	for b := byte(0); b < 255; b++ {
		log.Info("Compacting chain database", "range", fmt.Sprintf("0x%0.2X-0x%0.2X", b, b+1))
		err := ldb.LDB().CompactRange(util.Range{Start: []byte{b}, Limit: []byte{b + 1}})
		if err != nil {
			log.Error("Database compaction failed", "err", err)
			return err
		}
	}
	return nil
}

// SetHead rewinds the head of the blockchain to a previous block.
func (api *PrivateDebugAPI) SetHead(number hexutil.Uint64) {
	api.b.SetHead(uint64(number))
}

// PublicNetAPI offers network related RPC methods
type PublicNetAPI struct {
	net            *p2p.Server
	networkVersion uint64
}

// NewPublicNetAPI creates a new net API instance.
func NewPublicNetAPI(net *p2p.Server, networkVersion uint64) *PublicNetAPI {
	return &PublicNetAPI{net, networkVersion}
}

// Listening returns an indication if the node is listening for network connections.
func (s *PublicNetAPI) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
func (s *PublicNetAPI) PeerCount() hexutil.Uint {
	return hexutil.Uint(s.net.PeerCount())
}

// Version returns the current matrix protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}
