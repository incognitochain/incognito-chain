package txfull

import (
	"github.com/incognitochain/incognito-chain/operation"
	ota "github.com/incognitochain/incognito-chain/privacy/privacy-v2/onetime_address"
	"github.com/incognitochain/incognito-chain/privacy/privacy-v2/onetime_address/utxo"
)

func getSumBlindInput(this *RingCTFull) (*operation.Scalar, error) {
	sumBlindInput := new(operation.Scalar)
	for i := 0; i < len(this.inputs); i += 1 {
		blind, _, err := ota.ParseBlindAndMoneyFromUtxo(
			&this.fromAddress[i],
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

func (this *RingCTFull) getPrivateKeyOfInputs() *[]operation.Scalar {
	privateKeys := make([]operation.Scalar, len(this.inputs))
	for i := 0; i < len(privateKeys); i += 1 {
		privateKeys[i] = *ota.ParseUtxoPrivatekey(&this.fromAddress[i], &this.inputs[i])
	}
	return &privateKeys
}
