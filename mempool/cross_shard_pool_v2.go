package mempool

import (
	"errors"
	"github.com/ninjadotorg/constant/database"
	"sort"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
)

const (
	MAX_VALID_CROSS_SHARD_IN_POOL   = 20000
	MAX_PENDING_CROSS_SHARD_IN_POOL = 100 //per shardID

	VALID_CROSS_SHARD_BLOCK   = 0
	INVALID_CROSS_SHARD_BLOCK = -1
	PENDING_CROSS_SHARD_BLOCK = -2
)

// Cross shard pool only receive cross shard block when
// - we can validate block using beacon state (committee member)
// - we cannot validate block using beacon state (committee member), and beacon height is not too far from best state

// Valid pool: in-ordered cross shard block for each shard
// Pending pool: un-ordered cross shard block for each shard

// Whenever new beacon best state is updated, we should validate pending pool (check order)
// Whenever new cross shard block receive, validate it, and also validate pending pool (check order)

type CrossShardPool_v2 struct {
	shardID         byte
	validPool       map[byte][]*blockchain.CrossShardBlock
	pendingPool     map[byte][]*blockchain.CrossShardBlock
	crossShardState map[byte]uint64
	poolMu          *sync.RWMutex
	db              database.DatabaseInterface
}

var crossShardPoolMap = make(map[byte]*CrossShardPool_v2)

func InitCrossShardPool(pool map[byte]blockchain.CrossShardPool, db database.DatabaseInterface) {
	for i := 0; i < 255; i++ {
		crossShardPoolMap[byte(i)] = GetCrossShardPool(byte(i))
		pool[byte(i)] = crossShardPoolMap[byte(i)]
		crossShardPoolMap[byte(i)].db = db
	}
}

func GetCrossShardPool(shardID byte) *CrossShardPool_v2 {
	p, ok := crossShardPoolMap[shardID]
	if ok == false {
		p = new(CrossShardPool_v2)
		p.shardID = shardID
		p.validPool = make(map[byte][]*blockchain.CrossShardBlock)
		p.pendingPool = make(map[byte][]*blockchain.CrossShardBlock)
		p.poolMu = new(sync.RWMutex)
		crossShardPoolMap[shardID] = p
	}
	return p
}

// Validate pending pool again, to move pending block to valid block
// When receive new cross shard block or new beacon state arrive
func (pool *CrossShardPool_v2) UpdatePool() error {
	pool.poolMu.Lock()
	defer pool.poolMu.Unlock()
	return pool.updatePool()
}

func (pool *CrossShardPool_v2) getNextCrossShardHeight(fromShard, toShard byte, startHeight uint64) uint64 {
	nextHeight, err := pool.db.FetchCrossShardNextHeight(fromShard, toShard, startHeight)
	if err != nil {
		return 0
	}
	return nextHeight

}

func (pool *CrossShardPool_v2) updatePool() error {
	pool.crossShardState = blockchain.GetBestStateShard(pool.shardID).BestCrossShard.ShardHeight
	pool.removeBlockByHeight(pool.crossShardState)
	for blkShardID, blks := range pool.pendingPool {
		startHeight := pool.crossShardState[blkShardID]
		index := 0
		for _, blk := range blks {
			//only when beacon confirm (save next cross shard height), we make cross shard block valid
			waitHeight := pool.getNextCrossShardHeight(blkShardID, pool.shardID, startHeight)
			if waitHeight == blk.Header.Height {
				index++
				continue
			} else {
				break
			}
		}
		if index > 0 {
			var valid []*blockchain.CrossShardBlock
			valid, pool.pendingPool[blkShardID] = pool.pendingPool[blkShardID][:index], pool.pendingPool[blkShardID][index:]
			pool.validPool[blkShardID] = append(pool.validPool[blkShardID], valid...)
		}
	}
	return nil
}

