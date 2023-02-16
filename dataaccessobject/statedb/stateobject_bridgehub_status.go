package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeHubStatusState struct {
	statusType    []byte
	statusSuffix  []byte
	statusContent []byte
}

func (b BridgeHubStatusState) StatusType() []byte {
	return b.statusType
}

func (b *BridgeHubStatusState) SetStatusType(statusType []byte) {
	b.statusType = statusType
}

func (b BridgeHubStatusState) StatusSuffix() []byte {
	return b.statusSuffix
}

func (b *BridgeHubStatusState) SetStatusSuffix(statusSuffix []byte) {
	b.statusSuffix = statusSuffix
}

func (b BridgeHubStatusState) StatusContent() []byte {
	return b.statusContent
}

func (b *BridgeHubStatusState) SetStatusContent(statusContent []byte) {
	b.statusContent = statusContent
}

func (b BridgeHubStatusState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		StatusType    []byte
		StatusSuffix  []byte
		StatusContent []byte
	}{
		StatusType:    b.statusType,
		StatusSuffix:  b.statusSuffix,
		StatusContent: b.statusContent,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *BridgeHubStatusState) UnmarshalJSON(data []byte) error {
	temp := struct {
		StatusType    []byte
		StatusSuffix  []byte
		StatusContent []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.statusType = temp.StatusType
	b.statusContent = temp.StatusContent
	b.statusSuffix = temp.StatusSuffix
	return nil
}

func NewBridgeHubStatusState() *BridgeHubStatusState {
	return &BridgeHubStatusState{}
}

func NewBridgeHubStatusStateWithValue(statusType []byte, statusSuffix []byte, statusContent []byte) *BridgeHubStatusState {
	return &BridgeHubStatusState{statusType: statusType, statusSuffix: statusSuffix, statusContent: statusContent}
}

type BridgeHubStatusObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeHubStatusState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeHubStatusObject(db *StateDB, hash common.Hash) *BridgeHubStatusObject {
	return &BridgeHubStatusObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeHubStatusState(),
		objectType: BridgeHubStatusObjectType,
		deleted:    false,
	}
}

func newBridgeHubStatusObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeHubStatusObject, error) {
	var newBridgeHubStatus = NewBridgeHubStatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeHubStatus)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeHubStatus, ok = data.(*BridgeHubStatusState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubStatusStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeHubStatusObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeHubStatus,
		db:         db,
		objectType: BridgeHubStatusObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeHubStatusObjectKey(statusType []byte, statusSuffix []byte) common.Hash {
	prefixHash := GetBridgeHubStatusPrefix(statusType)
	valueHash := common.HashH(append(statusType, statusSuffix...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgeHubStatusObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeHubStatusObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeHubStatusObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeHubStatusObject) SetValue(data interface{}) error {
	newBridgeHubStatus, ok := data.(*BridgeHubStatusState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeHubStatusStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeHubStatus
	return nil
}

func (t BridgeHubStatusObject) GetValue() interface{} {
	return t.state
}

func (t BridgeHubStatusObject) GetValueBytes() []byte {
	bridgeHubState, ok := t.GetValue().(*BridgeHubStatusState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeHubState)
	if err != nil {
		panic("failed to marshal BridgeHubStatusState")
	}
	return value
}

func (t BridgeHubStatusObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeHubStatusObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeHubStatusObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeHubStatusObject) Reset() bool {
	t.state = NewBridgeHubStatusState()
	return true
}

func (t BridgeHubStatusObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeHubStatusObject) IsEmpty() bool {
	temp := NewBridgeHubStatusState()
	return reflect.DeepEqual(temp, t.state) || t.state == nil
}
