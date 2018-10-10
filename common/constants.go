package common

const (
	TxNormalType        = "n" // normal tx(send and receive coin)
	TxActionParamsType  = "a" // action tx to edit params
	IncMerkleTreeHeight = 29
)

// unit type use in tx
const (
	TxOutCoinType = "c" // coin
	TxOutBondType = "b" // bond
)

const (
	MAX_BLOCKSIZE           = 5000000 //byte 5MB
	MAX_TXSIZE              = 50000   //byte 50KB
	MAX_TXs_IN_BLOCK        = 1000
	MIN_TXs                 = 10                       // minium txs for block to get immediate process (meaning no wait time)
	MIN_BLOCK_WAIT_TIME     = 3                        // second
	MAX_BLOCK_WAIT_TIME     = 20 - MIN_BLOCK_WAIT_TIME // second
	MAX_SYNC_CHAINS_TIME    = 5                        // second
	MAX_BLOCKSIGN_WAIT_TIME = 20                       // second
	TOTAL_VALIDATORS        = 20                       // = TOTAL CHAINS
	TOTAL_CHAINS            = TOTAL_VALIDATORS - 1
	MINIMUM_BLOCKSIGS       = (TOTAL_VALIDATORS / 2) + 1
	DEFAULT_MINING_REWARD   = 50
	GETCHAINSTATE_INTERVAL  = 10 //second
)
