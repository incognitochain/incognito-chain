package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"testing"
)

func TestCreateAndSendCrossNormalTransaction(t *testing.T) {
	var err error
	_, err = readfile("./testsdata/transaction/cross_normal_transaction.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeTest("./testsdata/transaction/cross_normal_transaction.json")
	if err != nil {
		if rpcError, ok := err.(*rpcservice.RPCError); ok {
			if rpcError != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}
}

func TestCreateAndSendCrossCustomTokenPrivacyTransaction(t *testing.T) {
	var err error
	_, err = readfile("./testsdata/transaction/cross_custom_token_privacy_transaction.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeTest("./testsdata/transaction/cross_custom_token_privacy_transaction.json")
	if err != nil {
		if rpcError, ok := err.(*rpcservice.RPCError); ok {
			if rpcError != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}
}
