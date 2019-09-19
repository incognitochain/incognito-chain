package rpcservice

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
	"log"
)

type CoinService struct {
	BlockChain *blockchain.BlockChain
}

func (coinService CoinService) ListOutputCoinsByKeySet(keySet *incognitokey.KeySet, shardID byte) ([]*privacy.OutputCoin, error){
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, err
	}

	return coinService.BlockChain.GetListOutputCoinsByKeyset(keySet, shardID, prvCoinID)
}

func (coinService CoinService) ListUnspentOutputCoinsByKey(listKeyParams []interface{}) (*jsonresult.ListOutputCoins, *RPCError) {
	result := &jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain pri-key by deserializing
		if _, ok  := keys["PrivateKey"]; !ok{
			return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid private key"))
		}
		priKeyStr := keys["PrivateKey"].(string)
		keyWallet, err := wallet.Base58CheckDeserialize(priKeyStr)
		if err != nil {
			log.Println("Check Deserialize err", err)
			continue
		}
		if keyWallet.KeySet.PrivateKey == nil {
			log.Println("Private key empty")
			continue
		}

		keySetTmp, shardID, err := GetKeySetFromPrivateKey(keyWallet.KeySet.PrivateKey)
		if err != nil {
			return nil, NewRPCError(UnexpectedError, err)
		}
		keyWallet.KeySet = *keySetTmp

		outCoins, err := coinService.ListOutputCoinsByKeySet(&keyWallet.KeySet, shardID)
		if err != nil {
			return nil, NewRPCError(UnexpectedError, err)
		}

		item := make([]jsonresult.OutCoin, 0)
		for _, outCoin := range outCoins {
			if outCoin.CoinDetails.GetValue() == 0 {
				continue
			}
			item = append(item, jsonresult.NewOutCoin(outCoin))
		}
		result.Outputs[priKeyStr] = item
	}
	return result, nil
}

func (coinService CoinService) ListOutputCoinsByKey(listKeyParams []interface{}, tokenID common.Hash) (*jsonresult.ListOutputCoins, *RPCError) {
	result := &jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain readonly-key by deserializing
		if _, ok := keys["ReadonlyKey"]; !ok {
			return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid readonly key"))
		}

		readonlyKeyStr := keys["ReadonlyKey"].(string)
		readonlyKey, err := wallet.Base58CheckDeserialize(readonlyKeyStr)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(UnexpectedError, err)
		}

		// get keyset only contain pub-key by deserializing
		if _, ok := keys["PaymentAddress"]; !ok {
			return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid payment address"))
		}
		pubKeyStr := keys["PaymentAddress"].(string)
		pubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(UnexpectedError, err)
		}

		// create a key set
		keySet := incognitokey.KeySet{
			ReadonlyKey:    readonlyKey.KeySet.ReadonlyKey,
			PaymentAddress: pubKey.KeySet.PaymentAddress,
		}
		lastByte := keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		outputCoins, err := coinService.BlockChain.GetListOutputCoinsByKeyset(&keySet, shardIDSender, &tokenID)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(UnexpectedError, err)
		}
		item := make([]jsonresult.OutCoin, 0)

		for _, outCoin := range outputCoins {
			if outCoin.CoinDetails.GetValue() == 0 {
				continue
			}
			item = append(item, jsonresult.NewOutCoin(outCoin))
		}
		result.Outputs[readonlyKeyStr] = item
	}
	return result, nil
}
