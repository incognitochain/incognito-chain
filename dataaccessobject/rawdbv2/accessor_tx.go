package rawdbv2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreTransactionIndex(db incdb.Database, txHash common.Hash, blockHash common.Hash, index int) error {
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
