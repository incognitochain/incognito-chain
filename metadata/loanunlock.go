package metadata

import (
	"encoding/hex"

	"github.com/ninjadotorg/constant/common"
)

type LoanUnlock struct {
	LoanID []byte
	MetadataBase
}

func NewLoanUnlock(data map[string]interface{}) *LoanUnlock {
	result := LoanUnlock{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	return &result
}

func (lu *LoanUnlock) GetType() int {
	return LoanUnlockMeta
}

func (lu *LoanUnlock) Hash() *common.Hash {
	record := string(lu.LoanID)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lu *LoanUnlock) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	// TODO(@0xbunyip): validate that there's a corresponding TxLoanWithdraw in the same block
	return true, nil
}

func (lu *LoanUnlock) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, true, nil // continue checking for fee
}

func (lu *LoanUnlock) ValidateMetadataByItself() bool {
	return true
}
