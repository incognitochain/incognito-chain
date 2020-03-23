package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type CommitteeRewardState struct {
	// tokenid => amount
	incognitoPublicKey string
	reward             map[common.Hash]uint64
}

func NewCommitteeRewardState() *CommitteeRewardState {
	return &CommitteeRewardState{}
}

func NewCommitteeRewardStateWithValue(reward map[common.Hash]uint64, incognitoPublicKey string) *CommitteeRewardState {
	return &CommitteeRewardState{reward: reward, incognitoPublicKey: incognitoPublicKey}
}

func (cr CommitteeRewardState) Reward() map[common.Hash]uint64 {
	return cr.reward
}

func (cr *CommitteeRewardState) SetReward(reward map[common.Hash]uint64) {
	cr.reward = reward
}

func (cr CommitteeRewardState) IncognitoPublicKey() string {
	return cr.incognitoPublicKey
}

func (cr *CommitteeRewardState) SetIncognitoPublicKey(incognitoPublicKey string) {
	cr.incognitoPublicKey = incognitoPublicKey
}

func (c CommitteeRewardState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Reward             map[common.Hash]uint64
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

func (c *CommitteeRewardState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Reward             map[common.Hash]uint64
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

type CommitteeRewardObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                int
	committeePublicKeyHash common.Hash
	rewardReceiverState    *CommitteeRewardState
	objectType             int
	deleted                bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCommitteeRewardObject(db *StateDB, hash common.Hash) *CommitteeRewardObject {
	return &CommitteeRewardObject{
		version:                defaultVersion,
		db:                     db,
		committeePublicKeyHash: hash,
		rewardReceiverState:    NewCommitteeRewardState(),
		objectType:             CommitteeRewardObjectType,
		deleted:                false,
	}
}

func newCommitteeRewardObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*CommitteeRewardObject, error) {
	var newCommitteeRewardState = NewCommitteeRewardState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newCommitteeRewardState)
		if err != nil {
			return nil, err
		}
	} else {
		newCommitteeRewardState, ok = data.(*CommitteeRewardState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidCommitteeRewardStateType, reflect.TypeOf(data))
		}
	}
	if err := SoValidation.ValidateIncognitoPublicKeySanity(newCommitteeRewardState.incognitoPublicKey); err != nil {
		return nil, fmt.Errorf("%+v, got err %+v", ErrInvalidIncognitoPublicKeyType, err)
	}
	return &CommitteeRewardObject{
		version:                defaultVersion,
		committeePublicKeyHash: key,
		rewardReceiverState:    newCommitteeRewardState,
		db:                     db,
		objectType:             CommitteeRewardObjectType,
		deleted:                false,
	}, nil
}

func GenerateCommitteeRewardObjectKey(publicKey string) (common.Hash, error) {
	//err := SoValidation.ValidateIncognitoPublicKeySanity(publicKey)
	//if err != nil {
	//	return common.Hash{}, fmt.Errorf("%+v, got err %+v", ErrInvalidIncognitoPublicKeyType, err)
	//}
	publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%+v, got err %+v", ErrInvalidIncognitoPublicKeyType, err)
	}
	prefixHash := GetCommitteeRewardPrefix()
	valueHash := common.HashH(publicKeyBytes)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...)), nil
}

func (rr CommitteeRewardObject) GetVersion() int {
	return rr.version
}

// setError remembers the first non-nil error it is called with.
func (rr *CommitteeRewardObject) SetError(err error) {
	if rr.dbErr == nil {
		rr.dbErr = err
	}
}

func (rr CommitteeRewardObject) GetTrie(db DatabaseAccessWarper) Trie {
	return rr.trie
}

func (rr *CommitteeRewardObject) SetValue(data interface{}) error {
	var newCommitteeRewardState = NewCommitteeRewardState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newCommitteeRewardState)
		if err != nil {
			return err
		}
	} else {
		newCommitteeRewardState, ok = data.(*CommitteeRewardState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidCommitteeRewardStateType, reflect.TypeOf(data))
		}
	}
	if err := SoValidation.ValidateIncognitoPublicKeySanity(newCommitteeRewardState.incognitoPublicKey); err != nil {
		return fmt.Errorf("%+v, got err %+v", ErrInvalidIncognitoPublicKeyType, err)
	}
	rr.rewardReceiverState = newCommitteeRewardState
	return nil
}

func (rr CommitteeRewardObject) GetValue() interface{} {
	return rr.rewardReceiverState
}

func (rr CommitteeRewardObject) GetValueBytes() []byte {
	data := rr.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal committee reward state")
	}
	return []byte(value)
}

func (rr CommitteeRewardObject) GetHash() common.Hash {
	return rr.committeePublicKeyHash
}

func (rr CommitteeRewardObject) GetType() int {
	return rr.objectType
}

// MarkDelete will delete an object in trie
func (rr *CommitteeRewardObject) MarkDelete() {
	rr.deleted = true
}

func (rr *CommitteeRewardObject) Reset() bool {
	rr.rewardReceiverState = NewCommitteeRewardState()
	return true
}

func (rr CommitteeRewardObject) IsDeleted() bool {
	return rr.deleted
}

// value is either default or nil
func (rr CommitteeRewardObject) IsEmpty() bool {
	temp := NewCommitteeRewardState()
	return reflect.DeepEqual(temp, rr.rewardReceiverState) || rr.rewardReceiverState == nil
}
