package lvdb

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common"

	"github.com/ninjadotorg/constant/database"

	"math/big"

	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// StoreSerialNumbers - store list serialNumbers by chainID
func (db *db) StoreSerialNumbers(tokenID *common.Hash, serialNumber []byte, chainId byte) error {
	key := db.GetKey(string(serialNumbersPrefix), tokenID)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var arrayData [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &arrayData); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}

	lenData := int64(len(arrayData))
	newIndex := big.NewInt(lenData).Bytes()
	if lenData == 0 {
		newIndex = []byte{0}
	}
	keySpec1 := make([]byte, len(key))
	keySpec1 = append(key, serialNumber...)
	if err := db.lvdb.Put(keySpec1, newIndex, nil); err != nil {
		return err
	}

	arrayData = append(arrayData, serialNumber)
	b, err := json.Marshal(arrayData)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

// FetchSerialNumbers - Get list SerialNumbers by chainID
func (db *db) FetchSerialNumbers(tokenID *common.Hash, chainID byte) ([][]byte, error) {
	key := db.GetKey(string(serialNumbersPrefix), tokenID)
	key = append(key, chainID)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var arrayData [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &arrayData); err != nil {
			return make([][]byte, 0), errors.Wrap(err, "json.Unmarshal")
		}
	}
	return arrayData, nil
}

// HasSerialNumber - Check serialNumber in list SerialNumbers by chainID
func (db *db) HasSerialNumber(tokenID *common.Hash, serialNumber []byte, chainID byte) (bool, error) {
	key := db.GetKey(string(serialNumbersPrefix), tokenID)
	key = append(key, chainID)
	keySpec := append(key, serialNumber...)
	_, err := db.Get(keySpec)
	if err != nil {
		return false, nil
	} else {
		return true, nil
	}
	return false, nil
}

// HasSerialNumberIndex - Check serialNumber in list SerialNumbers by chainID
/*func (db *db) HasSerialNumberIndex(serialNumberIndex int64, chainID byte) (bool, error) {
	key := db.GetKey(string(serialNumbersPrefix), "")
	key = append(key, chainID)
	keySpec := append(key, big.NewInt(serialNumberIndex).Bytes()...)
	_, err := db.Get(keySpec)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
	return false, nil
}*/

/*func (db *db) GetSerialNumberByIndex(serialNumberIndex int64, chainID byte) ([]byte, error) {
	key := db.GetKey(string(serialNumbersPrefix), "")
	key = append(key, chainID)
	keySpec := append(key, big.NewInt(serialNumberIndex).Bytes()...)
	data, err := db.Get(keySpec)
	if err != nil {
		return data, err
	} else {
		return data, nil
	}
	return data, nil
}*/

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

func (db *db) StoreOutputCoins(tokenID *common.Hash, pubkey []byte, outputcoin []byte, chainID byte) error {
	key := db.GetKey(string(outcoinsPrefix), tokenID)
	key = append(key, chainID)

	// store for pubkey:[outcoint1, outcoint2, ...]
	key = append(key, pubkey...)
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
	arrDatabyPubkey = append(arrDatabyPubkey, outputcoin)
	resByPubkey, err = json.Marshal(arrDatabyPubkey)
	if err := db.lvdb.Put(key, resByPubkey, nil); err != nil {
		return err
	}

	return nil
}

// StoreCommitments - store list commitments by chainID
func (db *db) StoreCommitments(tokenID *common.Hash, pubkey []byte, commitments []byte, chainId byte) error {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var arrData [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &arrData); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}

	// use for create proof random
	lenData := uint64(len(arrData))
	newIndex := new(big.Int).SetUint64(lenData).Bytes()
	if lenData == 0 {
		newIndex = []byte{0}
	}
	keySpec1 := make([]byte, len(key))
	keySpec1 = append(key, newIndex...)
	if err := db.lvdb.Put(keySpec1, commitments, nil); err != nil {
		return err
	}

	// use for validate
	keySpec2 := make([]byte, len(key))
	keySpec2 = append(key, commitments...)
	if err := db.lvdb.Put(keySpec2, newIndex, nil); err != nil {
		return err
	}

	// store length of array commitment
	keySpec3 := make([]byte, len(key))
	keySpec3 = append(key, []byte("len")...)
	if err := db.lvdb.Put(keySpec3, newIndex, nil); err != nil {
		return err
	}

	// store for pubkey:[newindex1, newindex2]
	keySpec4 := make([]byte, len(key))
	keySpec4 = append(key, pubkey...)
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
	arrDatabyPubkey = append(arrDatabyPubkey, newIndex)
	resByPubkey, err = json.Marshal(arrDatabyPubkey)
	if err := db.lvdb.Put(keySpec4, resByPubkey, nil); err != nil {
		return err
	}

	arrData = append(arrData, commitments)
	b, err := json.Marshal(arrData)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

