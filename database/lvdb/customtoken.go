package lvdb

import (
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/ninjadotorg/constant/common"
)

func (db *db) StoreCustomToken(tokenID *common.Hash, txHash []byte) error {
	key := db.getKey(string(tokenInitPrefix), tokenID) // token-init-{tokenID}
	if err := db.lvdb.Put(key, txHash, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) StoreCustomTokenTx(tokenID *common.Hash, chainID byte, blockHeight int32, txIndex int32, txHash []byte) error {
	key := db.getKey(string(tokenPrefix), tokenID) // token-{tokenID}-chainID-(999999999-blockHeight)-txIndex
	key = append(key, tokenID[:]...)
	key = append(key, chainID)
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(999999999-blockHeight))
	key = append(key, bs...)
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(999999999-txIndex))
	key = append(key, bs...)
	if err := db.lvdb.Put(key, txHash, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) ListCustomToken() ([][]byte, error) {
	result := make([][]byte, 0)
	iter := db.lvdb.NewIterator(util.BytesPrefix(tokenInitPrefix), nil)
	for iter.Next() {
		result = append(result, iter.Value())
	}
	iter.Release()
	return result, nil
}

func (db *db) CustomTokenTxs(tokenID *common.Hash) ([][]byte, error) {
	result := make([][]byte, 0)
	key := tokenPrefix
	key = append(key, tokenID[:]...)
	// key = token-{tokenID}
	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	for iter.Next() {
		result = append(result, iter.Value())
	}
	iter.Release()
	return result, nil
}
