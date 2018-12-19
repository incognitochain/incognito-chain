package mempool

import (
	"errors"
	"github.com/ninjadotorg/constant/blockchain"
	"sync"
)

type shardToBeacon blockchain.ShardToBeaconBlock

type BlockPool struct {
	mtx  sync.RWMutex
	pool []shardToBeacon
}

type ShardPoolItem struct {
	state bool
	block shardToBeacon
}

var shardMap = map[byte][]ShardPoolItem{}

func (self *BlockPool) AddBlock(newBlock shardToBeacon) error {
	blockHeader := newBlock.Header
	ShardID := blockHeader.ShardID
	if ShardID <= 0 {
		return errors.New("invalid Shard ID")
	}

	if _, ok := shardMap[ShardID]; ok {
		return nil
	}
	shardPoolItem := ShardPoolItem{
		state: false,
		block: newBlock,
	}
	shardMap[ShardID] = append(shardMap[ShardID], shardPoolItem)

	return nil
}

func (self *BlockPool) GetBlock() {
}

func (self *BlockPool) RemoveBlock() {
}

func (self *BlockPool) GetAllBlocks() {
}

func (self *BlockPool) ValidateBlock() {

}
