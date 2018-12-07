package lvdb

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/voting"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) StoreCustomToken(tokenID *common.Hash, txHash []byte) error {
	key := db.GetKey(string(tokenInitPrefix), tokenID) // token-init-{tokenID}
	if err := db.lvdb.Put(key, txHash, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) StoreCustomTokenTx(tokenID *common.Hash, chainID byte, blockHeight int32, txIndex int32, txHash []byte) error {
	key := db.GetKey(string(tokenPrefix), tokenID) // token-{tokenID}-chainID-(999999999-blockHeight)-(999999999-txIndex)
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
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		result = append(result, value)
	}
	iter.Release()
	return result, nil
}

func (db *db) CustomTokenTxs(tokenID *common.Hash) ([]*common.Hash, error) {
	result := make([]*common.Hash, 0)
	key := db.GetKey(string(tokenPrefix), tokenID)
	// PubKey = token-{tokenID}
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
	Key: token-paymentAddress  -[-]-  {tokenId}  -[-]-  {paymentAddress}  -[-]-  {txHash}  -[-]-  {voutIndex}
  H: value-spent/unspent-rewarded/unreward

*/
func (db *db) StoreCustomTokenPaymentAddresstHistory(tokenID *common.Hash, tx *transaction.TxCustomToken) error {
	tokenKey := tokenPaymentAddressPrefix
	tokenKey = append(tokenKey, splitter...)
	tokenKey = append(tokenKey, (*tokenID)[:]...)
	for _, vin := range tx.TxTokenData.Vins {
		paymentAddress := vin.PaymentAddress.Pk
		utxoHash := &vin.TxCustomTokenID
		voutIndex := vin.VoutIndex
		paymentAddressKey := tokenKey
		paymentAddressKey = append(paymentAddressKey, splitter...)
		paymentAddressKey = append(paymentAddressKey, paymentAddress...)
		paymentAddressKey = append(paymentAddressKey, splitter...)
		paymentAddressKey = append(paymentAddressKey, utxoHash[:]...)
		paymentAddressKey = append(paymentAddressKey, splitter...)
		paymentAddressKey = append(paymentAddressKey, byte(voutIndex))
		fmt.Println(string(paymentAddressKey))
		ok, err := db.hasValue(paymentAddressKey)
		fmt.Println("Finding VIN in StoreCustomTokenPaymentAddresstHistory ", ok)
		if err != nil {
			fmt.Println("ERROR finding vin in DB, StoreCustomTokenPaymentAddresstHistory", tx.Hash(), err)
			return err
		}
		value, err := db.lvdb.Get(paymentAddressKey, nil)
		if err != nil {
			return err
		}
		// old VoteAmount: {VoteAmount}-unspent-unreward/reward
		values := strings.Split(string(value), string(splitter))
		fmt.Println("OldValues in StoreCustomTokenPaymentAddresstHistory", string(value))
		if strings.Compare(values[1], string(unspent)) != 0 {
			return errors.New("Double Spend Detected")
		}
		// new VoteAmount: {VoteAmount}-spent-unreward/reward
		newValues := values[0] + string(splitter) + string(spent) + string(splitter) + values[2]
		fmt.Println("NewValues in StoreCustomTokenPaymentAddresstHistory", newValues)
		if err := db.lvdb.Put(paymentAddressKey, []byte(newValues), nil); err != nil {
			return err
		}
	}
	for _, vout := range tx.TxTokenData.Vouts {
		paymentAddress := vout.PaymentAddress.Pk
		utxoHash := tx.Hash()
		voutIndex := vout.GetIndex()
		value := vout.Value
		paymentAddressKey := tokenKey
		paymentAddressKey = append(paymentAddressKey, splitter...)
		paymentAddressKey = append(paymentAddressKey, paymentAddress...)
		log.Println(hex.EncodeToString(paymentAddressKey))
		paymentAddressKey = append(paymentAddressKey, splitter...)
		paymentAddressKey = append(paymentAddressKey, utxoHash[:]...)
		paymentAddressKey = append(paymentAddressKey, splitter...)
		paymentAddressKey = append(paymentAddressKey, byte(voutIndex))
		fmt.Println(string(paymentAddressKey))
		ok, err := db.hasValue(paymentAddressKey)
		// Vout already exist
		if ok {
			return errors.New("UTXO already exist")
		}
		if err != nil {
			fmt.Println("ERROR finding vout in DB, StoreCustomTokenPaymentAddresstHistory", tx.Hash(), err)
			return err
		}
		// init VoteAmount: {VoteAmount}-unspent-unreward
		paymentAddressValue := strconv.Itoa(int(value)) + string(splitter) + string(unspent) + string(splitter) + string(unreward)
		fmt.Println("H in StoreCustomTokenPaymentAddresstHistory: ", paymentAddressValue)
		if err := db.lvdb.Put(paymentAddressKey, []byte(paymentAddressValue), nil); err != nil {
			return err
		}
	}
	return nil
}

/*
	Return a list of all address with balance > 0
*/
func (db *db) GetCustomTokenListPaymentAddress(tokenID *common.Hash) ([][]byte, error) {
	results := [][]byte{}
	tempsResult := make(map[string]bool)
	prefix := tokenPaymentAddressPrefix
	prefix = append(prefix, splitter...)
	prefix = append(prefix, (*tokenID)[:]...)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(splitter))
		values := strings.Split(value, string(splitter))
		if strings.Compare(values[1], string(unspent)) == 0 {
			paymentAddressStr := keys[2]
			tempsResult[paymentAddressStr] = true
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
func (db *db) GetCustomTokenListPaymentAddressesBalance(tokenID *common.Hash) (map[string]uint64, error) {
	results := make(map[string]uint64)
	//tempsResult := make(map[string]bool)
	prefix := tokenPaymentAddressPrefix
	prefix = append(prefix, splitter...)
	prefix = append(prefix, (*tokenID)[:]...)
	//fmt.Println("GetCustomTokenListPaymentAddressesBalance, prefix", prefix)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(splitter))
		values := strings.Split(value, string(splitter))
		fmt.Println("GetCustomTokenListPaymentAddressesBalance, utxo information", value)
		if strings.Compare(values[1], string(unspent)) == 0 {
			// Uncomment this to get balance of all account
			paymentAddress := privacy.PaymentAddress{}
			paymentAddress.FromBytes([]byte(keys[2]))
			i, ok := results[hex.EncodeToString(paymentAddress.Pk)]
			fmt.Println("GetCustomTokenListPaymentAddressesBalance, current balance", i)
			if ok == false {
				fmt.Println("ERROR geting VoteAmount in GetCustomTokenAccountHistory of account", paymentAddress)
			}
			balance, _ := strconv.Atoi(values[0])
			fmt.Println("GetCustomTokenListPaymentAddressesBalance, add balance", balance)
			i += uint64(balance)
			results[hex.EncodeToString(paymentAddress.Pk)] = i
			fmt.Println("GetCustomTokenListPaymentAddressesBalance, new balance", results[hex.EncodeToString(paymentAddress.Pk)])
		}
	}
	iter.Release()
	return results, nil
}

/*
	Get a list of UTXO that can be reward all payment address
	PubKey: payment address
	VoteAmount: a list of utxo
	Each utxo consist of two part: txHash-index
*/
/*func (db *db) GetCustomTokenListUnrewardUTXO(tokenID *common.Hash) (map[client.PaymentAddress][][]byte, error) {

	results := make(map[client.PaymentAddress][][]byte)
	prefix := tokenPaymentAddressPrefix
	prefix = append(prefix, splitter...)
	prefix = append(prefix, (*tokenID)[:]...)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.H())
		keys := strings.Split(key, string(splitter))
		values := strings.Split(value, string(splitter))
		// get unspent and unreward transaction output
		if (strings.Compare(values[1], string(unspent)) == 0) && (strings.Compare(values[2], string(unreward)) == 0) {
			paymentAddress := client.PaymentAddress{}
			paymentAddress.FromBytes([]byte(keys[2]))
			utxo := keys[4] + string(splitter) + keys[5]
			//utxo := append([]byte(keys[4]), []byte(keys[5])[:]...)
			results[paymentAddress] = append(results[paymentAddress], []byte(utxo))
		}
	}
	iter.Release()
	return results, nil
}*/

/*
	Get a list of UTXO of one address
	Return a list of UTXO, each UTXO has format: txHash-index
*/
func (db *db) GetCustomTokenPaymentAddressUTXO(tokenID *common.Hash, paymentAddress privacy.PaymentAddress) ([]transaction.TxTokenVout, error) {
	prefix := tokenPaymentAddressPrefix
	prefix = append(prefix, splitter...)
	prefix = append(prefix, (*tokenID)[:]...)
	prefix = append(prefix, splitter...)
	prefix = append(prefix, paymentAddress.Pk...)
	log.Println(hex.EncodeToString(prefix))
	results := []transaction.TxTokenVout{}
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		// token-paymentAddress  -[-]-  {tokenId}  -[-]-  {paymentAddress}  -[-]-  {txHash}  -[-]-  {voutIndex}
		value := string(iter.Value())
		keys := strings.Split(key, string(splitter))
		values := strings.Split(value, string(splitter))
		// get unspent and unreward transaction output
		if strings.Compare(values[1], string(unspent)) == 0 {

			vout := transaction.TxTokenVout{}
			vout.PaymentAddress = paymentAddress
			txHash, err := common.Hash{}.NewHash([]byte(keys[3]))
			if err != nil {
				return nil, err
			}
			vout.SetTxCustomTokenID(*txHash)
			voutIndexByte := []byte(keys[4])[0]
			voutIndex := int(voutIndexByte)
			vout.SetIndex(voutIndex)
			value, err := strconv.Atoi(values[0])
			if err != nil {
				return nil, err
			}
			vout.Value = uint64(value)
			fmt.Println("GetCustomTokenPaymentAddressUTXO VOUT", vout)
			results = append(results, vout)
		}
	}
	iter.Release()
	return results, nil
}

