package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3NftIndexState struct {
	nextIndex uint64
}

func (nftIndex *Pdexv3NftIndexState) NextIndex() uint64 {
	return nftIndex.nextIndex
}

func (nftIndex *Pdexv3NftIndexState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NextIndex uint64
	}{
		NextIndex: nftIndex.nextIndex,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (nftIndex *Pdexv3NftIndexState) UnmarshalJSON(data []byte) error {
	temp := struct {
		NextIndex uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	nftIndex.nextIndex = temp.NextIndex
	return nil
}

func NewPdexv3NftIndexState() *Pdexv3NftIndexState {
	return &Pdexv3NftIndexState{}
}

func NewPdexv3NftIndexWithValue(nextIndex uint64) *Pdexv3NftIndexState {
	return &Pdexv3NftIndexState{nextIndex: nextIndex}
}

type Pdexv3NftIndexObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	value      *Pdexv3NftIndexState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3NftIndexObject(db *StateDB, hash common.Hash) *Pdexv3NftIndexObject {
	return &Pdexv3NftIndexObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		value:      NewPdexv3NftIndexState(),
		objectType: Pdexv3NftIndexObjectType,
		deleted:    false,
	}
}

func newPdexv3NftIndexObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*Pdexv3NftIndexObject, error) {
	var newPdexv3NftIndexState = NewPdexv3NftIndexState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3NftIndexState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3NftIndexState, ok = data.(*Pdexv3NftIndexState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NFtIndexStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3NftIndexObject{
		version:    defaultVersion,
		hash:       key,
		value:      newPdexv3NftIndexState,
		db:         db,
		objectType: Pdexv3ParamsObjectType,
		deleted:    false,
	}, nil
}

func GeneratePdexv3NftIndexObjectKey() common.Hash {
	prefixHash := GetPdexv3NftIndexPrefix()
	return common.HashH(prefixHash)
}

func (n *Pdexv3NftIndexObject) GetVersion() int {
	return n.version
}

// setError remembers the first non-nil error it is called with.
func (n *Pdexv3NftIndexObject) SetError(err error) {
	if n.dbErr == nil {
		n.dbErr = err
	}
}

func (n *Pdexv3NftIndexObject) GetTrie(db DatabaseAccessWarper) Trie {
	return n.trie
}

func (n *Pdexv3NftIndexObject) SetValue(data interface{}) error {
	newPdexv3NftIndexState, ok := data.(*Pdexv3NftIndexState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NFtIndexStateType, reflect.TypeOf(data))
	}
	n.value = newPdexv3NftIndexState
	return nil
}

func (n *Pdexv3NftIndexObject) GetValue() interface{} {
	return n.value
}

func (n *Pdexv3NftIndexObject) GetValueBytes() []byte {
	pdexv3NftIndexState, ok := n.GetValue().(*Pdexv3NftIndexState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(pdexv3NftIndexState)
	if err != nil {
		panic("failed to marshal pdex v3 nft index state")
	}
	return value
}

func (n *Pdexv3NftIndexObject) GetHash() common.Hash {
	return n.hash
}

func (n *Pdexv3NftIndexObject) GetType() int {
	return n.objectType
}

// MarkDelete will delete an object in trie
func (n *Pdexv3NftIndexObject) MarkDelete() {
	n.deleted = true
}

// reset all shard committee value into default value
func (n *Pdexv3NftIndexObject) Reset() bool {
	n.value = NewPdexv3NftIndexState()
	return true
}

func (n *Pdexv3NftIndexObject) IsDeleted() bool {
	return n.deleted
}

// value is either default or nil
func (n *Pdexv3NftIndexObject) IsEmpty() bool {
	temp := NewPdexv3NftIndexState()
	return reflect.DeepEqual(temp, n.value) || n.value == nil
}
