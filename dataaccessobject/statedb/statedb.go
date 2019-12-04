package statedb

import (
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/incognitochain/incognito-chain/common"
	"time"
)

// StateDBs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db   DatabaseAccessWarper
	trie Trie

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects        map[common.Hash]StateObject
	stateObjectsPending map[common.Hash]struct{} // State objects finalized but not yet written to the trie
	stateObjectsDirty   map[common.Hash]struct{} // State objects modified in the current execution

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Measurements gathered during execution for debugging purposes
	StateObjectReads   time.Duration
	StateObjectHashes  time.Duration
	StateObjectUpdates time.Duration
	StateObjectCommits time.Duration
}

func New(root common.Hash, db DatabaseAccessWarper) (*StateDB, error) {
	tr, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:                  db,
		trie:                tr,
		stateObjects:        make(map[common.Hash]StateObject),
		stateObjectsPending: make(map[common.Hash]struct{}),
		stateObjectsDirty:   make(map[common.Hash]struct{}),
	}, nil
}

// setError remembers the first non-nil error it is called with.
func (stateDB *StateDB) setError(err error) {
	if stateDB.dbErr == nil {
		stateDB.dbErr = err
	}
}

func (stateDB *StateDB) Error() error {
	return stateDB.dbErr
}

// Reset clears out all ephemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (stateDB *StateDB) Reset(root common.Hash) error {
	tr, err := stateDB.db.OpenTrie(root)
	if err != nil {
		return err
	}
	stateDB.trie = tr
	stateDB.stateObjects = make(map[common.Hash]StateObject)
	stateDB.stateObjectsPending = make(map[common.Hash]struct{})
	stateDB.stateObjectsDirty = make(map[common.Hash]struct{})
	return nil
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	for addr := range s.stateObjectsPending {
		obj := s.stateObjects[addr]
		if obj.IsDeleted() {
			s.deleteStateObject(obj)
		} else {
			s.updateStateObject(obj)
		}
	}
	if len(s.stateObjectsPending) > 0 {
		s.stateObjectsPending = make(map[common.Hash]struct{})
	}
	// Track the amount of time wasted on hashing the account trie
	if metrics.EnabledExpensive {
		defer func(start time.Time) { s.StateObjectHashes += time.Since(start) }(time.Now())
	}
	return s.trie.Hash()
}

// Commit writes the state to the underlying in-memory trie database.
func (s *StateDB) Commit(deleteEmptyObjects bool) (common.Hash, error) {
	// Finalize any pending changes and merge everything into the tries
	s.IntermediateRoot(deleteEmptyObjects)

	if len(s.stateObjectsDirty) > 0 {
		s.stateObjectsDirty = make(map[common.Hash]struct{})
	}
	// Write the account trie changes, measuing the amount of wasted time
	if metrics.EnabledExpensive {
		defer func(start time.Time) { s.StateObjectCommits += time.Since(start) }(time.Now())
	}
	return s.trie.Commit(func(leaf []byte, parent common.Hash) error {
		return nil
	})
}

// ================================= STATE OBJECT =======================================
// Retrieve a state object or create a new state object if nil.
func (self *StateDB) GetOrNewStateObject(objectType int, hash common.Hash) StateObject {
	stateObject := self.getStateObject(objectType, hash)
	if stateObject == nil {
		stateObject, _ = self.createObject(objectType, hash)
	}
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given hash, it is overwritten and returned as the second return value.
func (stateDB *StateDB) createObject(objectType int, hash common.Hash) (newobj, prev StateObject) {
	prev = stateDB.getDeletedStateObject(objectType, hash) // Note, prev might have been deleted, we need that!
	newobj = newStateObject(stateDB, objectType, hash)
	stateDB.setStateObject(newobj)
	return newobj, prev
}

// add state object into statedb struct
func (stateDB *StateDB) setStateObject(object StateObject) {
	key := object.GetHash()
	stateDB.stateObjects[key] = object
}

// getStateObject retrieves a state object given by the address, returning nil if
// the object is not found or was deleted in this execution context. If you need
// to differentiate between non-existent/just-deleted, use getDeletedStateObject.
func (s *StateDB) getStateObject(objectType int, addr common.Hash) StateObject {
	if obj := s.getDeletedStateObject(objectType, addr); obj != nil && !obj.IsDeleted() {
		return obj
	}
	return nil
}

// getDeletedStateObject is similar to getStateObject, but instead of returning
// nil for a deleted state object, it returns the actual object with the deleted
// flag set. This is needed by the state journal to revert to the correct self-
// destructed object instead of wiping all knowledge about the state object.
func (stateDB *StateDB) getDeletedStateObject(objectType int, hash common.Hash) StateObject {
	// Prefer live objects if any is available
	if obj := stateDB.stateObjects[hash]; obj != nil {
		return obj
	}
	// Track the amount of time wasted on loading the object from the database
	if metrics.EnabledExpensive {
		defer func(start time.Time) { stateDB.StateObjectReads += time.Since(start) }(time.Now())
	}
	// Load the object from the database
	enc, err := stateDB.trie.TryGet(hash[:])
	if len(enc) == 0 {
		stateDB.setError(err)
		return nil
	}
	newValue := make([]byte, len(enc))
	copy(newValue, enc)
	// Insert into the live set
	obj := newStateObjectWithValue(stateDB, objectType, hash, newValue)
	stateDB.setStateObject(obj)
	return obj
}

// deleteStateObject removes the given object from the state trie.
func (s *StateDB) deleteStateObject(obj StateObject) {
	// Track the amount of time wasted on deleting the account from the trie
	if metrics.EnabledExpensive {
		defer func(start time.Time) { s.StateObjectUpdates += time.Since(start) }(time.Now())
	}
	// Delete the account from the trie
	addr := obj.GetHash()
	s.setError(s.trie.TryDelete(addr[:]))
}

// updateStateObject writes the given object to the trie.
func (s *StateDB) updateStateObject(obj StateObject) {
	// Track the amount of time wasted on updating the account from the trie
	if metrics.EnabledExpensive {
		defer func(start time.Time) { s.StateObjectUpdates += time.Since(start) }(time.Now())
	}
	// Encode the account and update the account trie
	addr := obj.GetHash()
	data := obj.GetValueBytes()
	s.setError(s.trie.TryUpdate(addr[:], data))
}
