package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type PortalV4ShieldInfoState struct {
	depositPubKey []byte
	txIDs         []string
}

func (s PortalV4ShieldInfoState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		DepositPubKey []byte
		TxIDs         []string
	}{
		DepositPubKey: s.depositPubKey,
		TxIDs:         s.txIDs,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *PortalV4ShieldInfoState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TokenID       string
		DepositPubKey []byte
		TxIDs         []string
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.depositPubKey = temp.DepositPubKey
	s.txIDs = temp.TxIDs
	return nil
}

func (s *PortalV4ShieldInfoState) ToString() string {
	pubKeyStr := base58.Base58Check{}.Encode(s.depositPubKey, 0)

	return fmt.Sprintf("DepositPubKey: %v\nTxIDs: %v", pubKeyStr, s.txIDs)
}

func NewPortalV4ShieldInfoState() *PortalV4ShieldInfoState {
	return &PortalV4ShieldInfoState{}
}

func NewPortalV4ShieldInfoStateWithValue(depositPubKey []byte, txIDs []string) *PortalV4ShieldInfoState {
	return &PortalV4ShieldInfoState{depositPubKey: depositPubKey, txIDs: txIDs}
}

type PortalV4ShieldInfoObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                int
	portalV4ShieldInfoHash common.Hash
	portalV4ShieldInfo     *PortalV4ShieldInfoState
	objectType             int
	deleted                bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPortalV4ShieldInfoObject(db *StateDB, hash common.Hash) *PortalV4ShieldInfoObject {
	return &PortalV4ShieldInfoObject{
		version:                defaultVersion,
		db:                     db,
		portalV4ShieldInfoHash: hash,
		portalV4ShieldInfo:     NewPortalV4ShieldInfoState(),
		objectType:             PortalV4ShieldInfoObjectType,
		deleted:                false,
	}
}

func newPortalV4ShieldInfoObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PortalV4ShieldInfoObject, error) {
	var newPortalStatus = NewPortalV4ShieldInfoState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPortalStatus)
		if err != nil {
			return nil, err
		}
	} else {
		newPortalStatus, ok = data.(*PortalV4ShieldInfoState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalV4ShieldInfoStateType, reflect.TypeOf(data))
		}
	}
	return &PortalV4ShieldInfoObject{
		version:                defaultVersion,
		portalV4ShieldInfoHash: key,
		portalV4ShieldInfo:     newPortalStatus,
		db:                     db,
		objectType:             PortalV4ShieldInfoObjectType,
		deleted:                false,
	}, nil
}

func GeneratePortalV4ShieldInfoObjectKey(depositPubKey []byte) common.Hash {
	prefixHash := GetPortalV4ShieldInfoPrefix()
	valueHash := common.HashH(depositPubKey)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PortalV4ShieldInfoObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PortalV4ShieldInfoObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PortalV4ShieldInfoObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PortalV4ShieldInfoObject) SetValue(data interface{}) error {
	newPortalStatus, ok := data.(*PortalV4ShieldInfoState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalV4ShieldInfoStateType, reflect.TypeOf(data))
	}
	t.portalV4ShieldInfo = newPortalStatus
	return nil
}

func (t PortalV4ShieldInfoObject) GetValue() interface{} {
	return t.portalV4ShieldInfo
}

func (t PortalV4ShieldInfoObject) GetValueBytes() []byte {
	PortalV4ShieldInfoState, ok := t.GetValue().(*PortalV4ShieldInfoState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(PortalV4ShieldInfoState)
	if err != nil {
		panic("failed to marshal PortalV4ShieldInfoState")
	}
	return value
}

func (t PortalV4ShieldInfoObject) GetHash() common.Hash {
	return t.portalV4ShieldInfoHash
}

func (t PortalV4ShieldInfoObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PortalV4ShieldInfoObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PortalV4ShieldInfoObject) Reset() bool {
	t.portalV4ShieldInfo = NewPortalV4ShieldInfoState()
	return true
}

func (t PortalV4ShieldInfoObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PortalV4ShieldInfoObject) IsEmpty() bool {
	temp := NewPortalV4ShieldInfoState()
	return reflect.DeepEqual(temp, t.portalV4ShieldInfo) || t.portalV4ShieldInfo == nil
}
