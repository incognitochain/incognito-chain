package bls

import (
	"crypto/sha256"
	"errors"
	bn256 "github.com/go-ethereum-master/crypto/bn256/cloudflare"
	"math/big"
)

func HashToG1(m []byte) *bn256.G1 {
	one := big.NewInt(1)

	h := sha256.Sum256(m)
	x := new(big.Int).SetBytes(h[:])
	x.Mod(x, bn256.P)

	for {
		xxx := new(big.Int).Mul(x, x)
		xxx.Mul(xxx, x)
		t := new(big.Int).Add(xxx, curveB)

		y := new(big.Int).ModSqrt(t, bn256.P)
		if y != nil {
			return SetXYG1(x, y)
		}
		x.Add(x, one)
	}
}

func HashToScalar(m []byte) *big.Int {
	h := sha256.Sum256(m)
	x := new(big.Int).SetBytes(h[:])
	return x.Mod(x, bn256.Order)
}

func SetXYG1(x, y *big.Int) *bn256.G1 {
	xBytes := x.Bytes()
	yBytes := y.Bytes()
	pBytes := make([]byte, CBigIntSize*2)
	copy(pBytes[1*CBigIntSize-len(xBytes):], xBytes)
	copy(pBytes[2*CBigIntSize-len(yBytes):], yBytes)
	res := new(bn256.G1)
	res.Unmarshal(pBytes)
	return res
}

func CompressG1(p *bn256.G1) []byte {
	pBytes := p.Marshal()
	xBytes := pBytes[:CBigIntSize]
	y := new(big.Int).SetBytes(pBytes[CBigIntSize:])
	yPrime := new(big.Int).Sub(bn256.P, y)
	if y.Cmp(yPrime) > 0 {
		xBytes[0] |= CMaskByte
	}
	return xBytes
}

func DecompressG1(data []byte) (*bn256.G1, error) {
	if len(data) != CCompressSize {
		return nil, errors.New(CInputError)
	}

	tmp := make([]byte, CCompressSize)
	copy(tmp[:], data[0:CCompressSize])

	flag := tmp[0] >> 7
	tmp[0] &= CNotMaskByte

	x := new(big.Int)
	x.SetBytes(tmp)

	xxx := new(big.Int).Mul(x, x)
	xxx.Mul(xxx, x)
	t := new(big.Int).Add(xxx, curveB)
	y := new(big.Int).ModSqrt(t, bn256.P)
	if y == nil {
		return nil, errors.New(CCompressError)
	}
	yPrime := new(big.Int).Sub(bn256.P, y)
	smaller := y.Cmp(yPrime) < 0

	if flag == 1 && smaller {
		return SetXYG1(x, yPrime), nil
	}
	if flag == 0 && !smaller {
		return SetXYG1(x, yPrime), nil
	}
	return SetXYG1(x, y), nil
}

func PairingCheck(a []*bn256.G1, b []*bn256.G2) bool {
	return bn256.PairingCheck(a, b)
}

func MultiScalarMultG1(pointLs []*bn256.G1, scalarLs []*big.Int) *bn256.G1 {
	Zero := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	if len(scalarLs) != len(pointLs) {
		return nil
	}

	digitsLs := make([][64]int8, len(scalarLs))
	for i := range digitsLs {
		digitsLs[i] = signedRadix16(scalarLs[i])
	}

	PiLs := make([][9]*bn256.G1, len(scalarLs))
	for i := 0; i < len(scalarLs); i++ {
		// 0P, 1P, 2P, ..., 8P
		PiLs[i][0] = new(bn256.G1).Set(Zero)
		PiLs[i][1] = new(bn256.G1).Set(pointLs[i])
		for j := 2; j < 9; j++ {
			PiLs[i][j] = new(bn256.G1).Add(pointLs[i], PiLs[i][j-1])
		}
	}

	res := new(bn256.G1).Set(Zero)
	for i := 0; i < len(scalarLs); i++ {
		xmask := digitsLs[i][63] >> 7
		xabs := uint8((digitsLs[i][63] + xmask) ^ xmask)
		for j := 0; j < 9; j++ {
			if int(xabs) == j {
				if digitsLs[i][63] >= 0 {
					res.Add(res, PiLs[i][j])
				} else {
					res.Add(res, new(bn256.G1).Neg(PiLs[i][j]))
				}
				break
			}
		}
	}

	tmp := new(bn256.G1)
	for k := 62; k >= 0; k-- {
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)

		for i := 0; i < len(scalarLs); i++ {
			xmask := digitsLs[i][k] >> 7
			xabs := uint8((digitsLs[i][k] + xmask) ^ xmask)
			for j := 0; j < 16; j++ {
				if int(xabs) == j {
					if digitsLs[i][k] >= 0 {
						res.Add(res, PiLs[i][j])
					} else {
						res.Add(res, new(bn256.G1).Neg(PiLs[i][j]))
					}
					break
				}
			}
		}
	}

	return res
}

