package config

import (
	"path/filepath"

	"github.com/incognitochain/incognito-chain/common"
)

// default config
const (
	DefaultConfigFilename              = "config.conf"
	DefaultDataDirname                 = "data"
	DefaultDatabaseDirname             = "block"
	DefaultDatabaseMempoolDirname      = "mempool"
	DefaultLogLevel                    = "info"
	DefaultLogDirname                  = "logs"
	DefaultLogFilename                 = "log.log"
	DefaultMaxPeers                    = 1000
	DefaultMaxPeersSameShard           = 300
	DefaultMaxPeersOtherShard          = 600
	DefaultMaxPeersOther               = 300
	DefaultMaxPeersNoShard             = 200
	DefaultMaxPeersBeacon              = 500
	DefaultMaxRPCClients               = 500
	DefaultRPCLimitRequestPerDay       = 0 // 0: unlimited
	DefaultRPCLimitErrorRequestPerHour = 0 // 0: unlimited
	DefaultMaxRPCWsClients             = 200
	DefaultMetricUrl                   = ""
	SampleConfigFilename               = "sample-config.conf"
	DefaultDisableRpcTLS               = true
	DefaultFastStartup                 = true
	// DefaultNodeMode                    = common.NodeModeRelay
	DefaultEnableMining = true
	DefaultTxPoolTTL    = uint(15 * 60) // 15 minutes
	DefaultTxPoolMaxTx  = uint64(100000)
	DefaultLimitFee     = uint64(1) // 1 nano PRV = 10^-9 PRV
	//DefaultLimitFee = uint64(100000) // 100000 nano PRV = 100000 * 10^-9 PRV
	// For wallet
	DefaultWalletName     = "wallet"
	DefaultPersistMempool = false
	DefaultBtcClient      = 0
	DefaultBtcClientPort  = "8332"
)

var (
	defaultHomeDir     = common.AppDataDir("incognito", false)
	defaultConfigFile  = filepath.Join(defaultHomeDir, DefaultConfigFilename)
	defaultDataDir     = filepath.Join(defaultHomeDir, DefaultDataDirname)
	defaultRPCKeyFile  = filepath.Join(defaultHomeDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(defaultHomeDir, "rpc.cert")
	defaultLogDir      = filepath.Join(defaultHomeDir, DefaultLogDirname)
)
