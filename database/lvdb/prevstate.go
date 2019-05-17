package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
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

func (db *db) BackupCommitments(tokenID *common.Hash, shardID byte) error {
	key := getPrevPrefix(false, shardID)
	return nil
}

func (db *db) RestoreCommitments(tokenID *common.Hash, shardID byte) error {
	key := getPrevPrefix(false, shardID)
	return nil
}

func (db *db) BackupOutputCoin(tokenID *common.Hash, pubkey []byte, shardID byte) error {
	key := getPrevPrefix(false, shardID)
	return nil
}

func (db *db) RestoreOutputCoin(tokenID *common.Hash, pubkey []byte, shardID byte) error {
	key := getPrevPrefix(false, shardID)
	return nil
}

func (db *db) BackupSNDerivators(tokenID *common.Hash, shardID byte) error {
	key := getPrevPrefix(false, shardID)
	return nil
}

func (db *db) RestoreSNDerivators(tokenID *common.Hash, shardID byte) error {
	key := getPrevPrefix(false, shardID)
	return nil
}

func (db *db) BackupSerialNumber(tokenID *common.Hash, shardID byte) error {
	key := getPrevPrefix(false, shardID)
	return nil
}

func (db *db) RestoreSerialNumber(tokenID *common.Hash, shardID byte) error {
	key := getPrevPrefix(false, shardID)
	return nil
}

func (db *db) DeleteTransactionIndex(txId *common.Hash) error {
	key := string(transactionKeyPrefix) + txId.String()
	err := db.Delete([]byte(key))
	if err != nil {
		return database.NewDatabaseError(database.ErrUnexpected, err)
	}
	return nil

}

func (db *db) DeleteSerialNumber(tokenID *common.Hash, serialNumber []byte, shardID byte) error {

	return nil
}

func (db *db) DeleteCustomToken(tokenID *common.Hash) error {

	return nil
}

func (db *db) DeleteCustomTokenTx(tokenID *common.Hash, shardID byte, blockHeight uint64) error {

	return nil
}

func (db *db) DeletePrivacyCustomToken(tokenID *common.Hash) error {

	return nil
}

func (db *db) DeletePrivacyCustomTokenTx(tokenID *common.Hash, shardID byte, blockHeight uint64) error {

	return nil
}

func (db *db) DeleteCrossShardNextHeight(byte, byte) error {

	return nil
}

func (db *db) DeleteCommitteeByEpoch(uint64) error {

	return nil
}
