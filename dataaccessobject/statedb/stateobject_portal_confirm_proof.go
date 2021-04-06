package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type PortalConfirmProofState struct {
	txID   common.Hash
	height uint64
}

func (b PortalConfirmProofState) Height() uint64 {
	return b.height
}

func (b *PortalConfirmProofState) SetHeight(height uint64) {
	b.height = height
}

func (b PortalConfirmProofState) TxID() common.Hash {
	return b.txID
}

func (b *PortalConfirmProofState) SetTxID(txID common.Hash) {
	b.txID = txID
}

func NewPortalConfirmProofState() *PortalConfirmProofState {
	return &PortalConfirmProofState{}
}

func NewPortalConfirmProofStateWithValue(txID common.Hash, height uint64) *PortalConfirmProofState {
	return &PortalConfirmProofState{txID: txID, height: height}
}

func (b PortalConfirmProofState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		TxID   common.Hash
		Height uint64
	}{
		TxID:   b.txID,
		Height: b.height,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (b *PortalConfirmProofState) UnmarshalJSON(data []byte) error {
	temp := struct {
		TxID   common.Hash
		Height uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	b.txID = temp.TxID
	b.height = temp.Height
	return nil
}

type NewPortalConfirmProofStateObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                     int
	portalConfirmProofStateHash common.Hash
	portalConfirmProofState     *PortalConfirmProofState
	objectType                  int
	deleted                     bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPortalConfirmProofStateObject(db *StateDB, hash common.Hash) *NewPortalConfirmProofStateObject {
	return &NewPortalConfirmProofStateObject{
		version:                     defaultVersion,
		db:                          db,
		portalConfirmProofStateHash: hash,
		portalConfirmProofState:     NewPortalConfirmProofState(),
		objectType:                  PortalConfirmProofObjectType,
		deleted:                     false,
	}
}
func newPortalConfirmProofStateObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*NewPortalConfirmProofStateObject, error) {
	var portalConfirmProofState = NewPortalConfirmProofState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, portalConfirmProofState)
		if err != nil {
			return nil, err
		}
	} else {
		portalConfirmProofState, ok = data.(*PortalConfirmProofState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalConfirmProofStateType, reflect.TypeOf(data))
		}
	}
	return &NewPortalConfirmProofStateObject{
		version:                     defaultVersion,
		portalConfirmProofStateHash: key,
		portalConfirmProofState:     portalConfirmProofState,
		db:                          db,
		objectType:                  PortalConfirmProofObjectType,
		deleted:                     false,
	}, nil
}

func GeneratePortalConfirmProofObjectKey(proofType []byte, txID common.Hash) common.Hash {
	prefixHash := GetPortalConfirmProofPrefixV3(proofType)
	valueHash := common.HashH(txID[:])
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ethtx NewPortalConfirmProofStateObject) GetVersion() int {
	return ethtx.version
}

// setError remembers the first non-nil error it is called with.
func (ethtx *NewPortalConfirmProofStateObject) SetError(err error) {
	if ethtx.dbErr == nil {
		ethtx.dbErr = err
	}
}

func (ethtx NewPortalConfirmProofStateObject) GetTrie(db DatabaseAccessWarper) Trie {
	return ethtx.trie
}

func (ethtx *NewPortalConfirmProofStateObject) SetValue(data interface{}) error {
	var newBurningConfirmState = NewPortalConfirmProofState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBurningConfirmState)
		if err != nil {
			return err
		}
	} else {
		newBurningConfirmState, ok = data.(*PortalConfirmProofState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalConfirmProofStateType, reflect.TypeOf(data))
		}
	}
	ethtx.portalConfirmProofState = newBurningConfirmState
	return nil
}

func (ethtx NewPortalConfirmProofStateObject) GetValue() interface{} {
	return ethtx.portalConfirmProofState
}

func (ethtx NewPortalConfirmProofStateObject) GetValueBytes() []byte {
	data := ethtx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal burning confirm state")
	}
	return []byte(value)
}

func (ethtx NewPortalConfirmProofStateObject) GetHash() common.Hash {
	return ethtx.portalConfirmProofStateHash
}

func (ethtx NewPortalConfirmProofStateObject) GetType() int {
	return ethtx.objectType
}

// MarkDelete will delete an object in trie
func (ethtx *NewPortalConfirmProofStateObject) MarkDelete() {
	ethtx.deleted = true
}

func (ethtx *NewPortalConfirmProofStateObject) Reset() bool {
	ethtx.portalConfirmProofState = NewPortalConfirmProofState()
	return true
}

func (ethtx NewPortalConfirmProofStateObject) IsDeleted() bool {
	return ethtx.deleted
}

// value is either default or nil
func (ethtx NewPortalConfirmProofStateObject) IsEmpty() bool {
	temp := NewPortalConfirmProofState()
	return reflect.DeepEqual(temp, ethtx.portalConfirmProofState) || ethtx.portalConfirmProofState == nil
}
