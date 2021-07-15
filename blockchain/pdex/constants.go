package pdex

const (
	BasicVersion = iota + 1
	RangeProvideVersion
)

const (
	RequestAcceptedChainStatus = "accepted"
	RequestRejectedChainStatus = "rejected"

	ParamsModifyingFailedStatus  = 0
	ParamsModifyingSucceedStatus = 1
)

// params
const (
	InitFeeRateBPS               = 30
	MaxFeeRateBPS                = 200
	InitPRVDiscountPercent       = 25
	MaxPRVDiscountPercent        = 75
	InitProtocolFeePercent       = 0
	InitStakingPoolRewardPercent = 10
	DefaultStakingPoolsShare     = 0
)
