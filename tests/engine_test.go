package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"log"
	"testing"
)

func TestExecuteTest(t *testing.T) {
	res, err := executeTest("./testsdata/sample.json")
	if err != nil {
		if rpcError, ok := err.(*rpcservice.RPCError); ok {
			if rpcError != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}
	log.Println(res)
}
