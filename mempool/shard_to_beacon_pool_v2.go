package mempool

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
)

/*
	ShardToBeaconPool contains block from all shard which is sent to beacon
	Whenever any shard block from any shard is created, shard node will create a corresponding block called shard to beacon block
	This block contains new state of shard, beacon use this block to update shard state globally
	ShardToBeacon block must be inserted to beacon chain as increased order by block height
	ShardToBeaconPool maintains block from all shard in this order, ex:
		- Shard 0: blockHeight 2, blockHeight 3, blockHeight 5
	ShardToBeaconPool has LatestValidHeight for each shard, this will be used as the pointer to
	record how many valid blocks (which can be inserted into beacon chain) can be get from this pool, ex:
		- Shard 0:
			+ Pool: blockHeight 2, blockHeight 3, blockHeight 5
			+ LatestvalidHeight: 3
			=> only get block with height equal or less than 3 (blockHeight 2 and 3) from pool,
			blockHeight 5 will be remained in pool until LatestvalidHeight is equal or greater than 5
*/
// Using Singleton object pattern
// shardToBeaconPool object will be used globally in this application
var shardToBeaconPool *ShardToBeaconPool = nil

type ShardToBeaconPool struct {
	pool                   map[byte][]*blockchain.ShardToBeaconBlock // shardID -> height -> block
	poolMutex              *sync.RWMutex
	latestValidHeight      map[byte]uint64
	latestValidHeightMutex *sync.RWMutex
}

func InitShardToBeaconPool() {
	GetShardToBeaconPool().SetShardState(blockchain.GetBestStateBeacon().GetBestShardHeight())
}

// get singleton instance of ShardToBeacon pool
func GetShardToBeaconPool() *ShardToBeaconPool {
	if shardToBeaconPool == nil {
		shardToBeaconPool = new(ShardToBeaconPool)
		shardToBeaconPool.pool = make(map[byte][]*blockchain.ShardToBeaconBlock)
		shardToBeaconPool.poolMutex = new(sync.RWMutex)
		shardToBeaconPool.latestValidHeight = make(map[byte]uint64)
		shardToBeaconPool.latestValidHeightMutex = new(sync.RWMutex)
	}
	return shardToBeaconPool
}

func (self *ShardToBeaconPool) SetShardState(latestShardState map[byte]uint64) {
	// fmt.Println("SetShardState")
	self.poolMutex.Lock()
	defer self.poolMutex.Unlock()

	self.latestValidHeightMutex.Lock()
	defer self.latestValidHeightMutex.Unlock()

	for shardID, latestHeight := range latestShardState {
		if latestHeight < 1 {
			latestShardState[shardID] = 1
		}
		self.latestValidHeight[shardID] = latestShardState[shardID]
	}
	// fmt.Println("SetShardState 1")
	//Remove pool base on new shardstate
	self.removePendingBlock(latestShardState)
	// fmt.Println("SetShardState 2")
	self.updateLatestShardState()
	// fmt.Println("SetShardState 3")
}

func (self *ShardToBeaconPool) GetShardState() map[byte]uint64 {
	return self.latestValidHeight
}

