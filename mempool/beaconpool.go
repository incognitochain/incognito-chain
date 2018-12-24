package mempool

import (
	"errors"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
)

var beaconPoolLock sync.RWMutex
var beaconPool = map[byte]map[uint64][]blockchain.ShardToBeaconBlock{}

// TEMPORARY FOR SHARDTOBEACONBLOCK STRUCT
type ShardToBeaconBlock interface{}

// type ShardToBeaconPool interface {
// 	RemoveBlock([]common.Hash) error
// 	GetBlock() map[byte][]ShardToBeaconBlock
// }

type ShardToBeaconPool struct{}

func (pool *ShardToBeaconPool) GetBlock() map[byte][]ShardToBeaconBlock {
	results := map[byte][]ShardToBeaconBlock{}
	return results
}

func (pool *ShardToBeaconPool) RemoveBlock(blockhashes []common.Hash) error {

	return nil
}

func (pool *ShardToBeaconPool) AddBeaconBlock(newBlock ShardToBeaconBlock) error {
	ShardID := byte("0")
	Height := uint64(1)

	if ShardID <= 0 {
		return errors.New("Invalid Shard ID")
	}
	if Height == 0 {
		return errors.New("Invalid Block Heght")
	}

	beaconPoolLock.Lock()
	// TODO validate block pool item
	beaconPoolShardItem, ok := beaconPool[ShardID]
	if beaconPoolShardItem == nil || !ok {
		beaconPoolShardItem = map[uint64][]ShardToBeaconBlock{}
	}

	items, ok := beaconPoolShardItem[Height]
	if len(items) <= 0 || !ok {
		items = []ShardToBeaconBlock{}
	}
	items = append(items, newBlock)
	beaconPoolShardItem[Height] = items

	beaconPool[ShardID] = beaconPoolShardItem
	beaconPoolLock.Unlock()

	// 	// TODO validate pool
	return nil
}

// func AddBeaconBlock(newBlock blockchain.ShardToBeaconBlock) error {

// 	blockHeader := newBlock.Header
// 	ShardID := blockHeader.ShardID
// 	Height := blockHeader.Height
// 	if ShardID <= 0 {
// 		return errors.New("Invalid Shard ID")
// 	}
// 	if Height == 0 {
// 		return errors.New("Invalid Block Heght")
// 	}

// 	beaconPoolLock.Lock()
// 	// TODO validate block pool item
// 	beaconPoolShardItem, ok := beaconPool[ShardID]
// 	if beaconPoolShardItem == nil || !ok {
// 		beaconPoolShardItem = map[uint64][]blockchain.ShardToBeaconBlock{}
// 	}

// 	items, ok := beaconPoolShardItem[Height]
// 	if len(items) <= 0 || !ok {
// 		items = []blockchain.ShardToBeaconBlock{}
// 	}
// 	items = append(items, newBlock)
// 	beaconPoolShardItem[Height] = items

// 	beaconPool[ShardID] = beaconPoolShardItem
// 	beaconPoolLock.Unlock()

// 	// TODO validate pool

// 	return nil
// }

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

// func GetAllBeaconBlocks() ([]blockchain.ShardToBeaconBlock, error) {
// 	results := []blockchain.ShardToBeaconBlock{}
// 	for _, shards := range beaconPool {
// 		if shards == nil {
// 			continue
// 		}
// 		for _, items := range shards {
// 			results = append(results, items...)
// 		}
// 	}
// 	return results, nil
// }

// func ReviseBeaconPool(blockchain.ShardToBeaconBlock) error {
// 	// TODO validate all block with same height
// 	return nil
// }
