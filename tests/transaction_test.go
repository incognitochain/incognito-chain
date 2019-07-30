package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver"
	"log"
	"testing"
)

func TestGetTransactionByHash(t *testing.T) {
	testResult, err := executeTest("./testsdata/transaction/get_transaction_by_hash.json")
	if err != nil {
		t.Fatal(err)
	} else {
		log.Println(testResult)
	}
}
func TestCreateAndSendNormalTransaction(t *testing.T) {
	res, err := readfile("./testsdata/transaction/transaction.json")
	if err != nil {
		t.Fatal(err)
	} else {
		for _, step := range res.steps {
			log.Println(step)
		}
	}
	testResult, err := executeTest("./testsdata/transaction/transaction.json")
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
func TestCreateAndSendStakeShardTransaction(t *testing.T) {
	res, err := readfile("./testsdata/transaction/stakeshard.json")
	if err != nil {
		t.Fatal(err)
	} else {
		for _, step := range res.steps {
			log.Println(step)
		}
	}
	testResult, err := executeTest("./testsdata/transaction/stakeshard.json")
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

func TestCreateAndSendStakeBeaconTransaction(t *testing.T) {
	res, err := readfile("./testsdata/transaction/stakebeacon.json")
	if err != nil {
		t.Fatal(err)
	} else {
		for _, step := range res.steps {
			log.Println(step)
		}
	}
	testResult, err := executeTest("./testsdata/transaction/stakebeacon.json")
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
