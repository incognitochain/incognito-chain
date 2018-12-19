package mempool

import (
	"errors"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
)

// type BlockPool struct {
// 	mtx  sync.RWMutex
// 	pool []blockchain.ShardToBeaconBlock
// }

var pLock sync.RWMutex
var beaconPool = map[byte]map[uint64][]blockchain.ShardToBeaconBlock{}

func AddBlock(newBlock blockchain.ShardToBeaconBlock) error {

	blockHeader := newBlock.Header
	ShardID := blockHeader.ShardID
	Height := blockHeader.Height
	if ShardID <= 0 {
		return errors.New("Invalid Shard ID")
	}
	if Height == 0 {
		return errors.New("Invalid Block Heght")
	}

	pLock.Lock()
	// TODO validate block pool item
	beaconPoolShardItem, ok := beaconPool[ShardID]
	if beaconPoolShardItem == nil || !ok {
		beaconPoolShardItem = map[uint64][]blockchain.ShardToBeaconBlock{}
	}

	items, ok := beaconPoolShardItem[Height]
	if len(items) <= 0 || !ok {
		items = []blockchain.ShardToBeaconBlock{}
	}
	items = items.append(newBlock)
	beaconPoolShardItem[Height] = items

	beaconPool[ShardID] = beaconPoolShardItem)
	pLock.Unlock()

	return nil
}

func GetBlock(ShardId byte, BlockHeight uint64) (blockchain.ShardToBeaconBlock, error){
	var result blockchain.ShardToBeaconBlock
	if ShardId < 0 || BlockHeight < 0 {
		return nil, errors.New("Invalid Shard ID or Block Heght")
	}
	shardItems, ok := beaconPool[ShardId]
	if shardItems == nil || !ok {
		return nil, errors.New("Shard not exist")
	}
	blocks, ok := shardItems[Height]
	if blocks == nil || len(blocks) <= 0 || !ok {
		return nil, errors.New("Block not exist")
	}

	result = blocks[0]
	return result, nil
}

func RemoveBlock() error {
	// TODO check condition for remove block
	return nil
}

func GetAllBlocks() ([]blockchain.ShardShardToBeaconBlock, error){
	results := []blockchain.ShardShardToBeaconBlock

	return results, nil
}

func ValidateBlock() bool {
	// TODO validate block
	return true
}
