package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"testing"
)

func TestStakeShard(t *testing.T) {
	var err error
	_, err = readfile("./testsdata/stake/stakeshard.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeTest("./testsdata/stake/stakeshard.json")
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

func TestStakeBeacon(t *testing.T) {
	var err error
	_, err = readfile("./testsdata/stake/stakebeacon.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeTest("./testsdata/stake/stakebeacon.json")
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

func TestStakeShardForOther(t *testing.T) {
	var err error
	_, err = readfile("./testsdata/stake/stakeshardforother.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeTest("./testsdata/stake/stakeshardforother.json")
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

func TestStakeBeaconForOther(t *testing.T) {
	var err error
	_, err = readfile("./testsdata/stake/stakebeaconforother.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeTest("./testsdata/stake/stakebeaconforother.json")
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
