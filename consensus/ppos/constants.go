package ppos

const (
	MAX_BLOCKSIZE           = 1000000 //byte
	MAX_TXSIZE              = 1000    //byte
	MAX_TXs_IN_BLOCK        = 300
	MIN_TXs                 = 10 // minium txs for block to get immediate process (meaning no wait time)
	MAX_BLOCK_WAIT_TIME     = 5  // second
	MAX_SYNC_CHAINS_TIME    = 5  // second
	CHAIN_VALIDATORS_LENGTH = 11
	TOTAL_VALIDATORS        = 20 // = TOTAL CHAINS
	DEFAULT_MINING_REWARD   = 50
)
