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
)

func open(dbPath string) (database.DB, error) {
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
		key  = append(blockKeyPrefix, hash[:]...)
	)
	if db.hasBlock(key) {
		return errors.Errorf("block %s already exists", hash.String())
	}
	val, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	if err := db.put(key, val); err != nil {
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
	if exists := db.hasBlock(db.getKey(hash)); exists {
		return true, nil
	}
	return false, nil
}

func (db *db) FetchBlock(hash *common.Hash) ([]byte, error) {
	block, err := db.lvdb.Get(db.getKey(hash), nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.lvdb.Get")
	}

	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

func (db *db) StoreNullifiers(nullifier []byte, typeJoinSplitDesc string) error {
	res, err := db.lvdb.Get(append(usedTxKey, []byte(typeJoinSplitDesc)...), nil)
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
	if err := db.lvdb.Put(append(usedTxKey, []byte(typeJoinSplitDesc)...), b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) StoreCommitments(commitments []byte, typeJoinSplitDesc string) error {
	res, err := db.lvdb.Get(append(notUsedTxKey, []byte(typeJoinSplitDesc)...), nil)
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
	if err := db.lvdb.Put(append(notUsedTxKey, []byte(typeJoinSplitDesc)...), b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) FetchNullifiers(typeJoinSplitDesc string) ([][]byte, error) {
	res, err := db.lvdb.Get(append(usedTxKey, []byte(typeJoinSplitDesc)...), nil)
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

func (db *db) FetchCommitments(typeJoinSplitDesc string) ([][]byte, error) {
	res, err := db.lvdb.Get(append(notUsedTxKey, []byte(typeJoinSplitDesc)...), nil)
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

func (db *db) HasNullifier(nullifier []byte, typeJoinSplitDesc string) (bool, error) {
	listNullifiers, err := db.FetchNullifiers(typeJoinSplitDesc)
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

func (db *db) HasCommitment(commitment []byte, typeJoinSplitDesc string) (bool, error) {
	listCommitments, err := db.FetchCommitments(typeJoinSplitDesc)
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

func (db *db) StoreBestBlock(v interface{}, chainID byte) error {
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

	if err := db.lvdb.Put(db.getKeyIdx(h), buf, nil); err != nil {
		return errors.Wrap(err, "db.lvdb.Put")
	}
	if err := db.lvdb.Put(buf, h[:], nil); err != nil {
		return errors.Wrap(err, "db.lvdb.Put")
	}
	return nil
}

func (db *db) GetIndexOfBlock(h *common.Hash) (int32, byte, error) {
	b, err := db.lvdb.Get(db.getKeyIdx(h), nil)
	if err != nil {
		return 0, 0, errors.Wrap(err, "db.lvdb.Get")
	}

	var idx int32
	var chainID byte
	if err := binary.Read(bytes.NewReader(b[:3]), binary.LittleEndian, &idx); err != nil {
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
	for chainID := byte(0); chainID <= 19; chainID++ {
		prefix := append(append(chainIDPrefix, chainID), blockKeyPrefix...)
		iter := db.lvdb.NewIterator(util.BytesPrefix(blockKeyPrefix), nil)
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
	iter := db.lvdb.NewIterator(util.BytesPrefix(blockKeyPrefix), nil)
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

func (db *db) getKey(h *common.Hash) []byte {
	var key []byte
	key = append(blockKeyPrefix, h[:]...)
	return key
}

func (db *db) getKeyIdx(h *common.Hash) []byte {
	var key []byte
	key = append(blockKeyIdxPrefix, h[:]...)
	return key
}

/*func (db *db) StoreUtxoEntry(op *transaction.OutPoint, v interface{}) error {
	val, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	if err := db.lvdb.Put([]byte(db.getUtxoKey(op)), val, nil); err != nil {
		return errors.Wrap(err, "db.lvdb.Put")
	}
	return nil
}

func (db *db) FetchUtxoEntry(op *transaction.OutPoint) ([]byte, error) {
	b, err := db.lvdb.Get([]byte(db.getUtxoKey(op)), nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.lvdb.Get")
	}
	return b, nil
}

func (db *db) DeleteUtxoEntry(op *transaction.OutPoint) error {
	if err := db.lvdb.Delete([]byte(db.getUtxoKey(op)), nil); err != nil {
		return errors.Wrap(err, "db.lvdb.Delete")
	}
	return nil
}

func (db *db) getUtxoKey(op *transaction.OutPoint) string {
	return fmt.Sprintf("%s%d", op.Hash.String(), op.Vout)
}*/
