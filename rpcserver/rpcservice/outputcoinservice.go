package rpcservice

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
)

type CoinService struct {
	BlockChain *blockchain.BlockChain
}

func (coinService CoinService) ListDecryptedOutputCoinsByKeySet(keySet *incognitokey.KeySet, shardID byte, shardHeight uint64) ([]coin.PlainCoin, error) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, err
	}
	return coinService.BlockChain.GetListDecryptedOutputCoinsByKeyset(keySet, shardID, prvCoinID, shardHeight)
}

func (coinService CoinService) ListUnspentOutputCoinsByKey(listKeyParams []interface{}) (*jsonresult.ListOutputCoins, *RPCError) {
	result := &jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}
	for _, keyParam := range listKeyParams {
		keys, ok := keyParam.(map[string]interface{})
		if !ok {
			return nil, NewRPCError(RPCInvalidParamsError, fmt.Errorf("Invalid Params %+v", keyParam))
		}
		// get keyset only contain private key by deserializing
		privateKeyStr, ok := keys["PrivateKey"].(string)
		if !ok {
			return nil, NewRPCError(RPCInvalidParamsError, errors.New("private key is invalid"))
		}
		keyWallet, err := wallet.Base58CheckDeserialize(privateKeyStr)
		if err != nil || keyWallet.KeySet.PrivateKey == nil {
			Logger.log.Error("Check Deserialize err", err)
			return nil, NewRPCError(RPCInvalidParamsError, fmt.Errorf("Private key is invalid, error %+v", err))
		}
		keySetTmp, shardID, err := GetKeySetFromPrivateKey(keyWallet.KeySet.PrivateKey)
		if err != nil {
			return nil, NewRPCError(ListUnspentOutputCoinsByKeyError, err)
		}
		keyWallet.KeySet = *keySetTmp

		// get shard height
		shardHeightTemp, ok := keys["StartHeight"].(float64)
		if !ok {
			return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid height param"))
		}
		shardHeight := uint64(shardHeightTemp)

		outCoins, err := coinService.ListDecryptedOutputCoinsByKeySet(&keyWallet.KeySet, shardID, shardHeight)
		if err != nil {
			return nil, NewRPCError(ListUnspentOutputCoinsByKeyError, err)
		}
		item := make([]jsonresult.OutCoin, 0)
		for _, outCoin := range outCoins {
			if outCoin.GetValue() != 0 {
				item = append(item, jsonresult.NewOutCoin(outCoin))
			}
		}
		result.Outputs[privateKeyStr] = item
	}
	return result, nil
}

func getKeysetFromKeyParam(keyParam interface{}) (*incognitokey.KeySet, uint64, error) {
	keys, ok := keyParam.(map[string]interface{})
	if !ok {
		return nil, 0, NewRPCError(RPCInvalidParamsError, fmt.Errorf("Invalid params: %+v", keyParam))
	}
	// get keyset only contain read only key by deserializing (optional)
	var readonlyKey *wallet.KeyWallet
	var err error
	readonlyKeyStr, ok := keys["ReadonlyKey"].(string)
	if !ok || readonlyKeyStr == "" {
		Logger.log.Info("Read onlyKey is optional")
	} else {
		readonlyKey, err = wallet.Base58CheckDeserialize(readonlyKeyStr)
		if err != nil {
			Logger.log.Debugf("Read onlyKey is invalid: err: %+v", err)
			return nil, 0, NewRPCError(ListDecryptedOutputCoinsByKeyError, err)
		}
	}
	// get keyset only contain public key by deserializing (required)
	paymentAddressStr, ok := keys["PaymentAddress"].(string)
	if !ok {
		return nil, 0, NewRPCError(RPCInvalidParamsError, errors.New("invalid payment address"))
	}
	paymentAddressKey, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
		return nil, 0, NewRPCError(ListDecryptedOutputCoinsByKeyError, err)
	}
	// create a key set
	keySet := &incognitokey.KeySet{
		PaymentAddress: paymentAddressKey.KeySet.PaymentAddress,
	}
	// readonly key is optional
	if readonlyKey != nil && len(readonlyKey.KeySet.ReadonlyKey.Rk) > 0 {
		keySet.ReadonlyKey = readonlyKey.KeySet.ReadonlyKey
	}

	fmt.Println("Checking readonlykey")
	fmt.Println("Do they the same?????")
	fmt.Println(keySet.GetReadOnlyKeyInBase58CheckEncode())
	fmt.Println(readonlyKeyStr)
	fmt.Println(readonlyKeyStr == keySet.GetReadOnlyKeyInBase58CheckEncode())
	fmt.Println("==============================")

	fmt.Println("Checking paymentAddressKey")
	fmt.Println("Do they the same?????")
	fmt.Println(keySet.GetPublicKeyInBase58CheckEncode())
	fmt.Println(paymentAddressStr)
	fmt.Println(paymentAddressStr == keySet.GetPublicKeyInBase58CheckEncode())
	fmt.Println("==============================")

	// get shard height
	shardHeightTemp, ok := keys["StartHeight"].(float64)
	if !ok {
		return nil, 0, NewRPCError(RPCInvalidParamsError, errors.New("invalid height param"))
	}
	shardHeight := uint64(shardHeightTemp)

	return keySet, shardHeight, nil
}

func (coinService CoinService) ListDecryptedOutputCoinsByKey(listKeyParams []interface{}, tokenID common.Hash) (*jsonresult.ListOutputCoins, *RPCError) {
	result := &jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}
	for _, keyParam := range listKeyParams {
		keySet, shardHeight, err := getKeysetFromKeyParam(keyParam)
		if err != nil {
			Logger.log.Debugf("ListDecryptedOutputCoinsByKeyError result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ListDecryptedOutputCoinsByKeyError, err)
		}
		shardID := common.GetShardIDFromLastByte(keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1])
		outputCoins, err := coinService.BlockChain.GetListDecryptedOutputCoinsByKeyset(keySet, shardID, &tokenID, shardHeight)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ListDecryptedOutputCoinsByKeyError, err)
		}
		item := make([]jsonresult.OutCoin, 0)

		for _, outCoin := range outputCoins {
			item = append(item, jsonresult.NewOutCoin(outCoin))
		}

		if len(keySet.ReadonlyKey.Rk) > 0 {
			result.Outputs[keySet.GetReadOnlyKeyInBase58CheckEncode()] = item
		} else {
			result.Outputs[keySet.GetPublicKeyInBase58CheckEncode()] = item
		}
	}
	return result, nil
}
