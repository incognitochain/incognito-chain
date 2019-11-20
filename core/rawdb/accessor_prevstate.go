package rawdb

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/opentracing/opentracing-go/log"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func StorePrevBestState(db incdb.Database, val []byte, isBeacon bool, shardID byte) error {
	key := getPrevPrefix(isBeacon, shardID)
	if err := db.Put(key, val); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func FetchPrevBestState(db incdb.Database, isBeacon bool, shardID byte) ([]byte, error) {
	key := getPrevPrefix(isBeacon, shardID)
	beststate, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(LvdbGetError, err)
	}
	return beststate, nil
}

func CleanBackup(db incdb.Database, isBeacon bool, shardID byte) error {
	iter := db.NewIteratorWithPrefix(getPrevPrefix(isBeacon, shardID))
	for iter.Next() {
		err := db.Delete(iter.Key())
		if err != nil {
			return NewRawdbError(LvdbGetError, err)
		}
	}
	iter.Release()
	return nil
}

func BackupCommitmentsOfPublicKey(db incdb.Database, tokenID common.Hash, shardID byte, pubkey []byte) error {
	//backup keySpec3 & keySpec4
	prevkey := getPrevPrefix(false, shardID)
	key := prefixWithHashKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	keySpec3 := append(key, []byte("len")...)
	backupKeySpec3 := append(prevkey, keySpec3...)
	res, err := db.Get(keySpec3)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return NewRawdbError(LvdbGetError, err)
		}
		return nil
	}
	if err := db.Put(backupKeySpec3, res); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func RestoreCommitmentsOfPubkey(db incdb.Database, tokenID common.Hash, shardID byte, pubkey []byte, commitments [][]byte) error {
	// restore keySpec3 & keySpec4
	// delete keySpec1 & keySpec2
	prevkey := getPrevPrefix(false, shardID)
	key := prefixWithHashKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)

	var lenData uint64
	lenCommittee, err := GetCommitmentLength(db, tokenID, shardID)
	if err != nil && lenCommittee == nil {
		return NewRawdbError(GetCommitmentLengthError, err)
	}
	if lenCommittee == nil {
		lenData = 0
	} else {
		lenData = lenCommittee.Uint64()
	}
	for _, c := range commitments {
		newIndex := new(big.Int).SetUint64(lenData).Bytes()
		if lenData == 0 {
			newIndex = []byte{0}
		}
		keySpec1 := append(key, newIndex...)
		err = db.Delete(keySpec1)
		if err != nil {
			log.Error(err)
		}

		keySpec2 := append(key, c...)
		err = db.Delete(keySpec2)
		if err != nil {
			log.Error(err)
		}
		lenData++
	}

	// keySpec3 store last index of array commitment
	keySpec3 := append(key, []byte("len")...)
	backupKeySpec3 := append(prevkey, keySpec3...)
	res, err := db.Get(backupKeySpec3)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return NewRawdbError(LvdbGetError, err)
		}
		if err := db.Delete(keySpec3); err != nil {
			return NewRawdbError(LvdbDeleteError, err)
		}
	}

	if err := db.Put(keySpec3, res); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}

	return nil
}

func DeleteOutputCoin(db incdb.Database, tokenID common.Hash, publicKey []byte, outputCoinArr [][]byte, shardID byte) error {
	key := prefixWithHashKey(string(outcoinsPrefix), tokenID)
	key = append(key, shardID)

	key = append(key, publicKey...)
	for _, outputCoin := range outputCoinArr {
		keyTemp := append(key, common.HashB(outputCoin)...)
		if err := db.Delete(keyTemp); err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	}

	return nil
}

func BackupSerialNumbersLen(db incdb.Database, tokenID common.Hash, shardID byte) error {
	current := prefixWithHashKey(string(serialNumbersPrefix), tokenID)
	current = append(current, shardID)
	current = append(current, []byte("len")...)
	res, err := db.Get(current)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return NewRawdbError(LvdbGetError, err)
		}
		return nil
	}
	key := getPrevPrefix(false, shardID)
	key = append(key, current...)
	if err := db.Put(key, res); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func RestoreSerialNumber(db incdb.Database, tokenID common.Hash, shardID byte, serialNumbers [][]byte) error {
	key := prefixWithHashKey(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)
	currentLenKey := append(key, []byte("len")...)
	prevLenKey := getPrevPrefix(false, shardID)
	prevLenKey = append(prevLenKey, currentLenKey...)

	prevLen, err := db.Get(prevLenKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return NewRawdbError(LvdbGetError, err)
	}
	if err := db.Put(currentLenKey, prevLen); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	for _, s := range serialNumbers {
		keySpec1 := append(key, s...)
		err = db.Delete(keySpec1)
		if err != nil {
			return NewRawdbError(LvdbDeleteError, err)
		}
	}

	return nil
}

func DeleteTransactionIndex(db incdb.Database, txId common.Hash) error {
	key := string(transactionKeyPrefix) + txId.String()
	err := db.Delete([]byte(key))
	if err != nil {
		return NewRawdbError(LvdbDeleteError, err)
	}
	return nil

}

func DeletePrivacyToken(db incdb.Database, tokenID common.Hash) error {
	key := prefixWithHashKey(string(privacyTokenInitPrefix), tokenID)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(LvdbDeleteError, err)
	}
	return nil
}

