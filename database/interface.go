package database

import (
	"github.com/ninjadotorg/cash-prototype/common"
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

	StoreNullifiers([]byte, string) error
	FetchNullifiers(string) ([][]byte, error)
	HasNullifier([]byte, string) (bool, error)
	StoreCommitments([]byte, string) error
	FetchCommitments(string) ([][]byte, error)
	HasCommitment([]byte, string) (bool, error)

	StoreBlockIndex(*common.Hash, int32, byte) error
	GetIndexOfBlock(*common.Hash) (int32, byte, error)
	GetBlockByIndex(int32, byte) (*common.Hash, error)

	//StoreUtxoEntry(*transaction.OutPoint, interface{}) error
	//FetchUtxoEntry(*transaction.OutPoint) ([]byte, error)
	//DeleteUtxoEntry(*transaction.OutPoint) error

	Close() error
}
