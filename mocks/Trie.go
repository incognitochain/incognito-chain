// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import common "github.com/incognitochain/incognito-chain/common"
import incdb "github.com/incognitochain/incognito-chain/incdb"
import mock "github.com/stretchr/testify/mock"

import trie "github.com/incognitochain/incognito-chain/trie"

// Trie is an autogenerated mock type for the Trie type
type Trie struct {
	mock.Mock
}

// Commit provides a mock function with given fields: onleaf
func (_m *Trie) Commit(onleaf trie.LeafCallback) (common.Hash, error) {
	ret := _m.Called(onleaf)

	var r0 common.Hash
	if rf, ok := ret.Get(0).(func(trie.LeafCallback) common.Hash); ok {
		r0 = rf(onleaf)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(trie.LeafCallback) error); ok {
		r1 = rf(onleaf)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetKey provides a mock function with given fields: _a0
func (_m *Trie) GetKey(_a0 []byte) []byte {
	ret := _m.Called(_a0)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]byte) []byte); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// Hash provides a mock function with given fields:
func (_m *Trie) Hash() common.Hash {
	ret := _m.Called()

	var r0 common.Hash
	if rf, ok := ret.Get(0).(func() common.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	return r0
}

// NodeIterator provides a mock function with given fields: startKey
func (_m *Trie) NodeIterator(startKey []byte) trie.NodeIterator {
	ret := _m.Called(startKey)

	var r0 trie.NodeIterator
	if rf, ok := ret.Get(0).(func([]byte) trie.NodeIterator); ok {
		r0 = rf(startKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(trie.NodeIterator)
		}
	}

	return r0
}

// Prove provides a mock function with given fields: key, fromLevel, proofDb
func (_m *Trie) Prove(key []byte, fromLevel uint, proofDb incdb.Database) error {
	ret := _m.Called(key, fromLevel, proofDb)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, uint, incdb.Database) error); ok {
		r0 = rf(key, fromLevel, proofDb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TryDelete provides a mock function with given fields: key
func (_m *Trie) TryDelete(key []byte) error {
	ret := _m.Called(key)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TryGet provides a mock function with given fields: key
func (_m *Trie) TryGet(key []byte) ([]byte, error) {
	ret := _m.Called(key)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]byte) []byte); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TryUpdate provides a mock function with given fields: key, value
func (_m *Trie) TryUpdate(key []byte, value []byte) error {
	ret := _m.Called(key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, []byte) error); ok {
		r0 = rf(key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
