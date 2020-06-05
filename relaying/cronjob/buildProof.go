package cronjob

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	relaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"strconv"
)

func BuildProof(txIndex int, blockHeight int64, url string) (string, error){
	bnbProof := new(relaying.BNBProof)
	err := bnbProof.Build(txIndex, blockHeight, url)
	if err != nil {
		return "", err
	}

	bnbProofBytes, err2 := json.Marshal(bnbProof)
	if err2 != nil {
		return "", err2
	}

	bnbProofStr := base64.StdEncoding.EncodeToString(bnbProofBytes)

	return bnbProofStr, nil
}

func PushBNBProofIntoIncognito(
	bnbProof string,
	blockHeight int64, uniqueID string, tokenID string,
	portingAmount uint64,
	urlIncognitoNode string) (map[string]interface{}, error) {
	params := `
		"112t8rnaqXpcge9BETLXdBnSVMq37pVzSr1i3tcvTJ3jQMs5NCWgv5VmMwRwtm9zzELKzz6WgtoPMR9PBgY95Cf15QMGVTFvpPii3TkW2tUB", 
    	null, 
    	-1,   
        0,
        {
        	"UniquePortingID" : "` + uniqueID + `",
        	"TokenID": "` + tokenID + `",
        	"IncogAddressStr": "12S5pBBRDf1GqfRHouvCV86sWaHzNfvakAWpVMvNnWu2k299xWCgQzLLc9wqPYUHfMYGDprPvQ794dbi6UU1hfRN4tPiU61txWWenhC",
        	"PortingAmount": ` + strconv.Itoa(int(portingAmount)) + `,
        	"PortingProof": "` + bnbProof + `"
        }
	`

	result, err := callRPCIncognito("createandsendtxwithreqptoken", params, urlIncognitoNode)
	fmt.Printf("Result: %v\n", result)
	fmt.Printf("err: %v\n", err)

	return result, err
}

func BuildAndPushBNBProof(
	txIndex int, blockHeight int64,
	url string, uniqueID string, tokenID string,
	portingAmount uint64, urlIncognitoNode string){
	bnbProofStr, err := BuildProof(txIndex, blockHeight, url)
	if err != nil {
		fmt.Printf("err BuildProof: %v\n", err)
	}

	result, err := PushBNBProofIntoIncognito(bnbProofStr, blockHeight, uniqueID, tokenID, portingAmount, urlIncognitoNode)
	if err != nil {
		fmt.Printf("err PushBNBProofIntoIncognito: %v\n", err)
	}

	fmt.Printf("result PushBNBProofIntoIncognito: %v\n", result)
}
