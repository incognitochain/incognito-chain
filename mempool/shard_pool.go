package mempool

import (
	"errors"
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/common"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"
)

const (
	MAX_VALID_PENDING_BLK_IN_POOL   = 10000
	MAX_INVALID_PENDING_BLK_IN_POOL = 10000
	CACHE_SIZE = 2000
)
type ShardPoolConfig struct {
	MaxValidBlock int
	MaxInvalidBlock int
	CacheSize int
}
type ShardPool struct {
	pool              []*blockchain.ShardBlock // shardID -> height -> block
	validPool         []*blockchain.ShardBlock
	pendingPool       map[uint64]*blockchain.ShardBlock
	conflictedPool    []*blockchain.ShardBlock
	blackListPool     map[common.Hash]*blockchain.ShardBlock
	shardID           byte
	latestValidHeight uint64
	mtx               *sync.RWMutex
	config            ShardPoolConfig
	cache             *lru.Cache
}

var shardPoolMap = make(map[byte]*ShardPool)
var defaultConfig = ShardPoolConfig{
	MaxValidBlock: MAX_VALID_PENDING_BLK_IN_POOL,
	MaxInvalidBlock: MAX_INVALID_PENDING_BLK_IN_POOL,
	CacheSize: CACHE_SIZE,
}
//@NOTICE: Shard pool will always be empty when node start
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
		shardPoolMap[byte(i)].mtx = new(sync.RWMutex)
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
		shardPool.validPool = []*blockchain.ShardBlock{}
		shardPool.blackListPool = make(map[common.Hash]*blockchain.ShardBlock)
		shardPool.conflictedPool = []*blockchain.ShardBlock{}
		shardPool.config = defaultConfig
		shardPool.pendingPool = make(map[uint64]*blockchain.ShardBlock)
		shardPool.cache, _ = lru.New(shardPool.config.CacheSize)
	}
	return shardPoolMap[shardID]
}

func (self *ShardPool) SetShardState(lastestShardHeight uint64) {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	if self.latestValidHeight < lastestShardHeight {
		self.latestValidHeight = lastestShardHeight
	}
}

func (self *ShardPool) GetShardState() uint64 {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	return self.latestValidHeight
}

func (self *ShardPool) AddShardBlock(block *blockchain.ShardBlock) error {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	var err error
	err = self.ValidateShardBlock(block)
	if err != nil {
		return err
	}
	isSuccess := self.insertNewShardBlockToPool(block)
	if isSuccess {
		self.promotePendingPool()
	}
	return nil
}
func(self *ShardPool) ValidateShardBlock(block *blockchain.ShardBlock) error {
	//If receive old block, it will ignore
	if _, ok := self.cache.Get(block.Header.Hash()); ok {
		return errors.New("Receive Old Block, this block maybe insert to blockchain already or invalid because of fork")
	}
	if block.Header.Height <= self.latestValidHeight {
		if self.latestValidHeight - block.Header.Height >= 1 {
			self.conflictedPool = append(self.conflictedPool, block)
		}
		return errors.New("receive old block")
	}
	//If block already in pool, it will ignore
	_, ok := self.pendingPool[block.Header.Height]
	if ok {
		return errors.New("receive duplicate block")
	}
	// if next valid block then check max valid pool
	if self.latestValidHeight+1 == block.Header.Height {
		if len(self.validPool) >= self.config.MaxValidBlock && len(self.pendingPool) >= self.config.MaxInvalidBlock {
			return errors.New("Exceed max valid pool and pending pool")
		}
	}
	// if not next valid block then check max pending pool
	if block.Header.Height > self.latestValidHeight {
		if len(self.pendingPool) >= self.config.MaxInvalidBlock {
			return errors.New("Exceed max invalid pending pool")
		}
	}
	return nil
}
func (self *ShardPool) insertNewShardBlockToPool(block *blockchain.ShardBlock) bool{
	if self.latestValidHeight+1 == block.Header.Height {
		if len(self.validPool) < self.config.MaxValidBlock {
			self.validPool = append(self.validPool, block)
			self.updateLatestShardState()
			return true
		} else if len(self.pendingPool) < self.config.MaxInvalidBlock {
			self.pendingPool[block.Header.Height] = block
			return false
		}
	} else {
		self.pendingPool[block.Header.Height] = block
		return false
	}
	return false
}
func (self *ShardPool) updateLatestShardState() {
	self.latestValidHeight = self.validPool[len(self.validPool)-1].Header.Height
}

func (self *ShardPool) promotePendingPool(){
	for {
		nextHeight := self.validPool[len(self.validPool)-1].Header.Height + 1
		block, ok := self.pendingPool[nextHeight]
		if !ok {
			break
		} else {
			err := self.ValidateShardBlock(block)
			if err != nil {
				break
			} else {
				isSuccess := self.insertNewShardBlockToPool(block)
				if !isSuccess {
					break
				}
			}
		}
	}
}
//@Notice: Remove should set latest valid height
//Because normal node may not have these block to remove
func (self *ShardPool) RemoveBlock(lastBlockHeight uint64) {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	self.removeBlock(lastBlockHeight)
}

// remove all block in valid pool that less or equal than input params
func (self *ShardPool) removeBlock(lastBlockHeight uint64) {
	for index, block := range self.validPool {
		if block.Header.Height <= lastBlockHeight {
			// if reach the end of pool then pool will be reset to empty array
			self.cache.Add(block.Header.Hash(), block)
			if index == len(self.validPool)-1 {
				self.validPool = self.validPool[index+1:]
			}
			continue
		} else {
			self.validPool = self.validPool[index:]
			break
		}
	}
}

func (self *ShardPool) GetValidBlock() []*blockchain.ShardBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	if len(self.validPool) > 1 {
		return self.validPool[:len(self.validPool)-1]
	} else {
		return []*blockchain.ShardBlock{}
	}
}

func (self *ShardPool) GetValidBlockHash() []common.Hash {
	blockHashes := []common.Hash{}
	blocks := self.GetValidBlock()
	for _, block := range blocks {
		blockHashes = append(blockHashes, *block.Hash())
	}
	return blockHashes
}

func (self *ShardPool) GetValidBlockHeight() []uint64 {
	blockHashes := []uint64{}
	blocks := self.GetValidBlock()
	for _, block := range blocks {
		blockHashes = append(blockHashes, block.Header.Height)
	}
	return blockHashes
}

func (self *ShardPool) GetLatestValidBlockHeight() uint64 {
	blocks := self.GetValidBlock()
	return blocks[len(blocks)-1].Header.Height
}

func (self *ShardPool) GetAllBlockHeight() []uint64 {
	blockHeights := []uint64{}
	blocks := self.GetValidBlock()
	for _, block := range blocks {
		blockHeights = append(blockHeights, block.Header.Height)
	}
	return blockHeights
}

func (self *ShardPool) GetBlockByHeight(height uint64) *blockchain.ShardBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	for _, block := range self.validPool {
		if block.Header.Height == height {
			return block
		}
	}
	for _, block := range self.pendingPool {
		if block.Header.Height == height {
			return block
		}
	}
	return nil
}
