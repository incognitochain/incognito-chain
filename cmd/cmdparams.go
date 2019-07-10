package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xsirrush/color"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/jessevdk/go-flags"
)

const (
	defaultConfigFilename = "component.conf"
	defaultDataDirname    = "data"
	defaultLogDirname     = "logs"
)

var (
	defaultHomeDir    = common.AppDataDir("cash", false)
	defaultConfigFile = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultDataDir    = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultLogDir     = filepath.Join(defaultHomeDir, defaultLogDirname)
)

// See loadParams for details on the configuration load process.
type params struct {
	Command string `long:"cmd" short:"c" description:"Command name"`
	DataDir string `short:"b" long:"datadir" description:"Directory to store data"`
	TestNet bool   `long:"testnet" description:"Use the test network"`

	// Chain
	Beacon bool `long:"beacon" description:"Process Beacon Chain"`
	// shardIDs:
	// "all": process all shards
	// 1,2,3,4: shard 1, shard 2, shard 3, shard 4
	ShardIDs     string `long:"shardids" description:"Process one or many Shard Chain with ShardID"`
	ChainDataDir string `long:"chaindatadir" description:"Directory of Stored Blockchain Database"`
	OutDataDir   string `long:"outdatadir" description:"Directory of Export Blockchain Data"`
	FileName     string `long:"filename" description:"Filename of Backup Blockchin Data"`
	// wallet
	WalletName        string `long:"wallet" description:"Wallet Database Name file, default is 'wallet'"`
	WalletPassphrase  string `long:"walletpassphrase" description:"Wallet passphrase"`
	WalletAccountName string `long:"walletaccountname" description:"Wallet account name"`
	ShardID           int8   `long:"shardid" description:"Process Shard Chain with ShardID"`

	// pToken
	PNetwork string `long:"pNetwork" description:"Bridge network"`
	PToken   string `long:"pToken" description:"Bridge token"`
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *params, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	return parser
}

func loadParams() (*params, error) {
	cfg := params{
		DataDir: defaultDataDir,
		TestNet: false,
	}

	preParser := newConfigParser(&cfg, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			red := color.New(color.FgRed).SprintFunc()
			fmt.Println(red("---------------------------------------"))
			fmt.Printf("List cmd: %+v \n", red(CmdList))
			fmt.Println(red("---------------------------------------"))
			return nil, err
		}
	}
	cfg.DataDir = common.CleanAndExpandPath(cfg.DataDir, defaultHomeDir)
	if cfg.TestNet {
		cfg.DataDir = filepath.Join(cfg.DataDir, blockchain.ChainTestParam.Name)
	} else {
		cfg.DataDir = filepath.Join(cfg.DataDir, blockchain.ChainMainParam.Name)
	}

	return &cfg, nil
}
