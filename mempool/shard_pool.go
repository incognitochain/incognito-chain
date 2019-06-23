package mempool

import (
	"errors"
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"sort"
	"sync"
	"time"
)

const (
	MAX_VALID_SHARD_BLK_IN_POOL   = 10000
	MAX_PENDING_SHARD_BLK_IN_POOL = 10000
	SHARD_CACHE_SIZE              = 2000
	SHARD_POOL_MAIN_LOOP_TIME     = 500 // count in milisecond
)

type ShardPoolConfig struct {
	MaxValidBlock   int
	MaxPendingBlock int
	CacheSize       int
}
type ShardPool struct {
	validPool             []*blockchain.ShardBlock          // valid, ready to insert into blockchain
	pendingPool           map[uint64]*blockchain.ShardBlock // not ready to insert into blockchain, there maybe many blocks exists at one height
	conflictedPool        map[common.Hash]*blockchain.ShardBlock
	shardID               byte
	latestValidHeight     uint64
	mtx                   *sync.RWMutex
	config                ShardPoolConfig
	cache                 *lru.Cache
	RoleInCommittees      int //Current Role of Node
	RoleInCommitteesEvent pubsub.Event
	PubsubManager         *pubsub.PubsubManager
}

var shardPoolMap = make(map[byte]*ShardPool)
var defaultConfig = ShardPoolConfig{
	MaxValidBlock:   MAX_VALID_SHARD_BLK_IN_POOL,
	MaxPendingBlock: MAX_PENDING_SHARD_BLK_IN_POOL,
	CacheSize:       SHARD_CACHE_SIZE,
}

//@NOTICE: Shard pool will always be empty when node start
func init() {
	go func() {
		mainLoopTime := time.Duration(SHARD_POOL_MAIN_LOOP_TIME) * time.Millisecond
		ticker := time.Tick(mainLoopTime)
		for _ = range ticker {
			for k, _ := range shardPoolMap {
				GetShardPool(k).RemoveBlock(blockchain.GetBestStateShard(k).ShardHeight)
				//GetShardPool(k).CleanOldBlock(blockchain.GetBestStateShard(k).ShardHeight)
				GetShardPool(k).PromotePendingPool()
			}
		}
	}()
}

func InitShardPool(pool map[byte]blockchain.ShardPool, pubsubManager *pubsub.PubsubManager) {
	for i := 0; i < 255; i++ {
		shardPoolMap[byte(i)] = GetShardPool(byte(i))
		//update last shard height
		shardPoolMap[byte(i)].mtx = new(sync.RWMutex)
		shardPoolMap[byte(i)].SetShardState(blockchain.GetBestStateShard(byte(i)).ShardHeight)
		pool[byte(i)] = shardPoolMap[byte(i)]
		shardPoolMap[byte(i)].PubsubManager = pubsubManager
		_, subChanRole, _ := shardPoolMap[byte(i)].PubsubManager.RegisterNewSubscriber(pubsub.ShardRoleTopic)
		shardPoolMap[byte(i)].RoleInCommitteesEvent = subChanRole
	}
}
func (self *ShardPool) Start(cQuit chan struct{}) {
	for {
		select {
		case msg := <-self.RoleInCommitteesEvent:
			role, ok := msg.Value.(int)
			if !ok {
				continue
			}
			self.mtx.Lock()
			self.RoleInCommittees = role
			self.mtx.Unlock()
		case <-cQuit:
			self.mtx.Lock()
			self.RoleInCommittees = -1
			self.mtx.Unlock()
			return
		}
	}
}

