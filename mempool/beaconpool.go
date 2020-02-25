package mempool

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
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
	mtx                   *sync.RWMutex
	config                BeaconPoolConfig
	cache                 *lru.Cache
	RoleInCommittees      bool //Current Role of Node
	RoleInCommitteesEvent pubsub.EventChannel
	PubSubManager         *pubsub.PubSubManager
}

var beaconPool *BeaconPool = nil

func init() {
	go func() {
		mainLoopTime := time.Duration(beaconPoolMainLoopTime)
		ticker := time.NewTicker(mainLoopTime)
		defer ticker.Stop()

		for _ = range ticker.C {
			GetBeaconPool().RemoveBlock(blockchain.GetBeaconBestState().BeaconHeight)
			GetBeaconPool().cleanOldBlock(blockchain.GetBeaconBestState().BeaconHeight)
			GetBeaconPool().PromotePendingPool()
		}
	}()
}

func InitBeaconPool(pubsubManager *pubsub.PubSubManager) {
	//do nothing
	beaconPool := GetBeaconPool()
	beaconPool.SetBeaconState(blockchain.GetBeaconBestState().BeaconHeight)
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
			MaxValidBlock:   maxValidBeaconBlockInPool,
			MaxPendingBlock: maxPendingBeaconBlockInPool,
			CacheSize:       beaconCacheSize,
		}
		beaconPool.cache, _ = lru.New(beaconPool.config.CacheSize)
		beaconPool.mtx = new(sync.RWMutex)
	}
	return beaconPool
}

func (beaconPool *BeaconPool) Start(cQuit chan struct{}) {
	for {
		select {
		case msg := <-beaconPool.RoleInCommitteesEvent:
			role, ok := msg.Value.(bool)
			if !ok {
				continue
			}
			beaconPool.mtx.Lock()
			beaconPool.RoleInCommittees = role
			beaconPool.mtx.Unlock()
		case <-cQuit:
			beaconPool.mtx.Lock()
			beaconPool.RoleInCommittees = false
			beaconPool.mtx.Unlock()
			return
		}
	}
}

func (beaconPool *BeaconPool) SetBeaconState(lastestBeaconHeight uint64) {
	if beaconPool.latestValidHeight < lastestBeaconHeight {
		beaconPool.latestValidHeight = lastestBeaconHeight
	}
}
func (beaconPool *BeaconPool) RevertBeconPool(latestValidHeight uint64) {
	beaconPool.mtx.Lock()
	defer beaconPool.mtx.Unlock()
	Logger.log.Infof("Begin Revert BeaconPool with latest valid height %+v", latestValidHeight)
	beaconBlocks := []*blockchain.BeaconBlock{}
	for _, shardBlock := range beaconPool.validPool {
		beaconBlocks = append(beaconBlocks, shardBlock)
	}
	beaconPool.validPool = []*blockchain.BeaconBlock{}
	for _, shardBlock := range beaconBlocks {
		err := beaconPool.addBeaconBlock(shardBlock)
		if err == nil {
			continue
		} else {
			return
		}
	}
}
func (beaconPool BeaconPool) GetBeaconState() uint64 {
	return beaconPool.latestValidHeight
}

func (beaconPool *BeaconPool) addBeaconBlock(block *blockchain.BeaconBlock) error {
	go beaconPool.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewBeaconBlockTopic, block))
	err := beaconPool.validateBeaconBlock(block, false)
	if err != nil {
		Logger.log.Infof("addBeaconBlock err: %+v", err)
		return err
	}
	beaconPool.insertNewBeaconBlockToPool(block)
	beaconPool.promotePendingPool()
	return nil
}

func (beaconPool *BeaconPool) AddBeaconBlock(block *blockchain.BeaconBlock) error {
	beaconPool.mtx.Lock()
	defer beaconPool.mtx.Unlock()
	return beaconPool.addBeaconBlock(block)
}

