package onetime_address

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
)

func CreateOutputs(addresses []address.PublicAddress, money []big.Int) ([]UTXO, *privacy.Scalar, error) {
	result := make([]UTXO, len(addresses))
	if len(addresses) > 256 {
		return nil, nil, errors.New("Error in tx_full CreateOutputs: Cannot create too much output (maximum is 256)")
	}
	if len(addresses) != len(money) {
		return nil, nil, errors.New("Error in tx_full CreateOutputs: addresses and money array length not the same")
	}
	sumBlind := new(privacy.Scalar)
	for i := 0; i < len(addresses); i += 1 {
		r := privacy.RandomScalar()              // Step 1 Monero One-time address
		blind := privacy.RandomScalar()          // Also create blind for money
		sumBlind = sumBlind.Add(sumBlind, blind) // The creator of outputs should know sum of blind for signature

		addressee, txData, cachedHash := parseOTAWithCached(r, addresses[i], byte(i))
		mask, amount, commitment, err := parseMoneyToCreateOutput(blind, cachedHash, money[i], byte(i))
		if err != nil {
			return nil, nil, errors.New("Error in tx_full CreateOutputs: money of the output is invalid")
		}
		result[i] = UTXO{uint8(i), mask, amount, txData, addressee, commitment}
	}
	return result, sumBlind, nil
}

func IsUtxoOfAddress(addr address.PrivateAddress, utxo UTXO) bool {
	rK := new(privacy.Point).ScalarMult(utxo.GetTxData(), addr.GetPrivateView())

	hashed := privacy.HashToScalar(
		append(rK.ToBytesS(), utxo.GetIndex()),
	)
	HnG := new(privacy.Point).ScalarMultBase(hashed)

	KCheck := new(privacy.Point).Sub(utxo.GetAddressee(), HnG)
	return *KCheck == *addr.GetPublicSpend()
}
