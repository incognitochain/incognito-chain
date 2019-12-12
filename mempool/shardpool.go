package mempool

import (
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"sort"
	"strconv"
	"sync"
	"time"
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
	RoleInCommitteesEvent pubsub.EventChannel
	PubSubManager         *pubsub.PubSubManager
}

var shardPoolMap = make(map[byte]*ShardPool)
var shardPoolMapMu sync.Mutex

var defaultConfig = ShardPoolConfig{
	MaxValidBlock:   maxValidShardBlockInPool,
	MaxPendingBlock: maxPendingShardBlockInPool,
	CacheSize:       shardCacheSize,
}

//@NOTICE: Shard pool will always be empty when node start
func init() {
	go func() {
		mainLoopTime := time.Duration(shardPoolMainLoopTime)
		ticker := time.NewTicker(mainLoopTime)
		for _ = range ticker.C {
			for k, _ := range shardPoolMap {
				GetShardPool(k).RemoveBlock(blockchain.GetBestStateShard(k).ShardHeight)
				GetShardPool(k).CleanOldBlock(blockchain.GetBestStateShard(k).ShardHeight)
				GetShardPool(k).PromotePendingPool()
			}
		}
	}()
}

func InitShardPool(pool map[byte]blockchain.ShardPool, pubsubManager *pubsub.PubSubManager) {
	shardPoolMapMu.Lock()
	defer shardPoolMapMu.Unlock()
	for i := 0; i < common.MaxShardNumber; i++ {
		shardPoolMap[byte(i)] = getShardPool(byte(i))
		shardPoolMap[byte(i)].mtx = new(sync.RWMutex)
		//update last shard height
		shardPoolMap[byte(i)].SetShardState(blockchain.GetBestStateShard(byte(i)).ShardHeight)
		pool[byte(i)] = shardPoolMap[byte(i)]
		shardPoolMap[byte(i)].PubSubManager = pubsubManager
		_, subChanRole, _ := shardPoolMap[byte(i)].PubSubManager.RegisterNewSubscriber(pubsub.ShardRoleTopic)
		shardPoolMap[byte(i)].RoleInCommitteesEvent = subChanRole
	}
}
func (shardPool *ShardPool) Start(cQuit chan struct{}) {
	for {
		select {
		case msg := <-shardPool.RoleInCommitteesEvent:
			role, ok := msg.Value.(int)
			if !ok {
				continue
			}
			shardPool.mtx.Lock()
			shardPool.RoleInCommittees = role
			shardPool.mtx.Unlock()
		case <-cQuit:
			shardPool.mtx.Lock()
			shardPool.RoleInCommittees = -1
			shardPool.mtx.Unlock()
			return
		}
	}
}

func getShardPool(shardID byte) *ShardPool {
	if shardPoolMap[shardID] == nil {
		shardPool := new(ShardPool)
		shardPool.shardID = shardID
		shardPool.latestValidHeight = 1
		shardPool.RoleInCommittees = -1
		shardPool.validPool = []*blockchain.ShardBlock{}
		shardPool.conflictedPool = make(map[common.Hash]*blockchain.ShardBlock)
		shardPool.config = defaultConfig
		shardPool.pendingPool = make(map[uint64]*blockchain.ShardBlock)
		shardPool.cache, _ = lru.New(shardPool.config.CacheSize)
		shardPool.mtx = new(sync.RWMutex)
		shardPoolMap[shardID] = shardPool
	}
	return shardPoolMap[shardID]
}

// get singleton instance of Shard Pool with lock
func GetShardPool(shardID byte) *ShardPool {
	shardPoolMapMu.Lock()
	defer shardPoolMapMu.Unlock()
	return getShardPool(shardID)
}

func (shardPool *ShardPool) SetShardState(lastestShardHeight uint64) {
	shardPool.mtx.Lock()
	defer shardPool.mtx.Unlock()
	if shardPool.latestValidHeight < lastestShardHeight {
		shardPool.latestValidHeight = lastestShardHeight
	}
}

