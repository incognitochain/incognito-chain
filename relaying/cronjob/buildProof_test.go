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
	portingID := "porting-10"
	memoPorting := PortingMemoBNB{PortingID: portingID}
	memoPortingBytes, err := json.Marshal(memoPorting)
	fmt.Printf("err: %v\n", err)
	memoPortingStr := base64.StdEncoding.EncodeToString(memoPortingBytes)
	fmt.Printf("memoPortingStr: %v\n", memoPortingStr)
	// eyJQb3J0aW5nSUQiOiIzIn0= // 3
	// eyJQb3J0aW5nSUQiOiIyIn0= // 2
	// eyJQb3J0aW5nSUQiOiIxIn0= // 1

	redeemID := "bnb13"
	custodianIncAddr := "12Rwz4HXkVABgRnSb5Gfu1FaJ7auo3fLNXVGFhxx1dSytxHpWhbkimT1Mv5Z2oCMsssSXTVsapY8QGBZd2J4mPiCTzJAtMyCzb4dDcy"
	memoRedeem := RedeemMemoBNB{RedeemID: redeemID, CustodianIncognitoAddress: custodianIncAddr}
	memoRedeemBytes, err := json.Marshal(memoRedeem)
	fmt.Printf("err: %v\n", err)
	memoRedeemHash := common.HashB(memoRedeemBytes)
	memoRedeemStr := base64.StdEncoding.EncodeToString(memoRedeemHash)
	fmt.Printf("memoRedeemStr: %v\n", memoRedeemStr)
}

func TestBuildAndPushBNBProof(t *testing.T) {
	txIndex := 0
	blockHeight := int64(80536994)
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
