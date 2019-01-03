package mempool

import (
	"errors"
	"sync"

	"github.com/ninjadotorg/constant/blockchain"
)

var nodeBeaconPoolLock sync.RWMutex
var nodeShardPoolLock sync.RWMutex

var nodeShardPool = map[uint64][]blockchain.ShardBlock{}
var nodeBeaconPool = map[uint64][]blockchain.BeaconBlock{}

type NodeShardPool struct{}

func (pool *NodeShardPool) PushBlock(block blockchain.ShardBlock) error {

	blockHeader := block.Header
	Height := blockHeader.Height
	if Height == 0 {
		return errors.New("Invalid Block Heght")
	}

	nodeShardPoolLock.Lock()
	nodeShardPool[Height] = append(nodeShardPool[Height], block)
	nodeShardPoolLock.UnLock()

	return nil
}

func (pool *NodeShardPool) GetBlocks(blockHeight uint64) []blockchain.ShardBlock {
	return nodeShardPool[blockHeight]
}

func (pool *NodeShardPool) RemoveBlocks(blockHeight uint64) error {
	delete(nodeShardPool, blockHeight)
	return nil
}

type NodeBeaconPool struct{}

func (pool *NodeBeaconPool) PushBlock(block blockchain.BeaconBlock) error {

	blockHeader := block.Header
	Height := blockHeader.Height
	if Height == 0 {
		return errors.New("Invalid Block Heght")
	}

	nodeBeaconPoolLock.Lock()
	NodeBeaconPool[Height] = append(NodeBeaconPool[Height], block)
	nodeBeaconPoolLock.UnLock()

	return nil
}

func (pool *NodeBeaconPool) GetBlocks(blockHeight uint64) []blockchain.BeaconBlock {
	return NodeBeaconPool[blockHeight]
}

func (pool *NodeBeaconPool) RemoveBlocks(blockHeight uint64) error {
	delete(NodeBeaconPool, blockHeight)
	return nil
}
