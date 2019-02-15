package mempool

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
)

const (
	MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL   = 100
	MAX_INVALID_SHARD_TO_BEACON_BLK_IN_POOL = 200
)

type ShardToBeaconPool struct {
	pool                   map[byte][]*blockchain.ShardToBeaconBlock // shardID -> height -> block
	poolMutex              *sync.RWMutex
	latestValidHeight      map[byte]uint64
	latestValidHeightMutex *sync.RWMutex
}

var shardToBeaconPool *ShardToBeaconPool = nil

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
	self.latestValidHeightMutex.Lock()
	defer self.latestValidHeightMutex.Unlock()

	for shardID, latestHeight := range latestShardState {
		if latestHeight < 1 {
			latestShardState[shardID] = 1
		}
		self.latestValidHeight[shardID] = latestShardState[shardID]
	}

	//Remove pool base on new shardstate
	self.RemovePendingBlock(latestShardState)
	self.UpdateLatestShardState()
}

func (self *ShardToBeaconPool) GetShardState() map[byte]uint64 {
	return self.latestValidHeight
}

//Add Shard to Beacon block to the pool, if it is new block and not yet in the pool, and satisfy pool capacity (for valid and invalid; also swap for better invalid block)
//#Return Param:
//#1 and #2: requested block from height to height
//#3 error
func (self *ShardToBeaconPool) AddShardToBeaconBlock(blk blockchain.ShardToBeaconBlock) (uint64, uint64, error) {
	blkShardID := blk.Header.ShardID
	blkHeight := blk.Header.Height
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
		numInValidPedingBlk := len(self.pool[blkShardID]) - numValidPedingBlk
		if numValidPedingBlk > MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL {
			return 0, 0, errors.New("exceed max valid pending block")
		}

		lastBlkInPool := self.pool[blkShardID][len(self.pool[blkShardID])-1]
		if numInValidPedingBlk > MAX_INVALID_SHARD_TO_BEACON_BLK_IN_POOL {
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
	self.pool[blkShardID] = append(self.pool[blkShardID], &blk)

	//sort pool
	sort.Slice(self.pool[blkShardID], func(i, j int) bool {
		return self.pool[blkShardID][i].Header.Height < self.pool[blkShardID][j].Header.Height
	})

	//update last valid pending ShardState
	self.UpdateLatestShardState()
	if self.pool[blkShardID][0].Header.Height > self.latestValidHeight[blkShardID] {
		offset := self.pool[blkShardID][0].Header.Height - self.latestValidHeight[blkShardID]
		if offset > MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL {
			offset = MAX_VALID_SHARD_TO_BEACON_BLK_IN_POOL
		}
		return self.latestValidHeight[blkShardID] + 1, self.latestValidHeight[blkShardID] + offset, nil
	}
	return 0, 0, nil
}
func (self *ShardToBeaconPool) UpdateLatestShardState() {
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
		fmt.Println("[[[[[[[[[[ShardToBeaconPool]]]]]]]]]]]: LastValidHeight", lastHeight)
	}
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (self *ShardToBeaconPool) RemovePendingBlock(blockItems map[byte]uint64) {
	self.poolMutex.Lock()
	defer self.poolMutex.Unlock()
	for shardID, blockHeight := range blockItems {
		for index, block := range self.pool[shardID] {
			fmt.Println("ShardToBeaconPool/Pool BEFORE Remove", block.Header.Height)
			if block.Header.Height <= blockHeight {
				fmt.Println("ShardToBeaconPool: RemovePendingBlock, Remove Shard Block", block.Header.Height)
				if index == len(self.pool[shardID])-1 {
					fmt.Println("ShardToBeaconPool: RemovePendingBlock, Update Expected Shard Height 1", block.Header.Height+1)
					self.pool[shardID] = self.pool[shardID][index+1:]
				}
				continue
			} else {
				fmt.Println("ShardToBeaconPool: RemovePendingBlock, Update Expected Shard Height 2", block.Header.Height)
				self.pool[shardID] = self.pool[shardID][index:]
				break
			}
		}
		fmt.Println("[[[[[[[[[[ShardToBeaconPool]]]]]]]]]]]: LastValidHeight", blockHeight)
	}
}

func (self *ShardToBeaconPool) GetValidPendingBlock() map[byte][]*blockchain.ShardToBeaconBlock {

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
			finalBlocks[shardID] = append(finalBlocks[shardID], blk)
		}
	}
	//UNCOMMENT FOR TESTING
	fmt.Println()
	for _, block := range finalBlocks[byte(0)] {
		fmt.Print("ShardToBeaconPool/ValidPendingBlock ")
		fmt.Printf(" %+v ", block.Header.Height)
	}
	fmt.Println()
	//==============
	return finalBlocks
}

func (self *ShardToBeaconPool) GetValidPendingBlockHash() map[byte][]common.Hash {
	finalBlocks := make(map[byte][]common.Hash)
	blks := self.GetValidPendingBlock()
	for shardID, blkItems := range blks {
		for _, blk := range blkItems {
			finalBlocks[shardID] = append(finalBlocks[shardID], *blk.Hash())
		}
	}
	return finalBlocks
}

func (self *ShardToBeaconPool) GetValidPendingBlockHeight() map[byte][]uint64 {
	finalBlocks := make(map[byte][]uint64)
	blks := self.GetValidPendingBlock()
	for shardID, blkItems := range blks {
		for _, blk := range blkItems {
			finalBlocks[shardID] = append(finalBlocks[shardID], blk.Header.Height)
		}
	}
	return finalBlocks
}
