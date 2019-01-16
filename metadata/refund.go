package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type Refund struct {
	MetadataBase
	SmallTxID common.Hash
}

func NewRefund(smallTxID common.Hash, metaType int) *Refund {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &Refund{
		SmallTxID:    smallTxID,
		MetadataBase: metadataBase,
	}
}

func (rf *Refund) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return common.TrueValue
}

func (rf *Refund) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via SmallTxID) in current block
	return common.FalseValue, nil
}

func (rf *Refund) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return common.FalseValue, common.TrueValue, nil
}

func (rf *Refund) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning common.TrueValue here
	return common.TrueValue
}

func (rf *Refund) Hash() *common.Hash {
	record := rf.SmallTxID.String()
	record += string(rf.MetadataBase.Hash()[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
