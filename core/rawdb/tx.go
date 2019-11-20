package rawdb

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

// StoreSerialNumbers - store list serialNumbers by shardID
func StoreSerialNumbers(db incdb.Database, tokenID common.Hash, serialNumbers [][]byte, shardID byte) error {
	key := addPrefixToKeyHash(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)

	var lenData int64
	lenSerialNumber, err := GetSerialNumbersLength(db, tokenID, shardID)
	if err != nil && lenSerialNumber == nil {
		return incdb.NewDatabaseError(incdb.StoreSerialNumbersError, err)
	}
	if lenSerialNumber == nil {
		lenData = 0
	} else {
		lenData = lenSerialNumber.Int64()
	}
	for _, s := range serialNumbers {
		newIndex := big.NewInt(lenData).Bytes()
		if lenData == 0 {
			newIndex = []byte{0}
		}
		// keySpec1 store serialNumber and index
		keySpec1 := append(key, s...)
		if err := db.Put(keySpec1, newIndex); err != nil {
			return incdb.NewDatabaseError(incdb.StoreSerialNumbersError, err)
		}
		// keyStoreLen store last index of array serialNumber
		keyStoreLen := append(key, []byte("len")...)
		if err := db.Put(keyStoreLen, newIndex); err != nil {
			return incdb.NewDatabaseError(incdb.StoreSerialNumbersError, err)
		}
		lenData++
	}
	return nil
}

// HasSerialNumber - Check serialNumber in list SerialNumbers by shardID
func HasSerialNumber(db incdb.Database, tokenID common.Hash, serialNumber []byte, shardID byte) (bool, error) {
	key := addPrefixToKeyHash(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, serialNumber...)
	hasValue, err := db.Has(keySpec)
	if err != nil {
		return false, incdb.NewDatabaseError(incdb.HasSerialNumberError, err, serialNumber, shardID, tokenID)
	} else {
		return hasValue, nil
	}
}

// ListSerialNumber -  return all serial number and its index
func ListSerialNumber(db incdb.Database, tokenID common.Hash, shardID byte) (map[string]uint64, error) {
	result := make(map[string]uint64)
	key := addPrefixToKeyHash(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)

	iterator := db.NewIteratorWithPrefix(key)
	for iterator.Next() {
		key1 := make([]byte, len(iterator.Key()))
		copy(key1, iterator.Key())
		if string(key1[len(key1)-3:]) == "len" {
			continue
		}
		serialNumberInByte := key1[len(key1)-privacy.Ed25519KeySize:]
		value := make([]byte, len(iterator.Value()))
		copy(value, iterator.Value())
		index := big.Int{}
		index.SetBytes(value)
		serialNumber := base58.Base58Check{}.Encode(serialNumberInByte, 0x0)
		result[serialNumber] = index.Uint64()
	}
	return result, nil
}

