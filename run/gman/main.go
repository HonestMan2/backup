// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
// Copyright 2014 The go-matrix Authors
// This file is part of go-matrix.
// go-matrix is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// go-matrix is distributed in the hope that it will be useful,
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with go-matrix. If not, see <http://www.gnu.org/licenses/>.

// gman is the official command-line client for Matrix.
package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/accounts/keystore"
	"github.com/matrix/go-matrix/run/utils"
	"github.com/matrix/go-matrix/console"
	"github.com/matrix/go-matrix/man"
	"github.com/matrix/go-matrix/manclient"
	"github.com/matrix/go-matrix/internal/debug"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/metrics"
	"github.com/matrix/go-matrix/pod"
	"github.com/matrix/go-matrix/p2p"
	"github.com/matrix/go-matrix/params"
)

const (
	clientIdentifier = "gman" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, "the go-matrix command line interface")
	// flags that configure the node
	nodeFlags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
		utils.NoUSBFlag,
		utils.DashboardEnabledFlag,
		utils.DashboardAddrFlag,
		utils.DashboardPortFlag,
		utils.DashboardRefreshFlag,
		utils.EthashCacheDirFlag,
		utils.EthashCachesInMemoryFlag,
		utils.EthashCachesOnDiskFlag,
		utils.EthashDatasetDirFlag,
		utils.EthashDatasetsInMemoryFlag,
		utils.EthashDatasetsOnDiskFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolAccountSlotsFlag,
		utils.TxPoolGlobalSlotsFlag,
		utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
		utils.FastSyncFlag,
		utils.LightModeFlag,
		utils.SyncModeFlag,
		utils.GCModeFlag,
		utils.LightServFlag,
		utils.LightPeersFlag,
		utils.LightKDFFlag,
		utils.CacheFlag,
		utils.CacheDatabaseFlag,
		utils.CacheGCFlag,
		utils.TrieCacheGenFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.EtherbaseFlag,
		utils.GasPriceFlag,
		utils.MinerThreadsFlag,
		utils.MiningEnabledFlag,
		utils.TargetGasLimitFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.DeveloperFlag,
		utils.DeveloperPeriodFlag,
		utils.TestnetFlag,
		utils.RinkebyFlag,
		utils.VMEnableDebugFlag,
		utils.NetworkIdFlag,
		utils.RPCCORSDomainFlag,
		utils.RPCVirtualHostsFlag,
		utils.EthStatsURLFlag,
		utils.MetricsEnabledFlag,
		utils.FakePoWFlag,
		utils.NoCompactionFlag,
		utils.GpoBlocksFlag,
		utils.GpoPercentileFlag,
		utils.ExtraDataFlag,
		configFileFlag,
		utils.TestLocalMiningFlag,
		utils.TestHeaderGenFlag,
		utils.TestChangeRoleFlag,
	}

	rpcFlags = []cli.Flag{
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCApiFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
	}

)

