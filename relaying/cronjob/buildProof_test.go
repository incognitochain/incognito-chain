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

	redeemID := "3"
	memoRedeem := RedeemMemoBNB{RedeemID: redeemID}
	memoRedeemBytes, err := json.Marshal(memoRedeem)
	fmt.Printf("err: %v\n", err)
	memoRedeemStr := base64.StdEncoding.EncodeToString(memoRedeemBytes)
	fmt.Printf("memoRedeemStr: %v\n", memoRedeemStr)
	// eyJSZWRlZW1JRCI6IjIifQ==   // 2

	// eyJSZWRlZW1JRCI6IjMifQ== // 3
	// eyJSZWRlZW1JRCI6IjQifQ==   // 4
}


func TestBuildAndPushBNBProof(t *testing.T) {
	txIndex := 0
	blockHeight := int64(1558)
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

	//eyJQcm9vZiI6eyJSb290SGFzaCI6IkRGMzE3NDAzODIzNzI4OEUwN0M2NzkyQzQ0NjFDMkI3OUYwNTAxQ0EwQTQ3REFBNjk4MUUyMEE0NTI1RkNFOEYiLCJEYXRhIjoiMkFId1lsM3VDa1lxTElmNkNoOEtGTTJHbk5mZjArSUlnNHl3NDN3YmR5LzlzWjFuRWdjS0EwSk9RaEFLRWg4S0ZHYWpnYzZRK3pQS1BZYzFDdjQ3WUUrdGFoc2RFZ2NLQTBKT1FoQUtFbXdLSnV0YTZZY2hBcUdMU0pHVlZjYWZqOFg2aWVRQUJ3dm5lNDZpT2F3enhyNVVuTms5VHZPNUVrQmprdXQxanZvQXBjbEozVmNPL215Sk5hbXIxNzBDdVRsZW80WmIybnhmd0F2Yi9tMmVtRXJ1bHdacWRqMjh2b2toRVFENTJJWW9oOTRsYzRLeEk4SUJJQVFhSEdWNVNsRmlNMG93WVZjMWJsTlZVV2xQYVVsNFRXcE5NRWx1TUQwPSIsIlByb29mIjp7InRvdGFsIjoxLCJpbmRleCI6MCwibGVhZl9oYXNoIjoiM3pGMEE0STNLSTRIeG5rc1JHSEN0NThGQWNvS1I5cW1tQjRncEZKZnpvOD0iLCJhdW50cyI6W119fSwiQmxvY2tIZWlnaHQiOjQ5N30=
	// redeemID = 2
	// eyJQcm9vZiI6eyJSb290SGFzaCI6IkE2MzJDOEMxQzQxODg5N0I0MUJCRTlBRkI1NDVBMURERTRERDY1QjQ5OEVDMzFBNTVFRTVCQTI3MjJDNjExQTAiLCJEYXRhIjoiMUFId1lsM3VDa1lxTElmNkNoOEtGTTJHbk5mZjArSUlnNHl3NDN3YmR5LzlzWjFuRWdjS0EwSk9RaEFLRWg4S0ZHYWpnYzZRK3pQS1BZYzFDdjQ3WUUrdGFoc2RFZ2NLQTBKT1FoQUtFbXdLSnV0YTZZY2hBcUdMU0pHVlZjYWZqOFg2aWVRQUJ3dm5lNDZpT2F3enhyNVVuTms5VHZPNUVrQmtqbk54WlNYNVZSenNQNklpLzVzaTZ6RnpjTEN6Q2tEWjVMSWhBVStNNFRQNVFDbkp6c2d5WEpyM2FUYnBTbUh1ODBtam91NzNHUnRKcWprUzNFNjRJQVVhR0dWNVNsTmFWMUpzV2xjeFNsSkRTVFpKYWtscFpsRTlQUT09IiwiUHJvb2YiOnsidG90YWwiOjEsImluZGV4IjowLCJsZWFmX2hhc2giOiJwakxJd2NRWWlYdEJ1K212dFVXaDNlVGRaYlNZN0RHbFh1VzZKeUxHRWFBPSIsImF1bnRzIjpbXX19LCJCbG9ja0hlaWdodCI6NTg5fQ==
	// redeemID = 3
	// eyJQcm9vZiI6eyJSb290SGFzaCI6IjM1NDA4MjMwM0U4NjY4ODM4MjZCODczODM4QzBBNjVGMjEwNzQ0NkE4RjVCNjMxNzUzRUYyMDgzNDdBQ0MxODEiLCJEYXRhIjoiMUFId1lsM3VDa1lxTElmNkNoOEtGTTJHbk5mZjArSUlnNHl3NDN3YmR5LzlzWjFuRWdjS0EwSk9RaEFLRWg4S0ZHemd3VWllVVQrdStoQ2YxdVV5MzQ1djlwOWpFZ2NLQTBKT1FoQUtFbXdLSnV0YTZZY2hBcUdMU0pHVlZjYWZqOFg2aWVRQUJ3dm5lNDZpT2F3enhyNVVuTms5VHZPNUVrQTRNSVdjdktvYXFuK1pzbCtybkt5VWhXYmIxWldFZHNlM2FxajlHRTJzNFhQb0xQbkJZY1duS0tpZVdqWVM5Q1hIYUNNQUpob3d4NzZRUURINmdSSWRJQVlhR0dWNVNsTmFWMUpzV2xjeFNsSkRTVFpKYWsxcFpsRTlQUT09IiwiUHJvb2YiOnsidG90YWwiOjEsImluZGV4IjowLCJsZWFmX2hhc2giOiJOVUNDTUQ2R2FJT0NhNGM0T01DbVh5RUhSR3FQVzJNWFUrOGdnMGVzd1lFPSIsImF1bnRzIjpbXX19LCJCbG9ja0hlaWdodCI6NjczfQ==
}
