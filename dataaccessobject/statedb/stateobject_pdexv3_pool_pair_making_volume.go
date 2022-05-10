package statedb

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairMakingVolumeState struct {
	nftID   string
	tokenID common.Hash
	value   *big.Int
}

func (state *Pdexv3PoolPairMakingVolumeState) Value() big.Int {
	return *state.value
}

func (state *Pdexv3PoolPairMakingVolumeState) NftID() string {
	return state.nftID
}

func (state *Pdexv3PoolPairMakingVolumeState) TokenID() common.Hash {
	return state.tokenID
}

func (state *Pdexv3PoolPairMakingVolumeState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NftID   string      `json:"NftID"`
		TokenID common.Hash `json:"TokenID"`
		Value   *big.Int    `json:"Value"`
	}{
		NftID:   state.nftID,
		TokenID: state.tokenID,
		Value:   state.value,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *Pdexv3PoolPairMakingVolumeState) UnmarshalJSON(data []byte) error {
	temp := struct {
		NftID   string      `json:"NftID"`
		TokenID common.Hash `json:"TokenID"`
		Value   *big.Int    `json:"Value"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.nftID = temp.NftID
	state.tokenID = temp.TokenID
	state.value = temp.Value
	return nil
}

func (state *Pdexv3PoolPairMakingVolumeState) Clone() *Pdexv3PoolPairMakingVolumeState {
	return &Pdexv3PoolPairMakingVolumeState{
		nftID:   state.nftID,
		tokenID: state.tokenID,
		value:   big.NewInt(0).Set(state.value),
	}
}

func NewPdexv3PoolPairMakingVolumeState() *Pdexv3PoolPairMakingVolumeState {
	return &Pdexv3PoolPairMakingVolumeState{}
}

func NewPdexv3PoolPairMakingVolumeStateWithValue(
	nftID string, tokenID common.Hash, value *big.Int,
) *Pdexv3PoolPairMakingVolumeState {
	return &Pdexv3PoolPairMakingVolumeState{
		nftID:   nftID,
		tokenID: tokenID,
		value:   value,
	}
}

type Pdexv3PoolPairMakingVolumeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3PoolPairMakingVolumeState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3PoolPairMakingVolumeObject(db *StateDB, hash common.Hash) *Pdexv3PoolPairMakingVolumeObject {
	return &Pdexv3PoolPairMakingVolumeObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3PoolPairMakingVolumeState(),
		objectType: Pdexv3PoolPairMakingVolumeObjectType,
		deleted:    false,
	}
}

func newPdexv3PoolPairMakingVolumeObjectWithValue(
	db *StateDB, key common.Hash, data interface{},
) (*Pdexv3PoolPairMakingVolumeObject, error) {
	var newPdexv3PoolPairMakingVolumeState = NewPdexv3PoolPairMakingVolumeState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3PoolPairMakingVolumeState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3PoolPairMakingVolumeState, ok = data.(*Pdexv3PoolPairMakingVolumeState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairMakingVolumeStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3PoolPairMakingVolumeObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3PoolPairMakingVolumeState,
		db:         db,
		objectType: Pdexv3PoolPairMakingVolumeObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3PoolPairMakingVolumeObjectPrefix(poolPairID string) []byte {
	b := append(GetPdexv3PoolPairMakingVolumePrefix(), []byte(poolPairID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3PoolPairMakingVolumeObjectKey(poolPairID string, tokenID common.Hash, nftID string) common.Hash {
	prefixHash := generatePdexv3PoolPairMakingVolumeObjectPrefix(poolPairID)
	valueHash := common.HashH(append([]byte(tokenID.String()), []byte(nftID)...))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3PoolPairMakingVolumeObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3PoolPairMakingVolumeObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3PoolPairMakingVolumeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3PoolPairMakingVolumeObject) SetValue(data interface{}) error {
	newPdexv3PoolPairMakingVolumeState, ok := data.(*Pdexv3PoolPairMakingVolumeState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3PoolPairMakingVolumeStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3PoolPairMakingVolumeState
	return nil
}

func (object *Pdexv3PoolPairMakingVolumeObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3PoolPairMakingVolumeObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3PoolPairMakingVolumeState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 pool pair making volume state")
	}
	return value
}

func (object *Pdexv3PoolPairMakingVolumeObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3PoolPairMakingVolumeObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3PoolPairMakingVolumeObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3PoolPairMakingVolumeObject) Reset() bool {
	object.state = NewPdexv3PoolPairMakingVolumeState()
	return true
}

func (object *Pdexv3PoolPairMakingVolumeObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3PoolPairMakingVolumeObject) IsEmpty() bool {
	return reflect.DeepEqual(NewPdexv3PoolPairMakingVolumeState(), object.state) || object.state == nil
}
