package blsmultisig

import (
	"math/big"

	"golang.org/x/crypto/bn256"
	"golang.org/x/crypto/sha3"
)

// P2I is Point to Big Int, in BLS-BFT, it called H1
func P2I(point *bn256.G1) *big.Int {
	return big.NewInt(int64(0))
}

// I2P is Big Int to Point, in BLS-BFT, it called H0
func I2P(bigInt *big.Int) *bn256.G1 {
	return nil
}

// Hash4Block is Hash function for calculate block hash
// this is different from hash function for calculate transaction hash
func Hash4Block(data []byte) []byte {
	hashMachine := sha3.NewLegacyKeccak256()
	hashMachine.Write(data)
	return hashMachine.Sum(nil)
}