// FetchCommitments - Get list commitments by chainID
func (db *db) FetchCommitments(tokenID *common.Hash, chainId byte) ([][]byte, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	return txs, nil
}

// HasCommitment - Check commitment in list commitments by chainID
func (db *db) HasCommitment(tokenID *common.Hash, commitment []byte, chainId byte) (bool, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, chainId)
	keySpec := append(key, commitment...)
	_, err := db.Get(keySpec)
	if err != nil {
		return false, nil
	} else {
		return true, nil
	}
	return false, nil
}

func (db *db) HasCommitmentIndex(tokenID *common.Hash, commitmentIndex uint64, chainId byte) (bool, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, chainId)
	keySpec := append(key, new(big.Int).SetUint64(commitmentIndex).Bytes()...)
	_, err := db.Get(keySpec)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
	return false, nil
}

func (db *db) GetCommitmentByIndex(tokenID *common.Hash, commitmentIndex uint64, chainId byte) ([]byte, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, chainId)
	keySpec := make([]byte, len(key))
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
func (db *db) GetCommitmentIndex(tokenID *common.Hash, commitment []byte, chainId byte) (*big.Int, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, chainId)
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
func (db *db) GetCommitmentLength(tokenID *common.Hash, chainId byte) (*big.Int, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, chainId)
	keySpec := append(key, []byte("len")...)
	data, err := db.Get(keySpec)
	if err != nil {
		return nil, err
	} else {
		lenArray := new(big.Int).SetBytes(data)
		lenArray = lenArray.Add(lenArray, new(big.Int).SetInt64(1))
		return lenArray, nil
	}
	return nil, nil
}

