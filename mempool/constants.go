package mempool

const (
	// UnminedHeight is the height used for the "block" height field of the
	// contextual transaction information provided in a transaction store
	// when it has not yet been mined into a block.
	UnminedHeight = 0x7fffffffffffffff
	MaxVersion    = 1
)
