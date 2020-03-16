package onetime_address

import (
	"errors"

	"github.com/incognitochain/incognito-chain/privacy/address"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
)

// Create output of coins and sum of blind values for later usage
func CreateOutputs(publicKeys []*address.PublicAddress, moneys []uint64) ([]*coin.Coin_v2, *operation.Scalar, error) {

	result := make([]*coin.Coin_v2, len(publicKeys))
	if len(publicKeys) > privacy_util.MaxOutputCoin {
		return nil, nil, errors.New("Error in tx_full CreateOutputs: Cannot create too much output (maximum is 256)")
	}
	if len(publicKeys) != len(moneys) {
		return nil, nil, errors.New("Error in tx_full CreateOutputs: addresses and money array length not the same")
	}
	sumBlind := new(operation.Scalar)
	for i := 0; i < len(publicKeys); i += 1 {
		r := operation.RandomScalar()            // Step 1 Monero One-time address
		blind := operation.RandomScalar()        // Also create blind for money
		sumBlind = sumBlind.Add(sumBlind, blind) // The creator of outputs should know sum of blind for signature

		oneTimeAddress, txData, cachedHash := parseOTAWithCached(r, publicKeys[i], byte(i))
		mask, amount, commitment, err := parseMoneyToCreateOutput(blind, cachedHash, moneys[i], byte(i))
		if err != nil {
			return nil, nil, errors.New("Error in tx_full CreateOutputs: money of the output is invalid")
		}
		result[i] = coin.NewCoinv2(mask, amount, txData, oneTimeAddress, commitment, uint8(i), []byte{})
	}
	return result, sumBlind, nil
}

// Check whether the utxo is from this address
func IsCoinOfAddress(addr *address.PrivateAddress, utxo *coin.Coin_v2) bool {
	rK := new(operation.Point).ScalarMult(utxo.GetTxRandom(), addr.GetPrivateView())

	hashed := operation.HashToScalar(
		append(rK.ToBytesS(), utxo.GetIndex()),
	)
	HnG := new(operation.Point).ScalarMultBase(hashed)

	KCheck := new(operation.Point).Sub(utxo.GetPublicKey(), HnG)

	// TODO
	return *KCheck == *addr.GetPublicSpend()
}
