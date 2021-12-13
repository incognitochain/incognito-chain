// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package chaindb

import (
	"context"
	"github.com/incognitochain/incognito-chain/incdb"
	"io"
)

// Database wraps all database operations. All methods are safe for concurrent use.
type Database interface {
	KeyValueReader
	KeyValueWriter
	io.Closer

	NewBatch() incdb.Batch
	Path() string
	incdb.Iteratee
	Subscribe(ctx context.Context, cb func(kv *KVList) error, prefixes []byte) error
	ClearAll() error
}

// Batch is a write-only operation.
type Batch interface {
	KeyValueWriter

	ValueSize() int
	Reset()
}

// Iterator iterates over key/value pairs in ascending key order.
// Must be released after use.
type Iterator interface {
	Next() bool
	Key() []byte
	Value() []byte
	Release()
}

// Reader interface
type KeyValueReader interface {
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
}

// Writer interface
type KeyValueWriter interface {
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	Write() error
}