func init() {
	// Initialize the CLI app and start Gman
	app.Action = gman
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2013-2018 The go-matrix Authors"
	app.Commands = []cli.Command{
		// See chaincmd.go:
		initCommand,
		importCommand,
		exportCommand,
		importPreimagesCommand,
		exportPreimagesCommand,
		copydbCommand,
		removedbCommand,
		dumpCommand,
		// See monitorcmd.go:
		monitorCommand,
		// See accountcmd.go:
		accountCommand,
		walletCommand,
		// See consolecmd.go:
		consoleCommand,
		attachCommand,
		javascriptCommand,
		// See misccmd.go:
		makecacheCommand,
		makedagCommand,
		versionCommand,
		bugCommand,
		licenseCommand,
		// See config.go
		dumpConfigCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, consoleFlags...)
	app.Flags = append(app.Flags, debug.Flags...)

	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		logdir := "debuglog"
		/*	if ctx.GlobalBool(utils.DashboardEnabledFlag.Name) {
			logdir = (&node.Config{DataDir: utils.MakeDataDir(ctx)}).ResolvePath("logs")
		logdir = "debuglog"
		if err := debug.Setup(ctx, logdir); err != nil {
			return err
		}

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(3 * time.Second)

		utils.SetupNetwork(ctx)
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		console.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

// gman is the main entry point into the system if no sprguments and runs it in
// blocking mode, waiting for it to be shut down.ecial subcommand is ran.
// It creates a default node based on the command line a
func gman(ctx *cli.Context) error {
	node := makeFullNode(ctx)
	startNode(ctx, node)
	node.Wait()
	return nil
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// miner.
func startNode(ctx *cli.Context, stack *pod.Node) {
	debug.Memsize.Add("node", stack)

	// Start up the node itself
	utils.StartNode(stack)

	// Unlock any account specifically requested
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	passwords := utils.MakePasswordList(ctx)
	unlocks := strings.Split(ctx.GlobalString(utils.UnlockedAccountFlag.Name), ",")
	for i, account := range unlocks {
		if trimmed := strings.TrimSpace(account); trimmed != "" {
			unlockAccount(ctx, ks, trimmed, i, passwords)
		}
	}

	signHelper := stack.SignalHelper()
	wallets := stack.AccountManager().Wallets()

	if len(wallets) > 0 && len(wallets[0].Accounts()) > 0 && len(passwords) > 0 {
		signHelper.SetAccountManager(stack.AccountManager(), wallets[0].Accounts()[0].Address, passwords[0])
	}
	if len(wallets) <= 0 {
		log.Error("无钱包", "请新建钱包", "")
	}
	if len(wallets) > 0 && len(wallets[0].Accounts()) <= 0 {
		log.Error("钱包无账户", "请新建账户", "")
	}
	if len(passwords) <= 0 {
		log.Error("password无密码", "请重启时输入密码", "")
	}

	// Register wallet event handlers to open and auto-derive wallets
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	go func() {
		// Create a chain state reader for self-derivation
		rpcClient, err := stack.Attach()
		if err != nil {
			utils.Fatalf("Failed to attach to self: %v", err)
		}
		stateReader := manclient.NewClient(rpcClient)

		// Open any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			switch event.Kind {
			case accounts.WalletArrived:
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				}
			case accounts.WalletOpened:
				status, _ := event.Wallet.Status()
				log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", status)

				if event.Wallet.URL().Scheme == "ledger" {
					event.Wallet.SelfDerive(accounts.DefaultLedgerBaseDerivationPath, stateReader)
				} else {
					event.Wallet.SelfDerive(accounts.DefaultBaseDerivationPath, stateReader)
				}

			case accounts.WalletDropped:
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()

	var matrix *man.Matrix
	if err := stack.Service(&matrix); err != nil {
		utils.Fatalf("Matrix service not running :%v", err)
	}
	log.INFO("MainBootNode", "data", params.MainnetBootnodes)
	log.INFO("BoradCastNode", "data", params.BroadCastNodes)
	log.Info("main", "nodeid", stack.Server().Self().ID.String())

	go func() {
		time.Sleep(3 * time.Second)
		mc.PublishEvent(mc.NewBlockMessage, matrix.BlockChain().GetBlockByNumber(0))
		log.INFO("MAIN", "创世区块插入消息已发送", matrix.BlockChain().GetBlockByNumber(0))
		log.INFO("Peer总量", "len", p2p.ServerP2p.PeerCount())

	}()

	log.INFO("创世文件选举信息", "data", matrix.BlockChain().GetBlockByNumber(0).Header().Elect)
	log.INFO("创世文件拓扑图", "data", matrix.BlockChain().GetBlockByNumber(0).Header().NetTopology)
	/*go func() {
		time.Sleep(10 * time.Second)
		valDep, err := depoistInfo.GetDepositList(big.NewInt(0), common.RoleValidator)
		log.INFO("验证者参选信息", "data", valDep, "err", err)
		mindep, err := depoistInfo.GetDepositList(big.NewInt(0), common.RoleMiner)
		log.INFO("矿工参选信息", "data", mindep, "err", err)

	// Start auxiliary services if enabled
	if ctx.GlobalBool(utils.MiningEnabledFlag.Name) || ctx.GlobalBool(utils.DeveloperFlag.Name) {
		// Mining only makes sense if a full Matrix node is running
		if ctx.GlobalBool(utils.LightModeFlag.Name) || ctx.GlobalString(utils.SyncModeFlag.Name) == "light" {
			utils.Fatalf("Light clients do not support mining")
		}
		var matrix *man.Matrix
		if err := stack.Service(&matrix); err != nil {
			utils.Fatalf("Matrix service not running: %v", err)
		}
		// Use a reduced number of threads if requested
		if threads := ctx.GlobalInt(utils.MinerThreadsFlag.Name); threads > 0 {
			type threaded interface {
				SetThreads(threads int)
			}
			if th, ok := matrix.Engine().(threaded); ok {
				th.SetThreads(threads)
			}
		}
		// Set the gas price to the limits from the CLI and start mining
		matrix.TxPool().SetGasPrice(utils.GlobalBig(ctx, utils.GasPriceFlag.Name))
		if err := matrix.StartMining(true); err != nil {
			utils.Fatalf("Failed to start mining: %v", err)

		}
	}
}
func Init_Config_PATH(ctx *cli.Context) {
	log.INFO("开始读取配置文件", "", "")
	config_dir := utils.MakeDataDir(ctx)
	if config_dir == "" {
		log.Error("无创世文件", "请在启动时使用--datadir", "")
	}

	params.Config_Init(config_dir + "/man.json")
}
