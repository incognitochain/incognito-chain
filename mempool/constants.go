package mempool

const (
	// UnminedHeight is the height used for the "block" height field of the
	// contextual transaction information provided in a transaction store
	// when it has not yet been mined into a block.
	UnminedHeight = 0x7fffffff
	MaxVersion    = 1
	//count in second
	TXPOOL_SCAN_TIME = 60
)