//Add Shard to Beacon block to the pool, if it is new block and not yet in the pool, and satisfy pool capacity (for valid and invalid; also swap for better invalid block)
//#Return Param:
//#1 and #2: requested block from height to height
//#3 error
func (self *ShardToBeaconPool) AddShardToBeaconBlock(blk *blockchain.ShardToBeaconBlock) (uint64, uint64, error) {

	blkShardID := blk.Header.ShardID
	blkHeight := blk.Header.Height

	fmt.Println("AddShardToBeaconBlock ", blkShardID, blkHeight)
	self.poolMutex.Lock()
	self.latestValidHeightMutex.Lock()

	defer self.poolMutex.Unlock()
	defer self.latestValidHeightMutex.Unlock()

	if self.latestValidHeight[blkShardID] == 0 {
		self.latestValidHeight[blkShardID] = 1
	}

	//If receive old block, it will ignore
	if blkHeight <= self.latestValidHeight[blkShardID] {
		return 0, 0, errors.New("receive old block")
	}

	//If block already in pool, it will ignore
	for _, blkItem := range self.pool[blkShardID] {
		if blkItem.Header.Height == blkHeight {
			return 0, 0, errors.New("receive duplicate block")
		}
	}

	//Check if satisfy pool capacity (for valid and invalid)
	if len(self.pool[blkShardID]) != 0 {
		numValidPedingBlk := int(self.latestValidHeight[blkShardID] - self.pool[blkShardID][0].Header.Height)
		if numValidPedingBlk < 0 {
			numValidPedingBlk = 0
		}
		numInValidPedingBlk := len(self.pool[blkShardID]) - numValidPedingBlk

		if numValidPedingBlk > MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL {
			// fmt.Println("exceed shard2beacon 1", blkShardID, numValidPedingBlk, numInValidPedingBlk, self.latestValidHeight[blkShardID], self.pool[blkShardID][0].Header.Height)
			return 0, 0, errors.New("exceed max valid pending block")
		}

		lastBlkInPool := self.pool[blkShardID][len(self.pool[blkShardID])-1]
		if numInValidPedingBlk > MAX_INVALID_SHARD_TO_BEACON_BLK_IN_POOL {
			// fmt.Println("exceed shard2beacon 2", blkShardID, numValidPedingBlk, numInValidPedingBlk, self.latestValidHeight[blkShardID], self.pool[blkShardID][0].Header.Height)
			//If invalid block is better than current invalid block
			if lastBlkInPool.Header.Height > blkHeight {
				//remove latest block and add better invalid to pool
				self.pool[blkShardID] = self.pool[blkShardID][:len(self.pool[blkShardID])-1]
			} else {

				return 0, 0, errors.New("exceed invalid pending block")
			}
		}
	}

	// add to pool
	if self.pool[blkShardID] == nil {
		self.pool[blkShardID] = []*blockchain.ShardToBeaconBlock{}
	}
	self.pool[blkShardID] = append(self.pool[blkShardID], blk)

	//sort pool
	sort.Slice(self.pool[blkShardID], func(i, j int) bool {
		return self.pool[blkShardID][i].Header.Height < self.pool[blkShardID][j].Header.Height
	})

	//update last valid pending ShardState
	self.updateLatestShardState()
	if self.pool[blkShardID][0].Header.Height > self.latestValidHeight[blkShardID] {
		offset := self.pool[blkShardID][0].Header.Height - self.latestValidHeight[blkShardID]
		if offset > MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL {
			offset = MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL
		}
		return self.latestValidHeight[blkShardID] + 1, self.latestValidHeight[blkShardID] + offset, nil
	}
	return 0, 0, nil
}

