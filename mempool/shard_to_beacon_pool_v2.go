package mempool

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
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
	mtx                    *sync.RWMutex
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
		// add to pool
		for i := 0; i < 255; i++ {
			shardID := byte(i)
			if shardToBeaconPool.pool[shardID] == nil {
				shardToBeaconPool.pool[shardID] = []*blockchain.ShardToBeaconBlock{}
			}
		}
		shardToBeaconPool.mtx = new(sync.RWMutex)
		shardToBeaconPool.latestValidHeight = make(map[byte]uint64)
		shardToBeaconPool.latestValidHeightMutex = new(sync.RWMutex)
	}
	return shardToBeaconPool
}

func (self *ShardToBeaconPool) SetShardState(latestShardState map[byte]uint64) {
	// fmt.Println("SetShardState")
	self.mtx.Lock()
	defer self.mtx.Unlock()

	self.latestValidHeightMutex.Lock()
	defer self.latestValidHeightMutex.Unlock()

	for shardID, latestHeight := range latestShardState {
		if latestHeight < 1 {
			latestShardState[shardID] = 1
		}
		self.latestValidHeight[shardID] = latestShardState[shardID]
	}
	//Remove pool base on new shardstate
	self.removeBlock(latestShardState)
	self.updateLatestShardState()
}

func (self *ShardToBeaconPool) GetShardState() map[byte]uint64 {
	return self.latestValidHeight
}
func (self *ShardToBeaconPool) checkLatestValidHeightValidity(shardID byte) {
	if self.latestValidHeight[shardID] == 0 {
		self.latestValidHeight[shardID] = 1
	}
}

/*
	Add Shard to Beacon block to the pool, if it match following condition
	1. New block to enter pool
	2. Not duplicate block in pool
	3. if block is next valid block then check max valid block in pool
		- if not full yet then push into valid block
		- if full then return error
	4. if it not next valid block then check max invalid block in pool
		- if full then check if it can replace any block in pool or not then replace if it match replacement condition
		- if not full then push into
	If block enter pool (valid or pending == invalid)
	Update pool state
	Return Param:
	#1 and #2: requested block from height to height
	#3 error
*/

func (self *ShardToBeaconPool) AddShardToBeaconBlock(block *blockchain.ShardToBeaconBlock) (uint64, uint64, error) {
	shardID := block.Header.ShardID
	blockHeight := block.Header.Height
	Logger.log.Infof("Add ShardToBeaconBlock from shard %+v, height %+v \n", shardID, blockHeight)
	self.mtx.Lock()
	defer self.mtx.Unlock()
	self.latestValidHeightMutex.Lock()
	defer self.latestValidHeightMutex.Unlock()
	
	self.checkLatestValidHeightValidity(shardID)
	//If receive old block, it will ignore
	if blockHeight <= self.latestValidHeight[shardID] {
		return 0, 0, NewBlockPoolError(OldBlockError, errors.New("Receive block " + strconv.Itoa(int(blockHeight))+ " but expect greater than "+ strconv.Itoa(int(self.latestValidHeight[shardID]))))
	}
	//If block already in pool, it will ignore
	for _, blkItem := range self.pool[shardID] {
		if blkItem.Header.Height == blockHeight {
			return 0, 0, NewBlockPoolError(DuplicateBlockError, errors.New("Receive Duplicate block " + strconv.Itoa(int(blockHeight))))
		}
	}
	//Check if satisfy pool capacity (for valid and invalid)
	if len(self.pool[shardID]) != 0 {
		numValidPedingBlk := int(self.latestValidHeight[shardID] - self.pool[shardID][0].Header.Height + 1)
		if numValidPedingBlk < 0 {
			numValidPedingBlk = 0
		}
		numInValidPedingBlk := len(self.pool[shardID]) - numValidPedingBlk + 1
		if numValidPedingBlk >= MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL {
			return 0, 0, NewBlockPoolError(MaxPoolSizeError, errors.New("exceed max valid block"))
		}
		lastBlockInPool := self.pool[shardID][len(self.pool[shardID])-1]
		if numInValidPedingBlk >= MAX_INVALID_SHARD_TO_BEACON_BLK_IN_POOL {
			//If invalid block is better than current invalid block
			if lastBlockInPool.Header.Height > blockHeight {
				//remove latest block and add better invalid to pool
				self.pool[shardID] = self.pool[shardID][:len(self.pool[shardID])-1]
			} else {
				return 0, 0,  NewBlockPoolError(MaxPoolSizeError,errors.New("exceed invalid pending block"))
			}
		}
	}
	self.pool[shardID] = append(self.pool[shardID], block)
	//sort pool
	sort.Slice(self.pool[shardID], func(i, j int) bool {
		return self.pool[shardID][i].Header.Height < self.pool[shardID][j].Header.Height
	})
	//update last valid pending ShardState
	self.updateLatestShardState()
	//@NOTICE: check logic again
	if self.pool[shardID][0].Header.Height > self.latestValidHeight[shardID] {
		offset := self.pool[shardID][0].Header.Height - self.latestValidHeight[shardID]
		if offset > MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL {
			offset = MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL
		}
		return self.latestValidHeight[shardID] + 1, self.latestValidHeight[shardID] + offset, nil
	}
	return 0, 0, nil
}