func (pool *CrossShardPool_v2) AddCrossShardBlock(blk blockchain.CrossShardBlock) error {
	pool.poolMu.Lock()
	defer pool.poolMu.Unlock()

	shardID := blk.Header.ShardID
	blkHeight := blk.Header.Height

	if blk.ToShardID != pool.shardID {
		return errors.New("This pool cannot receive this cross shard block, this block for another shard")
	}

	//If receive old block, it will ignore
	startHeight := pool.crossShardState[shardID]
	if blkHeight <= startHeight {
		return errors.New("receive old block")
	}

	//If block already in pool, it will ignore
	for _, blkItem := range pool.validPool[shardID] {
		if blkItem.Header.Height == blkHeight {
			return errors.New("receive duplicate block")
		}
	}
	for _, blkItem := range pool.pendingPool[shardID] {
		if blkItem.Header.Height == blkHeight {
			return errors.New("receive duplicate block")
		}
	}

	shouldStore := blk.ShouldStoreBlock()
	if shouldStore {
		if len(pool.pendingPool[shardID]) > MAX_PENDING_CROSS_SHARD_IN_POOL {
			//TODO: swap for better block
			return errors.New("Reach max pending cross shard block")
		}
		pool.pendingPool[shardID] = append(pool.pendingPool[shardID], &blk)
		sort.Slice(pool.pendingPool[shardID], func(i, j int) bool {
			return pool.pendingPool[shardID][i].Header.Height < pool.pendingPool[shardID][j].Header.Height
		})
	} else {
		return errors.New("Block is not valid")
	}

	pool.updatePool()
	return nil
}

func (self *CrossShardPool_v2) RemoveBlockByHeight(removeSinceBlkHeight map[byte]uint64) error {
	self.poolMu.Lock()
	defer self.poolMu.Unlock()
	return self.removeBlockByHeight(removeSinceBlkHeight)
}

func (self *CrossShardPool_v2) removeBlockByHeight(removeSinceBlkHeight map[byte]uint64) error {
	for shardID, blks := range self.validPool {
		removeIndex := 0
		for _, blk := range blks {
			if blk.Header.Height <= removeSinceBlkHeight[shardID] {
				removeIndex++
				continue
			} else {
				break
			}
		}
		self.validPool[shardID] = self.validPool[shardID][removeIndex:]
	}

	for shardID, blks := range self.pendingPool {
		removeIndex := 0
		for _, blk := range blks {
			if blk.Header.Height <= removeSinceBlkHeight[shardID] {
				removeIndex++
				continue
			} else {
				break
			}
		}
		self.pendingPool[shardID] = self.pendingPool[shardID][removeIndex:]
	}
	return nil
}

func (self *CrossShardPool_v2) GetValidBlock() map[byte][]*blockchain.CrossShardBlock {
	self.poolMu.Lock()
	defer self.poolMu.Unlock()
	finalBlocks := make(map[byte][]*blockchain.CrossShardBlock)
	for shardID, _ := range self.validPool {
		finalBlocks[shardID] = self.validPool[shardID]
	}
	return finalBlocks
}

func (self *CrossShardPool_v2) GetValidBlockHash() map[byte][]common.Hash {
	finalBlockHash := make(map[byte][]common.Hash)
	for shardID, blkItems := range self.validPool {
		for _, blk := range blkItems {
			finalBlockHash[shardID] = append(finalBlockHash[shardID], *blk.Hash())
		}
	}
	return finalBlockHash
}

func (self *CrossShardPool_v2) GetValidBlockHeight() map[byte][]uint64 {
	finalBlockHeight := make(map[byte][]uint64)
	for shardID, blkItems := range self.validPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}
	return finalBlockHeight
}

func (self *CrossShardPool_v2) GetPendingBlockHeight() map[byte][]uint64 {
	finalBlockHeight := make(map[byte][]uint64)
	for shardID, blkItems := range self.pendingPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}
	return finalBlockHeight
}

func (self *CrossShardPool_v2) GetAllBlockHeight() map[byte][]uint64 {
	finalBlockHeight := make(map[byte][]uint64)

	for shardID, blkItems := range self.validPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}

	for shardID, blkItems := range self.pendingPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}
	return finalBlockHeight
}
