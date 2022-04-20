package pdexv3

const (
	BaseAmplifier                      = 10000
	MaxPoolPairWithdrawalReceiver      = 5
	MaxStakingRewardWithdrawalReceiver = 15
)

const (
	RequestAcceptedChainStatus = "accepted"
	RequestRejectedChainStatus = "rejected"

	ParamsModifyingFailedStatus  = 0
	ParamsModifyingSuccessStatus = 1

	WithdrawLPFeeFailedStatus  = 0
	WithdrawLPFeeSuccessStatus = 1

	WithdrawProtocolFeeFailedStatus  = 0
	WithdrawProtocolFeeSuccessStatus = 1

	WithdrawStakingRewardFailedStatus  = 0
	WithdrawStakingRewardSuccessStatus = 1
)

// trade status
const (
	TradeAcceptedStatus         = 1
	TradeRefundedStatus         = 0
	OrderAcceptedStatus         = 1
	OrderRefundedStatus         = 0
	WithdrawOrderAcceptedStatus = 1
	WithdrawOrderRejectedStatus = 0

	MaxTradePathLength = 5
	MinPdexv3AccessBurn = 1
)
