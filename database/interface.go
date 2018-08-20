package database

import "github.com/internet-cash/prototype/blockchain"

type DB interface {
	GetChain() []*blockchain.Block
	PutChain(block []*blockchain.Block) (bool, error)
}
