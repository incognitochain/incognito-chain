package pdexv3

const (
	BaseAmplifier = 10000
)

const (
	RequestAcceptedChainStatus = "accepted"
	RequestRejectedChainStatus = "rejected"

	ParamsModifyingFailedStatus  = 0
	ParamsModifyingSuccessStatus = 1
)

// trade status
const (
	TradeAcceptedStatus = 1
	TradeRefundedStatus = 0
	OrderAcceptedStatus = 1
	OrderRefundedStatus = 0
	WithdrawOrderAcceptedStatus = 1
	WithdrawOrderRejectedStatus = 0

	MaxTradePathLength = 5
)
