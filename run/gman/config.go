// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"unicode"

	"encoding/base64"
	"encoding/json"
	"github.com/matrix/go-matrix/crypto/aes"
	"github.com/matrix/go-matrix/dashboard"
	"github.com/matrix/go-matrix/console"
	"github.com/matrix/go-matrix/man"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/pod"
	"github.com/matrix/go-matrix/run/utils"
	"github.com/naoina/toml"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/base58"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, rpcFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type manstatsConfig struct {
	URL string `toml:",omitempty"`
}

type gmanConfig struct {
	Man       man.Config
	Node      pod.Config
	Manstats  manstatsConfig
	Dashboard dashboard.Config
}

func loadConfig(file string, cfg *gmanConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() pod.Config {
	cfg := pod.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "man", "eth", "shh")
	cfg.WSModules = append(cfg.WSModules, "man", "eth", "shh")
	cfg.IPCPath = "gman.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*pod.Node, gmanConfig) {
	// Load defaults.
	cfg := gmanConfig{
		Man:       man.DefaultConfig,
		Node:      defaultNodeConfig(),
		Dashboard: dashboard.DefaultConfig,
	}

	// Load config file.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := pod.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetManConfig(ctx, stack, &cfg.Man)
	if ctx.GlobalIsSet(utils.ManStatsURLFlag.Name) {
		cfg.Manstats.URL = ctx.GlobalString(utils.ManStatsURLFlag.Name)
	}

	utils.SetDashboardConfig(ctx, &cfg.Dashboard)

	return stack, cfg
}

func makeFullNode(ctx *cli.Context) *pod.Node {
	Init_Config_PATH(ctx)
	stack, cfg := makeConfigNode(ctx)
	err := CheckEntrust(ctx)
	if err != nil {
		fmt.Println("检查委托交易失败 err", err)
		os.Exit(1)
	}
	utils.RegisterManService(stack, &cfg.Man)

	if ctx.GlobalBool(utils.DashboardEnabledFlag.Name) {
		utils.RegisterDashboardService(stack, &cfg.Dashboard, gitCommit)
	}
	// Add the Matrix Stats daemon if requested.
	if cfg.Manstats.URL != "" {
		utils.RegisterManStatsService(stack, cfg.Manstats.URL)
	}
	return stack
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.Man.Genesis != nil {
		cfg.Man.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}
	io.WriteString(os.Stdout, comment)
	os.Stdout.Write(out)
	return nil
}
func CheckEntrust(ctx *cli.Context) error {
	path := ctx.GlobalString(utils.AccountPasswordFileFlag.Name)
	if path == "" {
		return nil
	}

	password, err := ReadDecryptPassword(ctx)
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("文件失败", err, "path", path)
		return err
	}

	b, err := ioutil.ReadAll(f)
	bytesPass, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		fmt.Println("解密失败", err)
		return err
	}
	tpass, err := aes.AesDecrypt(bytesPass, []byte(password))
	if err != nil {
		fmt.Println("AedDecrypt失败")
		return err
	}

	var anss []mc.EntrustInfo
	err = json.Unmarshal(tpass, &anss)
	if err != nil {
		fmt.Println("加密文件解码失败 密码不正确")
		return err
	}

	entrustValue :=make(map[common.Address]string,0)
	for _, v := range anss {
		entrustValue[base58.Base58DecodeToAddress(v.Address)] = v.Password
	}
	manparams.EntrustAccountValue.SetEntrustValue(entrustValue)
	return nil
}

func ReadDecryptPassword(ctx *cli.Context) (string, error) {
	if password := ctx.GlobalString(utils.TestEntrustFlag.Name); password != "" {
		return password, nil
	}
	var passphrase string
	var err error
	InputCount:=0

	for true {
		InputCount++
		if InputCount>3{
			return "",errors.New("多次输入密码错误")
		}
		fmt.Printf("第 %d次密码输入 \n",InputCount)
		passphrase,err=GetPassword()
		if err!=nil{
			fmt.Println("获取密码错误",err)
			continue
		}
		if CheckPassword(passphrase) {
			break
		}
	}
	return passphrase, nil
}

func GetPassword()(string,error){
	password, err := console.Stdin.PromptPassword("Passphrase: ")
	if err != nil {
		return "", fmt.Errorf("Failed to read passphrase: %v", err)
	}
	confirm, err := console.Stdin.PromptPassword("Repeat passphrase: ")
	if err != nil {
		return "",fmt.Errorf("Failed to read passphrase confirmation: %v", err)
	}
	if password != confirm {
		return "",fmt.Errorf("Passphrases do not match")
	}
	return password,nil
}