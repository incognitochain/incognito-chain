package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) StoreBlock(v interface{}, chainID byte) error {
	h, ok := v.(hasher)
	if !ok {
		return database.NewDatabaseError(database.NotImplHashMethod, errors.New("v must implement Hash() method"))
	}
	var (
		hash = h.Hash()
		key  = append(append(chainIDPrefix, chainID), append(blockKeyPrefix, hash[:]...)...)
		// PubKey should look like this c10{b-[blockhash]}:{b-[blockhash]}
		keyB = append(blockKeyPrefix, hash[:]...)
		// PubKey should look like this {b-blockhash}:block
	)
	if ok, _ := db.HasValue(key); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %s already exists", hash.String()))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.Put(key, keyB); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.Put(keyB, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	//fmt.Println("Test Store Block keyB: ", string(keyB))
	return nil
}

func (db *db) StoreBlockHeader(v interface{}, hash *common.Hash, chainID byte) error {
	//fmt.Println("Log in StoreBlockHeader", v, hash, chainID)
	var (
		key = append(append(chainIDPrefix, chainID), append(blockKeyPrefix, hash[:]...)...)
		// PubKey should look like this c10{bh-[blockhash]}:{bh-[blockhash]}
		keyB = append(blockKeyPrefix, hash[:]...)
		// PubKey should look like this {bh-blockhash}:block
	)
	if ok, _ := db.HasValue(key); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("block %s already exists", hash.String()))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.Put(key, keyB); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.Put(keyB, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	//fmt.Println("Test StoreBlockHeader keyB: ", string(keyB))
	return nil
}

func (db *db) HasBlock(hash *common.Hash) (bool, error) {
	exists, err := db.HasValue(db.GetKey(string(blockKeyPrefix), hash))
	if err != nil {
		return false, err
	} else {
		return exists, nil
	}
}

func (db *db) FetchBlock(hash *common.Hash) ([]byte, error) {
	block, err := db.Get(db.GetKey(string(blockKeyPrefix), hash))
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

func (db *db) DeleteBlock(hash *common.Hash, idx int32, chainID byte) error {
	// Delete block
	err := db.lvdb.Delete(db.GetKey(string(blockKeyPrefix), hash), nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	// Delete block index
	err = db.lvdb.Delete(db.GetKey(string(blockKeyIdxPrefix), hash), nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	buf := make([]byte, 5)
	binary.LittleEndian.PutUint32(buf, uint32(idx))
	buf[4] = chainID
	err = db.lvdb.Delete(buf, nil)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return nil
}

func (db *db) StoreBestState(v interface{}, chainID byte) error {
	val, err := json.Marshal(v)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	key := append(bestBlockKey, chainID)
	if err := db.Put(key, val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	return nil
}

func (db *db) FetchBestState(chainID byte) ([]byte, error) {
	key := append(bestBlockKey, chainID)
	block, err := db.Get(key)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.get"))
	}
	return block, nil
}

func (db *db) CleanBestState() error {
	for chainID := byte(0); chainID < common.TotalValidators; chainID++ {
		key := append(bestBlockKey, chainID)
		err := db.lvdb.Delete(key, nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.delete"))
		}
	}
	return nil
}

func (db *db) StoreBlockIndex(h *common.Hash, idx int32, chainID byte) error {
	buf := make([]byte, 5)
	binary.LittleEndian.PutUint32(buf, uint32(idx))
	buf[4] = chainID
	//{i-[hash]}:index-chainid
	if err := db.lvdb.Put(db.GetKey(string(blockKeyIdxPrefix), h), buf, nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	//{index-chainid}:[hash]
	if err := db.lvdb.Put(buf, h[:], nil); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetIndexOfBlock(h *common.Hash) (int32, byte, error) {
	b, err := db.lvdb.Get(db.GetKey(string(blockKeyIdxPrefix), h), nil)
	//{i-[hash]}:index-chainid
	if err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.get"))
	}

	var idx int32
	var chainID byte
	if err := binary.Read(bytes.NewReader(b[:4]), binary.LittleEndian, &idx); err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	if err = binary.Read(bytes.NewReader(b[4:]), binary.LittleEndian, &chainID); err != nil {
		return 0, 0, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "binary.Read"))
	}
	return idx, chainID, nil
}

func (db *db) GetBlockByIndex(idx int32, chainID byte) (*common.Hash, error) {
	buf := make([]byte, 5)
	binary.LittleEndian.PutUint32(buf, uint32(idx))
	buf[4] = chainID
	// {index-chainid}: {blockhash}

	b, err := db.lvdb.Get(buf, nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	h := new(common.Hash)
	_ = h.SetBytes(b[:])
	return h, nil
}

func (db *db) FetchAllBlocks() (map[byte][]*common.Hash, error) {
	var keys map[byte][]*common.Hash
	for chainID := byte(0); chainID < 20; chainID++ {
		prefix := append(append(chainIDPrefix, chainID), blockKeyPrefix...)
		// prefix {c10{b-......}}
		iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
		for iter.Next() {
			h := new(common.Hash)
			_ = h.SetBytes(iter.Key()[len(prefix):])
			keys[chainID] = append(keys[chainID], h)
		}
		iter.Release()
		if err := iter.Error(); err != nil {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
		}
	}
	return keys, nil
}

func (db *db) FetchChainBlocks(chainID byte) ([]*common.Hash, error) {
	var keys []*common.Hash
	prefix := append(append(chainIDPrefix, chainID), blockKeyPrefix...)
	//prefix {c10{b-......}}
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		h := new(common.Hash)
		_ = h.SetBytes(iter.Key()[len(prefix):])
		keys = append(keys, h)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return keys, nil
}