func (shardPool *ShardPool) GetShardState() uint64 {
	shardPool.mtx.RLock()
	defer shardPool.mtx.RUnlock()
	return shardPool.latestValidHeight
}
func (shardPool *ShardPool) RevertShardPool(latestValidHeight uint64) {
	shardPool.mtx.Lock()
	defer shardPool.mtx.Unlock()
	Logger.log.Infof("Begin Revert ShardPool of Shard %+v with latest valid height %+v", shardPool.shardID, latestValidHeight)
	shardBlocks := []*blockchain.ShardBlock{}
	for _, shardBlock := range shardPool.validPool {
		shardBlocks = append(shardBlocks, shardBlock)
	}
	shardPool.validPool = []*blockchain.ShardBlock{}
	for _, shardBlock := range shardBlocks {
		err := shardPool.addShardBlock(shardBlock)
		if err == nil {
			continue
		} else {
			return
		}
	}
}

func (shardPool *ShardPool) addShardBlock(block *blockchain.ShardBlock) error {
	var err error
	err = shardPool.validateShardBlock(block, false)
	if err != nil {
		return err
	}
	shardPool.insertNewShardBlockToPool(block)
	shardPool.promotePendingPool()
	return nil
}
func (shardPool *ShardPool) AddShardBlock(block *blockchain.ShardBlock) error {
	shardPool.mtx.Lock()
	defer shardPool.mtx.Unlock()
	return shardPool.addShardBlock(block)
	return nil
}

func (shardPool *ShardPool) validateShardBlock(block *blockchain.ShardBlock, isPending bool) error {
	//If receive old block, it will ignore
	if _, ok := shardPool.cache.Get(block.Header.Hash()); ok {
		return NewBlockPoolError(OldBlockError, errors.New("Receive Old Block, this block maybe insert to blockchain already or invalid because of fork"))
	}
	if block.Header.Height <= shardPool.latestValidHeight {
		if shardPool.latestValidHeight-block.Header.Height < 2 {
			shardPool.conflictedPool[block.Header.Hash()] = block
		}
		return NewBlockPoolError(OldBlockError, errors.New("Receive old block"))
	}
	// if next valid block then check max valid pool
	if shardPool.latestValidHeight+1 == block.Header.Height {
		if isPending {
			if len(shardPool.validPool) >= shardPool.config.MaxValidBlock && len(shardPool.pendingPool) >= shardPool.config.MaxPendingBlock+1 {
				return NewBlockPoolError(MaxPoolSizeError, errors.New("Exceed max valid pool and pending pool"))
			}
		} else {
			if len(shardPool.validPool) >= shardPool.config.MaxValidBlock && len(shardPool.pendingPool) >= shardPool.config.MaxPendingBlock {
				return NewBlockPoolError(MaxPoolSizeError, errors.New("Exceed max valid pool and pending pool"))
			}
		}
	}
	if !isPending {
		//If block already in pool, it will ignore
		_, ok := shardPool.pendingPool[block.Header.Height]
		if ok {
			return NewBlockPoolError(DuplicateBlockError, errors.New("Receive duplicate block in pending pool"))
		}
		// if not next valid block then check max pending pool
		if block.Header.Height > shardPool.latestValidHeight {
			if len(shardPool.pendingPool) >= shardPool.config.MaxPendingBlock {
				return NewBlockPoolError(MaxPoolSizeError, errors.New("Exceed max invalid pending pool"))
			}
		}
	}
	return nil
}

func (shardPool *ShardPool) updateLatestShardState() {
	if len(shardPool.validPool) > 0 {
		shardPool.latestValidHeight = shardPool.validPool[len(shardPool.validPool)-1].Header.Height
	} else {
		shardPool.latestValidHeight = blockchain.GetBestStateShard(shardPool.shardID).ShardHeight
	}
}

func (shardPool *ShardPool) promotePendingPool() {
	for {
		nextHeight := shardPool.latestValidHeight + 1
		block, ok := shardPool.pendingPool[nextHeight]
		if !ok {
			break
		} else {
			err := shardPool.validateShardBlock(block, true)
			if err != nil {
				break
			} else {
				isSuccess := shardPool.insertNewShardBlockToPool(block)
				if !isSuccess {
					break
				} else {
					delete(shardPool.pendingPool, nextHeight)
				}
			}
		}
	}
}
func (shardPool *ShardPool) PromotePendingPool() {
	shardPool.mtx.Lock()
	defer shardPool.mtx.Unlock()
	shardPool.promotePendingPool()
	sort.Slice(shardPool.validPool, func(i, j int) bool {
		return shardPool.validPool[i].Header.Height < shardPool.validPool[j].Header.Height
	})
}

