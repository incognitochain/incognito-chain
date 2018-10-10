package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/blockchain"
)

var (
	chainIDPrefix     = []byte("c")
	blockKeyPrefix    = []byte("b-")
	blockKeyIdxPrefix = []byte("i-")
	usedTxKey         = []byte("usedTx")
	notUsedTxKey      = []byte("notusedTx")
	usedTxBondKey     = []byte("usedTxBond")
	notUsedBondTxKey  = []byte("notusedTxBond")
	bestBlockKey      = []byte("bestBlock")
	feeEstimator      = []byte("feeEstimator")
	bestCndListKey    = []byte("bestCndList")
)

func open(dbPath string) (database.DatabaseInterface, error) {
	lvdb, err := leveldb.OpenFile(filepath.Join(dbPath, "db"), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "levelvdb.OpenFile %s", dbPath)
	}
	return &db{lvdb: lvdb}, nil
}

type db struct {
	lvdb *leveldb.DB
}

func (db *db) hasBlock(key []byte) bool {
	ret, err := db.lvdb.Has(key, nil)
	if err != nil {
		return false
	}
	return ret
}

type hasher interface {
	Hash() *common.Hash
}

func (db *db) StoreBlock(v interface{}, chainID byte) error {
	h, ok := v.(hasher)
	if !ok {
		return errors.New("v must implement Hash() method")
	}
	var (
		hash = h.Hash()
		key  = append(append(chainIDPrefix, chainID), append(blockKeyPrefix, hash[:]...)...)
		// key should look like this c10{b-[blockhash]}:{b-[blockhash]}
		keyB = append(blockKeyPrefix, hash[:]...)
		// key should look like this {b-blockhash}:block
	)
	if db.hasBlock(key) {
		return errors.Errorf("block %s already exists", hash.String())
	}
	val, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	if err := db.put(key, keyB); err != nil {
		return errors.Wrap(err, "db.Put")
	}
	if err := db.put(keyB, val); err != nil {
		return errors.Wrap(err, "db.Put")
	}
	return nil
}

func (db *db) Close() error {
	return errors.Wrap(db.lvdb.Close(), "db.lvdb.Close")
}

func (db *db) put(key, value []byte) error {
	if err := db.lvdb.Put(key, value, nil); err != nil {
		return errors.Wrap(err, "db.lvdb.Put")
	}
	return nil
}

func (db *db) HasBlock(hash *common.Hash) (bool, error) {
	if exists := db.hasBlock(db.getKeyBlock(hash)); exists {
		return true, nil
	}
	return false, nil
}

func (db *db) FetchBlock(hash *common.Hash) ([]byte, error) {
	block, err := db.lvdb.Get(db.getKeyBlock(hash), nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.lvdb.Get")
	}

	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

func (db *db) StoreNullifiers(nullifier []byte, typeJoinSplitDesc string, chainId byte) error {
	key := append(usedTxKey, []byte(typeJoinSplitDesc)...)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return errors.Wrap(err, "db.lvdb.Get")
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return errors.Wrap(err, "json.Unmarshal")
		}
	}
	txs = append(txs, nullifier)
	b, err := json.Marshal(txs)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) StoreCommitments(commitments []byte, typeJoinSplitDesc string, chainId byte) error {
	key := append(notUsedTxKey, []byte(typeJoinSplitDesc)...)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return errors.Wrap(err, "db.lvdb.Get")
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return errors.Wrap(err, "json.Unmarshal")
		}
	}
	txs = append(txs, commitments)
	b, err := json.Marshal(txs)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) FetchNullifiers(typeJoinSplitDesc string, chainId byte) ([][]byte, error) {
	key := append(usedTxKey, []byte(typeJoinSplitDesc)...)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), errors.Wrap(err, "db.lvdb.Get")
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([][]byte, 0), errors.Wrap(err, "json.Unmarshal")
		}
	}
	return txs, nil
}

