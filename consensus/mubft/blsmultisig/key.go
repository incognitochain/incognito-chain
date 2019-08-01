package blsmultisig

import (
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/incognitochain/incognito-chain/common"
)

// KeyGen take an input seed and return BLS Key
func KeyGen(seed []byte) (*big.Int, *bn256.G1) {
	sk := SKGen(seed)
	return sk, PKGen(sk)
}

// SKGen take a seed and return BLS secret key
func SKGen(seed []byte) *big.Int {
	sk := big.NewInt(0)
	sk.SetBytes(common.HashB(seed))
	for {
		if sk.Cmp(bn256.Order) == -1 {
			break
		}
		sk.SetBytes(Hash4Bls(sk.Bytes()))
	}
	return sk
}

// PKGen take a secret key and return BLS public key
func PKGen(sk *big.Int) *bn256.G1 {
	pk := new(bn256.G1)
	pk = pk.ScalarBaseMult(sk)
	return pk
}

// AKGen take a seed and return BLS secret key
func AKGen(listPKPn []*bn256.G1, id int) *bn256.G1 {
	akByte := CmprG1(listPKPn[id])
	for i := 0; i < len(listPKPn); i++ {
		akByte = Hash4Bls(append(akByte, CmprG1(listPKPn[i])...))
	}
	akBInt := B2I(akByte)
	res := new(bn256.G1)
	res = res.ScalarMult(listPKPn[id], akBInt)
	return res
}

// SKBytes take input secretkey integer and return secretkey bytes
func SKBytes(sk *big.Int) SecretKey {
	return I2Bytes(sk, CSKSz)
}

// PKBytes take input publickey point and return publickey bytes
func PKBytes(pk *bn256.G1) PublicKey {
	return CmprG1(pk)
}
