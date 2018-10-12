package blockchain

import (
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
)

// DNSSeed identifies a DNS seed.
type DNSSeed struct {
	// Host defines the hostname of the seed.
	Host string

	// HasFiltering defines whether the seed supports filtering
	// by service flags (wire.ServiceFlag).
	HasFiltering bool
}

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

	// DNSSeeds defines a list of DNS seeds for the network that are used
	// as one method to discover peers.
	DNSSeeds []string

	// GenesisBlock defines the first block of the chain.
	GenesisBlock *Block

	// GenesisHash is the starting block hash.
	GenesisHash *common.Hash

	// SubsidyReductionInterval is the interval of blocks before the subsidy
	// is reduced.
	SubsidyReductionInterval int32

	// TargetTimespan is the desired amount of time that should elapse
	// before the block difficulty requirement is examined to determine how
	// it should be changed in order to maintain the desired block
	// generation rate.
	TargetTimespan time.Duration

	// TargetTimePerBlock is the desired amount of time to generate each
	// block.
	TargetTimePerBlock time.Duration

	// RetargetAdjustmentFactor is the adjustment factor used to limit
	// the minimum and maximum amount of adjustment that can occur between
	// difficulty retargets.
	RetargetAdjustmentFactor int64

	// ReduceMinDifficulty defines whether the network should reduce the
	// minimum required difficulty after a long enough period of time has
	// passed without finding a block.  This is really only useful for test
	// networks and should not be set on a main network.
	ReduceMinDifficulty bool

	// MinDiffReductionTime is the amount of time after which the minimum
	// required difficulty should be reduced when a block hasn't been found.
	//
	// NOTE: This only applies if ReduceMinDifficulty is true.
	MinDiffReductionTime time.Duration

	// GenerateSupported specifies whether or not CPU mining is allowed.
	GenerateSupported bool
}

var pposValidators = []string{
	"124sf2tJ4K6iVD6PS4dZzs3BNYuYmHmup3Q9MfhorDrJ6aiSr46",
	"1WG3ys2tsZKpAYV7UEMirmALrMe7wDijnZfTp2Nnd9Ei6upGhc",
	"12K2poTdqzStNZjKdvYzdTBihhigTRWimHWVd7nZ5wRjEPVEZ8n",
	"12VGen58VjKC8cT3hGhSohdb8n4kz3huXka9UNcYFbUzGdgnXKZ",
	"12nVJxbZnexTmkbqcs9huztH9kN4DBCbjZewHgoyH6kHsLnf9uE",
	"12TZJQbucHA97TJNVtp8xud2BUbrzt1Mgq8Kif1BEdf51BVPFwR",
	"12EgPTDbermjn6xSxtdjPiguaV4moStJK1uiHnSSnw6rdM9TEBe",
	"12ixtJSwVqvLrB4x14ux9c3h2DyUgdfvyjt5XooHkxh6vbcZomW",
	"12ogKMsmp3iwapFLa1HLeh2SmeEevF9bbUeLbcceD2TUY3MiJtW",
	"1QHok11tCavbA328ASuHWpoS3P5KuAkWoDLRJZXGbWpBK6xjAS",
	"1Jd94JYrqLGLUV6wEa43gdsDGc6JGcy2hYbsNptRuSS3iPz24e",
	"1Q7P7QZGfJSrzC3US1Eqw2iPYDX5rqEG2T8ADsjrML5cQbSaU8",
	"12mZfvHfV5h92TTF45EQgsKU7SkLNRZXLUf6WGLf24EcKfU5Xb6",
	"1n7Zch76tzjdQVLpJxeBmPkimBTWbFmQkSsDsvGAE7GMyUYmuh",
	"12uoLbfkWGiuYQ9JTRJMkJy1q3HUdrSr5Rzat7dBfHwYGfoDnta",
	"1YX8vFm8zkQEyHLMRSdr8LG4TS7Ua1xq7pWp8dzsbWkDZjsoZY",
	"12ts69QMg83g2v8tutoFPxaKbbxPzpSCCQ12k6XTtDxHzr4d46S",
	"1AH2pPWpF9TjmMaaAUT26WgfSJw31EhdyssHUecxKCmCzZGMB3",
	"12obfKTP2yTtQVx3mcHk2pKBZBoZEeyjmmcfA7SgtNwCFhHKLrB",
	"12k5BfodMQLMDZXmKNwd9gj7eqek3WQqmwYxyj37HBtJpMx1djR",
}

// MainNetParams defines the network parameters for the main coin network.
var MainNetParams = Params{
	Name:        MainetName,
	Net:         Mainnet,
	DefaultPort: MainnetDefaultPort,
	DNSSeeds: []string{
		/*{"seed.coin.sipa.be", true},
		{"dnsseed.bluematt.me", true},
		{"dnsseed.coin.dashjr.org", false},
		{"seed.coinstats.com", true},
		{"seed.bitnodes.io", false},
		{"seed.coin.jonasschnelli.ch", true},*/
		//"/ip4/127.0.0.1/tcp/9333/ipfs/QmRuvXN7BpTqxqpPLSftDFbKEYiRZRUb7iqZJcz2CxFvVS",
	},

	// blockChain parameters
	GenesisBlock:             GenesisBlockGenerator{}.CreateGenesisBlockPoSParallel(0x18aea41a, 0x1d00ffff, 1, MainnetGenesisBlockReward, MainnetGenesisblockPaymentAddress, pposValidators, MainnetGenesisBlockReward),
	SubsidyReductionInterval: 210000,
	TargetTimespan:           time.Hour * 24 * 14, // 14 days
	TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
	RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0,
	GenerateSupported:        false,
}

// TestNetParams defines the network parameters for the test coin network.
var TestNetParams = Params{
	Name:        TestnetName,
	Net:         Testnet,
	DefaultPort: TestnetDefaultPort,
	DNSSeeds: []string{
		/*{"seed.coin.sipa.be", true},
		{"dnsseed.bluematt.me", true},
		{"dnsseed.coin.dashjr.org", false},
		{"seed.coinstats.com", true},
		{"seed.bitnodes.io", false},
		{"seed.coin.jonasschnelli.ch", true},*/
		//"/ip4/127.0.0.1/tcp/9333/ipfs/QmRuvXN7BpTqxqpPLSftDFbKEYiRZRUb7iqZJcz2CxFvVS",
	},

	// blockChain parameters
	GenesisBlock:             GenesisBlockGenerator{}.CreateGenesisBlockPoSParallel(0x18aea41a, 0x1d00ffff, 1, TestnetGenesisBlockReward, TestnetGenesisBlockPaymentAddress, pposValidators, TestnetInitFundSalary),
	SubsidyReductionInterval: 210000,
	TargetTimespan:           time.Hour * 24 * 14, // 14 days
	TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
	RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0,
	GenerateSupported:        false,
}
