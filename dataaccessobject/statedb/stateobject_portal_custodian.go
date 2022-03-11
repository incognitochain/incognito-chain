package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type CustodianState struct {
	IncognitoAddress       string
	TotalCollateral        uint64            // prv
	FreeCollateral         uint64            // prv
	LockedAmountCollateral map[string]uint64 // ptokenID : amount
	HoldingPubTokens       map[string]uint64 // ptokenID : amount
	RemoteAddresses        map[string]string // ptokenID : remote address
	RewardAmount           map[string]uint64 // ptokenID : amount

	TotalTokenCollaterals  map[string]uint64            // publicTokenID : amount
	FreeTokenCollaterals   map[string]uint64            // publicTokenID : amount
	LockedTokenCollaterals map[string]map[string]uint64 // ptokenID: publicTokenID : amount
}

func (cs CustodianState) GetIncognitoAddress() string {
	return cs.IncognitoAddress
}

func (cs *CustodianState) SetIncognitoAddress(incognitoAddress string) {
	cs.IncognitoAddress = incognitoAddress
}

func (cs CustodianState) GetTotalCollateral() uint64 {
	return cs.TotalCollateral
}

func (cs *CustodianState) SetTotalCollateral(amount uint64) {
	cs.TotalCollateral = amount
}

func (cs CustodianState) GetHoldingPublicTokens() map[string]uint64 {
	return cs.HoldingPubTokens
}

func (cs *CustodianState) SetHoldingPublicTokens(holdingPublicTokens map[string]uint64) {
	cs.HoldingPubTokens = holdingPublicTokens
}

func (cs CustodianState) GetLockedAmountCollateral() map[string]uint64 {
	return cs.LockedAmountCollateral
}

func (cs *CustodianState) SetLockedAmountCollateral(lockedAmountCollateral map[string]uint64) {
	cs.LockedAmountCollateral = lockedAmountCollateral
}

func (cs CustodianState) GetRemoteAddresses() map[string]string {
	return cs.RemoteAddresses
}

func (cs *CustodianState) SetRemoteAddresses(remoteAddresses map[string]string) {
	cs.RemoteAddresses = remoteAddresses
}

func (cs CustodianState) GetFreeCollateral() uint64 {
	return cs.FreeCollateral
}

func (cs *CustodianState) SetFreeCollateral(amount uint64) {
	cs.FreeCollateral = amount
}

func (cs CustodianState) GetRewardAmount() map[string]uint64 {
	return cs.RewardAmount
}

func (cs *CustodianState) SetRewardAmount(amount map[string]uint64) {
	cs.RewardAmount = amount
}

func (cs CustodianState) GetTotalTokenCollaterals() map[string]uint64 {
	return cs.TotalTokenCollaterals
}

func (cs *CustodianState) SetTotalTokenCollaterals(totalTokenCollaterals map[string]uint64) {
	cs.TotalTokenCollaterals = totalTokenCollaterals
}

func (cs CustodianState) GetFreeTokenCollaterals() map[string]uint64 {
	return cs.FreeTokenCollaterals
}

func (cs *CustodianState) SetFreeTokenCollaterals(freeTokenCollaterals map[string]uint64) {
	cs.FreeTokenCollaterals = freeTokenCollaterals
}

func (cs CustodianState) GetLockedTokenCollaterals() map[string]map[string]uint64 {
	return cs.LockedTokenCollaterals
}

func (cs *CustodianState) SetLockedTokenCollaterals(lockedTokenCollaterals map[string]map[string]uint64) {
	cs.LockedTokenCollaterals = lockedTokenCollaterals
}

func (cs CustodianState) IsEmptyCollaterals() bool {
	if cs.TotalCollateral > 0 {
		return false
	}

	if len(cs.TotalTokenCollaterals) > 0 {
		for _, amount := range cs.TotalTokenCollaterals {
			if amount > 0 {
				return false
			}
		}
	}

	return true
}

