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
	key := db.GetKey(string(TokenPrefix), tokenID) // token-{tokenID}-chainID-(999999999-blockHeight)-(999999999-txIndex)
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
	key := db.GetKey(string(TokenPrefix), tokenID)
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
	Return a list of all address with balance > 0
*/
func (db *db) GetCustomTokenListPaymentAddress(tokenID *common.Hash) ([][]byte, error) {
	results := [][]byte{}
	tempsResult := make(map[string]bool)
	prefix := TokenPaymentAddressPrefix
	prefix = append(prefix, Splitter...)
	prefix = append(prefix, (*tokenID)[:]...)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(Splitter))
		values := strings.Split(value, string(Splitter))
		if strings.Compare(values[1], string(Unspent)) == 0 {
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
func (db *db) GetCustomTokenPaymentAddressesBalance(tokenID *common.Hash) (map[string]uint64, error) {
	results := make(map[string]uint64)
	//tempsResult := make(map[string]bool)
	prefix := TokenPaymentAddressPrefix
	prefix = append(prefix, Splitter...)
	prefix = append(prefix, (*tokenID)[:]...)
	//fmt.Println("GetCustomTokenPaymentAddressesBalance, prefix", prefix)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(Splitter))
		values := strings.Split(value, string(Splitter))
		fmt.Println("GetCustomTokenPaymentAddressesBalance, utxo information", value)
		if strings.Compare(values[1], string(Unspent)) == 0 {
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
/*func (db *db) GetCustomTokenListUnrewardUTXO(tokenID *common.Hash) (map[privacy.PaymentAddress][][]byte, error) {

	results := make(map[privacy.PaymentAddress][][]byte)
	prefix := TokenPaymentAddressPrefix
	prefix = append(prefix, Splitter...)
	prefix = append(prefix, (*tokenID)[:]...)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(Splitter))
		values := strings.Split(value, string(Splitter))
		// get unspent and unreward transaction output
		if (strings.Compare(values[1], string(Unspent)) == 0) && (strings.Compare(values[2], string(Unreward)) == 0) {
			paymentAddress := privacy.PaymentAddress{}
			paymentAddress.FromBytes([]byte(keys[2]))
			utxo := keys[4] + string(Splitter) + keys[5]
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
func (db *db) GetCustomTokenPaymentAddressUTXO(tokenID *common.Hash, pubkey []byte) (map[string]string, error) {
	prefix := TokenPaymentAddressPrefix
	prefix = append(prefix, Splitter...)
	prefix = append(prefix, (*tokenID)[:]...)
	prefix = append(prefix, Splitter...)
	prefix = append(prefix, pubkey...)
	log.Println(hex.EncodeToString(prefix))
	results := make(map[string]string)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		// token-paymentAddress  -[-]-  {tokenId}  -[-]-  {paymentAddress}  -[-]-  {txHash}  -[-]-  {voutIndex}
		value := string(iter.Value())
		results[key] = value
	}
	iter.Release()
	return results, nil
}

/*
	Update UTXO from unreward -> reward
*/
func (db *db) UpdateRewardAccountUTXO(tokenID *common.Hash, pubkey []byte, txHash *common.Hash, voutIndex int) error {
	key := TokenPaymentAddressPrefix
	key = append(key, Splitter...)
	key = append(key, (*tokenID)[:]...)
	key = append(key, Splitter...)
	key = append(key, pubkey...)
	key = append(key, Splitter...)
	key = append(key, (*txHash)[:]...)
	key = append(key, Splitter...)
	key = append(key, byte(voutIndex))
	_, err := db.HasValue([]byte(key))
	if err != nil {
		fmt.Println("ERROR finding PubKey in DB, UpdateRewardAccountUTXO", err)
		return err
	}
	res, err := db.Get([]byte(key))
	if err != nil {
		return err
	}
	reses := strings.Split(string(res), string(Splitter))
	// {value}-unspent-unreward
	value := reses[0] + reses[1] + string(rewared)
	if err := db.lvdb.Put([]byte(key), []byte(value), nil); err != nil {
		return err
	}
	return nil
}

func (db *db) SaveCrowdsaleData(saleID []byte, endBlock int32, buyingAsset []byte, amountBuying uint64, sellingAsset []byte, amountSelling uint64) error {
	return nil
}

func (db *db) LoadCrowdsaleData(saleID []byte) (int32, []byte, uint64, []byte, uint64, error) {
	return 0, nil, 0, nil, 0, nil
}
