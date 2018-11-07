package lvdb

import "encoding/binary"

func (db *db) StoreCustomToken(tokenID []byte, txHash []byte) error {
	key := db.getKey(string(tokenPrefix), tokenID) // token-{tokenID}
	if err := db.lvdb.Put(key, txHash, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) StoreCustomTokenTx(tokenID []byte, chainID byte, blockHeight int32, txIndex int32, txHash []byte) error {
	key := db.getKey(string(tokenPrefix), tokenID) // token-{tokenID}-chainID-(999999999-blockHeight)-txIndex
	key = append(key, tokenID...)
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

func (db *db) ListCustomToken() [][]byte {
	result := make([][]byte, 0)
	return result
}
