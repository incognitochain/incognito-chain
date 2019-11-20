package rawdb

import (
	"encoding/binary"
	"github.com/incognitochain/incognito-chain/incdb"
	"log"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/databasemp"
	"github.com/pkg/errors"
)

// StorePrivacyCustomToken - store data about privacy custom token when init
// Key: privacy-token-init-{tokenID}
// Value: txHash
func StorePrivacyToken(db incdb.Database, tokenID common.Hash, txHash []byte) error {
	key := addPrefixToKeyHash(string(privacyTokenInitPrefix), tokenID) // token-init-{tokenID}
	ok, _ := db.Has(key)
	if !ok {
		// not exist tx about init this token
		if err := db.Put(key, txHash); err != nil {
			return incdb.NewDatabaseError(incdb.UnexpectedError, err)
		}
	}
	return nil
}

func StorePrivacyTokenTx(db incdb.Database, tokenID common.Hash, shardID byte, blockHeight uint64, txIndex int32, txHash []byte) error {
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
		return incdb.NewDatabaseError(incdb.UnexpectedError, err)
	}
	return nil
}

func PrivacyTokenIDExisted(db incdb.Database, tokenID common.Hash) bool {
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

func PrivacyTokenIDCrossShardExisted(db incdb.Database, tokenID common.Hash) bool {
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
func ListPrivacyToken(db incdb.Database) ([][]byte, error) {
	result := make([][]byte, 0)
	iter := db.NewIteratorWithPrefix(privacyTokenInitPrefix)
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

func PrivacyTokenTxs(db incdb.Database, tokenID common.Hash) ([]common.Hash, error) {
	result := make([]common.Hash, 0)
	key := addPrefixToKeyHash(string(privacyTokenPrefix), tokenID)
	// PubKey = token-{tokenID}
	iter := db.NewIteratorWithPrefix(key)
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

func StorePrivacyTokenCrossShard(db incdb.Database, tokenID common.Hash, tokenValue []byte) error {
	key := addPrefixToKeyHash(string(privacyTokenCrossShardPrefix), tokenID)
	if err := db.Put(key, tokenValue); err != nil {
		return incdb.NewDatabaseError(incdb.UnexpectedError, err)
	}
	return nil
}

/*
	Return all data of token
*/
func ListPrivacyTokenCrossShard(db incdb.Database) ([][]byte, error) {
	result := make([][]byte, 0)
	iter := db.NewIteratorWithPrefix(privacyTokenCrossShardPrefix)
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
