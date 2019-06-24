package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
)

func (db *db) AddShardRewardRequest(epoch uint64, shardID byte, rewardAmount uint64, tokenID common.Hash) error {
	fmt.Printf("[ndh]-[DATABASE] AddShardRewardRequest- - - %+v %+v %+v %+v\n", epoch, shardID, rewardAmount, tokenID)
	key, err := NewKeyAddShardRewardRequest(epoch, shardID, tokenID)
	if err != nil {
		return err
	}
	oldValue, err := db.Get(key)
	if err != nil {
		err1 := db.Put(key, common.Uint64ToBytes(rewardAmount))
		//fmt.Printf("[ndh]-[ERROR] AddShardRewardRequest 1- - - %+v\n", err1)
		if err1 != nil {
			return err1
		}
	} else {
		newValue := common.BytesToUint64(oldValue)
		newValue += rewardAmount
		err = db.Put(key, common.Uint64ToBytes(newValue))
		//fmt.Printf("[ndh]-[ERROR] AddShardRewardRequest 2- - - %+v\n", err)
	}
	return nil
}

func (db *db) GetRewardOfShardByEpoch(epoch uint64, shardID byte, tokenID common.Hash) (uint64, error) {
	fmt.Printf("[ndh]-[DATABASE] GetRewardOfShardByEpoch- - - %+v %+v %+v\n", epoch, shardID, tokenID)
	key, _ := NewKeyAddShardRewardRequest(epoch, shardID, tokenID)
	rewardAmount, err := db.Get(key)
	if err != nil {
		//fmt.Printf("[ndh]-[ERROR] 1 --- %+v\n", err)
		return 0, nil
	}
	//fmt.Printf("[ndh] - - - %+v\n", rewardAmount)
	return common.BytesToUint64(rewardAmount), nil
}

func (db *db) AddCommitteeReward(committeeAddress []byte, amount uint64, tokenID common.Hash) error {
	key, err := NewKeyAddCommitteeReward(committeeAddress, tokenID)
	if err != nil {
		return err
	}
	oldValue, isExist := db.Get(key)
	if isExist != nil {
		err := db.Put(key, common.Uint64ToBytes(amount))
		if err != nil {
			return err
		}
	} else {
		newValue := common.BytesToUint64(oldValue)
		newValue += amount
		err := db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *db) GetCommitteeReward(committeeAddress []byte, tokenID common.Hash) (uint64, error) {
	key, err := NewKeyAddCommitteeReward(committeeAddress, tokenID)
	if err != nil {
		return 0, err
	}
	value, isExist := db.Get(key)
	if isExist != nil {
		return 0, nil
	}
	return common.BytesToUint64(value), nil
}

func (db *db) RemoveCommitteeReward(committeeAddress []byte, amount uint64, tokenID common.Hash) error {
	key, err := NewKeyAddCommitteeReward(committeeAddress, tokenID)
	if err != nil {
		return err
	}
	oldValue, isExist := db.Get(key)
	if isExist == nil {
		newValue := common.BytesToUint64(oldValue)
		if amount < newValue {
			newValue -= amount
		} else {
			newValue = 0
		}
		err := db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return err
		}
	}
	return nil
}
