package transaction

const (
	// txVersion is the current latest supported transaction version.
	txVersion = 1
)

const (
	TokenInit = iota
	TokenTransfer
	TokenCrossShard
)

const (
	NormalCoinType = iota
	CustomTokenPrivacyType
)

const MaxSizeInfo = 512
