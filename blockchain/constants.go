package blockchain

import "time"

// constant for network
const (
	//Network fixed params
	ThresholdRatioOfDCBCrisis = 9000
	ThresholdRatioOfGOVCrisis = 9000

	// ------------- Mainnet ---------------------------------------------
	Mainnet            = 0x01
	MainetName         = "mainnet"
	MainnetDefaultPort = "9333"

	MainNetShardsNum           = 4
	MainNetShardCommitteeSize  = 1
	MainNetBeaconCommitteeSize = 1
	MainNetActiveShards        = 2

	//board and proposal parameters
	MainnetInitFundSalary             = 1000000000000000
	MainnetInitDCBToken               = 10000
	MainnetInitGovToken               = 10000
	MainnetInitCmBToken               = 10000
	MainnetInitBondToken              = 10000
	MainnetGenesisblockPaymentAddress = "1UuyYcHgVFLMd8Qy7T1ZWRmfFvaEgogF7cEsqY98ubQjoQUy4VozTqyfSNjkjhjR85C6GKBmw1JKekgMwCeHtHex25XSKwzb9QPQ2g6a3"
	// ------------- end Mainnet --------------------------------------

	// ------------- Testnet ---------------------------------------------
	Testnet            = 0x02
	TestnetName        = "testnet"
	TestnetDefaultPort = "9444"

	TestNetShardsNum           = 4
	TestNetShardCommitteeSize  = 1
	TestNetBeaconCommitteeSize = 1
	TestNetActiveShards        = 2

	//board and proposal parameters
	TestnetInitFundSalary             = 1000000000000000
	TestnetInitDCBToken               = 10000
	TestnetInitGovToken               = 10000
	TestnetInitCmBToken               = 10000
	TestnetInitBondToken              = 10000
	TestnetGenesisBlockPaymentAddress = "1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba"
	// ------------- end Testnet --------------------------------------
)

const (
	// BlockVersion is the current latest supported block version.
	BlockVersion                = 1
	TransactionVersion          = 1
	defaultMaxBlkReqPerPeer     = 60
	defaultBroadcastStateTime   = 2 * time.Second  // in second
	defaultProcessPeerStateTime = 2 * time.Second  // in second
	defaultMaxBlockSyncTime     = 2 * time.Second  // in second
	defaultCacheCleanupTime     = 60 * time.Second // in second
)
