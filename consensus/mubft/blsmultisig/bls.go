package blsmultisig

import (
	"errors"
	"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
)

// Sign return BLS signature
func Sign(data []byte, sk *big.Int, selfIdx int) ([]byte, error) {
	if selfIdx >= len(CommonPKs) {
		return []byte{0}, errors.New(CErr + CErrInps)
	}
	dataPn := B2G1P(data)
	aiSk := big.NewInt(0)
	aiSk.Set(CommonAis[selfIdx])
	aiSk.Mul(aiSk, sk)
	aiSk.Mod(aiSk, bn256.Order)
	sig := dataPn.ScalarMult(dataPn, aiSk)
	return I2Bytes(G1P2I(sig), CCmprPnSz), nil
}

// Verify verify BLS sig on given data and list public key
func Verify(sig, data []byte, signersIdx []int) (bool, error) {
	gG2Pn := new(bn256.G2)
	gG2Pn.ScalarBaseMult(big.NewInt(1))
	sigPn := B2G1P(sig)
	lPair := bn256.Pair(sigPn, gG2Pn)
	apk := CalcAPK(signersIdx)
	dataPn := B2G1P(data)
	rPair := bn256.Pair(dataPn, apk)
	fmt.Println(lPair.Marshal())
	fmt.Println(rPair.Marshal())
	return true, nil
}

// Combine combine list of bls signature
func Combine(sigs [][]byte) ([]byte, error) {
	return []byte{0}, nil
}
