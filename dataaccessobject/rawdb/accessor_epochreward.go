package rawdb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incdb"
)

/**
 * AddShardRewardRequest save the amount of rewards for a shard X at epoch T.
 * @param epoch: epoch T
 * @param shardID: shard X
 * @param rewardAmount: the amount of rewards
 * @param tokenID: currency unit
 * @return error
 */
func AddShardRewardRequest(
	db incdb.Database,
	epoch uint64,
	shardID byte,
	rewardAmount uint64,
	tokenID common.Hash,
	bd *[]incdb.BatchData,
) error {
	key := addShardRewardRequestKey(epoch, shardID, tokenID)
	oldValue, err := db.Get(key)
	if err != nil {
		if bd != nil {
			*bd = append(*bd, incdb.BatchData{key, common.Uint64ToBytes(rewardAmount)})
			return nil
		}
		err := db.Put(key, common.Uint64ToBytes(rewardAmount))
		if err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	} else {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return NewRawdbError(UnexpectedError, err)
		}
		newValue += rewardAmount

		if bd != nil {
			*bd = append(*bd, incdb.BatchData{key, common.Uint64ToBytes(newValue)})
			return nil
		}
		err = db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return NewRawdbError(UnexpectedError, err)
		}
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
func GetRewardOfShardByEpoch(
	db incdb.Database,
	epoch uint64,
	shardID byte,
	tokenID common.Hash,
) (uint64, error) {
	key := addShardRewardRequestKey(epoch, shardID, tokenID)
	rewardAmount, err := db.Get(key)
	if err != nil {
		return 0, nil
	}
	value, err := common.BytesToUint64(rewardAmount)
	if err != nil {
		return 0, NewRawdbError(UnexpectedError, err)
	}
	return value, nil
}

/**
 * AddCommitteeReward increase the amount of rewards for a person in committee P.
 * @param committeeAddress: Public key of person P
 * @param amount: the amount of rewards
 * @param tokenID: currency unit
 * @return error
 */
func AddCommitteeReward(
	db incdb.Database,
	committeeAddress []byte,
	amount uint64,
	tokenID common.Hash,
) error {
	key := addCommitteeRewardKey(committeeAddress, tokenID)
	oldValue, isExist := db.Get(key)
	if isExist != nil {
		err := db.Put(key, common.Uint64ToBytes(amount))
		if err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	} else {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return NewRawdbError(UnexpectedError, err)
		}
		newValue += amount
		err = db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return NewRawdbError(UnexpectedError, err)
		}
	}
	return nil
}

// ListCommitteeReward - get reward on tokenID of all committee
func ListCommitteeReward(db incdb.Database) map[string]map[common.Hash]uint64 {
	result := make(map[string]map[common.Hash]uint64)
	iterator := db.NewIteratorWithPrefix(committeeRewardPrefix)
	for iterator.Next() {
		key := make([]byte, len(iterator.Key()))
		copy(key, iterator.Key())
		value := make([]byte, len(iterator.Value()))
		copy(value, iterator.Value())
		reward, _ := common.BytesToUint64(value)
		publicKeyInByte := key[len(committeeRewardPrefix) : len(committeeRewardPrefix)+common.PublicKeySize]
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
func GetCommitteeReward(
	db incdb.Database,
	committeeAddress []byte,
	tokenID common.Hash,
) (uint64, error) {
	key := addCommitteeRewardKey(committeeAddress, tokenID)
	value, isExist := db.Get(key)
	if isExist != nil {
		return 0, nil
	}

	val, err := common.BytesToUint64(value)
	if err != nil {
		return 0, NewRawdbError(GetCommitteeRewardError, err)
	}
	return val, nil
}

/**
 * RemoveCommitteeReward decrease the amount of rewards for a person in committee P.
 * @param committeeAddress: Public key of person P
 * @param amount: the amount of rewards
 * @param tokenID: currency unit
 * @return error
 */
func RemoveCommitteeReward(
	db incdb.Database,
	committeeAddress []byte,
	amount uint64,
	tokenID common.Hash,
	bd *[]incdb.BatchData,
) error {
	key := addCommitteeRewardKey(committeeAddress, tokenID)
	oldValue, isExist := db.Get(key)
	if isExist == nil {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return NewRawdbError(RemoveCommitteeRewardError, err)
		}
		if amount < newValue {
			newValue -= amount
		} else {
			newValue = 0
		}

		if bd != nil {
			*bd = append(*bd, incdb.BatchData{key, common.Uint64ToBytes(newValue)})
			return nil
		}
		err = db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return NewRawdbError(RemoveCommitteeRewardError, err)
		}
	}
	return nil
}
