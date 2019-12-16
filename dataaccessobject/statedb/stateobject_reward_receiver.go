package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type RewardReceiverState struct {
	PublicKey      string
	PaymentAddress string
}

func (rrs *RewardReceiverState) GetPublicKey() string {
	return rrs.PublicKey
}

func (rrs *RewardReceiverState) SetPublicKey(publicKey string) {
	rrs.PublicKey = publicKey
}

func (rrs *RewardReceiverState) GetPaymentAddress() string {
	return rrs.PublicKey
}

func (rrs *RewardReceiverState) SetPaymentAddress(publicKey string) {
	rrs.PublicKey = publicKey
}

func NewRewardReceiverState() *RewardReceiverState {
	return &RewardReceiverState{}
}

func NewRewardReceiverStateWithValue(publicKey string, paymentAddress string) *RewardReceiverState {
	return &RewardReceiverState{PublicKey: publicKey, PaymentAddress: paymentAddress}
}

type RewardReceiverObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	publicKeyHash       common.Hash
	rewardReceiverState *RewardReceiverState
	objectType          int
	deleted             bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newRewardReceiverObject(db *StateDB, hash common.Hash) *RewardReceiverObject {
	return &RewardReceiverObject{
		db:                  db,
		publicKeyHash:       hash,
		rewardReceiverState: NewRewardReceiverState(),
		objectType:          RewardReceiverObjectType,
		deleted:             false,
	}
}
func newRewardReceiverObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*RewardReceiverObject, error) {
	var newRewardReceiverState = NewRewardReceiverState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newRewardReceiverState)
		if err != nil {
			return nil, NewStatedbError(InvalidCommitteeStateTypeError, err)
		}
	} else {
		newRewardReceiverState, ok = data.(*RewardReceiverState)
		if !ok {
			return nil, NewStatedbError(InvalidCommitteeStateTypeError, fmt.Errorf("%+v", reflect.TypeOf(data)))
		}
	}
	if err := validatePaymentAddressSanity(newRewardReceiverState.PaymentAddress); err != nil {
		return nil, NewStatedbError(InvalidPaymentAddressTypeError, err)
	}
	if err := validateIncognitoPublicKeySanity(newRewardReceiverState.PublicKey); err != nil {
		return nil, NewStatedbError(InvalidIncognitoPublicKeyTypeError, err)
	}
	return &RewardReceiverObject{
		publicKeyHash:       key,
		rewardReceiverState: newRewardReceiverState,
		db:                  db,
		objectType:          RewardReceiverObjectType,
		deleted:             false,
	}, nil
}

func GenerateRewardReceiverObjectKey(publicKey string) (common.Hash, error) {
	err := validateIncognitoPublicKeySanity(publicKey)
	if err != nil {
		return common.Hash{}, NewStatedbError(InvalidIncognitoPublicKeyTypeError, err)
	}
	publicKeyBytes, _, _ := base58.Base58Check{}.Decode(publicKey)
	prefixHash := GetRewardReceiverPrefix()
	valueHash := common.HashH(publicKeyBytes)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...)), nil
}

// setError remembers the first non-nil error it is called with.
func (rr *RewardReceiverObject) SetError(err error) {
	if rr.dbErr == nil {
		rr.dbErr = err
	}
}

func (rr *RewardReceiverObject) GetTrie(db DatabaseAccessWarper) Trie {
	return rr.trie
}

func (rr *RewardReceiverObject) SetValue(data interface{}) error {
	var newRewardReceiverState = NewRewardReceiverState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newRewardReceiverState)
		if err != nil {
			return NewStatedbError(InvalidCommitteeStateTypeError, err)
		}
	} else {
		newRewardReceiverState, ok = data.(*RewardReceiverState)
		if !ok {
			return NewStatedbError(InvalidCommitteeStateTypeError, fmt.Errorf("%+v", reflect.TypeOf(data)))
		}
	}
	if err := validatePaymentAddressSanity(newRewardReceiverState.PaymentAddress); err != nil {
		return NewStatedbError(InvalidPaymentAddressTypeError, err)
	}
	if err := validateIncognitoPublicKeySanity(newRewardReceiverState.PublicKey); err != nil {
		return NewStatedbError(InvalidIncognitoPublicKeyTypeError, err)
	}
	rr.rewardReceiverState = newRewardReceiverState
	return nil
}

func (rr *RewardReceiverObject) GetValue() interface{} {
	return rr.rewardReceiverState
}

func (rr *RewardReceiverObject) GetValueBytes() []byte {
	data := rr.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal reward receiver state")
	}
	return []byte(value)
}

func (rr *RewardReceiverObject) GetHash() common.Hash {
	return rr.publicKeyHash
}

func (rr *RewardReceiverObject) GetType() int {
	return rr.objectType
}

// MarkDelete will delete an object in trie
func (rr *RewardReceiverObject) MarkDelete() {
	rr.deleted = true
}

// reset all shard committee value into default value
func (rr *RewardReceiverObject) Reset() bool {
	rr.rewardReceiverState = NewRewardReceiverState()
	return true
}

func (rr *RewardReceiverObject) IsDeleted() bool {
	return rr.deleted
}

// value is either default or nil
func (rr *RewardReceiverObject) IsEmpty() bool {
	temp := NewRewardReceiverState()
	return reflect.DeepEqual(temp, rr.rewardReceiverState) || rr.rewardReceiverState == nil
}
