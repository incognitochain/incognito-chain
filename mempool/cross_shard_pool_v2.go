package mempool

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"sync"

	"github.com/incognitochain/incognito-chain/database"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
)

/*
	Cross Shard pool served as pool for cross shard block from other shard
	Each node has 256 cross shard pool, each cross shard pool has an id (shardID) (each pool responsible for one particular shard corresponding with its shardID)
	Cross Shard pool only receive cross shard block when
	- Block contain valid signature along with some condition below (AddCrossShardBlock)
	Cross Shard Pool Contains:
	- Valid pool: ordered cross shard block for each shard
	- Pending pool: un-ordered cross shard block for each shard
	- CrossShardState: highest cross shard block height confirmed by beacon committee
	Whenever new beacon best state is updated, we should validate pending pool (check order)
	Whenever new cross shard block receive, validate it, and also validate pending pool (check order)

	Ex: Cross Shard Pool with Shard ID 0
	Valid Pool:
	- contain cross shard block sent to shard 0 (any shard except shard 0 can send cross shard block to shard 0)
	- valid cross shard block map contains 256 ordered list, each list is available to be processed
	Pending Pool:
	- contain cross shard block sent to shard 0 (any shard except shard 0 can send cross shard block to shard 0)
	- pending cross shard block map contain 256 un-ordered list, each list still not available yet to be processed
	- pending cross shard block will enter valid pool until
	 + Beacon state confirm the next valid cross shard block height
*/
type CrossShardPool_v2 struct {
	shardID         byte                                   // pool shard ID
	validPool       map[byte][]*blockchain.CrossShardBlock // cross shard block from all other shard to this shard
	pendingPool     map[byte][]*blockchain.CrossShardBlock // cross shard block from all other shard to this shard
	crossShardState map[byte]uint64                        // cross shard state (marked the current state of cross shard block from all shard)
	mtx             *sync.RWMutex
	db              database.DatabaseInterface
	// When beacon chain confirm new cross shard block, it will store these block height in database
	// Cross Shard Pool using database to detect either is valid or pending
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
		p.mtx = new(sync.RWMutex)
		crossShardPoolMap[shardID] = p
	}
	return p
}

// Validate pending pool again, to move pending block to valid block
// When receive new cross shard block or new beacon state arrive
func (pool *CrossShardPool_v2) UpdatePool() map[byte]uint64 {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()
	expectedHeight := pool.updatePool()
	return expectedHeight
}

func (pool *CrossShardPool_v2) GetNextCrossShardHeight(fromShard, toShard byte, startHeight uint64) uint64 {
	nextHeight, err := pool.db.FetchCrossShardNextHeight(fromShard, toShard, startHeight)
	if err != nil {
		return 0
	}
	return nextHeight

}
func (pool *CrossShardPool_v2) updatePool() map[byte]uint64 {
	pool.crossShardState = blockchain.GetBestStateShard(pool.shardID).BestCrossShard
	pool.removeBlockByHeight(pool.crossShardState)
	expectedHeight := make(map[byte]uint64)
	for blkShardID, blks := range pool.pendingPool {
		startHeight := pool.crossShardState[blkShardID]
		if len(pool.validPool[blkShardID]) > 0 {
			startHeight = pool.validPool[blkShardID][len(pool.validPool[blkShardID])-1].Header.Height
		}
		index := 0
		removeIndex := 0
		for _, blk := range blks {
			//only when beacon confirm (save next cross shard height), we make cross shard block valid
			//if waitHeight > blockHeight, remove that block
			waitHeight := pool.GetNextCrossShardHeight(blkShardID, pool.shardID, startHeight)
			if waitHeight > blk.Header.Height {
				removeIndex++
				index++
			} else if waitHeight == blk.Header.Height {
				index++
				startHeight = waitHeight
				continue
			} else {
				Logger.log.Info("crossshard next expectedHeight", waitHeight)
				expectedHeight[blkShardID] = waitHeight
				break
			}
		}
		if index > 0 || removeIndex > 0 {
			var valid []*blockchain.CrossShardBlock
			valid, pool.pendingPool[blkShardID] = pool.pendingPool[blkShardID][removeIndex:index], pool.pendingPool[blkShardID][index:]
			if len(valid) > 0 {
				pool.validPool[blkShardID] = append(pool.validPool[blkShardID], valid...)
			}
		}
	}

	//===============For log
	validPoolHeight := make(map[byte][]uint64)
	pendingPoolHeight := make(map[byte][]uint64)
	for shardID, blocks := range pool.validPool {
		for _, block := range blocks {
			validPoolHeight[shardID] = append(validPoolHeight[shardID], block.Header.Height)
		}
	}
	for shardID, blocks := range pool.pendingPool {
		for _, block := range blocks {
			pendingPoolHeight[shardID] = append(pendingPoolHeight[shardID], block.Header.Height)
		}
	}
	return expectedHeight
}

