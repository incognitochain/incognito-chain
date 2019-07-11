package lvdb

import (
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/database"

	"math/big"

	"github.com/pkg/errors"
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
		if err := db.Put(keySpec1, newIndex); err != nil {
			return err
		}
		// keyStoreLen store last index of array serialNumber
		keyStoreLen := append(key, []byte("len")...)
		if err := db.Put(keyStoreLen, newIndex); err != nil {
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
	hasValue, err := db.HasValue(keySpec)
	if err != nil {
		return false, err
	} else {
		return hasValue, nil
	}
}

// GetCommitmentIndex - return index of commitment in db list
func (db *db) GetSerialNumbersLength(tokenID common.Hash, shardID byte) (*big.Int, error) {
	key := db.GetKey(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)
	keyStoreLen := append(key, []byte("len")...)
	hasValue, err := db.HasValue(keyStoreLen)
	if err != nil {
		return nil, err
	} else {
		if !hasValue {
			return nil, nil
		} else {
			data, err := db.Get(keyStoreLen)
			if err != nil {
				return new(big.Int).SetInt64(0), nil
			} else {
				lenArray := new(big.Int).SetBytes(data)
				lenArray = lenArray.Add(lenArray, new(big.Int).SetInt64(1))
				return lenArray, nil
			}
		}
	}
}

// CleanSerialNumbers - clear all list serialNumber in DB
func (db *db) CleanSerialNumbers() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(serialNumbersPrefix), nil)
	for iter.Next() {
		err := db.Delete(iter.Key())
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
	batchData := []database.BatchData{}
	for _, outputCoin := range outputCoinArr {
		keyTemp := make([]byte, len(key))
		copy(keyTemp, key)
		keyTemp = append(keyTemp, common.HashB(outputCoin)...)
		/* deprecated
		if err := db.Put(keyTemp, outputCoin); err != nil {
			return err
		}*/
		// Put to batch
		batchData = append(batchData, database.BatchData{
			Key:   keyTemp,
			Value: outputCoin,
		})
	}
	if len(batchData) > 0 {
		err := db.PutBatch(batchData)
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreCommitments - store list commitments by shardID
func (db *db) StoreCommitments(tokenID common.Hash, pubkey []byte, commitments [][]byte, shardID byte) error {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)

	// keySpec3 store last index of array commitment
	keySpec3 := make([]byte, len(key)+len("len"))
	temp := append(key, []byte("len")...)
	copy(keySpec3, temp)

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
		if err := db.Put(keySpec1, c); err != nil {
			return err
		}
		// keySpec2 use for validate
		keySpec2 := append(key, c...)
		if err := db.Put(keySpec2, newIndex); err != nil {
			return err
		}

		// len of commitment array
		if err := db.Put(keySpec3, newIndex); err != nil {
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
	hasValue, err := db.HasValue(keySpec)
	if err != nil {
		return false, err
	} else {
		return hasValue, nil
	}
}

func (db *db) HasCommitmentIndex(tokenID common.Hash, commitmentIndex uint64, shardID byte) (bool, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	var keySpec []byte
	if commitmentIndex == 0 {
		keySpec = append(key, byte(0))
	} else {
		keySpec = append(key, new(big.Int).SetUint64(commitmentIndex).Bytes()...)
	}
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
	hasValue, err := db.HasValue(keySpec)
	if err != nil {
		return nil, err
	} else {
		if !hasValue {
			return nil, nil
		} else {
			data, err := db.Get(keySpec)
			if err != nil {
				return nil, err
			} else {
				lenArray := new(big.Int).SetBytes(data)
				lenArray = lenArray.Add(lenArray, new(big.Int).SetInt64(1))
				return lenArray, nil
			}
		}
	}
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
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		arrDatabyPubkey = append(arrDatabyPubkey, value)
	}
	iter.Release()
	return arrDatabyPubkey, nil
}

// CleanCommitments - clear all list commitments in DB
func (db *db) CleanCommitments() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(commitmentsPrefix), nil)
	for iter.Next() {
		err := db.Delete(iter.Key())
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
	batchData := []database.BatchData{}
	for _, snd := range sndArray {
		keySpec := make([]byte, len(key))
		copy(keySpec, key)
		keySpec = append(keySpec, snd...)
		// deprecated
		/*if err := db.Put(keySpec, []byte{}); err != nil {
			return err
		}*/
		batchData = append(batchData, database.BatchData{
			Key:   keySpec,
			Value: []byte{},
		})
	}
	if len(batchData) > 0 {
		err := db.PutBatch(batchData)
		if err != nil {
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
	hasValue, err := db.HasValue(keySpec)
	if err != nil {
		return false, err
	} else {
		return hasValue, nil
	}
}

// CleanCommitments - clear all list commitments in DB
func (db *db) CleanSNDerivator() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(snderivatorsPrefix), nil)
	for iter.Next() {
		err := db.Delete(iter.Key())
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
	b, err := db.Get(append(feeEstimator, shardID))
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return b, err
}

// CleanFeeEstimator - Clear FeeEstimator
func (db *db) CleanFeeEstimator() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(feeEstimator), nil)
	for iter.Next() {
		err := db.Delete(iter.Key())
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
	if err := db.Put([]byte(key), []byte(value)); err != nil {
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
		return common.Hash{}, -1, database.NewDatabaseError(database.UnexpectedError, err)
	}

	res, err := db.Get([]byte(key))
	if err != nil {
		return common.Hash{}, -1, database.NewDatabaseError(database.UnexpectedError, err)
	}
	reses := strings.Split(string(res), (string(Splitter)))
	hash, err := common.Hash{}.NewHashFromStr(reses[0])
	if err != nil {
		return common.Hash{}, -1, database.NewDatabaseError(database.UnexpectedError, err)
	}
	index, err := strconv.Atoi(reses[1])
	if err != nil {
		return common.Hash{}, -1, database.NewDatabaseError(database.UnexpectedError, err)
	}
	return *hash, index, nil
}
