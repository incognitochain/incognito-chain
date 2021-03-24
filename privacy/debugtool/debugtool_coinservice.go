package debugtool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var privJKeyList = [...]string{
	"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",
}

type InfoJSON struct{
	Start int `json:"Start"`
	Total int `json:"Total"`
	End   int `json:"End"`
}

type KeyInfoJSON struct {
	Result struct {
		ID           string         `json:"id"`
		CreatedAt    time.Time      `json:"created_at"`
		UpdatedAt    time.Time      `json:"updated_at"`
		Pubkey       string         `json:"pubkey"`
		Otakey       string         `json:"otakey"`
		V1Startindex map[string]InfoJSON `json:"v1startindex"`
		V2Startindex map[string]InfoJSON `json:"v2startindex"`
	} `json:"Result"`
	Error interface{} `json:"Error"`
}

type OutCoinJSON struct {
	Result struct {
		HighestIndex int                             `json:"HighestIndex"`
		Outputs      map[string][]jsonresult.OutCoin `json:"Outputs"`
	} `json:"Result"`
	Error interface{} `json:"Error"`
}

var URL = "http://51.161.119.66:9001"
var NoOfShard = 8

func SendPost(url, query string) ([]byte, error) {
	var jsonStr = []byte(query)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, err
		}
		return body, nil
	}
}

func CheckCoinsSpent(shardID byte, listKeyImages []string) ([]bool, error) {
	URL := "http://51.161.119.66:9001/checkkeyimages"
	if len(listKeyImages) == 0 {
		return nil, fmt.Errorf("no serial number provided to be checked")
	}

	snQueryList := make([]string, 0)
	for _, sn := range listKeyImages {
		snQueryList = append(snQueryList, fmt.Sprintf(`"%s"`, sn))
	}

	query := fmt.Sprintf(`{
		"Keyimages":[%v],
		"ShardID":%v
	}`, strings.Join(snQueryList, ","), shardID)

	b, err := SendPost(URL, query)
	if err != nil {
		return []bool{}, err
	}

	response, err := ParseResponse(b)
	if err != nil {
		return []bool{}, err
	}

	var tmp []bool
	err = json.Unmarshal(response.Result, &tmp)
	if err != nil {
		return []bool{}, err
	}

	if len(tmp) != len(listKeyImages) {
		return []bool{}, fmt.Errorf(fmt.Sprintf("Length of result and length of snList mismathc: len(Result) = %v, len(snList) = %v. Perhaps the shardID was wrong.", len(tmp), len(listKeyImages)))
	}

	return tmp, nil
}

