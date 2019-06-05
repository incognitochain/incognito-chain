package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/pkg/errors"
)

func (db *db) AddShardRewardRequest(
	epoch uint64,
	shardID byte,
	rewardAmount uint64,
) error {
	key, err := NewKeyAddShardRewardRequest(epoch, shardID)
	if err != nil {
		return err
	}
	oldValue, err := db.Get(key)
	if err != nil {
		err1 := db.Put(key, common.Uint64ToBytes(rewardAmount))
		if err1 != nil {
			return err1
		}
	} else {
		newValue := common.BytesToUint64(oldValue)
		newValue += rewardAmount
		db.Put(key, common.Uint64ToBytes(newValue))
	}
	return nil
}

func (db *db) AddShardCommitteeReward(reward uint64, shardPaymentAddress []byte) error {
	return nil
}

func (db *db) AddBeaconCommitteeReward(reward uint64, beaconPaymentAddress []byte) error {
	return nil
}

func (db *db) AddDevReward(reward uint64) error {
	return nil
}

func (db *db) GetRewardOfShardByEpoch(epoch uint64, shardID byte) (uint64, error) {
	key, err := NewKeyAddShardRewardRequest(epoch, shardID)
	if err != nil {
		return 0, err
	}
	rewardAmount, err := db.Get(key)
	return common.BytesToUint64(rewardAmount), err
}

func (db *db) AddBeaconBlockProposer(
	epoch uint64,
	beaconPaymentAddress []byte,
	beaconBlockHeight uint64,
) error {
	return nil
}

func (db *db) AddCommitteeReward(committeeAddress []byte, amount uint64) error {
	key, err := NewKeyAddCommitteeReward(committeeAddress)
	if err != nil {
		return err
	}
	oldValue, isExist := db.Get(key)
	if isExist != nil {
		db.Put(key, common.Uint64ToBytes(amount))
	} else {
		newValue := common.BytesToUint64(oldValue)
		newValue += amount
		db.Put(key, common.Uint64ToBytes(newValue))
	}
	return nil
}

func (db *db) GetCommitteeReward(committeeAddress []byte) (uint64, error) {
	key, err := NewKeyAddCommitteeReward(committeeAddress)
	if err != nil {
		return 0, err
	}
	value, isExist := db.Get(key)
	if isExist != nil {
		return 0, isExist
	}
	return common.BytesToUint64(value), isExist
}

func (db *db) RemoveCommitteeReward(committeeAddress []byte, amount uint64) error {
	key, err := NewKeyAddCommitteeReward(committeeAddress)
	if err != nil {
		return err
	}
	oldValue, isExist := db.Get(key)
	if isExist != nil {
		db.Put(key, common.Uint64ToBytes(amount))
	} else {
		newValue := common.BytesToUint64(oldValue)
		if amount > newValue {
			return errors.New("Not enough reward to remove")
		}
		newValue -= amount
		db.Put(key, common.Uint64ToBytes(newValue))
	}
	return nil
}
