package lvdb

import (
	"encoding/binary"
	"fmt"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"strconv"
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
	Key: token-paymentAddress-{tokenId}-{paymentAddress}-txHash-voutIndex
  Value: value-spent/unspent-rewarded/unreward

*/
func (db *db) StoreCustomTokenPaymentAddresstHistory(tokenID *common.Hash, tx *transaction.TxCustomToken) error {
	tokenKey := string(tokenPaymentAddressPrefix) + tokenID.String()
	for _, vin := range tx.TxTokenData.Vins {
		paymentAddress := string(vin.PaymentAddress.ToBytes())
		utxoHash := vin.Hash().String()
		voutIndex := vin.VoutIndex
		paymentAddressKey := tokenKey + string(spliter) + paymentAddress + string(spliter) + utxoHash + string(spliter) + strconv.Itoa(voutIndex)
		_, err := db.hasValue([]byte(paymentAddressKey))
		if err != nil {
			fmt.Println("ERROR finding vin in DB, StoreCustomTokenAccountHistory", tx.Hash(), err)
			return err
		}
		value, err := db.lvdb.Get([]byte(paymentAddressKey), nil)
		if err != nil {
			return err
		}
		// old value: {value}-unspent-unreward/reward
		values := strings.Split(string(value),string(spliter))
		// new value: {value}-spent-unreward/reward
		newValues := values[0] + string(spliter) + string(spent) + string(spliter) +  values[2]
		if err := db.lvdb.Put([]byte(paymentAddressKey), []byte(newValues), nil); err != nil {
			return err
		}
	}
	for _, vout := range tx.TxTokenData.Vouts {
		tokenKey := string(tokenPaymentAddressPrefix) + tokenID.String()
		paymentAddress := string(vout.PaymentAddress.ToBytes())
		utxoHash := vout.Hash().String()
		voutIndex := vout.GetIndex()
		value := vout.Value
		paymentAddressKey := tokenKey + string(spliter) + paymentAddress + string(spliter) + utxoHash + string(spliter) + strconv.Itoa(voutIndex)
		_, err := db.hasValue([]byte(paymentAddressKey))
		if err != nil {
			fmt.Println("ERROR finding vout in DB, StoreCustomTokenAccountHistory", tx.Hash(), err)
			return err
		}
		// init value: {value}-unspent-unreward
		paymentAddressValue := string(value) + string(spliter) + string(unspent) + string(spliter) + string(unreward)
		if err := db.lvdb.Put([]byte(paymentAddressKey), []byte(paymentAddressValue), nil); err != nil {
			return err
		}
	}
	return nil
}

/*
	Return a list of all address with balance > 0
*/
func (db *db) GetCustomTokenListPaymentAddress(tokenID *common.Hash) ([][]byte, error){
	results := [][]byte{}
	//tempResults := make(map[string]int)
	tempsResult := make(map[string]bool)
	prefix := string(tokenPaymentAddressPrefix) + tokenID.String()
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

/*
	Return a list of all address with balance > 0
*/
func (db *db) GetCustomTokenListPaymentAddressesBalance(tokenID *common.Hash) (map[client.PaymentAddress]int, error){
	results := make(map[client.PaymentAddress]int)
	//tempsResult := make(map[string]bool)
	prefix := string(tokenPaymentAddressPrefix) + tokenID.String()
	iter := db.lvdb.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(spliter))
		values := strings.Split(value, string(value))
		if strings.Compare(values[1],string(unspent)) == 0 {
			// Uncomment this to get balance of all account
			paymentAddress := client.PaymentAddress{}
			paymentAddress.FromBytes([]byte(keys[3]))
			i, ok := results[paymentAddress]
			if ok == false {
				fmt.Println("ERROR geting value in GetCustomTokenAccountHistory of account", paymentAddress)
			}
			balance,_ := strconv.Atoi(values[0])
			i += balance
			results[paymentAddress] = i

		}
	}
	iter.Release()
	return results, nil
}
/*
	Get a list of UTXO that can be reward all payment address
	key: payment address
	value: a list of utxo
	Each utxo consist of two part: first 32 byte is txHash, the rest is voutIndex
*/
func (db *db) GetCustomTokenListUnrewardUTXO(tokenID *common.Hash) (map[client.PaymentAddress][][]byte, error){

	results := make(map[client.PaymentAddress][][]byte)
	prefix := string(tokenPaymentAddressPrefix) + tokenID.String()
	iter := db.lvdb.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(spliter))
		values := strings.Split(value, string(value))
		// get unspent and unreward transaction output
		if (strings.Compare(values[1],string(unspent)) == 0) && (strings.Compare(values[2],string(unreward)) == 0) {
			paymentAddress := client.PaymentAddress{}
			paymentAddress.FromBytes([]byte(keys[3]))
			//utxo := keys[4] + string(spliter) + keys[5]
			utxo := append([]byte(keys[4]), []byte(keys[5])[:]... )
			results[paymentAddress] = append(results[paymentAddress], utxo)
		}
	}
	iter.Release()
	return results, nil
}
/*
	Get a list of UTXO of one address
	Return a list of UTXO, each UTXO has format: first 32 byte is txHash, the rest is voutIndex
*/
func (db *db) GetCustomTokenPaymentAddressUTXO(tokenID *common.Hash, paymentAddress client.PaymentAddress) ([][]byte, error){
	prefix := string(tokenPaymentAddressPrefix) + tokenID.String() + string(spliter) + string(paymentAddress.ToBytes())
	results := [][]byte{}
	iter := db.lvdb.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(spliter))
		values := strings.Split(value, string(value))
		// get unspent and unreward transaction output
		if (strings.Compare(values[1],string(unspent)) == 0) {
			//utxo := keys[4] + string(spliter) + keys[5]
			utxo := append([]byte(keys[4]), []byte(keys[5])[:]... )
			results = append(results, utxo)
		}
	}
	iter.Release()
	return results, nil
}
/*
	Update UTXO from unreward -> reward
*/
func (db *db) UpdateRewardAccountUTXO(tokenID *common.Hash, paymentAddress client.PaymentAddress, txHash *common.Hash, voutIndex int) (error){
	key := string(tokenPaymentAddressPrefix) + tokenID.String() + string(spliter) + string(paymentAddress.ToBytes()) + string(spliter) + txHash.String() + string(spliter) + strconv.Itoa(voutIndex)
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
