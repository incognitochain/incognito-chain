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
	log.Println("Test Make Ws Request")
	client := newClientWithHost("localhost", "19334")
	result, rpcErr := makeWsRequest(client, "subcribependingtransaction", 10*time.Second, "a9846c6545c62b11de7548ee29ad6d6b2adac0ffcdd385b7213ba628fbc1c08d")
	if rpcErr != nil {
		t.Fatal(rpcErr)
	}
	log.Println(result)
}