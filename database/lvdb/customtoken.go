package lvdb

import (
	"encoding/binary"
	"fmt"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"strings"
)

func (db *db) StoreCustomToken(tokenID *common.Hash, txHash []byte) error {
	key := db.getKey(string(tokenInitPrefix), tokenID) // token-init-{tokenID}
	if err := db.lvdb.Put(key, txHash, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) StoreCustomTokenTx(tokenID *common.Hash, chainID byte, blockHeight int32, txIndex int32, txHash []byte) error {
	bigNumber := int32(999999999)
	key := db.getKey(string(tokenPrefix), tokenID) // token-{tokenID}-chainID-(999999999-blockHeight)-(999999999-txIndex)
	key = append(key, chainID)
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(bigNumber-blockHeight))
	key = append(key, bs...)
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(bigNumber-txIndex))
	key = append(key, bs...)
	log.Println(string(key))
	if err := db.lvdb.Put(key, txHash, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) ListCustomToken() ([][]byte, error) {
	result := make([][]byte, 0)
	iter := db.lvdb.NewIterator(util.BytesPrefix(tokenInitPrefix), nil)
	for iter.Next() {
		result = append(result, iter.Value())
	}
	iter.Release()
	return result, nil
}

func (db *db) CustomTokenTxs(tokenID *common.Hash) ([]*common.Hash, error) {
	result := make([]*common.Hash, 0)
	key := db.getKey(string(tokenPrefix), tokenID)
	// key = token-{tokenID}
	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	log.Println(string(key))
	for iter.Next() {
		value := iter.Value()
		hash := common.Hash{}
		hash.SetBytes(value)
		result = append(result, &hash)
	}
	iter.Release()
	return result, nil
}

/*
	Key: token-account-{tokenId}-{account}-txHash-voutIndex
  Value: value-spent/unspent-rewarded/unreward

*/
func (db *db) StoreCustomTokenAccountHistory(tokenID *common.Hash, tx *transaction.TxCustomToken) error {
	tokenKey := string(tokenAccountPrefix) + tokenID.String()
	for _, vin := range tx.TxTokenData.Vins {
		address := string(vin.PaymentAddress.ToBytes())
		utxoHash := vin.Hash().String()
		voutIndex := vin.VoutIndex
		accountKey := tokenKey + string(spliter) + address + string(spliter) + utxoHash + string(spliter) + string(voutIndex)
		_, err := db.hasValue([]byte(accountKey))
		if err != nil {
			fmt.Println("ERROR finding vin in DB, StoreCustomTokenAccountHistory", tx.Hash(), err)
			return err
		}
		value, err := db.lvdb.Get([]byte(accountKey), nil)
		if err != nil {
			return err
		}
		values := strings.Split(string(value),string(spliter))
		// {value}-spent-unreward
		newValues := values[0] + string(spliter) + string(spent) + string(spliter) +  values[2]
		if err := db.lvdb.Put([]byte(accountKey), []byte(newValues), nil); err != nil {
			return err
		}
	}
	for _, vout := range tx.TxTokenData.Vouts {
		tokenKey := string(tokenAccountPrefix) + tokenID.String()
		address := string(vout.PaymentAddress.ToBytes())
		utxoHash := vout.Hash().String()
		voutIndex := vout.GetIndex()
		value := vout.Value
		accountKey := tokenKey + string(spliter) + address + string(spliter) + utxoHash + string(spliter) + string(voutIndex)
		_, err := db.hasValue([]byte(accountKey))
		if err != nil {
			fmt.Println("ERROR finding vout in DB, StoreCustomTokenAccountHistory", tx.Hash(), err)
			return err
		}
		// {value}-unspent-unreward
		accountValue := string(value) + string(spliter) + string(unspent) + string(spliter) + string(unreward)
		if err := db.lvdb.Put([]byte(accountKey), []byte(accountValue), nil); err != nil {
			return err
		}
	}
	return nil
}

func (db *db) GetCustomTokenAccountHistory(tokenID *common.Hash) ([][]byte, error){
	results := [][]byte{}
	//tempResults := make(map[string]int)
	tempsResult := make(map[string]bool)
	prefix := string(tokenAccountPrefix) + tokenID.String()
	iter := db.lvdb.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(spliter))
		values := strings.Split(value, string(value))
		if strings.Compare(values[1],string(unspent)) == 0 {
			// Uncomment this to get balance of all account
			//i, ok := tempResults[keys[3]]
			//if ok == false {
			//	fmt.Println("ERROR geting value in GetCustomTokenAccountHistory of account", key[3])
			//}
			//values0,_ := strconv.Atoi(values[0])
			//i += values0
			//tempResults[keys[3]] = i

			tempsResult[keys[3]] = true
		}
	}
	for key, value := range tempsResult {
		if value == true {
			results = append(results, []byte(key))
		}
	}
	iter.Release()
	return results, nil
}

func (db *db) GetCustomTokenAccountUTXO(tokenID *common.Hash, address []byte) ([][]byte, error){
	results := [][]byte{}
	prefix := string(tokenAccountPrefix) + tokenID.String() + string(spliter) + string(address)
	iter := db.lvdb.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(spliter))
		values := strings.Split(value, string(value))
		if strings.Compare(values[1],string(unspent)) == 0 {
			utxo := keys[4] + string(spliter) + keys[5]
			results = append(results, []byte(utxo))
		}
	}
	iter.Release()
	return results, nil
}

func (db *db) UpdateRewardAccountUTXO(tokenID *common.Hash, address []byte, txHash *common.Hash, voutIndex int) (error){
	key := string(tokenAccountPrefix) + tokenID.String() + string(spliter) + string(address) + string(spliter) + txHash.String() + string(spliter) + string(voutIndex)
	_, err := db.hasValue([]byte(key))
	if err != nil {
		fmt.Println("ERROR finding key in DB, UpdateRewardAccountUTXO", err)
		return err
	}
	res, err := db.lvdb.Get([]byte(key), nil)
	if err != nil {
		return err
	}
	reses := strings.Split(string(res), string(spliter))
	// {value}-unspent-unreward
	value := reses[0] + reses[1] + string(rewared)
	if err := db.lvdb.Put([]byte(key), []byte(value), nil); err != nil {
		return err
	}
	return nil
}
