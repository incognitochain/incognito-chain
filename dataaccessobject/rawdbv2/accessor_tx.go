package rawdbv2

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreTransactionIndex(db incdb.KeyValueWriter, txHash common.Hash, blockHash common.Hash, index int) error {
	key := GetTransactionHashKey(txHash)
	value := []byte(blockHash.String() + string(splitter) + strconv.Itoa(index))
	if err := db.Put(key, value); err != nil {
		return NewRawdbError(StoreTransactionIndexError, err)
	}
	return nil
}

func GetTransactionByHash(db incdb.Database, txHash common.Hash) (common.Hash, int, error) {
	key := GetTransactionHashKey(txHash)
	if has, err := db.Has(key); err != nil {
		return common.Hash{}, 0, NewRawdbError(GetTransactionByHashError, err)
	} else if !has {
		return common.Hash{}, 0, NewRawdbError(GetTransactionByHashError, fmt.Errorf("transaction %+v not found", txHash))
	}
	value, err := db.Get(key)
	if err != nil {
		return common.Hash{}, 0, NewRawdbError(GetTransactionByHashError, err)
	}
	strs := strings.Split(string(value), string(splitter))
	newHash := common.Hash{}
	blockHash, err := newHash.NewHashFromStr(strs[0])
	if err != nil {
		return common.Hash{}, 0, NewRawdbError(GetTransactionByHashError, err)
	}
	index, err := strconv.Atoi(strs[1])
	if err != nil {
		return common.Hash{}, 0, NewRawdbError(GetTransactionByHashError, err)
	}
	return *blockHash, index, nil
}

func DeleteTransactionIndex(db incdb.Database, txHash common.Hash) error {
	key := GetTransactionHashKey(txHash)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(DeleteTransactionByHashError, err)
	}
	return nil
}

// StoreTxByPublicKey - store txID by public key of receiver,
// use this data to get tx which send to receiver
// key format:
// 1st 33b bytes for pubkey
// 2nd 32 bytes fir txID which receiver get from
// 3nd 1 byte for shardID where sender send to receiver
func StoreTxByPublicKey(db incdb.Database, publicKey []byte, txID common.Hash, shardID byte) error {
	key := GetStoreTxByPublicKey(publicKey, txID, shardID)
	value := []byte{}
	if err := db.Put(key, value); err != nil {
		return NewRawdbError(StoreTxByPublicKeyError, err, txID.String(), publicKey, shardID)
	}
	return nil
}

// GetTxByPublicKey -  from public key, use this function to get list all txID which someone send use by txID from any shardID
func GetTxByPublicKey(db incdb.Database, publicKey []byte) (map[byte][]common.Hash, error) {
	iterator := db.NewIteratorWithPrefix(GetStoreTxByPublicPrefix(publicKey))
	result := make(map[byte][]common.Hash)
	for iterator.Next() {
		key := iterator.Key()
		tempKey := make([]byte, len(key))
		copy(tempKey, key)
		shardID := tempKey[len(tempKey)-1]
		if result[shardID] == nil {
			result[shardID] = make([]common.Hash, 0)
		}
		txID := common.Hash{}
		start := len(txByPublicKeyPrefix) + common.PublicKeySize
		end := len(txByPublicKeyPrefix) + common.PublicKeySize + common.HashSize
		err := txID.SetBytes(tempKey[start:end])
		if err != nil {
			return nil, NewRawdbError(GetTxByPublicKeyError, err, publicKey)
		}
		result[shardID] = append(result[shardID], txID)
	}
	return result, nil
}

