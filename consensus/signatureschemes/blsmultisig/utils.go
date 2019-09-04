package blsmultisig

import (
	"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

func printBit(bn *big.Int) {
	for i := 255; i >= 0; i-- {
		fmt.Printf("%+v", bn.Bit(i))
	}
	fmt.Println("")
}

// I2Bytes take an integer and return bytes arrays of it with fixed length
func I2Bytes(bn *big.Int, length int) []byte {
	res := bn.Bytes()
	for ; len(res) < length; res = append([]byte{0}, res...) {
	}
	return res
}

// CalcAPK calculate apk = pki^ai * ... * pkj^aj
func CalcAPK(signerIdx []int) *bn256.G2 {
	apk := new(bn256.G2)
	apk.ScalarMult(CommonAPs[signerIdx[0]], big.NewInt(1))
	for i := 1; i < len(signerIdx); i++ {
		apk.Add(apk, CommonAPs[signerIdx[i]])
	}
	return apk
}
