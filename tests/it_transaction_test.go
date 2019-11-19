package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"testing"
)

//func TestGetTransactionByHash(t *testing.T) {
//	testResult, err := executeTest("./testsdata/transaction/get_transaction_by_hash.json")
//	if err != nil {
//		t.Fatal(err)
//	} else {
//		log.Println(testResult)
//	}
//}
func TestCreateAndSendNormalTransaction(t *testing.T) {
	var err error
	_, err = readfile("./testsdata/transaction/normal_transaction.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeTest("./testsdata/transaction/normal_transaction.json")
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

func TestCreateAndSendCustomTokenPrivacyTransaction(t *testing.T) {
	var err error
	_, err = readfile("./testsdata/transaction/custom_token_privacy_transaction.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeTest("./testsdata/transaction/custom_token_privacy_transaction.json")
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
