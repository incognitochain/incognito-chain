package config

import "github.com/incognitochain/incognito-chain/common"

//Env variables key
const (
	NetworkKey          = "INCOGNITO_NETWORK_KEY"
	NetworkVersionKey   = "INCOGNITO_NETWORK_VERSION_KEY"
	ConfigFileKey       = "INCOGNITO_CONFIG_FILE_KEY"
	ConfigDirKey        = "INCOGNITO_CONFIG_DIR_KEY"
	ConfigFileTypeKey   = "INCOGNITO_CONFIG_FILE_TYPE_KEY"
	ParamFileKey        = "INCOGNITO_PARAM_FILE_KEY"
	InitTxFileKey       = "INCOGNITO_INIT_TX_FILE_KEY"
	UnifiedTokenFileKey = "INCOGNITO_UNIFIED_TOKEN_FILE_KEY"
	GethHostKey         = "GETH_NAME"
	GethPortKey         = "GETH_PORT"
	GethProtocolKey     = "GETH_PROTOCOL"
	BSCHostKey          = "BSC_HOST"
	PLGHostKey          = "PLG_HOST"
	FTMHostKey          = "FTM_HOST"
)

// default config
const (
	DefaultDataDirname                 = "data"
	DefaultDatabaseDirname             = "block"
	DefaultMempoolDirname              = "mempool"
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
	DefaultWalletName           = "wallet"
	DefaultPersistMempool       = false
	DefaultBtcClient            = 0
	DefaultBtcClientPort        = "8332"
	DefaultNetwork              = LocalNetwork
	DefaultConfigDir            = "config"
	DefaultConfigFile           = "config"
	DefaultConfigFileType       = "yaml"
	DefaultParamFile            = "param"
	DefaultInitTxFile           = "init_tx"
	DefaultInitTxFileType       = "json"
	DefaultUnifiedTokenFile     = "unified_token"
	DefaultUnifiedTokenFileType = "json"
)

const (
	LocalNetwork          = "local"
	LocalDCSNetwork       = "local-dcs"
	TestNetNetwork        = "testnet"
	MainnetNetwork        = "mainnet"
	TestNetVersion1       = "1"
	TestNetVersion2       = "2"
	TestNetVersion1Number = 1
	TestNetVersion2Number = 2
	DefaultPort           = "9444"
	DefaultRPCPort        = "9344"
	DefaultWSPort         = "19444"
	LocalNet              = 0x02
	Testnet2Net           = 0x32
	TestnetNet            = 0x16
	MainnetNet            = 0x01
	KeyListFileName       = "keylist.json"
	KeyListV2FileName     = "keylist-v2.json"
	DefaultOutcoinDirname = "_coins_"
	DefaultNumCIWorkers   = 0
)

var (
	defaultDataDir     = DefaultDataDirname
	defaultRPCKeyFile  = "rpc.key"
	defaultRPCCertFile = "rpc.cert"
	defaultLogDir      = DefaultLogDirname
)

var (
	configCache4GB = batchCommitSyncModeParam{
		TrieJournalCacheSize: 32,
		BlockTrieInMemory:    uint64(500),
		TrieNodeLimit:        common.StorageSize(128 * 1024 * 1024),
		TrieImgsLimit:        common.StorageSize(4 * 1024 * 1024),
	}
	configCache8GB = batchCommitSyncModeParam{
		TrieJournalCacheSize: 32,
		BlockTrieInMemory:    uint64(2000),
		TrieNodeLimit:        common.StorageSize(512 * 1024 * 1024),
		TrieImgsLimit:        common.StorageSize(4 * 1024 * 1024),
	}
	configCache16GB = batchCommitSyncModeParam{
		TrieJournalCacheSize: 32,
		BlockTrieInMemory:    uint64(10000),
		TrieNodeLimit:        common.StorageSize(2 * 1024 * 1024 * 1024),
		TrieImgsLimit:        common.StorageSize(4 * 1024 * 1024),
	}
	configCache32GB = batchCommitSyncModeParam{
		TrieJournalCacheSize: 32,
		BlockTrieInMemory:    uint64(20000),
		TrieNodeLimit:        common.StorageSize(4 * 1024 * 1024 * 1024),
		TrieImgsLimit:        common.StorageSize(4 * 1024 * 1024),
	}
)
