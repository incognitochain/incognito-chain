package lvdb

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/incognitochain/incognito-chain/database"
	"log"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// StoreCustomToken - store data about custom token
// Key: token-init-{tokenID}
// Value: txHash
func (db *db) StoreCustomToken(tokenID common.Hash, txHash []byte) error {
	key := db.GetKey(string(tokenInitPrefix), tokenID)
	if err := db.Put(key, txHash); err != nil {
		return err
	}
	return nil
}

// StorePrivacyCustomToken - store data about privacy custom token when init
// Key: privacy-token-init-{tokenID}
// Value: txHash
func (db *db) StorePrivacyCustomToken(tokenID common.Hash, txHash []byte) error {
	key := db.GetKey(string(privacyTokenInitPrefix), tokenID) // token-init-{tokenID}
	if err := db.Put(key, txHash); err != nil {
		return err
	}
	return nil
}

func (db *db) StoreCustomTokenTx(tokenID common.Hash, shardID byte, blockHeight uint64, txIndex int32, txHash []byte) error {
	key := db.GetKey(string(TokenPrefix), tokenID) // token-{tokenID}-shardID-(999999999-blockHeight)-(999999999-txIndex)
	key = append(key, shardID)
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, bigNumber-blockHeight)
	key = append(key, bs...)
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(bigNumber-txIndex))
	key = append(key, bs...)
	log.Println(string(key))
	if err := db.Put(key, txHash); err != nil {
		return err
	}
	return nil
}

func (db *db) StorePrivacyCustomTokenTx(tokenID common.Hash, shardID byte, blockHeight uint64, txIndex int32, txHash []byte) error {
	key := db.GetKey(string(PrivacyTokenPrefix), tokenID) // token-{tokenID}-shardID-(999999999-blockHeight)-(999999999-txIndex)
	key = append(key, shardID)
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, bigNumber-blockHeight)
	key = append(key, bs...)
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(bigNumber-txIndex))
	key = append(key, bs...)
	log.Println(string(key))
	if err := db.Put(key, txHash); err != nil {
		return err
	}
	return nil
}

func (db *db) CustomTokenIDExisted(tokenID common.Hash) bool {
	key := db.GetKey(string(tokenInitPrefix), tokenID)
	data, err := db.Get(key)
	if err != nil {
		return false
	}
	if data == nil || len(data) == 0 {
		return false
	}
	return true
}

func (db *db) PrivacyCustomTokenIDExisted(tokenID common.Hash) bool {
	key := db.GetKey(string(privacyTokenInitPrefix), tokenID) // token-init-{tokenID}
	data, err := db.Get(key)
	if err != nil {
		return false
	}
	if data == nil || len(data) == 0 {
		return false
	}
	return true
}

func (db *db) PrivacyCustomTokenIDCrossShardExisted(tokenID common.Hash) bool {
	key := db.GetKey(string(PrivacyTokenCrossShardPrefix), tokenID)
	data, err := db.Get(key)
	if err != nil {
		return false
	}
	if data == nil || len(data) == 0 {
		return false
	}
	return true
}

/*
	Return list of txhash
*/
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

/*
	Return list of txhash
*/
func (db *db) ListPrivacyCustomToken() ([][]byte, error) {
	result := make([][]byte, 0)
	iter := db.lvdb.NewIterator(util.BytesPrefix(privacyTokenInitPrefix), nil)
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		result = append(result, value)
	}
	iter.Release()
	return result, nil
}

func (db *db) CustomTokenTxs(tokenID common.Hash) ([]common.Hash, error) {
	result := make([]common.Hash, 0)
	key := db.GetKey(string(TokenPrefix), tokenID)
	// PubKey = token-{tokenID}
	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	log.Println(string(key))
	for iter.Next() {
		value := iter.Value()
		hash, _ := common.Hash{}.NewHash(value)
		result = append(result, *hash)
	}
	iter.Release()
	return result, nil
}

func (db *db) PrivacyCustomTokenTxs(tokenID common.Hash) ([]common.Hash, error) {
	result := make([]common.Hash, 0)
	key := db.GetKey(string(PrivacyTokenPrefix), tokenID)
	// PubKey = token-{tokenID}
	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	log.Println(string(key))
	for iter.Next() {
		value := iter.Value()
		hash, _ := common.Hash{}.NewHash(value)
		result = append(result, *hash)
	}
	iter.Release()
	return result, nil
}

/*
	Return a list of all address with balance > 0
*/
func (db *db) GetCustomTokenPaymentAddressesBalance(tokenID common.Hash) (map[string]uint64, error) {
	results := make(map[string]uint64)
	prefix := TokenPaymentAddressPrefix
	prefix = append(prefix, Splitter...)
	prefix = append(prefix, []byte(tokenID.String())...)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		keys := strings.Split(key, string(Splitter))
		values := strings.Split(value, string(Splitter))
		database.Logger.Log.Info("GetCustomTokenPaymentAddressesBalance, utxo information", value)
		if strings.Compare(values[1], string(Unspent)) == 0 {
			// Uncomment this to get balance of all account
			paymentAddress := privacy.PaymentAddress{}
			paymentAddressInBytes, _, _ := base58.Base58Check{}.Decode(keys[2])
			paymentAddress.SetBytes(paymentAddressInBytes)
			i, ok := results[base58.Base58Check{}.Encode(paymentAddress.Bytes(), 0x00)]
			database.Logger.Log.Info("GetCustomTokenListPaymentAddressesBalance, current balance", i)
			if !ok {
				database.Logger.Log.Info("ERROR geting VoteAmount in GetCustomTokenAccountHistory of account", paymentAddress)
			}
			balance, _ := strconv.Atoi(values[0])
			database.Logger.Log.Info("GetCustomTokenListPaymentAddressesBalance, add balance", balance)
			i += uint64(balance)
			results[base58.Base58Check{}.Encode(paymentAddress.Bytes(), 0x00)] = i
			database.Logger.Log.Info("GetCustomTokenListPaymentAddressesBalance, new balance", results[hex.EncodeToString(paymentAddress.Bytes())])
		}
	}
	iter.Release()
	return results, nil
}

/*
	Get a list of UTXO of one address
	Return a list of UTXO, each UTXO has format: txHash-index
*/
func (db *db) GetCustomTokenPaymentAddressUTXO(tokenID common.Hash, paymentAddress []byte) (map[string]string, error) {
	prefix := TokenPaymentAddressPrefix
	prefix = append(prefix, Splitter...)
	prefix = append(prefix, []byte(tokenID.String())...)
	prefix = append(prefix, Splitter...)
	prefix = append(prefix, base58.Base58Check{}.Encode(paymentAddress, 0x00)...)
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

func (db *db) StorePrivacyCustomTokenCrossShard(tokenID common.Hash, tokenValue []byte) error {
	key := db.GetKey(string(PrivacyTokenCrossShardPrefix), tokenID)
	if err := db.Put(key, tokenValue); err != nil {
		return err
	}
	return nil
}

/*
	Return all data of token

*/
func (db *db) ListPrivacyCustomTokenCrossShard() ([][]byte, error) {
	result := make([][]byte, 0)
	iter := db.lvdb.NewIterator(util.BytesPrefix(PrivacyTokenCrossShardPrefix), nil)
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		result = append(result, value)
	}
	iter.Release()
	return result, nil
}