func MultiScalarMultG2(pointLs []*bn256.G2, scalarLs []*big.Int) *bn256.G2 {
	Zero := new(bn256.G2).ScalarBaseMult(big.NewInt(0))
	if len(scalarLs) != len(pointLs) {
		return nil
	}

	digitsLs := make([][64]int8, len(scalarLs))
	for i := range digitsLs {
		digitsLs[i] = signedRadix16(scalarLs[i])
	}

	PiLs := make([][9]*bn256.G2, len(scalarLs))
	for i := 0; i < len(scalarLs); i++ {
		// 0P, 1P, 2P, ..., 8P
		PiLs[i][0] = new(bn256.G2).Set(Zero)
		PiLs[i][1] = new(bn256.G2).Set(pointLs[i])
		for j := 2; j < 9; j++ {
			PiLs[i][j] = new(bn256.G2).Add(pointLs[i], PiLs[i][j-1])
		}
	}

	res := new(bn256.G2).Set(Zero)
	for i := 0; i < len(scalarLs); i++ {
		xmask := digitsLs[i][63] >> 7
		xabs := uint8((digitsLs[i][63] + xmask) ^ xmask)
		for j := 0; j < 9; j++ {
			if int(xabs) == j {
				if digitsLs[i][63] >= 0 {
					res.Add(res, PiLs[i][j])
				} else {
					res.Add(res, new(bn256.G2).Neg(PiLs[i][j]))
				}
				break
			}
		}
	}

	tmp := new(bn256.G2)
	for k := 62; k >= 0; k-- {
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)

		for i := 0; i < len(scalarLs); i++ {
			xmask := digitsLs[i][k] >> 7
			xabs := uint8((digitsLs[i][k] + xmask) ^ xmask)
			for j := 0; j < 16; j++ {
				if int(xabs) == j {
					if digitsLs[i][k] >= 0 {
						res.Add(res, PiLs[i][j])
					} else {
						res.Add(res, new(bn256.G2).Neg(PiLs[i][j]))
					}
					break
				}
			}
		}
	}

	return res
}

func ScalarMultRadix16(point *bn256.G2, scalar *big.Int) *bn256.G2 {
	digits := signedRadix16(scalar)
	var Pi [9]*bn256.G2
	Pi[0] = new(bn256.G2).ScalarMult(point, big.NewInt(0))
	Pi[1] = new(bn256.G2).Set(point)
	for j := 2; j < 9; j++ {
		Pi[j] = new(bn256.G2).Add(point, Pi[j-1])
	}

	res := new(bn256.G2)
	for j := 0; j < 9; j++ {
		xmask := digits[63] >> 7
		xabs := uint8((digits[63] + xmask) ^ xmask)
		if int(xabs) == j {
			if digits[63] >= 0 {
				res.Set(Pi[j])
			} else {
				res.Set(new(bn256.G2).Neg(Pi[j]))
			}
			break
		}
	}
	tmp := new(bn256.G2)
	for k := 62; k >= 0; k-- {
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)
		tmp.Add(res, res)
		res.Set(tmp)

		xmask := digits[k] >> 7
		xabs := uint8((digits[k] + xmask) ^ xmask)
		for j := 0; j < 9; j++ {
			if int(xabs) == j {
				if digits[k] >= 0 {
					res.Add(res, Pi[j])
				} else {
					res.Add(res, new(bn256.G2).Neg(Pi[j]))
				}
				break
			}
		}
	}
	return res
}

func signedRadix16(digit *big.Int) [64]int8 {
	// Compute unsigned radix-16 digits:
	s := digit.Bytes()
	var digits [64]int8

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	for i := 0; i < 32; i++ {
		if i < len(s) {
			digits[2*i] = int8(s[i] & 15)
			digits[2*i+1] = int8((s[i] >> 4) & 15)
			continue
		}
		digits[2*i] = 0
		digits[2*i+1] = 0
	}

	// Recenter coefficients:
	for i := 0; i < 63; i++ {
		carry := (digits[i] + 8) >> 4
		digits[i] -= carry << 4
		digits[i+1] += carry
	}

	return digits
}
