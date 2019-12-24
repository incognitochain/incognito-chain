package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

func AddShardRewardRequest(stateDB *StateDB, epoch uint64, shardID byte, rewardAmount uint64, tokenID common.Hash) error {
	key := GenerateRewardRequestObjectKey(epoch, shardID, tokenID)
	r, has, err := stateDB.GetRewardRequestState(key)
	if err != nil {
		return NewStatedbError(StoreRewardRequestError, err)
	}
	if has {
		rewardAmount += r.amount
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
		return 0, NewStatedbError(GetRewardRequestError, fmt.Errorf("token %+v amount not found", tokenID))
	}
	return amount, nil
}

func AddCommitteeReward(stateDB *StateDB, incognitoPublicKey string, committeeReward uint64, tokenID common.Hash) error {
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
		return 0, NewStatedbError(GetCommitteeRewardError, fmt.Errorf("token %+v reward not exist", tokenID))
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

func RemoveCommitteeReward(stateDB *StateDB, incognitoPublicKey string, committeeWithdraw uint64, tokenID common.Hash) error {
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
	committeeRewardM := c.reward
	currentReward := committeeRewardM[tokenID]
	if committeeWithdraw > currentReward {
		return NewStatedbError(RemoveCommitteeRewardError, fmt.Errorf("Current Reward %+v but got withdraw %+v", currentReward, committeeWithdraw))
	}
	remain := currentReward - committeeWithdraw
	committeeRewardM[tokenID] = remain
	value := NewCommitteeRewardStateWithValue(committeeRewardM, incognitoPublicKey)
	err = stateDB.SetStateObject(CommitteeRewardObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	return nil
}
