package main

import (
	"encoding/json"
	"testing"
)

func TestMakeRPCRequest(t *testing.T) {
	client := newClientWithHost("http://localhost", "9334")
	res, rpcErr := makeRPCRequest(client, "getblockchaininfo", []string{})
	if rpcErr != nil {
		t.Fatal(rpcErr)
	}
	result := make(map[string]interface{})
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		t.Fatal(err)
	}
}