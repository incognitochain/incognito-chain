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
	ConstitutionPerBoard      = 30
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
	MainnetGenesisblockPaymentAddress = "1UuyYcHgVFLMd8Qy7T1ZWRmfFvaEgogF7cEsqY98ubQjoQUy4VozTqyfSNjkjhjR85C6GKBmw1JKekgMwCeHtHex25XSKwzb9QPQ2g6a3"
	// ------------- end Mainnet --------------------------------------
)

var MainnetInitConstant = []string{}

// for beacon
// public key
var PreSelectBeaconNodeMainnetSerializedPubkey = PreSelectBeaconNodeTestnetSerializedPubkey

// privatekey
var PreSelectBeaconNodeMainnet = PreSelectBeaconNodeTestnet

// For shard
// public key
var PreSelectShardNodeMainnetSerializedPubkey = PreSelectShardNodeTestnetSerializedPubkey

// privatekey
var PreSelectShardNodeMainnet = PreSelectShardNodeTestnet

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
	TestnetFeePerTxKb                 = 0
	TestnetGenesisBlockPaymentAddress = "1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba"
)

var TestnetInitConstant = []string{`{"Version":1,"Type":"s","LockTime":1549889112,"Fee":0,"Info":null,"SigPubKey":"A7GGbCnosNljq25A5o4VIGs7r6WOcs3OrDBJUFd28eEA","Sig":"4gzqBc1TnROMjEdGW1DdIlLRA6pAwbcC3r1macAVy8OaOQaWxcSQXubEgm3oKcJAyE7OnEckV35pwAWD4vr7+A==","Proof":"11111116WGHqpGSLR21nkwRaRVR2vJBD6DR8wKQfB5VCC4TNEXz1XeskmWDehJbmDvr4EeC8x5vGFSrNq4KRs4GoDgn85t7CHJPQWu6s8QWhQVRd621qqT5mBofPcB9WGgQPsD7i4WPxoPKVYhS3jaRXbT2C9S1tHQbW9TytbZKbASDgKygqeijEoWsLW4RXct1oGn2wat2Q1kdPX35AKW1B2R","PubKeyLastByteSender":0,"Metadata":null}`, `{"Version":1,"Type":"s","LockTime":1549889112,"Fee":0,"Info":null,"SigPubKey":"AySFA7ksPnDE7zG+ZKwyk8SaadPLOfJuIn5k4kqUgKcA","Sig":"0jcALduldAkey/6EmKW3EyUQGpJCZ5Vr1lmc7QlzOL3FYEHVwF3kXcDkuPXqqjaH8ueJjDGDqx4N8KpWDfSi7Q==","Proof":"11111116WGHqpGNRGpV3VBz1rndCx6TP4A8eLYeocjg8izynA2YAkx7x38mCir9Nm3oCubXdn25F4sj4jHryBtSbdwJj6o4X43YDftZ9nPsrw4m8DyF6NkxNXbvGj9egkUtypup34hdCXv2L8j5tB9cVUCXVqWeC9axqLLoibXEay4fLrroeRnfNhJ1moNDoQqyRVLrcC7yUjDQz6AUsdd3uFB","PubKeyLastByteSender":0,"Metadata":null}`}