func (db *db) GetCommitmentIndexsByPubkey(tokenID *common.Hash, pubkey []byte, chainID byte) ([][]byte, error) {
	key := db.GetKey(string(commitmentsPrefix), tokenID)
	key = append(key, chainID)

	keySpec4 := make([]byte, len(key))
	keySpec4 = append(key, pubkey...)
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

func (db *db) GetOutcoinsByPubkey(tokenID *common.Hash, pubkey []byte, chainID byte) ([][]byte, error) {
	key := db.GetKey(string(outcoinsPrefix), tokenID)
	key = append(key, chainID)

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
}

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

// StoreSNDerivators - store list serialNumbers by chainID
func (db *db) StoreSNDerivators(tokenID *common.Hash, data big.Int, chainID byte) error {
	key := db.GetKey(string(snderivatorsPrefix), tokenID)
	key = append(key, chainID)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	// "snderivator-data:data"
	snderivatorData := data.Bytes()
	keySpec := make([]byte, len(key))
	keySpec = append(key, snderivatorData...)
	if err := db.lvdb.Put(keySpec, snderivatorData, nil); err != nil {
		return err
	}

	var arrData []string
	if len(res) > 0 {
		if err := json.Unmarshal(res, &arrData); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	arrData = append(arrData, string(snderivatorData))
	b, err := json.Marshal(arrData)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

// FetchSerialNumbers - Get list all SnDerivators by chainID
func (db *db) FetchSNDerivator(tokenID *common.Hash, chainID byte) ([]big.Int, error) {
	key := db.GetKey(string(snderivatorsPrefix), tokenID)
	key = append(key, chainID)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([]big.Int, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var arrData []string
	if len(res) > 0 {
		if err := json.Unmarshal(res, &arrData); err != nil {
			return make([]big.Int, 0), errors.Wrap(err, "json.Unmarshal")
		}
	}
	result := []big.Int{}
	for _, data := range arrData {
		temp := big.Int{}
		temp.SetBytes([]byte(data))
		result = append(result, temp)
	}
	return result, nil
}

// HasSNDerivator - Check SnDerivator in list SnDerivators by chainID
func (db *db) HasSNDerivator(tokenID *common.Hash, data big.Int, chainID byte) (bool, error) {
	key := db.GetKey(string(snderivatorsPrefix), tokenID)
	key = append(key, chainID)
	snderivatorData := data.Bytes()
	keySpec := append(key, snderivatorData...)
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
func (db *db) StoreFeeEstimator(val []byte, chainId byte) error {
	if err := db.Put(append(feeEstimator, chainId), val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

// GetFeeEstimator - Get data for FeeEstimator object as a json in byte format
func (db *db) GetFeeEstimator(chainId byte) ([]byte, error) {
	b, err := db.lvdb.Get(append(feeEstimator, chainId), nil)
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
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
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
func (db *db) StoreTransactionIndex(txId *common.Hash, blockHash *common.Hash, index int) error {
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

func (db *db) GetTransactionIndexById(txId *common.Hash) (*common.Hash, int, error) {
	key := string(transactionKeyPrefix) + txId.String()
	_, err := db.HasValue([]byte(key))
	if err != nil {
		return nil, -1, err
	}

	res, err := db.lvdb.Get([]byte(key), nil)
	if err != nil {
		return nil, -1, err
	}
	reses := strings.Split(string(res), (string(Splitter)))
	hash, err := common.Hash{}.NewHashFromStr(reses[0])
	if err != nil {
		return nil, -1, err
	}
	index, err := strconv.Atoi(reses[1])
	if err != nil {
		return nil, -1, err
	}
	return hash, index, nil
}

/*
	Store Transaction in LightMode mode
	1. Key -> value : prefix(privateky)privateKey-[-]-chainId-[-]-(999999999 - blockHeight)-[-]-(999999999 - txIndex) 		-> 		tx
	2. Key -> value :							prefix(transaction)txHash 												->  	privateKey-chainId-blockHeight-txIndex

*/
func (db *db) StoreTransactionLightMode(privateKey *privacy.SpendingKey, chainId byte, blockHeight int32, txIndex int, unspentTxHash common.Hash, unspentTx []byte) error {
	tempChainId := []byte{}
	tempChainId = append(tempChainId, chainId)
	temp3ChainId := int(chainId)
	temp2ChainId := string(int(chainId))
	fmt.Println("StoreTransactionLightMode", privateKey, temp3ChainId, temp2ChainId, blockHeight, txIndex)
	reverseBlockHeight := make([]byte, 4)
	binary.LittleEndian.PutUint32(reverseBlockHeight, uint32(bigNumber-blockHeight))

	// Uncomment this to test little endian encoding and decoding
	/*
		buf := bytes.NewBuffer(reverseBlockHeight)
		var temp uint32
		err1 := binary.Read(buf, binary.LittleEndian, &temp)
		if err1 != nil {
			return err1
		}
		fmt.Println("Testing encoding and decoding uint32 little endian", uint32(999999999-temp), blockHeight)
	*/

	//fmt.Println("StoreTransactionLightMode reverseBlockHeight in byte", reverseBlockHeight, []byte(string(reverseBlockHeight)))

	reverseTxIndex := make([]byte, 4)
	binary.LittleEndian.PutUint32(reverseTxIndex, uint32(bigNumberTx-int32(txIndex)))

	key1 := string(privateKeyPrefix) + privateKey.String() + string(Splitter) + string(int(chainId)) + string(Splitter) + string(reverseBlockHeight) + string(Splitter) + string(reverseTxIndex)
	key2 := string(transactionKeyPrefix) + unspentTxHash.String()

	if ok, _ := db.HasValue([]byte(key1)); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("tx %s already exists", key1))
	}
	if ok, _ := db.HasValue([]byte(key2)); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("tx %s already exists", key2))
	}

	/*value, err := json.Marshal(unspentTx)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}*/
	value := unspentTx
	if err := db.Put([]byte(key1), value); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.Put([]byte(key2), []byte(key1)); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}

	return nil
}

/*
	Get Transaction in LightMode mode
	Get transaction by prefix(privateKey)privateKey, this prefix help to get all transaction belong to that privatekey
	1. Key -> value : prefix(privateky)-privateKey-chainId-(999999999 - blockHeight)-(999999999 - txIndex) 		-> 		tx

*/
/*func (db *db) GetTransactionLightModeByPrivateKey(privateKey *privacy.SpendingKey) (map[byte]([]([]byte)), error) {
	prefix := []byte(string(privateKeyPrefix) + privateKey.String())
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)

	results := make(map[byte]([]([]byte)))
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		reses := strings.Split(string(key), string(Splitter))
		tempChainId, _ := strconv.Atoi(reses[2])
		chainId := byte(tempChainId)
		*//*tx := transaction.Tx{}
		err := json.Unmarshal(value, &tx)
		if err != nil {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
		}*//*
		data := make([]byte, len(value))
		copy(data[:], value[:])
		results[chainId] = append(results[chainId], data)
	}

	iter.Release()
	return results, nil
}*/

/*
	Key: transactionPrefix-txHash
  H: txLocation
  tx: tx object in byte
*/
/*func (db *db) GetTransactionLightModeByHash(txId *common.Hash) ([]byte, []byte, error) {
	key := string(transactionKeyPrefix) + txId.String()
	_, err := db.HasValue([]byte(key))
	if err != nil {
		return nil, nil, err
	}
	value, err := db.lvdb.Get([]byte(key), nil)
	if err != nil {
		return nil, nil, err
	}
	_, err1 := db.HasValue([]byte(value))
	if err1 != nil {
		return nil, nil, err
	}
	tx, err := db.lvdb.Get([]byte(value), nil)
	if err != nil {
		return nil, nil, err
	}
	return value, tx, nil
}*/
