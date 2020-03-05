package cronjob

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	relaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"testing"
)

type PortingMemoBNB struct {
	PortingID string		`json:"PortingID"`
}

type RedeemMemoBNB struct {
	RedeemID string `json:"RedeemID"`
}

func TestB64EncodeMemo(t *testing.T) {
	portingID := "1"
	memoPorting := PortingMemoBNB{PortingID: portingID}
	memoPortingBytes, err := json.Marshal(memoPorting)
	fmt.Printf("err: %v\n", err)
	memoPortingStr := base64.StdEncoding.EncodeToString(memoPortingBytes)
	fmt.Printf("memoPortingStr: %v\n", memoPortingStr)
	//  eyJQb3J0aW5nSUQiOiIxIn0=

	redeemID := "2"
	memoRedeem := RedeemMemoBNB{RedeemID: redeemID}
	memoRedeemBytes, err := json.Marshal(memoRedeem)
	fmt.Printf("err: %v\n", err)
	memoRedeemStr := base64.StdEncoding.EncodeToString(memoRedeemBytes)
	fmt.Printf("memoRedeemStr: %v\n", memoRedeemStr)
	// eyJSZWRlZW1JRCI6IjIifQ==
}


func TestBuildAndPushBNBProof(t *testing.T) {
	txIndex := 0
	blockHeight := int64(532)
	url := relaying.TestnetURLRemote

	portingProof, err := BuildProof(txIndex, blockHeight, url)
	if err != nil {
		fmt.Printf("err BuildProof: %+v\n", err)
	}
	fmt.Printf("BNB portingProof: %+v\n", portingProof)



	redeemProof, err := BuildProof(txIndex, blockHeight, url)
	if err != nil {
		fmt.Printf("err BuildProof: %+v\n", err)
	}
	fmt.Printf("BNB redeemProof: %+v\n", redeemProof)

	//uniqueID := "123"
	//tokenID := "b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"
	//portingAmount := uint64(10000000000)
	//urlIncognitoNode := "http://localhost:9334"
	//BuildAndPushBNBProof(txIndex, blockHeight, url, uniqueID, tokenID, portingAmount, urlIncognitoNode)
}
