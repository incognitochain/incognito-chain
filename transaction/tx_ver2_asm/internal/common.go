package internal

import (
	"crypto/rand"
	"math/big"
	// "encoding/hex"
	// "errors"

	// "golang.org/x/crypto/sha3"
	"github.com/incognitochain/incognito-chain/common"
)

const(
	HashSize = 32
	MaxShardNumber = 1
	TxRandomGroupSize = 36
	MaxHashStringSize = HashSize * 2
)
const (
	TxNormalType          = "n"   // normal tx(send and receive coin)
	TxRewardType          = "s"   // reward tx
	TxReturnStakingType   = "rs"  //
	TxConversionType      = "cv"  // Convert 1 - 2 normal tx
	TxTokenConversionType = "tcv" // Convert 1 - 2 token tx
	//TxCustomTokenType        = "t"  // token  tx with no supporting privacy
	TxCustomTokenPrivacyType = "tp" // token  tx with supporting privacy
)
var(
	PRVCoinID = common.Hash{4}
)


func RandBigIntMaxRange(max *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, max)
}

func GetShardIDFromLastByte(b byte) byte {
	return byte(int(b) % MaxShardNumber)
}

func grabBytes(coinBytes *[]byte, offset *int) ([]byte, error) {
	b := *coinBytes
	if *offset >= len(b) {
		return nil, genericError
	}
	lenField := b[*offset]
	*offset += 1
	result := make([]byte, lenField)
	if lenField != 0 {
		if *offset+int(lenField) > len(b) {
			return nil, genericError
		}
		data := b[*offset : *offset+int(lenField)]
		copy(result, data)
		*offset += int(lenField)
	}
	return result, nil
}