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

//func AddCommitteeReward(stateDB *StateDB, paymentAddress string, amount uint64, tokenID common.Hash) error {
//
//}
