package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/**
 * NewKeyAddShardRewardRequest create a key for store reward of a shard X at epoch T in db.
 * @param epoch: epoch T
 * @param shardID: shard X
 * @param tokenID: currency unit
 * @return ([]byte, error): Key, error of this process
 */
func newKeyAddShardRewardRequest(
	epoch uint64,
	shardID byte,
	tokenID common.Hash,
) []byte {
	res := []byte{}
	res = append(res, shardRequestRewardPrefix...)
	res = append(res, common.Uint64ToBytes(epoch)...)
	res = append(res, shardID)
	res = append(res, tokenID.GetBytes()...)
	return res
}

/**
 * NewKeyAddCommitteeReward create a key for store reward of a person P in committee in db.
 * @param committeeAddress: Public key of person P
 * @param tokenID: currency unit
 * @return ([]byte, error): Key, error of this process
 */
func newKeyAddCommitteeReward(
	committeeAddress []byte,
	tokenID common.Hash,
) []byte {
	res := []byte{}
	res = append(res, committeeRewardPrefix...)
	res = append(res, committeeAddress...)
	res = append(res, tokenID.GetBytes()...)
	return res
}

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
	bd *[]database.BatchData,
) error {
	key := newKeyAddShardRewardRequest(epoch, shardID, tokenID)
	oldValue, err := db.Get(key)
	if err != nil {
		if bd != nil {
			*bd = append(*bd, database.BatchData{key, common.Uint64ToBytes(rewardAmount)})
			return nil
		}
		err1 := db.Put(key, common.Uint64ToBytes(rewardAmount))
		if err1 != nil {
			return database.NewDatabaseError(database.UnexpectedError, err1)
		}
	} else {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, err)
		}
		newValue += rewardAmount

		if bd != nil {
			*bd = append(*bd, database.BatchData{key, common.Uint64ToBytes(newValue)})
			return nil
		}
		err = db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, err)
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
func (db *db) GetRewardOfShardByEpoch(
	epoch uint64,
	shardID byte,
	tokenID common.Hash,
) (uint64, error) {
	key := newKeyAddShardRewardRequest(epoch, shardID, tokenID)
	rewardAmount, err := db.Get(key)
	if err != nil {
		return 0, nil
	}
	value, err := common.BytesToUint64(rewardAmount)
	if err != nil {
		return 0, database.NewDatabaseError(database.UnexpectedError, err)
	}
	return value, nil
}

func (db *db) GetAllTokenIDForReward(
	epoch uint64,
) (
	map[common.Hash]struct{},
	error,
) {
	keyForSearch := []byte{}
	keyForSearch = append(keyForSearch, shardRequestRewardPrefix...)
	keyForSearch = append(keyForSearch, common.Uint64ToBytes(epoch)...)
	result := map[common.Hash]struct{}{}
	iterator := db.lvdb.NewIterator(util.BytesPrefix(keyForSearch), nil)
	for iterator.Next() {
		key := make([]byte, len(iterator.Key()))
		copy(key, iterator.Key())
		value := make([]byte, len(iterator.Value()))
		copy(value, iterator.Value())
		tokenIDBytes := key[len(key)-32:]
		tokenID, err := common.Hash{}.NewHash(tokenIDBytes)
		if err != nil {
			database.Logger.Log.Errorf("Pasre token ID %v return error %v", tokenIDBytes, err)
			continue
		}
		result[*tokenID] = struct{}{}
	}
	return result, nil
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
	key := newKeyAddCommitteeReward(committeeAddress, tokenID)
	oldValue, isExist := db.Get(key)
	if isExist != nil {
		err := db.Put(key, common.Uint64ToBytes(amount))
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, err)
		}
	} else {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, err)
		}
		newValue += amount
		err = db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, err)
		}
	}
	return nil
}

// ListCommitteeReward - get reward on tokenID of all committee
func (db *db) ListCommitteeReward() map[string]map[common.Hash]uint64 {
	result := make(map[string]map[common.Hash]uint64)
	iterator := db.lvdb.NewIterator(util.BytesPrefix(committeeRewardPrefix), nil)
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
func (db *db) GetCommitteeReward(
	committeeAddress []byte,
	tokenID common.Hash,
) (uint64, error) {
	key := newKeyAddCommitteeReward(committeeAddress, tokenID)
	value, isExist := db.Get(key)
	if isExist != nil {
		return 0, nil
	}

	val, err := common.BytesToUint64(value)
	if err != nil {
		return 0, database.NewDatabaseError(database.GetCommitteeRewardError, err)
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
func (db *db) RemoveCommitteeReward(
	committeeAddress []byte,
	amount uint64,
	tokenID common.Hash,
	bd *[]database.BatchData,
) error {
	key := newKeyAddCommitteeReward(committeeAddress, tokenID)
	oldValue, isExist := db.Get(key)
	if isExist == nil {
		newValue, err := common.BytesToUint64(oldValue)
		if err != nil {
			return database.NewDatabaseError(database.RemoveCommitteeRewardError, err)
		}
		if amount < newValue {
			newValue -= amount
		} else {
			newValue = 0
		}

		if bd != nil {
			*bd = append(*bd, database.BatchData{key, common.Uint64ToBytes(newValue)})
			return nil
		}
		err = db.Put(key, common.Uint64ToBytes(newValue))
		if err != nil {
			return database.NewDatabaseError(database.RemoveCommitteeRewardError, err)
		}
	}
	return nil
}
