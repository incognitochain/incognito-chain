package rpcservice

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
)

type CoinService struct {
	BlockChain *blockchain.BlockChain
}

func (coinService CoinService) ListOutputCoinsByKeySet(keySet *incognitokey.KeySet, shardID byte, tokenID *common.Hash) ([]*privacy.OutputCoin, error) {
	if tokenID == nil {
		tokenID = &common.Hash{}
		err := tokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return nil, err
		}
	}
	return coinService.BlockChain.GetListOutputCoinsByKeyset(keySet, shardID, tokenID)
}

func (coinService CoinService) ListUnspentOutputCoinsByKey(listKeyParams []interface{}, tokenID *common.Hash) (*jsonresult.ListOutputCoins, *RPCError) {
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
		outCoins, err := coinService.ListOutputCoinsByKeySet(&keyWallet.KeySet, shardID, tokenID)
		if err != nil {
			return nil, NewRPCError(ListUnspentOutputCoinsByKeyError, err)
		}
		item := make([]jsonresult.OutCoin, 0)
		for _, outCoin := range outCoins {
			if outCoin.CoinDetails.GetValue() == 0 {
				continue
			}
			item = append(item, jsonresult.NewOutCoin(outCoin))
		}
		result.Outputs[privateKeyStr] = item
	}
	return result, nil
}

func (coinService CoinService) ListOutputCoinsByKey(listKeyParams []interface{}, tokenID common.Hash) (*jsonresult.ListOutputCoins, *RPCError) {
	result := &jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}
	for _, keyParam := range listKeyParams {
		keys, ok := keyParam.(map[string]interface{})
		if !ok {
			return nil, NewRPCError(RPCInvalidParamsError, fmt.Errorf("Invalid params: %+v", keyParam))
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
				return nil, NewRPCError(ListOutputCoinsByKeyError, err)
			}
		}
		// get keyset only contain public key by deserializing (required)
		paymentAddressStr, ok := keys["PaymentAddress"].(string)
		if !ok {
			return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid payment address"))
		}
		paymentAddressKey, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ListOutputCoinsByKeyError, err)
		}
		// create a key set
		keySet := incognitokey.KeySet{
			PaymentAddress: paymentAddressKey.KeySet.PaymentAddress,
		}
		// readonly key is optional
		if readonlyKey != nil && len(readonlyKey.KeySet.ReadonlyKey.Rk) > 0 {
			keySet.ReadonlyKey = readonlyKey.KeySet.ReadonlyKey
		}
		lastByte := keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		outputCoins, err := coinService.BlockChain.GetListOutputCoinsByKeyset(&keySet, shardIDSender, &tokenID)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ListOutputCoinsByKeyError, err)
		}
		item := make([]jsonresult.OutCoin, 0)

		for _, outCoin := range outputCoins {
			item = append(item, jsonresult.NewOutCoin(outCoin))
		}
		if readonlyKey != nil && len(readonlyKey.KeySet.ReadonlyKey.Rk) > 0 {
			result.Outputs[readonlyKeyStr] = item
		} else {
			result.Outputs[paymentAddressStr] = item
		}
	}
	return result, nil
}
