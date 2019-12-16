package statedb

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/trie"
	"time"
)

// StateDBs within the incognito protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * State Object
type StateDB struct {
	db    DatabaseAccessWarper
	trie  Trie
	rawdb incdb.Database
	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects        map[common.Hash]StateObject
	stateObjectsPending map[common.Hash]struct{} // State objects finalized but not yet written to the trie
	stateObjectsDirty   map[common.Hash]struct{} // State objects modified in the current execution

	// DB error.
	// State objects are used by the consensus core which are
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

// New return a new statedb attach with a state root
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

// New return a new statedb attach with a state root
func NewWithRawDB(root common.Hash, db DatabaseAccessWarper, rawdb incdb.Database) (*StateDB, error) {
	tr, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:                  db,
		trie:                tr,
		rawdb:               rawdb,
		stateObjects:        make(map[common.Hash]StateObject),
		stateObjectsPending: make(map[common.Hash]struct{}),
		stateObjectsDirty:   make(map[common.Hash]struct{}),
	}, nil
}

// New return a new statedb attach with a state root
func NewWithPrefixTrie(root common.Hash, db DatabaseAccessWarper) (*StateDB, error) {
	tr, err := db.OpenPrefixTrie(root)
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

// Error return statedb error
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
func (stateDB *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	stateDB.markDeleteEmptyStateObject(deleteEmptyObjects)
	for addr := range stateDB.stateObjectsPending {
		obj := stateDB.stateObjects[addr]
		if obj.IsDeleted() {
			stateDB.deleteStateObject(obj)
		} else {
			stateDB.updateStateObject(obj)
		}
	}
	if len(stateDB.stateObjectsPending) > 0 {
		stateDB.stateObjectsPending = make(map[common.Hash]struct{})
	}
	// Track the amount of time wasted on hashing the account trie
	if metrics.EnabledExpensive {
		defer func(start time.Time) { stateDB.StateObjectHashes += time.Since(start) }(time.Now())
	}
	return stateDB.trie.Hash()
}
func (stateDB *StateDB) markDeleteEmptyStateObject(deleteEmptyObjects bool) {
	for _, object := range stateDB.stateObjects {
		if object.IsEmpty() {
			object.MarkDelete()
		}
	}
}

// Commit writes the state to the underlying in-memory trie database.
func (stateDB *StateDB) Commit(deleteEmptyObjects bool) (common.Hash, error) {
	// Finalize any pending changes and merge everything into the tries
	stateDB.IntermediateRoot(deleteEmptyObjects)

	if len(stateDB.stateObjectsDirty) > 0 {
		stateDB.stateObjectsDirty = make(map[common.Hash]struct{})
	}
	// Write the account trie changes, measuing the amount of wasted time
	if metrics.EnabledExpensive {
		defer func(start time.Time) { stateDB.StateObjectCommits += time.Since(start) }(time.Now())
	}
	return stateDB.trie.Commit(func(leaf []byte, parent common.Hash) error {
		return nil
	})
}

// Database return current database access warper
func (stateDB *StateDB) Database() DatabaseAccessWarper {
	return stateDB.db
}

// TODO: implement duplicate current statedb
// Copy duplicate statedb and return new statedb instance
func (stateDB *StateDB) Copy() *StateDB {
	return &StateDB{}
}

// Exist check existence of a state object in statedb
func (stateDB *StateDB) Exist(objectType int, stateObjectHash common.Hash) (bool, error) {
	value, err := stateDB.getStateObject(objectType, stateObjectHash)
	if err != nil {
		return false, err
	}
	return value != nil, nil
}

// Empty check a state object in statedb is empty or not
func (stateDB *StateDB) Empty(objectType int, stateObjectHash common.Hash) bool {
	stateObject, err := stateDB.getStateObject(objectType, stateObjectHash)
	return stateObject == nil || stateObject.IsEmpty() || err != nil
}

// ================================= STATE OBJECT =======================================
// getDeletedStateObject is similar to getStateObject, but instead of returning
// nil for a deleted state object, it returns the actual object with the deleted
// flag set. This is needed by the state journal to revert to the correct self-
// destructed object instead of wiping all knowledge about the state object.
func (stateDB *StateDB) getDeletedStateObject(objectType int, hash common.Hash) (StateObject, error) {
	// Prefer live objects if any is available
	if obj := stateDB.stateObjects[hash]; obj != nil {
		return obj, nil
	}
	// Track the amount of time wasted on loading the object from the database
	if metrics.EnabledExpensive {
		defer func(start time.Time) { stateDB.StateObjectReads += time.Since(start) }(time.Now())
	}
	// Load the object from the database
	enc, err := stateDB.trie.TryGet(hash[:])
	if len(enc) == 0 {
		stateDB.setError(err)
		return nil, nil
	}
	newValue := make([]byte, len(enc))
	copy(newValue, enc)
	// Insert into the live set
	obj, err := newStateObjectWithValue(stateDB, objectType, hash, newValue)
	if err != nil {
		return nil, err
	}
	stateDB.setStateObject(obj)
	return obj, nil
}

// updateStateObject writes the given object to the trie.
func (stateDB *StateDB) updateStateObject(obj StateObject) {
	// Track the amount of time wasted on updating the account from the trie
	if metrics.EnabledExpensive {
		defer func(start time.Time) { stateDB.StateObjectUpdates += time.Since(start) }(time.Now())
	}
	// Encode the account and update the account trie
	addr := obj.GetHash()
	data := obj.GetValueBytes()
	stateDB.setError(stateDB.trie.TryUpdate(addr[:], data))
}

// deleteStateObject removes the given object from the state trie.
func (stateDB *StateDB) deleteStateObject(obj StateObject) {
	// Track the amount of time wasted on deleting the account from the trie
	if metrics.EnabledExpensive {
		defer func(start time.Time) { stateDB.StateObjectUpdates += time.Since(start) }(time.Now())
	}
	// Delete the account from the trie
	addr := obj.GetHash()
	stateDB.setError(stateDB.trie.TryDelete(addr[:]))
}

// createStateObject creates a new state object. If there is an existing account with
// the given hash, it is overwritten and returned as the second return value.
func (stateDB *StateDB) createStateObject(objectType int, hash common.Hash) (newobj, prev StateObject, err error) {
	prev, err = stateDB.getDeletedStateObject(objectType, hash) // Note, prev might have been deleted, we need that!
	if err != nil {
		return nil, nil, err
	}
	newobj = newStateObject(stateDB, objectType, hash)
	stateDB.stateObjectsPending[hash] = struct{}{}
	stateDB.setStateObject(newobj)
	return newobj, prev, err
}

func (stateDB *StateDB) createStateObjectWithValue(objectType int, hash common.Hash, value interface{}) (newobj, prev StateObject, err error) {
	newobj, err = newStateObjectWithValue(stateDB, objectType, hash, value)
	if err != nil {
		return nil, nil, err
	}
	stateDB.stateObjectsPending[hash] = struct{}{}
	stateDB.setStateObject(newobj)
	return newobj, prev, err
}

// SetStateObject add new stateobject into statedb
func (stateDB *StateDB) SetStateObject(objectType int, key common.Hash, value interface{}) error {
	obj, err := stateDB.getOrNewStateObjectWithValue(objectType, key, value)
	if err != nil {
		return err
	}
	err = obj.SetValue(value)
	if err != nil {
		return err
	}
	stateDB.stateObjectsPending[key] = struct{}{}
	return nil
}

// MarkDeleteStateObject add new stateobject into statedb
func (stateDB *StateDB) MarkDeleteStateObject(objectType int, key common.Hash) bool {
	stateObject, err := stateDB.getStateObject(objectType, key)
	if err == nil && stateObject != nil {
		stateObject.MarkDelete()
		stateDB.stateObjectsPending[key] = struct{}{}
		return true
	}
	return false
}

// Retrieve a state object or create a new state object if nil.
func (stateDB *StateDB) getOrNewStateObject(objectType int, hash common.Hash) (StateObject, error) {
	stateObject, err := stateDB.getStateObject(objectType, hash)
	if err != nil {
		return nil, err
	}
	if stateObject == nil {
		stateObject, _, err = stateDB.createStateObject(objectType, hash)
		if err != nil {
			return nil, err
		}
	}
	return stateObject, nil
}

func (stateDB *StateDB) getOrNewStateObjectWithValue(objectType int, hash common.Hash, value interface{}) (StateObject, error) {
	stateObject, err := stateDB.getStateObject(objectType, hash)
	if err != nil {
		return nil, err
	}
	if stateObject == nil {
		stateObject, _, err = stateDB.createStateObjectWithValue(objectType, hash, value)
		if err != nil {
			return nil, err
		}
	}
	return stateObject, nil
}

// add state object into statedb struct
func (stateDB *StateDB) setStateObject(object StateObject) {
	key := object.GetHash()
	stateDB.stateObjects[key] = object
}

// getStateObject retrieves a state object given by the address, returning nil if
// the object is not found or was deleted in this execution context. If you need
// to differentiate between non-existent/just-deleted, use getDeletedStateObject.
func (stateDB *StateDB) getStateObject(objectType int, addr common.Hash) (StateObject, error) {
	if obj, err := stateDB.getDeletedStateObject(objectType, addr); obj != nil && !obj.IsDeleted() {
		return obj, nil
	} else if err != nil {
		return nil, err
	}
	return nil, nil
}

// FOR TEST ONLY
// do not use this function for build feature
func (stateDB *StateDB) GetStateObjectMapForTestOnly() map[common.Hash]StateObject {
	return stateDB.stateObjects
}
func (stateDB *StateDB) GetStateObjectPendingMapForTestOnly() map[common.Hash]struct{} {
	return stateDB.stateObjectsPending
}

// =================================     Test Object     ========================================
func (stateDB *StateDB) GetTestObject(key common.Hash) ([]byte, error) {
	testObject, err := stateDB.getStateObject(TestObjectType, key)
	if err != nil {
		return []byte{}, err
	}
	if testObject != nil {
		return testObject.GetValueBytes(), nil
	}
	return []byte{}, nil
}
func (stateDB *StateDB) GetAllTestObjectList() ([]common.Hash, [][]byte) {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)
	keys := []common.Hash{}
	values := [][]byte{}
	for it.Next() {
		key := stateDB.trie.GetKey(it.Key)
		newKey := make([]byte, len(key))
		copy(newKey, key)
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		keys = append(keys, common.BytesToHash(key))
		values = append(values, value)
	}
	return keys, values
}
func (stateDB *StateDB) GetAllTestObjectMap() map[common.Hash][]byte {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)
	m := make(map[common.Hash][]byte)
	for it.Next() {
		key := stateDB.trie.GetKey(it.Key)
		newKey := make([]byte, len(key))
		copy(newKey, key)
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		m[common.BytesToHash(key)] = newValue
	}
	return m
}
func (stateDB *StateDB) GetByPrefixTestObjectList(prefix []byte) ([]common.Hash, [][]byte) {
	temp := stateDB.trie.NodeIterator(prefix)
	it := trie.NewIterator(temp)
	keys := []common.Hash{}
	values := [][]byte{}
	for it.Next() {
		key := stateDB.trie.GetKey(it.Key)
		newKey := make([]byte, len(key))
		copy(newKey, key)
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		keys = append(keys, common.BytesToHash(key))
		values = append(values, value)
	}
	return keys, values
}

