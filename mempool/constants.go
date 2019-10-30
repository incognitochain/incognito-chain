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
	maxValidShardToBeaconBlockInPool   = 10000
	maxInvalidShardToBeaconBlockInPool = 20000
)

// Shard pool
const (
	maxValidShardBlockInPool   = 10000
	maxPendingShardBlockInPool = 10000
	shardCacheSize             = 2000
	shardPoolMainLoopTime      = 500 * time.Millisecond // count in milisecond
)

// Cross Shard Pool
const (
	maxPendingCrossShardInPool = 2000 //per shardID
)
