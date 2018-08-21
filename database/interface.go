package database

import (
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
)

type DB interface {

	GetBlock(hash common.Hash) *blockchain.Block
	SaveBlock(block *blockchain.Block) (bool, error)
}
