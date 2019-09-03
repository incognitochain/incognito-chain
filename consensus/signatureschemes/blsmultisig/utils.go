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

// func Bytes2I(bytes []byte) *big.Int{
// 	res := bn.Bytes()
// 	for ; len(res) < length; res = append([]byte{0}, res...) {
// 	}
// 	return res
// }

// CacheCommonPKs convert list publickey in current epoch from byte-type to point-type
/*func CacheCommonPKs(ListPK []PublicKey) error {
	commonPKs := make([]*bn256.G2, len(ListPK))
	commonAPs := make([]*bn256.G2, len(ListPK))
	commonAis := make([]*big.Int, len(ListPK))
	var err error
	for i, pk := range ListPK {
		commonPKs[i], err = DecmprG2(pk)
		if err != nil {
			return err
		}
	}

	combinedPKByte := []byte{}
	for i := 0; i < len(ListPK); i++ {
		combinedPKByte = append(combinedPKByte, ListPK[i]...)
	}
	for i := 0; i < len(commonPKs); i++ {
		commonAPs[i], commonAis[i] = AKGen(ListPK[i], combinedPKByte)
	}
	CommonPKs = commonPKs
	CommonAPs = commonAPs
	CommonAis = commonAis
	return nil
}*/

// CalcAPK calculate apk = pki^ai * ... * pkj^aj
func CalcAPK(signerIdx []int) *bn256.G2 {
	apk := new(bn256.G2)
	apk.ScalarMult(CommonAPs[signerIdx[0]], big.NewInt(1))
	for i := 1; i < len(signerIdx); i++ {
		apk.Add(apk, CommonAPs[signerIdx[i]])
	}
	return apk
}

// func GetFunctionName(i interface{}) string {
// 	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
// }
