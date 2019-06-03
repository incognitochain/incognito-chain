package lvdb

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/constant-money/constant-chain/common"

	"github.com/constant-money/constant-chain/database"

	"math/big"

	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// StoreSerialNumbers - store list serialNumbers by shardID
func (db *db) StoreSerialNumbers(tokenID common.Hash, serialNumbers [][]byte, shardID byte) error {
	key := db.GetKey(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)

	var lenData int64
	len, err := db.GetSerialNumbersLength(tokenID, shardID)
	if err != nil && len == nil {
		return err
	}
	if len == nil {
		lenData = 0
	} else {
		lenData = len.Int64()
	}
	for _, s := range serialNumbers {
		newIndex := big.NewInt(lenData).Bytes()
		if lenData == 0 {
			newIndex = []byte{0}
		}
		// keySpec1 store serialNumber and index
		keySpec1 := append(key, s...)
		if err := db.lvdb.Put(keySpec1, newIndex, nil); err != nil {
			return err
		}
		// keyStoreLen store last index of array serialNumber
		keyStoreLen := append(key, []byte("len")...)
		if err := db.lvdb.Put(keyStoreLen, newIndex, nil); err != nil {
			return err
		}
		lenData++
	}
	return nil
}

// HasSerialNumber - Check serialNumber in list SerialNumbers by shardID
func (db *db) HasSerialNumber(tokenID common.Hash, serialNumber []byte, shardID byte) (bool, error) {
	key := db.GetKey(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, serialNumber...)
	_, err := db.Get(keySpec)
	if err != nil {
		return false, nil
	} else {
		return true, nil
	}
	return false, nil
}

// GetCommitmentIndex - return index of commitment in db list
func (db *db) GetSerialNumbersLength(tokenID common.Hash, shardID byte) (*big.Int, error) {
	key := db.GetKey(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)
	keyStoreLen := append(key, []byte("len")...)
	data, err := db.Get(keyStoreLen)
	if err != nil {
		return new(big.Int).SetInt64(0), nil
	} else {
		lenArray := new(big.Int).SetBytes(data)
		lenArray = lenArray.Add(lenArray, new(big.Int).SetInt64(1))
		return lenArray, nil
	}
	return new(big.Int).SetInt64(0), nil
}

// CleanSerialNumbers - clear all list serialNumber in DB
func (db *db) CleanSerialNumbers() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(serialNumbersPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

//StoreOutputCoins - store all output coin of pubkey
// key: [outcoinsPrefix][tokenID][shardID][hash(output)]
// value: output in bytes
func (db *db) StoreOutputCoins(tokenID common.Hash, publicKey []byte, outputCoinArr [][]byte, shardID byte) error {
	key := db.GetKey(string(outcoinsPrefix), tokenID)
	key = append(key, shardID)

	key = append(key, publicKey...)
	for _, outputCoin := range outputCoinArr {
		keyTemp := append(key, common.HashB(outputCoin)...)
		if err := db.lvdb.Put(keyTemp, outputCoin, nil); err != nil {
			return err
		}
	}

	return nil
}

/* deprecated func
func (db *db) StoreOutputCoins(tokenID common.Hash, publicKey []byte, outputCoinArr [][]byte, shardID byte) error {
	key := db.GetKey(string(outcoinsPrefix), tokenID)
	key = append(key, shardID)

	// store for pubkey:[outcoint1, outcoint2, ...]
	key = append(key, publicKey...)
	var arrDatabyPubkey [][]byte
	resByPubkey, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if len(resByPubkey) > 0 {
		if err := json.Unmarshal(resByPubkey, &arrDatabyPubkey); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	arrDatabyPubkey = append(arrDatabyPubkey, outputCoinArr...)
	resByPubkey, err = json.Marshal(arrDatabyPubkey)
	if err != nil {
		return err
	}
	if err := db.lvdb.Put(key, resByPubkey, nil); err != nil {
		return err
	}

	return nil
}*/

// StoreCommitments - store list commitments by shardID
func (db *db) StoreCommitments(tokenID common.Hash, pubkey []byte, commitments [][]byte, shardID byte) error {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)

	keySpec4 := append(key, pubkey...)
	var arrDatabyPubkey [][]byte
	resByPubkey, err := db.lvdb.Get(keySpec4, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if len(resByPubkey) > 0 {
		if err := json.Unmarshal(resByPubkey, &arrDatabyPubkey); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}

	var lenData uint64
	len, err := db.GetCommitmentLength(tokenID, shardID)
	if err != nil && len == nil {
		return err
	}
	if len == nil {
		lenData = 0
	} else {
		lenData = len.Uint64()
	}
	for _, c := range commitments {
		newIndex := new(big.Int).SetUint64(lenData).Bytes()
		if lenData == 0 {
			newIndex = []byte{0}
		}
		// keySpec1 use for create proof random
		keySpec1 := append(key, newIndex...)
		if err := db.lvdb.Put(keySpec1, c, nil); err != nil {
			return err
		}
		// keySpec2 use for validate
		keySpec2 := append(key, c...)
		if err := db.lvdb.Put(keySpec2, newIndex, nil); err != nil {
			return err
		}
		// keySpec3 store last index of array commitment
		keySpec3 := append(key, []byte("len")...)
		if err := db.lvdb.Put(keySpec3, newIndex, nil); err != nil {
			return err
		}
		// keySpec4 store for pubkey:[newindex1, newindex2]
		arrDatabyPubkey = append(arrDatabyPubkey, newIndex)
		resByPubkey, err = json.Marshal(arrDatabyPubkey)
		if err != nil {
			return err
		}
		if err := db.lvdb.Put(keySpec4, resByPubkey, nil); err != nil {
			return err
		}
		lenData++
	}

	return nil
}

// HasCommitment - Check commitment in list commitments by shardID
func (db *db) HasCommitment(tokenID common.Hash, commitment []byte, shardID byte) (bool, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, commitment...)
	_, err := db.Get(keySpec)
	if err != nil {
		return false, nil
	} else {
		return true, nil
	}
	return false, nil
}