// ================================= Serial Number OBJECT =======================================
func (stateDB *StateDB) GetSerialNumber(key common.Hash) ([]byte, error) {
	serialNumberObject, err := stateDB.getStateObject(SerialNumberObjectType, key)
	if err != nil {
		return []byte{}, err
	}
	if serialNumberObject != nil {
		return serialNumberObject.GetValueBytes(), nil
	}
	return []byte{}, nil
}

func (stateDB *StateDB) GetAllSerialNumberKeyValueList() ([]common.Hash, [][]byte) {
	temp := stateDB.trie.NodeIterator(GetSerialNumberPrefix())
	it := trie.NewIterator(temp)
	keys := []common.Hash{}
	values := [][]byte{}
	for it.Next() {
		key := stateDB.trie.GetKey(it.Key)
		newKey := make([]byte, len(key))
		copy(newKey, key)
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		keys = append(keys, common.BytesToHash(key))
		values = append(values, value)
	}
	return keys, values
}
func (stateDB *StateDB) GetAllSerialNumberValueList() [][]byte {
	temp := stateDB.trie.NodeIterator(GetSerialNumberPrefix())
	it := trie.NewIterator(temp)
	values := [][]byte{}
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		values = append(values, value)
	}
	return values
}

// ================================= Committee OBJECT =======================================
func (stateDB *StateDB) GetCommitteeState(key common.Hash) (*CommitteeState, bool, error) {
	committeeStateObject, err := stateDB.getStateObject(CommitteeObjectType, key)
	if err != nil {
		return nil, false, err
	}
	if committeeStateObject != nil {
		return committeeStateObject.GetValue().(*CommitteeState), true, nil
	}
	return NewCommitteeState(), false, nil
}
func (stateDB *StateDB) GetAllCommitteeState(ids []int) map[int][]incognitokey.CommitteePublicKey {
	m := make(map[int][]incognitokey.CommitteePublicKey)
	for _, id := range ids {
		prefix := GetCommitteePrefixByShardID(id)
		temp := stateDB.trie.NodeIterator(prefix)
		it := trie.NewIterator(temp)
		for it.Next() {
			value := it.Value
			newValue := make([]byte, len(value))
			copy(newValue, value)
			committeeState := NewCommitteeState()
			err := json.Unmarshal(newValue, committeeState)
			if err != nil {
				panic("wrong value type")
			}
			m[committeeState.ShardID] = append(m[committeeState.ShardID], committeeState.CommitteePublicKey)
		}
	}
	return m
}
func (stateDB *StateDB) GetByShardIDCommitteeState(shardID int) []incognitokey.CommitteePublicKey {
	committees := []incognitokey.CommitteePublicKey{}
	prefix := GetCommitteePrefixByShardID(shardID)
	temp := stateDB.trie.NodeIterator(prefix)
	it := trie.NewIterator(temp)
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		committeeState := NewCommitteeState()
		err := json.Unmarshal(newValue, committeeState)
		if err != nil {
			panic("wrong value type")
		}
		if committeeState.ShardID != shardID {
			panic("wrong expected shard id")
		}
		committees = append(committees, committeeState.CommitteePublicKey)
	}
	return committees
}

// ================================= Reward Receiver OBJECT =======================================
func (stateDB *StateDB) GetRewardReceiverState(key common.Hash) (*RewardReceiverState, bool, error) {
	rewardReceiverObject, err := stateDB.getStateObject(RewardReceiverObjectType, key)
	if err != nil {
		return nil, false, err
	}
	if rewardReceiverObject != nil {
		return rewardReceiverObject.GetValue().(*RewardReceiverState), true, nil
	}
	return NewRewardReceiverState(), false, nil
}
func (stateDB *StateDB) GetAllRewardReceiverState() map[string]string {
	m := make(map[string]string)
	prefix := GetRewardReceiverPrefix()
	temp := stateDB.trie.NodeIterator(prefix)
	it := trie.NewIterator(temp)
	for it.Next() {
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		rewardReceiverState := NewRewardReceiverState()
		err := json.Unmarshal(newValue, rewardReceiverState)
		if err != nil {
			panic("wrong value type")
		}
		m[rewardReceiverState.PublicKey] = rewardReceiverState.PaymentAddress
	}
	return m
}