// for beacon
// public key
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{
	"15NmWBEbc8faj4QxHjBh1ugpkuBC8qaoRAp2mktKiwcKiaQgV8i",
	"16QMc6ARYki7eL3p8cj8T8b54ZAhPrnBcfaTY9CgPBDKEtwcm2u",
	"16S3Db9V2kqmmogfggKAD2bpJjXcveJcdUQmx9S3ewEGQBE3rrv",
	"158haewyeNr4WXGk4Bao2MUonNNaAjSpYeUTJ8JoD4at2AjVS45",
	"177ZHyh2WpZeVcFPJULigwkH6dahem4jDHkXQMZGNgVDijyKjKJ",
	"1671hBGTAT1ui2BQGqpzYyy3pVLPvdDTPEMLfoLix7igUyzG6sE",

	"15QDhbfx5bE7HZPkpmsayKPD13ZM1zfrCJXdbMm8y63LZU8fbG9",
	"15yb9gJC7QH3hTdjxNyV4Mdr7FJkjia5SQ5Bbwb4EHAtBTXSWC1",
	"15MC1D4orEfnJFaEeebVuUmxq2UDAwcogd8efPMDxC2cJW2MFDD",
	"15nAkpuo4BKCPk3VFey81WZaHV1YYV8HW8925eNb24mhTTtHFqV",
	"18E2LKY4PND2tAwcM9hNsZsvK1tUpKUwtLjMVRmjCBURKKD14Du",
	"17AHQTTHCnTDrt1wj3t7rThjEQndrSpk75Pv4xEpAsMw4caD4xa",
	"16XruBoQVSdMwvLzo28cL4SmdQJyBLrZD3Y9ezjAdcxaBp6GXUv",
	"16rsKSNDMGUcr7Xpt5PpkA1KiHkujEcF9BPMN9AW4gS8ZZNBu9R",
	"15QRR2WDLQcpru794E7GFjsJthhJWLhSDo2t9DToizWBAXZkCyi",
	"17k3AdAH1crag9gV4VNT7SLK76zbnc8jMUw9WPmztk7GG5n5YWt",
	"17HZ2km8EHB94EnxC9o4nP14GbzrB4Ci7yULmS7DcooY6eP4wNG",
	"17p8PgmAUpmm6CJqAtw8B5W3hLCdVYiq1W3fZ7v4hy183Pq9xYf",
	"17PKTxtJUtFLb73tu77qPqYNVvYmGMPLPCh9A1ukG3VvWqWSHQL",
	"16eykkm9y72xbrG9pLhdCWm7FVEY77gCrbpaJMgp8Ck7SZXDDkn",
	"188cBXFqnKoTroNAPtwYThbbmEoy1gAZnLCYQPL4q6gEFbz1xCs",
	"164DBoH1SDYY5xnzctuE7KEnut48tbeSTQ27zHXaTv8vvbUc2GF",
	"15rfZautPLY7mxe2MrBCmgpmRCoXfCXe4Ay1mKRKETcxabGQeDq",
	"15PMBJz2vQDPWXmDUZY33SN4uG5BKgsHk2sTxkJo156PJPy28dm",
}

// privatekey
var PreSelectBeaconNodeTestnet = []string{
	"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",
}

// For shard
// public key
var PreSelectShardNodeTestnetSerializedPubkey = []string{
	"177KNe6pRhi97hD9LqjUvGxLoNeKh9F5oSeh99V6Td2sQcm7qEu", //shard 0
	"16W9eKEqyJqKKDkzxcSAKu4G2b1HvZh9FDRmM3ZyC4tN3MkVx6z", //shard 0
	"17zmxXqnwTK1YE42eNqVJ51mvRaCFoqzm6HogpQQBBt8dWwaUgV", //shard 0
	"17wqq26DuTQ6Hr7ocuMBdeu934rLqSoMyxib4RQdQhLUQ7Le3KD", //shard 0

	"17S44aXG7y9yEmb932MWQrrRT4Rc6pehK2UMC5np84QB2UYZdZM", //shard 1
	"18YNhMumBmeWE8GJJGbW19esqtB22zUiQx73Rwifxkyt1YKCp1s", //shard 1
	"15QYRykFuiFhoU56EAJYFRXn5UWurSuyGiZox9y7rCoSzpKW62H", //shard 1
	"15Mjx8UwK9hG2xFrmyVxco39BSDLidvcy7MBDTivuPBme6JJ24e", //shard 1

	"17E9zkHtf495WBkdo47vDB2AVTLLtSq5QtpFU2X7sQcEgHSLmfB",
	"16zsNt8d4UEtGR5c5gLfW4GvhWe3NXQv9K3tBEfom8FTYKNDeim",
	"173HS3C7RFGJDWH8YwtDvMG1s9tgrHu69DtMxVi9NsunwWnrWjk",
	"16uYeyZyRe3pzpWsyjJvqVGsa65R3A4myCzYYX9qr2Gw6L4YBnu",
	"18eckf9WNsj4hrm9goesUqeXgDzracNbr7m86qpHDLbB3jEC4wt",
	"17bgRBWmoNUCRZFAtmRKo7af98t53AWWmVagB5cwPj26Ri7ipPg",
}

