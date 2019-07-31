package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver"
	"log"
	"testing"
)

func TestCreateAndSendCrossNormalTransaction(t *testing.T) {
	_, err := readfile("./testsdata/transaction/cross_normal_transaction.json")
	if err != nil {
		t.Fatal(err)
	}
	testResult, err := executeTest("./testsdata/transaction/cross_normal_transaction.json")
	if err != nil {
		if rpcError, ok := err.(*rpcserver.RPCError); ok {
			if rpcError != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		log.Println(testResult)
	}
}
func TestCreateAndSendCrossCustomTokenTransaction(t *testing.T) {
	_, err := readfile("./testsdata/transaction/cross_custom_token_transaction.json")
	if err != nil {
		t.Fatal(err)
	}
	testResult, err := executeTest("./testsdata/transaction/cross_custom_token_transaction.json")
	if err != nil {
		if rpcError, ok := err.(*rpcserver.RPCError); ok {
			if rpcError != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		log.Println(testResult)
	}
}
func TestCreateAndSendCrossCustomTokenPrivacyTransaction(t *testing.T) {
	_, err := readfile("./testsdata/transaction/cross_custom_token_privacy_transaction.json")
	if err != nil {
		t.Fatal(err)
	}
	testResult, err := executeTest("./testsdata/transaction/cross_custom_token_privacy_transaction.json")
	if err != nil {
		if rpcError, ok := err.(*rpcserver.RPCError); ok {
			if rpcError != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		log.Println(testResult)
	}
}
