package blsmultisig

import (
	"errors"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

// CmprG1 take a point in G1 group and return bytes array
func CmprG1(pn *bn256.G1) []byte {
	pnBytesArr := pn.Marshal()
	xCoorBytes := pnBytesArr[:CBigIntSz]
	if pnBytesArr[CBigIntSz*2-1]&1 == 1 {
		xCoorBytes[0] |= CMaskByte
	}
	return xCoorBytes
}

// DecmprG1 is
func DecmprG1(bytes []byte) (*bn256.G1, error) {
	if len(bytes) != CCmprPnSz {
		return nil, errors.New(GetFunctionName(DecmprG1) + CErr + CErrInLn)
	}

	oddPoint := ((bytes[0] & CMaskByte) != 0x00)
	if oddPoint {
		bytes[0] &= CNotMaskB
	}
	xCoor := big.NewInt(1)
	xCoor.SetBytes(bytes)
	pn, err := xCoor2G1P(xCoor, oddPoint)
	if err != nil {
		return nil, errors.New(GetFunctionName(DecmprG1) + CErr + err.Error())
	}
	return pn, nil
}

func xCoor2G1P(xCoor *big.Int, oddPoint bool) (*bn256.G1, error) {
	pnBytesArr := I2Bytes(xCoor, CBigIntSz)
	xCoorPow3 := big.NewInt(1)
	xCoorPow3.Exp(xCoor, big.NewInt(3), bn256.P)
	yCoorPow2 := big.NewInt(3)
	yCoorPow2.Add(xCoorPow3, yCoorPow2)
	yCoorPow2.Mod(yCoorPow2, bn256.P)

	yCoor := big.NewInt(0)
	yCoor.Exp(yCoorPow2, pAdd1Div4, bn256.P)
	pn := new(bn256.G1)
	yCoorByte := I2Bytes(yCoor, CBigIntSz)
	pnBytesArr = append(pnBytesArr, yCoorByte...)
	_, err := pn.Unmarshal(pnBytesArr)
	if err != nil {
		return nil, err
	}
	if ((yCoorByte[CBigIntSz-1]&1 == 0) && oddPoint) || ((yCoorByte[CBigIntSz-1]&1 == 1) && !oddPoint) {
		pn = pn.Neg(pn)
	}
	return pn, nil
}