/*
	Validate Condition:
	1. Block come into exact destination shardID
	2. Greater than current pool cross shard state
	3. Duplicate block in pending or valid
	4. Signature
*/
func (pool *CrossShardPool_v2) AddCrossShardBlock(block *blockchain.CrossShardBlock) (map[byte]uint64, byte, error) {
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	shardID := block.Header.ShardID
	blockHeight := block.Header.Height

	Logger.log.Criticalf("Receiver Block %+v from shard %+v at Cross Shard Pool \n", blockHeight, shardID)
	if block.ToShardID != pool.shardID {
		return nil, pool.shardID, NewBlockPoolError(WrongShardIDError, errors.New("This pool cannot receive this cross shard block, this block for another shard "+strconv.Itoa(int(block.Header.Height))))
	}

	//If receive old block, it will ignore
	startHeight := pool.crossShardState[shardID]
	if blockHeight <= startHeight {
		return nil, pool.shardID, NewBlockPoolError(OldBlockError, errors.New("receive old block"))
	}

	//If block already in pool, it will ignore
	for _, blkItem := range pool.validPool[shardID] {
		if blkItem.Header.Height == blockHeight {
			return nil, pool.shardID, NewBlockPoolError(DuplicateBlockError, errors.New("receive duplicate block"))
		}
	}
	for _, blkItem := range pool.pendingPool[shardID] {
		if blkItem.Header.Height == blockHeight {
			return nil, pool.shardID, NewBlockPoolError(DuplicateBlockError, errors.New("receive duplicate block"))
		}
	}
	shardCommitteeByte, err := pool.db.FetchCommitteeByEpoch(block.Header.BeaconHeight)
	if err != nil {
		return nil, pool.shardID, NewBlockPoolError(DatabaseError, errors.New("No committee for this epoch"))
	}
	shardCommittee := make(map[byte][]string)
	if err := json.Unmarshal(shardCommitteeByte, &shardCommittee); err != nil {
		return nil, pool.shardID, NewBlockPoolError(UnmarshalError, errors.New("Fail to unmarshal shard committee"))
	}
	if err := blockchain.ValidateAggSignature(block.ValidatorsIdx, shardCommittee[shardID], block.AggregatedSig, block.R, block.Hash()); err != nil {
		return nil, pool.shardID, err
	}

	if len(pool.pendingPool[shardID]) > MAX_PENDING_CROSS_SHARD_IN_POOL {
		if pool.pendingPool[shardID][len(pool.pendingPool[shardID])-1].Header.Height > block.Header.Height {
			pool.pendingPool[shardID] = pool.pendingPool[shardID][:len(pool.pendingPool[shardID])-1]
		} else {
			return nil, pool.shardID, NewBlockPoolError(MaxPoolSizeError, errors.New("Reach max pending cross shard block"))
		}
	}

	pool.pendingPool[shardID] = append(pool.pendingPool[shardID], block)
	sort.Slice(pool.pendingPool[shardID], func(i, j int) bool {
		return pool.pendingPool[shardID][i].Header.Height < pool.pendingPool[shardID][j].Header.Height
	})
	Logger.log.Infof("Finish Verify Cross Shard Block %+v from shard %+v \n", blockHeight, shardID)
	expectedHeight := pool.updatePool()
	return expectedHeight, pool.shardID, nil
}

func (self *CrossShardPool_v2) RemoveBlockByHeight(removeSinceBlkHeight map[byte]uint64) {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	self.removeBlockByHeight(removeSinceBlkHeight)
}

func (self *CrossShardPool_v2) removeBlockByHeight(removeSinceBlkHeight map[byte]uint64) {
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
}

func (self *CrossShardPool_v2) GetValidBlock(limit map[byte]uint64) map[byte][]*blockchain.CrossShardBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	finalBlocks := make(map[byte][]*blockchain.CrossShardBlock)
	for shardID, blks := range self.validPool {
		for _, blk := range blks {
			if limit != nil && limit[shardID] != 0 && limit[shardID] < blk.Header.Height {
				break
			}
			finalBlocks[shardID] = append(finalBlocks[shardID], blk)
		}

	}
	return finalBlocks
}

func (self *CrossShardPool_v2) GetValidBlockHash() map[byte][]common.Hash {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	finalBlockHash := make(map[byte][]common.Hash)
	for shardID, blkItems := range self.validPool {
		for _, blk := range blkItems {
			finalBlockHash[shardID] = append(finalBlockHash[shardID], *blk.Hash())
		}
	}
	return finalBlockHash
}

func (self *CrossShardPool_v2) GetValidBlockHeight() map[byte][]uint64 {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	finalBlockHeight := make(map[byte][]uint64)
	for shardID, blkItems := range self.validPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}
	return finalBlockHeight
}

func (self *CrossShardPool_v2) GetPendingBlockHeight() map[byte][]uint64 {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	finalBlockHeight := make(map[byte][]uint64)
	for shardID, blkItems := range self.pendingPool {
		for _, blk := range blkItems {
			finalBlockHeight[shardID] = append(finalBlockHeight[shardID], blk.Header.Height)
		}
	}
	return finalBlockHeight
}

func (self *CrossShardPool_v2) GetAllBlockHeight() map[byte][]uint64 {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
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

func (self *CrossShardPool_v2) GetLatestValidBlockHeight() map[byte]uint64 {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	finalBlockHeight := make(map[byte]uint64)
	for shardID, blkItems := range self.validPool {
		if len(blkItems) > 0 {
			finalBlockHeight[shardID] = blkItems[len(blkItems)-1].Header.Height
		} else {
			finalBlockHeight[shardID] = 0
		}

	}
	return finalBlockHeight
}

func (self *CrossShardPool_v2) GetBlockByHeight(_shardID byte, height uint64) *blockchain.CrossShardBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	for shardID, blkItems := range self.validPool {
		if shardID != _shardID {
			continue
		}
		for _, blk := range blkItems {
			if blk.Header.Height == height {
				return blk
			}
		}
	}

	for shardID, blkItems := range self.pendingPool {
		if shardID != _shardID {
			continue
		}
		for _, blk := range blkItems {
			if blk.Header.Height == height {
				return blk
			}
		}
	}

	return nil
}
