package blsmultisig

import (
	"errors"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

// this file implement some helpful function for work with complex number
// For reduce 128bytes of G2 Point to 64 bytes

// GfP2 - Gf(p^2) is an extension field of the finite field GF(p)= Z/pZ
// A + Bi, with A, B in Zp; i^2 = -1
type GfP2 struct {
	A, B *big.Int
}

// SetBytes return GfP2 element from bytes
func (gfP2 *GfP2) SetBytes(bytes []byte) error {
	if len(bytes) != CBigIntSz*2 {
		return errors.New(CErr + CErrInps)
	}
	aInt := big.NewInt(0)
	bInt := big.NewInt(0)
	aInt.SetBytes(bytes[:CBigIntSz])
	bInt.SetBytes(bytes[CBigIntSz:])
	if aInt.Cmp(bn256.P) != -1 {
		aInt.Mod(aInt, bn256.P)
	}
	if bInt.Cmp(bn256.P) != -1 {
		bInt.Mod(bInt, bn256.P)
	}
	gfP2.A = aInt
	gfP2.B = bInt
	return nil
}

// Bytes return
func (gfP2 *GfP2) Bytes() []byte {
	// res := make([]byte, CBigIntSz*2)
	res := I2Bytes(gfP2.A, CBigIntSz)
	res = append(res, I2Bytes(gfP2.B, CBigIntSz)...)
	return res
}

// Add Complex number addition
func (gfP2 *GfP2) Add(x *GfP2, y *GfP2) *GfP2 {
	res := new(GfP2)
	res.A = big.NewInt(0)
	res.B = big.NewInt(0)
	res.A.Add(x.A, y.A)
	res.B.Add(x.B, y.B)
	if res.A.Cmp(bn256.P) != -1 {
		res.A.Mod(res.A, bn256.P)
	}
	if res.B.Cmp(bn256.P) != -1 {
		res.B.Mod(res.B, bn256.P)
	}
	return res
}

// Mul Complex number multiplication
func (gfP2 *GfP2) Mul(x *GfP2, y *GfP2) *GfP2 {
	return nil
}

// Sqr Complex number power 2
func (gfP2 *GfP2) Sqr(x *GfP2) *GfP2 {
	return nil
}
