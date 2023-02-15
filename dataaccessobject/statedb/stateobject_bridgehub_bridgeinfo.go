package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeInfoState struct {
	extChainID    string
	briValidators []string // array of bridgePubKey
	briPubKey     string   // Public key of TSS that used to validate sig from validators by TSS

	// info of previous bridge validators that are used to slashing if they haven't completed their remain tasks
	prevBriValidators []string // array of bridgePubKey
	prevBriPubKey     string   // Public key of TSS that used to validate sig from validators by TSS
}

func (b BridgeInfoState) ExtChainID() string {
	return b.extChainID
}

func (b *BridgeInfoState) SetExtChainID(extChainID string) {
	b.extChainID = extChainID
}

func (b BridgeInfoState) BriValidators() []string {
	return b.briValidators
}

func (b *BridgeInfoState) SetBriValidators(briValidators []string) {
	b.briValidators = briValidators
}

func (b BridgeInfoState) PrevBriValidators() []string {
	return b.prevBriValidators
}

func (b *BridgeInfoState) SetPrevBriValidators(prevBriValidators []string) {
	b.prevBriValidators = prevBriValidators
}
func (b BridgeInfoState) PrevBriPubKey() string {
	return b.prevBriPubKey
}

func (b *BridgeInfoState) SetPrevBriPubKey(prevBriPubKey string) {
	b.prevBriPubKey = prevBriPubKey
}

func (b BridgeInfoState) Clone() *BridgeInfoState {
	briValidatorsCopy := make([]string, len(b.briValidators))
	copy(briValidatorsCopy, b.briValidators)
	prevBriValidatorsCopy := make([]string, len(b.prevBriValidators))
	copy(prevBriValidatorsCopy, b.prevBriValidators)

	return &BridgeInfoState{
		extChainID:        b.extChainID,
		briValidators:     briValidatorsCopy,
		briPubKey:         b.briPubKey,
		prevBriValidators: prevBriValidatorsCopy,
		prevBriPubKey:     b.prevBriPubKey,
	}
}

func (b *BridgeInfoState) IsDiff(compareParam *BridgeInfoState) bool {
	if compareParam == nil {
		return true
	}
	return b.extChainID != compareParam.extChainID ||
		!reflect.DeepEqual(b.briValidators, compareParam.briValidators) ||
		b.briPubKey != compareParam.briPubKey ||
		!reflect.DeepEqual(b.prevBriValidators, compareParam.prevBriValidators) ||
		b.prevBriPubKey != compareParam.prevBriPubKey

}

func (b BridgeInfoState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ExtChainID        string
		BriValidators     []string
		BriPubKey         string
		PrevBriValidators []string
		PrevBriPubKey     string
	}{
		ExtChainID:        b.extChainID,
		BriValidators:     b.briValidators,
		BriPubKey:         b.briPubKey,
		PrevBriValidators: b.prevBriValidators,
		PrevBriPubKey:     b.prevBriPubKey,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeInfoState) UnmarshalJSON(data []byte) error {
	temp := struct {
		ExtChainID        string
		BriValidators     []string
		BriPubKey         string
		PrevBriValidators []string
		PrevBriPubKey     string
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.extChainID = temp.ExtChainID
	b.briValidators = temp.BriValidators
	b.briPubKey = temp.BriPubKey
	b.prevBriPubKey = temp.PrevBriPubKey
	b.prevBriValidators = temp.PrevBriValidators
	return nil
}

func NewBridgeInfoState() *BridgeInfoState {
	return &BridgeInfoState{}
}

func NewBridgeInfoStateWithValue(
	extChainID string,
	briValidators []string,
	briPubKey string,
	prevBriValidators []string,
	prevBriPubKey string,
) *BridgeInfoState {
	return &BridgeInfoState{
		extChainID:        extChainID,
		briValidators:     briValidators,
		briPubKey:         briPubKey,
		prevBriValidators: prevBriValidators,
		prevBriPubKey:     prevBriPubKey,
	}
}

type BridgeHubBridgeInfoObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeInfoState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubBridgeInfoObject(db *StateDB, hash common.Hash) *BridgeHubBridgeInfoObject {
	return &BridgeHubBridgeInfoObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeInfoState(),
		objectType: BridgeHubBridgeInfoObjectType,
		deleted:    false,
	}
}

func newBridgeHubBridgeInfoObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeHubBridgeInfoObject, error) {
	var newBridgeInfo = NewBridgeInfoState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeInfo)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeInfo, ok = data.(*BridgeInfoState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubBridgeInfoStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeHubBridgeInfoObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeInfo,
		db:         db,
		objectType: BridgeHubBridgeInfoObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeHubBridgeInfoObjectKey(bridgeID string) common.Hash {
	prefixHash := GetBridgeHubBridgeInfoPrefix([]byte(bridgeID))
	valueHash := common.HashH([]byte{})
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgeHubBridgeInfoObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeHubBridgeInfoObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeHubBridgeInfoObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeHubBridgeInfoObject) SetValue(data interface{}) error {
	newBridgeHubParam, ok := data.(*BridgeInfoState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubBridgeInfoStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeHubParam
	return nil
}

func (t BridgeHubBridgeInfoObject) GetValue() interface{} {
	return t.state
}

func (t BridgeHubBridgeInfoObject) GetValueBytes() []byte {
	bridgeInfoState, ok := t.GetValue().(*BridgeInfoState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeInfoState)
	if err != nil {
		panic("failed to marshal BridgeInfoState")
	}
	return value
}

func (t BridgeHubBridgeInfoObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeHubBridgeInfoObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeHubBridgeInfoObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeHubBridgeInfoObject) Reset() bool {
	t.state = NewBridgeInfoState()
	return true
}

func (t BridgeHubBridgeInfoObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeHubBridgeInfoObject) IsEmpty() bool {
	return t.state == nil
}
