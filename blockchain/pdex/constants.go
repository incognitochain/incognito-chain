package pdex

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
	GenesisMintingAmount = 5000000 // without mulitply with denominating rate
	DecayIntervals       = 30
	DecayRateBPS         = 50000 // 5%

)
