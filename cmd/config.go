package main

import (
	"github.com/jessevdk/go-flags"
	"path/filepath"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/blockchain"
)

const (
	defaultConfigFilename  = "config.conf"
	defaultDataDirname     = "data"
	defaultDatabaseDirname = "block"
	defaultLogLevel        = "info"
	defaultLogDirname      = "logs"
	defaultLogFilename     = "log.log"
	defaultMaxPeers        = 125
	defaultMaxRPCClients   = 10
	defaultGenerate        = false
	sampleConfigFilename   = "sample-config.conf"
	defaultDisableRpcTls   = true
	defaultFastMode        = false
	// For wallet
	defaultWalletName = "wallet"
)

var (
	defaultHomeDir     = common.AppDataDir("prototype", false)
	defaultConfigFile  = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultDataDir     = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultRPCKeyFile  = filepath.Join(defaultHomeDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(defaultHomeDir, "rpc.cert")
	defaultLogDir      = filepath.Join(defaultHomeDir, defaultLogDirname)
)

// See loadConfig for details on the configuration load process.
type config struct {
	Command string `long:"cmd" short:"c" description:"Command name"`
	DataDir string `short:"b" long:"datadir" description:"Directory to store data"`
	TestNet bool   `long:"testnet" description:"Use the test network"`

	// For Wallet
	WalletName       string `long:"wallet" description:"Wallet Database Name file, default is 'wallet'"`
	WalletPassphrase string `long:"walletpassphrase" description:"Wallet passphrase"`
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *config, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	return parser
}

func loadConfig() (*config, error) {
	cfg := config{
		DataDir: defaultDataDir,
		TestNet: false,
	}

	preParser := newConfigParser(&cfg, flags.HelpFlag)
	preParser.Parse()
	cfg.DataDir = common.CleanAndExpandPath(cfg.DataDir, defaultHomeDir)
	if cfg.TestNet {
		cfg.DataDir = filepath.Join(cfg.DataDir, blockchain.TestNetParams.Name)
	} else {
		cfg.DataDir = filepath.Join(cfg.DataDir, blockchain.MainNetParams.Name)
	}

	return &cfg, nil
}
