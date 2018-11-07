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

	// Transaction Index
	StoreTransactionIndex(*common.Hash, *common.Hash, int) error
	GetTransactionIndexById(*common.Hash) (*common.Hash, int, error)
	// Best state of chain
	StoreBestState(interface{}, byte) error
	FetchBestState(byte) ([]byte, error)
	CleanBestState() error

	// Nullifier
	StoreNullifiers([]byte, byte) error
	FetchNullifiers(byte) ([][]byte, error)
	HasNullifier([]byte, byte) (bool, error)
	CleanNullifiers() error

	// Commitment
	StoreCommitments([]byte, byte) error
	FetchCommitments(byte) ([][]byte, error)
	HasCommitment([]byte, byte) (bool, error)
	CleanCommitments() error

	// Fee estimator
	StoreFeeEstimator([]byte, byte) error
	GetFeeEstimator(byte) ([]byte, error)
	CleanFeeEstimator() error

	// Custom token
	StoreCustomToken(*common.Hash, []byte) error                       // param: tokenID, txInitToken-id, data
	StoreCustomTokenTx(*common.Hash, byte, int32, int32, []byte) error // param: tokenID, chainID, block height, tx-id, data
	ListCustomToken() ([][]byte, error)

	Close() error
}
