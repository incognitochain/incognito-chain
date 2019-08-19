package bridgesig

import (
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

func Sign(hexkey string, data []byte) (string, error) {
	sk, err := ethcrypto.HexToECDSA(hexkey)
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