// privatekey
var PreSelectShardNodeTestnet = []string{
	"112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV", //shard 0
	"112t8s2UkZEwS7JtqLHFruRrh4Drj53UzH4A6DrairctKutxVb8Vw2DMzxCReYsAZkXi9ycaSNRHEcB7TJaTwPhyPvqRzu5NnUgTMN9AEKwo", //shard 0
	"112t8rnYY8UbXGVJ3PsrWxssjr1JXaTPNCPDrneXcQgVQs2MFYwgCzPmTsgqPPbeq8c4QxkrjpHYRaG39ZjtwCmHMJBNh2MxaQvKWw5eUGTM", //shard 0
	"112t8rniLPf1ten3xsPpFx2P6Ug7D4UVz4yYGKQfk8YpzY3RDEqmQ9FxTXd9G6r6dJXb9URqvNApfQzVQqqRqvmfarRnaR4tVUATrxbmcQD6", //shard 0

	"112t8rnYoj4LesSwRsseGCCYi4J2Py5QxytKKF2WixwEYP4opKUNL2Av9bR2zjfLewf3PQeKcNnuRTTPKgZSJaZH8dfoqY2rmHNekmGMBNDX", //shard 1
	"112t8rnZ5aGQqJw9bg6fR8AiGe9NFRtSmn73Scd4oNJcE5BNY4Rbju2amkTRW5PUaFpETkKAdSJUMqptjFYb3B8PVAcQhrqooieNFXe5jzTj", //shard 1
	"112t8rnZUKcW5CBDojVmMD6PmDJzR3VtfqFGWG6HRT9PocB6aewekjebWMm9aQnSncgwDV2GMqAWzspzFYL2vs3C3KnZB9H5YSE4s1SdotHb", //shard 1
	"112t8rnhvfmkbNz6bmrtBM74pHiUFgML4yhuSZ431teYD3S1sEUVPKpB6gGFjgrPu8e9uc7eqfnmqqtj9kTxYXTRbYkvyrsv3ydxVPmRM8fX", //shard 1

	"112t8rnZdou7TJBdGsWUJ3jWxuQYHdEKndzmKHhHzjdHzckLf7dAz4uBr2oVPF3ChNjs9owpobjaySzPrK3nUsZukVWv2MybKiajw6kD6M69",
	"112t8rna913eNyB7uyfi6Nbpg9Fqv4ic8uyCyC79S8MhkTgVQYnxpEJFBQZsEveNa3AGWqHoBiEp1dgMH5e2UUpcN6XLvbVo6jaiy3UiiaUY",
	"112t8rnaTDoXRzYbiB5BZKdZcxjEEKoZ7W4h5QFJ7iwgQ1MqDALCL5c7sexj42GvMLHsXbCmMcjx4JZEUW2UramvgrTwVr9TCp16obmuwTCs",
	"112t8rnaet4nhVpq517eXmCNnE4JAd2EsTZgfzn6SVKgfSQ6rS7h6AYETMBUNkiZ8PpXqRwCYjpGCLk5DpPhHQNqa8tcRacMKffbYoTWGK9W",
	"112t8rnb1VhdWUR4SwVNTAokxntpNT5EcLFg6w6DovD9ZptT1DFsAXfrorLofP9uzCZC3JechZowMnc7fcXJ8nvsjdSEr3M6tzWVYBdLJmNW",
	"112t8rnbDuvxqCrnzQbRkBLrrGoaqTHnvKSBa4tdt4585gJHJHcsm4shE4yBardCsLkXV2Rtogom6Gy8rn4Z5vQXXmanBoVPn2wQhFLTYz4E",
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
	salaryPerTx = "salaryPerTx"
	basicSalary = "basicSalary"
	salaryFund  = "salaryFund"
	feePerTxKb  = "feePerTxKb"
)

// ---------------------------------------------