func GetListToken(viewingKey string) (map[string]InfoJSON, error) {
	method := "getkeyinfo"
	resp, err := http.Get(fmt.Sprintf("%v/%v?key=%v", URL, method, viewingKey))
	if err != nil {
		fmt.Println(fmt.Sprintf("cannot get list outcoin. Error %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(fmt.Sprintf("cannot parse body from response text. Error %v", err))
		return nil, err
	}
	tmp := new(KeyInfoJSON)
	if err := json.Unmarshal(respBytes, &tmp); err != nil {
		fmt.Println(fmt.Sprintf("cannot unmarshal json. Error %v", err))
		return nil, err
	}

	listToken := tmp.Result.V1Startindex
	return listToken, nil
}

func GetOutputCoins(viewingKey, tokenID string) ([]jsonresult.ICoinInfo, int, error) {
	method := "getcoins"
	query := fmt.Sprintf("%v/%v?viewkey=%v&offset=%v&limit=%v&tokenid=%v",
		URL,
		method,
		viewingKey,
		0,
		1000000,
		tokenID)

	resp, err := http.Get(query)
	if err != nil {
		fmt.Println(fmt.Sprintf("cannot get list outcoin. Error %v", err))
		return nil, 0, nil
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(fmt.Sprintf("cannot parse body from response text. Error %v", err))
		return nil, 0, nil
	}

	listOutputCoins := make([]jsonresult.ICoinInfo, 0)
	latestIndex := 0

	tmp := new(OutCoinJSON)
	if err := json.Unmarshal(respBytes, &tmp); err != nil {
		err := fmt.Errorf("cannot unmarshal response. Error %v", err)
		fmt.Println(err)
		return nil, 0, err
	} else if tmp.Error != nil {
		err := fmt.Errorf("cannot unmarshal response. Error %v", err)
		fmt.Println(err)
		return nil, 0, err
	} else {
		for _, outCoins := range tmp.Result.Outputs {
			for _, coin := range outCoins {
				out, _, err := jsonresult.NewCoinFromJsonOutCoin(coin)
				if err != nil {
					return nil, 0, err
				}
				listOutputCoins = append(listOutputCoins, out)
			}
		}
		latestIndex = tmp.Result.HighestIndex
	}
	return listOutputCoins, latestIndex, nil
}

func GetUTXO(listPlainCoins []coin.PlainCoin, listKeyImages []string, shardID byte, tokenID string) ([]coin.PlainCoin, error) {

	checkSpentList, err := CheckCoinsSpent(shardID, listKeyImages)
	if err != nil {
		return nil, err
	}

	listUnspentCoins := make([]coin.PlainCoin, 0)
	for i, plainCoin := range listPlainCoins {
		if !checkSpentList[i] {
			listUnspentCoins = append(listUnspentCoins, plainCoin)
		}
	}
	return listUnspentCoins, nil
}


func CheckCoinSpentFromRPC(tool *DebugTool, listSN []string, paymentAddress, tokenID string) ([]bool, error){
	snQueryList := make([]string, 0)
	for _, sn := range listSN {
		snQueryList = append(snQueryList, fmt.Sprintf(`"%s"`, sn))
	}

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "hasserialnumbers",
		"params": [
			"%v",
			[%v],
			"%v"
		],
		"id": 1
	}`, paymentAddress, strings.Join(snQueryList, ","), tokenID)

	resp, err := tool.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, fmt.Errorf("cannot get list outputcoin from rpc. Error %v", err)
	}
	response, err := ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result []bool
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return []bool{}, err
	}
	return result, nil
}


func GetOutputCoinsFromRPC(tool *DebugTool, paymentAddress, readOnlyKey string, height int, tokenID string) ([]jsonresult.ICoinInfo, error) {
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "listoutputcoins",
		"params": [
			0,
			999999,
			[
				{
			  "PaymentAddress": "%s",
			  "ReadOnlyKey":"%s",
			  "StartHeight": %d
				}
			],
		  "%s"
		  ],
		"id": 1
	}`, paymentAddress, readOnlyKey, height, tokenID)

	resp, err := tool.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, fmt.Errorf("cannot get list outputcoin from rpc. Error %v", err)
	}

	outputCoins, _, err := ParseCoinFromJsonResponse(resp)
	return outputCoins, err
}



func GetBalanceFromRPC(tool *DebugTool, privateKey, paymentAddressStr, readOnlyKey, tokenID string, shardID byte, height uint64) (uint64, error) {
	listOutputCoins, err := GetOutputCoinsFromRPC(tool, paymentAddressStr, readOnlyKey, 0, tokenID)
	if err != nil {
		return 0, fmt.Errorf("error cannot get list output coin from rpc %v. error %v", tokenID, err)
	}
	listPlainCoins, listKeyImages, err := GetListDecryptedCoins(privateKey, listOutputCoins)
	if err != nil {
		return 0, fmt.Errorf("error cannot get plain coins from output coins")
	}

	checkSpentList, err := CheckCoinSpentFromRPC(tool, listKeyImages, paymentAddressStr, tokenID)

	utxos := make([]coin.PlainCoin, 0)
	for i, decryptedCoin := range listPlainCoins {
		if !checkSpentList[i] {
			utxos = append(utxos, decryptedCoin)
		}
	}

	balance := uint64(0)
	for _, coin := range utxos {
		balance += coin.GetValue()
	}
	return balance, nil
}


func GetBalanceFromCS(privateKey, viewingKeyStr, tokenID string, totalCoin int,  shardID byte) (uint64, error){
	listOutputCoins, _, err :=GetOutputCoins(viewingKeyStr, tokenID)
	if err != nil {
		return 0, fmt.Errorf("cannot get list output coins from CS. Error %v", err)
	}
	if totalCoin != len(listOutputCoins) {
		fmt.Println("Wrong total coin")
		return 0, fmt.Errorf("wrong total coin")
	}
	listPlainCoins, listKeyImages, err := GetListDecryptedCoins(privateKey, listOutputCoins)
	if err != nil {
		return 0, fmt.Errorf("cannot get plain coins from CS. Error %v", err)
	}

	utxos, err := GetUTXO(listPlainCoins, listKeyImages, shardID, tokenID)

	balance := uint64(0)
	for _, coin := range utxos {
		balance += coin.GetValue()
	}
	return balance, err
}