func (self *ShardToBeaconPool) updateLatestShardState() {
	for shardID, blks := range self.pool {
		if self.latestValidHeight[shardID] == 0 {
			self.latestValidHeight[shardID] = 1
		}
		lastHeight := self.latestValidHeight[shardID]
		for i, blk := range blks {
			if blks[i].Header.Height > lastHeight && lastHeight+1 != blk.Header.Height {
				break
			}
			lastHeight = blk.Header.Height
		}
		self.latestValidHeight[shardID] = lastHeight
		fmt.Printf("ShardToBeaconPool: Updated/LastValidHeight %+v of Shard %+v \n", lastHeight, shardID)
	}
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (self *ShardToBeaconPool) RemovePendingBlock(blockItems map[byte]uint64) {
	self.poolMutex.Lock()
	defer self.poolMutex.Unlock()
	self.removePendingBlock(blockItems)
}

func (self *ShardToBeaconPool) removePendingBlock(blockItems map[byte]uint64) {

	for shardID, blockHeight := range blockItems {
		for index, block := range self.pool[shardID] {
			fmt.Println("ShardToBeaconPool/Pool BEFORE Remove", block.Header.Height)
			if block.Header.Height <= blockHeight {
				if index == len(self.pool[shardID])-1 {
					self.pool[shardID] = self.pool[shardID][index+1:]
				}
				continue
			} else {
				self.pool[shardID] = self.pool[shardID][index:]
				break
			}
		}
		fmt.Printf("ShardToBeaconPool: Removed/LastValidHeight %+v of shard %+v \n", blockHeight, shardID)
	}
}

func (self *ShardToBeaconPool) GetValidPendingBlock(limit map[byte]uint64) map[byte][]*blockchain.ShardToBeaconBlock {

	self.poolMutex.RLock()
	defer self.poolMutex.RUnlock()

	self.latestValidHeightMutex.Lock()
	defer self.latestValidHeightMutex.Unlock()

	finalBlocks := make(map[byte][]*blockchain.ShardToBeaconBlock)
	for shardID, blks := range self.pool {
		if self.latestValidHeight[shardID] == 0 {
			self.latestValidHeight[shardID] = 1
		}
		for i, blk := range blks {
			if blks[i].Header.Height > self.latestValidHeight[shardID] {
				break
			}
			if i >= 50 {
				break
			}
			if limit != nil && limit[shardID] != 0 && limit[shardID] < blks[i].Header.Height {
				break
			}
			finalBlocks[shardID] = append(finalBlocks[shardID], blk)
		}
	}
	//UNCOMMENT FOR TESTING
	// fmt.Println()
	// fmt.Print("ShardToBeaconPool/ValidPendingBlock ")
	// for _, block := range finalBlocks[byte(0)] {
	// 	fmt.Printf(" %+v ", block.Header.Height)
	// }
	// fmt.Println()
	//==============

	return finalBlocks
}

func (self *ShardToBeaconPool) GetValidPendingBlockHash() map[byte][]common.Hash {
	finalBlocks := make(map[byte][]common.Hash)
	blks := self.GetValidPendingBlock(nil)
	for shardID, blkItems := range blks {
		for _, blk := range blkItems {
			finalBlocks[shardID] = append(finalBlocks[shardID], *blk.Hash())
		}
	}
	return finalBlocks
}

func (self *ShardToBeaconPool) GetValidPendingBlockHeight() map[byte][]uint64 {
	finalBlocks := make(map[byte][]uint64)
	blks := self.GetValidPendingBlock(nil)
	for shardID, blkItems := range blks {
		for _, blk := range blkItems {
			finalBlocks[shardID] = append(finalBlocks[shardID], blk.Header.Height)
		}
	}
	return finalBlocks
}

func (self *ShardToBeaconPool) GetLatestValidPendingBlockHeight() map[byte]uint64 {
	finalBlocks := make(map[byte]uint64)
	self.latestValidHeightMutex.Lock()
	for shardID, height := range self.latestValidHeight {
		finalBlocks[shardID] = height
	}
	self.latestValidHeightMutex.Unlock()
	return finalBlocks
}

func (self *ShardToBeaconPool) GetAllBlockHeight() map[byte][]uint64 {
	self.poolMutex.RLock()
	defer self.poolMutex.RUnlock()
	finalBlocks := make(map[byte][]uint64)
	for shardID, blks := range self.pool {
		for _, blk := range blks {
			finalBlocks[shardID] = append(finalBlocks[shardID], blk.Header.Height)
		}
	}
	return finalBlocks
}

func (self *ShardToBeaconPool) GetBlockByHeight(shardID byte, height uint64) *blockchain.ShardToBeaconBlock {
	self.poolMutex.RLock()
	defer self.poolMutex.RUnlock()
	for _shardID, blks := range self.pool {
		if _shardID != shardID {
			continue
		}
		for _, blk := range blks {
			if blk.Header.Height == height {
				return blk
			}
		}
	}
	return nil
}
