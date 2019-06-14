package blockchain

import (
	"time"
)

//Network fixed params
const (
	// BlockVersion is the current latest supported block version.
	BlockVersion            = 1
	defaultMaxBlkReqPerPeer = 600
	defaultMaxBlkReqPerTime = 1200

	defaultBroadcastStateTime = 2 * time.Second  // in second
	defaultStateUpdateTime    = 3 * time.Second  // in second
	defaultMaxBlockSyncTime   = 1 * time.Second  // in second
	defaultCacheCleanupTime   = 30 * time.Second // in second
	workerNum                 = 5
)

// CONSTANT for network MAINNET
const (
	// ------------- Mainnet ---------------------------------------------
	Mainnet            = 0x01
	MainetName         = "mainnet"
	MainnetDefaultPort = "9333"

	MainNetShardCommitteeSize  = 3
	MainNetBeaconCommitteeSize = 3
	MainNetActiveShards        = 2
	MainNetStakingAmountShard  = 175000

	//board and proposal parameters
	MainnetBasicReward                = 50      //50 mili PRV
	MainnetRewardHalflife             = 6307200 //1 year, reduce 10% per year
	MainnetGenesisblockPaymentAddress = "1Uv2zzR4LgfX8ToQe8ub3bYcCLk3uDU1sm9U9hiu9EKYXoS77UdikfT9s8d5YjhsTJm61eazsMwk2otFZBYpPHwiMn8z6bKWWJRspsLky"
	// ------------- end Mainnet --------------------------------------
)

// VARIABLE for mainnet
var (
	MainnetInitConstant = []string{}
	// for beacon
	// public key
	PreSelectBeaconNodeMainnetSerializedPubkey = PreSelectBeaconNodeTestnetSerializedPubkey
	// For shard
	// public key
	PreSelectShardNodeMainnetSerializedPubkey = PreSelectShardNodeTestnetSerializedPubkey
	MaxTxsInBlock                             = 600
	MaxTxsProcessTimeInBlockCreation          = float64(0.85)
	TxsAverageProcessTime                     = int64(5000) // count in nano second ~= 5 mili seconds
	DefaultTxsAverageProcessTime              = int64(5000) // count in nano second
)

// END CONSTANT for network MAINNET

// CONSTANT for network TESTNET
const (
	Testnet            = 0x255
	TestnetName        = "testnet"
	TestnetDefaultPort = "9444"

	TestNetShardCommitteeSize  = 4
	TestNetBeaconCommitteeSize = 4
	TestNetActiveShards        = 2
	TestNetStakingAmountShard  = 175000

	//board and proposal parameters
	TestnetBasicReward                = 50      //50 mili PRV
	TestnetRewardHalflife             = 6307200 //1 year, reduce 10% per year
	TestnetGenesisBlockPaymentAddress = "1Uv46Pu4pqBvxCcPw7MXhHfiAD5Rmi2xgEE7XB6eQurFAt4vSYvfyGn3uMMB1xnXDq9nRTPeiAZv5gRFCBDroRNsXJF1sxPSjNQtivuHk"
)

// for beacon
// public key
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{
	"17YiepCpN6tMwD91MGSXbEBjLBjeUysVcQJm87kCSHfXA2bWM46",
	"16jWkWY5xRqZDkwDQaA32uvnakS27YyyWqeWtap6a6ELEc2Vbwi",
	"15gXn3sEVp41Kdquy696ftT5bi3fRRbExFT7p5ZftEQusaJGi3y",
	"16WQfZBp27g9AUYPXujLqNF7JtZhxdEUfVdox5wEuqM9dC1VYz6",
	"1651Vc7vjFBj69mUZ6BB8KY7uziRWZ57Fi4thXAVqtySGzCqfpq",
	"15EhQjnhchcsozbySChn6sv5ZonVYCC459TkJ6oiTEiprPPMf6n",
	"16ui4k2kPpZrWZaeMq9MnRJocEtoGNJeBN7LWhVFwVpmfukhy8B",
	"18PnGK8M7zj892syajS9H7c7BX8CJfQjyswpLAsHnwQuDzkVSni",
}

// For shard
// public key
var PreSelectShardNodeTestnetSerializedPubkey = []string{
	"183GBqPhSfcEFZP7MQFTnuLVuX2PRkd5HFA3qkqkLN4STghvxpw", //shard 0
	"15ezEJs61P8qq6F8Zrhbcd2RpuqrtDWtzPheJWiEM6ct1sWjFTi", //shard 0
	"16VVUEPJR3uwbkgyVXcwiifsJLcqqR95onn7sZ3jzfs1QofLv11", //shard 1
	"183cdDa9XXppiTTxF4HBiFyRABWtB8v8EBy556WNfX2RnW87VRC", //shard 0

	"14zf4SMg7Jfmmaq64jkjcfRBY8NB9xkg9adSBkXisoEiXUWxxs3", //shard 1
	"16H5t5ezMF16S5j5ZEyHP3N4nBBcsppg5bRfU5Ft8N1VZYBQu38", //shard 1
	"156qsnqcYWPUb8PLbdowV4TtUhS8kuEboABfHgVeh4MguoPwqVj", //shard 1
	"177cGseHedBzrvBTqP2boXBjXb84JgKJEs7fGbCmgJ47gcgNwoK", //shard 1

	"177wqpiaSaswghv2z2y13KR6RPwfMm6mbeTtnfMEdH2iPhmxEbv",
	"16HxssV6VKrGs9qNnCoA1bXi5Uqjco8DyhYLLLqmhgJPAGHyk9A",
	"1771T9b7vo426iizqfyjTVfKz5DM76eQvCdxREJBkEuCD7xXyaF",
	"17wUTdX3qLdyoiw6LAcQmBQYEnDpkYCCKir22WRzfcSXQ1CCNug",
	"15FVc7gKiP9hrazFSQDmJ2TkBi3s9qD3FQBcqCGzvZhLFHxKLLD",
	"17K1jyVmJ94gKmH5eok9XAzCUjuCk64bFzZ1UFtQFTTz6duue8d",
}

// END CONSTANT for network TESTNET

// -------------- FOR INSTRUCTION --------------
// Action for instruction
const (
	SetAction    = "set"
	DeleteAction = "del"
	SwapAction   = "swap"
	RandomAction = "random"
	StakeAction  = "stake"
)

// ---------------------------------------------
var TestnetInitConstant = []string{
	`{  
   "Version":1,
   "Type":"s",
   "LockTime":1557799670,
   "Fee":0,
   "Info":null,
   "SigPubKey":"AsKCsGYkt3JthzFysVzWHxkESGfEoSRFeWafGB+DZRQA",
   "Sig":"P1wbiDpmn2PK9G3FNILqu3JrU5E4ekfrnOz9X7Dd9HRHwp+YDFiEAMLicj7mhKcp3RCR+SWsWOaFxenbrmzXdA==",
   "Proof":"1111111RMhr5Bpy8zSZm7bQEnJcEEbeMSYh6wX9LdwSBjWhESroPN9mvBuwapr4DfKH26bQm9Eu8jtKR3saFoseZj46YQbf7iyuiA6JhAKmekHK1ds4qtFw1ipFzhLYNvp4MYXEupErvQGvZ6bvd9sxDxbwrEFJuV7i8QnHMLftsAAwDAEpDr8MkuxwDXAr5rEjoo9h6SDBHo4c1X6VRBT2GSe3",
   "PubKeyLastByteSender":0,
   "Metadata":null
	}`,
}
