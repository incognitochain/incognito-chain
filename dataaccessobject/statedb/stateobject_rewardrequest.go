package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type RewardRequestState struct {
	// tokenid => amount
	epoch   uint64
	shardID byte
	tokenID common.Hash
	amount  uint64
}

func (rr RewardRequestState) Amount() uint64 {
	return rr.amount
}

func (rr *RewardRequestState) SetAmount(amount uint64) {
	rr.amount = amount
}

func (rr RewardRequestState) TokenID() common.Hash {
	return rr.tokenID
}

func (rr *RewardRequestState) SetTokenID(tokenID common.Hash) {
	rr.tokenID = tokenID
}

func (rr RewardRequestState) ShardID() byte {
	return rr.shardID
}

func (rr *RewardRequestState) SetShardID(shardID byte) {
	rr.shardID = shardID
}

func (rr RewardRequestState) Epoch() uint64 {
	return rr.epoch
}

func (rr *RewardRequestState) SetEpoch(epoch uint64) {
	rr.epoch = epoch
}

func NewRewardRequestState() *RewardRequestState {
	return &RewardRequestState{}
}

func NewRewardRequestStateWithValue(epoch uint64, shardID byte, tokenID common.Hash, amount uint64) *RewardRequestState {
	return &RewardRequestState{epoch: epoch, shardID: shardID, tokenID: tokenID, amount: amount}
}

func (c RewardRequestState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Epoch   uint64
		ShardID byte
		TokenID common.Hash
		Amount  uint64
	}{
		Epoch:   c.epoch,
		ShardID: c.shardID,
		TokenID: c.tokenID,
		Amount:  c.amount,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *RewardRequestState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Epoch   uint64
		ShardID byte
		TokenID common.Hash
		Amount  uint64
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.epoch = temp.Epoch
	c.shardID = temp.ShardID
	c.tokenID = temp.TokenID
	c.amount = temp.Amount
	return nil
}

type RewardRequestObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version             int
	rewardRequestHash   common.Hash
	rewardReceiverState *RewardRequestState
	objectType          int
	deleted             bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newRewardRequestObject(db *StateDB, hash common.Hash) *RewardRequestObject {
	return &RewardRequestObject{
		version:             defaultVersion,
		db:                  db,
		rewardRequestHash:   hash,
		rewardReceiverState: NewRewardRequestState(),
		objectType:          RewardRequestObjectType,
		deleted:             false,
	}
}

func newRewardRequestObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*RewardRequestObject, error) {
	var newRewardRequestState = NewRewardRequestState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newRewardRequestState)
		if err != nil {
			return nil, err
		}
	} else {
		newRewardRequestState, ok = data.(*RewardRequestState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidRewardRequestStateType, reflect.TypeOf(data))
		}
	}
	return &RewardRequestObject{
		version:             defaultVersion,
		rewardRequestHash:   key,
		rewardReceiverState: newRewardRequestState,
		db:                  db,
		objectType:          RewardRequestObjectType,
		deleted:             false,
	}, nil
}

func GenerateRewardRequestObjectKey(epoch uint64, shardID byte, tokenID common.Hash) common.Hash {
	prefixHash := GetRewardRequestPrefix(epoch)
	valueHash := common.HashH(append([]byte{shardID}, tokenID[:]...))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (rr RewardRequestObject) GetVersion() int {
	return rr.version
}

// setError remembers the first non-nil error it is called with.
func (rr *RewardRequestObject) SetError(err error) {
	if rr.dbErr == nil {
		rr.dbErr = err
	}
}

func (rr RewardRequestObject) GetTrie(db DatabaseAccessWarper) Trie {
	return rr.trie
}

func (rr *RewardRequestObject) SetValue(data interface{}) error {
	var newRewardRequestState = NewRewardRequestState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newRewardRequestState)
		if err != nil {
			return err
		}
	} else {
		newRewardRequestState, ok = data.(*RewardRequestState)
		if !ok {
			return fmt.Errorf("%+v, got type %+v", ErrInvalidRewardRequestStateType, reflect.TypeOf(data))
		}
	}
	rr.rewardReceiverState = newRewardRequestState
	return nil
}

func (rr RewardRequestObject) GetValue() interface{} {
	return rr.rewardReceiverState
}

func (rr RewardRequestObject) GetValueBytes() []byte {
	data := rr.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal reward request state")
	}
	return []byte(value)
}

func (rr RewardRequestObject) GetHash() common.Hash {
	return rr.rewardRequestHash
}

func (rr RewardRequestObject) GetType() int {
	return rr.objectType
}

// MarkDelete will delete an object in trie
func (rr *RewardRequestObject) MarkDelete() {
	rr.deleted = true
}

func (rr *RewardRequestObject) Reset() bool {
	rr.rewardReceiverState = NewRewardRequestState()
	return true
}

func (rr RewardRequestObject) IsDeleted() bool {
	return rr.deleted
}

// value is either default or nil
func (rr RewardRequestObject) IsEmpty() bool {
	temp := NewRewardRequestState()
	return reflect.DeepEqual(temp, rr.rewardReceiverState) || rr.rewardReceiverState == nil
}