func (db *db) HasCommitmentIndex(tokenID common.Hash, commitmentIndex uint64, shardID byte) (bool, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, new(big.Int).SetUint64(commitmentIndex).Bytes()...)
	_, err := db.Get(keySpec)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
	return false, nil
}

func (db *db) GetCommitmentByIndex(tokenID common.Hash, commitmentIndex uint64, shardID byte) ([]byte, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	//keySpec := make([]byte, len(key))
	var keySpec []byte
	if commitmentIndex == 0 {
		keySpec = append(key, byte(0))
	} else {
		keySpec = append(key, new(big.Int).SetUint64(commitmentIndex).Bytes()...)
	}
	data, err := db.Get(keySpec)
	if err != nil {
		return data, err
	} else {
		return data, nil
	}
	return data, nil
}

// GetCommitmentIndex - return index of commitment in db list
func (db *db) GetCommitmentIndex(tokenID common.Hash, commitment []byte, shardID byte) (*big.Int, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, commitment...)
	data, err := db.Get(keySpec)
	if err != nil {
		return nil, err
	} else {
		return new(big.Int).SetBytes(data), nil
	}
	return nil, nil
}

// GetCommitmentIndex - return index of commitment in db list
func (db *db) GetCommitmentLength(tokenID common.Hash, shardID byte) (*big.Int, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, []byte("len")...)
	data, err := db.Get(keySpec)
	if err != nil {
		return new(big.Int).SetInt64(0), err
	} else {
		lenArray := new(big.Int).SetBytes(data)
		lenArray = lenArray.Add(lenArray, new(big.Int).SetInt64(1))
		return lenArray, nil
	}
	return new(big.Int).SetInt64(0), nil
}

func (db *db) GetCommitmentIndexsByPubkey(tokenID common.Hash, pubkey []byte, shardID byte) ([][]byte, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)

	//keySpec4 := make([]byte, len(key))
	keySpec4 := append(key, pubkey...)
	var arrDatabyPubkey [][]byte
	resByPubkey, err := db.lvdb.Get(keySpec4, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if len(resByPubkey) > 0 {
		if err := json.Unmarshal(resByPubkey, &arrDatabyPubkey); err != nil {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	return arrDatabyPubkey, nil
}

//GetOutcoinsByPubkey - get all output coin of pubkey
// key: [outcoinsPrefix][tokenID][shardID][hash(output)]
// value: output in bytes
func (db *db) GetOutcoinsByPubkey(tokenID common.Hash, pubkey []byte, shardID byte) ([][]byte, error) {
	key := db.GetKey(string(outcoinsPrefix), tokenID)
	key = append(key, shardID)

	key = append(key, pubkey...)
	arrDatabyPubkey := make([][]byte, 0)
	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	if iter.Error() != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(iter.Error(), "db.lvdb.NewIterator"))
	}
	for iter.Next() {
		value := iter.Value()
		arrDatabyPubkey = append(arrDatabyPubkey, value)
	}
	return arrDatabyPubkey, nil
}

