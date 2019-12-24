package rawdbv2

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"strconv"
	"strings"
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
