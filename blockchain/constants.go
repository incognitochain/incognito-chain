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
	MainnetBasicReward                = 50       //50 mili PRV
	MainnetRewardHalflife             = 31536000 //5 year
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
	Testnet            = 0x16
	TestnetName        = "testnet"
	TestnetDefaultPort = "9444"
	
	TestNetShardCommitteeSize  = 16
	TestNetBeaconCommitteeSize = 4
	TestNetActiveShards        = 2
	TestNetStakingAmountShard  = 175000

	//board and proposal parameters
	TestnetBasicReward                = 50       //50 mili PRV
	TestnetRewardHalflife             = 31536000 //5 year
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
	// Committee of shard 0
	"183GBqPhSfcEFZP7MQFTnuLVuX2PRkd5HFA3qkqkLN4STghvxpw", //shard 0
	"15ezEJs61P8qq6F8Zrhbcd2RpuqrtDWtzPheJWiEM6ct1sWjFTi", //shard 0
	"16VVUEPJR3uwbkgyVXcwiifsJLcqqR95onn7sZ3jzfs1QofLv11", //shard 1
	"183cdDa9XXppiTTxF4HBiFyRABWtB8v8EBy556WNfX2RnW87VRC", //shard 0
	"17FYHpLygyHdtQcf6oUnJCLYGDMgswYwDWHBtWPA9Di6oM9yvqB", //shard 0
	"15LMbkYQUTdoWP2FWodStn7eq9zkdX46WdyY92j6G24vv4nruiK", //shard 0
	"17QSVjBq4LYf4H9fsdrtjHAAAGrjxivfxCRA66sM8veqBLoWMhm", //shard 0
	"17bs8G9gygxFNasCKGQ6sZfS4pL7dUc5cqxmvErg7w2jGitYwC5", //shard 0
	"176kYQgu6ngaAeecqypGpBgomByTbFCYp5wRGKcVQdC6xowD2eN", //shard 0
	"16gGNuW3e7jf4125TWNKyhm8wkEb5sggbC3juJ15YcPuA6TFe5J", //shard 0
	"15wAKpAHph28Mr3HcPsVyQmXDm4GvdE7E3aNG5q2g8rJH7fjRy2", //shard 0
	"14xigQHFhQfvt6edebYrfovAwgVUG3BUh1Ax3gWvE6CW3VZUwjJ", //shard 0
	"14xGoftt4xsYjSN63rhs2mWU4feqKCoe3f1FpztW8sW78BLSMZW", //shard 0
	"15WxuiLV4XN4UjtgBCfxRPQKpMPh2wzi7BrUhrwbVKASJMnToxP", //shard 0
	"17JLAMk5aXmgDRoRZ6XNk26nS5o54sVwu7DuB94nmwpmQR94dME", //shard 0
	"17c1m1MWp9Kyu4dmghQCWc4eM2kwNSaHE2i9aaLkhVd5nrNeUuH", //shard 0
	
	// Committee of shard 1
	"14zf4SMg7Jfmmaq64jkjcfRBY8NB9xkg9adSBkXisoEiXUWxxs3", //shard 1
	"16H5t5ezMF16S5j5ZEyHP3N4nBBcsppg5bRfU5Ft8N1VZYBQu38", //shard 1
	"156qsnqcYWPUb8PLbdowV4TtUhS8kuEboABfHgVeh4MguoPwqVj", //shard 1
	"177cGseHedBzrvBTqP2boXBjXb84JgKJEs7fGbCmgJ47gcgNwoK", //shard 1
	"15MdBPBqMmrZ7MtZxjzPBKc8oNC3kLisAvtiEs6vrNFPnByCSx7", //shard 1
	"18E1jmMz3R2vidv3MC5MCtS1LpZTshnh5YSjnJqBrpZrnDxEpjj", //shard 1
	"16a14WSkpGsF5o4B6pk3TbNMFbW4s6pXN9haPLjZfu8rPMSzHTa", //shard 1
	"15FpmWXjGaLPVmBfAguBcJNzLdpbW32joGJdTTDqVCrUHRMwJmt", //shard 1
	"16FJGLmgmX7HXQXLZpFxpAtSRyAsnjVW5wyQ5q631qAdjtY2LiG", //shard 1
	"186dkkeebuaHEjo99BX1omYSKdJNwah2dLsCRadHFxSqyY2E7bP", //shard 1
	"166S7YKvwYWYwaykFPGLnewqBwf5bydg7cR7Livog9TokvzXsqA", //shard 1
	"17qg9rs5PysAQJfhwYAAgBWAVTxq91vVq3h8sNAzpWeWVSRdKmN", //shard 1
	"15JPGWmC5ebFxg5zs6tXbWuAJBok3kd7pNfpnDZCesNLkbJbJsZ", //shard 1
	"17RM5fdYn7yAhxe3yU4mSR18y3TGkTaoa4Ph3HaVwfY3sRCNMVZ", //shard 1
	"168RLoj25TCmzq32LEARGWX2d8SAHvS4egiLYtwLkp41csGWHUG", //shard 1
	"165bmHpowcdfmpLzc2RYPjDqBckmNwDo6fXXNKVTJhCG5huuXq5", //shard 1
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
