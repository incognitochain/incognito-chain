package trie

import (
	"github.com/incognitochain/incognito-chain/common"
)

// PrefixTrie wraps a trie with key hashing. In a secure trie, all
// access operations hash the key using keccak256. This prevents
// calling code from creating long chains of nodes that
// increase the access time.
//
// Contrary to a regular trie, a SecureTrie can only be created with
// New and must have an attached database. The database also stores
// the preimage of each key.
//
// SecureTrie is not safe for concurrent use.
type PrefixTrie struct {
	trie             Trie
	hashKeyBuf       [common.HashSize]byte
	secKeyCache      map[string][]byte
	secKeyCacheOwner *PrefixTrie // Pointer to self, replace the key cache on mismatch
}

// NewSecure creates a trie with an existing root node from a backing database
// and optional intermediate in-memory node pool.
//
// If root is the zero hash or the sha3 hash of an empty string, the
// trie is initially empty. Otherwise, New will panic if iw is nil
// and returns MissingNodeError if the root node cannot be found.
//
// Accessing the trie loads nodes from the database or node pool on demand.
// Loaded nodes are kept around until their 'cache generation' expires.
// A new cache generation is created by each call to Commit.
// cachelimit sets the number of past cache generations to keep.
func NewPrefixTrie(root common.Hash, intermediateWriter *IntermediateWriter) (*PrefixTrie, error) {
	if intermediateWriter == nil {
		panic("trie.NewSecure called without a database")
	}
	trie, err := New(root, intermediateWriter)
	if err != nil {
		return nil, err
	}
	return &PrefixTrie{trie: *trie}, nil
}

// Get returns the value for key stored in the trie.
// The value bytes must not be modified by the caller.
func (t *PrefixTrie) Get(key []byte) []byte {
	res, err := t.TryGet(key)
	if err != nil {
		Logger.log.Errorf("Unhandled trie error: %v", err)
	}
	return res
}

// TryGet returns the value for key stored in the trie.
// The value bytes must not be modified by the caller.
// If a node was not found in the database, a MissingNodeError is returned.
func (t *PrefixTrie) TryGet(key []byte) ([]byte, error) {
	return t.trie.TryGet(key)
}

// Update associates key with value in the trie. Subsequent calls to
// Get will return value. If value has length zero, any existing value
// is deleted from the trie and calls to Get will return nil.
//
// The value bytes must not be modified by the caller while they are
// stored in the trie.
func (t *PrefixTrie) Update(key, value []byte) {
	if err := t.TryUpdate(key, value); err != nil {
		Logger.log.Error("Unhandled trie error: %v", err)
	}
}

// TryUpdate associates key with value in the trie. Subsequent calls to
// Get will return value. If value has length zero, any existing value
// is deleted from the trie and calls to Get will return nil.
//
// The value bytes must not be modified by the caller while they are
// stored in the trie.
//
// If a node was not found in the database, a MissingNodeError is returned.
func (t *PrefixTrie) TryUpdate(key, value []byte) error {
	//hk := t.hashKey(key)
	err := t.trie.TryUpdate(key, value)
	if err != nil {
		return err
	}
	t.getSecKeyCache()[string(key)] = common.CopyBytes(key)
	return nil
}

// Delete removes any existing value for key from the trie.
func (t *PrefixTrie) Delete(key []byte) {
	if err := t.TryDelete(key); err != nil {
		Logger.log.Error("Unhandled trie error: %v", err)
	}
}

// TryDelete removes any existing value for key from the trie.
// If a node was not found in the database, a MissingNodeError is returned.
func (t *PrefixTrie) TryDelete(key []byte) error {
	//hk := t.hashKey(key)
	delete(t.getSecKeyCache(), string(key))
	return t.trie.TryDelete(key)
}

// GetKey returns the sha3 preimage of a hashed key that was
// previously used to store a value.
func (t *PrefixTrie) GetKey(shaKey []byte) []byte {
	if key, ok := t.getSecKeyCache()[string(shaKey)]; ok {
		return key
	}
	key, _ := t.trie.iw.preimage(common.BytesToHash(shaKey))
	return key
}

// Commit writes all nodes and the secure hash pre-images to the trie's database.
// Nodes are stored with their sha3 hash as the key.
//
// Committing flushes nodes from memory. Subsequent Get calls will load nodes
// from the database.
func (t *PrefixTrie) Commit(onleaf LeafCallback) (root common.Hash, err error) {
	// Write all the pre-images to the actual disk database
	if len(t.getSecKeyCache()) > 0 {
		t.trie.iw.lock.Lock()
		for hk, key := range t.secKeyCache {
			t.trie.iw.insertPreimage(common.BytesToHash([]byte(hk)), key)
		}
		t.trie.iw.lock.Unlock()

		t.secKeyCache = make(map[string][]byte)
	}
	// Commit the trie to its intermediate node database
	return t.trie.Commit(onleaf)
}

// Hash returns the root hash of PrefixTrie. It does not write to the
// database and can be used even if the trie doesn't have one.
func (t *PrefixTrie) Hash() common.Hash {
	return t.trie.Hash()
}

// Copy returns a copy of PrefixTrie.
func (t *PrefixTrie) Copy() *PrefixTrie {
	cpy := *t
	return &cpy
}

// NodeIterator returns an iterator that returns nodes of the underlying trie. Iteration
// starts at the key after the given start key.
func (t *PrefixTrie) NodeIterator(start []byte) NodeIterator {
	return t.trie.NodeIterator(start)
}

// hashKey returns the hash of key as an ephemeral buffer.
// The caller must not hold onto the return value because it will become
// invalid on the next call to hashKey or secKey.
func (t *PrefixTrie) hashKey(key []byte) []byte {
	h := newHasher(nil)
	h.sha.Reset()
	h.sha.Write(key)
	buf := h.sha.Sum(t.hashKeyBuf[:0])
	returnHasherToPool(h)
	return buf
}

// getSecKeyCache returns the current secure key cache, creating a new one if
// ownership changed (i.e. the current secure trie is a copy of another owning
// the actual cache).
func (t *PrefixTrie) getSecKeyCache() map[string][]byte {
	if t != t.secKeyCacheOwner {
		t.secKeyCacheOwner = t
		t.secKeyCache = make(map[string][]byte)
	}
	return t.secKeyCache
}
