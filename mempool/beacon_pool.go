package mempool

import (
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"sort"
	"sync"
	"time"
)

type BeaconPoolConfig struct {
	MaxValidBlock   int
	MaxPendingBlock int
	CacheSize       int
}
type BeaconPool struct {
	validPool             []*blockchain.BeaconBlock          // valid, ready to insert into blockchain
	pendingPool           map[uint64]*blockchain.BeaconBlock // not ready to insert into blockchain, there maybe many blocks exists at one height
	conflictedPool        map[common.Hash]*blockchain.BeaconBlock
	latestValidHeight     uint64
	mtx                   sync.RWMutex
	config                BeaconPoolConfig
	cache                 *lru.Cache
	RoleInCommittees      bool //Current Role of Node
	RoleInCommitteesEvent pubsub.EventChannel
	PubSubManager         *pubsub.PubSubManager
}

var beaconPool *BeaconPool = nil

func init() {
	go func() {
		mainLoopTime := time.Duration(BEACON_POOL_MAIN_LOOP_TIME) * time.Millisecond
		ticker := time.Tick(mainLoopTime)
		for _ = range ticker {
			GetBeaconPool().RemoveBlock(blockchain.GetBestStateBeacon().BeaconHeight)
			GetBeaconPool().CleanOldBlock(blockchain.GetBestStateBeacon().BeaconHeight)
			GetBeaconPool().PromotePendingPool()
		}
	}()
}

func InitBeaconPool(pubsubManager *pubsub.PubSubManager) {
	//do nothing
	beaconPool := GetBeaconPool()
	beaconPool.SetBeaconState(blockchain.GetBestStateBeacon().BeaconHeight)
	beaconPool.PubSubManager = pubsubManager
	_, subChanRole, _ := beaconPool.PubSubManager.RegisterNewSubscriber(pubsub.BeaconRoleTopic)
	beaconPool.RoleInCommitteesEvent = subChanRole
}

// get singleton instance of ShardToBeacon pool
func GetBeaconPool() *BeaconPool {
	if beaconPool == nil {
		beaconPool = new(BeaconPool)
		beaconPool.latestValidHeight = 1
		beaconPool.validPool = []*blockchain.BeaconBlock{}
		beaconPool.pendingPool = make(map[uint64]*blockchain.BeaconBlock)
		beaconPool.conflictedPool = make(map[common.Hash]*blockchain.BeaconBlock)
		beaconPool.config = BeaconPoolConfig{
			MaxValidBlock:   MAX_VALID_BEACON_BLK_IN_POOL,
			MaxPendingBlock: MAX_PENDING_BEACON_BLK_IN_POOL,
			CacheSize:       BEACON_CACHE_SIZE,
		}
		beaconPool.cache, _ = lru.New(beaconPool.config.CacheSize)
	}
	return beaconPool
}
func (self *BeaconPool) Start(cQuit chan struct{}) {
	for {
		select {
		case msg := <-self.RoleInCommitteesEvent:
			role, ok := msg.Value.(bool)
			if !ok {
				continue
			}
			self.mtx.Lock()
			self.RoleInCommittees = role
			self.mtx.Unlock()
		case <-cQuit:
			self.mtx.Lock()
			self.RoleInCommittees = false
			self.mtx.Unlock()
			return
		}
	}
}
func (self *BeaconPool) SetBeaconState(lastestBeaconHeight uint64) {
	if self.latestValidHeight < lastestBeaconHeight {
		self.latestValidHeight = lastestBeaconHeight
	}
}

func (self *BeaconPool) GetBeaconState() uint64 {
	return self.latestValidHeight
}

func (self *BeaconPool) AddBeaconBlock(block *blockchain.BeaconBlock) error {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	go self.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewBeaconBlockTopic, block))
	err := self.validateBeaconBlock(block, false)
	if err != nil {
		return err
	}
	self.insertNewBeaconBlockToPool(block)
	self.promotePendingPool()
	//self.CleanOldBlock(blockchain.GetBestStateBeacon().BeaconHeight)
	return nil
}

func (self *BeaconPool) validateBeaconBlock(block *blockchain.BeaconBlock, isPending bool) error {
	//If receive old block, it will ignore
	if _, ok := self.cache.Get(block.Header.Hash()); ok {
		return NewBlockPoolError(OldBlockError, errors.New("Receive Old Block, this block maybe insert to blockchain already or invalid because of fork: "+fmt.Sprintf("%d", block.Header.Height)))
	}
	if block.Header.Height <= self.latestValidHeight {
		if self.latestValidHeight-block.Header.Height < 2 {
			self.conflictedPool[block.Header.Hash()] = block
		}
		return NewBlockPoolError(OldBlockError, errors.New("Receive old block: "+fmt.Sprintf("%d", block.Header.Height)))
	}
	// if next valid block then check max valid pool
	if self.latestValidHeight+1 == block.Header.Height {
		if len(self.validPool) >= self.config.MaxValidBlock && len(self.pendingPool) >= self.config.MaxPendingBlock {
			return NewBlockPoolError(MaxPoolSizeError, errors.New("Exceed max valid pool and pending pool"))
		}
	}
	if !isPending {
		//If block already in pool, it will ignore
		_, ok := self.pendingPool[block.Header.Height]
		if ok {
			return NewBlockPoolError(DuplicateBlockError, errors.New("Receive duplicate block in pending pool: "+fmt.Sprintf("%d", block.Header.Height)))
		}
		// if not next valid block then check max pending pool
		if block.Header.Height > self.latestValidHeight {
			if len(self.pendingPool) >= self.config.MaxPendingBlock {
				return NewBlockPoolError(MaxPoolSizeError, errors.New("Exceed max invalid pending pool"))
			}
		}
	}
	return nil
}

