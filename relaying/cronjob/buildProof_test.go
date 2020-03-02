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

func TestB64EncodeMemo(t *testing.T) {
	portingID := "1234"
	memo := PortingMemoBNB{PortingID:portingID}
	memoBytes, err := json.Marshal(memo)
	fmt.Printf("err: %v\n", err)
	memoStr := base64.StdEncoding.EncodeToString(memoBytes)
	fmt.Printf("memoStr: %v\n", memoStr)
}


func TestBuildAndPushBNBProof(t *testing.T) {
	//txhash1 := "A74198A1FC252397C04D556A23BFD669F239B18F0D37FF073D10070BCCEF0518"  28
	// txHash2: = "B785F1F9CF24355EF02EE389876C302AD44D1B7D50BD79F1B9AE2AAA19B81F06"  532
	// proof2: "eyJQcm9vZiI6eyJSb290SGFzaCI6IkVDQjQ0MTJDODk0ODY2NTA1OUI4QjVBN0VCMjNDMjNERTQxRjhEMDMxMTJDQUE3MDExQ0U4MDdBOTQ2QUYzNjQiLCJEYXRhIjoiMkFId1lsM3VDa1lxTElmNkNoOEtGRmdGYXpvNXZQZ0pIYVRCYXEzOHk5MUEzMnVERWdjS0EwSk9RaEFLRWg4S0ZHYWpnYzZRK3pQS1BZYzFDdjQ3WUUrdGFoc2RFZ2NLQTBKT1FoQUtFbXdLSnV0YTZZY2hBdElZUDgraHpBVHZDMkMzYWo1K3Q3ZkhyK05zNUxyMXJ1NHcveSt1QmZxakVrQi91VFhyRkRINmt2aFg5bWx1TVR1alBkVWp1bG5Od2UxZlNCRnNNb1lkYmdDdkVlQ05JRHowNlZVb1ZHaEdKRGlGQ1ZCcEVzWVZnUm9GaHpDWUM2VEJJQUlhSEdWNVNsRmlNMG93WVZjMWJsTlZVV2xQYVVsNFRXcE5NRWx1TUQwPSIsIlByb29mIjp7InRvdGFsIjoxLCJpbmRleCI6MCwibGVhZl9oYXNoIjoiN0xSQkxJbElabEJadUxXbjZ5UENQZVFmalFNUkxLcHdFYzZBZXBScTgyUT0iLCJhdW50cyI6W119fSwiQmxvY2tIZWlnaHQiOjUzMn0="
	txIndex := 0
	blockHeight := int64(532)
	url := relaying.TestnetURLRemote
	//uniqueID := "123"
	//tokenID := "b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"
	//portingAmount := uint64(10000000000)
	//urlIncognitoNode := "http://localhost:9334"

	proof, err := BuildProof(txIndex, blockHeight, url)
	if err != nil {
		fmt.Printf("err BuildProof: %+v\n", err)
	}
	fmt.Printf("BNB proof: %+v\n", proof)

	//var bnbProof relaying.BNBProof
	//proofBytes, _ := base64.StdEncoding.DecodeString(proof)
	//json.Unmarshal(proofBytes, &bnbProof)
	//fmt.Printf("[buildInstructionsForReqPTokens] bnbProof: %v\n", bnbProof)
	//
	//txBNB, _ := relaying.ParseTxFromData(bnbProof.Proof.Data)
	//
	//type PortingMemoBNB struct {
	//	PortingID string		`json:"PortingID"`
	//}
	//memo := txBNB.Memo
	//fmt.Printf("[buildInstructionsForReqPTokens] memo: %v\n", memo)
	//memoBytes, err2 := base64.StdEncoding.DecodeString(memo)
	//if err2 != nil {
	//	fmt.Printf("Can not decode memo in tx bnb proof %v\n", err2)
	//}
	//fmt.Printf("[buildInstructionsForReqPTokens] memoBytes: %v\n", memoBytes)
	//
	//var portingMemo PortingMemoBNB
	//err2 = json.Unmarshal(memoBytes, &portingMemo)
	//if err2 != nil {
	//	fmt.Printf("Can not unmarshal memo in tx bnb proof  %v\n", err2)
	//}
	//
	//fmt.Printf("[buildInstructionsForReqPTokens] portingMemo: %v\n", portingMemo)

	// eyJQcm9vZiI6eyJSb290SGFzaCI6IjBFQ0UwRTgzRkQ3OTE3NEE2Qjg1OTA4M0FDMEEwRDhCMjU5NjMxRkRBN0U1Nzg2MzEyQTM0RjU3MDFCMEYxMTAiLCJEYXRhIjoidWdId1lsM3VDa1lxTElmNkNoOEtGT0k1RWg0Tk9PMUMxS09lbU1wSVFUanlvQVBNRWdjS0EwSk9RaEFLRWg4S0ZHYWpnYzZRK3pQS1BZYzFDdjQ3WUUrdGFoc2RFZ2NLQTBKT1FoQUtFbXdLSnV0YTZZY2hBNlJwVUF3TDBqbGVGdGIxMmhYUzIvU2pNd2RHU1ZQOHhBNEt6N2FENFlDckVrQ0J6M2VTY252UGtONm9qSktFY1dpckU1ZGRIK2E1UjR2eE5DWnhlMHlOTTJxWHhvM1FQTExoT1d3S3paV1pzSjJ3Z3JVQm1LNExaVWdhT0JaVFY2MzBJQUk9IiwiUHJvb2YiOnsidG90YWwiOjEsImluZGV4IjowLCJsZWFmX2hhc2giOiJEczRPZy8xNUYwcHJoWkNEckFvTml5V1dNZjJuNVhoakVxTlBWd0d3OFJBPSIsImF1bnRzIjpbXX19LCJCbG9ja0hlaWdodCI6OTV9
	//BuildAndPushBNBProof(txIndex, blockHeight, url, uniqueID, tokenID, portingAmount, urlIncognitoNode)
	//fmt.Printf("err BuildProof: %+v\n", err)
}
