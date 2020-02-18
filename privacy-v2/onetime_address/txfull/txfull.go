package txfull

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
)

type UTXO struct {
	mask      *privacy.Scalar
	amount    *privacy.Scalar
	txData    *privacy.Point
	addressee *privacy.Point // K^o = H_n(r * K_B^v )G + K_B^s
}

func CreateOutputs(addresses []*address.PublicAddress, money []*big.Int) ([]UTXO, error) {
	result := make([]UTXO, len(addresses))
	if len(addresses) > 256 {
		return nil, errors.New("Error in tx_full CreateOutputs: Cannot create too much output (maximum is 256)")
	}
	for i := 0; i < len(addresses); i += 1 {
		r := privacy.RandomScalar() // Step 1 Monero One-time address
		addressee, txData, cachedHash := ParseOTAWithCached(r, addresses[i], byte(i))
		mask, amount, err := ParseMoneyToCreateOutput(cachedHash, money[i], byte(i))
		if err != nil {
			return nil, errors.New("Error in tx_full CreateOutputs: money of the output is invalid")
		}
		result[i] = UTXO{mask, amount, txData, addressee}
	}
	return result, nil
}

func 