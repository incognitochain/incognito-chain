package onetime_address

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/incognitochain/incognito-chain/privacy/address"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowOnetimeAddress(t *testing.T) {
	const n int = 3
	money := make([]big.Int, n)
	peopleAddresses := make([]address.PrivateAddress, n)
	peoplePublicAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		peopleAddresses[i] = *address.GenerateRandomAddress()
		peoplePublicAddresses[i] = *peopleAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString("100", 10)
		money[i] = *curMoney
	}
	outputsPtr, _, err := CreateOutputs(&peoplePublicAddresses, &money)
	outputs := *outputsPtr
	assert.Equal(t, nil, err, "Should not have error when creating outputs")

	for i := 0; i < len(outputs); i += 1 {
		_, amount, err := ParseBlindAndMoneyFromUtxo(&peopleAddresses[i], &outputs[i])
		assert.Equal(t, nil, err, "Should not have any error when parsing money")
		assert.Equal(t, uint8(100), amount.GetKey()[0], "Amount when parse should be 100")
	}
}

func TestParseUtxoPrivateKey(t *testing.T) {
	const n int = 3
	money := make([]big.Int, n)
	peopleAddresses := make([]address.PrivateAddress, n)
	peoplePublicAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		peopleAddresses[i] = *address.GenerateRandomAddress()
		peoplePublicAddresses[i] = *peopleAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString("100", 10)
		money[i] = *curMoney
	}
	outputsPointer, _, err := CreateOutputs(&peoplePublicAddresses, &money)
	outputs := *outputsPointer
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(outputs); i += 1 {
		priv := ParseUtxoPrivatekey(&peopleAddresses[i], &outputs[i])
		check := new(operation.Point).ScalarMultBase(priv)
		assert.Equal(t, check.GetKey(), outputs[i].GetAddressee().GetKey(), "Parse private key of utxo should return correctly")
	}
}

func TestParseBlindAndMoneyFromUtxo(t *testing.T) {
	const n int = 4
	money := make([]big.Int, n)
	peopleAddresses := make([]address.PrivateAddress, n)
	peoplePublicAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		peopleAddresses[i] = *address.GenerateRandomAddress()
		peoplePublicAddresses[i] = *peopleAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString("10", 10)
		money[i] = *curMoney
	}
	outputsPointer, sumBlind, err := CreateOutputs(&peoplePublicAddresses, &money)
	assert.Equal(t, nil, err, "There should not be any error in creating output")

	outputs := *outputsPointer

	sumMoney := new(operation.Scalar)
	for i := 0; i < len(outputs); i += 1 {
		blind, money, _ := ParseBlindAndMoneyFromUtxo(&peopleAddresses[i], &outputs[i])
		sumBlind = sumBlind.Sub(sumBlind, blind)
		sumMoney = sumMoney.Add(sumMoney, money)
	}
	keySum := sumMoney.GetKey()
	// n * 10 = 4 * 10 = 40
	assert.Equal(t, uint8(40), keySum[0], "Money should have right amount")
	assert.Equal(t, true, sumBlind.IsZero(), "Blind should have right amount")
}

func TestIsUtxoOfAddress(t *testing.T) {
	const n int = 4
	money := make([]big.Int, n)
	peopleAddresses := make([]address.PrivateAddress, n)
	peoplePublicAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		peopleAddresses[i] = *address.GenerateRandomAddress()
		peoplePublicAddresses[i] = *peopleAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString("10", 10)
		money[i] = *curMoney
	}
	outputsPointer, _, err := CreateOutputs(&peoplePublicAddresses, &money)
	assert.Equal(t, nil, err, "There should not be any error in creating output")

	outputs := *outputsPointer
	for i := 0; i < len(outputs); i += 1 {
		check := IsUtxoOfAddress(&peopleAddresses[i], &outputs[i])
		assert.Equal(t, true, check, "IsUtxo should detect correct")

		another := address.GenerateRandomAddress()
		check = IsUtxoOfAddress(another, &outputs[i])
		assert.Equal(t, false, check, "IsUtxo should detect correct")
	}
}

func TestCreateOutputFail1(t *testing.T) {
	const n int = 300
	money := make([]big.Int, n)
	peopleAddresses := make([]address.PrivateAddress, n)
	peoplePublicAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		peopleAddresses[i] = *address.GenerateRandomAddress()
		peoplePublicAddresses[i] = *peopleAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString("10", 10)
		money[i] = *curMoney
	}
	_, _, err := CreateOutputs(&peoplePublicAddresses, &money)
	assert.NotEqual(t, nil, err, "Should have error because length is too long")
}

func TestCreateOutputFail2(t *testing.T) {
	const n int = 10
	money := make([]big.Int, n+1)
	peopleAddresses := make([]address.PrivateAddress, n)
	peoplePublicAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		peopleAddresses[i] = *address.GenerateRandomAddress()
		peoplePublicAddresses[i] = *peopleAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString("10", 10)
		money[i] = *curMoney
	}
	_, _, err := CreateOutputs(&peoplePublicAddresses, &money)
	assert.NotEqual(t, nil, err, "Should have error because length of money and address is not the same")
}

func TestCreateOutputFail3(t *testing.T) {
	const n int = 10
	money := make([]big.Int, n)
	peopleAddresses := make([]address.PrivateAddress, n)
	peoplePublicAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n; i += 1 {
		peopleAddresses[i] = *address.GenerateRandomAddress()
		peoplePublicAddresses[i] = *peopleAddresses[i].GetPublicAddress()
		curMoney, _ := new(big.Int).SetString("99999999999999999999999999999999999999999999999999999999999999999999999999999999", 10)
		money[i] = *curMoney
	}
	_, _, err := CreateOutputs(&peoplePublicAddresses, &money)
	assert.NotEqual(t, nil, err, "Should have error because the money is too big")
}
