package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type RewardRequestMultisetState struct {
	// tokenid => amount
	epoch    uint64
	shardID  byte
	subsetID byte
	tokenID  common.Hash
	amount   uint64
}

func (rr RewardRequestMultisetState) Amount() uint64 {
	return rr.amount
}

func (rr *RewardRequestMultisetState) SetAmount(amount uint64) {
	rr.amount = amount
}

func (rr RewardRequestMultisetState) TokenID() common.Hash {
	return rr.tokenID
}

func (rr *RewardRequestMultisetState) SetTokenID(tokenID common.Hash) {
	rr.tokenID = tokenID
}

func (rr RewardRequestMultisetState) ShardID() byte {
	return rr.shardID
}

func (rr *RewardRequestMultisetState) SetShardID(shardID byte) {
	rr.shardID = shardID
}

func (rr RewardRequestMultisetState) SubsetID() byte {
	return rr.subsetID
}

func (rr *RewardRequestMultisetState) SetSubsetID(subsetID byte) {
	rr.subsetID = subsetID
}

func (rr RewardRequestMultisetState) Epoch() uint64 {
	return rr.epoch
}

func (rr *RewardRequestMultisetState) SetEpoch(epoch uint64) {
	rr.epoch = epoch
}

func NewRewardRequestStateV3() *RewardRequestMultisetState {
	return &RewardRequestMultisetState{}
}

func NewRewardRequestStateV3WithValue(
	epoch uint64,
	shardID, subsetID byte,
	tokenID common.Hash,
	amount uint64,
) *RewardRequestMultisetState {
	return &RewardRequestMultisetState{
		epoch:    epoch,
		shardID:  shardID,
		subsetID: subsetID,
		tokenID:  tokenID,
		amount:   amount,
	}
}

func (c RewardRequestMultisetState) MarshalJSON() ([]byte, error) {
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

func (c *RewardRequestMultisetState) UnmarshalJSON(data []byte) error {
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

type RewardRequestMultisetObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version             int
	rewardRequestHash   common.Hash
	rewardReceiverState *RewardRequestMultisetState
	objectType          int
	deleted             bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newRewardRequestMultisetObject(db *StateDB, hash common.Hash) *RewardRequestMultisetObject {
	return &RewardRequestMultisetObject{
		version:             defaultVersion,
		db:                  db,
		rewardRequestHash:   hash,
		rewardReceiverState: NewRewardRequestStateV3(),
		objectType:          RewardRequestV3ObjectType,
		deleted:             false,
	}
}

func newRewardRequestMultisetObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*RewardRequestMultisetObject, error) {
	var newRewardRequestState = NewRewardRequestStateV3()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newRewardRequestState)
		if err != nil {
			return nil, err
		}
	} else {
		newRewardRequestState, ok = data.(*RewardRequestMultisetState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidRewardRequestStateType, reflect.TypeOf(data))
		}
	}
	return &RewardRequestMultisetObject{
		version:             defaultVersion,
		rewardRequestHash:   key,
		rewardReceiverState: newRewardRequestState,
		db:                  db,
		objectType:          RewardRequestV3ObjectType,
		deleted:             false,
	}, nil
}

func (rr RewardRequestMultisetObject) GetVersion() int {
	return rr.version
}

// setError remembers the first non-nil error it is called with.
func (rr *RewardRequestMultisetObject) SetError(err error) {
	if rr.dbErr == nil {
		rr.dbErr = err
	}
}

func (rr RewardRequestMultisetObject) GetTrie(db DatabaseAccessWarper) Trie {
	return rr.trie
}

func (rr *RewardRequestMultisetObject) SetValue(data interface{}) error {
	var newRewardRequestState = NewRewardRequestStateV3()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newRewardRequestState)
		if err != nil {
			return err
		}
	} else {
		newRewardRequestState, ok = data.(*RewardRequestMultisetState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidRewardRequestStateType, reflect.TypeOf(data))
		}
	}
	rr.rewardReceiverState = newRewardRequestState
	return nil
}

func (rr RewardRequestMultisetObject) GetValue() interface{} {
	return rr.rewardReceiverState
}

func (rr RewardRequestMultisetObject) GetValueBytes() []byte {
	data := rr.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal reward request state")
	}
	return []byte(value)
}

func (rr RewardRequestMultisetObject) GetHash() common.Hash {
	return rr.rewardRequestHash
}

func (rr RewardRequestMultisetObject) GetType() int {
	return rr.objectType
}

// MarkDelete will delete an object in trie
func (rr *RewardRequestMultisetObject) MarkDelete() {
	rr.deleted = true
}

func (rr *RewardRequestMultisetObject) Reset() bool {
	rr.rewardReceiverState = NewRewardRequestStateV3()
	return true
}

func (rr RewardRequestMultisetObject) IsDeleted() bool {
	return rr.deleted
}

// value is either default or nil
func (rr RewardRequestMultisetObject) IsEmpty() bool {
	temp := NewRewardRequestStateV3()
	return reflect.DeepEqual(temp, rr.rewardReceiverState) || rr.rewardReceiverState == nil
}

func generateRewardRequestMultisetObjectKey(epoch uint64, shardID, subsetID byte, tokenID common.Hash) common.Hash {
	prefixHash := GetRewardRequestPrefix(epoch)
	tempPrefix := append([]byte{shardID}, []byte{subsetID}...)
	valueHash := common.HashH(append(tempPrefix, tokenID[:]...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}
