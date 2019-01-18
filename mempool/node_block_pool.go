package mempool

import (
	"errors"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
)

var nodeBeaconPoolLock sync.RWMutex
var nodeShardPoolLock sync.RWMutex

var nodeShardPool = map[byte]map[uint64][]blockchain.ShardBlock{}
var nodeBeaconPool = map[uint64][]blockchain.BeaconBlock{}

type NodeShardPool struct{}

func (pool *NodeShardPool) PushBlock(block blockchain.ShardBlock) error {

	blockHeader := block.Header
	shardID := blockHeader.ShardID
	height := blockHeader.Height
	if height == 0 {
		return errors.New("Invalid Block Heght")
	}

	nodeShardPoolLock.Lock()
	shardItems := nodeShardPool[shardID]
	if shardItems == nil {
		shardItems = map[uint64][]blockchain.ShardBlock{}
	}
	shardItems[height] = append(shardItems[height], block)
	nodeShardPool[shardID] = shardItems
	nodeShardPoolLock.Unlock()

	return nil
}

func (pool *NodeShardPool) GetBlocks(shardID byte, blockHeight uint64) ([]blockchain.ShardBlock, error) {

	if blockHeight == 0 {
		return []blockchain.ShardBlock{}, errors.New("Invalid ShardId or block Height")
	}
	shardItems := nodeShardPool[shardID]

	return shardItems[blockHeight], nil
}

func (pool *NodeShardPool) RemoveBlocks(shardID byte, blockHeight uint64) error {
	if shardID <= 0 || blockHeight == 0 {
		return errors.New("Invalid ShardId or block Height")
	}

	nodeShardPoolLock.Lock()
	shardItems := nodeShardPool[shardID]
	delete(shardItems, blockHeight)
	nodeShardPool[shardID] = shardItems
	nodeShardPoolLock.Unlock()

	return nil
}

type NodeBeaconPool struct{}

func (pool *NodeBeaconPool) PushBlock(block blockchain.BeaconBlock) error {

	blockHeader := block.Header
	height := blockHeader.Height
	if height == 0 {
		return errors.New("Invalid Block Heght")
	}

	nodeBeaconPoolLock.Lock()
	defer nodeBeaconPoolLock.Unlock()
	if _, ok := nodeBeaconPool[height]; ok {
		isNew := true
		for _, poolblk := range nodeBeaconPool[height] {
			if poolblk.Hash() == block.Hash() {
				isNew = false
				return nil
			}
		}
		if isNew {
			nodeBeaconPool[height] = append(nodeBeaconPool[height], block)
		}
	}
	return nil
}

func (pool *NodeBeaconPool) GetBlocks(blockHeight uint64) ([]blockchain.BeaconBlock, error) {
	return nodeBeaconPool[blockHeight], nil
}

func (pool *NodeBeaconPool) RemoveBlocks(blockHeight uint64) error {
	delete(nodeBeaconPool, blockHeight)
	return nil
}
