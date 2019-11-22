package lvdb

import (
	"encoding/binary"
	"log"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/databasemp"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// StorePrivacyCustomToken - store data about privacy custom token when init
// Key: privacy-token-init-{tokenID}
// Value: txHash
func (db *db) StorePrivacyToken(tokenID common.Hash, txHash []byte) error {
	key := addPrefixToKeyHash(string(privacyTokenInitPrefix), tokenID) // token-init-{tokenID}
	ok, _ := db.HasValue(key)
	if !ok {
		// not exist tx about init this token
		if err := db.Put(key, txHash); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, err)
		}
	}
	return nil
}

func (db *db) StorePrivacyTokenTx(tokenID common.Hash, shardID byte, blockHeight uint64, txIndex int32, txHash []byte) error {
	key := addPrefixToKeyHash(string(privacyTokenPrefix), tokenID) // token-{tokenID}-shardID-(999999999-blockHeight)-(999999999-txIndex)
	key = append(key, shardID)
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, bigNumber-blockHeight)
	key = append(key, bs...)
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(bigNumber-txIndex))
	key = append(key, bs...)
	log.Println(string(key))
	if err := db.Put(key, txHash); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, err)
	}
	return nil
}

func (db *db) PrivacyTokenIDExisted(tokenID common.Hash) bool {
	key := addPrefixToKeyHash(string(privacyTokenInitPrefix), tokenID) // token-init-{tokenID}
	data, err := db.Get(key)
	if err != nil {
		return false
	}
	if data == nil || len(data) == 0 {
		return false
	}
	return true
}

func (db *db) PrivacyTokenIDCrossShardExisted(tokenID common.Hash) bool {
	key := addPrefixToKeyHash(string(privacyTokenCrossShardPrefix), tokenID)
	data, err := db.Get(key)
	if err != nil {
		return false
	}
	if data == nil || len(data) == 0 {
		return false
	}
	return true
}

/*
	Return list of txhash
*/
func (db *db) ListPrivacyToken() ([][]byte, error) {
	result := make([][]byte, 0)
	iter := db.lvdb.NewIterator(util.BytesPrefix(privacyTokenInitPrefix), nil)
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		result = append(result, value)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return result, nil
}

func (db *db) PrivacyTokenTxs(tokenID common.Hash) ([]common.Hash, error) {
	result := make([]common.Hash, 0)
	key := addPrefixToKeyHash(string(privacyTokenPrefix), tokenID)
	// PubKey = token-{tokenID}
	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	log.Println(string(key))
	for iter.Next() {
		value := iter.Value()
		hash, _ := common.Hash{}.NewHash(value)
		result = append(result, *hash)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return result, nil
}

func (db *db) StorePrivacyTokenCrossShard(tokenID common.Hash, tokenValue []byte) error {
	key := addPrefixToKeyHash(string(privacyTokenCrossShardPrefix), tokenID)
	if err := db.Put(key, tokenValue); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, err)
	}
	return nil
}

/*
	Return all data of token

*/
func (db *db) ListPrivacyTokenCrossShard() ([][]byte, error) {
	result := make([][]byte, 0)
	iter := db.lvdb.NewIterator(util.BytesPrefix(privacyTokenCrossShardPrefix), nil)
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		result = append(result, value)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return result, nil
}