// GetTxByPublicKeyV2 returns list of all tx IDs in paging fashion for a given public key
func GetTxByPublicKeyV2(
	db incdb.Database, publicKey []byte,
	skip, limit uint,
) (map[byte][]common.Hash, uint, uint, error) {
	iterator := db.NewIteratorWithPrefix(GetStoreTxByPublicPrefix(publicKey))
	result := make(map[byte][]common.Hash)
	for iterator.Next() {
		if skip > 0 {
			skip--
			continue
		}
		if limit == 0 {
			return result, skip, limit, nil
		}
		key := iterator.Key()
		tempKey := make([]byte, len(key))
		copy(tempKey, key)
		shardID := tempKey[len(tempKey)-1]
		if result[shardID] == nil {
			result[shardID] = make([]common.Hash, 0)
		}
		txID := common.Hash{}
		start := len(txByPublicKeyPrefix) + common.PublicKeySize
		end := len(txByPublicKeyPrefix) + common.PublicKeySize + common.HashSize
		err := txID.SetBytes(tempKey[start:end])
		if err != nil {
			return nil, skip, limit, NewRawdbError(GetTxByPublicKeyError, err, publicKey)
		}
		result[shardID] = append(result[shardID], txID)
		limit--
	}
	return result, skip, limit, nil
}

func StoreIndexedOutCoins(db incdb.Database, tokenID common.Hash, publicKey []byte, outputCoins [][]byte, shardID byte) error {
	for _, outputCoin := range outputCoins {
		key := generateIndexedOutputCoinObjectKey(tokenID, shardID, publicKey, outputCoin)
		value := outputCoin
		err := db.Put(key, value)
		if err != nil {
			return NewRawdbError(StoreOutcoinByOTAKeyError, err)
		}
	}
	return nil
}

func GetOutCoinsByIndexedOTAKey(db incdb.Database, tokenID common.Hash, shardID byte, publicKey []byte) ([][]byte, error) {
	it := db.NewIteratorWithPrefix(getIndexedOutputCoinPrefix(tokenID, shardID, publicKey))
	var outputCoins [][]byte
	for it.Next() {
		value := it.Value()
		newValue := make([]byte, len(value))
		copy(newValue, value)
		outputCoins = append(outputCoins, newValue)
	}
	return outputCoins, nil
}

func StoreIndexedOTAKey(db incdb.Database, theKey []byte, state int) error {
	key := generateIndexedOTAKeyObjectKey(theKey)
	// only care about `PublicKey` field
	value := append(theKey, byte(state))
	err := db.Put(key, value)
	if err != nil {
		return NewRawdbError(StoreOTAKeyError, err)
	}
	return nil
}

func StoreWhitelistOTAKeys(db incdb.Database, whitelist [][OTAKeyBytesLength]byte) error {
	key := GetWhitelistOTAKey()
	if value, err := json.Marshal(whitelist); err != nil {
		return err
	} else {
		return db.Put(key, value)
	}
}

func GetWhitelistOTAKeys(db incdb.Database) ([][OTAKeyBytesLength]byte, error) {
	key := GetWhitelistOTAKey()
	value, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	whitelist := [][OTAKeyBytesLength]byte{}
	err = json.Unmarshal(value, &whitelist)
	return whitelist, err
}

func DeleteIndexedOTAKey(db incdb.Database, theKey []byte) error {
	key := generateIndexedOTAKeyObjectKey(theKey)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(DeleteOTAKeyError, err)
	}
	return nil
}

func GetIndexedOTAKeys(db incdb.Database) ([][]byte, error) {
	whitelistMap := map[[OTAKeyBytesLength]byte]interface{}{}
	whitelist, err := GetWhitelistOTAKeys(db)
	if err == nil {
		for _, kBytes := range whitelist {
			whitelistMap[kBytes] = nil
		}
	}
	it := db.NewIteratorWithPrefix(getIndexedKeysPrefix())
	var otaKeys [][]byte
	for it.Next() {
		value := it.Value()
		newValue := make([]byte, len(value))
		if len(whitelistMap) != 0 {
			otaKey := [OTAKeyBytesLength]byte{}
			copy(otaKey[:], newValue[:OTAKeyBytesLength])
			if _, has := whitelistMap[[64]byte(otaKey)]; !has {
				continue
			}
		}
		copy(newValue, value)
		otaKeys = append(otaKeys, newValue)
	}
	return otaKeys, nil
}

