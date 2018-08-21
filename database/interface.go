package database

import (
	"github.com/ninjadotorg/money-prototype/blockchain"
	"github.com/ninjadotorg/money-prototype/common"
)

type DB interface {

	GetBlock(hash common.Hash) *blockchain.Block
	SaveBlock(block *blockchain.Block) (bool, error)
}
