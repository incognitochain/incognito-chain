package txfull

import (
	"github.com/incognitochain/incognito-chain/privacy"
	ota "github.com/incognitochain/incognito-chain/privacy-v2/onetime_address"
)

func getSumBlindInput(this *RingCTFull) (*privacy.Scalar, error) {
	sumBlindInput := new(privacy.Scalar)
	for i := 0; i < len(this.inputs); i += 1 {
		blind, _, err := ota.ParseBlindAndMoneyFromUtxo(
			this.fromAddress[i],
			this.inputs[i],
		)
		if err != nil {
			return nil, err
		}
		sumBlindInput = sumBlindInput.Add(sumBlindInput, blind)
	}
	return sumBlindInput, nil
}

func getSumCommitment(arr []ota.UTXO) *privacy.Point {
	sum := new(privacy.Point)
	for i := 0; i < len(arr); i += 1 {
		sum = sum.Add(sum, arr[i].GetCommitment())
	}
	return sum
}

func getPrivateKeyOfInputs(this *RingCTFull) []privacy.Scalar {
	privateKeys := make([]privacy.Scalar, len(this.inputs))
	for i := 0; i < len(privateKeys); i += 1 {
		privateKeys[i] = *ota.ParseUTXOPrivatekey(this.fromAddress[i], this.inputs[i])
	}
	return privateKeys
}
