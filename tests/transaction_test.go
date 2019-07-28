package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver"
	"log"
	"testing"
)

func TestGetTransactionByHash(t *testing.T) {
	testResult, err := executeTest("./testsdata/get_transaction_by_hash.json")
	if err != nil {
		t.Fatal(err)
	} else {
		log.Println(testResult)
	}
}
func TestCreateAndSendNormalTransaction(t *testing.T) {
	res, err := readfile("./testsdata/transaction.json")
	if err != nil {
		t.Fatal(err)
	} else {
		log.Println(res.steps[0], res.steps[1], res.steps[2], res.steps[3])
	}
	testResult, err := executeTest("./testsdata/transaction.json")
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
