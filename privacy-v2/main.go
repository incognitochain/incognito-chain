package main

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy-v2/mlsag"
	ota "github.com/incognitochain/incognito-chain/privacy-v2/onetime_address"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
	"github.com/incognitochain/incognito-chain/privacy-v2/txfull"
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
	ring := mlsag.NewRandomRing(&keyInputs, numFake, pi)
	signer := mlsag.NewMlsagWithDefinedRing(&keyInputs, ring, pi)

	signature, err := signer.Sign("Hello")
	if err != nil {
		fmt.Println("There is something wrong with signing")
		fmt.Println(err)
	}
	// ring = mlsag.NewRandomRing(&keyInputs, numFake, pi)
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

func testOTA() {
	const n int = 3
	money := make([]big.Int, n)
	peopleAddresses := make([]address.PrivateAddress, n)
	peoplePublicAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		peopleAddresses[i] = address.GenerateRandomAddress()
		peoplePublicAddresses[i] = peopleAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString("100", 10)
		money[i] = *curMoney
	}
	outputsPointer, _, err := ota.CreateOutputs(&peoplePublicAddresses, &money)
	outputs := *outputsPointer
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(outputs); i += 1 {
		mask, amount, err := ota.ParseBlindAndMoneyFromUtxo(peopleAddresses[i], outputs[i])
		fmt.Println("Mask =")
		fmt.Println(mask)
		fmt.Println("Amount =")
		fmt.Println(amount)
		fmt.Println("Err =")
		fmt.Println(err)
	}
}

func testTxFull() {
	// Example of Alice has 5 inputs, give 10 outputs to Bob
	moneyInput := "100"
	moneyOutput := "49"

	// 5*100 = 500 = 50*10
	n_input := 5
	n_output := 10

	// Initialize params to make input
	money := make([]big.Int, n_input)
	aliceAddresses := make([]address.PrivateAddress, n_input)
	alicePublicAddresses := make([]address.PublicAddress, n_input)
	for i := 0; i < n_input; i += 1 {
		aliceAddresses[i] = address.GenerateRandomAddress()
		alicePublicAddresses[i] = aliceAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString(moneyInput, 10)
		money[i] = *curMoney
	}

	// Initialize params to make output
	n_output = n_output + 1
	money_output := make([]big.Int, n_output)
	bobAddresses := make([]address.PrivateAddress, n_output)
	bobPublicAddresses := make([]address.PublicAddress, n_output)
	for i := 0; i < n_output; i += 1 {
		bobAddresses[i] = address.GenerateRandomAddress()
		bobPublicAddresses[i] = bobAddresses[i].GetPublicAddress()
		var curMoney *big.Int
		if i == n_output-1 {
			curMoney, _ = new(big.Int).SetString("10", 10)
		} else {
			curMoney, _ = new(big.Int).SetString(moneyOutput, 10)
		}
		money_output[i] = *curMoney
	}

	// Create inputs, outputs
	inputs, _, err_inp := ota.CreateOutputs(&alicePublicAddresses, &money)
	outputs, sumBlindOutput, err_out := ota.CreateOutputs(&bobPublicAddresses, &money_output)
	if err_inp != nil || err_out != nil {
		fmt.Println(err_inp)
		fmt.Println(err_out)
		return
	}

	// Create signature
	ringctfull := txfull.NewRingCTFull(
		inputs,
		&aliceAddresses, //private keys
		sumBlindOutput,
		outputs,
		&bobPublicAddresses,
	)
	message := "Some f******* message that can be changed with the transaction message :D"
	ring, privateKeys, pi, err := ringctfull.CreateRandomRing(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	signer := mlsag.NewMlsagWithDefinedRing(privateKeys, ring, pi)
	signature, err_sig := signer.Sign(message)
	if err_sig != nil {
		fmt.Println(err_sig)
		return
	}

	check, err := mlsag.Verify(signature, ring, message)
	fmt.Println(err)
	fmt.Println(check)
}

func main() {
	testTxFull()
}
