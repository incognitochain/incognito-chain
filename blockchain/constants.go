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
	TestnetFeePerTxKb                 = 1
	TestnetGenesisBlockPaymentAddress = "1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba"
)

var TestnetInitConstant = []string{`{"Version":1,"Type":"s","LockTime":1549889112,"Fee":0,"Info":null,"SigPubKey":"A7GGbCnosNljq25A5o4VIGs7r6WOcs3OrDBJUFd28eEA","Sig":"4gzqBc1TnROMjEdGW1DdIlLRA6pAwbcC3r1macAVy8OaOQaWxcSQXubEgm3oKcJAyE7OnEckV35pwAWD4vr7+A==","Proof":"11111116WGHqpGSLR21nkwRaRVR2vJBD6DR8wKQfB5VCC4TNEXz1XeskmWDehJbmDvr4EeC8x5vGFSrNq4KRs4GoDgn85t7CHJPQWu6s8QWhQVRd621qqT5mBofPcB9WGgQPsD7i4WPxoPKVYhS3jaRXbT2C9S1tHQbW9TytbZKbASDgKygqeijEoWsLW4RXct1oGn2wat2Q1kdPX35AKW1B2R","PubKeyLastByteSender":0,"Metadata":null}`, `{"Version":1,"Type":"s","LockTime":1549889112,"Fee":0,"Info":null,"SigPubKey":"AySFA7ksPnDE7zG+ZKwyk8SaadPLOfJuIn5k4kqUgKcA","Sig":"0jcALduldAkey/6EmKW3EyUQGpJCZ5Vr1lmc7QlzOL3FYEHVwF3kXcDkuPXqqjaH8ueJjDGDqx4N8KpWDfSi7Q==","Proof":"11111116WGHqpGNRGpV3VBz1rndCx6TP4A8eLYeocjg8izynA2YAkx7x38mCir9Nm3oCubXdn25F4sj4jHryBtSbdwJj6o4X43YDftZ9nPsrw4m8DyF6NkxNXbvGj9egkUtypup34hdCXv2L8j5tB9cVUCXVqWeC9axqLLoibXEay4fLrroeRnfNhJ1moNDoQqyRVLrcC7yUjDQz6AUsdd3uFB","PubKeyLastByteSender":0,"Metadata":null}`}

// for beacon
// public key
var PreSelectBeaconNodeTestnetSerializedPubkey = []string{
	"15NmWBEbc8faj4QxHjBh1ugpkuBC8qaoRAp2mktKiwcKiaQgV8i",
	"16QMc6ARYki7eL3p8cj8T8b54ZAhPrnBcfaTY9CgPBDKEtwcm2u",
	"16S3Db9V2kqmmogfggKAD2bpJjXcveJcdUQmx9S3ewEGQBE3rrv",
}

// privatekey
var PreSelectBeaconNodeTestnet = []string{
	"112t8rxTdWfGCtgWvAMHnnEw9vN3R1D7YgD1SSHjAnVGL82HCrMq9yyXrHv3kB4gr84cejnMZRQ973RyHhq2G3MksoTWejNKdSWoQYDFf4gQ",
	"112t8rnXDNYL1RyTuT85JXeX7mJg1Sc6tCby5akSM7pfEGApgAx83X8C46EDu6dFAK6MVcWfQv2sfTk5nV9HqU3jrkrWdhrmi9z34jEhgHak",
	"112t8rnXmEeG5zsS7rExURJfqaRZhm6r4Pypkeag2gprdhtgDpen3LwV68x1nDPRYz2zhyhJTJCGvq1tUx4P1dvrdxF9W9DH7ME7PeGN2ohZ",
}

// For shard
// public key
var PreSelectShardNodeTestnetSerializedPubkey = []string{
	"177KNe6pRhi97hD9LqjUvGxLoNeKh9F5oSeh99V6Td2sQcm7qEu",
	"16W9eKEqyJqKKDkzxcSAKu4G2b1HvZh9FDRmM3ZyC4tN3MkVx6z",
	"17zmxXqnwTK1YE42eNqVJ51mvRaCFoqzm6HogpQQBBt8dWwaUgV",
	"17S44aXG7y9yEmb932MWQrrRT4Rc6pehK2UMC5np84QB2UYZdZM",
	"18YNhMumBmeWE8GJJGbW19esqtB22zUiQx73Rwifxkyt1YKCp1s",
	"15QYRykFuiFhoU56EAJYFRXn5UWurSuyGiZox9y7rCoSzpKW62H",
	"17E9zkHtf495WBkdo47vDB2AVTLLtSq5QtpFU2X7sQcEgHSLmfB",
	"16zsNt8d4UEtGR5c5gLfW4GvhWe3NXQv9K3tBEfom8FTYKNDeim",
	"173HS3C7RFGJDWH8YwtDvMG1s9tgrHu69DtMxVi9NsunwWnrWjk",
	"16uYeyZyRe3pzpWsyjJvqVGsa65R3A4myCzYYX9qr2Gw6L4YBnu",
	"18eckf9WNsj4hrm9goesUqeXgDzracNbr7m86qpHDLbB3jEC4wt",
	"17bgRBWmoNUCRZFAtmRKo7af98t53AWWmVagB5cwPj26Ri7ipPg",
}

// privatekey
var PreSelectShardNodeTestnet = []string{
	"112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV",
	"112t8s2UkZEwS7JtqLHFruRrh4Drj53UzH4A6DrairctKutxVb8Vw2DMzxCReYsAZkXi9ycaSNRHEcB7TJaTwPhyPvqRzu5NnUgTMN9AEKwo",
	"112t8rnYY8UbXGVJ3PsrWxssjr1JXaTPNCPDrneXcQgVQs2MFYwgCzPmTsgqPPbeq8c4QxkrjpHYRaG39ZjtwCmHMJBNh2MxaQvKWw5eUGTM",
	"112t8rnYoj4LesSwRsseGCCYi4J2Py5QxytKKF2WixwEYP4opKUNL2Av9bR2zjfLewf3PQeKcNnuRTTPKgZSJaZH8dfoqY2rmHNekmGMBNDX",
	"112t8rnZ5aGQqJw9bg6fR8AiGe9NFRtSmn73Scd4oNJcE5BNY4Rbju2amkTRW5PUaFpETkKAdSJUMqptjFYb3B8PVAcQhrqooieNFXe5jzTj",
	"112t8rnZUKcW5CBDojVmMD6PmDJzR3VtfqFGWG6HRT9PocB6aewekjebWMm9aQnSncgwDV2GMqAWzspzFYL2vs3C3KnZB9H5YSE4s1SdotHb",
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
