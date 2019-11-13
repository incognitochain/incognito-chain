package mempool

import (
	"errors"
	"fmt"
	"reflect"
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
	GetShardToBeaconPool().SetShardState(blockchain.GetBeaconBestState().GetBestShardHeight())
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
func (shardToBeaconPool *ShardToBeaconPool) RevertShardToBeaconPool(shardID byte, latestValidHeight uint64) {
	shardToBeaconPool.mtx.Lock()
	defer shardToBeaconPool.mtx.Unlock()
	shardToBeaconPool.latestValidHeightMutex.Lock()
	defer shardToBeaconPool.latestValidHeightMutex.Unlock()
	Logger.log.Infof("Begin Revert ShardToBeaconPool of Shard %+v with latest valid height %+v", shardID, latestValidHeight)
	shardToBeaconBlocks := []*blockchain.ShardToBeaconBlock{}
	if _, ok := shardToBeaconPool.pool[shardID]; ok {
		for _, shardToBeaconBlock := range shardToBeaconPool.pool[shardID] {
			shardToBeaconBlocks = append(shardToBeaconBlocks, shardToBeaconBlock)
		}
		shardToBeaconPool.pool[shardID] = []*blockchain.ShardToBeaconBlock{}
		for _, shardToBeaconBlock := range shardToBeaconBlocks {
			_, _, err := shardToBeaconPool.addShardToBeaconBlock(shardToBeaconBlock)
			if err == nil {
				continue
			} else {
				return
			}
		}
	} else {
		return
	}
}
func (shardToBeaconPool *ShardToBeaconPool) SetShardState(latestShardState map[byte]uint64) {
	// Logger.log.Info("SetShardState")
	shardToBeaconPool.mtx.Lock()
	defer shardToBeaconPool.mtx.Unlock()

	shardToBeaconPool.latestValidHeightMutex.Lock()
	defer shardToBeaconPool.latestValidHeightMutex.Unlock()

	for shardID, latestHeight := range latestShardState {
		if latestHeight < 1 {
			latestShardState[shardID] = 1
		}
		shardToBeaconPool.latestValidHeight[shardID] = latestShardState[shardID]
	}
	//Remove pool base on new shardstate
	shardToBeaconPool.removeBlock(latestShardState)
	shardToBeaconPool.updateLatestShardState()
}

func (shardToBeaconPool *ShardToBeaconPool) GetShardState() map[byte]uint64 {
	return shardToBeaconPool.latestValidHeight
}
func (shardToBeaconPool *ShardToBeaconPool) checkLatestValidHeightValidity(shardID byte) {
	if shardToBeaconPool.latestValidHeight[shardID] == 0 {
		shardToBeaconPool.latestValidHeight[shardID] = 1
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

func (shardToBeaconPool *ShardToBeaconPool) addShardToBeaconBlock(block *blockchain.ShardToBeaconBlock) (uint64, uint64, error) {
	shardID := block.Header.ShardID
	blockHeight := block.Header.Height
	Logger.log.Infof("Add ShardToBeaconBlock from shard %+v, height %+v \n", shardID, blockHeight)

	shardToBeaconPool.checkLatestValidHeightValidity(shardID)
	//If receive old block, it will ignore
	if blockHeight <= shardToBeaconPool.latestValidHeight[shardID] {
		// if old block has round > current block in pool then swap
		if _, ok := shardToBeaconPool.pool[shardID]; ok {
			for index, existedBlock := range shardToBeaconPool.pool[shardID] {
				if existedBlock.Header.Height == blockHeight {
					if existedBlock.Header.Round < block.Header.Round {
						// replace current existed block in pool
						shardToBeaconPool.pool[shardID][index] = block
						return 0, 0, nil
					}
				}
			}
		}
		return 0, 0, NewBlockPoolError(OldBlockError, errors.New("Receive block "+strconv.Itoa(int(blockHeight))+" but expect greater than "+strconv.Itoa(int(shardToBeaconPool.latestValidHeight[shardID]))))
	}
	//If block already in pool, it will ignore
	for i, blkItem := range shardToBeaconPool.pool[shardID] {
		if blkItem.Header.Height == blockHeight {
			if i+1 < len(shardToBeaconPool.pool[shardID]) {
				if !reflect.DeepEqual(*blkItem.Hash(), shardToBeaconPool.pool[shardID][i+1].Header.PreviousBlockHash) {
					shardToBeaconPool.pool[shardID][i] = block
					return 0, 0, nil
				}
			}
			if blkItem.Header.Round < block.Header.Round {
				shardToBeaconPool.pool[shardID][i] = block
				return 0, 0, nil
			}
			return 0, 0, NewBlockPoolError(DuplicateBlockError, errors.New("Receive Duplicate block "+strconv.Itoa(int(blockHeight))))
		}
	}

	//Check if satisfy pool capacity (for valid and invalid)
	if len(shardToBeaconPool.pool[shardID]) != 0 {
		numValidPedingBlk := int(shardToBeaconPool.latestValidHeight[shardID] - shardToBeaconPool.pool[shardID][0].Header.Height + 1)
		if numValidPedingBlk < 0 {
			numValidPedingBlk = 0
		}
		numInValidPedingBlk := len(shardToBeaconPool.pool[shardID]) - numValidPedingBlk + 1
		if numValidPedingBlk >= maxValidShardToBeaconBlockInPool {
			return 0, 0, NewBlockPoolError(MaxPoolSizeError, errors.New("exceed max valid block"))
		}
		lastBlockInPool := shardToBeaconPool.pool[shardID][len(shardToBeaconPool.pool[shardID])-1]
		if numInValidPedingBlk >= maxInvalidShardToBeaconBlockInPool {
			//If invalid block is better than current invalid block
			if lastBlockInPool.Header.Height > blockHeight {
				//remove latest block and add better invalid to pool
				shardToBeaconPool.pool[shardID] = shardToBeaconPool.pool[shardID][:len(shardToBeaconPool.pool[shardID])-1]
			} else {
				return 0, 0, NewBlockPoolError(MaxPoolSizeError, errors.New("exceed invalid pending block"))
			}
		}
	}
	shardToBeaconPool.pool[shardID] = append(shardToBeaconPool.pool[shardID], block)
	//sort pool
	sort.Slice(shardToBeaconPool.pool[shardID], func(i, j int) bool {
		return shardToBeaconPool.pool[shardID][i].Header.Height < shardToBeaconPool.pool[shardID][j].Header.Height
	})
	//update last valid pending ShardState
	shardToBeaconPool.updateLatestShardState()
	//@NOTICE: check logic again
	if shardToBeaconPool.pool[shardID][0].Header.Height > shardToBeaconPool.latestValidHeight[shardID] {
		offset := shardToBeaconPool.pool[shardID][0].Header.Height - shardToBeaconPool.latestValidHeight[shardID]
		if offset > maxValidShardToBeaconBlockInPool {
			offset = maxValidShardToBeaconBlockInPool
		}
		return shardToBeaconPool.latestValidHeight[shardID] + 1, shardToBeaconPool.latestValidHeight[shardID] + offset, nil
	}
	return 0, 0, nil
}

func (shardToBeaconPool *ShardToBeaconPool) AddShardToBeaconBlock(block *blockchain.ShardToBeaconBlock) (uint64, uint64, error) {
	shardID := block.Header.ShardID
	blockHeight := block.Header.Height
	Logger.log.Infof("Add ShardToBeaconBlock from shard %+v, height %+v, hash %v \n", shardID, blockHeight, block.Hash().String())
	shardToBeaconPool.mtx.Lock()
	defer shardToBeaconPool.mtx.Unlock()
	shardToBeaconPool.latestValidHeightMutex.Lock()
	defer shardToBeaconPool.latestValidHeightMutex.Unlock()
	return shardToBeaconPool.addShardToBeaconBlock(block)
}

func (shardToBeaconPool *ShardToBeaconPool) updateLatestShardState() {
	for shardID, blks := range shardToBeaconPool.pool {
		shardToBeaconPool.checkLatestValidHeightValidity(shardID)
		lastHeight := shardToBeaconPool.latestValidHeight[shardID]
		for i, blk := range blks {
			// if block height is not next expected
			if blk.Header.Height < lastHeight+1 {
				continue
			}
			if blk.Header.Height > lastHeight+1 {
				break
			}

			if blk.Header.Height != 2 {
				if i == (len(blks) - 1) {
					break
				} else {
					if blks[i+1].Header.Height != blk.Header.Height+1 {
						break
					}
					if !reflect.DeepEqual(blks[i+1].Header.PreviousBlockHash, *blk.Hash()) {
						fmt.Println("Not equal", blk.Header.ShardID, blk.Header.Height, blks[i+1].Header.Height, (*blk.Hash()).String(), blks[i+1].Header.PreviousBlockHash.String(), lastHeight)
						shardToBeaconPool.pool[shardID] = append(blks[:i], blks[i+1:]...)
						break
					}
				}
			}
			lastHeight = blk.Header.Height
		}
		if shardToBeaconPool.latestValidHeight[shardID] != lastHeight {
			Logger.log.Infof("ShardToBeaconPool: Updated/LastValidHeight %+v of Shard %+v \n", lastHeight, shardID)
			shardToBeaconPool.latestValidHeight[shardID] = lastHeight
		}

	}
}

//@Notice: Remove should set latest valid height
//Because normal beacon node may not have these block to remove
func (shardToBeaconPool *ShardToBeaconPool) RemoveBlock(blockItems map[byte]uint64) {
	shardToBeaconPool.mtx.Lock()
	defer shardToBeaconPool.mtx.Unlock()
	shardToBeaconPool.removeBlock(blockItems)
}

func (shardToBeaconPool *ShardToBeaconPool) removeBlock(blockItems map[byte]uint64) {
	for shardID, blockHeight := range blockItems {
		for index, block := range shardToBeaconPool.pool[shardID] {
			Logger.log.Debugf("ShardToBeaconPool/Pool BEFORE Remove", block.Header.Height)
			if block.Header.Height <= blockHeight {
				if index == len(shardToBeaconPool.pool[shardID])-1 {
					shardToBeaconPool.pool[shardID] = shardToBeaconPool.pool[shardID][index+1:]
				}
				continue
			} else {
				shardToBeaconPool.pool[shardID] = shardToBeaconPool.pool[shardID][index:]
				break
			}
		}
		if blockHeight != 1 {
			Logger.log.Debugf("ShardToBeaconPool: Removed/LastValidHeight %+v of shard %+v \n", blockHeight, shardID)
		}
	}
}

func (shardToBeaconPool *ShardToBeaconPool) GetValidBlock(limit map[byte]uint64) map[byte][]*blockchain.ShardToBeaconBlock {
	shardToBeaconPool.mtx.RLock()
	defer shardToBeaconPool.mtx.RUnlock()
	shardToBeaconPool.latestValidHeightMutex.Lock()
	defer shardToBeaconPool.latestValidHeightMutex.Unlock()
	finalBlocks := make(map[byte][]*blockchain.ShardToBeaconBlock)
	Logger.log.Infof("In GetValidBlock pool: %+v", shardToBeaconPool.pool)
	for shardID, blks := range shardToBeaconPool.pool {
		shardToBeaconPool.checkLatestValidHeightValidity(shardID)
		for i, blk := range blks {
			Logger.log.Infof("In GetValidBlock blks[i]Height && latestValidHeight: %+v %+v", blks[i].Header.Height, shardToBeaconPool.latestValidHeight[shardID])
			if blks[i].Header.Height > shardToBeaconPool.latestValidHeight[shardID] {
				break
			}
			// ?
			if i >= blockchain.GetValidBlock {
				break
			}
			if limit != nil && limit[shardID] != 0 && limit[shardID] < blks[i].Header.Height {
				break
			}
			finalBlocks[shardID] = append(finalBlocks[shardID], blk)
		}
	}
	////UNCOMMENT FOR TESTING
	//Logger.log.Infof("ShardToBeaconPool, Valid Block ")
	//for _, block := range finalBlocks[byte(0)] {
	//	fmt.Printf(" %+v ", block.Header.Height)
	//}
	////==============
	//fmt.Println("GetValidBlock", limit, finalBlocks)
	return finalBlocks
}

func (shardToBeaconPool *ShardToBeaconPool) GetValidBlockHash() map[byte][]common.Hash {
	finalBlocks := make(map[byte][]common.Hash)
	blks := shardToBeaconPool.GetValidBlock(nil)
	for shardID, blkItems := range blks {
		for _, blk := range blkItems {
			finalBlocks[shardID] = append(finalBlocks[shardID], *blk.Hash())
		}
	}
	return finalBlocks
}

func (shardToBeaconPool *ShardToBeaconPool) GetValidBlockHeight() map[byte][]uint64 {
	finalBlocks := make(map[byte][]uint64)
	blks := shardToBeaconPool.GetValidBlock(nil)
	for shardID, blkItems := range blks {
		for _, blk := range blkItems {
			finalBlocks[shardID] = append(finalBlocks[shardID], blk.Header.Height)
		}
	}
	return finalBlocks
}

func (shardToBeaconPool *ShardToBeaconPool) GetLatestValidPendingBlockHeight() map[byte]uint64 {
	finalBlocks := make(map[byte]uint64)
	shardToBeaconPool.latestValidHeightMutex.Lock()
	for shardID, height := range shardToBeaconPool.latestValidHeight {
		finalBlocks[shardID] = height
	}
	shardToBeaconPool.latestValidHeightMutex.Unlock()
	return finalBlocks
}

func (shardToBeaconPool *ShardToBeaconPool) GetAllBlockHeight() map[byte][]uint64 {
	shardToBeaconPool.mtx.RLock()
	defer shardToBeaconPool.mtx.RUnlock()
	finalBlocks := make(map[byte][]uint64)
	for shardID, blks := range shardToBeaconPool.pool {
		for _, blk := range blks {
			finalBlocks[shardID] = append(finalBlocks[shardID], blk.Header.Height)
		}
	}
	return finalBlocks
}

func (shardToBeaconPool *ShardToBeaconPool) GetBlockByHeight(shardID byte, height uint64) *blockchain.ShardToBeaconBlock {
	shardToBeaconPool.mtx.RLock()
	defer shardToBeaconPool.mtx.RUnlock()
	for _shardID, blks := range shardToBeaconPool.pool {
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
