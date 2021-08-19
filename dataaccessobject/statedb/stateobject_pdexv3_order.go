package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type Pdexv3OrderState struct {
	poolPairID string
	value      rawdbv2.Pdexv3Order
}

func (s *Pdexv3OrderState) PoolPairID() string {
	return s.poolPairID
}

func (s *Pdexv3OrderState) Value() rawdbv2.Pdexv3Order {
	return s.value
}

func (s *Pdexv3OrderState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PoolPairID string              `json:"PoolPairID"`
		Value      *rawdbv2.Pdexv3Order `json:"Value"`
	}{
		PoolPairID: s.poolPairID,
		Value:      &s.value,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *Pdexv3OrderState) UnmarshalJSON(data []byte) error {
	temp := struct {
		PoolPairID string              `json:"PoolPairID"`
		Value      *rawdbv2.Pdexv3Order `json:"Value"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.poolPairID = temp.PoolPairID
	if temp.Value != nil {
		s.value = *temp.Value
	}
	return nil
}

func NewPdexv3OrderState() *Pdexv3OrderState {
	return &Pdexv3OrderState{}
}

func NewPdexv3OrderStateWithValue(
	pairID string, v rawdbv2.Pdexv3Order,
) *Pdexv3OrderState {
	return &Pdexv3OrderState{
		poolPairID: pairID,
		value:      v,
	}
}

func (s *Pdexv3OrderState) Clone() *Pdexv3OrderState {
	return &Pdexv3OrderState{
		poolPairID: s.poolPairID,
		value:      s.value,
	}
}

type Pdexv3OrderObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3OrderState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3OrderObject(db *StateDB, hash common.Hash) *Pdexv3OrderObject {
	return &Pdexv3OrderObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3OrderState(),
		objectType: Pdexv3OrderObjectType,
		deleted:    false,
	}
}

func newPdexv3OrderObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*Pdexv3OrderObject, error,
) {
	var newPdexv3OrderState = NewPdexv3OrderState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3OrderState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3OrderState, ok = data.(*Pdexv3OrderState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3OrderStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3OrderObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3OrderState,
		db:         db,
		objectType: Pdexv3OrderObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3OrderObjectPrefix(poolPairID string) []byte {
	str := string(GetPdexv3OrdersPrefix()) + "-" + poolPairID
	temp := []byte(str)
	h := common.HashH(temp)
	return h[:][:prefixHashKeyLength]
}

func GeneratePdexv3OrderObjectKey(poolPairID, orderID string) common.Hash {
	prefixHash := generatePdexv3OrderObjectPrefix(poolPairID)
	valueHash := common.HashH([]byte(orderID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (obj *Pdexv3OrderObject) GetVersion() int {
	return obj.version
}

// setError remembers the first non-nil error it is called with.
func (obj *Pdexv3OrderObject) SetError(err error) {
	if obj.dbErr == nil {
		obj.dbErr = err
	}
}

func (obj *Pdexv3OrderObject) GetTrie(db DatabaseAccessWarper) Trie {
	return obj.trie
}

func (obj *Pdexv3OrderObject) SetValue(data interface{}) error {
	newPdexv3OrderState, ok := data.(*Pdexv3OrderState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3OrderStateType, reflect.TypeOf(data))
	}
	obj.state = newPdexv3OrderState
	return nil
}

func (obj *Pdexv3OrderObject) GetValue() interface{} {
	return obj.state
}

func (obj *Pdexv3OrderObject) GetValueBytes() []byte {
	state, ok := obj.GetValue().(*Pdexv3OrderState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 order state")
	}
	return value
}

func (obj *Pdexv3OrderObject) GetHash() common.Hash {
	return obj.hash
}

func (obj *Pdexv3OrderObject) GetType() int {
	return obj.objectType
}

// MarkDelete will delete an object in trie
func (obj *Pdexv3OrderObject) MarkDelete() {
	obj.deleted = true
}

// reset all shard committee value into default value
func (obj *Pdexv3OrderObject) Reset() bool {
	obj.state = NewPdexv3OrderState()
	return true
}

func (obj *Pdexv3OrderObject) IsDeleted() bool {
	return obj.deleted
}

// value is either default or nil
func (obj *Pdexv3OrderObject) IsEmpty() bool {
	temp := NewPdexv3OrderState()
	return reflect.DeepEqual(temp, obj.state) || obj.state == nil
}
