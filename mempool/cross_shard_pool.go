package mempool

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
)

var shardPoolLock sync.RWMutex

var shardPool = make(map[byte][]blockchain.CrossShardBlock)
var poolState = make(map[byte]uint64)

type CrossShardPool struct{}

func (pool *CrossShardPool) GetBlock(bestStateInfos map[byte]uint64) map[byte][]blockchain.CrossShardBlock {
	fmt.Println("Try to GetBlock From CrossShardPool", shardPool)
	results := map[byte][]blockchain.CrossShardBlock{}
	shardPoolLock.Lock()
	defer shardPoolLock.Unlock()
	for shardId, shardItems := range shardPool {
		// shardBestState, ok := bestStateInfos[shardId]
		// if !ok || shardBestState < 0 {
		// 	continue
		// }

		items := []blockchain.CrossShardBlock{}

		for _, item := range shardItems {
			// fmt.Printf("item Height %+v && shardBestState %+v \n", item.Header.Height, shardBestState)
			// if item.Header.Height > shardBestState {
			// 	continue
			// }
			items = append(items, item)
		}
		results[shardId] = items
	}
	// fmt.Println("Return result of GetBlock From CrossShardPool", results)
	return results
}

func (pool *CrossShardPool) RemoveBlock(blockItems map[byte]uint64) error {
	if len(blockItems) <= 0 {
		log.Println("Block items empty")
		return nil
	}

	shardPoolLock.Lock()
	for shardID, blockHeight := range blockItems {
		shardItems, ok := shardPool[shardID]
		if !ok || len(shardItems) <= 0 {
			log.Println("Shard is not exist")
			continue
		}
		index := 0
		for index, block := range shardPool[shardID] {
			if block.Header.Height > blockHeight {
				if index != 0 {
					poolState[shardID] = shardPool[shardID][index-1].Header.Height
				}
				break
			}
		}
		shardPool[shardID] = shardPool[shardID][index:]
	}
	shardPoolLock.Unlock()
	return nil
}

func (pool *CrossShardPool) AddCrossShardBlock(newBlock blockchain.CrossShardBlock) error {

	blockHeader := newBlock.Header
	shardID := blockHeader.ShardID
	height := blockHeader.Height

	if height == 0 {
		return errors.New("Invalid Block Heght")
	}

	shardPoolLock.Lock()

	shardPool[shardID] = append(shardPool[shardID], newBlock)

	fmt.Println("CrossShardPool", shardPool)
	shardPoolLock.Unlock()

	return nil
}

func GetCrossShardPoolState() map[byte]uint64 {
	return poolState
}
