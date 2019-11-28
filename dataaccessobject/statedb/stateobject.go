package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
)

type StateObject interface {
	GetValue() interface{}
	GetValueBytes() []byte
	GetKey() common.Hash
	GetType() int
	SetValue(interface{})
	GetTrie(DatabaseAccessWarper) Trie
	SetError(error)
	Delete() error
	Exist() bool
	Reset() bool
}
