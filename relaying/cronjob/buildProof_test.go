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

	redeemID := "4"
	memoRedeem := RedeemMemoBNB{RedeemID: redeemID}
	memoRedeemBytes, err := json.Marshal(memoRedeem)
	fmt.Printf("err: %v\n", err)
	memoRedeemStr := base64.StdEncoding.EncodeToString(memoRedeemBytes)
	fmt.Printf("memoRedeemStr: %v\n", memoRedeemStr)
	// eyJSZWRlZW1JRCI6IjIifQ==   // 2

	// eyJSZWRlZW1JRCI6IjQifQ==   // 4
}


func TestBuildAndPushBNBProof(t *testing.T) {
	txIndex := 0
	blockHeight := int64(450)
	url := relaying.TestnetURLRemote

	portingProof, err := BuildProof(txIndex, blockHeight, url)
	if err != nil {
		fmt.Printf("err BuildProof: %+v\n", err)
	}
	fmt.Printf("BNB portingProof: %+v\n", portingProof)
	// eyJQcm9vZiI6eyJSb290SGFzaCI6IjdGMjY5NjRFRjc3RDFFMDg3NkE2Rjg1MkFBOTEyNTEyNUJFODMyQUZENDkxMkNCRDRERTY5QkZENjM2NDBBQTgiLCJEYXRhIjoiZ1FMd1lsM3VDbk1xTElmNkNpTUtGTTJHbk5mZjArSUlnNHl3NDN3YmR5LzlzWjFuRWdzS0EwSk9RaENBdk1HV0N4SWpDaFJtbzRIT2tQc3p5ajJITlFyK08yQlByV29iSFJJTENnTkNUa0lRZ042Z3l3VVNJd29VTndnandFd1d1RVhhSFNzU29ZUThabitFbHFRU0N3b0RRazVDRUlEZW9Nc0ZFbXdLSnV0YTZZY2hBcUdMU0pHVlZjYWZqOFg2aWVRQUJ3dm5lNDZpT2F3enhyNVVuTms5VHZPNUVrQmxQZUJWNnczdDhOVmNjRW9MQ3VkWG10TUpkeCtRTTN3Nkp0bTNCMDRRdWlpVm9laG1NRnVHN0t1VUhlZUJFTW9tbTZYYXYzNXJFd3hXeFdxNkY4ZHZJQUVhR0dWNVNsRmlNMG93WVZjMWJsTlZVV2xQYVVsNFNXNHdQUT09IiwiUHJvb2YiOnsidG90YWwiOjEsImluZGV4IjowLCJsZWFmX2hhc2giOiJmeWFXVHZkOUhnaDJwdmhTcXBFbEVsdm9NcS9Va1N5OVRlYWIvV05rQ3FnPSIsImF1bnRzIjpbXX19LCJCbG9ja0hlaWdodCI6MjQ3fQ==



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
