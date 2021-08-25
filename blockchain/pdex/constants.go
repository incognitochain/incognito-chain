package pdex

const (
	BasicVersion = iota + 1
	AmplifierVersion
)

// params
const (
	MaxFeeRateBPS         = 200
	MaxPRVDiscountPercent = 75
)

// nft hash prefix
var (
	hashPrefix = []byte("pdex-v3")
)

const (
	addOperator = byte(0)
	subOperator = byte(1)
)
