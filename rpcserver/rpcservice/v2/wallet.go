package rpcservicev2

import (
    "encoding/json"
    "errors"
    // "fmt"

    "github.com/incognitochain/incognito-chain/blockchain"
    // "github.com/incognitochain/incognito-chain/common"
    "github.com/incognitochain/incognito-chain/privacy"
    "github.com/incognitochain/incognito-chain/wallet"
)

type WalletService struct {
	BlockChain *blockchain.BlockChain
}

// this is marshalled into a base58 string
type KeyHolder struct{
	Key wallet.KeyWallet
}

func (kh KeyHolder) MarshalJSON() ([]byte, error) {
	rawStr := kh.Key.Base58CheckSerialize(wallet.PaymentAddressType)
	return json.Marshal(rawStr)
}

func (kh *KeyHolder) UnmarshalJSON(rawBytes []byte) error {
	var rawStr string
	err := json.Unmarshal(rawBytes, &rawStr)
	if err != nil {
		return nil
	}
	keyWallet, err := wallet.Base58CheckDeserialize(rawStr)
	if err != nil {
		return nil
	}
	kh.Key = *keyWallet
	return nil
}

/*
this submits a chain-facing `OTA key` to view its balances later
Parameter #1—the OTA key that will be submitted
		  #2-(optional) the height to sync from. If not present, the indexer assumes this key has no coins on the chain yet
Result—success or error
*/
func (walletService WalletService) SubmitKey(keyHolder KeyHolder, syncFrom *uint64) (interface{}, error) {
	type response struct {
		SyncingState int8
	}
	failure := &response{SyncingState : -1}
	success := &response{SyncingState : 0}
    // this function accepts a private key or a hex-encoded OTA key
    var otaKey privacy.OTAKey = keyHolder.Key.KeySet.OTAKey
    if otaKey.GetOTASecretKey()==nil || otaKey.GetPublicSpend()==nil {
        return failure, errors.New("OTA key not found")
    }
    var heightToSyncFrom uint64
    if syncFrom != nil {
    	heightToSyncFrom = *syncFrom
    	success.SyncingState = 1
    }
	
    err := walletService.BlockChain.SubmitOTAKey(otaKey, syncFrom != nil, heightToSyncFrom)
    if err != nil {
        return failure, err
    }

    return success, nil
}