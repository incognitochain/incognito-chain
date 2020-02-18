package main

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy-v2/mlsag"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address"
	ota "github.com/incognitochain/incognito-chain/privacy-v2/onetime_address"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
)

func testMlsag() {
	fmt.Println("Running test")
	keyInputs := []privacy.Scalar{}
	for i := 0; i < 8; i += 1 {
		privateKey := privacy.RandomScalar()
		keyInputs = append(keyInputs, *privateKey)
	}
	numFake := 3
	pi := common.RandInt() % numFake
	ring := mlsag.NewRandomRing(keyInputs, numFake, pi)
	signer := mlsag.NewMlsagWithDefinedRing(keyInputs, ring, pi, numFake)

	signature, err := signer.Sign("Hello")
	if err != nil {
		fmt.Println("There is something wrong with signing")
		fmt.Println(err)
	}
	// ring = mlsag.NewRandomRing(keyInputs, numFake, pi)
	check, err := mlsag.Verify(signature, ring, "Hello")
	if err != nil {
		fmt.Println("There is something wrong with verifying")
		fmt.Println(err)
	}
	fmt.Println("Check signature:")
	fmt.Println(check)

	b, err := signature.ToHex()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Printing signature")
	fmt.Println(signature)
	fmt.Println("==============")
	fmt.Println("Printing signature hex")
	fmt.Println(b)
	sig, _ := new(mlsag.Signature).FromHex(b)

	fmt.Println("==============")
	fmt.Println("Printing signature from hex")
	fmt.Println(sig)

	fmt.Println("==============")
	fmt.Println("Checking 2 signatures are the same")
	b1, _ := sig.ToBytes()
	b2, _ := signature.ToBytes()
	fmt.Println(bytes.Equal(b1, b2))
}

func main() {
	// aliceAddress := address.GenerateRandomAddress()

	const n int = 10
	money := make([]*big.Int, n)
	peopleAddresses := make([]*address.PrivateAddress, n)
	peoplePublicAddresses := make([]*address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		money[i], _ = new(big.Int).SetString("100", 10)
		peopleAddresses[i] = address.GenerateRandomAddress()
		peoplePublicAddresses[i] = peopleAddresses[i].GetPublicAddress()
	}
	outputs, err := onetime_address.CreateOutputs(peoplePublicAddresses, money)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(outputs); i += 1 {
		fmt.Println("===========")
		fmt.Println(ota.IsUtxoOfAddress(peopleAddresses[i], outputs[i]))
		break
		// nxt := (i + 1) % len(outputs)
		// fmt.Println(ota.IsUtxoOfAddress(peopleAddresses[i], outputs[nxt]))
	}
}