func StoreCachedCoinHash(db incdb.Database, theCoinHash []byte) error {
	key := generateCachedCoinHashObjectKey(theCoinHash)
	value := theCoinHash
	err := db.Put(key, value)
	if err != nil {
		return NewRawdbError(StoreCoinHashError, err)
	}

	return nil
}

func GetCachedCoinHashes(db incdb.Database) ([][]byte, error) {
	it := db.NewIteratorWithPrefix(getCoinHashKeysPrefix())
	var otaKeys [][]byte
	for it.Next() {
		value := it.Value()
		newValue := make([]byte, len(value))
		copy(newValue, value)
		otaKeys = append(otaKeys, newValue)
	}
	return otaKeys, nil
}

// These functions are used for storing and getting a transaction by an output coin index
// TODO: refactor these functions for more efficient read/write
func StoreTxByCoinIndex(db incdb.Database, index []byte, tokenID common.Hash, shardID byte, txID common.Hash) error {
	key := generateTxByCoinIndexObjectKey(index, tokenID, shardID)
	value := txID.Bytes()
	if err := db.Put(key, value); err != nil {
		return NewRawdbError(StoreTxByCoinIndexError, err, txID.String(), index, tokenID.String(), shardID)
	}
	return nil
}

func GetTxByCoinIndex(db incdb.Database, index []byte, tokenID common.Hash, shardID byte) (*common.Hash, error) {
	iterator := db.NewIteratorWithPrefix(generateTxByCoinIndexObjectKey(index, tokenID, shardID))
	if iterator.Next() {
		value := iterator.Value()
		txHash, err := new(common.Hash).NewHash(value)
		if err != nil {
			return nil, NewRawdbError(GetTxByCoinIndexError, err, index, tokenID.String(), shardID)
		}

		return txHash, nil
	}

	//If this is a token transaction, try with assetTag
	if tokenID.String() != common.PRVIDStr {
		iterator := db.NewIteratorWithPrefix(generateTxByCoinIndexObjectKey(index, common.ConfidentialAssetID, shardID))
		if iterator.Next() {
			value := iterator.Value()
			txHash, err := new(common.Hash).NewHash(value)
			if err != nil {
				return nil, NewRawdbError(GetTxByCoinIndexError, err, index, tokenID.String(), shardID)
			}

			return txHash, nil
		}
	}

	return nil, NewRawdbError(GetTxByCoinIndexError, fmt.Errorf("no tx found for index %v, tokenID %v, shardID %v", index, tokenID.String(), shardID))
}

// These functions are used for storing/retrieving a transaction by an input coin serial number
// TODO: refactor these functions for more efficient read/write
func StoreTxBySerialNumber(db incdb.Database, serialNumber []byte, tokenID common.Hash, shardID byte, txID common.Hash) error {
	key := generateTxBySerialNumberObjectKey(serialNumber, tokenID, shardID)
	value := txID.Bytes()
	if err := db.Put(key, value); err != nil {
		return NewRawdbError(StoreTxBySerialNumberError, err, txID.String(), serialNumber, tokenID.String(), shardID)
	}
	return nil
}

func GetTxBySerialNumber(db incdb.Database, serialNumber []byte, tokenID common.Hash, shardID byte) (*common.Hash, error) {
	iterator := db.NewIteratorWithPrefix(generateTxBySerialNumberObjectKey(serialNumber, tokenID, shardID))
	if iterator.Next() {
		value := iterator.Value()
		txHash, err := new(common.Hash).NewHash(value)
		if err != nil {
			return nil, NewRawdbError(GetTxBySerialNumberError, err, serialNumber, tokenID.String(), shardID)
		}

		return txHash, nil
	}

	return nil, NewRawdbError(GetTxBySerialNumberError, fmt.Errorf("no tx found for serialNumber %v, tokenID %v, shardID %v", serialNumber, tokenID.String(), shardID))
}
