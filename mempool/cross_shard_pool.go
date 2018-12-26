package mempool

import (
	"errors"
	"log"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
)

var shardPoolLock sync.RWMutex

var shardPool = map[byte][]blockchain.ShardBlock{}

func AddBlock(newBlock blockchain.ShardBlock) error {

	blockHeader := newBlock.Header
	ShardID := blockHeader.ShardID
	// Height := blockHeader.Height

	if ShardID <= 0 {
		return errors.New("Invalid Shard ID")
	}
	// if Height == 0 {
	// 	return errors.New("Invalid Block Heght")
	// }

	shardPoolLock.Lock()
	// TODO validate block pool item
	shardItems, ok := shardPool[ShardID]
	if shardItems == nil || !ok {
		shardItems = []blockchain.ShardBlock{}
	}

	// TODO validate input block
	shardItems = append(shardItems, newBlock)
	shardPool[ShardID] = shardItems
	shardPoolLock.Unlock()

	// TODO validate pool

	return nil
}

// func GetBlock() {

// }

func RemoveBlock(blockhash common.Hash) error {
	// TODO remove block from pool
	return nil
}

func GetAllBlocks() map[byte][]common.Hash {
	results := map[byte][]common.Hash{}

	for shardId, shardItems := range shardPool {
		if len(shardItems) <= 0 {
			continue
		}
		resultItems := []common.Hash{}
		for _, item := range shardItems {
			log.Printf("shard block %+v\n", item)
			value := common.Hash{}
			resultItems = append(resultItems, value)
		}
		results[shardId] = resultItems
	}

	return results
}

func ValidateInputBlock() {
	// TODO validate input block
}

func ValidateOutputBlock() {
	// TODO validate output block
}
