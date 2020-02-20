package txfull

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy-v2/mlsag"
	ota "github.com/incognitochain/incognito-chain/privacy-v2/onetime_address"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
)

func Create_Addresses_With_Specified_Money(n, sumSpecifiedMoney int) (*[]address.PrivateAddress, *[]address.PublicAddress, *[]big.Int) {
	// Initialize params to make input
	sum := 0
	money := make([]big.Int, n)
	privAddresses := make([]address.PrivateAddress, n)
	pubAddresses := make([]address.PublicAddress, n)
	for i := 0; i < n-1; i += 1 {
		privAddresses[i] = address.GenerateRandomAddress()
		pubAddresses[i] = privAddresses[i].GetPublicAddress()

		curMoneyInt := common.RandIntInterval(10, 100)
		sum += curMoneyInt

		curMoney, _ := new(big.Int).SetString(strconv.Itoa(curMoneyInt), 10)
		money[i] = *curMoney
	}
	privAddresses[n-1] = address.GenerateRandomAddress()
	pubAddresses[n-1] = privAddresses[n-1].GetPublicAddress()

	curMoneyInt := sumSpecifiedMoney - sum
	curMoney, _ := new(big.Int).SetString(strconv.Itoa(curMoneyInt), 10)
	money[n-1] = *curMoney

	return &privAddresses, &pubAddresses, &money
}

func TestMultipleWorkflow(t *testing.T) {
	// When testing with hand should make it iterate 500 times for sure
	for i := 0; i < 5; i += 1 {
		j := common.RandInt() % 2
		if j == 0 {
			TestFailTxFullWorkflow(t)
		} else {
			TestCorrectTxFullWorkflow(t)
		}
	}
}

func TestFailTxFullWorkflow(t *testing.T) {
	n_input := common.RandIntInterval(20, 40)
	n_output := common.RandIntInterval(20, 40)

	// 1000 coin
	sumMoneyShouldBe := common.RandIntInterval(50000, 100000)

	aliceAddresses, alicePublicAddresses, money_input := Create_Addresses_With_Specified_Money(n_input, sumMoneyShouldBe)
	_, bobPublicAddresses, money_output := Create_Addresses_With_Specified_Money(n_output, sumMoneyShouldBe+1)

	// Create inputs, outputs
	inputs, _, err_inp := ota.CreateOutputs(alicePublicAddresses, money_input)
	outputs, sumBlindOutput, err_out := ota.CreateOutputs(bobPublicAddresses, money_output)

	assert.Equal(t, nil, err_inp, "Should not have any error in correct workflow txfull")
	assert.Equal(t, nil, err_out, "Should not have any error in correct workflow txfull")

	// Create signature
	ringctfull := NewRingCTFull(
		inputs,
		aliceAddresses, //private keys to spend
		sumBlindOutput,
		outputs,
		bobPublicAddresses,
	)
	message := "Some f******* message that can be changed with the transaction message :D"

	ring, privateKeys, pi, err := ringctfull.CreateRandomRing(message)
	assert.Equal(t, nil, err, "Should not have any error in correct workflow txfull")

	signer := mlsag.NewMlsagWithDefinedRing(privateKeys, ring, pi)
	signature, err_sig := signer.Sign(message)

	assert.Equal(t, nil, err_sig, "Should not have any error in correct workflow txfull")

	check, err := mlsag.Verify(signature, ring, message)
	assert.Equal(t, false, check, "Should verify fail when workflow is wrong")
}

func TestCorrectTxFullWorkflow(t *testing.T) {
	n_input := common.RandIntInterval(20, 40)
	n_output := common.RandIntInterval(20, 40)

	// 1000 coin
	sumMoneyShouldBe := common.RandIntInterval(50000, 100000)

	aliceAddresses, alicePublicAddresses, money_input := Create_Addresses_With_Specified_Money(n_input, sumMoneyShouldBe)
	_, bobPublicAddresses, money_output := Create_Addresses_With_Specified_Money(n_output, sumMoneyShouldBe)

	// Create inputs, outputs
	inputs, _, err_inp := ota.CreateOutputs(alicePublicAddresses, money_input)
	outputs, sumBlindOutput, err_out := ota.CreateOutputs(bobPublicAddresses, money_output)

	assert.Equal(t, nil, err_inp, "Should not have any error in correct workflow txfull")
	assert.Equal(t, nil, err_out, "Should not have any error in correct workflow txfull")

	// Create signature
	ringctfull := NewRingCTFull(
		inputs,
		aliceAddresses, //private keys to spend
		sumBlindOutput,
		outputs,
		bobPublicAddresses,
	)
	message := "Some f******* message that can be changed with the transaction message :D"

	ring, privateKeys, pi, err := ringctfull.CreateRandomRing(message)
	assert.Equal(t, nil, err, "Should not have any error in correct workflow txfull")

	signer := mlsag.NewMlsagWithDefinedRing(privateKeys, ring, pi)
	signature, err_sig := signer.Sign(message)

	assert.Equal(t, nil, err_sig, "Should not have any error in correct workflow txfull")

	check, err := mlsag.Verify(signature, ring, message)
	assert.Equal(t, true, check, "Should verify true when workflow correct")

	if check == false {
		fmt.Println("Money Inputs")
		fmt.Println(money_input)
		fmt.Println("============")

		fmt.Println("Alice addresses")
		fmt.Println(aliceAddresses)
		fmt.Println("============")

		fmt.Println("SumBlind")
		fmt.Println(sumBlindOutput)
		fmt.Println("============")

		fmt.Println("Money Outputs")
		fmt.Println(money_output)
		fmt.Println("============")

		fmt.Println("Bob public addresses")
		fmt.Println(bobPublicAddresses)
		fmt.Println("============")
	}
}
