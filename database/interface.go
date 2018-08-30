package database

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

// DB provides the interface that is used to store blocks.
type DB interface {
	StoreBlock(v interface{}) error
	FetchBlock(*common.Hash) ([]byte, error)
	HasBlock(*common.Hash) (bool, error)

	StoreBestBlock(v interface{}) error
	FetchBestBlock() ([]byte, error)

	StoreTx([]byte) error

	StoreBlockIndex(*common.Hash, int32) error
	GetIndexOfBlock(*common.Hash) (int32, error)
	GetBlockByIndex(int32) ([]byte, error)

	Close() error
}
