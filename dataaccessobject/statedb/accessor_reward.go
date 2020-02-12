package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incdb"
	"strings"
)

// Reward in Beacon
func AddShardRewardRequest(stateDB *StateDB, epoch uint64, shardID byte, tokenID common.Hash, rewardAmount uint64) error {
	key := GenerateRewardRequestObjectKey(epoch, shardID, tokenID)
	r, has, err := stateDB.GetRewardRequestState(key)
	if err != nil {
		return NewStatedbError(StoreRewardRequestError, err)
	}
	if has {
		rewardAmount += r.Amount()
	}
	value := NewRewardRequestStateWithValue(epoch, shardID, tokenID, rewardAmount)
	err = stateDB.SetStateObject(RewardRequestObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreRewardRequestError, err)
	}
	return nil
}

func GetRewardOfShardByEpoch(stateDB *StateDB, epoch uint64, shardID byte, tokenID common.Hash) (uint64, error) {
	key := GenerateRewardRequestObjectKey(epoch, shardID, tokenID)
	amount, has, err := stateDB.GetRewardRequestAmount(key)
	if err != nil {
		return 0, NewStatedbError(GetRewardRequestError, err)
	}
	if !has {
		return 0, nil
	}
	return amount, nil
}

func GetAllTokenIDForReward(stateDB *StateDB, epoch uint64) []common.Hash {
	_, rewardRequestStates := stateDB.GetAllRewardRequestState(epoch)
	tokenIDs := []common.Hash{}
	for _, rewardRequestState := range rewardRequestStates {
		tokenIDs = append(tokenIDs, rewardRequestState.TokenID())
	}
	return tokenIDs
}

func RemoveRewardOfShardByEpoch(stateDB *StateDB, epoch uint64) {
	rewardRequestKeys, _ := stateDB.GetAllRewardRequestState(epoch)
	for _, k := range rewardRequestKeys {
		stateDB.MarkDeleteStateObject(RewardRequestObjectType, k)
	}
}

