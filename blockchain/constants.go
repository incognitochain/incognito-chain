package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/metrics"
)

//Network fixed params
const (
	// SHARD_BLOCK_VERSION is the current latest supported block version.
	VERSION                       = 1
	RANDOM_NUMBER                 = 3
	SHARD_BLOCK_VERSION           = 1
	DefaultMaxBlkReqPerPeer       = 900
	MinCommitteeSize              = 3 // min size to run bft
	WorkerNumber                  = 5
	MAX_S2B_BLOCK                 = 30
	MAX_BEACON_BLOCK              = 5
	LowerBoundPercentForIncDAO    = 3
	UpperBoundPercentForIncDAO    = 10
	TestRandom                    = true
	ValidateTimeForSpamRequestTxs = 1581565837 // GMT: Thursday, February 13, 2020 3:50:37 AM. From this time, block will be checked spam request-reward tx
	TransactionBatchSize          = 30
	SpareTime                     = 1000             // in mili-second
	DefaultMaxBlockSyncTime       = 30 * time.Second // in second
)

// burning addresses
const (
	burningAddress  = "15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs"
	burningAddress2 = "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
)

// CONSTANT for network MAINNET
const (
	// ------------- Mainnet ---------------------------------------------
	Mainnet                  = 0x01
	MainetName               = "mainnet"
	MainnetDefaultPort       = "9333"
	MainnetGenesisBlockTime  = "2019-10-29T00:00:00.000Z"
	MainnetEpoch             = 350
	MainnetRandomTime        = 175
	MainnetEpochV2BreakPoint = 10e9
	MainnetEpochV2           = 350
	MainnetRandomTimeV2      = 175
	MainnetOffset            = 4
	MainnetSwapOffset        = 4
	MainnetAssignOffset      = 8
	MainnetMaxSwapOrAssign   = 10

	MainNetShardCommitteeSize     = 32
	MainNetMinShardCommitteeSize  = 22
	MainNetBeaconCommitteeSize    = 32
	MainNetMinBeaconCommitteeSize = 7
	MainNetActiveShards           = 8
	MainNetStakingAmountShard     = 1750000000000 // 1750 PRV = 1750 * 10^9 nano PRV

	MainnetMinBeaconBlkInterval = 40 * time.Second //second
	MainnetMaxBeaconBlkCreation = 10 * time.Second //second
	MainnetMinShardBlkInterval  = 40 * time.Second //second
	MainnetMaxShardBlkCreation  = 10 * time.Second //second

	//board and proposal parameters
	MainnetBasicReward = 1386666000 //1.386666 PRV
	//MainETHContractAddressStr = "0x0261DB5AfF8E5eC99fBc8FBBA5D4B9f8EcD44ec7" // v2-main - mainnet, branch master-temp-B-deploy, support erc20 with decimals > 18
	//MainETHContractAddressStr               = "0x3c8ec94213f09A1575f773470830124dfb40042e"                                                              // v3-main - mainnet
	//MainETHContractAddressStr               = "0x6CC3873C3ca91cf5500DaD8B1A2c620B4f20507c"                                                              // v4-main - mainnet
	//MainETHContractAddressStr               = "0xED5309daac912a52d985c317576a1b3f5020FDc9"                                                              // v5-main - mainnet
	MainETHContractAddressStr               = "0x97875355eF55Ae35613029df8B1C8Cf8f89c9066"                                                              // v6-main - mainnet
	MainnetIncognitoDAOAddress              = "12S32fSyF4h8VxFHt4HfHvU1m9KHvBQsab5zp4TpQctmMdWuveXFH9KYWNemo7DRKvaBEvMgqm4XAuq1a1R4cNk2kfUfvXR3DdxCho3" // community fund
	MainnetCentralizedWebsitePaymentAddress = "12Rvjw6J3FWY3YZ1eDZ5uTy6DTPjFeLhCK7SXgppjivg9ShX2RRq3s8pdoapnH8AMoqvUSqZm1Gqzw7rrKsNzRJwSK2kWbWf1ogy885"

	// relaying header chain
	MainnetBNBChainID        = "Binance-Chain-Tigris"
	MainnetBTCChainID        = "Bitcoin-Mainnet"
	MainnetBTCDataFolderName = "btcrelayingv7"

	// BNB fullnode
	MainnetBNBFullNodeHost     = "dataseed1.ninicoin.io"
	MainnetBNBFullNodeProtocol = "https"
	MainnetBNBFullNodePort     = "443"

	MainnetPortalFeeder = "12RwJVcDx4SM4PvjwwPrCRPZMMRT9g6QrnQUHD54EbtDb6AQbe26ciV6JXKyt4WRuFQVqLKqUUbb7VbWxR5V6KaG9HyFbKf6CrRxhSm"

	// Enable Feature Flag
	MainnetEnablePortalRelaying = 1
	MainnetEnablePortalV3       = 0
	// ------------- end Mainnet --------------------------------------
)

const (
	TestnetBNBChainID         = "Binance-Chain-Ganges"
	TestnetBTCChainID         = "Bitcoin-Testnet"
	Testnet2BNBChainID        = "Binance-Chain-Ganges"
	Testnet2BTCChainID        = "Bitcoin-Testnet-2"
	Testnet2BTCDataFolderName = "btcrelayingv11"
)

var (
	shardInsertBlockTimer                  = metrics.NewRegisteredTimer("shard/insert", nil)
	shardVerifyPreprocesingTimer           = metrics.NewRegisteredTimer("shard/verify/preprocessing", nil)
	shardVerifyPreprocesingForPreSignTimer = metrics.NewRegisteredTimer("shard/verify/preprocessingpresign", nil)
	shardVerifyWithBestStateTimer          = metrics.NewRegisteredTimer("shard/verify/withbeststate", nil)
	shardVerifyPostProcessingTimer         = metrics.NewRegisteredTimer("shard/verify/postprocessing", nil)
	shardStoreBlockTimer                   = metrics.NewRegisteredTimer("shard/storeblock", nil)
	shardUpdateBestStateTimer              = metrics.NewRegisteredTimer("shard/updatebeststate", nil)

	beaconInsertBlockTimer                  = metrics.NewRegisteredTimer("beacon/insert", nil)
	beaconVerifyPreprocesingTimer           = metrics.NewRegisteredTimer("beacon/verify/preprocessing", nil)
	beaconVerifyPreprocesingForPreSignTimer = metrics.NewRegisteredTimer("beacon/verify/preprocessingpresign", nil)
	beaconVerifyWithBestStateTimer          = metrics.NewRegisteredTimer("beacon/verify/withbeststate", nil)
	beaconVerifyPostProcessingTimer         = metrics.NewRegisteredTimer("beacon/verify/postprocessing", nil)
	beaconStoreBlockTimer                   = metrics.NewRegisteredTimer("beacon/storeblock", nil)
	beaconUpdateBestStateTimer              = metrics.NewRegisteredTimer("beacon/updatebeststate", nil)
)

const (
	Duration = 1000000
)
