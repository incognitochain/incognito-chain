package databasemp

import (
	"github.com/constant-money/constant-chain/common"
)

type DatabaseInterface interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	HasValue(key []byte) (bool, error)
	
	AddTransaction(key *common.Hash, value []byte) error
	RemoveTransaction(key *common.Hash) error
	GetTransaction(key *common.Hash) ([]byte, error)
	Reset() error
	
	Close() error
}
