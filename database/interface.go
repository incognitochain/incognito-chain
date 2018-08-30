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
	StoreBestBlockIndex(int32) error
	FetchBestBlockIndex() (int32, error)
	StoreTx([]byte) error
	Close() error
}