// GetCommitmentIndex - return index of commitment in db list
func GetSerialNumbersLength(db incdb.Database, tokenID common.Hash, shardID byte) (*big.Int, error) {
	key := addPrefixToKeyHash(string(serialNumbersPrefix), tokenID)
	key = append(key, shardID)
	keyStoreLen := append(key, []byte("len")...)
	hasValue, err := db.Has(keyStoreLen)
	if err != nil {
		return nil, incdb.NewDatabaseError(incdb.GetSerialNumbersLengthError, err)
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
func CleanSerialNumbers(db incdb.Database) error {
	iter := db.NewIteratorWithPrefix(serialNumbersPrefix)
	for iter.Next() {
		err := db.Delete(iter.Key())
		if err != nil {
			return incdb.NewDatabaseError(incdb.CleanSerialNumbersError, err)
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return incdb.NewDatabaseError(incdb.CleanSerialNumbersError, err)
	}
	return nil
}

//StoreOutputCoins - store all output coin of pubkey
// key: [outcoinsPrefix][tokenID][shardID][hash(output)]
func StoreOutputCoins(db incdb.Database, tokenID common.Hash, publicKey []byte, outputCoinArr [][]byte, shardID byte) error {
	// value: output in bytes
	key := addPrefixToKeyHash(string(outcoinsPrefix), tokenID)
	key = append(key, shardID)

	key = append(key, publicKey...)
	batchData := []incdb.BatchData{}
	for _, outputCoin := range outputCoinArr {
		keyTemp := make([]byte, len(key))
		copy(keyTemp, key)
		keyTemp = append(keyTemp, common.HashB(outputCoin)...)
		// Put to batch
		batchData = append(batchData, incdb.BatchData{
			Key:   keyTemp,
			Value: outputCoin,
		})
	}
	if len(batchData) > 0 {
		err := db.PutBatch(batchData)
		if err != nil {
			return incdb.NewDatabaseError(incdb.StoreOutputCoinsError, err)
		}
	}

	return nil
}

// StoreCommitments - store list commitments by shardID
func StoreCommitments(db incdb.Database, tokenID common.Hash, pubkey []byte, commitments [][]byte, shardID byte) error {
	key := addPrefixToKeyHash(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)

	// keySpec3 store last index of array commitment
	keySpec3 := make([]byte, len(key)+len("len"))
	temp := append(key, []byte("len")...)
	copy(keySpec3, temp)

	var lenData uint64
	lenCommitment, err := GetCommitmentLength(db, tokenID, shardID)
	if err != nil && lenCommitment == nil {
		return incdb.NewDatabaseError(incdb.StoreCommitmentsError, err)
	}
	if lenCommitment == nil {
		lenData = 0
	} else {
		lenData = lenCommitment.Uint64()
	}
	for _, c := range commitments {

		newIndex := new(big.Int).SetUint64(lenData).Bytes()
		if lenData == 0 {
			newIndex = []byte{0}
		}
		// keySpec1 use for create proof random
		keySpec1 := append(key, newIndex...)
		if err := db.Put(keySpec1, c); err != nil {
			return incdb.NewDatabaseError(incdb.StoreCommitmentsError, err)
		}
		// keySpec2 use for validate
		keySpec2 := append(key, c...)
		if err := db.Put(keySpec2, newIndex); err != nil {
			return incdb.NewDatabaseError(incdb.StoreCommitmentsError, err)
		}

		// len of commitment array
		if err := db.Put(keySpec3, newIndex); err != nil {
			return incdb.NewDatabaseError(incdb.StoreCommitmentsError, err)
		}
		lenData++
	}

	return nil
}

// HasCommitment - Check commitment in list commitments by shardID
func HasCommitment(db incdb.Database, tokenID common.Hash, commitment []byte, shardID byte) (bool, error) {
	key := addPrefixToKeyHash(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, commitment...)
	hasValue, err := db.Has(keySpec)
	if err != nil {
		return false, incdb.NewDatabaseError(incdb.HasCommitmentError, err, commitment, shardID, tokenID.String())
	} else {
		return hasValue, nil
	}
}

// ListCommitment -  return all commitment and its index
func ListCommitment(db incdb.Database, tokenID common.Hash, shardID byte) (map[string]uint64, error) {
	result := make(map[string]uint64)
	key := addPrefixToKeyHash(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)

	iterator := db.NewIteratorWithPrefix(key)
	for iterator.Next() {
		key1 := make([]byte, len(iterator.Key()))
		copy(key1, iterator.Key())
		if string(key1[len(key1)-3:]) == "len" {
			continue
		}
		if len(key1) < len(key)+privacy.Ed25519KeySize {
			continue
		}
		commitmentInByte := key1[len(key1)-privacy.Ed25519KeySize:]
		value := make([]byte, len(iterator.Value()))
		copy(value, iterator.Value())
		index := big.Int{}
		index.SetBytes(value)
		commitment := base58.Base58Check{}.Encode(commitmentInByte, 0x0)
		result[commitment] = index.Uint64()
	}
	return result, nil
}

// ListCommitmentIndices -  return all commitment index and its value
func ListCommitmentIndices(db incdb.Database, tokenID common.Hash, shardID byte) (map[uint64]string, error) {
	result := make(map[uint64]string)
	key := addPrefixToKeyHash(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)

	iterator := db.NewIteratorWithPrefix(key)
	for iterator.Next() {
		key1 := make([]byte, len(iterator.Key()))
		copy(key1, iterator.Key())
		if string(key1[len(key1)-3:]) == "len" {
			continue
		}

		commitmentInByte := make([]byte, len(iterator.Value()))
		copy(commitmentInByte, iterator.Value())
		if len(commitmentInByte) != privacy.Ed25519KeySize {
			continue
		}
		indexInByte := key1[45:]
		index := big.Int{}
		index.SetBytes(indexInByte)
		commitment := base58.Base58Check{}.Encode(commitmentInByte, 0x0)
		result[index.Uint64()] = commitment
	}
	return result, nil
}

func HasCommitmentIndex(db incdb.Database, tokenID common.Hash, commitmentIndex uint64, shardID byte) (bool, error) {
	key := addPrefixToKeyHash(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	var keySpec []byte
	if commitmentIndex == 0 {
		keySpec = append(key, byte(0))
	} else {
		keySpec = append(key, new(big.Int).SetUint64(commitmentIndex).Bytes()...)
	}
	_, err := db.Get(keySpec)
	if err != nil {
		return false, incdb.NewDatabaseError(incdb.HasCommitmentInexError, err, commitmentIndex, shardID, tokenID)
	} else {
		return true, nil
	}
}

func GetCommitmentByIndex(db incdb.Database, tokenID common.Hash, commitmentIndex uint64, shardID byte) ([]byte, error) {
	key := addPrefixToKeyHash(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	var keySpec []byte
	if commitmentIndex == 0 {
		keySpec = append(key, byte(0))
	} else {
		keySpec = append(key, new(big.Int).SetUint64(commitmentIndex).Bytes()...)
	}
	data, err := db.Get(keySpec)
	if err != nil {
		return data, incdb.NewDatabaseError(incdb.GetCommitmentByIndexError, err, commitmentIndex, shardID, tokenID)
	} else {
		return data, nil
	}
}

// GetCommitmentIndex - return index of commitment in db list
func GetCommitmentIndex(db incdb.Database, tokenID common.Hash, commitment []byte, shardID byte) (*big.Int, error) {
	key := addPrefixToKeyHash(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, commitment...)
	data, err := db.Get(keySpec)
	if err != nil {
		return nil, incdb.NewDatabaseError(incdb.GetCommitmentIndexError, err, commitment, shardID, tokenID)
	} else {
		return new(big.Int).SetBytes(data), nil
	}
}

// GetCommitmentIndex - return index of commitment in db list
func GetCommitmentLength(db incdb.Database, tokenID common.Hash, shardID byte) (*big.Int, error) {
	key := addPrefixToKeyHash(string(commitmentsPrefix), tokenID)
	key = append(key, shardID)
	keySpec := append(key, []byte("len")...)
	hasValue, err := db.Has(keySpec)
	if err != nil {
		return nil, incdb.NewDatabaseError(incdb.GetCommitmentLengthError, err)
	} else {
		if !hasValue {
			return nil, nil
		} else {
			data, err := db.Get(keySpec)
			if err != nil {
				return nil, incdb.NewDatabaseError(incdb.GetCommitmentLengthError, err)
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
func GetOutcoinsByPubkey(db incdb.Database, tokenID common.Hash, pubkey []byte, shardID byte) ([][]byte, error) {
	key := addPrefixToKeyHash(string(outcoinsPrefix), tokenID)
	key = append(key, shardID)

	key = append(key, pubkey...)
	arrDatabyPubkey := make([][]byte, 0)
	iter := db.NewIteratorWithPrefix(key)
	if iter.Error() != nil {
		return nil, incdb.NewDatabaseError(incdb.GetOutputCoinByPublicKeyError, errors.Wrap(iter.Error(), "db.lvdb.NewIterator"))
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
func CleanCommitments(db incdb.Database) error {
	iter := db.NewIteratorWithPrefix(commitmentsPrefix)
	for iter.Next() {
		err := db.Delete(iter.Key())
		if err != nil {
			return incdb.NewDatabaseError(incdb.CleanCommitmentError, err)
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return incdb.NewDatabaseError(incdb.CleanCommitmentError, err)
	}
	return nil
}

// HasSNDerivator - Check SnDerivator in list SnDerivators by shardID
func HasSNDerivator(db incdb.Database, tokenID common.Hash, data []byte) (bool, error) {
	key := addPrefixToKeyHash(string(snderivatorsPrefix), tokenID)
	keySpec := append(key, data...)
	hasValue, err := db.Has(keySpec)
	if err != nil {
		return false, incdb.NewDatabaseError(incdb.HasSNDerivatorError, err, data, -1, tokenID)
	} else {
		return hasValue, nil
	}
}

// StoreSNDerivators - store list serialNumbers by shardID
func StoreSNDerivators(db incdb.Database, tokenID common.Hash, sndArray [][]byte) error {
	key := addPrefixToKeyHash(string(snderivatorsPrefix), tokenID)

	// "snderivator-data:nil"
	batchData := []incdb.BatchData{}
	for _, snd := range sndArray {
		keySpec := make([]byte, len(key))
		copy(keySpec, key)
		keySpec = append(keySpec, snd...)
		batchData = append(batchData, incdb.BatchData{
			Key:   keySpec,
			Value: []byte{},
		})
	}
	if len(batchData) > 0 {
		err := db.PutBatch(batchData)
		if err != nil {
			return incdb.NewDatabaseError(incdb.StoreSNDerivatorsError, err)
		}
	}
	return nil
}

func ListSNDerivator(db incdb.Database, tokenID common.Hash) ([][]byte, error) {
	result := make([][]byte, 0)
	key := addPrefixToKeyHash(string(snderivatorsPrefix), tokenID)

	iterator := db.NewIteratorWithPrefix(key)
	for iterator.Next() {
		key1 := make([]byte, len(iterator.Key()))
		copy(key1, iterator.Key())

		sndInByte := key1[len(key)-1:]
		result = append(result, sndInByte)
	}
	return result, nil
}

// CleanCommitments - clear all list commitments in DB
func CleanSNDerivator(db incdb.Database) error {
	iter := db.NewIteratorWithPrefix(snderivatorsPrefix)
	for iter.Next() {
		err := db.Delete(iter.Key())
		if err != nil {
			return incdb.NewDatabaseError(incdb.CleanSNDerivatorError, err)
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return incdb.NewDatabaseError(incdb.CleanSNDerivatorError, err)
	}
	return nil
}

// StoreFeeEstimator - Store data for FeeEstimator object
func StoreFeeEstimator(db incdb.Database, val []byte, shardID byte) error {
	if err := db.Put(append(feeEstimatorPrefix, shardID), val); err != nil {
		return incdb.NewDatabaseError(incdb.UnexpectedError, errors.Wrap(err, "StoreFeeEstimator"))
	}
	return nil
}

// GetFeeEstimator - Get data for FeeEstimator object as a json in byte format
func GetFeeEstimator(db incdb.Database, shardID byte) ([]byte, error) {
	b, err := db.Get(append(feeEstimatorPrefix, shardID))
	if err != nil {
		return nil, incdb.NewDatabaseError(incdb.UnexpectedError, errors.Wrap(err, "GetFeeEstimator"))
	}
	return b, err
}

// CleanFeeEstimator - Clear FeeEstimator
func CleanFeeEstimator(db incdb.Database) error {
	iter := db.NewIteratorWithPrefix(feeEstimatorPrefix)
	for iter.Next() {
		err := db.Delete(iter.Key())
		if err != nil {
			return incdb.NewDatabaseError(incdb.UnexpectedError, errors.Wrap(err, "CleanFeeEstimator"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return incdb.NewDatabaseError(incdb.UnexpectedError, errors.Wrap(err, "CleanFeeEstimator"))
	}
	return nil
}

/*
	StoreTransactionIndex
	Store tx detail location
  Key: prefixTx-txHash
	H: blockHash-blockIndex
*/
func StoreTransactionIndex(db incdb.Database, txId common.Hash, blockHash common.Hash, index int, bd *[]incdb.BatchData) error {
	key := string(transactionKeyPrefix) + txId.String()
	value := blockHash.String() + string(Splitter) + strconv.Itoa(index)

	if bd != nil {
		*bd = append(*bd, incdb.BatchData{[]byte(key), []byte(value)})
		return nil
	}
	if err := db.Put([]byte(key), []byte(value)); err != nil {
		return incdb.NewDatabaseError(incdb.StoreTransactionIndexError, err, txId.String(), blockHash.String(), index)
	}

	return nil
}

/*
  Get Transaction by ID
*/

func GetTransactionIndexById(db incdb.Database, txId common.Hash) (common.Hash, int, error) {
	key := string(transactionKeyPrefix) + txId.String()
	_, err := db.Has([]byte(key))
	if err != nil {
		return common.Hash{}, -1, incdb.NewDatabaseError(incdb.GetTransactionIndexByIdError, err, txId.String())
	}

	res, err := db.Get([]byte(key))
	if err != nil {
		return common.Hash{}, -1, incdb.NewDatabaseError(incdb.GetTransactionIndexByIdError, err, txId.String())
	}
	reses := strings.Split(string(res), (string(Splitter)))
	hash, err := common.Hash{}.NewHashFromStr(reses[0])
	if err != nil {
		return common.Hash{}, -1, incdb.NewDatabaseError(incdb.GetTransactionIndexByIdError, err, txId.String())
	}
	index, err := strconv.Atoi(reses[1])
	if err != nil {
		return common.Hash{}, -1, incdb.NewDatabaseError(incdb.GetTransactionIndexByIdError, err, txId.String())
	}
	return *hash, index, nil
}

// StoreTxByPublicKey - store txID by public key of receiver,
// use this data to get tx which send to receiver
func StoreTxByPublicKey(db incdb.Database, publicKey []byte, txID common.Hash, shardID byte) error {
	key := make([]byte, 0)
	key = append(key, publicKey...)       // 1st 33b bytes for pubkey
	key = append(key, txID.GetBytes()...) // 2nd 32 bytes fir txID which receiver get from
	key = append(key, shardID)            // 3nd 1 byte for shardID where sender send to receiver

	if err := db.Put(key, []byte{}); err != nil {
		incdb.Logger.Log.Debug("StoreTxByPublicKey", err)
		return incdb.NewDatabaseError(incdb.StoreTxByPublicKeyError, err, txID.String(), publicKey, shardID)
	}

	return nil
}

// GetTxByPublicKey -  from public key, use this function to get list all txID which someone send use by txID from any shardID
func GetTxByPublicKey(db incdb.Database, publicKey []byte) (map[byte][]common.Hash, error) {
	itertor := db.NewIteratorWithPrefix(publicKey)
	result := make(map[byte][]common.Hash)
	for itertor.Next() {
		iKey := itertor.Key()
		key := make([]byte, len(iKey))
		copy(key, iKey)
		shardID := key[len(key)-1]
		if result[shardID] == nil {
			result[shardID] = make([]common.Hash, 0)
		}
		txID := common.Hash{}
		err := txID.SetBytes(key[common.PublicKeySize : common.PublicKeySize+common.HashSize])
		if err != nil {
			incdb.Logger.Log.Debugf("Err at GetTxByPublicKey", err)
			return nil, incdb.NewDatabaseError(incdb.GetTxByPublicKeyError, err, publicKey)
		}
		result[shardID] = append(result[shardID], txID)
	}
	return result, nil
}
