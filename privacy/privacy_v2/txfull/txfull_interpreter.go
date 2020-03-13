package txfull

import (
	"github.com/incognitochain/incognito-chain/privacy/operation"
	ota "github.com/incognitochain/incognito-chain/privacy/privacy_v2/onetime_address"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/onetime_address/utxo"
)

func getSumBlindInput(this *RingCTFull) (*operation.Scalar, error) {
	sumBlindInput := new(operation.Scalar)
	for i := 0; i < len(this.inputs); i += 1 {
		blind, _, err := ota.ParseBlindAndMoneyFromUtxo(
			this.privateAddress,
			&this.inputs[i],
		)
		if err != nil {
			return nil, err
		}
		sumBlindInput = sumBlindInput.Add(sumBlindInput, blind)
	}
	return sumBlindInput, nil
}

func getSumCommitment(arr []utxo.Utxo) *operation.Point {
	sum := new(operation.Point)
	for i := 0; i < len(arr); i += 1 {
		sum = sum.Add(sum, arr[i].GetCommitment())
	}
	return sum
}

func getTxPrivateKeys(this *RingCTFull) *[]operation.Scalar {
	privAddress := this.privateAddress
	inputCoins := this.inputs
	privateKeys := make([]operation.Scalar, len(inputCoins))
	for i := 0; i < len(privateKeys); i += 1 {
		privateKeys[i] = *ota.ParseUtxoPrivatekey(privAddress, &inputCoins[i])
	}
	return &privateKeys
}