func (cs *CustodianState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		IncognitoAddress       string
		TotalCollateral        uint64
		FreeCollateral         uint64
		HoldingPubTokens       map[string]uint64
		LockedAmountCollateral map[string]uint64
		RemoteAddresses        map[string]string
		RewardAmount           map[string]uint64
		TotalTokenCollaterals  map[string]uint64
		FreeTokenCollaterals   map[string]uint64
		LockedTokenCollaterals map[string]map[string]uint64
	}{
		IncognitoAddress:       cs.IncognitoAddress,
		TotalCollateral:        cs.TotalCollateral,
		FreeCollateral:         cs.FreeCollateral,
		HoldingPubTokens:       cs.HoldingPubTokens,
		LockedAmountCollateral: cs.LockedAmountCollateral,
		RemoteAddresses:        cs.RemoteAddresses,
		RewardAmount:           cs.RewardAmount,
		TotalTokenCollaterals:  cs.TotalTokenCollaterals,
		FreeTokenCollaterals:   cs.FreeTokenCollaterals,
		LockedTokenCollaterals: cs.LockedTokenCollaterals,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (cs *CustodianState) UnmarshalJSON(data []byte) error {
	temp := struct {
		IncognitoAddress       string
		TotalCollateral        uint64
		FreeCollateral         uint64
		HoldingPubTokens       map[string]uint64
		LockedAmountCollateral map[string]uint64
		RemoteAddresses        map[string]string
		RewardAmount           map[string]uint64
		TotalTokenCollaterals  map[string]uint64
		FreeTokenCollaterals   map[string]uint64
		LockedTokenCollaterals map[string]map[string]uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	cs.IncognitoAddress = temp.IncognitoAddress
	cs.TotalCollateral = temp.TotalCollateral
	cs.FreeCollateral = temp.FreeCollateral
	cs.HoldingPubTokens = temp.HoldingPubTokens
	cs.LockedAmountCollateral = temp.LockedAmountCollateral
	cs.RemoteAddresses = temp.RemoteAddresses
	cs.RewardAmount = temp.RewardAmount
	cs.TotalTokenCollaterals = temp.TotalTokenCollaterals
	cs.FreeTokenCollaterals = temp.FreeTokenCollaterals
	cs.LockedTokenCollaterals = temp.LockedTokenCollaterals
	return nil
}

func NewCustodianState() *CustodianState {
	return &CustodianState{
		IncognitoAddress:       "",
		TotalCollateral:        0,
		FreeCollateral:         0,
		LockedAmountCollateral: map[string]uint64{},
		HoldingPubTokens:       map[string]uint64{},
		RemoteAddresses:        map[string]string{},
		RewardAmount:           map[string]uint64{},
		TotalTokenCollaterals:  map[string]uint64{},
		FreeTokenCollaterals:   map[string]uint64{},
		LockedTokenCollaterals: map[string]map[string]uint64{},
	}
}

func NewCustodianStateWithValue(
	incognitoAddress string,
	totalCollateral uint64,
	freeCollateral uint64,
	holdingPubTokens map[string]uint64,
	lockedAmountCollateral map[string]uint64,
	remoteAddresses map[string]string,
	rewardAmount map[string]uint64,
	totalTokenCollaterals map[string]uint64,
	freeTokenCollaterals map[string]uint64,
	lockedTokenCollaterals map[string]map[string]uint64,
) *CustodianState {

	return &CustodianState{
		IncognitoAddress:       incognitoAddress,
		TotalCollateral:        totalCollateral,
		FreeCollateral:         freeCollateral,
		LockedAmountCollateral: lockedAmountCollateral,
		HoldingPubTokens:       holdingPubTokens,
		RemoteAddresses:        remoteAddresses,
		RewardAmount:           rewardAmount,
		TotalTokenCollaterals:  totalTokenCollaterals,
		FreeTokenCollaterals:   freeTokenCollaterals,
		LockedTokenCollaterals: lockedTokenCollaterals,
	}
}

type CustodianStateObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version            int
	custodianStateHash common.Hash
	custodianState     *CustodianState
	objectType         int
	deleted            bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCustodianStateObject(db *StateDB, hash common.Hash) *CustodianStateObject {
	return &CustodianStateObject{
		version:            defaultVersion,
		db:                 db,
		custodianStateHash: hash,
		custodianState:     NewCustodianState(),
		objectType:         CustodianStateObjectType,
		deleted:            false,
	}
}

func newCustodianStateObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*CustodianStateObject, error) {
	var custodianState = NewCustodianState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, custodianState)
		if err != nil {
			return nil, err
		}
	} else {
		custodianState, ok = data.(*CustodianState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalCustodianStateType, reflect.TypeOf(data))
		}
	}
	return &CustodianStateObject{
		version:            defaultVersion,
		custodianStateHash: key,
		custodianState:     custodianState,
		db:                 db,
		objectType:         CustodianStateObjectType,
		deleted:            false,
	}, nil
}

func GenerateCustodianStateObjectKey(custodianIncognitoAddress string) common.Hash {
	prefixHash := GetPortalCustodianStatePrefix()
	valueHash := common.HashH([]byte(custodianIncognitoAddress))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t CustodianStateObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *CustodianStateObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t CustodianStateObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *CustodianStateObject) SetValue(data interface{}) error {
	newCustodianState, ok := data.(*CustodianState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalCustodianStateType, reflect.TypeOf(data))
	}
	t.custodianState = newCustodianState
	return nil
}

func (t CustodianStateObject) GetValue() interface{} {
	return t.custodianState
}

func (t CustodianStateObject) GetValueBytes() []byte {
	custodianState, ok := t.GetValue().(*CustodianState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(custodianState)
	if err != nil {
		panic("failed to marshal custodian state")
	}
	return value
}

func (t CustodianStateObject) GetHash() common.Hash {
	return t.custodianStateHash
}

func (t CustodianStateObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *CustodianStateObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *CustodianStateObject) Reset() bool {
	t.custodianState = NewCustodianState()
	return true
}

func (t CustodianStateObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t CustodianStateObject) IsEmpty() bool {
	temp := NewCustodianState()
	return reflect.DeepEqual(temp, t.custodianState) || t.custodianState == nil
}
