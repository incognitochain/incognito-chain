package database

import (
	"github.com/ninjadotorg/constant/common"
)

// DatabaseInterface provides the interface that is used to store blocks.
type DatabaseInterface interface {
	// Block
	StoreBlock(interface{}, byte) error
	FetchBlock(*common.Hash) ([]byte, error)
	HasBlock(*common.Hash) (bool, error)
	FetchAllBlocks() (map[byte][]*common.Hash, error)
	FetchChainBlocks(byte) ([]*common.Hash, error)
	DeleteBlock(*common.Hash, int32, byte) error

	// Block index
	StoreBlockIndex(*common.Hash, int32, byte) error
	GetIndexOfBlock(*common.Hash) (int32, byte, error)
	GetBlockByIndex(int32, byte) (*common.Hash, error)

	// Best state of chain
	StoreBestState(interface{}, byte) error
	FetchBestState(byte) ([]byte, error)
	CleanBestState() error

	// Nullifier
	StoreNullifiers([]byte, string, byte) error
	FetchNullifiers(string, byte) ([][]byte, error)
	HasNullifier([]byte, string, byte) (bool, error)
	CleanNullifiers() error

	// Commitment
	StoreCommitments([]byte, string, byte) error
	FetchCommitments(string, byte) ([][]byte, error)
	HasCommitment([]byte, string, byte) (bool, error)
	CleanCommitments() error

	// Fee estimator
	StoreFeeEstimator([]byte, byte) error
	GetFeeEstimator(byte) ([]byte, error)
	CleanFeeEstimator() error

	Close() error
}
