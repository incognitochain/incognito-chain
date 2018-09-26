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
	"4hOjrRzLoL3q7i5NYSRjFH7embnyTMOQhgj7ikNl/34=",
	"luV9cuPksZotZpOiAWkspV4kfsSjsyMQChMT03iO3TM=",
	"WfDZi+0P9h/8ui6ofNMS7se0hDxEUdcqYMWOobMxU8I=",
	"Tce7lH0wAHnDyCGSuBz5Z1bVcig9NPhqMUYEAHhgSEI=",
	"YmzDfFLmrFr6SvJAqgeh3v/F6jyRTk3HlCAutQ5W0GA=",
	"Ms+ZcC5ocEqfEiV6i/UEQ9G7W4Y/W9oUxWYApUl5uSE=",
	"ux7KOfoe2R7tBCoDXOFdSfS0KIcgQgf2oo6qkoez0yk=",
	"RoY/GhBbnlmBW5GapRBCCShye3Mf/G6Bkn529DzY6nc=",
	"lnkX2PS6NjfdYi/8pjEHHNqSqrOl7W0dhTYryt1I/zo=",
	"QV5QKQ4bT1wOqpTFjGXaZU9M4+nKRsfJy4F5+Ae0T+U=",
	"SnuJ8K0pPv59rbrua0WtdwCzKWOPk4H9xgbtKkW+2XY=",
	"z+XnuEY20KbNTkfskHaBZRikIpZMX9nSaOssmdatei4=",
	"85l9TUx76n6yAo7QcVnOED6jQEkDMIfB1rBvdJCzCdQ=",
	"qq4SOlJMXoAqb2TOU4PhV8F1Cj/6kykPFVgLc52TqUg=",
	"L1SLjEb4gDZJDCzakPQ1kuV1zh8E5yqCVaOMTGH3BYo=",
	"3a20MZszzMwSWBAOZskEiF6iyYD7q3/S9AL1pGY5kP4=",
	"G27V1uKpi3Ha/dnuO952jF8RNViFmmKyH0OUspynw4E=",
	"GhQua2kxRMUj5t3haRUYgZ+14pi/WixKZFbqA4Ksveo=",
	"8D3SIDDtpnrMqwmDI5FuJKztMfUQQrwuM9bXXlFAMeQ=",
	"p/9U84IGafEWlZKESm5gq+JY8s4Vl9esNkNCI0g2/4o=",
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
