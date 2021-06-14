package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type RewardRequestStateV3 struct {
	// tokenid => amount
	epoch    uint64
	shardID  byte
	subsetID byte
	tokenID  common.Hash
	amount   uint64
}

func (rr RewardRequestStateV3) Amount() uint64 {
	return rr.amount
}

func (rr *RewardRequestStateV3) SetAmount(amount uint64) {
	rr.amount = amount
}

func (rr RewardRequestStateV3) TokenID() common.Hash {
	return rr.tokenID
}

func (rr *RewardRequestStateV3) SetTokenID(tokenID common.Hash) {
	rr.tokenID = tokenID
}

func (rr RewardRequestStateV3) ShardID() byte {
	return rr.shardID
}

func (rr *RewardRequestStateV3) SetShardID(shardID byte) {
	rr.shardID = shardID
}

func (rr RewardRequestStateV3) SubsetID() byte {
	return rr.subsetID
}

func (rr *RewardRequestStateV3) SetSubsetID(subsetID byte) {
	rr.subsetID = subsetID
}

func (rr RewardRequestStateV3) Epoch() uint64 {
	return rr.epoch
}

func (rr *RewardRequestStateV3) SetEpoch(epoch uint64) {
	rr.epoch = epoch
}

func NewRewardRequestStateV3() *RewardRequestStateV3 {
	return &RewardRequestStateV3{}
}

func NewRewardRequestStateV3WithValue(
	epoch uint64,
	shardID, subsetID byte,
	tokenID common.Hash,
	amount uint64,
) *RewardRequestStateV3 {
	return &RewardRequestStateV3{
		epoch:    epoch,
		shardID:  shardID,
		subsetID: subsetID,
		tokenID:  tokenID,
		amount:   amount,
	}
}

func (c RewardRequestStateV3) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Epoch    uint64
		ShardID  byte
		SubsetID byte
		TokenID  common.Hash
		Amount   uint64
	}{
		Epoch:    c.epoch,
		ShardID:  c.shardID,
		SubsetID: c.subsetID,
		TokenID:  c.tokenID,
		Amount:   c.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *RewardRequestStateV3) UnmarshalJSON(data []byte) error {
	temp := struct {
		Epoch    uint64
		ShardID  byte
		SubsetID byte
		TokenID  common.Hash
		Amount   uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.epoch = temp.Epoch
	c.shardID = temp.ShardID
	c.subsetID = temp.SubsetID
	c.tokenID = temp.TokenID
	c.amount = temp.Amount
	return nil
}

type RewardRequestObjectV3 struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version             int
	rewardRequestHash   common.Hash
	rewardReceiverState *RewardRequestStateV3
	objectType          int
	deleted             bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newRewardRequestV3Object(db *StateDB, hash common.Hash) *RewardRequestObjectV3 {
	return &RewardRequestObjectV3{
		version:             defaultVersion,
		db:                  db,
		rewardRequestHash:   hash,
		rewardReceiverState: NewRewardRequestStateV3(),
		objectType:          RewardRequestV3ObjectType,
		deleted:             false,
	}
}

func newRewardRequestV3ObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*RewardRequestObjectV3, error) {
	var newRewardRequestState = NewRewardRequestStateV3()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newRewardRequestState)
		if err != nil {
			return nil, err
		}
	} else {
		newRewardRequestState, ok = data.(*RewardRequestStateV3)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidRewardRequestStateType, reflect.TypeOf(data))
		}
	}
	return &RewardRequestObjectV3{
		version:             defaultVersion,
		rewardRequestHash:   key,
		rewardReceiverState: newRewardRequestState,
		db:                  db,
		objectType:          RewardRequestV3ObjectType,
		deleted:             false,
	}, nil
}

func (rr RewardRequestObjectV3) GetVersion() int {
	return rr.version
}

// setError remembers the first non-nil error it is called with.
func (rr *RewardRequestObjectV3) SetError(err error) {
	if rr.dbErr == nil {
		rr.dbErr = err
	}
}

func (rr RewardRequestObjectV3) GetTrie(db DatabaseAccessWarper) Trie {
	return rr.trie
}

func (rr *RewardRequestObjectV3) SetValue(data interface{}) error {
	var newRewardRequestState = NewRewardRequestStateV3()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newRewardRequestState)
		if err != nil {
			return err
		}
	} else {
		newRewardRequestState, ok = data.(*RewardRequestStateV3)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidRewardRequestStateType, reflect.TypeOf(data))
		}
	}
	rr.rewardReceiverState = newRewardRequestState
	return nil
}

func (rr RewardRequestObjectV3) GetValue() interface{} {
	return rr.rewardReceiverState
}

func (rr RewardRequestObjectV3) GetValueBytes() []byte {
	data := rr.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal reward request state")
	}
	return []byte(value)
}

func (rr RewardRequestObjectV3) GetHash() common.Hash {
	return rr.rewardRequestHash
}

func (rr RewardRequestObjectV3) GetType() int {
	return rr.objectType
}

// MarkDelete will delete an object in trie
func (rr *RewardRequestObjectV3) MarkDelete() {
	rr.deleted = true
}

func (rr *RewardRequestObjectV3) Reset() bool {
	rr.rewardReceiverState = NewRewardRequestStateV3()
	return true
}

func (rr RewardRequestObjectV3) IsDeleted() bool {
	return rr.deleted
}

// value is either default or nil
func (rr RewardRequestObjectV3) IsEmpty() bool {
	temp := NewRewardRequestStateV3()
	return reflect.DeepEqual(temp, rr.rewardReceiverState) || rr.rewardReceiverState == nil
}

func generateRewardRequestObjectKeyV3(epoch uint64, shardID, subsetID byte, tokenID common.Hash) common.Hash {
	prefixHash := GetRewardRequestPrefix(epoch)
	tempPrefix := append([]byte{shardID}, []byte{subsetID}...)
	valueHash := common.HashH(append(tempPrefix, tokenID[:]...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}
