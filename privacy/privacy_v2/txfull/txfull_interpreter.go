package txfull

import (
	"github.com/incognitochain/incognito-chain/privacy/address"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	ota "github.com/incognitochain/incognito-chain/privacy/privacy_v2/onetime_address"
)

func parsePublicKey(privateKey *operation.Scalar) *operation.Point {
	return new(operation.Point).ScalarMultBase(privateKey)
}

func getSumBlindInput(this *RingCTFull) (*operation.Scalar, error) {
	sumBlindInput := new(operation.Scalar)
	for i := 0; i < len(this.inputs); i += 1 {
		blind, _, err := ota.ParseBlindAndMoneyFromUtxo(
			this.privateAddress,
			this.inputs[i],
		)
		if err != nil {
			return nil, err
		}
		sumBlindInput = sumBlindInput.Add(sumBlindInput, blind)
	}
	return sumBlindInput, nil
}

func getBlindInput(privAddress *address.PrivateAddress, coin *coin.Coin_v2) (*operation.Scalar, error) {
	blind, _, err := ota.ParseBlindAndMoneyFromUtxo(privAddress, coin)
	if err != nil {
		return nil, err
	} else {
		return blind, nil
	}

}

func getSumCommitment(arr []*coin.Coin_v2) *operation.Point {
	sum := new(operation.Point).Identity()
	for i := 0; i < len(arr); i += 1 {
		sum.Add(sum, arr[i].GetCommitment())
	}
	return sum
}

func getTxPrivateKey(privAddress *address.PrivateAddress, inputCoin *coin.Coin_v2) *operation.Scalar {
	privKey := ota.ParseUtxoPrivatekey(privAddress, inputCoin)
	return privKey
}

