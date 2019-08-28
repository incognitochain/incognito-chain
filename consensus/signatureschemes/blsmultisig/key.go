package blsmultisig

import (
	"math/big"
	"sync"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/google"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

// KeyGen take an input seed and return BLS Key
func KeyGen(seed []byte) (*big.Int, *bn256.G2) {
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
func PKGen(sk *big.Int) *bn256.G2 {
	pk := new(bn256.G2)
	pk = pk.ScalarBaseMult(sk)
	return pk
}

// AKGen take a seed and return BLS secret key
func AKGen(listPKBytes []PublicKey, id int) (*bn256.G2, *big.Int) {
	akByte := []byte{}
	akByte = append(akByte, listPKBytes[id]...)
	for i := 0; i < len(listPKBytes); i++ {
		akByte = append(akByte, listPKBytes[i]...)
	}
	akByte = Hash4Bls(akByte)
	akBInt := B2I(akByte)
	res := new(bn256.G2)
	PKPn, _ := DecmprG2(listPKBytes[id])
	res = res.ScalarMult(PKPn, akBInt)
	return res, akBInt
}

// ListAKGen take a seed and return BLS secret key
// func APKGen(committee []PublicKey, idx []int) *bn256.G2 {
// 	// apk := new(bn256.G2)
// 	apk, _ := AKGen(committee, idx[0])
// 	// apk.ScalarMult(CommonAPs[signerIdx[0]], big.NewInt(1))
// 	wg := sync.WaitGroup{}
// 	apkTmpList := make([]*bn256.G2, len(idx)-1)
// 	for i := 1; i < len(idx); i++ {
// 		wg.Add(1)
// 		go func(index int) {
// 			apkTmp, _ := AKGen(committee, idx[index])
// 			apkTmpList[index-1] = apkTmp
// 			wg.Done()
// 		}(i)
// 	}
// 	wg.Wait()
// 	for _, apkTmp := range apkTmpList {
// 		apk.Add(apk, apkTmp)
// 	}

// 	return apk
// }

func APKGen(committee []PublicKey, idx []int) *bn256.G2 {
	wg := sync.WaitGroup{}
	apkTmpList := make([]*bn256.G2, len(idx))
	for i := 0; i < len(idx); i++ {
		wg.Add(1)
		go func(index int) {
			apkTmpList[index], _ = AKGen(committee, idx[index])
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i := 1; i < len(idx); i++ {
		apkTmpList[0].Add(apkTmpList[0], apkTmpList[i])
	}
	return apkTmpList[0]
}

func AiGen(listPKBytes []PublicKey, id int) *big.Int {
	akByte := []byte{}
	akByte = append(akByte, listPKBytes[id]...)
	for i := 0; i < len(listPKBytes); i++ {
		akByte = append(akByte, listPKBytes[i]...)
	}
	akByte = Hash4Bls(akByte)
	akBInt := B2I(akByte)
	return akBInt
}

// SKBytes take input secretkey integer and return secretkey bytes
func SKBytes(sk *big.Int) SecretKey {
	return I2Bytes(sk, CSKSz)
}

// PKBytes take input publickey point and return publickey bytes
func PKBytes(pk *bn256.G2) PublicKey {
	return CmprG2(pk)
}

// ChkPKSt Check input string is BLS PublicKey string-type
func ChkPKSt(pkSt string) bool {
	pkBytes, ver, err := base58.Base58Check{}.Decode(pkSt)
	if err != nil {
		return false
	}
	pkPn := new(bn256.G2)
	if _, err := pkPn.Unmarshal(pkBytes); err != nil {
		return false
	}
	if ver != common.ZeroByte {
		return false
	}
	return true
}

func IncSK2BLSPKBytes(sk []byte) []byte {
	_, pk := KeyGen(sk)
	return CmprG2(pk)
}

func ListPKBytes2ListPKPoints(listPKBytes []PublicKey) ([]*bn256.G2, error) {
	listPKs := make([]*bn256.G2, len(listPKBytes))
	var err error
	for i, pk := range listPKBytes {
		// fmt.Println(pk, len(pk))
		listPKs[i], err = DecmprG2(pk)
		if err != nil {
			return nil, err
		}
	}
	return listPKs, nil
}
