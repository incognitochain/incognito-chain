package cronjob

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	relaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"testing"
)

type PortingMemoBNB struct {
	PortingID string `json:"PortingID"`
}

type RedeemMemoBNB struct {
	RedeemID                  string `json:"RedeemID"`
	CustodianIncognitoAddress string `json:"CustodianIncognitoAddress"`
}

func TestB64EncodeMemo(t *testing.T) {
	portingID := "3"
	memoPorting := PortingMemoBNB{PortingID: portingID}
	memoPortingBytes, err := json.Marshal(memoPorting)
	fmt.Printf("err: %v\n", err)
	memoPortingStr := base64.StdEncoding.EncodeToString(memoPortingBytes)
	fmt.Printf("memoPortingStr: %v\n", memoPortingStr)
	//  eyJQb3J0aW5nSUQiOiIxIn0=
	// eyJQb3J0aW5nSUQiOiIzIn0= // 3

	redeemID := "13"
	custodianIncAddr := "12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ"
	memoRedeem := RedeemMemoBNB{RedeemID: redeemID, CustodianIncognitoAddress: custodianIncAddr}
	memoRedeemBytes, err := json.Marshal(memoRedeem)
	fmt.Printf("err: %v\n", err)
	memoRedeemHash := common.HashB(memoRedeemBytes)
	memoRedeemStr := base64.StdEncoding.EncodeToString(memoRedeemHash)
	fmt.Printf("memoRedeemStr: %v\n", memoRedeemStr)
	// eyJSZWRlZW1JRCI6IjIifQ==   // 2

	// eyJSZWRlZW1JRCI6IjMifQ== // 3
	// eyJSZWRlZW1JRCI6IjQifQ==   // 4

	// eyJSZWRlZW1JRCI6IjExIn0= // 11

}

func TestBuildAndPushBNBProof(t *testing.T) {
	txIndex := 0
	blockHeight := int64(6918)
	url := relaying.TestnetURLRemote

	portingProof, err := BuildProof(txIndex, blockHeight, url)
	if err != nil {
		fmt.Printf("err BuildProof: %+v\n", err)
	}
	fmt.Printf("BNB portingProof: %+v\n", portingProof)

	//redeemProof, err := BuildProof(txIndex, blockHeight, url)
	//if err != nil {
	//	fmt.Printf("err BuildProof: %+v\n", err)
	//}
	//fmt.Printf("BNB redeemProof: %+v\n", redeemProof)

	//uniqueID := "123"
	//tokenID := "b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"
	//portingAmount := uint64(10000000000)
	//urlIncognitoNode := "http://localhost:9334"
	//BuildAndPushBNBProof(txIndex, blockHeight, url, uniqueID, tokenID, portingAmount, urlIncognitoNode)
}
