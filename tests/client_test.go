package main

import (
	"encoding/json"
	"testing"
)

func TestMakeRPCRequest(t *testing.T) {
	res, rpcErr := makeRPCRequest("http://localhost", "9334", "getblockchaininfo", []string{})
	if rpcErr != nil {
		t.Fatal(rpcErr)
	}
	result := make(map[string]interface{})
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		t.Fatal(err)
	}
}