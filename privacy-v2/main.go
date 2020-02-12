package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy-v2/mlsag"
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
)

func main() {
	fmt.Println("Running test")
	keyInputs := []C25519.Key{}
	for i := 0; i < 3; i += 1 {
		privateKey := C25519.RandomScalar()
		keyInputs = append(keyInputs, *privateKey)
	}
	signature := mlsag.SignCore(keyInputs, "Hello", 5)
	fmt.Println(signature)
}
