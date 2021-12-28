package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3NftState struct {
	id          common.Hash
	burntAmount uint64
}

func (state *Pdexv3NftState) ID() common.Hash {
	return state.id
}

func (state *Pdexv3NftState) BurntAmount() uint64 {
	return state.burntAmount
}

func (state *Pdexv3NftState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ID          common.Hash `json:"ID"`
		BurntAmount uint64      `json:"BurntAmount"`
	}{
		ID:          state.id,
		BurntAmount: state.burntAmount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *Pdexv3NftState) UnmarshalJSON(data []byte) error {
	temp := struct {
		ID          common.Hash `json:"ID"`
		BurntAmount uint64      `json:"BurntAmount"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.id = temp.ID
	state.burntAmount = temp.BurntAmount
	return nil
}

func (state *Pdexv3NftState) Clone() *Pdexv3NftState {
	return &Pdexv3NftState{
		id:          state.id,
		burntAmount: state.burntAmount,
	}
}

func NewPdexv3NftState() *Pdexv3NftState {
	return &Pdexv3NftState{}
}

func NewPdexv3NftStateWithValue(id common.Hash, burntAmount uint64) *Pdexv3NftState {
	return &Pdexv3NftState{
		id:          id,
		burntAmount: burntAmount,
	}
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
		objectType: Pdexv3ContributionObjectType,
		deleted:    false,
	}
}

func newPdexv3NftObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*Pdexv3NftObject, error,
) {
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
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NftStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3NftObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3NftState,
		db:         db,
		objectType: Pdexv3NftObjectType,
		deleted:    false,
	}, nil
}

func GeneratePdexv3NftObjectKey(nftID string) common.Hash {
	prefixHash := GetPdexv3NftPrefix()
	valueHash := common.HashH([]byte(nftID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3NftObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3NftObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3NftObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3NftObject) SetValue(data interface{}) error {
	newPdexv3NftState, ok := data.(*Pdexv3NftState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3NftStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3NftState
	return nil
}

func (object *Pdexv3NftObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3NftObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3NftState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 nft state")
	}
	return value
}

func (object *Pdexv3NftObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3NftObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3NftObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3NftObject) Reset() bool {
	object.state = NewPdexv3NftState()
	return true
}

func (object *Pdexv3NftObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3NftObject) IsEmpty() bool {
	temp := NewPdexv3NftState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