// Reward in Shard
func AddCommitteeReward(stateDB *StateDB, incognitoPublicKey string, committeeReward uint64, tokenID common.Hash, db incdb.Database, height uint64) error {
	key, err := GenerateCommitteeRewardObjectKey(incognitoPublicKey)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	c, has, err := stateDB.GetCommitteeRewardState(key)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	committeeRewardM := make(map[common.Hash]uint64)
	if has {
		committeeRewardM = c.Reward()
	}
	amount, ok := committeeRewardM[tokenID]
	if ok {
		committeeReward += amount
	}
	committeeRewardM[tokenID] = committeeReward
	value := NewCommitteeRewardStateWithValue(committeeRewardM, incognitoPublicKey)
	err = stateDB.SetStateObject(CommitteeRewardObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	publicKeyBytes, _, _ := base58.Base58Check{}.Decode(incognitoPublicKey)
	err = AddTestCommitteeReward(db, height, publicKeyBytes, committeeReward, tokenID)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	return nil
}

func GetCommitteeReward(stateDB *StateDB, incognitoPublicKey string, tokenID common.Hash) (uint64, error) {
	key, err := GenerateCommitteeRewardObjectKey(incognitoPublicKey)
	if err != nil {
		return 0, NewStatedbError(GetCommitteeRewardError, err)
	}
	r, has, err := stateDB.GetCommitteeRewardAmount(key)
	if err != nil {
		return 0, NewStatedbError(GetCommitteeRewardError, err)
	}
	if !has {
		return 0, nil
	}
	if amount, ok := r[tokenID]; !ok {
		return 0, nil
	} else {
		return amount, nil
	}
}

func ListCommitteeReward(stateDB *StateDB) map[string]map[common.Hash]uint64 {
	return stateDB.GetAllCommitteeReward()
}

func RemoveCommitteeReward(stateDB *StateDB, incognitoPublicKeyBytes []byte, withdrawAmount uint64, tokenID common.Hash) error {
	incognitoPublicKey := base58.Base58Check{}.Encode(incognitoPublicKeyBytes, common.Base58Version)
	key, err := GenerateCommitteeRewardObjectKey(incognitoPublicKey)
	if err != nil {
		return NewStatedbError(RemoveCommitteeRewardError, err)
	}
	c, has, err := stateDB.GetCommitteeRewardState(key)
	if err != nil {
		return NewStatedbError(RemoveCommitteeRewardError, err)
	}
	if !has {
		return nil
	}
	committeeRewardM := c.Reward()
	currentReward := committeeRewardM[tokenID]
	if withdrawAmount > currentReward {
		return NewStatedbError(RemoveCommitteeRewardError, fmt.Errorf("Current Reward %+v but got withdraw %+v", currentReward, withdrawAmount))
	}
	remain := currentReward - withdrawAmount
	if remain == 0 {
		delete(committeeRewardM, tokenID)
	} else {
		committeeRewardM[tokenID] = remain
	}
	if len(committeeRewardM) == 0 {
		stateDB.MarkDeleteStateObject(CommitteeRewardObjectType, key)
		return nil
	}
	value := NewCommitteeRewardStateWithValue(committeeRewardM, incognitoPublicKey)
	err = stateDB.SetStateObject(CommitteeRewardObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	return nil
}

//================================= Testing ======================================
func GetRewardRequestInfoByEpoch(stateDB *StateDB, epoch uint64) []*RewardRequestState {
	_, rewardRequestStates := stateDB.GetAllRewardRequestState(epoch)
	return rewardRequestStates
}

// ======================================================== TESTING
var testCommitteeRewardPrefix = []byte("test-committee-reward")
var Splitter = []byte("-[-]-")

func addTestCommitteeRewardKey(height uint64, committeeAddress []byte, tokenID common.Hash) []byte {
	res := []byte{}
	res = append(res, testCommitteeRewardPrefix...)
	res = append(res, Splitter...)
	res = append(res, common.Uint64ToBytes(height)...)
	res = append(res, Splitter...)
	res = append(res, committeeAddress...)
	res = append(res, Splitter...)
	res = append(res, tokenID.GetBytes()...)
	return res
}

func AddTestCommitteeReward(
	db incdb.Database,
	height uint64,
	committeeAddress []byte,
	amount uint64,
	tokenID common.Hash,
) error {
	key := addTestCommitteeRewardKey(height, committeeAddress, tokenID)
	oldValue, isExist := db.Get(key)
	if isExist != nil {
		err := db.Put(key, common.Uint64ToBytes(amount))
		if err != nil {
			return err
		}
	} else {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return err
		}
		newValue += amount
		err = db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return err
		}
	}
	return nil
}

// ListCommitteeReward - get reward on tokenID of all committee
func ListTestCommitteeReward(db incdb.Database) map[string]map[uint64]map[common.Hash]uint64 {
	result := make(map[string]map[uint64]map[common.Hash]uint64)
	iterator := db.NewIteratorWithPrefix(testCommitteeRewardPrefix)
	for iterator.Next() {
		key := make([]byte, len(iterator.Key()))
		copy(key, iterator.Key())
		value := make([]byte, len(iterator.Value()))
		copy(value, iterator.Value())
		reses := strings.Split(string(key), string(Splitter))
		reward, _ := common.BytesToUint64(value)
		publicKeyInByte := []byte(reses[2])
		publicKeyInBase58Check := base58.Base58Check{}.Encode(publicKeyInByte, 0x0)
		heightBytes := []byte(reses[1])
		height, _ := common.BytesToUint64(heightBytes)
		tokenIDBytes := []byte(reses[3])
		tokenID, _ := common.Hash{}.NewHash(tokenIDBytes)
		rewardByHeight, ok := result[publicKeyInBase58Check]
		if !ok {
			rewardByHeight = make(map[uint64]map[common.Hash]uint64)
		}
		rewards, ok := rewardByHeight[height]
		if !ok {
			rewards = make(map[common.Hash]uint64)
		}
		rewards[*tokenID] = reward
		rewardByHeight[height] = rewards
		result[publicKeyInBase58Check] = rewardByHeight
	}
	return result
}
