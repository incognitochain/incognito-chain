package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3NftState struct {
	id common.Hash
}

func (nft *Pdexv3NftState) ID() common.Hash {
	return nft.id
}

func (nft *Pdexv3NftState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ID common.Hash `json:"ID"`
	}{
		ID: nft.id,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (nft *Pdexv3NftState) UnmarshalJSON(data []byte) error {
	temp := struct {
		ID common.Hash `json:"ID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	nft.id = temp.ID
	return nil
}

func (nft *Pdexv3NftState) Clone() *Pdexv3NftState {
	return &Pdexv3NftState{
		id: nft.ID(),
	}
}

func NewPdexv3NftState() *Pdexv3NftState {
	return &Pdexv3NftState{}
}

func NewPdexv3NftStateWithValue(id common.Hash) *Pdexv3NftState {
	return &Pdexv3NftState{id: id}
}

type Pdexv3NftObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3NftState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3NftObject(db *StateDB, hash common.Hash) *Pdexv3NftObject {
	return &Pdexv3NftObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3NftState(),
		objectType: Pdexv3NftObjectType,
		deleted:    false,
	}
}

func newPdexv3NftObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*Pdexv3NftObject, error) {
	var newPdexv3NftState = NewPdexv3NftState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3NftState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3NftState, ok = data.(*Pdexv3NftState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NFtStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3NftObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3NftState,
		db:         db,
		objectType: Pdexv3PoolPairObjectType,
		deleted:    false,
	}, nil
}

func GeneratePdexv3NftObjectKey(id common.Hash) common.Hash {
	prefixHash := GetPdexv3NftPrefix()
	valueHash := id
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (nft *Pdexv3NftObject) GetVersion() int {
	return nft.version
}

// setError remembers the first non-nil error it is called with.
func (nft *Pdexv3NftObject) SetError(err error) {
	if nft.dbErr == nil {
		nft.dbErr = err
	}
}

func (nft *Pdexv3NftObject) GetTrie(db DatabaseAccessWarper) Trie {
	return nft.trie
}

func (nft *Pdexv3NftObject) SetValue(data interface{}) error {
	newPdexv3NftState, ok := data.(*Pdexv3NftState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NFtStateType, reflect.TypeOf(data))
	}
	nft.state = newPdexv3NftState
	return nil
}

func (nft *Pdexv3NftObject) GetValue() interface{} {
	return nft.state
}

func (nft *Pdexv3NftObject) GetValueBytes() []byte {
	state, ok := nft.GetValue().(*Pdexv3NftObject)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 nft state")
	}
	return value
}

func (nft *Pdexv3NftObject) GetHash() common.Hash {
	return nft.hash
}

func (nft *Pdexv3NftObject) GetType() int {
	return nft.objectType
}

// MarkDelete will delete an object in trie
func (nft *Pdexv3NftObject) MarkDelete() {
	nft.deleted = true
}

// reset all shard committee value into default value
func (nft *Pdexv3NftObject) Reset() bool {
	nft.state = NewPdexv3NftState()
	return true
}

func (nft *Pdexv3NftObject) IsDeleted() bool {
	return nft.deleted
}

// value is either default or nil
func (nft *Pdexv3NftObject) IsEmpty() bool {
	temp := NewPdexv3NftState()
	return reflect.DeepEqual(temp, nft.state) || nft.state == nil
}
