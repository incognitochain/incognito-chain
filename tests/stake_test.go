package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver"
	"log"
	"testing"
)

func TestStakeShard(t *testing.T) {
	res, err := readfile("./testsdata/stake/stakeshard.json")
	if err != nil {
		t.Fatal(err)
	} else {
		for _, step := range res.steps {
			log.Println(step)
		}
	}
	testResult, err := executeTest("./testsdata/stake/stakeshard.json")
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

func TestStakeBeacon(t *testing.T) {
	res, err := readfile("./testsdata/stake/stakebeacon.json")
	if err != nil {
		t.Fatal(err)
	} else {
		for _, step := range res.steps {
			log.Println(step)
		}
	}
	testResult, err := executeTest("./testsdata/stake/stakebeacon.json")
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

