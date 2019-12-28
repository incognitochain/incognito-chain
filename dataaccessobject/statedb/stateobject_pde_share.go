package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type PDEShareState struct {
	beaconHeight       uint64
	token1ID           string
	token2ID           string
	contributorAddress string
	amount             uint64
}

func (s PDEShareState) BeaconHeight() uint64 {
	return s.beaconHeight
}

func (s *PDEShareState) SetBeaconHeight(beaconHeight uint64) {
	s.beaconHeight = beaconHeight
}

func (s PDEShareState) Token1ID() string {
	return s.token1ID
}

func (s *PDEShareState) SetToken1ID(token1ID string) {
	s.token1ID = token1ID
}

func (s PDEShareState) Token2ID() string {
	return s.token2ID
}

func (s *PDEShareState) SetToken2ID(token2ID string) {
	s.token2ID = token2ID
}

func (s PDEShareState) Amount() uint64 {
	return s.amount
}

func (s *PDEShareState) SetAmount(amount uint64) {
	s.amount = amount
}

func (s PDEShareState) ContributorAddress() string {
	return s.contributorAddress
}

func (s *PDEShareState) SetContributorAddress(contributorAddress string) {
	s.contributorAddress = contributorAddress
}

func (s PDEShareState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		BeaconHeight       uint64
		Token1ID           string
		Token2ID           string
		ContributorAddress string
		Amount             uint64
	}{
		BeaconHeight:       s.beaconHeight,
		Token1ID:           s.token1ID,
		Token2ID:           s.token2ID,
		ContributorAddress: s.contributorAddress,
		Amount:             s.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *PDEShareState) UnmarshalJSON(data []byte) error {
	temp := struct {
		BeaconHeight       uint64
		Token1ID           string
		Token2ID           string
		ContributorAddress string
		Amount             uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.beaconHeight = temp.BeaconHeight
	s.token1ID = temp.Token1ID
	s.token2ID = temp.Token2ID
	s.contributorAddress = temp.ContributorAddress
	s.amount = temp.Amount
	return nil
}

func NewPDEShareState() *PDEShareState {
	return &PDEShareState{}
}

func NewPDEShareStateWithValue(beaconHeight uint64, token1ID string, token2ID string, contributorAddress string, amount uint64) *PDEShareState {
	return &PDEShareState{beaconHeight: beaconHeight, token1ID: token1ID, token2ID: token2ID, contributorAddress: contributorAddress, amount: amount}
}

type PDEShareObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version       int
	pdeShareHash  common.Hash
	pdeShareState *PDEShareState
	objectType    int
	deleted       bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newPDEShareObject(db *StateDB, hash common.Hash) *PDEShareObject {
	return &PDEShareObject{
		version:       defaultVersion,
		db:            db,
		pdeShareHash:  hash,
		pdeShareState: NewPDEShareState(),
		objectType:    PDEShareObjectType,
		deleted:       false,
	}
}
func newPDEShareObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*PDEShareObject, error) {
	var newPDEShareState = NewPDEShareState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newPDEShareState)
		if err != nil {
			return nil, err
		}
	} else {
		newPDEShareState, ok = data.(*PDEShareState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPDEShareStateType, reflect.TypeOf(data))
		}
	}
	return &PDEShareObject{
		version:       defaultVersion,
		pdeShareHash:  key,
		pdeShareState: newPDEShareState,
		db:            db,
		objectType:    PDEShareObjectType,
		deleted:       false,
	}, nil
}

func GeneratePDEShareObjectKey(beaconHeight uint64, token1ID, token2ID, contributorAddress string) common.Hash {
	prefixHash := GetPDESharePrefix(beaconHeight)
	valueHash := common.HashH([]byte(token1ID + token2ID + contributorAddress))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t PDEShareObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *PDEShareObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t PDEShareObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *PDEShareObject) SetValue(data interface{}) error {
	newPDEShareState, ok := data.(*PDEShareState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPDEShareStateType, reflect.TypeOf(data))
	}
	t.pdeShareState = newPDEShareState
	return nil
}

func (t PDEShareObject) GetValue() interface{} {
	return t.pdeShareState
}

func (t PDEShareObject) GetValueBytes() []byte {
	pdeShareState, ok := t.GetValue().(*PDEShareState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(pdeShareState)
	if err != nil {
		panic("failed to marshal pde share state")
	}
	return value
}

func (t PDEShareObject) GetHash() common.Hash {
	return t.pdeShareHash
}

func (t PDEShareObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *PDEShareObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *PDEShareObject) Reset() bool {
	t.pdeShareState = NewPDEShareState()
	return true
}

func (t PDEShareObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t PDEShareObject) IsEmpty() bool {
	temp := NewPDEShareState()
	return reflect.DeepEqual(temp, t.pdeShareState) || t.pdeShareState == nil
}
