package blockchain

import "time"

//Network fixed params
const (
	// BlockVersion is the current latest supported block version.
	BlockVersion                = 1
	defaultMaxBlkReqPerPeer     = 60
	defaultMaxBlkReqPerTime     = 600
	defaultBroadcastStateTime   = 2 * time.Second  // in second
	defaultProcessPeerStateTime = 5 * time.Second  // in second
	defaultMaxBlockSyncTime     = 2 * time.Second  // in second
	defaultCacheCleanupTime     = 60 * time.Second // in second

	// Threshold ratio
	ThresholdRatioOfDCBCrisis = 9000
	ThresholdRatioOfGOVCrisis = 9000
	ConstitutionPerBoard      = 10
	EndOfFirstBoard           = 25
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
	"17dTfsw6VVuN2Ebf6AxjU2ewPC9DtVCnjPzoPRpSiexqVLtFuZa",
	"16Vi9kjv1RDRpBLdVqc1i3wdnbqPntLL8AxkzxwBU1iCRuCPUbr",
	"16zLUs4RcJ6HhMEn26bfeZdw224BQxAm5RpefKNidowJR3KmZ5u",
	"165Nbz2ZZm52JdzUiSdQ73qTKbQvjBoU88zrhBGej64oGoQoQZ1",
	"16MaJTDAgw4jxpiReBu61uRqkm84ywNwzpwx2xo6Jj6AUh2scU8",
	"16Ekwi3fCqCStjzi1rbgV2S4kniQgbHSEiWpoanetue2dKaJXbb",

	"15L68KC5u26ZqMXjFMvTZgMJcZAKjtVvvjN4cUPso7DDYwepSHa",
	"16sU7voKM75cKkn6weD6DBTTRXD8uqqnz5yYbdLy9X8SA7RWVJf",
}

// For shard
// public key
var PreSelectShardNodeTestnetSerializedPubkey = []string{
	"183GBqPhSfcEFZP7MQFTnuLVuX2PRkd5HFA3qkqkLN4STghvxpw", //shard 0
	"15ezEJs61P8qq6F8Zrhbcd2RpuqrtDWtzPheJWiEM6ct1sWjFTi", //shard 1
	"152nVbYDgLDYve2RA2CQLEJMTUTAHSuZmT6s4DnsaeeDQ8bDD82",
	"16HXR5Jp2LJVV1vV9NTqpPVsVAZvQigv8huJLC9j4TZXnkWc5cw",

	"16VVUEPJR3uwbkgyVXcwiifsJLcqqR95onn7sZ3jzfs1QofLv11",
	"14zf4SMg7Jfmmaq64jkjcfRBY8NB9xkg9adSBkXisoEiXUWxxs3",
	"15tZEEk7qvyFUN9rdLWfSMZj28VDWTywTW41WTLcpYNCpRGpqqu",
	"188xSvTr2ktocRgLPfKaAtw8vgqrgKSMjJD7VPxEoopCLyKxHFi",

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

var TestnetInitConstant = []string{`{"Version":1,"Type":"s","LockTime":1553765098,"Fee":0,"Info":null,"SigPubKey":"AietXs9VLcsFnJMORtdmhdZZ0m0CreSXLOurqKCjzcwB","Sig":"7E90vcQuND1/Usw/2kz8M7C6VotyL2m+i9KcfeXvUEzYkrHMTKHcdivpT00vR4gnxMS3+ClcBmWuXUHBg8TUIA==","Proof":"11111112F9nPGboX8Fh1HmhP6KMy17RokQSRxazSGKZhbQZ9FX7QzAdFp3KrrcqshUHfQ1BsWtp24TfQp24cvucwhgSWCckG32qF5MUPpwFcRz5NPJ31kiLnddkcdCbdPugKeysqkNjhPaotgDa8tSMbusrsKnE5iLCEE1kQ5JLwqwAANmsoKy4DnHik6kfM79kyuF7qeVSja3iWZ572jmBZ5","PubKeyLastByteSender":1,"Metadata":null}`}
