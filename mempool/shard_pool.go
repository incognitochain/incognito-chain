package mempool

import (
	"errors"
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/common"
	"sort"
	"sync"
	"time"
)

const (
	MAX_VALID_SHARD_BLK_IN_POOL   = 10000
	MAX_PENDING_SHARD_BLK_IN_POOL = 10000
)

type ShardPool struct {
	pool              []*blockchain.ShardBlock // shardID -> height -> block
	shardID           byte
	latestValidHeight uint64
	poolMu            *sync.RWMutex
}

var shardPoolMap = make(map[byte]*ShardPool)

func init() {
	go func() {
		ticker := time.Tick(5 * time.Second)
		for _ = range ticker {
			for k, _ := range shardPoolMap {
				GetShardPool(k).RemoveBlock(blockchain.GetBestStateShard(k).ShardHeight)
			}
		}
	}()
}

func InitShardPool(pool map[byte]blockchain.ShardPool) {
	for i := 0; i < 255; i++ {
		shardPoolMap[byte(i)] = GetShardPool(byte(i))
		//update last shard height
		shardPoolMap[byte(i)].poolMu = new(sync.RWMutex)
		shardPoolMap[byte(i)].SetShardState(blockchain.GetBestStateShard(byte(i)).ShardHeight)
		pool[byte(i)] = shardPoolMap[byte(i)]

	}
}

// get singleton instance of ShardToBeacon pool
func GetShardPool(shardID byte) *ShardPool {
	if shardPoolMap[shardID] == nil {
		shardPool := new(ShardPool)
		shardPool.shardID = shardID
		shardPool.pool = []*blockchain.ShardBlock{}
		shardPool.latestValidHeight = 1
		shardPoolMap[shardID] = shardPool
	}
	return shardPoolMap[shardID]
}

func (self *ShardPool) SetShardState(lastestShardHeight uint64) {
	if self.latestValidHeight < lastestShardHeight {
		self.latestValidHeight = lastestShardHeight
	}
}

func (self *ShardPool) GetShardState() uint64 {
	self.poolMu.RLock()
	defer self.poolMu.RUnlock()
	return self.latestValidHeight
}

func (self *ShardPool) AddShardBlock(blk *blockchain.ShardBlock) error {
	//TODO: validate aggregated signature
	self.poolMu.Lock()
	defer self.poolMu.Unlock()

	blkHeight := blk.Header.Height

	//If receive old block, it will ignore
	if blkHeight <= self.latestValidHeight {
		return errors.New("receive old block")
	}

	//If block already in pool, it will ignore
	for _, blkItem := range self.pool {
		if blkItem.Header.Height == blkHeight {
			return errors.New("receive duplicate block")
		}
	}

	//Check if satisfy pool capacity (for valid and invalid)
	if len(self.pool) != 0 {
		numValidPedingBlk := int(self.latestValidHeight - self.pool[0].Header.Height)
		if numValidPedingBlk < 0 {
			numValidPedingBlk = 0
		}
		numInValidPedingBlk := len(self.pool) - numValidPedingBlk
		if numValidPedingBlk > MAX_VALID_SHARD_BLK_IN_POOL {
			return errors.New("exceed max valid pending block")
		}

		lastBlkInPool := self.pool[len(self.pool)-1]
		if numInValidPedingBlk > MAX_PENDING_SHARD_BLK_IN_POOL {
			//If invalid block is better than current invalid block
			if lastBlkInPool.Header.Height > blkHeight {
				//remove latest block and add better invalid to pool
				self.pool = self.pool[:len(self.pool)-1]
			} else {
				return errors.New("exceed invalid pending block")
			}
		}
	}

	// add to pool
	self.pool = append(self.pool, blk)

	//sort pool
	sort.Slice(self.pool, func(i, j int) bool {
		return self.pool[i].Header.Height < self.pool[j].Header.Height
	})

	//update last valid pending ShardState
	self.updateLatestShardState()
	return nil
}

func (self *ShardPool) updateLatestShardState() {
	lastHeight := self.latestValidHeight
	for i, blk := range self.pool {
		if blk.Header.Height <= lastHeight {
			continue
		}
		if lastHeight+1 != blk.Header.Height {
			break
		}
		if i == len(self.pool)-1 {
			break
		}
		if self.pool[i+1].Header.PrevBlockHash != *blk.Hash() {
			break
		}
		lastHeight = blk.Header.Height
	}
	self.latestValidHeight = lastHeight
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (self *ShardPool) RemoveBlock(lastBlockHeight uint64) {
	self.poolMu.Lock()
	defer self.poolMu.Unlock()
	self.removeBlock(lastBlockHeight)
}

func (self *ShardPool) removeBlock(lastBlockHeight uint64) {
	for index, block := range self.pool {
		if block.Header.Height <= lastBlockHeight {
			if index == len(self.pool)-1 {
				self.pool = self.pool[index+1:]
			}
			continue
		} else {
			self.pool = self.pool[index:]
			break
		}
	}
}

func (self *ShardPool) GetValidBlock() []*blockchain.ShardBlock {
	self.poolMu.RLock()
	defer self.poolMu.RUnlock()
	finalBlocks := []*blockchain.ShardBlock{}
	for _, blk := range self.pool {
		if blk.Header.Height <= blockchain.GetBestStateShard(self.shardID).ShardHeight {
			continue
		}
		if blk.Header.Height > self.latestValidHeight {
			break
		}
		finalBlocks = append(finalBlocks, blk)
	}

	return finalBlocks
}

func (self *ShardPool) GetValidBlockHash() []common.Hash {
	finalBlocks := []common.Hash{}
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = append(finalBlocks, *blk.Hash())
	}
	return finalBlocks
}

func (self *ShardPool) GetValidBlockHeight() []uint64 {
	finalBlocks := []uint64{}
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = append(finalBlocks, blk.Header.Height)
	}
	return finalBlocks
}

func (self *ShardPool) GetLatestValidBlockHeight() uint64 {
	finalBlocks := uint64(0)
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = blk.Header.Height
	}
	return finalBlocks
}

func (self *ShardPool) GetAllBlockHeight() []uint64 {
	self.poolMu.RLock()
	defer self.poolMu.RUnlock()
	finalBlocks := []uint64{}
	for _, blk := range self.pool {
		finalBlocks = append(finalBlocks, blk.Header.Height)
	}
	return finalBlocks
}

func (self *ShardPool) GetBlockByHeight(height uint64) *blockchain.ShardBlock {
	self.poolMu.RLock()
	defer self.poolMu.RUnlock()
	for _, blk := range self.pool {
		if blk.Header.Height == height {
			return blk
		}
	}
	return nil
}
