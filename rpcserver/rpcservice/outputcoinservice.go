package rpcservice

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"log"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
)

type OutputCoinService struct {
	BlockChain *blockchain.BlockChain
}

func (outputCounService OutputCoinService) ListUnspentOutputCoinsByKey(listKeyParams []interface{}) (*jsonresult.ListOutputCoins, *RPCError) {
	result := &jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain pri-key by deserializing
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

		err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		shardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])
		tokenID := &common.Hash{}
		err = tokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return nil, NewRPCError(TokenIsInvalidError, err)
		}
		outCoins, err := outputCounService.BlockChain.GetListOutputCoinsByKeyset(&keyWallet.KeySet, shardID, tokenID)
		if err != nil {
			return nil, NewRPCError(UnexpectedError, err)
		}
		item := make([]jsonresult.OutCoin, 0)
		for _, outCoin := range outCoins {
			if outCoin.CoinDetails.GetValue() == 0 {
				continue
			}
			item = append(item, jsonresult.OutCoin{
				SerialNumber:   base58.Base58Check{}.Encode(outCoin.CoinDetails.GetSerialNumber().Compress(), common.ZeroByte),
				PublicKey:      base58.Base58Check{}.Encode(outCoin.CoinDetails.GetPublicKey().Compress(), common.ZeroByte),
				Value:          strconv.FormatUint(outCoin.CoinDetails.GetValue(), 10),
				Info:           base58.Base58Check{}.Encode(outCoin.CoinDetails.GetInfo()[:], common.ZeroByte),
				CoinCommitment: base58.Base58Check{}.Encode(outCoin.CoinDetails.GetCoinCommitment().Compress(), common.ZeroByte),
				Randomness:     base58.Base58Check{}.Encode(outCoin.CoinDetails.GetRandomness().Bytes(), common.ZeroByte),
				SNDerivator:    base58.Base58Check{}.Encode(outCoin.CoinDetails.GetSNDerivator().Bytes(), common.ZeroByte),
			})
		}
		result.Outputs[priKeyStr] = item
	}
	return result, nil
}

func (outputCounService OutputCoinService) ListOutputCoinsByKey(listKeyParams []interface{}, tokenID common.Hash) (*jsonresult.ListOutputCoins, *RPCError) {
	result := &jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain readonly-key by deserializing
		readonlyKeyStr := keys["ReadonlyKey"].(string)
		readonlyKey, err := wallet.Base58CheckDeserialize(readonlyKeyStr)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}

		// get keyset only contain pub-key by deserializing
		pubKeyStr := keys["PaymentAddress"].(string)
		pubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}

		// create a key set
		keySet := incognitokey.KeySet{
			ReadonlyKey:    readonlyKey.KeySet.ReadonlyKey,
			PaymentAddress: pubKey.KeySet.PaymentAddress,
		}
		lastByte := keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		outputCoins, err := outputCounService.BlockChain.GetListOutputCoinsByKeyset(&keySet, shardIDSender, &tokenID)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
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
