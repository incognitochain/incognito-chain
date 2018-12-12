package blockchain

/*
Params defines a network by its params. These params may be used by Applications
to differentiate network as well as addresses and keys for one network
from those intended for use on another network
*/
type ParamsNew struct {
	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net uint32

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort string

	// GenesisBlock defines the first block of the chain.
	GenesisBlock *BlockV2
}

//"124sf2tJ4K6iVD6PS4dZzs3BNYuYmHmup3Q9MfhorDrJ6aiSr46",
//"1WG3ys2tsZKpAYV7UEMirmALrMe7wDijnZfTp2Nnd9Ei6upGhc",
//"12K2poTdqzStNZjKdvYzdTBihhigTRWimHWVd7nZ5wRjEPVEZ8n",
var preSelectBeaconNodeTestnet = []string{
	"124sf2tJ4K6iVD6PS4dZzs3BNYuYmHmup3Q9MfhorDrJ6aiSr46",
	"1WG3ys2tsZKpAYV7UEMirmALrMe7wDijnZfTp2Nnd9Ei6upGhc",
	"12K2poTdqzStNZjKdvYzdTBihhigTRWimHWVd7nZ5wRjEPVEZ8n",
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

var TestNetParamsNew = ParamsNew{
	Name:        TestnetName,
	Net:         Testnet,
	DefaultPort: TestnetDefaultPort,

	// blockChain parameters
	GenesisBlock: BeaconBlockGenerator{}.CreateBeaconGenesisBlock(1, preSelectBeaconNodeTestnet, preSelectShardNodeTestnet, icoParamsTestnetNew, 1000, 1000),
}