// get singleton instance of ShardToBeacon pool
func GetShardPool(shardID byte) *ShardPool {
	if shardPoolMap[shardID] == nil {
		shardPool := new(ShardPool)
		shardPool.shardID = shardID
		shardPool.latestValidHeight = 1
		shardPoolMap[shardID] = shardPool
		shardPool.validPool = []*blockchain.ShardBlock{}
		shardPool.conflictedPool = make(map[common.Hash]*blockchain.ShardBlock)
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
	err = self.ValidateShardBlock(block, false)
	if err != nil {
		return err
	}
	self.insertNewShardBlockToPoolV2(block)
	self.promotePendingPool()
	//self.CleanOldBlock(blockchain.GetBestStateShard(self.shardID).ShardHeight)
	return nil
}

func (self *ShardPool) ValidateShardBlock(block *blockchain.ShardBlock, isPending bool) error {
	//If receive old block, it will ignore
	if _, ok := self.cache.Get(block.Header.Hash()); ok {
		return NewBlockPoolError(OldBlockError, errors.New("Receive Old Block, this block maybe insert to blockchain already or invalid because of fork"))
	}
	if block.Header.Height <= self.latestValidHeight {
		if self.latestValidHeight-block.Header.Height > 2 {
			self.conflictedPool[block.Header.Hash()] = block
		}
		return NewBlockPoolError(OldBlockError, errors.New("Receive old block"))
	}
	if !isPending {
		//If block already in pool, it will ignore
		_, ok := self.pendingPool[block.Header.Height]
		if ok {
			return NewBlockPoolError(DuplicateBlockError, errors.New("Receive duplicate block in pending pool"))
		}
		// if not next valid block then check max pending pool
		if block.Header.Height > self.latestValidHeight {
			if len(self.pendingPool) >= self.config.MaxPendingBlock {
				return NewBlockPoolError(MaxPoolSizeError, errors.New("Exceed max invalid pending pool"))
			}
		}
	}
	// if next valid block then check max valid pool
	if self.latestValidHeight+1 == block.Header.Height {
		if len(self.validPool) >= self.config.MaxValidBlock && len(self.pendingPool) >= self.config.MaxPendingBlock {
			return NewBlockPoolError(MaxPoolSizeError, errors.New("Exceed max valid pool and pending pool"))
		}
	}
	return nil
}

func (self *ShardPool) updateLatestShardState() {
	if len(self.validPool) > 0 {
		self.latestValidHeight = self.validPool[len(self.validPool)-1].Header.Height
	} else {
		self.latestValidHeight = blockchain.GetBestStateShard(self.shardID).ShardHeight
	}
}

func (self *ShardPool) promotePendingPool() {
	for {
		nextHeight := self.latestValidHeight + 1
		block, ok := self.pendingPool[nextHeight]
		if !ok {
			break
		} else {
			err := self.ValidateShardBlock(block, true)
			if err != nil {
				break
			} else {
				isSuccess := self.insertNewShardBlockToPoolV2(block)
				if !isSuccess {
					break
				} else {
					delete(self.pendingPool, nextHeight)
				}
			}
		}
	}
}
func (self *ShardPool) PromotePendingPool() {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	self.promotePendingPool()
}

/*
	Description:
	In case we have two block at one height: Ex: 5A and 5B
	Block 5A or AB is final blocks if there's exist block 6 that point to either 5A or 5B
	=> so new version will wait for the next block to determine which previous block is final

	- New Block Will be inserted into valid pool if match these condition:
		1 New Block Height = LatestValidHeight + 1
		2 Block with Height = New Block Height + 1 exist in pool -> skip
		3 New Block Previous Hash = Latest block hash in valid pool (if valid pool is not empty)
		- If new block pre hash does not point to latest block hash in pool then,
			+ Delete current latest block hash in pool
			+ Add new block to pending pool
			+ Find conflicted block with recent delete latest block in pool if possbile then try to add conflicted block into pool
*/
func (self *ShardPool) insertNewShardBlockToPoolV2(block *blockchain.ShardBlock) bool {
	//If unknown to beacon best state store in pending
	if block.Header.Height > blockchain.GetBestStateBeacon().GetBestHeightOfShard(block.Header.ShardID) {
		self.pendingPool[block.Header.Height] = block
		return false
	}
	// Condition 1
	if self.latestValidHeight+1 == block.Header.Height {
		// if pool still has available room
		if len(self.validPool) < self.config.MaxValidBlock {
			// Condition 2
			//if _, ok := self.pendingPool[block.Header.Height+1]; ok {
			if len(self.validPool) > 0 {
				latestBlock := self.validPool[len(self.validPool)-1]
				latestBlockHash := latestBlock.Header.Hash()
				// condition 3
				preHash := &block.Header.PrevBlockHash
				if preHash.IsEqual(&latestBlockHash) {
					self.validPool = append(self.validPool, block)
					self.updateLatestShardState()
					return true
				} else {
					// add new block to pending pool
					self.pendingPool[block.Header.Height] = block
					// delete latest block in pool
					self.validPool = self.validPool[:len(self.validPool)-1]
					// update latest state
					self.updateLatestShardState()
					// add delete block to cache
					self.cache.Add(latestBlockHash, latestBlock)
					// find previous block of new block
					previousBlock, ok := self.conflictedPool[block.Header.PrevBlockHash]
					if ok {
						// try to add previous block of new block
						err := self.AddShardBlock(previousBlock)
						if err == nil {
							delete(self.conflictedPool, previousBlock.Header.Hash())
							return true
						}
					}
					return false
				}
				// if valid pool is empty then add block to valid pool
			} else {
				self.validPool = append(self.validPool, block)
				self.updateLatestShardState()
				return true
			}
		}
		//} else if len(self.pendingPool) < self.config.MaxPendingBlock{
		//	self.pendingPool[block.Header.Height] = block
		//}
	} else {
		self.pendingPool[block.Header.Height] = block
	}
	return false
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
				self.validPool = []*blockchain.ShardBlock{}
			}
			continue
		} else {
			self.validPool = self.validPool[index:]
			break
		}
	}
	self.updateLatestShardState()
}

