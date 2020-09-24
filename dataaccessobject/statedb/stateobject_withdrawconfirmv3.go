package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type WithdrawCollateralConfirmStateV3 struct {
	txID   common.Hash
	height uint64
}

func (b WithdrawCollateralConfirmStateV3) Height() uint64 {
	return b.height
}

func (b *WithdrawCollateralConfirmStateV3) SetHeight(height uint64) {
	b.height = height
}

func (b WithdrawCollateralConfirmStateV3) TxID() common.Hash {
	return b.txID
}

func (b *WithdrawCollateralConfirmStateV3) SetTxID(txID common.Hash) {
	b.txID = txID
}

func NewWithdrawCollateralConfirmStateV3() *WithdrawCollateralConfirmStateV3 {
	return &WithdrawCollateralConfirmStateV3{}
}

func NewWithdrawCollateralConfirmStateV3WithValue(txID common.Hash, height uint64) *WithdrawCollateralConfirmStateV3 {
	return &WithdrawCollateralConfirmStateV3{txID: txID, height: height}
}

func (b WithdrawCollateralConfirmStateV3) MarshalJSON() ([]byte, error) {
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

func (b *WithdrawCollateralConfirmStateV3) UnmarshalJSON(data []byte) error {
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

type WithdrawCollateralConfirmStateV3Object struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                  int
	withdrawConfirmStateHash common.Hash
	withdrawConfirmState     *WithdrawCollateralConfirmStateV3
	objectType               int
	deleted                  bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newWithdrawCollateralConfirmStateV3Object(db *StateDB, hash common.Hash) *WithdrawCollateralConfirmStateV3Object {
	return &WithdrawCollateralConfirmStateV3Object{
		version:                  defaultVersion,
		db:                       db,
		withdrawConfirmStateHash: hash,
		withdrawConfirmState:     NewWithdrawCollateralConfirmStateV3(),
		objectType:               WithdrawCollateralConfirmObjectType,
		deleted:                  false,
	}
}
func newWithdrawCollateralConfirmStateV3ObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*WithdrawCollateralConfirmStateV3Object, error) {
	var newBurningConfirmState = NewWithdrawCollateralConfirmStateV3()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBurningConfirmState)
		if err != nil {
			return nil, err
		}
	} else {
		newBurningConfirmState, ok = data.(*WithdrawCollateralConfirmStateV3)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidWithdrawCollateralConfirmStateType, reflect.TypeOf(data))
		}
	}
	return &WithdrawCollateralConfirmStateV3Object{
		version:                  defaultVersion,
		withdrawConfirmStateHash: key,
		withdrawConfirmState:     newBurningConfirmState,
		db:                       db,
		objectType:               WithdrawCollateralConfirmObjectType,
		deleted:                  false,
	}, nil
}

func GenerateWithdrawCollateralConfirmObjectKey(txID common.Hash) common.Hash {
	prefixHash := GetWithdrawCollateralConfirmPrefixV3()
	valueHash := common.HashH(txID[:])
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (ethtx WithdrawCollateralConfirmStateV3Object) GetVersion() int {
	return ethtx.version
}

// setError remembers the first non-nil error it is called with.
func (ethtx *WithdrawCollateralConfirmStateV3Object) SetError(err error) {
	if ethtx.dbErr == nil {
		ethtx.dbErr = err
	}
}

func (ethtx WithdrawCollateralConfirmStateV3Object) GetTrie(db DatabaseAccessWarper) Trie {
	return ethtx.trie
}

func (ethtx *WithdrawCollateralConfirmStateV3Object) SetValue(data interface{}) error {
	var newBurningConfirmState = NewWithdrawCollateralConfirmStateV3()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBurningConfirmState)
		if err != nil {
			return err
		}
	} else {
		newBurningConfirmState, ok = data.(*WithdrawCollateralConfirmStateV3)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidWithdrawCollateralConfirmStateType, reflect.TypeOf(data))
		}
	}
	ethtx.withdrawConfirmState = newBurningConfirmState
	return nil
}

func (ethtx WithdrawCollateralConfirmStateV3Object) GetValue() interface{} {
	return ethtx.withdrawConfirmState
}

func (ethtx WithdrawCollateralConfirmStateV3Object) GetValueBytes() []byte {
	data := ethtx.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal burning confirm state")
	}
	return []byte(value)
}

func (ethtx WithdrawCollateralConfirmStateV3Object) GetHash() common.Hash {
	return ethtx.withdrawConfirmStateHash
}

func (ethtx WithdrawCollateralConfirmStateV3Object) GetType() int {
	return ethtx.objectType
}

// MarkDelete will delete an object in trie
func (ethtx *WithdrawCollateralConfirmStateV3Object) MarkDelete() {
	ethtx.deleted = true
}

func (ethtx *WithdrawCollateralConfirmStateV3Object) Reset() bool {
	ethtx.withdrawConfirmState = NewWithdrawCollateralConfirmStateV3()
	return true
}

func (ethtx WithdrawCollateralConfirmStateV3Object) IsDeleted() bool {
	return ethtx.deleted
}

// value is either default or nil
func (ethtx WithdrawCollateralConfirmStateV3Object) IsEmpty() bool {
	temp := NewWithdrawCollateralConfirmStateV3()
	return reflect.DeepEqual(temp, ethtx.withdrawConfirmState) || ethtx.withdrawConfirmState == nil
}
