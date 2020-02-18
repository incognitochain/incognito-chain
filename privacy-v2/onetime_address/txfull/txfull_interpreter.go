package txfull

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy-v2/onetime_address/address"
)

// Step 1 Monero One-time Address
func getAddresseeWithCached(r *privacy.Scalar, addr *address.PublicAddress, index byte) (*privacy.Point, *privacy.Scalar) {
	rK := new(privacy.Point).ScalarMult(addr.GetPublicView(), r)
	cachedHash := privacy.HashToScalar(
		append(rK.ToBytesS(), index),
	)
	HrKG := new(privacy.Point).ScalarMultBase(cachedHash)
	addressee := new(privacy.Point).Add(HrKG, addr.GetPublicSpend())

	return addressee, cachedHash
}

// Step 2 Monero One-time Address
func ParseOTAWithCached(r *privacy.Scalar, addr *address.PublicAddress, index byte) (addressee *privacy.Point, txData *privacy.Point, cachedHash *privacy.Scalar) {
	addressee, cachedHash = getAddresseeWithCached(r, addr, index)
	txData = new(privacy.Point).ScalarMultBase(r)
	return
}

func ParseMoneyToCreateOutput(cachedHash *privacy.Scalar, money *big.Int, index byte) (mask *privacy.Scalar, amount *privacy.Scalar, err error) {
	blind := privacy.RandomScalar()
	scMoney, err := ParseBigIntToScalar(money)
	if err != nil {
		return nil, nil, err
	}

	mask = privacy.HashToScalar(scMoney.ToBytesS())
	amount = privacy.HashToScalar(mask.ToBytesS())

	mask = mask.Add(blind, mask)
	amount = amount.Add(scMoney, amount)

	return mask, amount, nil
}

func ParseBigIntToScalar(number *big.Int) (*privacy.Scalar, error) {
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

// func ParseMaskAmount(mask *privacy.Scalar, amount *privacy.Scalar) (blind *privacy.Scalar, money *privacy.Scalar) {

// }

func ParseCommitment(blind *privacy.Scalar, money *privacy.Scalar) *privacy.Point {
	// Get blind*G
	blindHalf := new(privacy.Point).ScalarMultBase(blind)

	// Get value pedersen H in privacy library
	H := privacy.PedCom.G[privacy.PedersenValueIndex]
	moneyHalf := new(privacy.Point).ScalarMult(H, money)

	return new(privacy.Point).Add(blindHalf, moneyHalf)
}
