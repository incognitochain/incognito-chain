package pdexv3

const (
	BaseAmplifier                 = 10000
	MaxPoolPairWithdrawalReceiver = 5
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
)

// receiving token type
const (
	Token0Type   = "Token0"
	Token1Type   = "Token1"
	PRVType      = "PRV"
	PDEXType     = "PDEX"
	NftTokenType = "NftToken"
)

// trade status
const (
	TradeAcceptedStatus = 1
	TradeRefundedStatus = 0
	OrderAcceptedStatus = 1
	OrderRefundedStatus = 0
)
