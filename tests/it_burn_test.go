package main

import (
	"testing"

	"github.com/incognitochain/incognito-chain/rpcserver"
)

func TestBurnPrivateETH(t *testing.T) {
	_, err := executeTest("./testsdata/burn/burneth.json")
	checkError(t, err)
}

func checkError(t *testing.T, err error) {
	if err == nil {
		return
	}

	if rpcError, ok := err.(*rpcserver.RPCError); !ok || rpcError != nil {
		t.Fatal(err)
	}
}
