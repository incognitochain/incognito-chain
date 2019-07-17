package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/**
 * AddShardRewardRequest save the amount of rewards for a shard X at epoch T.
 * @param epoch: epoch T
 * @param shardID: shard X
 * @param rewardAmount: the amount of rewards
 * @param tokenID: currency unit
 * @return error
 */
func (db *db) AddShardRewardRequest(
	epoch uint64,
	shardID byte,
	rewardAmount uint64,
	tokenID common.Hash,
) error {
	key, err := NewKeyAddShardRewardRequest(epoch, shardID, tokenID)
	if err != nil {
		return err
	}
	oldValue, err := db.Get(key)
	if err != nil {
		err1 := db.Put(key, common.Uint64ToBytes(rewardAmount))
		////fmt.Printf("[ndh]-[ERROR] AddShardRewardRequest 1- - - %+v\n", err1)
		if err1 != nil {
			return err1
		}
	} else {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return err
		}
		newValue += rewardAmount
		err = db.Put(key, common.Uint64ToBytes(newValue))
		////fmt.Printf("[ndh]-[ERROR] AddShardRewardRequest 2- - - %+v\n", err)
	}
	return nil
}

/**
 * GetRewardOfShardByEpoch get the amount of rewards for a shard X at epoch T.
 * @param epoch: epoch T
 * @param shardID: shard X
 * @param tokenID: currency unit
 * @return (uint64, error): the amount of rewards, error of this process
 */
func (db *db) GetRewardOfShardByEpoch(
	epoch uint64,
	shardID byte,
	tokenID common.Hash,
) (uint64, error) {
	key, _ := NewKeyAddShardRewardRequest(epoch, shardID, tokenID)
	rewardAmount, err := db.Get(key)
	if err != nil {
		////fmt.Printf("[ndh]-[ERROR] 1 --- %+v\n", err)
		return 0, nil
	}
	////fmt.Printf("[ndh] - - - %+v\n", rewardAmount)
	return common.BytesToUint64(rewardAmount)
}

/**
 * AddCommitteeReward increase the amount of rewards for a person in committee P.
 * @param committeeAddress: Public key of person P
 * @param amount: the amount of rewards
 * @param tokenID: currency unit
 * @return error
 */
func (db *db) AddCommitteeReward(
	committeeAddress []byte,
	amount uint64,
	tokenID common.Hash,
) error {
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
func (db *db) ListCommitteeReward() map[string]map[common.Hash]uint64 {
	result := make(map[string]map[common.Hash]uint64)
	iterator := db.lvdb.NewIterator(util.BytesPrefix(CommitteeRewardPrefix), nil)
	for iterator.Next() {
		key := make([]byte, len(iterator.Key()))
		copy(key, iterator.Key())
		value := make([]byte, len(iterator.Value()))
		copy(value, iterator.Value())
		reward, _ := common.BytesToUint64(value)
		publicKeyInByte := key[len(CommitteeRewardPrefix) : len(CommitteeRewardPrefix)+33]
		publicKeyInBase58Check := base58.Base58Check{}.Encode(publicKeyInByte, 0x0)
		tokenIDBytes := key[len(key)-32:]
		tokenID, _ := common.Hash{}.NewHash(tokenIDBytes)
		if result[publicKeyInBase58Check] == nil {
			result[publicKeyInBase58Check] = make(map[common.Hash]uint64)
		}
		result[publicKeyInBase58Check][*tokenID] = reward
	}
	return result
}

/**
 * AddCommitteeReward get the amount of rewards for a person in committee P.
 * @param committeeAddress: Public key of person P
 * @param tokenID: currency unit
 * @return (uint64, error): the amount of rewards, error of this process
 */
func (db *db) GetCommitteeReward(
	committeeAddress []byte,
	tokenID common.Hash,
) (uint64, error) {
	key, err := NewKeyAddCommitteeReward(committeeAddress, tokenID)
	if err != nil {
		return 0, err
	}
	value, isExist := db.Get(key)
	if isExist != nil {
		return 0, nil
	}

	return common.BytesToUint64(value)
}

/**
 * RemoveCommitteeReward decrease the amount of rewards for a person in committee P.
 * @param committeeAddress: Public key of person P
 * @param amount: the amount of rewards
 * @param tokenID: currency unit
 * @return error
 */
func (db *db) RemoveCommitteeReward(
	committeeAddress []byte,
	amount uint64,
	tokenID common.Hash,
) error {
	key, err := NewKeyAddCommitteeReward(committeeAddress, tokenID)
	if err != nil {
		return err
	}
	oldValue, isExist := db.Get(key)
	if isExist == nil {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return err
		}
		if amount < newValue {
			newValue -= amount
		} else {
			newValue = 0
		}
		err = db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return err
		}
	}
	return nil
}
