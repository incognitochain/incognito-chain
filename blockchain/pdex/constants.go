package pdex

import (
	"math"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	BasicVersion = iota + 1
	AmplifierVersion
)

// common
const (
	BPS = 10000
)

// params
const (
	MaxFeeRateBPS         = 200
	MaxPRVDiscountPercent = 75
)

// PDEX token
const (
	MintingBlockReward = 45000000                 // without multiply with denominating rate
	MintingBlocks      = 3600 * 24 * 30 * 60 / 40 // 60 months
	DecayIntervals     = 30
	DecayRateBPS       = 500 // 5%
)

// nft hash prefix
var (
	hashPrefix            = []byte("pdex-v3")
	inscriptionHashPrefix = []byte("inscription-")

	INSCRIPTIONS_FEATURE = "inscriptions"
	BaseLPFeesPerShare   = new(big.Int).SetUint64(1e18)

	TotalPDEXReward         = MintingBlockReward * math.Pow(10, common.PDEXDenominatingDecimal)
	DecayRate               = float64(DecayRateBPS) / float64(BPS)
	PDEXRewardFirstInterval = uint64(TotalPDEXReward * DecayRate / (1 - math.Pow(1-DecayRate, DecayIntervals)))
)

const (
	addOperator = byte(0)
	subOperator = byte(1)
)
