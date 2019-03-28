package main

import (
	"github.com/constant-money/constant-chain/utility/generateKeys/generator"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

type KeyPairs struct {
	KeyPair []KeyPair
}

func main() {
	generator.GenerateAddressByShard(0)
}
