package mempool

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
)

var beaconPoolLock sync.RWMutex
var beaconPool = map[byte]map[uint64][]blockchain.ShardToBeaconBlock{}

type ShardToBeaconPool struct{}

func (pool *ShardToBeaconPool) GetFinalBlock() map[byte][]blockchain.ShardToBeaconBlock {
	results := map[byte][]blockchain.ShardToBeaconBlock{}

	for ShardId, shardItems := range beaconPool {
		if shardItems == nil || len(shardItems) <= 0 {
			continue
		}
		items := []blockchain.ShardToBeaconBlock{}
		for _, blks := range shardItems {
			items = append(items, blks[0])
		}
		results[ShardId] = items
	}
	return results

}

func (pool *ShardToBeaconPool) RemoveBlock(blockItems map[byte]uint64) error {
	if len(blockItems) <= 0 {
		log.Println("ShardToBeaconPool: Block items empty")
		return nil
	}

	beaconPoolLock.Lock()
	for shardID, blockHeight := range blockItems {
		shardItems, ok := beaconPool[shardID]
		if !ok || len(shardItems) <= 0 {
			log.Println("Shard is not exist")
			continue
		}

		items := map[uint64][]blockchain.ShardToBeaconBlock{}
		for i := blockHeight + 1; i < uint64(len(shardItems)); i++ {
			items[i] = shardItems[i]
		}

		beaconPool[shardID] = items
	}
	beaconPoolLock.Unlock()
	return nil
}

func (pool *ShardToBeaconPool) AddShardBeaconBlock(newBlock blockchain.ShardToBeaconBlock) error {
	fmt.Println()
	fmt.Println("BLAH BLAH aaaaaaaaa", beaconPool)
	fmt.Println()
	blockHeader := newBlock.Header
	ShardID := blockHeader.ShardID
	Height := blockHeader.Height
	PrevBlockHash := blockHeader.PrevBlockHash

	if Height == 0 {
		return errors.New("Invalid Block Heght")
	}

	beaconPoolLock.Lock()
	// TODO validate block pool item
	beaconPoolShardItem, ok := beaconPool[ShardID]
	if beaconPoolShardItem == nil || !ok {
		beaconPoolShardItem = map[uint64][]blockchain.ShardToBeaconBlock{}
	}

	items, ok := beaconPoolShardItem[Height]
	if len(items) <= 0 || !ok {
		items = []blockchain.ShardToBeaconBlock{}
	}
	items = append(items, newBlock)
	beaconPoolShardItem[Height] = items

	beaconPool[ShardID] = beaconPoolShardItem

	err := UpdateBeaconPool(ShardID, Height, PrevBlockHash)
	if err != nil {
		log.Println("update beacon pool err: ", err)
	}
	log.Println("update previous block items with same height")

	beaconPoolLock.Unlock()

	return nil
}

func UpdateBeaconPool(shardID byte, blockHeight uint64, preBlockHash common.Hash) error {
	if blockHeight == 0 {
		return errors.New("Invalid Block Heght")
	}
	if len(preBlockHash) <= 0 {
		return errors.New("Invalid Previous Block Hash")
	}
	shardItems, ok := beaconPool[shardID]
	if !ok || len(shardItems) <= 0 {
		log.Println("pool shard items not exists")
		return nil
	}
	prevBlockHeight := blockHeight - 1
	if prevBlockHeight < 0 {
		return nil
	}
	blocks, ok := shardItems[prevBlockHeight]
	if !ok || len(blocks) <= 0 {
		return nil
	}

	for _, block := range blocks {
		header := block.Header
		hash := header.Hash()
		if hash == preBlockHash {
			shardItems[prevBlockHeight] = []blockchain.ShardToBeaconBlock{block}
			beaconPool[shardID] = shardItems
			break
		}
	}

	return nil
}

func (pool *ShardToBeaconPool) GetDistinctBlockMap() map[byte]map[uint64][]common.Hash {
	var poolBlksMap map[byte]map[uint64][]common.Hash
	poolBlksMap = make(map[byte]map[uint64][]common.Hash)
	beaconPoolLock.Lock()
	defer beaconPoolLock.Unlock()
	for ShardId, shardItems := range beaconPool {
		if shardItems == nil || len(shardItems) <= 0 {
			continue
		}
		items := map[uint64][]common.Hash{}
		items = make(map[uint64][]common.Hash)
		for height, blks := range shardItems {
			for _, blk := range blks {
				items[height] = append(items[height], *blk.Hash())
			}

		}
		poolBlksMap[ShardId] = items
	}
	return poolBlksMap
}

// func GetBeaconBlock(ShardId byte, BlockHeight uint64) (blockchain.ShardToBeaconBlock, error) {
// 	result := blockchain.ShardToBeaconBlock{}
// 	if ShardId < 0 || BlockHeight < 0 {
// 		return blockchain.ShardToBeaconBlock{}, errors.New("Invalid Shard ID or Block Heght")
// 	}
// 	shardItems, ok := beaconPool[ShardId]
// 	if shardItems == nil || !ok {
// 		return blockchain.ShardToBeaconBlock{}, errors.New("Shard not exist")
// 	}
// 	blocks, ok := shardItems[BlockHeight]
// 	if blocks == nil || len(blocks) <= 0 || !ok {
// 		return blockchain.ShardToBeaconBlock{}, errors.New("Block not exist")
// 	}

// 	result = blocks[0]
// 	return result, nil
// }
