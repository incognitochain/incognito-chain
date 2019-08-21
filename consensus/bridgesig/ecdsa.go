package bridgesig

import (
	"reflect"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

func Sign(keyBytes []byte, data []byte) (string, error) {
	sk, err := ethcrypto.ToECDSA(keyBytes)
	if err != nil {
		return "", err
	}
	hash := ethcrypto.Keccak256Hash(data)
	sig, err := ethcrypto.Sign(hash.Bytes(), sk)
	if err != nil {
		return "", err
	}
	sigStr := base58.Base58Check{}.Encode(sig, common.ZeroByte)
	return sigStr, nil
}

func Verify(pubkeyBytes []byte, data []byte, sig []byte) (bool, error) {
	pk, err := ethcrypto.SigToPub(data, sig)
	if err != nil {
		return false, err
	}
	if !reflect.DeepEqual(pubkeyBytes, ethcrypto.CompressPubkey(pk)) {
		return false, nil
	}
	return true, nil
}
