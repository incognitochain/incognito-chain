package btcrelaying

import (
	"github.com/incognitochain/incognito-chain/common/base58"
	"golang.org/x/crypto/sha3"
)

// ConvertExternalBTCAmountToIncAmount converts amount in bTC chain (decimal 8) to amount in inc chain (decimal 9)
func ConvertExternalBTCAmountToIncAmount(externalBTCAmount int64) int64 {
	return externalBTCAmount * 10 // externalBTCAmount / 1^8 * 1^9
}

// ConvertIncPBTCAmountToExternalBTCAmount converts amount in inc chain (decimal 9) to amount in bTC chain (decimal 8)
func ConvertIncPBTCAmountToExternalBTCAmount(incPBTCAmount int64) int64 {
	return incPBTCAmount / 10 // incPBTCAmount / 1^9 * 1^8
}

func HashAndEncodeBase58(msg string) string {
	hash := make([]byte, 16)
	h := sha3.NewShake128()
	h.Write([]byte(msg))
	h.Read(hash)
	b58 := new(base58.Base58)
	return b58.Encode(hash)
}
