package mempool

import "time"

const (
	// unminedHeight is the height used for the "block" height field of the
	// contextual transaction information provided in a transaction store
	// when it has not yet been mined into a block.
	unminedHeight = 0x7fffffffffffffff
	maxVersion    = 1
)

// Beacon pool
const (
	maxValidBeaconBlockInPool   = 10000
	maxPendingBeaconBlockInPool = 10000
	beaconCacheSize             = 2000
	beaconPoolMainLoopTime      = 500 * time.Millisecond // count in milisecond
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
	MAX_PENDING_CROSS_SHARD_IN_POOL = 2000 //per shardID

	/*MAX_VALID_CROSS_SHARD_IN_POOL = 1000
	VALID_CROSS_SHARD_BLOCK       = 0
	INVALID_CROSS_SHARD_BLOCK     = -1
	PENDING_CROSS_SHARD_BLOCK     = -2*/
)
