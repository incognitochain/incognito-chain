package cronjob

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	relaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/tendermint/tendermint/types"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func GetBNBHeaderFromBinanceNetwork(blockHeight int64, url string) (string, error) {
	block, err := relaying.GetBlock(blockHeight, url)
	if err != nil {
		return "", err
	}
	if block == nil {
		return "", errors.New("Can not get block from bnb chain")
	}

	blockHeader := types.Block{
		Header:     block.Header,
		LastCommit: block.LastCommit,
	}
	bnbHeaderBytes, err2 := json.Marshal(blockHeader)
	if err2 != nil {
		return "", err2
	}

	bnbHeaderStr := base64.StdEncoding.EncodeToString(bnbHeaderBytes)
	return bnbHeaderStr, nil
}

func callRPCIncognito(method string, params string, url string) (map[string]interface{}, error) {
	var err error
	var result = make(map[string]interface{})
	body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"1\",\"method\":\"" + method + "\",\"params\":[" + params + "]}")
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return result, err
	}
	//req.SetBasicAuth(btcClient.User, btcClient.Password)
	req.Header.Set("Content-Type", "text/plain;")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func PushBNBHeaderIntoIncognito(bnbHeaderStr string, blockHeight int64, urlIncognitoNode string) (map[string]interface{}, error) {
	params := `
		"", 
    	null, 
    	-1,   
        0,
        {
        	"SenderAddress" : "12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci",
        	"Header": "` + bnbHeaderStr + `",
        	"BlockHeight": ` + strconv.Itoa(int(blockHeight)) + `
        }
	`

	result, err := callRPCIncognito("createandsendtxwithrelayingbnbheader", params, urlIncognitoNode)
	fmt.Printf("Result: %v\n", result)
	fmt.Printf("err: %v\n", err)

	return result, err
}

func GetAndPushBNBHeader() {
	url := relaying.TestnetURLRemote
	//urlIncognitoNode := "http://localhost:9334"
	blockHeight := relaying.TestnetGenesisBlockHeight
	for i := blockHeight; i <= blockHeight; i++ {
		bnbHeaderStr, err := GetBNBHeaderFromBinanceNetwork(int64(i), url)
		if err != nil {
			fmt.Printf("Error GetBNBHeaderFromBinanceNetwork: %v\n", err)
			panic(nil)
		}
		if bnbHeaderStr == "" {
			fmt.Printf("Error GetBNBHeaderFromBinanceNetwork: %v\n", err)
			panic(nil)
		}

		fmt.Printf("bnbHeaderStr: %v\n", bnbHeaderStr)

		//result, err2 := PushBNBHeaderIntoIncognito(bnbHeaderStr, int64(i), urlIncognitoNode)
		//if err2 != nil {
		//	fmt.Printf("Error PushBNBHeaderIntoIncognito: %v\n", err2)
		//	panic(nil)
		//}
		//
		//fmt.Printf("Result PushBNBHeaderIntoIncognito: %v\n", result)
		//fmt.Printf("====== Push successfully %v\n\n\n", i)
		//
		//time.Sleep(15*1000 * time.Millisecond)
	}
}
