package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type Pdexv3PoolPairState struct {
	poolPairID string
	value      rawdbv2.Pdexv3PoolPair
}

func (pp *Pdexv3PoolPairState) PoolPairID() string {
	return pp.poolPairID
}

func (pp *Pdexv3PoolPairState) Value() rawdbv2.Pdexv3PoolPair {
	return pp.value
}

func (pp *Pdexv3PoolPairState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID string                 `json:"PoolPairID"`
		Value      rawdbv2.Pdexv3PoolPair `json:"Value"`
	}{
		PoolPairID: pp.poolPairID,
		Value:      pp.value,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *Pdexv3PoolPairState) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID string                 `json:"PoolPairID"`
		Value      rawdbv2.Pdexv3PoolPair `json:"Value"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.poolPairID = temp.PoolPairID
	pp.value = temp.Value
	return nil
}

func (pp *Pdexv3PoolPairState) Clone() *Pdexv3PoolPairState {
	return &Pdexv3PoolPairState{
		poolPairID: pp.poolPairID,
		value:      *pp.value.Clone(),
	}
}

func NewPdexv3PoolPairState() *Pdexv3PoolPairState {
	return &Pdexv3PoolPairState{}
}

func NewPdexv3PoolPairStateWithValue(
	poolPairID string, value rawdbv2.Pdexv3PoolPair,
) *Pdexv3PoolPairState {
	return &Pdexv3PoolPairState{
		poolPairID: poolPairID,
		value:      value,
	}
}

type Pdexv3PoolPairObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairObject {
	return &Pdexv3PoolPairObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairState(),
		objectType: Pdexv3PoolPairObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*Pdexv3PoolPairObject, error,
) {
	var newPdexv3PoolPairState = NewPdexv3PoolPairState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairState, ok = data.(*Pdexv3PoolPairState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairState,
		db:         db,
		objectType: Pdexv3PoolPairObjectType,
		deleted:    false,
	}, nil
}

func GeneratePdexv3PoolPairObjectKey(poolPairID string) common.Hash {
	prefixHash := GetPdexv3PoolPairsPrefix()
	valueHash := common.HashH([]byte(poolPairID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (pp *Pdexv3PoolPairObject) GetVersion() int {
	return pp.version
}

// setError remembers the first non-nil error it is called with.
func (pp *Pdexv3PoolPairObject) SetError(err error) {
	if pp.dbErr == nil {
		pp.dbErr = err
	}
}

func (pp *Pdexv3PoolPairObject) GetTrie(db DatabaseAccessWarper) Trie {
	return pp.trie
}

func (pp *Pdexv3PoolPairObject) SetValue(data interface{}) error {
	newPdexv3PoolPairState, ok := data.(*Pdexv3PoolPairState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPDEPoolPairStateType, reflect.TypeOf(data))
	}
	pp.state = newPdexv3PoolPairState
	return nil
}

func (pp *Pdexv3PoolPairObject) GetValue() interface{} {
	return pp.state
}

func (pp *Pdexv3PoolPairObject) GetValueBytes() []byte {
	state, ok := pp.GetValue().(*Pdexv3PoolPairState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair state")
	}
	return value
}

func (pp *Pdexv3PoolPairObject) GetHash() common.Hash {
	return pp.hash
}

func (pp *Pdexv3PoolPairObject) GetType() int {
	return pp.objectType
}

// MarkDelete will delete an object in trie
func (pp *Pdexv3PoolPairObject) MarkDelete() {
	pp.deleted = true
}

// reset all shard committee value into default value
func (pp *Pdexv3PoolPairObject) Reset() bool {
	pp.state = NewPdexv3PoolPairState()
	return true
}

func (pp *Pdexv3PoolPairObject) IsDeleted() bool {
	return pp.deleted
}

// value is either default or nil
func (pp *Pdexv3PoolPairObject) IsEmpty() bool {
	temp := NewPdexv3PoolPairState()
	return reflect.DeepEqual(temp, pp.state) || pp.state == nil
}
