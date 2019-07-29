package main

import (
	"encoding/json"
	"log"
	"testing"
	"time"
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

func TestMakeWsRequest(t *testing.T) {
	client := newClientWithHost("127.0.0.1", "19334")
	result, rpcErr := makeWsRequest(client, "gettransactionbyhash", 100*time.Second, []interface{}{"fb48d9ae4736d2a1ac698920710770496d8c44dff858d1d8f6c55858b5580a74"})
	if rpcErr != nil {
		t.Fatal(rpcErr)
	}
	log.Println(result)
}