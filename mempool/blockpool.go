package mempool

import (
	"errors"
	"sync"

	"github.com/ninjadotorg/constant/blockchain/blocktype"
)

type BlockPool struct {
	mtx  sync.RWMutex
	pool []blocktype.ShardToBeaconBlock
}

type ShardPoolItem struct {
	state bool
	block blocktype.ShardToBeaconBlock
}

var shardMap = map[byte][]ShardPoolItem{}

func (self *BlockPool) AddBlock(newBlock blocktype.ShardToBeaconBlock) error {
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
