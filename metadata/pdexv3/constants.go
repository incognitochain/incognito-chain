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

// receiving token type
const (
	Token0Str    = "Token0"
	Token1Str    = "Token1"
	PRVStr       = "PRV"
	PDEXStr      = "PDEX"
	NcftTokenStr = "NcftToken"
)

// trade status
const (
	TradeAcceptedStatus = 1
	TradeRefundedStatus = 0
	OrderAcceptedStatus = 1
	OrderRefundedStatus = 0
)
