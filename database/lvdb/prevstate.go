package lvdb

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func getPrevPrefix(isBeacon bool, shardID byte) []byte {
	key := []byte{}
	if isBeacon {
		key = append(key, prevBeaconPrefix...)
	} else {
		key = append(key, append(prevShardPrefix, append([]byte{shardID}, byte('-'))...)...)
	}
	return key
}

func (db *db) StorePrevBestState(val []byte, isBeacon bool, shardID byte) error {
	key := getPrevPrefix(isBeacon, shardID)
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	return nil
}

func (db *db) FetchPrevBestState(isBeacon bool, shardID byte) ([]byte, error) {
	key := getPrevPrefix(isBeacon, shardID)
	beststate, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return beststate, nil
}

func (db *db) CleanBackup(isBeacon bool, shardID byte) error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(getPrevPrefix(isBeacon, shardID)), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
		}
	}
	iter.Release()
	return nil
}

func (db *db) BackupCommitmentsOfPubkey(tokenID common.Hash, shardID byte, pubkey []byte) error {
	current := db.GetKey(string(commitmentsPrefix), tokenID)
	current = append(current, shardID)

	keySpec4 := append(current, pubkey...)
	resByPubkey, err := db.lvdb.Get(keySpec4, nil)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		return nil
	}
	if len(resByPubkey) > 0 {
		key := getPrevPrefix(false, shardID)
		key = append(key, keySpec4...)
		if err := db.lvdb.Put(key, resByPubkey, nil); err != nil {
			return err
		}
	}
	return nil
}
func (db *db) RestoreCommitmentsOfPubkey(tokenID common.Hash, shardID byte, pubkey []byte, commitments []byte) error {
	current := db.GetKey(string(commitmentsPrefix), tokenID)
	current = append(current, shardID)

	keySpec2 := append(current, commitments...)
	err := db.Delete(keySpec2)
	if err != nil {
		return database.NewDatabaseError(database.ErrUnexpected, err)
	}

	keySpec4 := append(current, pubkey...)
	prevkey := getPrevPrefix(false, shardID)
	prevkey = append(prevkey, keySpec4...)

	resByPubkey, err := db.lvdb.Get(prevkey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if len(resByPubkey) > 0 {
		if err := db.lvdb.Put(keySpec4, resByPubkey, nil); err != nil {
			return err
		}
	} else {
		if err := db.Delete(keySpec4); err != nil {
			return err
		}
	}

	return nil
}

func (db *db) BackupCommitments(tokenID common.Hash, shardID byte) error {
	current := db.GetKey(string(commitmentsPrefix), tokenID)
	current = append(current, shardID)
	key := getPrevPrefix(false, shardID)
	key = append(key, current...)

	res, err := db.lvdb.Get(current, nil)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		return nil
	}

	if err := db.Put(key, res); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	//bk
	currentkeySpec3 := append(current, []byte("len")...)
	keySpec3 := getPrevPrefix(false, shardID)
	keySpec3 = append(keySpec3, currentkeySpec3...)
	res, err = db.lvdb.Get(currentkeySpec3, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if err := db.lvdb.Put(keySpec3, res, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) DeleteCommitmentsIndex(tokenID common.Hash, shardID byte) error {
	current := db.GetKey(string(commitmentsPrefix), tokenID)
	current = append(current, shardID)

	currentkeySpec3 := append(current, []byte("len")...)
	keySpec3 := getPrevPrefix(false, shardID)
	keySpec3 = append(keySpec3, currentkeySpec3...)
	currentLen, err := db.lvdb.Get(currentkeySpec3, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	prevLen, err := db.lvdb.Get(keySpec3, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	currentLenInt := new(big.Int).SetBytes(currentLen)
	prevLenInt := new(big.Int).SetBytes(prevLen)
	for index := prevLenInt; index.Int64() <= currentLenInt.Int64(); index.Add(index, big.NewInt(1)) {
		keySpec1 := append(current, index.Bytes()...)
		err = db.Delete(keySpec1)
		if err != nil {
			return database.NewDatabaseError(database.ErrUnexpected, err)
		}
	}

	return nil
}

func (db *db) RestoreCommitments(tokenID common.Hash, shardID byte) error {
	current := db.GetKey(string(commitmentsPrefix), tokenID)
	current = append(current, shardID)
	prevkey := getPrevPrefix(false, shardID)
	prevkey = append(prevkey, current...)

	res, err := db.lvdb.Get(prevkey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	if err := db.Put(current, res); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}

	currentkeySpec3 := append(current, []byte("len")...)
	keySpec3 := getPrevPrefix(false, shardID)
	keySpec3 = append(keySpec3, currentkeySpec3...)
	res, err = db.lvdb.Get(keySpec3, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if err := db.lvdb.Put(currentkeySpec3, res, nil); err != nil {
		return err
	}

	return nil
}

func (db *db) BackupOutputCoin(tokenID common.Hash, pubkey []byte, shardID byte) error {
	current := db.GetKey(string(outcoinsPrefix), tokenID)
	current = append(current, shardID)
	current = append(current, pubkey...)
	resByPubkey, err := db.lvdb.Get(current, nil)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		return nil
	}
	if len(resByPubkey) > 0 {
		key := getPrevPrefix(false, shardID)
		key = append(key, current...)
		if err := db.lvdb.Put(key, resByPubkey, nil); err != nil {
			return err
		}
	}
	return nil
}

func (db *db) RestoreOutputCoin(tokenID common.Hash, pubkey []byte, shardID byte) error {
	current := db.GetKey(string(outcoinsPrefix), tokenID)
	current = append(current, shardID)
	current = append(current, pubkey...)
	prevkey := getPrevPrefix(false, shardID)
	prevkey = append(prevkey, current...)

	resByPubkey, err := db.lvdb.Get(prevkey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if len(resByPubkey) > 0 {
		if err := db.lvdb.Put(current, resByPubkey, nil); err != nil {
			return err
		}
	} else {
		err = db.Delete(current)
		if err != nil {
			return database.NewDatabaseError(database.ErrUnexpected, err)
		}
	}
	return nil
}

// func (db *db) BackupSNDerivators(tokenID common.Hash, shardID byte) error {
// 	key := getPrevPrefix(false, shardID)

// 	return nil
// }

// func (db *db) RestoreSNDerivators(tokenID common.Hash, shardID byte) error {
// 	key := getPrevPrefix(false, shardID)

// 	return nil
// }

func (db *db) BackupSerialNumber(tokenID common.Hash, shardID byte) error {
	current := db.GetKey(string(serialNumbersPrefix), tokenID)
	current = append(current, shardID)
	res, err := db.lvdb.Get(current, nil)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		return nil
	}
	key := getPrevPrefix(false, shardID)
	key = append(key, current...)
	if err := db.lvdb.Put(key, res, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) RestoreSerialNumber(tokenID common.Hash, shardID byte) error {
	current := db.GetKey(string(serialNumbersPrefix), tokenID)
	current = append(current, shardID)
	prevkey := getPrevPrefix(false, shardID)
	prevkey = append(prevkey, current...)
	res, err := db.lvdb.Get(prevkey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if err := db.lvdb.Put(current, res, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) DeleteSerialNumber(tokenID common.Hash, serialNumber []byte, shardID byte) error {
	key := db.GetKey(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)

	err := db.Delete(key)
	if err != nil {
		return err
	}

	keySpec1 := append(key, serialNumber...)
	err = db.Delete(keySpec1)
	if err != nil {
		return err
	}

	return nil
}

func (db *db) DeleteTransactionIndex(txId common.Hash) error {
	key := string(transactionKeyPrefix) + txId.String()
	err := db.Delete([]byte(key))
	if err != nil {
		return database.NewDatabaseError(database.ErrUnexpected, err)
	}
	return nil

}

func (db *db) DeleteCustomToken(tokenID common.Hash) error {
	key := db.GetKey(string(tokenInitPrefix), tokenID)
	err := db.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (db *db) DeleteCustomTokenTx(tokenID common.Hash, txIndex int32, shardID byte, blockHeight uint64) error {
	key := db.GetKey(string(TokenPrefix), tokenID)
	key = append(key, shardID)
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, bigNumber-blockHeight)
	key = append(key, bs...)
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(bigNumber-txIndex))
	key = append(key, bs...)
	err := db.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (db *db) DeletePrivacyCustomToken(tokenID common.Hash) error {
	key := db.GetKey(string(privacyTokenInitPrefix), tokenID)
	err := db.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (db *db) DeletePrivacyCustomTokenTx(tokenID common.Hash, txIndex int32, shardID byte, blockHeight uint64) error {
	key := db.GetKey(string(PrivacyTokenPrefix), tokenID)
	key = append(key, shardID)
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, bigNumber-blockHeight)
	key = append(key, bs...)
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(bigNumber-txIndex))
	key = append(key, bs...)
	err := db.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (db *db) DeletePrivacyCustomTokenCrossShard(tokenID common.Hash) error {
	key := db.GetKey(string(PrivacyTokenCrossShardPrefix), tokenID)
	err := db.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (db *db) RestoreCrossShardNextHeights(fromShard byte, toShard byte, curHeight uint64) error {
	key := append(nextCrossShardKeyPrefix, fromShard)
	key = append(key, []byte("-")...)
	key = append(key, toShard)
	key = append(key, []byte("-")...)
	curHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(curHeightBytes, curHeight)
	heightKey := append(key, curHeightBytes...)
	for {
		nextHeightBytes, err := db.lvdb.Get(heightKey, nil)
		if err != nil && err != lvdberr.ErrNotFound {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		err = db.Delete(heightKey)
		if err != nil {
			return err
		}

		var nextHeight uint64
		binary.Read(bytes.NewReader(nextHeightBytes[:8]), binary.LittleEndian, &nextHeight)

		if nextHeight == 0 {
			break
		}
		heightKey = append(key, nextHeightBytes...)
	}
	nextHeightBytes := make([]byte, 8)
	heightKey = append(key, curHeightBytes...)
	if err := db.lvdb.Put(heightKey, nextHeightBytes, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) DeleteCommitteeByEpoch(blkEpoch uint64) error {
	key := append(beaconPrefix, shardIDPrefix...)
	key = append(key, committeePrefix...)
	key = append(key, epochPrefix...)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, blkEpoch)
	key = append(key, buf[:]...)
	err := db.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (db *db) DeleteAcceptedShardToBeacon(shardID byte, shardBlkHash common.Hash) error {
	prefix := append([]byte{shardID}, shardBlkHash[:]...)
	key := append(shardToBeaconKeyPrefix, prefix...)
	if err := db.Delete(key); err != nil {
		return nil
	}
	return nil
}

func (db *db) DeleteIncomingCrossShard(shardID byte, crossShardID byte, crossBlkHash common.Hash) error {
	prefix := append([]byte{shardID}, append([]byte{crossShardID}, crossBlkHash[:]...)...)
	// csh-ShardID-CrossShardID-CrossShardBlockHash : ShardBlockHeight
	key := append(crossShardKeyPrefix, prefix...)
	if err := db.Delete(key); err != nil {
		return err
	}
	return nil
}
