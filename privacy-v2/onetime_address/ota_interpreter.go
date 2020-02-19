package onetime_address

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
)

func ParseUTXOPrivatekey(addr address.PrivateAddress, utxo UTXO) *privacy.Scalar {
	rK := new(privacy.Point).ScalarMult(utxo.GetTxData(), addr.GetPrivateView())
	hashed := privacy.HashToScalar(
		append(rK.ToBytesS(), utxo.GetIndex()),
	)
	return new(privacy.Scalar).Add(hashed, addr.GetPrivateSpend())
}

// Step 1 Monero One-time Address
func parseAddresseeWithCached(r *privacy.Scalar, addr address.PublicAddress, index byte) (*privacy.Point, *privacy.Scalar) {
	rK := new(privacy.Point).ScalarMult(addr.GetPublicView(), r)
	cachedHash := privacy.HashToScalar(
		append(rK.ToBytesS(), index),
	)
	HrKG := new(privacy.Point).ScalarMultBase(cachedHash)
	addressee := new(privacy.Point).Add(HrKG, addr.GetPublicSpend())

	return addressee, cachedHash
}

// Step 2 Monero One-time Address
func parseOTAWithCached(r *privacy.Scalar, addr address.PublicAddress, index byte) (addressee *privacy.Point, txData *privacy.Point, cachedHash *privacy.Scalar) {
	addressee, cachedHash = parseAddresseeWithCached(r, addr, index)
	txData = new(privacy.Point).ScalarMultBase(r)
	return
}

func parseMoneyToCreateOutput(blind *privacy.Scalar, cachedHash *privacy.Scalar, money big.Int, index byte) (mask *privacy.Scalar, amount *privacy.Scalar, commitment *privacy.Point, err error) {
	scMoney, err := ParseBigIntToScalar(money)
	if err != nil {
		return nil, nil, nil, err
	}

	mask = privacy.HashToScalar(cachedHash.ToBytesS())
	amount = privacy.HashToScalar(mask.ToBytesS())

	mask = mask.Add(blind, mask)
	amount = amount.Add(scMoney, amount)
	commitment = ParseCommitment(blind, scMoney)

	return mask, amount, commitment, nil
}

func ParseBigIntToScalar(number big.Int) (*privacy.Scalar, error) {
	b := number.Bytes()
	if len(b) > 32 {
		return nil, errors.New("Error in tx_full ParseBigIntToScalar: BigInt too big (length larger than 32)")
	}
	zeroPadding := make([]byte, 32-len(b))
	b = append(zeroPadding, b...)

	// Reverse key of scalar
	scalar := new(privacy.Scalar).FromBytesS(b)
	keyReverse := privacy.Reverse(scalar.GetKey())
	result, err := scalar.SetKey(&keyReverse)
	if err != nil {
		return nil, errors.New("Error in txfull ParseBigIntToScalar: scalar.SetKet got error")
	}

	return result, nil
}

// Get Mask and Amount from UTXO if we have privateAddress
func ParseBlindAndMoneyFromUtxo(addr address.PrivateAddress, utxo UTXO) (blind *privacy.Scalar, money *privacy.Scalar, err error) {
	if IsUtxoOfAddress(addr, utxo) == false {
		return nil, nil, errors.New("Error in ota_interpreter ParseBlindAndMoneyFromUtxo: utxo is not from this address")
	}
	shared_secret := new(privacy.Point).ScalarMult(utxo.GetTxData(), addr.GetPrivateView())
	hashed_offset := privacy.HashToScalar(
		append(shared_secret.ToBytesS(), utxo.GetIndex()),
	)

	// Get blind value
	blind_offset := privacy.HashToScalar(
		hashed_offset.ToBytesS(),
	)
	blind = new(privacy.Scalar).Sub(utxo.GetMask(), blind_offset)

	// Get amount value
	money_offset := privacy.HashToScalar(
		blind_offset.ToBytesS(),
	)
	money = new(privacy.Scalar).Sub(utxo.GetAmount(), money_offset)

	return blind, money, nil
}

func ParseCommitment(blind *privacy.Scalar, money *privacy.Scalar) *privacy.Point {
	// Get blind*G
	blindHalf := new(privacy.Point).ScalarMultBase(blind)

	// Get value pedersen H in privacy library
	H := privacy.PedCom.G[privacy.PedersenValueIndex]
	moneyHalf := new(privacy.Point).ScalarMult(H, money)

	return new(privacy.Point).Add(blindHalf, moneyHalf)
}