func (self *ShardPool) CleanOldBlock(latestBlockHeight uint64) {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	toBeRemovedHeight := []uint64{}
	toBeRemovedHash := []common.Hash{}
	for height, _ := range self.pendingPool {
		if height <= latestBlockHeight {
			toBeRemovedHeight = append(toBeRemovedHeight, height)
		}
	}
	for hash, block := range self.conflictedPool {
		if block.Header.Height < latestBlockHeight-2 {
			toBeRemovedHash = append(toBeRemovedHash, hash)
		}
	}
	for _, height := range toBeRemovedHeight {
		delete(self.pendingPool, height)
	}
	for _, hash := range toBeRemovedHash {
		delete(self.conflictedPool, hash)
	}
}

// @NOTICE: this function only serve insert block from pool
func (self *ShardPool) GetValidBlock() []*blockchain.ShardBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	if self.RoleInCommittees != -1 {
		if len(self.validPool) == 0 {
			if block, ok := self.pendingPool[self.latestValidHeight+1]; ok {
				//delete(self.pendingPool, self.latestValidHeight+1)
				return []*blockchain.ShardBlock{block}
			}
		}
	}
	return self.validPool
}
func (self *ShardPool) GetPendingBlock() []*blockchain.ShardBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	blocks := []*blockchain.ShardBlock{}
	for _, block := range self.pendingPool {
		blocks = append(blocks, block)
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Header.Height < blocks[j].Header.Height
	})
	return blocks
}
func (self *ShardPool) GetPendingBlockHeight() []uint64 {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	blocks := []uint64{}
	for _, block := range self.pendingPool {
		blocks = append(blocks, block.Header.Height)
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i] < blocks[j]
	})
	return blocks
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
	if len(blocks) < 1 {
		return 0
	}
	return blocks[len(blocks)-1].Header.Height
}

func (self *ShardPool) GetAllBlockHeight() []uint64 {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	blockHeights := []uint64{}
	for _, block := range self.validPool {
		blockHeights = append(blockHeights, block.Header.Height)
	}
	for _, block := range self.pendingPool {
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
