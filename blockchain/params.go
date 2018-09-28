package blockchain

import (
	"math/big"
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
)

var (
	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// mainPowLimit is the highest proof of work value a coin block can
	// have for the main network.  It is the value 2^224 - 1.
	mainPowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 224), bigOne)
)

// DNSSeed identifies a DNS seed.
type DNSSeed struct {
	// Host defines the hostname of the seed.
	Host string

	// HasFiltering defines whether the seed supports filtering
	// by service flags (wire.ServiceFlag).
	HasFiltering bool
}

/**
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

	// PowLimit defines the highest allowed proof of work value for a block
	// as a uint256.
	PowLimit *big.Int

	// PowLimitBits defines the highest allowed proof of work value for a
	// block in compact form.
	PowLimitBits uint32

	// CoinbaseMaturity is the number of blocks required before newly mined
	// coins (coinbase transactions) can be spent.
	CoinbaseMaturity uint16

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

	// Mempool parameters
	RelayNonStdTxs bool
}

var pposValidators = []string{
	"GDWWWkXmfmmQhyQar7gnRnLTSz5VxVNnuv7F4342AUiH",
	"BA37mxoKiHuZ2xndjeW7xhW7Qu5ZZcceTCC6LrHh5PUW",
	"746J3bxx7kCz7mpijYP1ycTRwRoRTLZJHTseQcbs5deH",
	"6Ed3AiSZvZW9a8Upmy93Y5pAL7J3X9ksrzdwDSb1Xx6u",
	"7dD9SNgG5DEKRK7BS85VkjodD7wzwmE7eD2dx8dcP7rF",
	"4RLzEfiorGa5FtrHpQhm5GiZtHLhGDPqGQUur3aN9zkp",
	"DbSUGygJwnLjBAZpoPF41k2vVPbQ1FCCvghykaeeN848",
	"5kJGVbVKQ8tCRPEVPiveLdFYeHoBrkSkWE8D68ATYANS",
	"B8PFWDRQmwzaUYzy6vWj1ue1bvfPZJxN4g6gm4Wz69kd",
	"5QAurpbAkyyw7HM6QdwS9toZhSEtzg3CpusjNf73hHsr",
	"61kRpo51pPhJigi7Ho2CmWk4q7D33YcB8FptQEpozj8H",
	"EzYivEfPChfM6cwWVb3HfU8t2whoiYgxCZpY17wPNEeH",
	"HPupQn7h31b2bTpzYadThvUnLAUVhfrahYojqpDtxZi3",
	"CVGFyAC2d2u7sC4aTU8ZxKHpKnGbz3cdstJW3qaLGbwV",
	"4BkwAhbp5KVdcKQBgEmDeZXimmrYxw6sM1bCHCQuVyNh",
	"FvLj6b6LKa7ws9M78PxBu5CFUt2RySDvWq4nn7mkQdKK",
	"2r62ongZ7ZPXNqFUA9yxxKrkDnnFi5ieB7mGtMeY4Jzk",
	"2koT283THFvGwDcFzw2w92gQBDEnmhNrxYc3R3s6w649",
	"HAoXHKyev1nE1jg2WM73XFrnWGTLGDnQMK5S5BCNTbiK",
	"CJnuLLdRduaTFbkuxcg1ZRGtC7Z8sU6yk45NKMetN3p5",
}

// MainNetParams defines the network parameters for the main coin network.
var MainNetParams = Params{
	Name:        MAINNET_NAME,
	Net:         MAINNET,
	DefaultPort: MAINET_DEFAULT_PORT,
	DNSSeeds:    []string{
		/*{"seed.coin.sipa.be", true},
		{"dnsseed.bluematt.me", true},
		{"dnsseed.coin.dashjr.org", false},
		{"seed.coinstats.com", true},
		{"seed.bitnodes.io", false},
		{"seed.coin.jonasschnelli.ch", true},*/
		//"/ip4/127.0.0.1/tcp/9333/ipfs/QmRuvXN7BpTqxqpPLSftDFbKEYiRZRUb7iqZJcz2CxFvVS",
	},

	// BlockChain parameters
	GenesisBlock:             GenesisBlockGenerator{}.CreateGenesisBlockPoSParallel(time.Date(2018, 8, 1, 0, 0, 0, 0, time.Local), 0x18aea41a, 0x1d00ffff, 1, GENESIS_BLOCK_REWARD, pposValidators),
	PowLimit:                 mainPowLimit,
	PowLimitBits:             486604799,
	CoinbaseMaturity:         100,
	SubsidyReductionInterval: 210000,
	TargetTimespan:           time.Hour * 24 * 14, // 14 days
	TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
	RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0,
	GenerateSupported:        false,

	// Mempool parameters
	RelayNonStdTxs: false,
}
