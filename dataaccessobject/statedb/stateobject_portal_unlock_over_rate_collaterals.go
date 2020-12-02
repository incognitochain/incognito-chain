package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type UnlockOverRateCollaterals struct {
	CustodianAddressStr string
	TokenID             string
	UnlockAmounts       map[string]uint64
}

func (u *UnlockOverRateCollaterals) GetCustodianAddress() string {
	return u.CustodianAddressStr
}

func (u *UnlockOverRateCollaterals) SetCustodianAddress(cusAddress string) {
	u.CustodianAddressStr = cusAddress
}

func (u *UnlockOverRateCollaterals) GetTokenID() string {
	return u.TokenID
}

func (u *UnlockOverRateCollaterals) SetTokenID(tokenID string) {
	u.TokenID = tokenID
}

func (u *UnlockOverRateCollaterals) GetUnlockAmount() map[string]uint64 {
	return u.UnlockAmounts
}

func (u *UnlockOverRateCollaterals) SetUnlockAmount(unlockAmount map[string]uint64) {
	u.UnlockAmounts = unlockAmount
}

func NewUnlockOverRateCollaterals() *UnlockOverRateCollaterals {
	return &UnlockOverRateCollaterals{}
}

func NewUnlockOverRateCollateralsWithValue(
	cusAddress string,
	tokenID string,
	unlockAmounts map[string]uint64,
) *UnlockOverRateCollaterals {
	return &UnlockOverRateCollaterals{
		CustodianAddressStr: cusAddress,
		TokenID:             tokenID,
		UnlockAmounts:       unlockAmounts,
	}
}

func GeneratePortalUnlockOverRateCollateralsStateObjectKey() common.Hash {
	suffix := "unlockoverratecollaterals"
	prefixHash := GetPortalUnlockOverRateCollateralsPrefix()
	valueHash := common.HashH([]byte(suffix))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (u *UnlockOverRateCollaterals) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		CustodianAddressStr string
		TokenID             string
		UnlockAmounts       map[string]uint64
	}{
		CustodianAddressStr: u.CustodianAddressStr,
		TokenID:             u.TokenID,
		UnlockAmounts:       u.UnlockAmounts,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (u *UnlockOverRateCollaterals) UnmarshalJSON(data []byte) error {
	temp := struct {
		CustodianAddressStr string
		TokenID             string
		UnlockAmounts       map[string]uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	u.CustodianAddressStr = temp.CustodianAddressStr
	u.TokenID = temp.TokenID
	u.UnlockAmounts = temp.UnlockAmounts
	return nil
}

type UnlockOverRateCollateralsStateObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                            int
	unlockOverRateCollateralsStateHash common.Hash
	unlockOverRateCollateralsState     *UnlockOverRateCollaterals
	objectType                         int
	deleted                            bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newUnlockOverRateCollateralsStateObjectWithValue(db *StateDB, unlockOverRateCollateralsStateHash common.Hash, data interface{}) (*UnlockOverRateCollateralsStateObject, error) {
	var newUnlockOverRateCollaterals = NewUnlockOverRateCollaterals()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newUnlockOverRateCollaterals)
		if err != nil {
			return nil, err
		}
	} else {
		newUnlockOverRateCollaterals, ok = data.(*UnlockOverRateCollaterals)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidUnlockOverRateCollateralsStateType, reflect.TypeOf(data))
		}
	}
	return &UnlockOverRateCollateralsStateObject{
		db:                                 db,
		version:                            defaultVersion,
		unlockOverRateCollateralsStateHash: unlockOverRateCollateralsStateHash,
		unlockOverRateCollateralsState:     newUnlockOverRateCollaterals,
		objectType:                         PortalUnlockOverRateCollaterals,
		deleted:                            false,
	}, nil
}

func newUnlockOverRateCollateralsStateObject(db *StateDB, unlockOverRateCollateralsStateHash common.Hash) *UnlockOverRateCollateralsStateObject {
	return &UnlockOverRateCollateralsStateObject{
		db:                                 db,
		version:                            defaultVersion,
		unlockOverRateCollateralsStateHash: unlockOverRateCollateralsStateHash,
		unlockOverRateCollateralsState:     NewUnlockOverRateCollaterals(),
		objectType:                         PortalUnlockOverRateCollaterals,
		deleted:                            false,
	}
}

func (u UnlockOverRateCollateralsStateObject) GetVersion() int {
	return u.version
}

// setError remembers the first non-nil error it is called with.
func (u *UnlockOverRateCollateralsStateObject) SetError(err error) {
	if u.dbErr == nil {
		u.dbErr = err
	}
}

func (u UnlockOverRateCollateralsStateObject) GetTrie(db DatabaseAccessWarper) Trie {
	return u.trie
}

func (u *UnlockOverRateCollateralsStateObject) SetValue(data interface{}) error {
	unlockOverRateCollateralsState, ok := data.(*UnlockOverRateCollaterals)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidUnlockOverRateCollateralsStateType, reflect.TypeOf(data))
	}
	u.unlockOverRateCollateralsState = unlockOverRateCollateralsState
	return nil
}

func (u UnlockOverRateCollateralsStateObject) GetValue() interface{} {
	return u.unlockOverRateCollateralsState
}

func (u UnlockOverRateCollateralsStateObject) GetValueBytes() []byte {
	unlockOverRateCollateralsState, ok := u.GetValue().(*UnlockOverRateCollaterals)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(unlockOverRateCollateralsState)
	if err != nil {
		panic("failed to marshal UnlockOverRateCollaterals")
	}
	return value
}

func (u UnlockOverRateCollateralsStateObject) GetHash() common.Hash {
	return u.unlockOverRateCollateralsStateHash
}

func (u UnlockOverRateCollateralsStateObject) GetType() int {
	return u.objectType
}

// MarkDelete will delete an object in trie
func (u *UnlockOverRateCollateralsStateObject) MarkDelete() {
	u.deleted = true
}

// reset all shard committee value into default value
func (u *UnlockOverRateCollateralsStateObject) Reset() bool {
	u.unlockOverRateCollateralsState = NewUnlockOverRateCollaterals()
	return true
}

func (u UnlockOverRateCollateralsStateObject) IsDeleted() bool {
	return u.deleted
}

// value is either default or nil
func (u UnlockOverRateCollateralsStateObject) IsEmpty() bool {
	temp := NewUnlockOverRateCollaterals()
	return reflect.DeepEqual(temp, u.unlockOverRateCollateralsState) || u.unlockOverRateCollateralsState == nil
}