/*
 New block only become valid after
	1. This block height is next block height ( latest valid height + 1)
	2. Valid Pool still has avaiable capacity
	3. Pending pool has next block,
	4. and next block has previous hash == this block hash
*/
func (self *BeaconPool) insertNewBeaconBlockToPool(block *blockchain.BeaconBlock) bool {
	// Condition 1: check height
	if block.Header.Height == self.latestValidHeight+1 {
		// Condition 2: check pool capacity
		if len(self.validPool) < self.config.MaxValidBlock {
			nextHeight := block.Header.Height + 1
			// Condition 3: check next block
			if nextBlock, ok := self.pendingPool[nextHeight]; ok {
				preHash := &nextBlock.Header.PrevBlockHash
				blockHeader := block.Header.Hash()
				// Condition 4: next block should point to this block
				if preHash.IsEqual(&blockHeader) {
					self.validPool = append(self.validPool, block)
					self.updateLatestBeaconState()
					return true
				} else {
					self.cache.Add(block.Header.Hash(), block)
				}
			} else {
				// no next block found then push to pending pool
				self.pendingPool[block.Header.Height] = block
			}
		} else if len(self.pendingPool) < self.config.MaxPendingBlock {
			self.pendingPool[block.Header.Height] = block
			return false
		}
	} else {
		self.pendingPool[block.Header.Height] = block
		return false
	}
	return false
}
func (self *BeaconPool) updateLatestBeaconState() {
	if len(self.validPool) > 0 {
		self.latestValidHeight = self.validPool[len(self.validPool)-1].Header.Height
	} else {
		self.latestValidHeight = blockchain.GetBestStateBeacon().BeaconHeight
	}
}

// Check block in pending block then add to valid pool if block is valid
func (self *BeaconPool) promotePendingPool() {
	for {
		// get next height
		nextHeight := self.latestValidHeight + 1
		// retrieve next height block in pending
		if block, ok := self.pendingPool[nextHeight]; ok {
			// validate next block
			err := self.validateBeaconBlock(block, true)
			if err != nil {
				break
			}
			// insert next block into valid pool
			isSuccess := self.insertNewBeaconBlockToPool(block)
			if isSuccess {
				delete(self.pendingPool, nextHeight)
			} else {
				break
			}
		} else {
			break
		}
	}
}
func (self *BeaconPool) PromotePendingPool() {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	self.promotePendingPool()
}
func (self *BeaconPool) RemoveBlock(lastBlockHeight uint64) {
	self.mtx.Lock()
	defer self.mtx.Unlock()
	self.removeBlock(lastBlockHeight)
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (self *BeaconPool) removeBlock(latestBlockHeight uint64) {
	for index, block := range self.validPool {
		if block.Header.Height <= latestBlockHeight {
			if index == len(self.validPool)-1 {
				self.validPool = []*blockchain.BeaconBlock{}
			}
			continue
		} else {
			self.validPool = self.validPool[index:]
			break
		}
	}
	self.updateLatestBeaconState()
}

func (self *BeaconPool) CleanOldBlock(latestBlockHeight uint64) {
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

func (self *BeaconPool) GetValidBlock() []*blockchain.BeaconBlock {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	if self.RoleInCommittees {
		if len(self.validPool) == 0 {
			if block, ok := self.pendingPool[self.latestValidHeight+1]; ok {
				return []*blockchain.BeaconBlock{block}
			}
		}
	}
	return self.validPool
}

func (self *BeaconPool) GetValidBlockHash() []common.Hash {
	finalBlocks := []common.Hash{}
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = append(finalBlocks, *blk.Hash())
	}
	return finalBlocks
}

func (self *BeaconPool) GetValidBlockHeight() []uint64 {
	finalBlocks := []uint64{}
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = append(finalBlocks, blk.Header.Height)
	}
	return finalBlocks
}

func (self *BeaconPool) GetLatestValidBlockHeight() uint64 {
	finalBlocks := uint64(0)
	blks := self.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = blk.Header.Height
	}
	return finalBlocks
}

func (self *BeaconPool) GetPoolLen() uint64 {
	self.mtx.RLock()
	defer self.mtx.RUnlock()
	return uint64(len(self.validPool) + len(self.pendingPool))
}

func (self *BeaconPool) GetAllBlockHeight() []uint64 {
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
func (self *BeaconPool) GetPendingBlockHeight() []uint64 {
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
func (self *BeaconPool) GetBlockByHeight(height uint64) *blockchain.BeaconBlock {
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