/*
	Description:
	In case we have two block at one height: Ex: 5A and 5B
	Block 5A or AB is final blocks if there's exist block 6 that point to either 5A or 5B
	=> so new version will wait for the next block to determine which previous block is final

	- New Block Will be inserted into valid pool if match these condition:
		1 Beacon Best State is greater than block height
		2 New Block Height = LatestValidHeight + 1
		3 Valid block still have room for next block
		4 Block with Height = New Block Height + 1 exist in pool -> skip
		5 New Block Previous Hash = Latest block hash in valid pool (if valid pool is not empty)
		- If new block pre hash does not point to latest block hash in pool then,
			+ Delete current latest block hash in pool
			+ Add new block to pending pool
			+ Find conflicted block with recent delete latest block in pool if possbile then try to add conflicted block into pool
*/
func (shardPool *ShardPool) insertNewShardBlockToPool(block *blockchain.ShardBlock) bool {
	//If unknown to beacon best state store in pending
	// Condition 1
	if block.Header.Height > blockchain.GetBeaconBestState().GetBestHeightOfShard(block.Header.ShardID) {
		shardPool.pendingPool[block.Header.Height] = block
		return false
	}
	// Condition 2
	if shardPool.latestValidHeight+1 == block.Header.Height {
		// if pool still has available room
		// condition 3
		if len(shardPool.validPool) < shardPool.config.MaxValidBlock {
			// Condition 4
			if len(shardPool.validPool) > 0 {
				latestBlock := shardPool.validPool[len(shardPool.validPool)-1]
				latestBlockHash := latestBlock.Header.Hash()
				// condition 5
				preHash := &block.Header.PreviousBlockHash
				if preHash.IsEqual(&latestBlockHash) {
					shardPool.validPool = append(shardPool.validPool, block)
					shardPool.updateLatestShardState()
					return true
				} else {
					// add new block to pending pool
					if len(shardPool.pendingPool) < shardPool.config.MaxPendingBlock {
						shardPool.pendingPool[block.Header.Height] = block
					}
					// delete latest block in pool
					shardPool.validPool = shardPool.validPool[:len(shardPool.validPool)-1]
					// update latest state
					shardPool.updateLatestShardState()
					// add delete block to cache
					shardPool.cache.Add(latestBlockHash, latestBlock)
					// find previous block of new block
					previousBlock, ok := shardPool.conflictedPool[block.Header.PreviousBlockHash]
					if ok {
						// try to add previous block of new block
						err := shardPool.AddShardBlock(previousBlock)
						if err == nil {
							delete(shardPool.conflictedPool, previousBlock.Header.Hash())
							return true
						}
					} else {
						msg := strconv.Itoa(int(block.Header.ShardID))
						msg += fmt.Sprintf("%+v", block.Header.PreviousBlockHash)
						shardPool.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.RequestShardBlockByHashTopic, msg))
					}
					return false
				}
				// if valid pool is empty then add block to valid pool
			} else {
				shardPool.validPool = append(shardPool.validPool, block)
				shardPool.updateLatestShardState()
				return true
			}
		} else if len(shardPool.pendingPool) < shardPool.config.MaxPendingBlock {
			shardPool.pendingPool[block.Header.Height] = block
			return false
		}
	} else if len(shardPool.pendingPool) < shardPool.config.MaxPendingBlock {
		shardPool.pendingPool[block.Header.Height] = block
		return false
	}
	return false
}

//@Notice: Remove should set latest valid height
//Because normal node may not have these block to remove
func (shardPool *ShardPool) RemoveBlock(lastBlockHeight uint64) {
	shardPool.mtx.Lock()
	defer shardPool.mtx.Unlock()
	shardPool.removeBlock(lastBlockHeight)
}

