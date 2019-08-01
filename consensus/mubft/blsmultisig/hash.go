package blsmultisig

import (
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"golang.org/x/crypto/sha3"
)

// P2I is Point to Big Int, in BLS-BFT, it called H1
func P2I(point *bn256.G1) *big.Int {
	pnByte := CmprG1(point)
	res := big.NewInt(0)
	res.SetBytes(pnByte)
	for res.Cmp(bn256.Order) != -1 {
		pnByte = Hash4Bls(pnByte)
		res.SetBytes(pnByte)
	}
	return res
}

func B2I(bytes []byte) *big.Int {
	res := big.NewInt(0)
	res.SetBytes(bytes)
	for res.Cmp(bn256.Order) != -1 {
		bytes = Hash4Bls(bytes)
		res.SetBytes(bytes)
	}
	return res
}

// I2P is Big Int to Point, in BLS-BFT, it called H0
func I2P(bigInt *big.Int) *bn256.G1 {
	x := big.NewInt(0)
	x.Set(bigInt)
	for i := 0; ; i++ {
		res, err := xCoor2G1P(x, bigInt.Bit(0) == 1)
		if err == nil {
			return res
		}
		x.SetBytes(Hash4Bls(x.Bytes()))
	}
}

// Hash4Bls is Hash function for calculate block hash
// this is different from hash function for calculate transaction hash
func Hash4Bls(data []byte) []byte {
	hashMachine := sha3.NewLegacyKeccak256()
	hashMachine.Write(data)
	return hashMachine.Sum(nil)
}
