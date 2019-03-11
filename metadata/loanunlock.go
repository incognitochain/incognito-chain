package metadata

import (
	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/database"
)

type LoanUnlock struct {
	LoanID []byte
	MetadataBase
}

func (lu *LoanUnlock) Hash() *common.Hash {
	record := string(lu.LoanID)

	// final hash
	record += lu.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lu *LoanUnlock) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// TODO(@0xbunyip): validate that there's a corresponding TxLoanWithdraw in the same block
	return true, nil
}

func (lu *LoanUnlock) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, true, nil // continue checking for fee
}

func (lu *LoanUnlock) ValidateMetadataByItself() bool {
	return true
}

func (lu *LoanUnlock) CalculateSize() uint64 {
	return calculateSize(lu)
}
