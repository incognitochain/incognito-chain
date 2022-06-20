package statedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggVaultState struct {
	reserve                  uint64
	lastUpdatedRewardReserve uint64
	currentRewardReserve     uint64
	decimal                  uint
	isPaused                 bool
}

func (state *BridgeAggVaultState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Reserve                  uint64 `json:"Reserve"`
		LastUpdatedRewardReserve uint64 `json:"LastUpdatedRewardReserve"`
		CurrentRewardReserve     uint64 `json:"CurrentRewardReserve"`
		Decimal                  uint   `json:"Decimal"`
		IsPaused                 bool   `json:"IsPaused"`
	}{
		Reserve:                  state.reserve,
		LastUpdatedRewardReserve: state.lastUpdatedRewardReserve,
		CurrentRewardReserve:     state.currentRewardReserve,
		Decimal:                  state.decimal,
		IsPaused:                 state.isPaused,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *BridgeAggVaultState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Reserve                  uint64 `json:"Reserve"`
		LastUpdatedRewardReserve uint64 `json:"LastUpdatedRewardReserve"`
		CurrentRewardReserve     uint64 `json:"CurrentRewardReserve"`
		Decimal                  uint   `json:"Decimal"`
		IsPaused                 bool   `json:"IsPaused"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.reserve = temp.Reserve
	state.lastUpdatedRewardReserve = temp.LastUpdatedRewardReserve
	state.currentRewardReserve = temp.CurrentRewardReserve
	state.decimal = temp.Decimal
	state.isPaused = temp.IsPaused
	return nil
}

func NewBridgeAggVaultState() *BridgeAggVaultState {
	return &BridgeAggVaultState{}
}

func NewBridgeAggVaultStateWithValue(
	reserve, lastUpdatedRewardReserve, currentRewardReserve uint64, decimal uint, isPaused bool,
) *BridgeAggVaultState {
	return &BridgeAggVaultState{
		reserve:                  reserve,
		lastUpdatedRewardReserve: lastUpdatedRewardReserve,
		currentRewardReserve:     currentRewardReserve,
		decimal:                  decimal,
		isPaused:                 isPaused,
	}
}

func (b *BridgeAggVaultState) Decimal() uint {
	return b.decimal
}

func (b *BridgeAggVaultState) Reserve() uint64 {
	return b.reserve
}

func (b *BridgeAggVaultState) LastUpdatedRewardReserve() uint64 {
	return b.lastUpdatedRewardReserve
}

func (b *BridgeAggVaultState) CurrentRewardReserve() uint64 {
	return b.currentRewardReserve
}

func (b *BridgeAggVaultState) IsPaused() bool {
	return b.isPaused
}

func (b *BridgeAggVaultState) SetReserve(reserve uint64) {
	b.reserve = reserve
}

func (b *BridgeAggVaultState) SetLastUpdatedRewardReserve(lastUpdatedRewardReserve uint64) {
	b.lastUpdatedRewardReserve = lastUpdatedRewardReserve
}

func (b *BridgeAggVaultState) SetCurrentRewardReserve(currentRewardReserve uint64) {
	b.currentRewardReserve = currentRewardReserve
}

func (b *BridgeAggVaultState) SetDecimal(decimal uint) {
	b.decimal = decimal
}

func (b *BridgeAggVaultState) SetIsPaused(isPaused bool) {
	b.isPaused = isPaused
}

func (b *BridgeAggVaultState) Clone() *BridgeAggVaultState {
	return &BridgeAggVaultState{
		reserve:                  b.reserve,
		lastUpdatedRewardReserve: b.lastUpdatedRewardReserve,
		currentRewardReserve:     b.currentRewardReserve,
		decimal:                  b.decimal,
		isPaused:                 b.isPaused,
	}
}

func (b *BridgeAggVaultState) GetDiff(compareState *BridgeAggVaultState) (*BridgeAggVaultState, error) {
	if compareState == nil {
		return nil, errors.New("compareState is nil")
	}
	if b.reserve != compareState.reserve ||
		b.currentRewardReserve != compareState.currentRewardReserve ||
		b.lastUpdatedRewardReserve != compareState.lastUpdatedRewardReserve ||
		b.decimal != compareState.decimal || b.isPaused != compareState.isPaused {
		return b.Clone(), nil
	}
	return nil, nil
}

func (b *BridgeAggVaultState) IsEmpty() bool {
	return b.reserve == 0 && b.currentRewardReserve == 0 && b.lastUpdatedRewardReserve == 0 && b.decimal == 0
}

type BridgeAggVaulltObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeAggVaultState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAggVaultObject(db *StateDB, hash common.Hash) *BridgeAggVaulltObject {
	return &BridgeAggVaulltObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeAggVaultState(),
		objectType: BridgeAggVaultObjectType,
		deleted:    false,
	}
}

func newBridgeAggVaultObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*BridgeAggVaulltObject, error,
) {
	var newBridgeAggVaultState = NewBridgeAggVaultState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAggVaultState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeAggVaultState, ok = data.(*BridgeAggVaultState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggVaultStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeAggVaulltObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeAggVaultState,
		db:         db,
		objectType: BridgeAggVaultObjectType,
		deleted:    false,
	}, nil
}

func generateBridgeAggVaultObjectPrefix(unifiedTokenID common.Hash) []byte {
	b := append(GetBridgeAggVaultPrefix(), unifiedTokenID.Bytes()...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GenerateBridgeAggVaultObjectKey(unifiedTokenID, tokenID common.Hash) common.Hash {
	prefixHash := generateBridgeAggVaultObjectPrefix(unifiedTokenID)
	valueHash := common.HashH(tokenID.Bytes())
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *BridgeAggVaulltObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *BridgeAggVaulltObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *BridgeAggVaulltObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *BridgeAggVaulltObject) SetValue(data interface{}) error {
	newBridgeAggVaultState, ok := data.(*BridgeAggVaultState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggVaultStateType, reflect.TypeOf(data))
	}
	object.state = newBridgeAggVaultState
	return nil
}

func (object *BridgeAggVaulltObject) GetValue() interface{} {
	return object.state
}

func (object *BridgeAggVaulltObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*BridgeAggVaultState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal bridge agg vault state")
	}
	return value
}

func (object *BridgeAggVaulltObject) GetHash() common.Hash {
	return object.hash
}

func (object *BridgeAggVaulltObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *BridgeAggVaulltObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *BridgeAggVaulltObject) Reset() bool {
	object.state = NewBridgeAggVaultState()
	return true
}

func (object *BridgeAggVaulltObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *BridgeAggVaulltObject) IsEmpty() bool {
	temp := NewBridgeAggVaultState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
