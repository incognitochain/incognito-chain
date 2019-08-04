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
	client := newClientWithFullInform("127.0.0.1", "9334", "19334")
	result, rpcErr := makeWsRequest(client, "subcribependingtransaction", 10*time.Second, "ac9db9a149a892da81b5d6521f3296b6524e331893a133adc1e77c15186c1907")
	if rpcErr != nil {
		t.Log(result, rpcErr)
	} else {
		t.Fatal(result, rpcErr)
	}
}