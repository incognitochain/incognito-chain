package statedb

import (
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type RewardReceiverObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	publicKeyHash                common.Hash
	publicKey                    string
	rewardReceiverPaymentAddress string
	objectType                   int
	deleted                      bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newRewardReceiverObject(db *StateDB, hash common.Hash) *RewardReceiverObject {
	return &RewardReceiverObject{
		db:                           db,
		publicKeyHash:                hash,
		rewardReceiverPaymentAddress: "",
		objectType:                   RewardReceiverObjectType,
		deleted:                      false,
	}
}
func newRewardReceiverObjectWithValue(db *StateDB, key common.Hash, data interface{}) *RewardReceiverObject {
	var newRewardReceiverPaymentAddress string
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		newRewardReceiverPaymentAddress = string(dataBytes)
	} else {
		newRewardReceiverPaymentAddress, ok = data.(string)
		if !ok {
			panic("Wrong expected value")
		}
	}
	return &RewardReceiverObject{
		publicKeyHash:                key,
		rewardReceiverPaymentAddress: newRewardReceiverPaymentAddress,
		db:                           db,
		objectType:                   RewardReceiverObjectType,
		deleted:                      false,
	}
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
	newRewardReceiverPaymentAddress, ok := data.(string)
	if !ok {
		return NewStatedbError(InvalidPaymentAddressTypeError, fmt.Errorf("%+v", reflect.TypeOf(data)))
	}
	rr.rewardReceiverPaymentAddress = newRewardReceiverPaymentAddress
	return nil
}

func (rr *RewardReceiverObject) GetValue() interface{} {
	return rr.rewardReceiverPaymentAddress
}

func (rr *RewardReceiverObject) GetValueBytes() []byte {
	data := rr.GetValue()
	value, ok := data.(string)
	if !ok {
		panic("Wrong expected value")
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
	rr.rewardReceiverPaymentAddress = ""
	return true
}

func (rr *RewardReceiverObject) IsDeleted() bool {
	return rr.deleted
}

// value is either default or nil
func (rr *RewardReceiverObject) IsEmpty() bool {
	return rr.rewardReceiverPaymentAddress == ""
}

func validateValueSanity(v string) error {
	keyWalletReceiver, err := wallet.Base58CheckDeserialize(v)
	if err != nil {
		return err
	}
	if len(keyWalletReceiver.KeySet.PaymentAddress.Pk) == 0 || len(keyWalletReceiver.KeySet.PaymentAddress.Tk) == 0 {
		return fmt.Errorf("length public key %+v, length transmission key %+v", len(keyWalletReceiver.KeySet.PaymentAddress.Pk), len(keyWalletReceiver.KeySet.PaymentAddress.Tk))
	}
	return nil
}
