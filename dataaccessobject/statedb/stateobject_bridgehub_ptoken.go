package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeHubPTokenState struct {
	pTokenAmount uint64 // pTokenID : amount
}

func (b BridgeHubPTokenState) PTokenAmount() uint64 {
	return b.pTokenAmount
}

func (b *BridgeHubPTokenState) SetPTokenAmount(pTokenAmount uint64) {
	b.pTokenAmount = pTokenAmount
}

func (b BridgeHubPTokenState) Clone() *BridgeHubPTokenState {
	return &BridgeHubPTokenState{
		pTokenAmount: b.pTokenAmount,
	}
}

func (b *BridgeHubPTokenState) IsDiff(compareParam *BridgeHubPTokenState) bool {
	if compareParam == nil {
		return true
	}
	return b.pTokenAmount != compareParam.pTokenAmount
}

func (b BridgeHubPTokenState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PTokenAmount uint64
	}{
		PTokenAmount: b.pTokenAmount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeHubPTokenState) UnmarshalJSON(data []byte) error {
	temp := struct {
		PTokenAmount uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.pTokenAmount = temp.PTokenAmount
	return nil
}

func NewBridgeHubPTokenState() *BridgeHubPTokenState {
	return &BridgeHubPTokenState{}
}

func NewBridgeHubPTokenStateWithValue(pTokenAmount uint64) *BridgeHubPTokenState {
	return &BridgeHubPTokenState{
		pTokenAmount: pTokenAmount,
	}
}

type BridgePTokenObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeHubPTokenState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubPTokenObject(db *StateDB, hash common.Hash) *BridgePTokenObject {
	return &BridgePTokenObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeHubPTokenState(),
		objectType: BridgeHubPTokenObjectType,
		deleted:    false,
	}
}

func newBridgeHubPTokenObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgePTokenObject, error) {
	var newBridgePToken = NewBridgeHubPTokenState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgePToken)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgePToken, ok = data.(*BridgeHubPTokenState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubPTokenStateType, reflect.TypeOf(data))
		}
	}
	return &BridgePTokenObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgePToken,
		db:         db,
		objectType: BridgeHubPTokenObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgePTokenObjectKey(bridgeID, pTokenID string) common.Hash {
	prefixHash := GetBridgeHubPTokenPrefix([]byte(bridgeID))
	valueHash := common.HashH([]byte(pTokenID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgePTokenObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgePTokenObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgePTokenObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgePTokenObject) SetValue(data interface{}) error {
	newBridgeHubPToken, ok := data.(*BridgeHubPTokenState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubPTokenStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeHubPToken
	return nil
}

func (t BridgePTokenObject) GetValue() interface{} {
	return t.state
}

func (t BridgePTokenObject) GetValueBytes() []byte {
	bridgeHubPTokenState, ok := t.GetValue().(*BridgeHubPTokenState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeHubPTokenState)
	if err != nil {
		panic("failed to marshal BridgeHubPTokenState")
	}
	return value
}

func (t BridgePTokenObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgePTokenObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgePTokenObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgePTokenObject) Reset() bool {
	t.state = NewBridgeHubPTokenState()
	return true
}

func (t BridgePTokenObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgePTokenObject) IsEmpty() bool {
	return t.state == nil
}
