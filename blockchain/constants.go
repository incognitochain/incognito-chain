package blockchain

import "time"

// constant for network
const (
	//Network fixed params
	ThresholdRatioOfDCBCrisis = 9000
	ThresholdRatioOfGOVCrisis = 9000

	// Mainnet
	Mainnet                           = 0x01
	MainetName                        = "mainnet"
	MainnetDefaultPort                = "9333"
	MainnetInitFundSalary             = 0
	MainnetInitDCBToken               = 0
	MainnetInitGovToken               = 0
	MainnetInitCmBToken               = 0
	MainnetInitBondToken              = 0
	MainnetGenesisblockPaymentAddress = "1UuyYcHgVFLMd8Qy7T1ZWRmfFvaEgogF7cEsqY98ubQjoQUy4VozTqyfSNjkjhjR85C6GKBmw1JKekgMwCeHtHex25XSKwzb9QPQ2g6a3"

	// Testnet
	Testnet               = 0x02
	TestnetName           = "testnet"
	TestnetDefaultPort    = "9444"
	TestnetInitFundSalary = 1000000000000000
	TestnetInitDCBToken   = 10000
	TestnetInitGovToken   = 10000

	//board and proposal parameters
	TestnetInitCmBToken               = 10000
	TestnetInitBondToken              = 10000
	TestnetGenesisBlockPaymentAddress = "1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba"
)

const (
	// BlockVersion is the current latest supported block version.
	BlockVersion                = 1
	TransactionVersion          = 1
	defaultMaxBlkReqPerPeer     = 40
	defaultBroadcastStateTime   = 2 * time.Second  // in second
	defaultProcessPeerStateTime = 2 * time.Second  // in second
	defaultMaxBlockSyncTime     = 2 * time.Second  // in second
	defaultCacheCleanupTime     = 60 * time.Second // in second
)
