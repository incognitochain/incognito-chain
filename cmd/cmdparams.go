package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"os"
	"path/filepath"
)

const (
	defaultConfigFilename = "params.conf"
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

	// For Wallet
	WalletName        string `long:"wallet" description:"Wallet Database Name file, default is 'wallet'"`
	WalletPassphrase  string `long:"walletpassphrase" description:"Wallet passphrase"`
	WalletAccountName string `long:"walletaccountname" description:"Wallet account name"`
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
		cfg.DataDir = filepath.Join(cfg.DataDir, blockchain.ChainParam.Name)
	} else {
		cfg.DataDir = filepath.Join(cfg.DataDir, blockchain.MainNetParams.Name)
	}

	return &cfg, nil
}
