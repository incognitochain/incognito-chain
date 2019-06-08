package lvdb

import (
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) AddShardRewardRequest(
	epoch uint64,
	shardID byte,
	rewardAmount uint64,
) error {
	fmt.Printf("[ndh]-[DATABASE] AddShardRewardRequest- - - %+v %+v %+v\n", epoch, shardID, rewardAmount)
	key, err := NewKeyAddShardRewardRequest(epoch, shardID)
	if err != nil {
		return err
	}
	oldValue, err := db.Get(key)
	if err != nil {
		fmt.Printf("[ndh]-[ERROR] AddShardRewardRequest 0- - - %+v\n", err)
		err1 := db.Put(key, common.Uint64ToBytes(rewardAmount))
		fmt.Printf("[ndh]-[ERROR] AddShardRewardRequest 1- - - %+v\n", err1)
		if err1 != nil {
			return err1
		}
	} else {
		newValue := common.BytesToUint64(oldValue)
		newValue += rewardAmount
		err = db.Put(key, common.Uint64ToBytes(newValue))
		fmt.Printf("[ndh]-[ERROR] AddShardRewardRequest 2- - - %+v\n", err)
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
	fmt.Printf("[ndh]-[DATABASE] GetRewardOfShardByEpoch- - - %+v %+v\n", epoch, shardID)
	key, _ := NewKeyAddShardRewardRequest(epoch, shardID)
	rewardAmount, err := db.Get(key)
	if err != nil {
		fmt.Printf("[ndh]-[ERROR] 1 --- %+v\n", err)
		return 0, nil
	}
	fmt.Printf("[ndh] - - - %+v\n", rewardAmount)
	return common.BytesToUint64(rewardAmount), nil
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

func (db *db) GetCommitteeReward(committeeAddress []byte) (uint64, error) {
	key, err := NewKeyAddCommitteeReward(committeeAddress)
	if err != nil {
		return 0, err
	}
	value, isExist := db.Get(key)
	if isExist != nil {
		return 0, nil
	}
	return common.BytesToUint64(value), nil
}

func (db *db) RemoveCommitteeReward(committeeAddress []byte, amount uint64) error {
	key, err := NewKeyAddCommitteeReward(committeeAddress)
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
		if amount > newValue {
			return errors.New("Not enough reward to remove")
		}
		newValue -= amount
		err := db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *db) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return db.lvdb.NewIterator(slice, ro)
}

func ViewDBByPrefix(db database.DatabaseInterface, prefix []byte) map[string]string {
	begin := prefix
	// +1 to search in that range
	end := common.BytesPlusOne(prefix)

	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}
	iter := db.NewIterator(&searchRange, nil)
	res := make(map[string]string)
	for iter.Next() {
		res[string(iter.Key())] = string(iter.Value())
	}
	return res
}

func ViewDetailDBByPrefix(db database.DatabaseInterface, prefix []byte) map[string][]byte {
	begin := prefix
	// +1 to search in that range
	end := common.BytesPlusOne(prefix)

	searchRange := util.Range{
		Start: begin,
		Limit: end,
	}
	iter := db.NewIterator(&searchRange, nil)
	res := make(map[string][]byte)
	for iter.Next() {
		res[string(iter.Key())] = iter.Key()
	}
	return res
}
