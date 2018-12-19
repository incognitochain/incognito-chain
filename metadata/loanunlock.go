package metadata

import (
	"encoding/hex"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type LoanUnlock struct {
	LoanID []byte
	MetadataBase
}

func NewLoanUnlock(data map[string]interface{}) *LoanUnlock {
	result := LoanUnlock{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	result.Type = LoanUnlockMeta
	return &result
}

func (lu *LoanUnlock) Hash() *common.Hash {
	record := string(lu.LoanID)

	// final hash
	record += string(lu.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lu *LoanUnlock) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// TODO(@0xbunyip): validate that there's a corresponding TxLoanWithdraw in the same block
	return true, nil
}

func (lu *LoanUnlock) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, true, nil // continue checking for fee
}

func (lu *LoanUnlock) ValidateMetadataByItself() bool {
	return true
}
