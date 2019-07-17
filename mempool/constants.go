package mempool

const (
	// UnminedHeight is the height used for the "block" height field of the
	// contextual transaction information provided in a transaction store
	// when it has not yet been mined into a block.
	UnminedHeight = 0x7fffffffffffffff
	MaxVersion    = 1
)

// Beacon pool
const (
	MAX_VALID_BEACON_BLK_IN_POOL   = 10000
	MAX_PENDING_BEACON_BLK_IN_POOL = 10000
	BEACON_CACHE_SIZE              = 2000
	BEACON_POOL_MAIN_LOOP_TIME     = 500 // count in milisecond
)

// Shard to beacon pool
const (
	MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL   = 1000
	MAX_INVALID_SHARD_TO_BEACON_BLK_IN_POOL = 2000
)

// Shard pool
const (
	MAX_VALID_SHARD_BLK_IN_POOL   = 10000
	MAX_PENDING_SHARD_BLK_IN_POOL = 10000
	SHARD_CACHE_SIZE              = 2000
	SHARD_POOL_MAIN_LOOP_TIME     = 500 // count in milisecond
)

// Cross Shard Pool
const (
	MAX_VALID_CROSS_SHARD_IN_POOL   = 1000
	MAX_PENDING_CROSS_SHARD_IN_POOL = 2000 //per shardID
	
	VALID_CROSS_SHARD_BLOCK   = 0
	INVALID_CROSS_SHARD_BLOCK = -1
	PENDING_CROSS_SHARD_BLOCK = -2
)
