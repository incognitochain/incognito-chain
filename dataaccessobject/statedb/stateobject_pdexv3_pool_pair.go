package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type Pdexv3PoolPairState struct {
	token0ID              string
	token1ID              string
	token0RealAmount      uint64
	token1RealAmount      uint64
	currentContributionID uint64
	token0VirtualAmount   uint64
	token1VirtualAmount   uint64
	amplifier             uint
}

func (pp *Pdexv3PoolPairState) Amplifier() uint {
	return pp.amplifier
}

func (pp *Pdexv3PoolPairState) Token0ID() string {
	return pp.token0ID
}

func (pp *Pdexv3PoolPairState) Token1ID() string {
	return pp.token1ID
}

func (pp *Pdexv3PoolPairState) Token0RealAmount() uint64 {
	return pp.token0RealAmount
}

func (pp *Pdexv3PoolPairState) Token1RealAmount() uint64 {
	return pp.token1RealAmount
}

func (pp *Pdexv3PoolPairState) CurrentContributionID() uint64 {
	return pp.currentContributionID
}

func (pp *Pdexv3PoolPairState) Token0VirtualAmount() uint64 {
	return pp.token0VirtualAmount
}

func (pp *Pdexv3PoolPairState) Token1VirtualAmount() uint64 {
	return pp.token1VirtualAmount
}

func (pp *Pdexv3PoolPairState) SetToken0RealAmount(amount uint64) {
	pp.token0RealAmount = amount
}

func (pp *Pdexv3PoolPairState) SetToken1RealAmount(amount uint64) {
	pp.token1RealAmount = amount
}

func (pp *Pdexv3PoolPairState) SetCurrentContributionID(id uint64) {
	pp.currentContributionID = id
}

func (pp *Pdexv3PoolPairState) SetToken0VirtualAmount(amount uint64) {
	pp.token0VirtualAmount = amount
}

func (pp *Pdexv3PoolPairState) SetToken1VirtualAmount(amount uint64) {
	pp.token1VirtualAmount = amount
}

func (pp *Pdexv3PoolPairState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Token0ID              string `json:"Token0ID"`
		Token1ID              string `json:"Token1ID"`
		Token0RealAmount      uint64 `json:"Token0RealAmount"`
		Token1RealAmount      uint64 `json:"Token1RealAmount"`
		CurrentContributionID uint64 `json:"CurrentContributionID"`
		Token0VirtualAmount   uint64 `json:"Token0VirtualAmount"`
		Token1VirtualAmount   uint64 `json:"Token1VirtualAmount"`
		Amplifier             uint   `json:"Amplifier"`
	}{
		Token0ID:              pp.token0ID,
		Token1ID:              pp.token1ID,
		Token0RealAmount:      pp.token0RealAmount,
		Token1RealAmount:      pp.token1RealAmount,
		CurrentContributionID: pp.currentContributionID,
		Token0VirtualAmount:   pp.token0VirtualAmount,
		Token1VirtualAmount:   pp.token1VirtualAmount,
		Amplifier:             pp.amplifier,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (pp *Pdexv3PoolPairState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Token0ID              string `json:"Token0ID"`
		Token1ID              string `json:"Token1ID"`
		Token0RealAmount      uint64 `json:"Token0RealAmount"`
		Token1RealAmount      uint64 `json:"Token1RealAmount"`
		CurrentContributionID uint64 `json:"CurrentContributionID"`
		Token0VirtualAmount   uint64 `json:"Token0VirtualAmount"`
		Token1VirtualAmount   uint64 `json:"Token1VirtualAmount"`
		Amplifier             uint   `json:"Amplifier"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	pp.token0ID = temp.Token0ID
	pp.token1ID = temp.Token1ID
	pp.token0RealAmount = temp.Token0RealAmount
	pp.token1RealAmount = temp.Token1RealAmount
	pp.currentContributionID = temp.CurrentContributionID
	pp.token0VirtualAmount = temp.Token0VirtualAmount
	pp.token1VirtualAmount = temp.Token1VirtualAmount
	pp.amplifier = temp.Amplifier
	return nil
}

func (pp *Pdexv3PoolPairState) Clone() *Pdexv3PoolPairState {
	return NewPdexv3PoolPairStateWithValue(
		pp.token0ID, pp.token1ID,
		pp.token0RealAmount, pp.token1RealAmount, pp.currentContributionID,
		pp.token0VirtualAmount, pp.token1VirtualAmount, pp.amplifier,
	)
}

func NewPdexv3PoolPairState() *Pdexv3PoolPairState {
	return &Pdexv3PoolPairState{}
}

func NewPdexv3PoolPairStateWithValue(
	token0ID, token1ID string,
	token0RealAmount, token1RealAmount, currentContributionID,
	token0VirtualAmount, token1VirtualAmount uint64,
	amplifier uint,
) *Pdexv3PoolPairState {
	return &Pdexv3PoolPairState{
		token0ID:              token0ID,
		token1ID:              token1ID,
		token0RealAmount:      token0RealAmount,
		token1RealAmount:      token1RealAmount,
		currentContributionID: currentContributionID,
		token0VirtualAmount:   token0VirtualAmount,
		token1VirtualAmount:   token1VirtualAmount,
		amplifier:             amplifier,
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
