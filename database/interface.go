package database

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

// DB provides the interface that is used to store blocks.
type DB interface {
	StoreBlock(v interface{}) error
	FetchBlock(*common.Hash) ([]byte, error)
	HasBlock(*common.Hash) (bool, error)
	FetchAllBlocks() ([]*common.Hash, error)

	StoreBestBlock(v interface{}) error
	FetchBestState() ([]byte, error)

	StoreNullifiers([]byte) error
	FetchNullifiers() ([][]byte, error)

	StoreBlockIndex(*common.Hash, int32) error
	GetIndexOfBlock(*common.Hash) (int32, error)
	GetBlockByIndex(int32) (*common.Hash, error)

	//StoreUtxoEntry(*transaction.OutPoint, interface{}) error
	//FetchUtxoEntry(*transaction.OutPoint) ([]byte, error)
	//DeleteUtxoEntry(*transaction.OutPoint) error

	Close() error
}
