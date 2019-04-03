package blockchain

import "time"

//Network fixed params
const (
	// BlockVersion is the current latest supported block version.
	BlockVersion                = 1
	defaultMaxBlkReqPerPeer     = 60
	defaultMaxBlkReqPerTime     = 600
	defaultBroadcastStateTime   = 2 * time.Second  // in second
	defaultProcessPeerStateTime = 4 * time.Second  // in second
	defaultMaxBlockSyncTime     = 2 * time.Second  // in second
	defaultCacheCleanupTime     = 60 * time.Second // in second

	// Threshold ratio
	ThresholdRatioOfDCBCrisis = 9000
	ThresholdRatioOfGOVCrisis = 9000
	ConstitutionPerBoard      = 10
	EndOfFirstBoard           = 90
	BaseSalaryBoard           = 10000
)

// CONSTANT for network MAINNET
const (
	// ------------- Mainnet ---------------------------------------------
	Mainnet            = 0x01
	MainetName         = "mainnet"
	MainnetDefaultPort = "9333"

	MainNetShardCommitteeSize  = 1
	MainNetBeaconCommitteeSize = 1
	MainNetActiveShards        = 2

	//board and proposal parameters
	MainnetSalaryPerTx                = 0
	MainnetBasicSalary                = 0
	MainnetInitFundSalary             = 0
	MainnetInitDCBToken               = 0
	MainnetInitGovToken               = 0
	MainnetInitCmBToken               = 0
	MainnetInitBondToken              = 0
	MainnetFeePerTxKb                 = 0
	MainnetGenesisblockPaymentAddress = "1Uv2zzR4LgfX8ToQe8ub3bYcCLk3uDU1sm9U9hiu9EKYXoS77UdikfT9s8d5YjhsTJm61eazsMwk2otFZBYpPHwiMn8z6bKWWJRspsLky"
	// ------------- end Mainnet --------------------------------------
)

var MainnetInitConstant = []string{}

// for beacon
// public key
var PreSelectBeaconNodeMainnetSerializedPubkey = PreSelectBeaconNodeTestnetSerializedPubkey

// For shard
// public key
var PreSelectShardNodeMainnetSerializedPubkey = PreSelectShardNodeTestnetSerializedPubkey

// END CONSTANT for network MAINNET

// CONSTANT for network TESTNET
const (
	Testnet            = 0x02
	TestnetName        = "testnet"
	TestnetDefaultPort = "9444"

	TestNetShardCommitteeSize  = 1
	TestNetBeaconCommitteeSize = 1
	TestNetActiveShards        = 2

	//board and proposal parameters
	TestnetSalaryPerTx                = 10
	TestnetBasicSalary                = 10
	TestnetInitFundSalary             = 1000000
	TestnetInitDCBToken               = 10000
	TestnetInitGovToken               = 10000
	TestnetInitCmBToken               = 10000
	TestnetInitBondToken              = 10000
	TestnetFeePerTxKb                 = 1
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
	InitAction   = "init"
	DeleteAction = "del"
	SwapAction   = "swap"
	RandomAction = "random"
	StakeAction  = "stake"
)

// Key param for instruction
const (
	salaryFund = "salaryFund"
)

// ---------------------------------------------

// Todo @hy don't merge this
var TestnetInitConstant = []string{`{"Version":1,"Type":"s","LockTime":1554203055,"Fee":0,"Info":null,"SigPubKey":"A58CKhajw92BH9mFlRBEbCrAOl8nzD9Ey4TdMLWDKL0A","Sig":"ZjxxKZv2v7nPTnpq8x4cZNjUg27PollvtiwjzJly2xcnNUJtOg8LpzGHF5QP47/jc1x9FeXZdaSE2siT31PWcA==","Proof":"11111116WGHqpGRpb156mNjZAF55nAeqtfP68nb2hQYXCzLzACytCMb85wsvLd7aTE6FxWQK88KPhnZEg9YxpAzcvi8xypaxXkeq2hxTTbBZin4qgFheBPEinySyKHJoudMEKtFrZBxoB8jCLTUmD9t4iEjdHS8mG8Khmdbx5Ms9UWDzW3oWTaeQQFSnHoUh7TFT6BhpHhsMj7S7nSQgGFGMNx","PubKeyLastByteSender":0,"Metadata":null}`}
