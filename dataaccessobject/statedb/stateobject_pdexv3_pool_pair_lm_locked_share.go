package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairLmLockedShareState struct {
	nftID        string
	beaconHeight uint64
	amount       uint64
}

func (state *Pdexv3PoolPairLmLockedShareState) BeaconHeight() uint64 {
	return state.amount
}

func (state *Pdexv3PoolPairLmLockedShareState) NftID() string {
	return state.nftID
}

func (state *Pdexv3PoolPairLmLockedShareState) Amount() uint64 {
	return state.amount
}

func (state *Pdexv3PoolPairLmLockedShareState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NftID        string `json:"NftID"`
		BeaconHeight uint64 `json:"BeaconHeight"`
		Amount       uint64 `json:"Amount"`
	}{
		NftID:        state.nftID,
		BeaconHeight: state.beaconHeight,
		Amount:       state.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *Pdexv3PoolPairLmLockedShareState) UnmarshalJSON(data []byte) error {
	temp := struct {
		NftID        string `json:"NftID"`
		BeaconHeight uint64 `json:"BeaconHeight"`
		Amount       uint64 `json:"Amount"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.beaconHeight = temp.BeaconHeight
	state.nftID = temp.NftID
	state.amount = temp.Amount
	return nil
}

func (state *Pdexv3PoolPairLmLockedShareState) Clone() *Pdexv3PoolPairLmLockedShareState {
	return &Pdexv3PoolPairLmLockedShareState{
		nftID:        state.nftID,
		beaconHeight: state.beaconHeight,
		amount:       state.amount,
	}
}

func NewPdexv3PoolPairLmLockedShareState() *Pdexv3PoolPairLmLockedShareState {
	return &Pdexv3PoolPairLmLockedShareState{}
}

func NewPdexv3PoolPairLmLockedShareStateWithValue(
	nftID string, beaconHeight, amount uint64,
) *Pdexv3PoolPairLmLockedShareState {
	return &Pdexv3PoolPairLmLockedShareState{
		nftID:        nftID,
		beaconHeight: beaconHeight,
		amount:       amount,
	}
}

type Pdexv3PoolPairLmLockedShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairLmLockedShareState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairLmLockedShareObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairLmLockedShareObject {
	return &Pdexv3PoolPairLmLockedShareObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairLmLockedShareState(),
		objectType: Pdexv3PoolPairLmLockedShareObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairLmLockedShareObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairLmLockedShareObject, error) {
	var newPdexv3PoolPairLmLockedShareState = NewPdexv3PoolPairLmLockedShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairLmLockedShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairLmLockedShareState, ok = data.(*Pdexv3PoolPairLmLockedShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairLmLockedShareStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairLmLockedShareObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairLmLockedShareState,
		db:         db,
		objectType: Pdexv3PoolPairLmLockedShareObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairLmLockedShareObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3PoolPairLmLockedSharePrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairLmLockedShareObjectKey(poolPairID, nftID string, beaconHeight uint64) common.Hash {
	prefixHash := generatePdexv3PoolPairLmLockedShareObjectPrefix(poolPairID)
	valueHash := common.HashH(append([]byte(nftID), common.Uint64ToBytes(beaconHeight)...))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairLmLockedShareObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairLmLockedShareObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairLmLockedShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairLmLockedShareObject) SetValue(data interface{}) error {
	newPdexv3PoolPairLmLockedShareState, ok := data.(*Pdexv3PoolPairLmLockedShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairLmLockedShareStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairLmLockedShareState
	return nil
}

func (object *Pdexv3PoolPairLmLockedShareObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairLmLockedShareObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairLmLockedShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair lp fee per share state")
	}
	return value
}

func (object *Pdexv3PoolPairLmLockedShareObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairLmLockedShareObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairLmLockedShareObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairLmLockedShareObject) Reset() bool {
	object.state = NewPdexv3PoolPairLmLockedShareState()
	return true
}

func (object *Pdexv3PoolPairLmLockedShareObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairLmLockedShareObject) IsEmpty() bool {
	temp := NewPdexv3PoolPairLmLockedShareState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
