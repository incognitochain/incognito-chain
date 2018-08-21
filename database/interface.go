package database

import (
	"github.com/internet-cash/prototype/blockchain"
	"github.com/internet-cash/prototype/common"
)

type DB interface {

	GetBlock(hash common.Hash) *blockchain.Block
	SaveBlock(block *blockchain.Block) (bool, error)
}
