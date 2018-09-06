package database

import (
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

// DB provides the interface that is used to store blocks.
type DB interface {
	StoreBlock(interface{}, byte) error
	FetchBlock(*common.Hash) ([]byte, error)
	HasBlock(*common.Hash) (bool, error)
	FetchAllBlocks() ([][]*common.Hash, error)
	FetchChainBlocks(byte) ([]*common.Hash, error)

	StoreBestBlock(interface{}, byte) error
	FetchBestState(byte) ([]byte, error)

	StoreTx([]byte) error

	StoreBlockIndex(*common.Hash, int32, byte) error
	GetIndexOfBlock(*common.Hash) (int32, byte, error)
	GetBlockByIndex(int32, byte) (*common.Hash, error)

	StoreUtxoEntry(*transaction.OutPoint, interface{}) error
	FetchUtxoEntry(*transaction.OutPoint) ([]byte, error)
	DeleteUtxoEntry(*transaction.OutPoint) error

	Close() error
}