func (beaconPool *BeaconPool) validateBeaconBlock(block *blockchain.BeaconBlock, isPending bool) error {
	//If receive old block, it will ignore
	if _, ok := beaconPool.cache.Get(block.Header.Hash()); ok {
		if nextBlock, ok := beaconPool.pendingPool[block.Header.Height+1]; ok {
			beaconPool.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.RequestBeaconBlockByHashTopic, nextBlock.Header.PreviousBlockHash))
		}
		return NewBlockPoolError(OldBlockError, errors.New("Receive Old Block, this block maybe insert to blockchain already or invalid because of fork: "+fmt.Sprintf("%d", block.Header.Height)))
	}
	if block.Header.Height <= beaconPool.latestValidHeight {
		if beaconPool.latestValidHeight-block.Header.Height < 2 {
			beaconPool.conflictedPool[block.Header.Hash()] = block
		}
		return NewBlockPoolError(OldBlockError, errors.New("Receive old block: "+fmt.Sprintf("%d", block.Header.Height)))
	}
	// if next valid block then check max valid pool
	if beaconPool.latestValidHeight+1 == block.Header.Height {
		if len(beaconPool.validPool) >= beaconPool.config.MaxValidBlock && len(beaconPool.pendingPool) >= beaconPool.config.MaxPendingBlock {
			return NewBlockPoolError(MaxPoolSizeError, errors.New("Exceed max valid pool and pending pool"))
		}
	}
	if !isPending {
		//If block already in pool, it will ignore
		_, ok := beaconPool.pendingPool[block.Header.Height]
		if ok {
			return NewBlockPoolError(DuplicateBlockError, errors.New("Receive duplicate block in pending pool: "+fmt.Sprintf("%d", block.Header.Height)))
		}
		// if not next valid block then check max pending pool
		if block.Header.Height > beaconPool.latestValidHeight {
			if len(beaconPool.pendingPool) >= beaconPool.config.MaxPendingBlock {
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
func (beaconPool *BeaconPool) insertNewBeaconBlockToPool(block *blockchain.BeaconBlock) bool {
	Logger.log.Infof("insertNewBeaconBlockToPool blk.Height latestValid: %+v %+v", block.Header.Height, beaconPool.latestValidHeight+1)
	// Condition 1: check height
	if block.Header.Height == beaconPool.latestValidHeight+1 {
		// Condition 2: check pool capacity
		if len(beaconPool.validPool) < beaconPool.config.MaxValidBlock {
			nextHeight := block.Header.Height + 1
			// Condition 3: check next block
			Logger.log.Infof("insertNewBeaconBlockToPool nextHeight: %+v", nextHeight)
			if nextBlock, ok := beaconPool.pendingPool[nextHeight]; ok {
				preHash := &nextBlock.Header.PreviousBlockHash
				blockHeader := block.Header.Hash()
				// Condition 4: next block should point to this block
				if preHash.IsEqual(&blockHeader) {
					Logger.log.Infof("Condition 4: next block should point to this block")
					beaconPool.validPool = append(beaconPool.validPool, block)
					beaconPool.updateLatestBeaconState()
					return true
				} else {
					fmt.Println("BPool: block is fork at height %v with hash %v (block hash should be %v)", block.Header.Height, blockHeader, preHash)
					delete(beaconPool.pendingPool, block.Header.Height)
					beaconPool.cache.Add(block.Header.Hash(), block) // mark as wrong block for validating later
					beaconPool.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.RequestBeaconBlockByHashTopic, preHash))
				}
			} else {
				Logger.log.Infof("no next block found then push to pending pool")
				// no next block found then push to pending pool
				beaconPool.pendingPool[block.Header.Height] = block
			}
		} else if len(beaconPool.pendingPool) < beaconPool.config.MaxPendingBlock {
			beaconPool.pendingPool[block.Header.Height] = block
			return false
		}
	} else {
		beaconPool.pendingPool[block.Header.Height] = block
		return false
	}
	return false
}

func (beaconPool *BeaconPool) updateLatestBeaconState() {
	if len(beaconPool.validPool) > 0 {
		beaconPool.latestValidHeight = beaconPool.validPool[len(beaconPool.validPool)-1].Header.Height
	} else {
		beaconPool.latestValidHeight = blockchain.GetBeaconBestState().BeaconHeight
	}
}

// Check block in pending block then add to valid pool if block is valid
func (beaconPool *BeaconPool) promotePendingPool() {
	for {
		// get next height
		nextHeight := beaconPool.latestValidHeight + 1
		// retrieve next height block in pending
		if block, ok := beaconPool.pendingPool[nextHeight]; ok {
			// validate next block
			err := beaconPool.validateBeaconBlock(block, true)
			if err != nil {
				break
			}
			// insert next block into valid pool
			isSuccess := beaconPool.insertNewBeaconBlockToPool(block)
			if isSuccess {
				delete(beaconPool.pendingPool, nextHeight)
			} else {
				break
			}
		} else {
			break
		}
	}
}

func (beaconPool *BeaconPool) PromotePendingPool() {
	beaconPool.mtx.Lock()
	defer beaconPool.mtx.Unlock()
	beaconPool.promotePendingPool()
}

func (beaconPool *BeaconPool) RemoveBlock(lastBlockHeight uint64) {
	beaconPool.mtx.Lock()
	defer beaconPool.mtx.Unlock()
	beaconPool.removeBlock(lastBlockHeight)
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (beaconPool *BeaconPool) removeBlock(latestBlockHeight uint64) {
	for index, block := range beaconPool.validPool {
		if block.Header.Height <= latestBlockHeight {
			if index == len(beaconPool.validPool)-1 {
				beaconPool.validPool = []*blockchain.BeaconBlock{}
			}
			continue
		} else {
			beaconPool.validPool = beaconPool.validPool[index:]
			break
		}
	}
	beaconPool.updateLatestBeaconState()
}

func (beaconPool *BeaconPool) cleanOldBlock(latestBlockHeight uint64) {
	beaconPool.mtx.Lock()
	defer beaconPool.mtx.Unlock()
	toBeRemovedHeight := []uint64{}
	toBeRemovedHash := []common.Hash{}
	for height := range beaconPool.pendingPool {
		if height <= latestBlockHeight {
			toBeRemovedHeight = append(toBeRemovedHeight, height)
		}
	}
	for hash, block := range beaconPool.conflictedPool {
		if block.Header.Height < latestBlockHeight-2 {
			toBeRemovedHash = append(toBeRemovedHash, hash)
		}
	}
	for _, height := range toBeRemovedHeight {
		delete(beaconPool.pendingPool, height)
	}
	for _, hash := range toBeRemovedHash {
		delete(beaconPool.conflictedPool, hash)
	}
}

func (beaconPool *BeaconPool) GetValidBlock() []*blockchain.BeaconBlock {
	beaconPool.mtx.RLock()
	defer beaconPool.mtx.RUnlock()
	if beaconPool.RoleInCommittees {
		if len(beaconPool.validPool) == 0 {
			if block, ok := beaconPool.pendingPool[beaconPool.latestValidHeight+1]; ok {
				return []*blockchain.BeaconBlock{block}
			}
		}
	}
	return beaconPool.validPool
}

func (beaconPool *BeaconPool) GetValidBlockHeight() []uint64 {
	finalBlocks := []uint64{}
	blks := beaconPool.GetValidBlock()
	for _, blk := range blks {
		finalBlocks = append(finalBlocks, blk.Header.Height)
	}
	return finalBlocks
}

func (beaconPool *BeaconPool) GetPoolLen() uint64 {
	beaconPool.mtx.RLock()
	defer beaconPool.mtx.RUnlock()
	return uint64(len(beaconPool.validPool) + len(beaconPool.pendingPool))
}

func (beaconPool *BeaconPool) GetAllBlockHeight() []uint64 {
	beaconPool.mtx.RLock()
	defer beaconPool.mtx.RUnlock()
	blockHeights := []uint64{}
	for _, block := range beaconPool.validPool {
		blockHeights = append(blockHeights, block.Header.Height)
	}
	for _, block := range beaconPool.pendingPool {
		blockHeights = append(blockHeights, block.Header.Height)
	}

	return blockHeights
}

func (beaconPool *BeaconPool) GetPendingBlockHeight() []uint64 {
	beaconPool.mtx.RLock()
	defer beaconPool.mtx.RUnlock()
	blocks := []uint64{}
	for _, block := range beaconPool.pendingPool {
		blocks = append(blocks, block.Header.Height)
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i] < blocks[j]
	})
	return blocks
}