// remove all block in valid pool that less or equal than input params
func (shardPool *ShardPool) removeBlock(lastBlockHeight uint64) {
	for index, block := range shardPool.validPool {
		if block.Header.Height <= lastBlockHeight {
			// if reach the end of pool then pool will be reset to empty array
			shardPool.cache.Add(block.Header.Hash(), block)
			if index == len(shardPool.validPool)-1 {
				shardPool.validPool = []*blockchain.ShardBlock{}
			}
			continue
		} else {
			shardPool.validPool = shardPool.validPool[index:]
			break
		}
	}
	//if len(shardPool.validPool) > 0 {
	//	fmt.Println("Remove block routine", shardPool.shardID, shardPool.validPool[0].Header.Height)
	//}
	shardPool.updateLatestShardState()
}

func (shardPool *ShardPool) CleanOldBlock(latestBlockHeight uint64) {
	shardPool.mtx.Lock()
	defer shardPool.mtx.Unlock()
	toBeRemovedHeight := []uint64{}
	toBeRemovedHash := []common.Hash{}
	for height := range shardPool.pendingPool {
		if height <= latestBlockHeight {
			toBeRemovedHeight = append(toBeRemovedHeight, height)
		}
	}
	for hash, block := range shardPool.conflictedPool {
		if block.Header.Height < latestBlockHeight-2 {
			toBeRemovedHash = append(toBeRemovedHash, hash)
		}
	}
	for _, height := range toBeRemovedHeight {
		delete(shardPool.pendingPool, height)
	}
	for _, hash := range toBeRemovedHash {
		delete(shardPool.conflictedPool, hash)
	}
}

// @NOTICE: this function only serve insert block from pool
func (shardPool *ShardPool) GetValidBlock() []*blockchain.ShardBlock {
	shardPool.mtx.RLock()
	defer shardPool.mtx.RUnlock()
	if shardPool.RoleInCommittees != -1 {
		if len(shardPool.validPool) == 0 {
			if block, ok := shardPool.pendingPool[shardPool.latestValidHeight+1]; ok {
				//delete(shardPool.pendingPool, shardPool.latestValidHeight+1)
				return []*blockchain.ShardBlock{block}
			}
		}
	}
	return shardPool.validPool
}
func (shardPool *ShardPool) GetPendingBlock() []*blockchain.ShardBlock {
	shardPool.mtx.RLock()
	defer shardPool.mtx.RUnlock()
	blocks := []*blockchain.ShardBlock{}
	for _, block := range shardPool.pendingPool {
		blocks = append(blocks, block)
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Header.Height < blocks[j].Header.Height
	})
	return blocks
}
func (shardPool *ShardPool) GetPendingBlockHeight() []uint64 {
	shardPool.mtx.RLock()
	defer shardPool.mtx.RUnlock()
	blocks := []uint64{}
	for _, block := range shardPool.pendingPool {
		blocks = append(blocks, block.Header.Height)
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i] < blocks[j]
	})
	return blocks
}

func (shardPool *ShardPool) GetValidBlockHash() []common.Hash {
	blockHashes := []common.Hash{}
	blocks := shardPool.GetValidBlock()
	for _, block := range blocks {
		blockHashes = append(blockHashes, *block.Hash())
	}
	return blockHashes
}

func (shardPool *ShardPool) GetValidBlockHeight() []uint64 {
	blockHashes := []uint64{}
	blocks := shardPool.GetValidBlock()
	for _, block := range blocks {
		blockHashes = append(blockHashes, block.Header.Height)
	}
	return blockHashes
}

func (shardPool *ShardPool) GetLatestValidBlockHeight() uint64 {
	return shardPool.latestValidHeight
}

func (shardPool *ShardPool) GetAllBlockHeight() []uint64 {
	shardPool.mtx.RLock()
	defer shardPool.mtx.RUnlock()
	blockHeights := []uint64{}
	for _, block := range shardPool.validPool {
		blockHeights = append(blockHeights, block.Header.Height)
	}
	for _, block := range shardPool.pendingPool {
		blockHeights = append(blockHeights, block.Header.Height)
	}
	return blockHeights
}

func (shardPool *ShardPool) GetBlockByHeight(height uint64) *blockchain.ShardBlock {
	shardPool.mtx.RLock()
	defer shardPool.mtx.RUnlock()
	for _, block := range shardPool.validPool {
		if block.Header.Height == height {
			return block
		}
	}
	for _, block := range shardPool.pendingPool {
		if block.Header.Height == height {
			return block
		}
	}
	return nil
}
