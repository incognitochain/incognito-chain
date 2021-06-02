package debugtool

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
	"math/big"
)


func (tool *DebugTool) GetOutputCoins(outCoinKey *OutCoinKey, tokenID string, height uint64) ([]jsonresult.ICoinInfo, []*big.Int, error) {
	b, err := tool.GetListOutputCoinsByRPC(outCoinKey, tokenID, height)
	if err != nil{
		return nil, nil, err
	}

	//fmt.Println(string(b))

	return ParseCoinFromJsonResponse(b)
}

//CheckCoinsSpent checks if the provided serial numbers have been spent or not.
//
//Returned result in boolean list.
func (tool *DebugTool) CheckCoinsSpent(shardID byte, tokenID string, snList []string) ([]bool, error){
	b, err := tool.HasSerialNumberByRPC(shardID, tokenID, snList)
	if err != nil{
		return []bool{}, err
	}

	response, err := ParseResponse(b)
	if err != nil{
		return []bool{}, err
	}

	var tmp []bool
	err = json.Unmarshal(response.Result, &tmp)
	if err != nil {
		return []bool{}, err
	}

	if len(tmp) != len(snList){
		return []bool{}, errors.New(fmt.Sprintf("Length of result and length of snList mismathc: len(Result) = %v, len(snList) = %v. Perhaps the shardID was wrong.", len(tmp), len(snList)))
	}

	return tmp, nil
}

//GetUnspentOutputCoins retrieves all unspent coins of a private key, without sending the private key to the remote full node.
func (tool *DebugTool) GetUnspentOutputCoins(privateKey, tokenID string, height uint64) ([]coin.PlainCoin, []*big.Int, error){
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil{
		return nil, nil, err
	}
	outCoinKey, err := NewOutCoinKeyFromPrivateKey(privateKey)
	if err != nil{
		return nil, nil, err
	}
	outCoinKey.SetReadonlyKey("") // call this if you do not want the remote full node to decrypt your coin

	listOutputCoins, _, err := tool.GetOutputCoins(outCoinKey, tokenID, height)
	if err != nil{
		return nil, nil, err
	}

	if len(listOutputCoins) == 0{
		return nil, nil, nil
	}

	listDecryptedOutCoins, listKeyImages, err := GetListDecryptedCoins(privateKey, listOutputCoins)
	if err != nil {
		return nil, nil, err
	}

	shardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])
	checkSpentList, err := tool.CheckCoinsSpent(shardID, tokenID, listKeyImages)
	if err != nil {
		return nil, nil, err
	}

	listUnspentOutputCoins := make([]coin.PlainCoin, 0)
	for i, decryptedCoin := range listDecryptedOutCoins{
		if !checkSpentList[i] {
			listUnspentOutputCoins = append(listUnspentOutputCoins, decryptedCoin)
		}
	}

	return listUnspentOutputCoins, nil, nil
}

//GetBalance retrieves balance of a private key without sending this private key to the remote full node.
func (tool *DebugTool) GetBalance(privateKey, tokenID string) (uint64, error){
	unspentCoins, _, err := tool.GetUnspentOutputCoins(privateKey, tokenID, 0)
	if err != nil{
		return 0, err
	}

	balance := uint64(0)
	for _, unspentCoin := range unspentCoins {
		balance += unspentCoin.GetValue()
	}

	return balance, nil
}

//===================== OUTPUT COINS RPC =====================//
//These RPCs return raw JSON bytes.

//GetListOutputCoinsByRPC retrieves list of output coins of an OutCoinKey and returns the result in raw json bytes.
func (this *DebugTool) GetListOutputCoinsByRPC(outCoinKey *OutCoinKey, tokenID string, h uint64) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "listoutputcoins",
		"params": [
			0,
			999999,
			[
				{
			  "PaymentAddress": "%s",
			  "OTASecretKey": "%s",
			  "ReadonlyKey" : "%s",
			  "StartHeight": %d
				}
			],
		  "%s"
		  ],
		"id": 1
	}`, outCoinKey.paymentAddress, outCoinKey.otaKey, outCoinKey.readonlyKey, h, tokenID)

	return this.SendPostRequestWithQuery(query)
}

//GetListOutputCoinsCachedByRPC retrieves list of output coins (which have been cached at the fullnode) of an OutCoinKey and returns the result in raw json bytes.
func (this *DebugTool) GetListOutputCoinsCachedByRPC(privKeyStr, tokenID string, h uint64) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}

	keyWallet, _ := wallet.Base58CheckDeserialize(privKeyStr)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	otaSecretKey := keyWallet.Base58CheckSerialize(wallet.OTAKeyType)
	viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "listoutputcoinsfromcache",
		"params": [
			0,
			999999,
			[
				{
			  "PaymentAddress": "%s",
			  "OTASecretKey": "%s",
			  "ReadonlyKey" : "%s",
			  "StartHeight": %d
				}
			],
		  "%s"
		  ],
		"id": 1
	}`, paymentAddStr, otaSecretKey, viewingKeyStr, h, tokenID)

	//fmt.Println("==============")

	return this.SendPostRequestWithQuery(query)
}

//ListUnspentOutputCoinsByRPC retrieves list of output coins of an OutCoinKey and returns the result in raw json bytes.
//
//NOTE: PrivateKey must be supplied.
func (this *DebugTool) ListUnspentOutputCoinsByRPC(privKeyStr string) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}

	query := fmt.Sprintf(`{
	   "jsonrpc":"1.0",
	   "method":"listunspentoutputcoins",
	   "params":[
		  0,
		  999999,
		  [
			 {
				"PrivateKey":"%s",
				"StartHeight": 0
			 }

		  ]
	   ],
	   "id":1
	}`, privKeyStr)

	return this.SendPostRequestWithQuery(query)
}

//ListPrivacyCustomTokenByRPC lists all tokens currently present on the blockchain
func (this *DebugTool) ListPrivacyCustomTokenByRPC() ([]byte, error) {
	query := `{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "listprivacycustomtoken",
		"params": []
	}`
	return this.SendPostRequestWithQuery(query)
}

//HasSerialNumberByRPC checks if the provided serial numbers have been spent or not.
//
//Returned result in raw json bytes.
func (tool *DebugTool) HasSerialNumberByRPC(shardID byte, tokenID string, snList []string) ([]byte, error){
	if len(snList) == 0 {
		return nil, errors.New("no serial number provided to be checked")
	}
	snQueryList := make([]string, 0)
	for _, sn := range snList{
		snQueryList = append(snQueryList, fmt.Sprintf(`"%s"`, sn))
	}


	method := "hasserialnumbers"

	params := make([]interface{}, 0)
	params = append(params, shardID)
	params = append(params, snList)
	params = append(params, tokenID)

	request := CreateJsonRequest("1.0", method, params, 1)


	query, err := json.Marshal(request)
	if err != nil{
		return nil, err
	}

	return tool.SendPostRequestWithQuery(string(query))

}

//===================== END OF OUTPUT COINS RPC =====================//


//func (tool *DebugTool) GetUnspentOutputCoins(priva)