/*
	Update UTXO from unreward -> reward
*/
func (db *db) UpdateRewardAccountUTXO(tokenID *common.Hash, paymentAddress privacy.PaymentAddress, txHash *common.Hash, voutIndex int) error {
	key := tokenPaymentAddressPrefix
	key = append(key, splitter...)
	key = append(key, (*tokenID)[:]...)
	key = append(key, splitter...)
	key = append(key, (paymentAddress.Pk)[:]...)
	key = append(key, splitter...)
	key = append(key, (*txHash)[:]...)
	key = append(key, splitter...)
	key = append(key, byte(voutIndex))
	_, err := db.hasValue([]byte(key))
	if err != nil {
		fmt.Println("ERROR finding PubKey in DB, UpdateRewardAccountUTXO", err)
		return err
	}
	res, err := db.lvdb.Get([]byte(key), nil)
	if err != nil {
		return err
	}
	reses := strings.Split(string(res), string(splitter))
	// {VoteAmount}-unspent-unreward
	value := reses[0] + reses[1] + string(rewared)
	if err := db.lvdb.Put([]byte(key), []byte(value), nil); err != nil {
		return err
	}
	return nil
}

func (db *db) SaveCrowdsaleData(saleData *voting.SaleData) error {
	return nil
}

func (db *db) LoadCrowdsaleData(saleID []byte) (*voting.SaleData, error) {
	return nil, nil
}
