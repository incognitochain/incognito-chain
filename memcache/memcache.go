// Package memorydb implements the key-value database layer based on memory maps.
// Reference go-etherium memorydb
package memcache

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

var (
	// errMemorydbClosed is returned if a memory database was already closed at the
	// invocation of a data access operation.
	errMemorydbClosed = errors.New("database closed")

	// errMemorydbNotFound is returned if a key is requested that is not found in
	// the provided memory database.
	errMemorydbNotFound = errors.New("not found")

	errExpired = errors.New("expired time")
)

// MemoryCache is an ephemeral key-value store. Apart from basic data storage
// functionality it also supports batch writes and iterating over the keyspace in
// binary-alphabetical order.
type MemoryCache struct {
	db      map[string][]byte
	expired map[string]time.Time
	lock    sync.RWMutex
}

// New returns a wrapped map with all the required database interface methods
// implemented.
func New() *MemoryCache {
	return &MemoryCache{
		db:      make(map[string][]byte),
		expired: make(map[string]time.Time),
	}
}

// NewWithCap returns a wrapped map pre-allocated to the provided capcity with
// all the required database interface methods implemented.
func NewWithCap(size int) *MemoryCache {
	return &MemoryCache{
		db: make(map[string][]byte, size),
	}
}

// Close deallocates the internal map and ensures any consecutive data access op
// failes with an error.
func (db *MemoryCache) Close() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.db = nil
	return nil
}

// Has retrieves if a key is present in the key-value store.
func (db *MemoryCache) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if db.db == nil {
		return false, errMemorydbClosed
	}
	_, ok := db.db[string(key)]
	return ok, nil
}

// Get retrieves the given key if it's present in the key-value store.
func (db *MemoryCache) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	//defer db.lock.RUnlock()

	if db.db == nil {
		db.lock.RUnlock()
		return nil, errMemorydbClosed
	}
	if entry, ok := db.db[string(key)]; ok {
		// check expired time
		if expired, ok1 := db.expired[string(key)]; ok1 {
			if expired.Before(time.Now()) {
				// is expired
				db.lock.RUnlock()
				db.Delete(key)
				return nil, errExpired
			}
		}
		db.lock.RUnlock()
		return common.CopyBytes(entry), nil
	}
	db.lock.RUnlock()
	return nil, errMemorydbNotFound
}

// Put inserts the given value into the key-value store.
func (db *MemoryCache) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return errMemorydbClosed
	}
	db.db[string(key)] = common.CopyBytes(value)
	return nil
}

// Put inserts the given value into the key-value store. expired is mili second
func (db *MemoryCache) PutExpired(key []byte, value []byte, expired time.Duration) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return errMemorydbClosed
	}
	db.db[string(key)] = common.CopyBytes(value)
	db.expired[string(key)] = time.Now().Add(expired * time.Millisecond)
	return nil
}

// Delete removes the key from the key-value store.
func (db *MemoryCache) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if db.db == nil {
		return errMemorydbClosed
	}
	delete(db.db, string(key))
	return nil
}

// NewIterator creates a binary-alphabetical iterator over the entire keyspace
// contained within the memory database.
func (db *MemoryCache) NewIterator() Iterator {
	return db.NewIteratorWithStart(nil)
}

// NewIteratorWithStart creates a binary-alphabetical iterator over a subset of
// database content starting at a particular initial key (or after, if it does
// not exist).
func (db *MemoryCache) NewIteratorWithStart(start []byte) Iterator {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var (
		st     = string(start)
		keys   = make([]string, 0, len(db.db))
		values = make([][]byte, 0, len(db.db))
	)
	// Collect the keys from the memory database corresponding to the given start
	for key := range db.db {
		if key >= st {
			keys = append(keys, key)
		}
	}
	// Sort the items and retrieve the associated values
	sort.Strings(keys)
	for _, key := range keys {
		values = append(values, db.db[key])
	}
	return &iterator{
		keys:   keys,
		values: values,
	}
}

// NewIteratorWithPrefix creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix.
func (db *MemoryCache) NewIteratorWithPrefix(prefix []byte) Iterator {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var (
		pr     = string(prefix)
		keys   = make([]string, 0, len(db.db))
		values = make([][]byte, 0, len(db.db))
	)
	// Collect the keys from the memory database corresponding to the given prefix
	for key := range db.db {
		if strings.HasPrefix(key, pr) {
			keys = append(keys, key)
		}
	}
	// Sort the items and retrieve the associated values
	sort.Strings(keys)
	for _, key := range keys {
		values = append(values, db.db[key])
	}
	return &iterator{
		keys:   keys,
		values: values,
	}
}

// Stat returns a particular internal stat of the database.
func (db *MemoryCache) Stat(property string) (string, error) {
	return "", errors.New("unknown property")
}

// Compact is not supported on a memory database.
func (db *MemoryCache) Compact(start []byte, limit []byte) error {
	return errors.New("unsupported operation")
}

// Len returns the number of entries currently present in the memory database.
//
// Note, this method is only used for testing (i.e. not public in general) and
// does not have explicit checks for closed-ness to allow simpler testing code.
func (db *MemoryCache) Len() int {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return len(db.db)
}

// keyvalue is a key-value tuple tagged with a deletion field to allow creating
// memory-database write batches.
type keyvalue struct {
	key    []byte
	value  []byte
	delete bool
}

// iterator can walk over the (potentially partial) keyspace of a memory key
// value store. Internally it is a deep copy of the entire iterated state,
// sorted by keys.
type iterator struct {
	inited bool
	keys   []string
	values [][]byte
}

// Next moves the iterator to the next key/value pair. It returns whether the
// iterator is exhausted.
func (it *iterator) Next() bool {
	// If the iterator was not yet initialized, do it now
	if !it.inited {
		it.inited = true
		return len(it.keys) > 0
	}
	// Iterator already initialize, advance it
	if len(it.keys) > 0 {
		it.keys = it.keys[1:]
		it.values = it.values[1:]
	}
	return len(it.keys) > 0
}

// Error returns any accumulated error. Exhausting all the key/value pairs
// is not considered to be an error. A memory iterator cannot encounter errors.
func (it *iterator) Error() error {
	return nil
}

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (it *iterator) Key() []byte {
	if len(it.keys) > 0 {
		return []byte(it.keys[0])
	}
	return nil
}

// Value returns the value of the current key/value pair, or nil if done. The
// caller should not modify the contents of the returned slice, and its contents
// may change on the next call to Next.
func (it *iterator) Value() []byte {
	if len(it.values) > 0 {
		return it.values[0]
	}
	return nil
}

// Release releases associated resources. Release should always succeed and can
// be called multiple times without causing error.
func (it *iterator) Release() {
	it.keys, it.values = nil, nil
}
