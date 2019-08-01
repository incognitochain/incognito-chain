package blockchain

import "github.com/incognitochain/incognito-chain/common"

type BlockChainLogger struct {
	log common.Logger
}

func (self *BlockChainLogger) Init(inst common.Logger) {
	self.log = inst
}

type DeBridgeLogger struct {
	log common.Logger
}

func (self *DeBridgeLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = BlockChainLogger{}
var BLogger = DeBridgeLogger{}