/* deprecated func
func (db *db) GetOutcoinsByPubkey(tokenID common.Hash, pubkey []byte, shardID byte) ([][]byte, error) {
	key := db.GetKey(string(outcoinsPrefix), tokenID)
	key = append(key, shardID)

	key = append(key, pubkey...)
	var arrDatabyPubkey [][]byte
	resByPubkey, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	if len(resByPubkey) > 0 {
		if err := json.Unmarshal(resByPubkey, &arrDatabyPubkey); err != nil {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	return arrDatabyPubkey, nil
}*/

// CleanCommitments - clear all list commitments in DB
func (db *db) CleanCommitments() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(commitmentsPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

// StoreSNDerivators - store list serialNumbers by shardID
func (db *db) StoreSNDerivators(tokenID common.Hash, sndArray [][]byte, shardID byte) error {
	key := db.GetKey(string(snderivatorsPrefix), tokenID)
	key = append(key, shardID)

	// "snderivator-data:nil"
	for _, snd := range sndArray {
		keySpec := append(key, snd...)
		if err := db.lvdb.Put(keySpec, []byte{}, nil); err != nil {
			return err
		}
	}
	return nil
}

// HasSNDerivator - Check SnDerivator in list SnDerivators by shardID
func (db *db) HasSNDerivator(tokenID common.Hash, data []byte, shardID byte) (bool, error) {
	key := db.GetKey(string(snderivatorsPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, data...)
	_, err := db.Get(keySpec)
	if err != nil {
		return false, nil
	} else {
		return true, nil
	}
	return false, nil
}

// CleanCommitments - clear all list commitments in DB
func (db *db) CleanSNDerivator() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(snderivatorsPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

// StoreFeeEstimator - Store data for FeeEstimator object
func (db *db) StoreFeeEstimator(val []byte, shardID byte) error {
	if err := db.Put(append(feeEstimator, shardID), val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

// GetFeeEstimator - Get data for FeeEstimator object as a json in byte format
func (db *db) GetFeeEstimator(shardID byte) ([]byte, error) {
	b, err := db.lvdb.Get(append(feeEstimator, shardID), nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return b, err
}

// CleanFeeEstimator - Clear FeeEstimator
func (db *db) CleanFeeEstimator() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(feeEstimator), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

/*
	StoreTransactionIndex
	Store tx detail location
  Key: prefixTx-txHash
	H: blockHash-blockIndex
*/
func (db *db) StoreTransactionIndex(txId common.Hash, blockHash common.Hash, index int) error {
	key := string(transactionKeyPrefix) + txId.String()
	value := blockHash.String() + string(Splitter) + strconv.Itoa(index)
	if err := db.lvdb.Put([]byte(key), []byte(value), nil); err != nil {
		return err
	}

	return nil
}

/*
  Get Transaction by ID
*/

func (db *db) GetTransactionIndexById(txId common.Hash) (common.Hash, int, *database.DatabaseError) {
	key := string(transactionKeyPrefix) + txId.String()
	_, err := db.HasValue([]byte(key))
	if err != nil {
		return common.Hash{}, -1, database.NewDatabaseError(database.ErrUnexpected, err)
	}

	res, err := db.Get([]byte(key))
	if err != nil {
		return common.Hash{}, -1, database.NewDatabaseError(database.ErrUnexpected, err)
	}
	reses := strings.Split(string(res), (string(Splitter)))
	hash, err := common.Hash{}.NewHashFromStr(reses[0])
	if err != nil {
		return common.Hash{}, -1, database.NewDatabaseError(database.ErrUnexpected, err)
	}
	index, err := strconv.Atoi(reses[1])
	if err != nil {
		return common.Hash{}, -1, database.NewDatabaseError(database.ErrUnexpected, err)
	}
	return *hash, index, nil
}
