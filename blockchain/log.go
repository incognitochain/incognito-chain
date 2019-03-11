package blockchain

import "github.com/big0t/constant-chain/common"

type BlockChainLogger struct {
	log common.Logger
}

func (self *BlockChainLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = BlockChainLogger{}
