package mempool

import (
	"errors"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"sort"
	"sync"
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
// Cross shard process: process of including output coin from other shard into the shard block

// Whenever new beacon best state is updated, we should validate pending pool (check order)
// Whenever new cross shard block receive, validate it, and also validate pending pool (check order)

type CrossShardPool_v2 struct {
	shardID     byte
	validPool   map[byte][]*blockchain.CrossShardBlock
	pendingPool map[byte][]*blockchain.CrossShardBlock
	poolMu      *sync.RWMutex
}

func GetCrossShardPool(shardID byte) *CrossShardPool_v2 {
	var crossShardPool *CrossShardPool_v2
	if crossShardPool == nil {
		crossShardPool = new(CrossShardPool_v2)
		crossShardPool.shardID = shardID

		crossShardPool.validPool = make(map[byte][]*blockchain.CrossShardBlock)
		crossShardPool.pendingPool = make(map[byte][]*blockchain.CrossShardBlock)
		crossShardPool.poolMu = new(sync.RWMutex)

		//go func() {
		//	for {
		//		select {
		//
		//		case newBestState := <-crossShardPool.beaconBestStateUpdateCh: // when receive signal of updating beacon state
		//			crossShardPool.curBestStateBeacon = newBestState
		//			crossShardPool.UpdateValidBlock()
		//
		//		}
		//	}
		//}()
	}

	return crossShardPool
}

////calculate waitingCrossShardBlock from beacon info
//func (self *CrossShardPool_v2) CalculateWaitingCrossShardBlock(crossShardProcessState map[byte]uint64) error {
//	return nil
//}

//TODO: When start node, we should get cross shard process to know which block we should store in pool & include in the shard block
//TODO: When shard block is inserted to chain, we should update the cross shard process -> to validate block
func (self *CrossShardPool_v2) SetCrossShardProcessState(crossShardProcessState map[byte]uint64) error {

	self.RemoveValidBlockByHeight(crossShardProcessState)
	return nil
}

//func (self *CrossShardPool_v2) RemoveWaitingCrossShardBlock(crossShardProcessState map[byte]uint64) error {
//	//remove waiting cross shard block hash
//	return nil
//}

// Validate pending pool again, to move pending block to valid block
// When receive new cross shard block or new beacon state arrive
func (self *CrossShardPool_v2) UpdateValidBlock() error {
	//TODO: check cross shard block that in order (follow waitingCrossShardBlock)
	//TODO: traverse pending block and validate it

	return nil
}

func (self *CrossShardPool_v2) AddCrossShardBlock(blk blockchain.CrossShardBlock) error {
	self.poolMu.Lock()
	defer self.poolMu.Unlock()

	shardID := blk.Header.ShardID

	if shardID != self.shardID {
		return errors.New("This pool cannot receive this cross shard block, this block for another shard")
	}

	shouldStore := blk.ShouldStoreBlock()

	if shouldStore {

		if len(self.pendingPool[shardID]) > MAX_PENDING_CROSS_SHARD_IN_POOL {
			//TODO: swap for better block
			return errors.New("Reach max pending cross shard block")
		}

		self.pendingPool[shardID] = append(self.pendingPool[shardID], &blk)

		sort.Slice(self.pendingPool[shardID], func(i, j int) bool {
			return self.pendingPool[shardID][i].Header.Height < self.pendingPool[shardID][j].Header.Height
		})

	} else {
		return errors.New("Block is not valid")
	}

	self.UpdateValidBlock()
	return nil
}

func (self *CrossShardPool_v2) RemoveValidBlockByHeight(removeSinceBlkHeight map[byte]uint64) error {
	self.poolMu.Lock()
	defer self.poolMu.Unlock()

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
