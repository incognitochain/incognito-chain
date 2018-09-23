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
	"bibqG4qhHBSvoWjTDJ4SUSHsxR1yrzL5Cq63CVAokfA=",
	"gOUNRnLIfZrk277zA46gYC99iM5y0U4y0deCWwycFag=",
	"CuIsMnls26Th8TajOimsNs+yTGawtb/3pVQM1xxvXgE=",
	"Sii/5l/rTabKHE9rj9NnAySNqrp1RzTa9uFP2dGG8ow=",
	"gHg6jPFCrdGG5klnbV3LA4utDINJlSJ9DurvywAdGZU=",
	"AbQii9GKwDJSe6s9w/Mlv2h4VkGmtwZ3CPJBjwxdW4A=",
	"DAMnkqVd0PX7Q+E5r2owGQlGVjdk0bG/M87ffqnvULA=",
	"KpYlKd7UmsZ9iaE/qY6u/7MmhCEAW3CVoHbKR2ZmhWM=",
	"y1oFkksv9qIPxdeomroU31VxqSfKe1Gre97k0BSOQZU=",
	"aZsJNkBlp3IOCBjS4t79rSeojlsnnMFyeAJQgtGhWko=",
	"RUBAsgpgxpQw1zADW5ZnzqJQPatDgaXbNRm1NNM92Ss=",
	"clAV3X7SuDS58Zkud4fySEsaj+BSdB8P9SeHt5N5AY8=",
	"GzYpI8XPUHmI+0zfiWufbn/hilpd2DEz+Wn9n+R5/4c=",
	"9JH5TjwrL1IdgKUa1a4n2KqkCblr8Fg75qUUDB+hbYg=",
	"AFEYlD3onHq0PUpKq/1Ib4AYvabzaCCSa/eyVLqqJXc=",
	"ZoE1MbR4QFq6HhCA3nrgIbOSLHVZBEzFNy6XNYOpqaM=",
	"tEHomcoX3Zv4wyM3QJ43Epv7Mhc/PBsb4Hcm0t1AMBc=",
	"mDZGD/hH3Qe3J/TFvp47ehNu+ZT2WHhpitUjR9/VQCs=",
	"89AbTl86qYLXq5d/5s+tCMq6RYmZdAuGBW4oiTShz2E=",
	"IOgarHxmJzVKLYDb/vciPlPx4MuadMAl+Yj0aDRbU/U=",
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
	GenesisBlock: GenesisBlockGenerator{}.CreateGenesisBlockPoSParallel(time.Date(2018, 8, 1, 0, 0, 0, 0, time.Local), 0x18aea41a, 0x1d00ffff, 1, GENESIS_BLOCK_REWARD, pposValidators),
	PowLimit:     mainPowLimit,
	PowLimitBits: 486604799,
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