func (self *ShardToBeaconPool) updateLatestShardState() {
	for shardID, blks := range self.pool {
		self.checkLatestValidHeightValidity(shardID)
		lastHeight := self.latestValidHeight[shardID]
		for _, blk := range blks {
			// if block height is greater than lastHeight 2 value then break
			if blk.Header.Height > lastHeight && blk.Header.Height != lastHeight+1 {
				break
			}
			// if block height is greater than lastHeight only 1 value than set new lastHeight to block height
			lastHeight = blk.Header.Height
		}
		self.latestValidHeight[shardID] = lastHeight
		Logger.log.Infof("ShardToBeaconPool: Updated/LastValidHeight %+v of Shard %+v \n", lastHeight, shardID)
	}
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (self *ShardToBeaconPool) RemoveBlock(blockItems map[byte]uint64) {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	self.removeBlock(blockItems)
}

func (self *ShardToBeaconPool) removeBlock(blockItems map[byte]uint64) {
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
		Logger.log.Infof("ShardToBeaconPool: Removed/LastValidHeight %+v of shard %+v \n", blockHeight, shardID)
	}
}

func (self *ShardToBeaconPool) GetValidBlock(limit map[byte]uint64) map[byte][]*blockchain.ShardToBeaconBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	self.latestValidHeightMutex.Lock()
	defer self.latestValidHeightMutex.Unlock()
	finalBlocks := make(map[byte][]*blockchain.ShardToBeaconBlock)
	for shardID, blks := range self.pool {
		self.checkLatestValidHeightValidity(shardID)
		for i, blk := range blks {
			if blks[i].Header.Height > self.latestValidHeight[shardID] {
				break
			}
			// ?
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

func (self *ShardToBeaconPool) GetValidBlockHash() map[byte][]common.Hash {
	finalBlocks := make(map[byte][]common.Hash)
	blks := self.GetValidBlock(nil)
	for shardID, blkItems := range blks {
		for _, blk := range blkItems {
			finalBlocks[shardID] = append(finalBlocks[shardID], *blk.Hash())
		}
	}
	return finalBlocks
}

func (self *ShardToBeaconPool) GetValidBlockHeight() map[byte][]uint64 {
	finalBlocks := make(map[byte][]uint64)
	blks := self.GetValidBlock(nil)
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
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	finalBlocks := make(map[byte][]uint64)
	for shardID, blks := range self.pool {
		for _, blk := range blks {
			finalBlocks[shardID] = append(finalBlocks[shardID], blk.Header.Height)
		}
	}
	return finalBlocks
}

func (self *ShardToBeaconPool) GetBlockByHeight(shardID byte, height uint64) *blockchain.ShardToBeaconBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
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
