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
	portingID := "1"
	memoPorting := PortingMemoBNB{PortingID: portingID}
	memoPortingBytes, err := json.Marshal(memoPorting)
	fmt.Printf("err: %v\n", err)
	memoPortingStr := base64.StdEncoding.EncodeToString(memoPortingBytes)
	fmt.Printf("memoPortingStr: %v\n", memoPortingStr)
	//  eyJQb3J0aW5nSUQiOiIxIn0=

	redeemID := "11"
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
	blockHeight := int64(446)
	url := relaying.TestnetURLRemote

	portingProof, err := BuildProof(txIndex, blockHeight, url)
	if err != nil {
		fmt.Printf("err BuildProof: %+v\n", err)
	}
	fmt.Printf("BNB portingProof: %+v\n", portingProof)

	//eyJQcm9vZiI6eyJSb290SGFzaCI6IjZEQzE2NjA2QkFCOUI4OTJDMTM0MjYyMUUzNzMzOTc2N0U2QzdDNTlGQjEwOUVDQTVFOTU1NTJBNzZFNTMyM0EiLCJEYXRhIjoiM0FId1lsM3VDazRxTElmNkNpTUtGUFlCSW1pVXNEQnlSNjdEbjlVVUVtTmRNSEdJRWdzS0EwSk9RaENBbE92Y0F4SWpDaFEzQ0NQQVRCYTRSZG9kS3hLaGhEeG1mNFNXcEJJTENnTkNUa0lRZ0pUcjNBTVNiQW9tNjFycGh5RUNaekxjSnNLR0d3eVFibkNaT2g1Y2sxcENNYzdHR1lKdGk3REFSMzZHYXQwU1FINE9XM1B5YW10V0hqN1pSQWh2Z0M3UzVESDBBOUZFck5lS3BLandZWjh2YWhIVldUTUhPWC9qdFFwTkM3OEVyeFBaL2QzdGRCekhkdG8xMi9FQUJJQWdBUm9ZWlhsS1VXSXpTakJoVnpWdVUxVlJhVTlwU1hoSmJqQTkiLCJQcm9vZiI6eyJ0b3RhbCI6MSwiaW5kZXgiOjAsImxlYWZfaGFzaCI6ImJjRm1CcnE1dUpMQk5DWWg0M001ZG41c2ZGbjdFSjdLWHBWVktuYmxNam89IiwiYXVudHMiOltdfX0sIkJsb2NrSGVpZ2h0IjoxMDZ9

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

	//eyJQcm9vZiI6eyJSb290SGFzaCI6IkRGMzE3NDAzODIzNzI4OEUwN0M2NzkyQzQ0NjFDMkI3OUYwNTAxQ0EwQTQ3REFBNjk4MUUyMEE0NTI1RkNFOEYiLCJEYXRhIjoiMkFId1lsM3VDa1lxTElmNkNoOEtGTTJHbk5mZjArSUlnNHl3NDN3YmR5LzlzWjFuRWdjS0EwSk9RaEFLRWg4S0ZHYWpnYzZRK3pQS1BZYzFDdjQ3WUUrdGFoc2RFZ2NLQTBKT1FoQUtFbXdLSnV0YTZZY2hBcUdMU0pHVlZjYWZqOFg2aWVRQUJ3dm5lNDZpT2F3enhyNVVuTms5VHZPNUVrQmprdXQxanZvQXBjbEozVmNPL215Sk5hbXIxNzBDdVRsZW80WmIybnhmd0F2Yi9tMmVtRXJ1bHdacWRqMjh2b2toRVFENTJJWW9oOTRsYzRLeEk4SUJJQVFhSEdWNVNsRmlNMG93WVZjMWJsTlZVV2xQYVVsNFRXcE5NRWx1TUQwPSIsIlByb29mIjp7InRvdGFsIjoxLCJpbmRleCI6MCwibGVhZl9oYXNoIjoiM3pGMEE0STNLSTRIeG5rc1JHSEN0NThGQWNvS1I5cW1tQjRncEZKZnpvOD0iLCJhdW50cyI6W119fSwiQmxvY2tIZWlnaHQiOjQ5N30=
	// redeemID = 2
	// eyJQcm9vZiI6eyJSb290SGFzaCI6IkE2MzJDOEMxQzQxODg5N0I0MUJCRTlBRkI1NDVBMURERTRERDY1QjQ5OEVDMzFBNTVFRTVCQTI3MjJDNjExQTAiLCJEYXRhIjoiMUFId1lsM3VDa1lxTElmNkNoOEtGTTJHbk5mZjArSUlnNHl3NDN3YmR5LzlzWjFuRWdjS0EwSk9RaEFLRWg4S0ZHYWpnYzZRK3pQS1BZYzFDdjQ3WUUrdGFoc2RFZ2NLQTBKT1FoQUtFbXdLSnV0YTZZY2hBcUdMU0pHVlZjYWZqOFg2aWVRQUJ3dm5lNDZpT2F3enhyNVVuTms5VHZPNUVrQmtqbk54WlNYNVZSenNQNklpLzVzaTZ6RnpjTEN6Q2tEWjVMSWhBVStNNFRQNVFDbkp6c2d5WEpyM2FUYnBTbUh1ODBtam91NzNHUnRKcWprUzNFNjRJQVVhR0dWNVNsTmFWMUpzV2xjeFNsSkRTVFpKYWtscFpsRTlQUT09IiwiUHJvb2YiOnsidG90YWwiOjEsImluZGV4IjowLCJsZWFmX2hhc2giOiJwakxJd2NRWWlYdEJ1K212dFVXaDNlVGRaYlNZN0RHbFh1VzZKeUxHRWFBPSIsImF1bnRzIjpbXX19LCJCbG9ja0hlaWdodCI6NTg5fQ==
	// redeemID = 3
	// eyJQcm9vZiI6eyJSb290SGFzaCI6IjM1NDA4MjMwM0U4NjY4ODM4MjZCODczODM4QzBBNjVGMjEwNzQ0NkE4RjVCNjMxNzUzRUYyMDgzNDdBQ0MxODEiLCJEYXRhIjoiMUFId1lsM3VDa1lxTElmNkNoOEtGTTJHbk5mZjArSUlnNHl3NDN3YmR5LzlzWjFuRWdjS0EwSk9RaEFLRWg4S0ZHemd3VWllVVQrdStoQ2YxdVV5MzQ1djlwOWpFZ2NLQTBKT1FoQUtFbXdLSnV0YTZZY2hBcUdMU0pHVlZjYWZqOFg2aWVRQUJ3dm5lNDZpT2F3enhyNVVuTms5VHZPNUVrQTRNSVdjdktvYXFuK1pzbCtybkt5VWhXYmIxWldFZHNlM2FxajlHRTJzNFhQb0xQbkJZY1duS0tpZVdqWVM5Q1hIYUNNQUpob3d4NzZRUURINmdSSWRJQVlhR0dWNVNsTmFWMUpzV2xjeFNsSkRTVFpKYWsxcFpsRTlQUT09IiwiUHJvb2YiOnsidG90YWwiOjEsImluZGV4IjowLCJsZWFmX2hhc2giOiJOVUNDTUQ2R2FJT0NhNGM0T01DbVh5RUhSR3FQVzJNWFUrOGdnMGVzd1lFPSIsImF1bnRzIjpbXX19LCJCbG9ja0hlaWdodCI6NjczfQ==
}
