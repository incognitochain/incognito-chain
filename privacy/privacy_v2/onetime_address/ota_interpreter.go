package onetime_address

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/onetime_address/address"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/onetime_address/utxo"
)

func ParseUtxoPrivatekey(addr *address.PrivateAddress, utxo *utxo.Utxo) *operation.Scalar {
	rK := new(operation.Point).ScalarMult(utxo.GetTxData(), addr.GetPrivateView())
	hashed := operation.HashToScalar(
		append(rK.ToBytesS(), utxo.GetIndex()),
	)
	return new(operation.Scalar).Add(hashed, addr.GetPrivateSpend())
}

// Step 1 Monero One-time Address
func parseAddresseeWithCached(r *operation.Scalar, addr *address.PublicAddress, index byte) (*operation.Point, *operation.Scalar) {
	rK := new(operation.Point).ScalarMult(addr.GetPublicView(), r)
	cachedHash := operation.HashToScalar(
		append(rK.ToBytesS(), index),
	)
	HrKG := new(operation.Point).ScalarMultBase(cachedHash)
	addressee := new(operation.Point).Add(HrKG, addr.GetPublicSpend())

	return addressee, cachedHash
}

// Step 2 Monero One-time Address
func parseOTAWithCached(r *operation.Scalar, addr *address.PublicAddress, index byte) (*operation.Point, *operation.Point, *operation.Scalar) {
	addressee, cachedHash := parseAddresseeWithCached(r, addr, index)
	txData := new(operation.Point).ScalarMultBase(r)
	return addressee, txData, cachedHash
}

func parseMoneyToCreateOutput(blind *operation.Scalar, cachedHash *operation.Scalar, money *big.Int, index byte) (mask *operation.Scalar, amount *operation.Scalar, commitment *operation.Point, err error) {
	scMoney := privacy_util.BigIntToScalar(money)
	if scMoney == nil {
		return nil, nil, nil, errors.New("OTA error parseMoneyToCreateInput: Cannot parse BigInt To Scalar")
	}

	mask = operation.HashToScalar(cachedHash.ToBytesS())
	amount = operation.HashToScalar(mask.ToBytesS())

	mask = mask.Add(blind, mask)
	amount = amount.Add(scMoney, amount)
	commitment = ParseCommitment(blind, scMoney)

	return mask, amount, commitment, nil
}

// Get Mask and Amount from UTXO if we have privateAddress
func ParseBlindAndMoneyFromUtxo(addr *address.PrivateAddress, utxo *utxo.Utxo) (blind *operation.Scalar, money *operation.Scalar, err error) {
	if IsUtxoOfAddress(addr, utxo) == false {
		return nil, nil, errors.New("Error in ota_interpreter ParseBlindAndMoneyFromUtxo: utxo is not from this address")
	}
	shared_secret := new(operation.Point).ScalarMult(utxo.GetTxData(), addr.GetPrivateView())
	hashed_offset := operation.HashToScalar(
		append(shared_secret.ToBytesS(), utxo.GetIndex()),
	)

	// Get blind value
	blind_offset := operation.HashToScalar(
		hashed_offset.ToBytesS(),
	)
	blind = new(operation.Scalar).Sub(utxo.GetMask(), blind_offset)

	// Get amount value
	money_offset := operation.HashToScalar(
		blind_offset.ToBytesS(),
	)
	money = new(operation.Scalar).Sub(utxo.GetAmount(), money_offset)

	return blind, money, nil
}

func ParseCommitment(blind *operation.Scalar, money *operation.Scalar) *operation.Point {
	// Get blind*G
	blindHalf := new(operation.Point).ScalarMultBase(blind)

	// Get value pedersen H in privacy library
	H := operation.PedCom.G[operation.PedersenValueIndex]
	moneyHalf := new(operation.Point).ScalarMult(H, money)

	return new(operation.Point).Add(blindHalf, moneyHalf)
}
