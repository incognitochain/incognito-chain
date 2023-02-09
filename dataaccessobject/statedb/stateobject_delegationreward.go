package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type DelegateInfo struct {
	BeaconUID string
	Amount    int
}
type DelegationRewardState struct {
	//reward receiver
	incognitoPublicKey string
	// shard cpk => affect epoch => info
	reward map[string]map[int]DelegateInfo
}

func NewDelegationRewardState() *DelegationRewardState {
	return &DelegationRewardState{
		reward: map[string]map[int]DelegateInfo{},
	}
}

func (cr DelegationRewardState) IncognitoPublicKey() string {
	return cr.incognitoPublicKey
}

func (cr *DelegationRewardState) SetIncognitoPublicKey(incognitoPublicKey string) {
	cr.incognitoPublicKey = incognitoPublicKey
}

func (c DelegationRewardState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Reward             map[string]map[int]DelegateInfo
		IncognitoPublicKey string
	}{
		Reward:             c.reward,
		IncognitoPublicKey: c.incognitoPublicKey,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *DelegationRewardState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Reward             map[string]map[int]DelegateInfo
		IncognitoPublicKey string
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.reward = temp.Reward
	c.incognitoPublicKey = temp.IncognitoPublicKey
	return nil
}

type DelegationRewardObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                int
	committeePublicKeyHash common.Hash
	rewardReceiverState    *DelegationRewardState
	objectType             int
	deleted                bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newDelegationRewardObject(db *StateDB, hash common.Hash) *DelegationRewardObject {
	return &DelegationRewardObject{
		version:                defaultVersion,
		db:                     db,
		committeePublicKeyHash: hash,
		rewardReceiverState:    NewDelegationRewardState(),
		objectType:             CommitteeRewardObjectType,
		deleted:                false,
	}
}

func newDelegationRewardObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*DelegationRewardObject, error) {
	var newDelegationRewardState = NewDelegationRewardState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newDelegationRewardState)
		if err != nil {
			return nil, err
		}
	} else {
		newDelegationRewardState, ok = data.(*DelegationRewardState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidDelegationRewardStateType, reflect.TypeOf(data))
		}
	}
	if err := SoValidation.ValidateIncognitoPublicKeySanity(newDelegationRewardState.incognitoPublicKey); err != nil {
		return nil, fmt.Errorf("%+v, got err %+v", ErrInvalidIncognitoPublicKeyType, err)
	}
	return &DelegationRewardObject{
		version:                defaultVersion,
		committeePublicKeyHash: key,
		rewardReceiverState:    newDelegationRewardState,
		db:                     db,
		objectType:             CommitteeRewardObjectType,
		deleted:                false,
	}, nil
}

func GenerateDelegateRewardObjectKey(publicKey string) (common.Hash, error) {
	publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%+v, got err %+v", ErrInvalidIncognitoPublicKeyType, err)
	}
	prefixHash := GetDelegationRewardPrefix()
	valueHash := common.HashH(publicKeyBytes)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...)), nil
}

func (rr DelegationRewardObject) GetVersion() int {
	return rr.version
}

// setError remembers the first non-nil error it is called with.
func (rr *DelegationRewardObject) SetError(err error) {
	if rr.dbErr == nil {
		rr.dbErr = err
	}
}

func (rr DelegationRewardObject) GetTrie(db DatabaseAccessWarper) Trie {
	return rr.trie
}

func (rr *DelegationRewardObject) SetValue(data interface{}) error {
	var newDelegationRewardState = NewDelegationRewardState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newDelegationRewardState)
		if err != nil {
			return err
		}
	} else {
		newDelegationRewardState, ok = data.(*DelegationRewardState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidDelegationRewardStateType, reflect.TypeOf(data))
		}
	}
	if err := SoValidation.ValidateIncognitoPublicKeySanity(newDelegationRewardState.incognitoPublicKey); err != nil {
		return fmt.Errorf("%+v, got err %+v", ErrInvalidIncognitoPublicKeyType, err)
	}
	rr.rewardReceiverState = newDelegationRewardState
	return nil
}

func (rr DelegationRewardObject) GetValue() interface{} {
	return rr.rewardReceiverState
}

func (rr DelegationRewardObject) GetValueBytes() []byte {
	data := rr.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal committee reward state")
	}
	return []byte(value)
}

func (rr DelegationRewardObject) GetHash() common.Hash {
	return rr.committeePublicKeyHash
}

func (rr DelegationRewardObject) GetType() int {
	return rr.objectType
}

// MarkDelete will delete an object in trie
func (rr *DelegationRewardObject) MarkDelete() {
	rr.deleted = true
}

func (rr *DelegationRewardObject) Reset() bool {
	rr.rewardReceiverState = NewDelegationRewardState()
	return true
}

func (rr DelegationRewardObject) IsDeleted() bool {
	return rr.deleted
}

// value is either default or nil
func (rr DelegationRewardObject) IsEmpty() bool {
	temp := NewDelegationRewardState()
	return reflect.DeepEqual(temp, rr.rewardReceiverState) || rr.rewardReceiverState == nil
}
