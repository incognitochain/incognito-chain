package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3StakerState struct {
	nftID     common.Hash
	liquidity uint64
	accessOTA []byte
}

func (staker *Pdexv3StakerState) NftID() common.Hash {
	return staker.nftID
}

func (staker *Pdexv3StakerState) Liquidity() uint64 {
	return staker.liquidity
}

func (staker *Pdexv3StakerState) AccessOTA() []byte {
	return staker.accessOTA
}

func (staker *Pdexv3StakerState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		NftID     common.Hash `json:"NftID"`
		Liquidity uint64      `json:"Liquidity"`
		AccessOTA []byte      `json:"AccessOTA,omitempty"`
	}{
		NftID:     staker.nftID,
		Liquidity: staker.liquidity,
		AccessOTA: staker.accessOTA,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (staker *Pdexv3StakerState) UnmarshalJSON(data []byte) error {
	temp := struct {
		NftID     common.Hash `json:"NftID"`
		Liquidity uint64      `json:"Liquidity"`
		AccessOTA []byte      `json:"AccessOTA,omitempty"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	staker.accessOTA = temp.AccessOTA
	staker.nftID = temp.NftID
	staker.liquidity = temp.Liquidity
	return nil
}

func (staker *Pdexv3StakerState) Clone() *Pdexv3StakerState {
	return &Pdexv3StakerState{
		nftID:     staker.nftID,
		liquidity: staker.liquidity,
	}
}

func NewPdexv3StakerState() *Pdexv3StakerState { return &Pdexv3StakerState{} }

func NewPdexv3StakerStateWithValue(nftID common.Hash, liquidity uint64, accessOTA []byte) *Pdexv3StakerState {
	return &Pdexv3StakerState{
		nftID:     nftID,
		liquidity: liquidity,
		accessOTA: accessOTA,
	}
}

type Pdexv3StakerObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *Pdexv3StakerState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPdexv3StakerObject(db *StateDB, hash common.Hash) *Pdexv3StakerObject {
	return &Pdexv3StakerObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewPdexv3StakerState(),
		objectType: Pdexv3StakerObjectType,
		deleted:    false,
	}
}

func newPdexv3StakerObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*Pdexv3StakerObject, error) {
	var newPdexv3StakerState = NewPdexv3StakerState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPdexv3StakerState)
		if err != nil {
			return nil, err
		}
	} else {
		newPdexv3StakerState, ok = data.(*Pdexv3StakerState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3StakerStateType, reflect.TypeOf(data))
		}
	}
	return &Pdexv3StakerObject{
		version:    defaultVersion,
		hash:       key,
		state:      newPdexv3StakerState,
		db:         db,
		objectType: Pdexv3StakerObjectType,
		deleted:    false,
	}, nil
}

func generatePdexv3StakerObjectPrefix(stakingPoolID string) []byte {
	b := append(GetPdexv3StakersPrefix(), []byte(stakingPoolID)...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GeneratePdexv3StakerObjectKey(stakingPoolID, nftID string) common.Hash {
	prefixHash := generatePdexv3StakerObjectPrefix(stakingPoolID)
	valueHash := common.HashH([]byte(nftID))
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *Pdexv3StakerObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *Pdexv3StakerObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *Pdexv3StakerObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *Pdexv3StakerObject) SetValue(data interface{}) error {
	newPdexv3StakerState, ok := data.(*Pdexv3StakerState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPdexv3ContributionStateType, reflect.TypeOf(data))
	}
	object.state = newPdexv3StakerState
	return nil
}

func (object *Pdexv3StakerObject) GetValue() interface{} {
	return object.state
}

func (object *Pdexv3StakerObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*Pdexv3StakerState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal pdexv3 staker state")
	}
	return value
}

func (object *Pdexv3StakerObject) GetHash() common.Hash {
	return object.hash
}

func (object *Pdexv3StakerObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *Pdexv3StakerObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *Pdexv3StakerObject) Reset() bool {
	object.state = NewPdexv3StakerState()
	return true
}

func (object *Pdexv3StakerObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *Pdexv3StakerObject) IsEmpty() bool {
	temp := NewPdexv3StakerState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
