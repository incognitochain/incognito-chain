package pdex

const (
	BasicVersion = iota + 1
	AmplifierVersion
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

// nft hash prefix
var (
	hashPrefix = []byte("pdex-v3")
)

const (
	addOperator = byte(0)
	subOperator = byte(1)
)