func DeletePrivacyTokenTx(db incdb.Database, tokenID common.Hash, txIndex int32, shardID byte, blockHeight uint64) error {
	key := prefixWithHashKey(string(privacyTokenPrefix), tokenID)
	key = append(key, shardID)
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, bigNumber-blockHeight)
	key = append(key, bs...)
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(bigNumber-txIndex))
	key = append(key, bs...)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(LvdbDeleteError, err)
	}
	return nil
}

func DeletePrivacyTokenCrossShard(db incdb.Database, tokenID common.Hash) error {
	key := prefixWithHashKey(string(privacyTokenCrossShardPrefix), tokenID)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(LvdbDeleteError, err)
	}
	return nil
}

func RestoreCrossShardNextHeights(db incdb.Database, fromShard byte, toShard byte, curHeight uint64) error {
	key := append(nextCrossShardKeyPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	curHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(curHeightBytes, curHeight)
	heightKey := append(key, curHeightBytes...)
	for {
		nextHeightBytes, err := db.Get(heightKey)
		if err != nil && err != lvdberr.ErrNotFound {
			return NewRawdbError(LvdbGetError, err)
		}
		err = db.Delete(heightKey)
		if err != nil {
			return NewRawdbError(LvdbDeleteError, err)
		}

		var nextHeight uint64
		err = binary.Read(bytes.NewReader(nextHeightBytes[:8]), binary.LittleEndian, &nextHeight)
		if err != nil {
			log.Error(err)
		}

		if nextHeight == 0 {
			break
		}
		heightKey = append(key, nextHeightBytes...)
	}
	nextHeightBytes := make([]byte, 8)
	heightKey = append(key, curHeightBytes...)
	if err := db.Put(heightKey, nextHeightBytes); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func DeleteCommitteeByHeight(db incdb.Database, blkEpoch uint64) error {
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, heightPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkEpoch)
	key = append(key, buf[:]...)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(LvdbDeleteError, err)
	}
	return nil
}

func DeleteAcceptedShardToBeacon(db incdb.Database, shardID byte, shardBlkHash common.Hash) error {
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	key := append(shardToBeaconKeyPrefix, prefix...)
	if err := db.Delete(key); err != nil {
		return NewRawdbError(LvdbDeleteError, err)
	}
	return nil
}

func DeleteIncomingCrossShard(db incdb.Database, shardID byte, crossShardID byte, crossBlkHash common.Hash) error {
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if err := db.Delete(key); err != nil {
		return err
	}
	return nil
}

func BackupBridgedTokenByTokenID(db incdb.Database, tokenID common.Hash) error {
	key := append(centralizedBridgePrefix, tokenID[:]...)
	backupKey := getPrevPrefix(true, 0)
	backupKey = append(backupKey, key...)
	tokenWithAmtBytes, err := db.Get(key)
	if err != nil {
		if err := db.Put(backupKey, []byte{}); err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	} else {
		if err := db.Put(backupKey, tokenWithAmtBytes); err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	}
	return nil
}

func RestoreBridgedTokenByTokenID(db incdb.Database, tokenID common.Hash) error {
	key := append(centralizedBridgePrefix, tokenID[:]...)
	backupKey := getPrevPrefix(true, 0)
	backupKey = append(backupKey, key...)

	tokenWithAmtBytes, err := db.Get(backupKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return NewRawdbError(LvdbGetError, err)
	}

	if err := db.Put(key, tokenWithAmtBytes); err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

// REWARD

func BackupShardRewardRequest(db incdb.Database, epoch uint64, shardID byte, tokenID common.Hash) error {
	backupKey := getPrevPrefix(true, 0)
	key := addShardRewardRequestKey(epoch, shardID, tokenID)
	backupKey = append(backupKey, key...)
	curValue, err := db.Get(key)
	if err != nil {
		err := db.Put(backupKey, common.Uint64ToBytes(0))
		if err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	} else {
		err := db.Put(backupKey, curValue)
		if err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	}

	return nil
}
func BackupCommitteeReward(db incdb.Database, committeeAddress []byte, tokenID common.Hash) error {
	backupKey := getPrevPrefix(true, 0)
	key := addCommitteeRewardKey(committeeAddress, tokenID)
	backupKey = append(backupKey, key...)
	curValue, err := db.Get(key)
	if err != nil {
		err := db.Put(backupKey, common.Uint64ToBytes(0))
		if err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	} else {
		err := db.Put(backupKey, curValue)
		if err != nil {
			return NewRawdbError(LvdbPutError, err)
		}
	}

	return nil
}
func RestoreShardRewardRequest(db incdb.Database, epoch uint64, shardID byte, tokenID common.Hash) error {
	backupKey := getPrevPrefix(true, 0)
	key := addShardRewardRequestKey(epoch, shardID, tokenID)
	backupKey = append(backupKey, key...)
	bakValue, err := db.Get(backupKey)
	if err != nil {
		return NewRawdbError(LvdbGetError, err)
	}
	err = db.Put(key, bakValue)
	if err != nil {
		return NewRawdbError(LvdbPutError, err)
	}

	return nil
}
func RestoreCommitteeReward(db incdb.Database, committeeAddress []byte, tokenID common.Hash) error {
	backupKey := getPrevPrefix(true, 0)
	key := addCommitteeRewardKey(committeeAddress, tokenID)
	backupKey = append(backupKey, key...)
	bakValue, err := db.Get(backupKey)
	if err != nil {
		return NewRawdbError(LvdbGetError, err)
	}
	err = db.Put(key, bakValue)
	if err != nil {
		return NewRawdbError(LvdbPutError, err)
	}

	return nil
}
