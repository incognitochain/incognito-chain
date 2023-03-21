package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeHubPTokenkState struct {
	pTokenID     common.Hash
	pTokenAmount uint64 // pTokenID : amount
}

func (b BridgeHubPTokenkState) PTokenAmount() uint64 {
	return b.pTokenAmount
}

func (b *BridgeHubPTokenkState) SetPTokenAmount(pTokenAmount uint64) {
	b.pTokenAmount = pTokenAmount
}

func (b BridgeHubPTokenkState) PTokenID() common.Hash {
	return b.pTokenID
}

func (b *BridgeHubPTokenkState) SetPTokenID(pTokenID common.Hash) {
	b.pTokenID = pTokenID
}

func (b BridgeHubPTokenkState) Clone() *BridgeHubPTokenkState {
	return &BridgeHubPTokenkState{
		pTokenID:     b.pTokenID,
		pTokenAmount: b.pTokenAmount,
	}
}

func (b *BridgeHubPTokenkState) IsDiff(compareParam *BridgeHubPTokenkState) bool {
	if compareParam == nil {
		return true
	}
	return b.pTokenAmount != compareParam.pTokenAmount || b.pTokenID != compareParam.pTokenID
}

func (b BridgeHubPTokenkState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		PTokenID     common.Hash
		PTokenAmount uint64
	}{
		PTokenID:     b.pTokenID,
		PTokenAmount: b.pTokenAmount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeHubPTokenkState) UnmarshalJSON(data []byte) error {
	temp := struct {
		PTokenID     common.Hash
		PTokenAmount uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.pTokenID = temp.PTokenID
	b.pTokenAmount = temp.PTokenAmount
	return nil
}

func NewBridgeHubPTokenkState() *BridgeHubPTokenkState {
	return &BridgeHubPTokenkState{}
}

func NewBridgeHubPTokenkStateWithValue(pTokenAmount uint64, pTokenID common.Hash) *BridgeHubPTokenkState {
	return &BridgeHubPTokenkState{
		pTokenID:     pTokenID,
		pTokenAmount: pTokenAmount,
	}
}

type BridgeHubNetworkObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeHubPTokenkState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubNetworkObject(db *StateDB, hash common.Hash) *BridgeHubNetworkObject {
	return &BridgeHubNetworkObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeHubPTokenkState(),
		objectType: BridgeHubBridgeInfoNetworkObjectType,
		deleted:    false,
	}
}

func newBridgeHubNetworkObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeHubNetworkObject, error) {
	var newBridgePToken = NewBridgeHubPTokenkState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgePToken)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgePToken, ok = data.(*BridgeHubPTokenkState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubPTokenStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeHubNetworkObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgePToken,
		db:         db,
		objectType: BridgeHubBridgeInfoNetworkObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeHubPTokenObjectKey(bridgeID string, networkId int, pTokenId common.Hash) common.Hash {
	prefixHash := GetBridgeHubPTokenPrefix([]byte(bridgeID))
	valueHash := common.HashH(common.IntToBytes(networkId))
	pTokenHash := common.HashH(pTokenId.Bytes())
	return common.BytesToHash(append(prefixHash, append(valueHash[:][:prefixKeyLength/2], pTokenHash[:][:prefixKeyLength/2]...)...))
}

func (t BridgeHubNetworkObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeHubNetworkObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeHubNetworkObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeHubNetworkObject) SetValue(data interface{}) error {
	newBridgeHubPToken, ok := data.(*BridgeHubPTokenkState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubPTokenStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeHubPToken
	return nil
}

func (t BridgeHubNetworkObject) GetValue() interface{} {
	return t.state
}

func (t BridgeHubNetworkObject) GetValueBytes() []byte {
	bridgeHubPTokenState, ok := t.GetValue().(*BridgeHubPTokenkState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeHubPTokenState)
	if err != nil {
		panic("failed to marshal BridgeHubPTokenState")
	}
	return value
}

func (t BridgeHubNetworkObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeHubNetworkObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeHubNetworkObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeHubNetworkObject) Reset() bool {
	t.state = NewBridgeHubPTokenkState()
	return true
}

func (t BridgeHubNetworkObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeHubNetworkObject) IsEmpty() bool {
	return t.state == nil
}