func (db *db) FetchCommitments(typeJoinSplitDesc string, chainId byte) ([][]byte, error) {
	key := append(notUsedTxKey, []byte(typeJoinSplitDesc)...)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), errors.Wrap(err, "db.lvdb.Get")
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([][]byte, 0), errors.Wrap(err, "json.Unmarshal")
		}
	}
	return txs, nil
}

func (db *db) HasNullifier(nullifier []byte, typeJoinSplitDesc string, chainId byte) (bool, error) {
	listNullifiers, err := db.FetchNullifiers(typeJoinSplitDesc, chainId)
	if err != nil {
		return false, err
	}
	for _, item := range listNullifiers {
		if bytes.Equal(item, nullifier) {
			return true, nil
		}
	}
	return false, nil
}

func (db *db) HasCommitment(commitment []byte, typeJoinSplitDesc string, chainId byte) (bool, error) {
	listCommitments, err := db.FetchCommitments(typeJoinSplitDesc, chainId)
	if err != nil {
		return false, err
	}
	for _, item := range listCommitments {
		if bytes.Equal(item, commitment) {
			return true, nil
		}
	}
	return false, nil
}

func (db *db) StoreBestState(v interface{}, chainID byte) error {
	val, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	key := append(bestBlockKey, chainID)
	if err := db.put(key, val); err != nil {
		return errors.Wrap(err, "db.Put")
	}
	return nil
}

func (db *db) FetchBestState(chainID byte) ([]byte, error) {
	key := append(bestBlockKey, chainID)
	block, err := db.lvdb.Get(key, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.lvdb.Get")
	}
	return block, nil
}

func (db *db) StoreBlockIndex(h *common.Hash, idx int32, chainID byte) error {
	buf := make([]byte, 5)
	binary.LittleEndian.PutUint32(buf, uint32(idx))
	buf[4] = chainID
	//{i-[hash]}:index-chainid
	if err := db.lvdb.Put(db.getKeyIdx(h), buf, nil); err != nil {
		return errors.Wrap(err, "db.lvdb.Put")
	}
	//{index-chainid}:[hash]
	if err := db.lvdb.Put(buf, h[:], nil); err != nil {
		return errors.Wrap(err, "db.lvdb.Put")
	}
	return nil
}

func (db *db) GetIndexOfBlock(h *common.Hash) (int32, byte, error) {
	b, err := db.lvdb.Get(db.getKeyIdx(h), nil)
	//{i-[hash]}:index-chainid
	if err != nil {
		return 0, 0, errors.Wrap(err, "db.lvdb.Get")
	}

	var idx int32
	var chainID byte
	if err := binary.Read(bytes.NewReader(b[:4]), binary.LittleEndian, &idx); err != nil {
		return 0, 0, errors.Wrap(err, "binary.Read")
	}
	if err = binary.Read(bytes.NewReader(b[4:]), binary.LittleEndian, &chainID); err != nil {
		return 0, 0, errors.Wrap(err, "binary.Read")
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
		return nil, errors.Wrap(err, "db.lvdb.Get")
	}
	h := new(common.Hash)
	_ = h.SetBytes(b[:])
	return h, nil
}

func (db *db) FetchAllBlocks() ([][]*common.Hash, error) {
	var keys [][]*common.Hash
	for chainID := byte(0); chainID < blockchain.ChainCount; chainID++ {
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
			return nil, errors.Wrap(err, "iter.Error")
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
		return nil, errors.Wrap(err, "iter.Error")
	}
	return keys, nil
}

func (db *db) StoreFeeEstimator(val []byte, chainId byte) error {
	if err := db.put(append(feeEstimator, chainId), val); err != nil {
		return errors.Wrap(err, "db.Put")
	}
	return nil
}

func (db *db) GetFeeEstimator(chainId byte) ([]byte, error) {
	b, err := db.lvdb.Get(append(feeEstimator, chainId), nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.ldb.Get")
	}
	return b, err
}

func (db *db) getKeyBlock(h *common.Hash) []byte {
	var key []byte
	key = append(blockKeyPrefix, h[:]...)
	return key
}

func (db *db) getKeyIdx(h *common.Hash) []byte {
	var key []byte
	key = append(blockKeyIdxPrefix, h[:]...)
	return key
}
