package bridgesig

import (
	"reflect"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func Sign(keyBytes []byte, data []byte) ([]byte, error) {
	sk, err := ethcrypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, err
	}
	hash := ethcrypto.Keccak256Hash(data)
	sig, err := ethcrypto.Sign(hash.Bytes(), sk)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func Verify(pubkeyBytes []byte, data []byte, sig []byte) (bool, error) {
	hash := ethcrypto.Keccak256Hash(data)
	pk, err := ethcrypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return false, err
	}
	if !reflect.DeepEqual(pubkeyBytes, ethcrypto.CompressPubkey(pk)) {
		return false, nil
	}
	return true, nil
}
