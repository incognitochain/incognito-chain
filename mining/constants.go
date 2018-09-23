package mining

const (
	DEFAULT_ADDRESS_FOR_BURNING      = "0x0000000000"
	NUMBER_OF_MAKING_DECISION_AGENTS = 3
	MAX_OF_MAKING_DECISION_AGENTS    = 21
	DEFAULT_COINS                    = 5
	DEFAULT_BONDS                    = 0
)

const (
	// UnminedHeight is the height used for the "block" height field of the
	// contextual transaction information provided in a transaction store
	// when it has not yet been mined into a block.
	UnminedHeight = 0x7fffffff
)
