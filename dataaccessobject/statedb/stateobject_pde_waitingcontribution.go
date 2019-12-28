package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type WaitingPDEContributionState struct {
	beaconHeight       uint64
	pairID             string
	contributorAddress string
	tokenID            string
	amount             uint64
	txReqID            common.Hash
}

func (wc WaitingPDEContributionState) TxReqID() common.Hash {
	return wc.txReqID
}

func (wc *WaitingPDEContributionState) SetTxReqID(txReqID common.Hash) {
	wc.txReqID = txReqID
}

func (wc WaitingPDEContributionState) Amount() uint64 {
	return wc.amount
}

func (wc *WaitingPDEContributionState) SetAmount(amount uint64) {
	wc.amount = amount
}

func (wc WaitingPDEContributionState) TokenID() string {
	return wc.tokenID
}

func (wc *WaitingPDEContributionState) SetTokenID(tokenID string) {
	wc.tokenID = tokenID
}

func (wc WaitingPDEContributionState) ContributorAddress() string {
	return wc.contributorAddress
}

func (wc *WaitingPDEContributionState) SetContributorAddress(contributorAddress string) {
	wc.contributorAddress = contributorAddress
}

func (wc WaitingPDEContributionState) PairID() string {
	return wc.pairID
}

func (wc *WaitingPDEContributionState) SetPairID(pairID string) {
	wc.pairID = pairID
}

func (wc WaitingPDEContributionState) BeaconHeight() uint64 {
	return wc.beaconHeight
}

func (wc *WaitingPDEContributionState) SetBeaconHeight(beaconHeight uint64) {
	wc.beaconHeight = beaconHeight
}

func (wc WaitingPDEContributionState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		BeaconHeight       uint64
		PairID             string
		ContributorAddress string
		TokenID            string
		Amount             uint64
		TxReqID            common.Hash
	}{
		BeaconHeight:       wc.beaconHeight,
		PairID:             wc.pairID,
		ContributorAddress: wc.contributorAddress,
		TokenID:            wc.tokenID,
		Amount:             wc.amount,
		TxReqID:            wc.txReqID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (wc *WaitingPDEContributionState) UnmarshalJSON(data []byte) error {
	temp := struct {
		BeaconHeight       uint64
		PairID             string
		ContributorAddress string
		TokenID            string
		Amount             uint64
		TxReqID            common.Hash
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	wc.beaconHeight = temp.BeaconHeight
	wc.pairID = temp.PairID
	wc.contributorAddress = temp.ContributorAddress
	wc.tokenID = temp.TokenID
	wc.amount = temp.Amount
	wc.txReqID = temp.TxReqID
	return nil
}

func NewWaitingPDEContributionState() *WaitingPDEContributionState {
	return &WaitingPDEContributionState{}
}
func NewWaitingPDEContributionStateWithValue(beaconHeight uint64, pairID string, contributorAddress string, tokenID string, amount uint64, txReqID common.Hash) *WaitingPDEContributionState {
	return &WaitingPDEContributionState{beaconHeight: beaconHeight, pairID: pairID, contributorAddress: contributorAddress, tokenID: tokenID, amount: amount, txReqID: txReqID}
}

type WaitingPDEContributionObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                     int
	waitingPDEContributionHash  common.Hash
	waitingPDEContributionState *WaitingPDEContributionState
	objectType                  int
	deleted                     bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newWaitingPDEContributionObject(db *StateDB, hash common.Hash) *WaitingPDEContributionObject {
	return &WaitingPDEContributionObject{
		version:                     defaultVersion,
		db:                          db,
		waitingPDEContributionHash:  hash,
		waitingPDEContributionState: NewWaitingPDEContributionState(),
		objectType:                  WaitingPDEContributionObjectType,
		deleted:                     false,
	}
}
func newWaitingPDEContributionObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*WaitingPDEContributionObject, error) {
	var newWaitingPDEContributionState = NewWaitingPDEContributionState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newWaitingPDEContributionState)
		if err != nil {
			return nil, err
		}
	} else {
		newWaitingPDEContributionState, ok = data.(*WaitingPDEContributionState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidWaitingPDEContributionStateType, reflect.TypeOf(data))
		}
	}
	return &WaitingPDEContributionObject{
		version:                     defaultVersion,
		waitingPDEContributionHash:  key,
		waitingPDEContributionState: newWaitingPDEContributionState,
		db:                          db,
		objectType:                  WaitingPDEContributionObjectType,
		deleted:                     false,
	}, nil
}

func GenerateWaitingPDEContributionObjectKey(beaconHeight uint64, pairID string) common.Hash {
	prefixHash := GetWaitingPDEContributionPrefix(beaconHeight)
	valueHash := common.HashH([]byte(pairID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t WaitingPDEContributionObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *WaitingPDEContributionObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t WaitingPDEContributionObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *WaitingPDEContributionObject) SetValue(data interface{}) error {
	newWaitingPDEContributionState, ok := data.(*WaitingPDEContributionState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidWaitingPDEContributionStateType, reflect.TypeOf(data))
	}
	t.waitingPDEContributionState = newWaitingPDEContributionState
	return nil
}

func (t WaitingPDEContributionObject) GetValue() interface{} {
	return t.waitingPDEContributionState
}

func (t WaitingPDEContributionObject) GetValueBytes() []byte {
	waitingPDEcontributionState, ok := t.GetValue().(*WaitingPDEContributionState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(waitingPDEcontributionState)
	if err != nil {
		panic("failed to marshal waiting pde contribution state")
	}
	return value
}

func (t WaitingPDEContributionObject) GetHash() common.Hash {
	return t.waitingPDEContributionHash
}

func (t WaitingPDEContributionObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *WaitingPDEContributionObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *WaitingPDEContributionObject) Reset() bool {
	t.waitingPDEContributionState = NewWaitingPDEContributionState()
	return true
}

func (t WaitingPDEContributionObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t WaitingPDEContributionObject) IsEmpty() bool {
	temp := NewWaitingPDEContributionState()
	return reflect.DeepEqual(temp, t.waitingPDEContributionState) || t.waitingPDEContributionState == nil
}
