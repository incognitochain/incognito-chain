package database

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

// DB provides the interface that is used to store blocks.
type DB interface {
	StoreBlock(v interface{}) error
	FetchBlock(*common.Hash) ([]byte, error)
	HasBlock(*common.Hash) (bool, error)
	StoreTx([]byte) error
	Close() error
}
