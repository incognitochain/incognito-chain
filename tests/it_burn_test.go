package main

import (
	"testing"

	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func TestBurnPrivateETH(t *testing.T) {
	_, err := executeTest("./testsdata/burn/burneth.json")
	checkError(t, err)
}

func TestBurnPrivateERC20(t *testing.T) {
	_, err := executeTest("./testsdata/burn/burnerc20.json")
	checkError(t, err)
}

func checkError(t *testing.T, err error) {
	if err == nil {
		return
	}

	if rpcError, ok := err.(*rpcservice.RPCError); !ok || rpcError != nil {
		t.Fatalf("%+v", err)
	}
}
