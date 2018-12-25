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
	DefaultPort string
	ShardsNum   int //max 256 shards
	// GenesisBlock defines the first block of the chain.
	GenesisBlockBeacon *BeaconBlock

	// GenesisBlock defines the first block of the chain.
	GenesisBlockShard *ShardBlock
}

// public key
var preSelectBeaconNodeTestnetSerializedPubkey = []string{
	"14uPLoQ8GneVYYyyYUyacQENfZREyX5vxzNCWKwLLyfbwhW1dpX",
	"18AjeCbiownZuCzk7aoJHqPVkVhzNVJhwbGVHb8iecYSLRVxuaA",
	"17ndcst8UgcCf9kxReSKqgizvAjrFHS7nfF3Fxgf3GEMajtpAUH",
}

// private key - seed
var preSelectBeaconNodeTestnet = []string{
	"124sf2tJ4K6iVD6PS4dZzs3BNYuYmHmup3Q9MfhorDrJ6aiSr46",
	"1WG3ys2tsZKpAYV7UEMirmALrMe7wDijnZfTp2Nnd9Ei6upGhc",
	"12K2poTdqzStNZjKdvYzdTBihhigTRWimHWVd7nZ5wRjEPVEZ8n",
}

var preSelectShardNodeTestnetSerializedPubkey = []string{
	"16hYym14N2sa65hhkry54xK3pPN98LsPKkmHa6Us6Ks5NaUAakN",
	"18CSro8nyJ27WxXGdWFS1EZhSH3P8CAHejYpG2GNnJ7KFHGb2km",
	"17LQYtypjiPNgSBxnx7nxG7fmEqjnyCRoELM1a36bormfjwLtTo",
	"18mJ399aMwpnMw3B4iiUZn7kHQ4mNdPfH8i5xobVtGyDBuVKdjG",
	"17E9zkHtf495WBkdo47vDB2AVTLLtSq5QtpFU2X7sQcEgHSLmfB",
	"16N6NMe4uH557JSuLVrqNUjiy1LqfQyJrCMUQNg5gwF6qbm3tPL",
	"18ccr4NMSe2ia7VZRuzuktjfzdo7ykxBD5HhXhywpNpU2UQ5J7G",
	"17ttef56d1DPfRh5YeYiF4WEy5qQ2Rvi31A1gvBZSMVP2GmANxA",
	"18cBb3xhyE8KYNrZZ5u7KFf2eugoawa4aUcUs7xY1nk7q4Z3hkS",
	"17oJNVFcYArTDikiczRmrELYDSexyY6bNxG14tLEkmwdFw5jUFP",
	"15GYHWSDWuvFwBcvxp1iusVcrwsFH4YCVcpFgXyA5XWQJLGytKw",
	"16zrxSeZ79ptRT9uM95XyV3o9speULRsFmKWfGme3dpMEeb682e",
	"15gKTiArrngNJ9cz9Tf3K8XSjGbDXTAFfcf26UeTmzT6pPVz3si",
	"17Dk9dcbrSGondrMEYyjKVgbJusbxx316PmsVJMW5aYEXSw5PrV",
	"17yhuk5kFtXvUyC15QQsfTDvHpRr4vHxivKfeUbn9119Chy8BQF",
	"17FHrUgdsvfETgYHSdM8rAZpSPgaRiAMJaAWNQr7diMMeWrEaYv",
	"153NVGa5GgDkeoidR6BC4c2zPT9AwMbYDteGXWddnncH2FyKon4",
}

var preSelectShardNodeTestnet = []string{
	"12VGen58VjKC8cT3hGhSohdb8n4kz3huXka9UNcYFbUzGdgnXKZ",
	"12nVJxbZnexTmkbqcs9huztH9kN4DBCbjZewHgoyH6kHsLnf9uE",
	"12TZJQbucHA97TJNVtp8xud2BUbrzt1Mgq8Kif1BEdf51BVPFwR",
	"112hmH8nGFpJoqbevB7pmXGqyHenzxuP67tSyh4jfGqr5PbC4yNQ",
	"12ixtJSwVqvLrB4x14ux9c3h2DyUgdfvyjt5XooHkxh6vbcZomW",
	"1cizgU9GeDuEiH7GddwnV2YhPBB3aD1DMir3dynDQahjwQyqTk",
	"17EMNk6W3QpgmjxdtCaZAYmG7sBqN4XxC9bo6YfnAu587ASGv9g",
	"1Jd94JYrqLGLUV6wEa43gdsDGc6JGcy2hYbsNptRuSS3iPz24e",
	"1Q7P7QZGfJSrzC3US1Eqw2iPYDX5rqEG2T8ADsjrML5cQbSaU8",
	"12mZfvHfV5h92TTF45EQgsKU7SkLNRZXLUf6WGLf24EcKfU5Xb6",
	"1n7Zch76tzjdQVLpJxeBmPkimBTWbFmQkSsDsvGAE7GMyUYmuh",
	"17V5TXkUr12JvDrChUQ1kHaQPVFUoVCGGQji9qphTS8asVJBwdF", // me
	"1YX8vFm8zkQEyHLMRSdr8LG4TS7Ua1xq7pWp8dzsbWkDZjsoZY",
	"12ts69QMg83g2v8tutoFPxaKbbxPzpSCCQ12k6XTtDxHzr4d46S",
	"1AH2pPWpF9TjmMaaAUT26WgfSJw31EhdyssHUecxKCmCzZGMB3",
	"12obfKTP2yTtQVx3mcHk2pKBZBoZEeyjmmcfA7SgtNwCFhHKLrB",
	"12k5BfodMQLMDZXmKNwd9gj7eqek3WQqmwYxyj37HBtJpMx1djR",
}

var icoParamsTestnetNew = IcoParams{
	InitialPaymentAddress: TestnetGenesisBlockPaymentAddress,
	InitFundSalary:        TestnetInitFundSalary,
	InitialBondToken:      TestnetInitBondToken,
	InitialCMBToken:       TestnetInitCmBToken,
	InitialDCBToken:       TestnetInitDCBToken,
	InitialGOVToken:       TestnetInitGovToken,
}

var TestNetParams = Params{
	Name:        TestnetName,
	Net:         Testnet,
	DefaultPort: TestnetDefaultPort,
	ShardsNum:   4,
	// blockChain parameters
	GenesisBlockBeacon: CreateBeaconGenesisBlock(1, preSelectBeaconNodeTestnetSerializedPubkey, icoParamsTestnetNew, 1000, 1000, 0),
	GenesisBlockShard:  CreateShardGenesisBlock(1, preSelectShardNodeTestnetSerializedPubkey, 4, icoParamsTestnetNew, 1000, 1000),
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
