package tests

import (
	"encoding/json"
	"log"
	"testing"
)

func TestMakeRPCRequest(t *testing.T) {
	res, err := makeRPCRequest("http://localhost", "9334", "getblockchaininfo", []string{})
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(res.Result, &result)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(result)
}