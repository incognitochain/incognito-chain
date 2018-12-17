package mempool

import (
	"github.com/ninjadotorg/constant/blockchain"
	"sync"
)

type BlockPool struct {
	mtx  sync.RWMutex
	pool []blockchain.BlockV2
}

func (self *BlockPool) addBlock() {
}

func (self *BlockPool) getBlock() {
}

func (self *BlockPool) removeBlock() {
}

func (self *BlockPool) getAllBlocks() {
}

func (self *BlockPool) validateBlock() {

}
