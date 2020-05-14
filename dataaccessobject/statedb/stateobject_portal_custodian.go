package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

type CustodianState struct {
	incognitoAddress       string
	totalCollateral        uint64            // prv
	freeCollateral         uint64            // prv
	holdingPubTokens       map[string]uint64 // tokenID : amount
	lockedAmountCollateral map[string]uint64 // tokenID : amount
	remoteAddresses        map[string]string // tokenID : remote address
	rewardAmount           map[string]uint64 // tokenID : amount
}

func (cs CustodianState) GetIncognitoAddress() string {
	return cs.incognitoAddress
}

func (cs *CustodianState) SetIncognitoAddress(incognitoAddress string) {
	cs.incognitoAddress = incognitoAddress
}

func (cs CustodianState) GetTotalCollateral() uint64 {
	return cs.totalCollateral
}

func (cs *CustodianState) SetTotalCollateral(amount uint64) {
	cs.totalCollateral = amount
}

func (cs CustodianState) GetHoldingPublicTokens() map[string]uint64 {
	return cs.holdingPubTokens
}

func (cs *CustodianState) SetHoldingPublicTokens(holdingPublicTokens map[string]uint64) {
	cs.holdingPubTokens = holdingPublicTokens
}

func (cs CustodianState) GetLockedAmountCollateral() map[string]uint64 {
	return cs.lockedAmountCollateral
}

func (cs *CustodianState) SetLockedAmountCollateral(lockedAmountCollateral map[string]uint64) {
	cs.lockedAmountCollateral = lockedAmountCollateral
}

func (cs CustodianState) GetRemoteAddresses() map[string]string {
	return cs.remoteAddresses
}

func (cs *CustodianState) SetRemoteAddresses(remoteAddresses map[string]string) {
	cs.remoteAddresses = remoteAddresses
}

func (cs CustodianState) GetFreeCollateral() uint64 {
	return cs.freeCollateral
}

func (cs *CustodianState) SetFreeCollateral(amount uint64) {
	cs.freeCollateral = amount
}

func (cs CustodianState) GetRewardAmount() map[string]uint64 {
	return cs.rewardAmount
}

func (cs *CustodianState) SetRewardAmount(amount map[string]uint64) {
	cs.rewardAmount = amount
}

func (cs CustodianState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		IncognitoAddress       string
		TotalCollateral        uint64
		FreeCollateral         uint64
		HoldingPubTokens       map[string]uint64
		LockedAmountCollateral map[string]uint64
		RemoteAddresses        map[string]string
		RewardAmount           map[string]uint64
	}{
		IncognitoAddress:       cs.incognitoAddress,
		TotalCollateral:        cs.totalCollateral,
		FreeCollateral:         cs.freeCollateral,
		HoldingPubTokens:       cs.holdingPubTokens,
		LockedAmountCollateral: cs.lockedAmountCollateral,
		RemoteAddresses:        cs.remoteAddresses,
		RewardAmount:           cs.rewardAmount,
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
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	cs.incognitoAddress = temp.IncognitoAddress
	cs.totalCollateral = temp.TotalCollateral
	cs.freeCollateral = temp.FreeCollateral
	cs.holdingPubTokens = temp.HoldingPubTokens
	cs.lockedAmountCollateral = temp.LockedAmountCollateral
	cs.remoteAddresses = temp.RemoteAddresses
	cs.rewardAmount = temp.RewardAmount
	return nil
}

// GetRemoteAddressByTokenID returns remote address for tokenID
//func GetRemoteAddressByTokenID(addresses []RemoteAddress, tokenID string) (string, error) {
//	for _, addr := range addresses {
//		if addr.GetPTokenID() == tokenID {
//			return addr.GetAddress(), nil
//		}
//	}
//
//	return "", errors.New("Can not found address with tokenID")
//}

func NewCustodianState() *CustodianState {
	return &CustodianState{
		rewardAmount:           map[string]uint64{},
		holdingPubTokens:       map[string]uint64{},
		lockedAmountCollateral: map[string]uint64{},
	}
}

func NewCustodianStateWithValue(
	incognitoAddress string,
	totalCollateral uint64,
	freeCollateral uint64,
	holdingPubTokens map[string]uint64,
	lockedAmountCollateral map[string]uint64,
	remoteAddresses map[string]string,
	rewardAmount map[string]uint64) *CustodianState {

	return &CustodianState{
		incognitoAddress:       incognitoAddress,
		totalCollateral:        totalCollateral,
		freeCollateral:         freeCollateral,
		holdingPubTokens:       holdingPubTokens,
		lockedAmountCollateral: lockedAmountCollateral,
		remoteAddresses:        remoteAddresses,
		rewardAmount:           rewardAmount,
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
