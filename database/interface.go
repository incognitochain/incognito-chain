package database

import (
	"github.com/internet-cash/prototype/blockchain"
	"github.com/internet-cash/prototype/common"
)

type DB interface {

	GetChain(hash common.Hash) *blockchain.Block
	PutChain(block *blockchain.Block) (bool, error)
}
