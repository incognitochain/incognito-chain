package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type CommitteeState struct {
	shardID            int
	role               int
	committeePublicKey incognitokey.CommitteePublicKey
	rewardReceiver     privacy.PaymentAddress
	autoStaking        bool
	enterTime          int64 // unix time
}

func NewCommitteeState() *CommitteeState {
	return &CommitteeState{}
}

func NewCommitteeStateWithValue(shardID int, role int, committeePublicKey incognitokey.CommitteePublicKey, rewardReceiver privacy.PaymentAddress, autoStaking bool) *CommitteeState {
	return &CommitteeState{shardID: shardID, role: role, committeePublicKey: committeePublicKey, rewardReceiver: rewardReceiver, autoStaking: autoStaking, enterTime: time.Now().UnixNano()}
}

func NewCommitteeStateWithValueAndTime(shardID int, role int, committeePublicKey incognitokey.CommitteePublicKey, rewardReceiver privacy.PaymentAddress, autoStaking bool, enterTime int64) *CommitteeState {
	return &CommitteeState{shardID: shardID, role: role, committeePublicKey: committeePublicKey, rewardReceiver: rewardReceiver, autoStaking: autoStaking, enterTime: enterTime}
}

func (c CommitteeState) EnterTime() int64 {
	return c.enterTime
}

func (c *CommitteeState) SetEnterTime(enterTime int64) {
	c.enterTime = enterTime
}

func (c CommitteeState) AutoStaking() bool {
	return c.autoStaking
}

func (c *CommitteeState) SetAutoStaking(autoStaking bool) {
	c.autoStaking = autoStaking
}

func (c CommitteeState) RewardReceiver() privacy.PaymentAddress {
	return c.rewardReceiver
}

func (c *CommitteeState) SetRewardReceiver(rewardReceiver privacy.PaymentAddress) {
	c.rewardReceiver = rewardReceiver
}

func (c CommitteeState) CommitteePublicKey() incognitokey.CommitteePublicKey {
	return c.committeePublicKey
}

func (c *CommitteeState) SetCommitteePublicKey(committeePublicKey incognitokey.CommitteePublicKey) {
	c.committeePublicKey = committeePublicKey
}

func (c CommitteeState) Role() int {
	return c.role
}

func (c *CommitteeState) SetRole(role int) {
	c.role = role
}

func (c CommitteeState) ShardID() int {
	return c.shardID
}

func (c *CommitteeState) SetShardID(shardID int) {
	c.shardID = shardID
}

func (c CommitteeState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ShardID            int
		Role               int
		CommitteePublicKey incognitokey.CommitteePublicKey
		RewardReceiver     privacy.PaymentAddress
		AutoStaking        bool
		EnterTime          int64
	}{
		ShardID:            c.shardID,
		Role:               c.role,
		CommitteePublicKey: c.committeePublicKey,
		RewardReceiver:     c.rewardReceiver,
		AutoStaking:        c.autoStaking,
		EnterTime:          c.enterTime,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *CommitteeState) UnmarshalJSON(data []byte) error {
	temp := struct {
		ShardID            int
		Role               int
		CommitteePublicKey incognitokey.CommitteePublicKey
		RewardReceiver     privacy.PaymentAddress
		AutoStaking        bool
		EnterTime          int64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.shardID = temp.ShardID
	c.role = temp.Role
	c.committeePublicKey = temp.CommitteePublicKey
	c.rewardReceiver = temp.RewardReceiver
	c.autoStaking = temp.AutoStaking
	c.enterTime = temp.EnterTime
	return nil
}

type CommitteeObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                int
	committeePublicKeyHash common.Hash
	committeeState         *CommitteeState
	objectType             int
	deleted                bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCommitteeObject(db *StateDB, hash common.Hash) *CommitteeObject {
	return &CommitteeObject{
		version:                defaultVersion,
		db:                     db,
		committeePublicKeyHash: hash,
		committeeState:         NewCommitteeState(),
		objectType:             CommitteeObjectType,
		deleted:                false,
	}
}

func newCommitteeObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*CommitteeObject, error) {
	var newCommitteeState = NewCommitteeState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newCommitteeState)
		if err != nil {
			return nil, err
		}
	} else {
		newCommitteeState, ok = data.(*CommitteeState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidCommitteeStateType, reflect.TypeOf(data))
		}
	}
	if err := SoValidation.ValidatePaymentAddressSanity(newCommitteeState.rewardReceiver); err != nil {
		return nil, fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	}
	return &CommitteeObject{
		version:                defaultVersion,
		committeePublicKeyHash: key,
		committeeState:         newCommitteeState,
		db:                     db,
		objectType:             CommitteeObjectType,
		deleted:                false,
	}, nil
}

func GenerateCommitteeObjectKeyWithRole(role int, shardID int, committee incognitokey.CommitteePublicKey) (common.Hash, error) {
	committeeBytes, err := committee.Bytes()
	if err != nil {
		return common.Hash{}, err
	}
	prefixHash := GetCommitteePrefixWithRole(role, shardID)
	valueHash := common.HashH(committeeBytes)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...)), nil
}

func (c CommitteeObject) GetVersion() int {
	return c.version
}

// setError remembers the first non-nil error it is called with.
func (c *CommitteeObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c CommitteeObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *CommitteeObject) SetValue(data interface{}) error {
	newCommitteeState, ok := data.(*CommitteeState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidCommitteeStateType, reflect.TypeOf(data))
	}
	if err := SoValidation.ValidatePaymentAddressSanity(newCommitteeState.rewardReceiver); err != nil {
		return fmt.Errorf("%+v, got err %+v", ErrInvalidPaymentAddressType, err)
	}
	c.committeeState = newCommitteeState
	return nil
}

func (c CommitteeObject) GetValue() interface{} {
	return c.committeeState
}

func (c CommitteeObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c CommitteeObject) GetHash() common.Hash {
	return c.committeePublicKeyHash
}

func (c CommitteeObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *CommitteeObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *CommitteeObject) Reset() bool {
	c.committeeState = NewCommitteeState()
	return true
}

func (c CommitteeObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c CommitteeObject) IsEmpty() bool {
	temp := NewCommitteeState()
	return reflect.DeepEqual(temp, c.committeeState) || c.committeeState == nil
}
