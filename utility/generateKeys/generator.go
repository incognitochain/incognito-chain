package main

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/utility/generateKeys/generator"
	"io/ioutil"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

type KeyPairs struct {
	KeyPair []KeyPair
}

func main() {
	keys, _ := generator.GenerateAddressByShard(1)
	file, _ := json.MarshalIndent(keys, "", " ")
	_ = ioutil.WriteFile("private-keys-shard-1-1.json", file, 0644)
}
