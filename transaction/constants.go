package transaction

const (
	// txVersion is the current latest supported transaction version.
	txVersion = 1
)

const (
	CustomTokenInit = iota
	CustomTokenTransfer
	CustomTokenCrossShard
)

const (
	NormalCoinType = iota
	CustomTokenType
	CustomTokenPrivacyType
)
