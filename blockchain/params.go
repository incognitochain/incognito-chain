package blockchain

/*
Params defines a network by its params. These params may be used by Applications
to differentiate network as well as addresses and keys for one network
from those intended for use on another network
*/
type Params struct {
	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net uint32

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort         string
	ShardCommitteeSize  int
	BeaconCommitteeSize int
	ActiveShards        int
	// GenesisBlock defines the first block of the chain.
	GenesisBeaconBlock *BeaconBlock

	// GenesisBlock defines the first block of the chain.
	GenesisShardBlock *ShardBlock
}

type IcoParams struct {
	InitialPaymentAddress string
	InitFundSalary        uint64
	InitialDCBToken       uint64
	InitialCMBToken       uint64
	InitialGOVToken       uint64
	InitialBondToken      uint64
	InitialVoteDCBToken   uint64
	InitialVoteGOVToken   uint64
}

// FOR TESTNET
const (
	TestNetShardsNum           = 4
	TestNetShardCommitteeSize  = 3
	TestNetBeaconCommitteeSize = 3
	TestNetActiveShards        = 1
)

// for beacon
// public key
var preSelectBeaconNodeTestnetSerializedPubkey = []string{
	"15NmWBEbc8faj4QxHjBh1ugpkuBC8qaoRAp2mktKiwcKiaQgV8i",
	"16QMc6ARYki7eL3p8cj8T8b54ZAhPrnBcfaTY9CgPBDKEtwcm2u",
	"16S3Db9V2kqmmogfggKAD2bpJjXcveJcdUQmx9S3ewEGQBE3rrv",
}

// privatekey
var preSelectBeaconNodeTestnet = []string{
	"112t8rxTdWfGCtgWvAMHnnEw9vN3R1D7YgD1SSHjAnVGL82HCrMq9yyXrHv3kB4gr84cejnMZRQ973RyHhq2G3MksoTWejNKdSWoQYDFf4gQ",
	"112t8rnXDNYL1RyTuT85JXeX7mJg1Sc6tCby5akSM7pfEGApgAx83X8C46EDu6dFAK6MVcWfQv2sfTk5nV9HqU3jrkrWdhrmi9z34jEhgHak",
	"112t8rnXmEeG5zsS7rExURJfqaRZhm6r4Pypkeag2gprdhtgDpen3LwV68x1nDPRYz2zhyhJTJCGvq1tUx4P1dvrdxF9W9DH7ME7PeGN2ohZ",
}

// For shard
// public key
var preSelectShardNodeTestnetSerializedPubkey = []string{
	"177KNe6pRhi97hD9LqjUvGxLoNeKh9F5oSeh99V6Td2sQcm7qEu",
	"1671hBGTAT1ui2BQGqpzYyy3pVLPvdDTPEMLfoLix7igUyzG6sE",
	"17zmxXqnwTK1YE42eNqVJ51mvRaCFoqzm6HogpQQBBt8dWwaUgV",
	"17S44aXG7y9yEmb932MWQrrRT4Rc6pehK2UMC5np84QB2UYZdZM",
	"17E9zkHtf495WBkdo47vDB2AVTLLtSq5QtpFU2X7sQcEgHSLmfB",
	"18YNhMumBmeWE8GJJGbW19esqtB22zUiQx73Rwifxkyt1YKCp1s",
	"15QYRykFuiFhoU56EAJYFRXn5UWurSuyGiZox9y7rCoSzpKW62H",
	"16zsNt8d4UEtGR5c5gLfW4GvhWe3NXQv9K3tBEfom8FTYKNDeim",
	"173HS3C7RFGJDWH8YwtDvMG1s9tgrHu69DtMxVi9NsunwWnrWjk",
	"16uYeyZyRe3pzpWsyjJvqVGsa65R3A4myCzYYX9qr2Gw6L4YBnu",
	"18eckf9WNsj4hrm9goesUqeXgDzracNbr7m86qpHDLbB3jEC4wt",
	"17bgRBWmoNUCRZFAtmRKo7af98t53AWWmVagB5cwPj26Ri7ipPg",
}

// privatekey
var preSelectShardNodeTestnet = []string{
	"112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV",
	"112t8rnYBW9trs5rzxrMzLU5AnzngQhbp6X4c3xyamFkWU7PwWRq6gprDkm6mf3ZjxaeYQmSpe3xorpWHo3JLLZFHCHSgqd8u19XkVuMGz1M",
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

var icoParamsTestnetNew = IcoParams{
	InitialPaymentAddress: TestnetGenesisBlockPaymentAddress,
	InitFundSalary:        TestnetInitFundSalary,
	InitialBondToken:      TestnetInitBondToken,
	InitialCMBToken:       TestnetInitCmBToken,
	InitialDCBToken:       TestnetInitDCBToken,
	InitialGOVToken:       TestnetInitGovToken,
}

var ChainTestParam = Params{
	Name:                TestnetName,
	Net:                 Testnet,
	DefaultPort:         TestnetDefaultPort,
	ShardCommitteeSize:  TestNetShardCommitteeSize,  //TestNetShardCommitteeSize,
	BeaconCommitteeSize: TestNetBeaconCommitteeSize, //TestNetBeaconCommitteeSize,
	ActiveShards:        TestNetActiveShards,
	// blockChain parameters
	GenesisBeaconBlock: CreateBeaconGenesisBlock(1, preSelectBeaconNodeTestnetSerializedPubkey[:], icoParamsTestnetNew, 1000, 1000, 0),
	GenesisShardBlock:  CreateShardGenesisBlock(1, preSelectShardNodeTestnetSerializedPubkey[:], icoParamsTestnetNew),
}

// END TESTNET

// FOR MAINNET
const (
	MainNetShardsNum           = TestNetShardsNum
	MainNetShardCommitteeSize  = TestNetShardCommitteeSize
	MainNetBeaconCommitteeSize = TestNetBeaconCommitteeSize
	MainNetActiveShards        = TestNetActiveShards
)

// for beacon
// public key
var preSelectBeaconNodeMainnetSerializedPubkey = preSelectBeaconNodeTestnetSerializedPubkey

// For shard
// public key
var preSelectShardNodeMainnetSerializedPubkey = preSelectShardNodeTestnetSerializedPubkey

var icoParamsMainnetNew = icoParamsTestnetNew

var ChainMainParam = Params{
	Name:                MainetName,
	Net:                 Mainnet,
	DefaultPort:         MainnetDefaultPort,
	ShardCommitteeSize:  MainNetShardCommitteeSize,  //MainNetShardCommitteeSize,
	BeaconCommitteeSize: MainNetBeaconCommitteeSize, //MainNetBeaconCommitteeSize,
	ActiveShards:        MainNetActiveShards,
	// blockChain parameters
	GenesisBeaconBlock: CreateBeaconGenesisBlock(1, preSelectBeaconNodeMainnetSerializedPubkey[:], icoParamsMainnetNew, 1000, 1000, 0),
	GenesisShardBlock:  CreateShardGenesisBlock(1, preSelectShardNodeMainnetSerializedPubkey[:], icoParamsMainnetNew),
}
