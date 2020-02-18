package onetime_address

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
)

type UTXO struct {
	index     uint8
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
		result[i] = UTXO{uint8(i), mask, amount, txData, addressee}
	}
	return result, nil
}

func (this *UTXO) GetIndex() uint8              { return this.index }
func (this *UTXO) GetMask() *privacy.Scalar     { return this.mask }
func (this *UTXO) GetAmount() *privacy.Scalar   { return this.amount }
func (this *UTXO) GetTxData() *privacy.Point    { return this.txData }
func (this *UTXO) GetAddressee() *privacy.Point { return this.addressee }

func (this *UTXO) SetMask(mask *privacy.Scalar)          { this.mask = mask }
func (this *UTXO) SetAmount(amount *privacy.Scalar)      { this.amount = amount }
func (this *UTXO) SetTxData(txData *privacy.Point)       { this.txData = txData }
func (this *UTXO) SetAddressee(addressee *privacy.Point) { this.addressee = addressee }

func IsUtxoOfAddress(this *address.PrivateAddress, output UTXO) bool {
	rK := new(privacy.Point).ScalarMult(output.GetTxData(), this.GetPrivateView())

	hashed := privacy.HashToScalar(
		append(rK.ToBytesS(), output.GetIndex()),
	)
	HnG := new(privacy.Point).ScalarMultBase(hashed)

	KCheck := new(privacy.Point).Sub(output.GetAddressee(), HnG)
	return *KCheck == *this.GetPublicSpend()
}
