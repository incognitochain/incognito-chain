package pdex

import (
	"math"

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
	InitFeeRateBPS               = 30
	MaxFeeRateBPS                = 200
	InitPRVDiscountPercent       = 25
	MaxPRVDiscountPercent        = 75
	InitProtocolFeePercent       = 0
	InitStakingPoolRewardPercent = 10
	InitStakingPoolsShare        = 0
)

// PDEX token
const (
	GenesisMintingAmount = 5000000                  // without mulitply with denominating rate
	MintingBlockReward   = 45000000                 // without multiply with denominating rate
	MintingBlocks        = 3600 * 24 * 30 * 60 / 40 // 60 months
	DecayIntervals       = 30
	DecayRateBPS         = 500 // 5%
)

var (
	hashPrefix = []byte("pdex-v3")

	TotalPDEXReward         = MintingBlockReward * math.Pow(10, common.PDEXDenominatingDecimal)
	DecayRate               = float64(DecayRateBPS) / float64(BPS)
	PDEXRewardFirstInterval = uint64(TotalPDEXReward * DecayRate / (1 - math.Pow(1-DecayRate, DecayIntervals)))
)
