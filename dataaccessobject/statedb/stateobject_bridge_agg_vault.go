package statedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggVaultState struct {
	amount     uint64
	extDecimal uint
	networkID  uint
	tokenID    common.Hash
}

func (state *BridgeAggVaultState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount     uint64      `json:"Amount"`
		ExtDecimal uint        `json:"ExtDecimal"`
		NetworkID  uint        `json:"NetworkID"`
		TokenID    common.Hash `json:"TokenID"`
	}{
		Amount:     state.amount,
		ExtDecimal: state.extDecimal,
		NetworkID:  state.networkID,
		TokenID:    state.tokenID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *BridgeAggVaultState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount    uint64      `json:"Amount"`
		Decimal   uint        `json:"ExtDecimal"`
		NetworkID uint        `json:"NetworkID"`
		TokenID   common.Hash `json:"TokenID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.amount = temp.Amount
	state.extDecimal = temp.Decimal
	state.networkID = temp.NetworkID
	state.tokenID = temp.TokenID
	return nil
}

func NewBridgeAggVaultState() *BridgeAggVaultState {
	return &BridgeAggVaultState{}
}

func NewBridgeAggVaultStateWithValue(
	amount, lastUpdatedRewardReserve, currentRewardReserve uint64, extDecimal uint, isPaused bool, networkID uint, tokenID common.Hash,
) *BridgeAggVaultState {
	return &BridgeAggVaultState{
		amount:     amount,
		extDecimal: extDecimal,
		networkID:  networkID,
		tokenID:    tokenID,
	}
}

func (b *BridgeAggVaultState) ExtDecimal() uint {
	return b.extDecimal
}

func (b *BridgeAggVaultState) Amount() uint64 {
	return b.amount
}

func (b *BridgeAggVaultState) NetworkID() uint {
	return b.networkID
}

func (b *BridgeAggVaultState) TokenID() common.Hash {
	return b.tokenID
}

func (b *BridgeAggVaultState) SetAmount(amount uint64) {
	b.amount = amount
}

func (b *BridgeAggVaultState) SetExtDecimal(extDecimal uint) {
	b.extDecimal = extDecimal
}

func (b *BridgeAggVaultState) Clone() *BridgeAggVaultState {
	return &BridgeAggVaultState{
		amount:     b.amount,
		extDecimal: b.extDecimal,
		networkID:  b.networkID,
		tokenID:    b.tokenID,
	}
}

func (b *BridgeAggVaultState) GetDiff(compareState *BridgeAggVaultState) (*BridgeAggVaultState, error) {
	if compareState == nil {
		return nil, errors.New("compareState is nil")
	}
	if b.amount != compareState.amount ||
		b.extDecimal != compareState.extDecimal ||
		b.networkID != compareState.networkID || b.tokenID != compareState.tokenID {
		return b.Clone(), nil
	}
	return nil, nil
}

func (b *BridgeAggVaultState) IsEmpty() bool {
	return b.amount == 0 && b.extDecimal == 0 && b.tokenID == common.Hash{}
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
