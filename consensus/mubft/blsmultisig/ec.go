package blsmultisig

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

// CmprG1 take a point in G1 group and return bytes array
func CmprG1(pn *bn256.G1) []byte {
	pnBytesArr := pn.Marshal()
	xCoorBytes := pnBytesArr[:CBigIntSz]
	fmt.Println(xCoorBytes)
	if pnBytesArr[CBigIntSz*2-1]&1 == 1 {
		xCoorBytes[0] |= 0x80
	}
	return xCoorBytes
}

// DecmprG1 is
func DecmprG1(bytes []byte) *bn256.G1 {
	// xCoorByte := bytes[:]
	seed := time.Now().UnixNano()
	reader := rand.New(rand.NewSource(int64(seed)))
	_, x, _ := bn256.RandomG1(reader)
	return x
}

func xCoor2G1P(xCoor *big.Int, oddPoint bool) *bn256.G1 {
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
	pn.Unmarshal(pnBytesArr)
	if ((yCoorByte[CBigIntSz-1]&1 == 0) && oddPoint) || ((yCoorByte[CBigIntSz-1]&1 == 1) && !oddPoint) {
		pn = pn.Neg(pn)
	}
	return pn
}
