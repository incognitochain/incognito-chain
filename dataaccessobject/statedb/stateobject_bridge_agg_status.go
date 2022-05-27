package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggStatusState struct {
	statusType    []byte
	statusSuffix  []byte
	statusContent []byte
}

func (b BridgeAggStatusState) StatusType() []byte {
	return b.statusType
}

func (b *BridgeAggStatusState) SetStatusType(statusType []byte) {
	b.statusType = statusType
}

func (b BridgeAggStatusState) StatusSuffix() []byte {
	return b.statusSuffix
}

func (b *BridgeAggStatusState) SetStatusSuffix(statusSuffix []byte) {
	b.statusSuffix = statusSuffix
}

func (b BridgeAggStatusState) StatusContent() []byte {
	return b.statusContent
}

func (b *BridgeAggStatusState) SetStatusContent(statusContent []byte) {
	b.statusContent = statusContent
}

func (b BridgeAggStatusState) MarshalJSON() ([]byte, error) {
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

func (b *BridgeAggStatusState) UnmarshalJSON(data []byte) error {
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

func NewBridgeAggStatusState() *BridgeAggStatusState {
	return &BridgeAggStatusState{}
}

func NewBridgeAggStatusStateWithValue(statusType []byte, statusSuffix []byte, statusContent []byte) *BridgeAggStatusState {
	return &BridgeAggStatusState{statusType: statusType, statusSuffix: statusSuffix, statusContent: statusContent}
}

type BridgeAggStatusObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeAggStatusState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAggStatusObject(db *StateDB, hash common.Hash) *BridgeAggStatusObject {
	return &BridgeAggStatusObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeAggStatusState(),
		objectType: BridgeAggStatusObjectType,
		deleted:    false,
	}
}

func newBridgeAggStatusObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*BridgeAggStatusObject, error) {
	var newBridgeAggStatus = NewBridgeAggStatusState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAggStatus)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeAggStatus, ok = data.(*BridgeAggStatusState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggStatusStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeAggStatusObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeAggStatus,
		db:         db,
		objectType: BridgeAggStatusObjectType,
		deleted:    false,
	}, nil
}

func GenerateBridgeAggStatusObjectKey(statusType []byte, statusSuffix []byte) common.Hash {
	prefixHash := GetBridgeAggStatusPrefix(statusType)
	valueHash := common.HashH(append(statusType, statusSuffix...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t BridgeAggStatusObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *BridgeAggStatusObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t BridgeAggStatusObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *BridgeAggStatusObject) SetValue(data interface{}) error {
	newBridgeAggStatus, ok := data.(*BridgeAggStatusState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggStatusStateType, reflect.TypeOf(data))
	}
	t.state = newBridgeAggStatus
	return nil
}

func (t BridgeAggStatusObject) GetValue() interface{} {
	return t.state
}

func (t BridgeAggStatusObject) GetValueBytes() []byte {
	bridgeAggState, ok := t.GetValue().(*BridgeAggStatusState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(bridgeAggState)
	if err != nil {
		panic("failed to marshal BridgeAggStatusState")
	}
	return value
}

func (t BridgeAggStatusObject) GetHash() common.Hash {
	return t.hash
}

func (t BridgeAggStatusObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *BridgeAggStatusObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *BridgeAggStatusObject) Reset() bool {
	t.state = NewBridgeAggStatusState()
	return true
}

func (t BridgeAggStatusObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t BridgeAggStatusObject) IsEmpty() bool {
	temp := NewBridgeAggStatusState()
	return reflect.DeepEqual(temp, t.state) || t.state == nil
}
