package metadata

import (
	"github.com/ninjadotorg/constant/common"
)

type BuyBackResponse struct {
	MetadataBase
	RequestedTxID *common.Hash
}

func (bbRes *BuyBackResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (bbRes *BuyBackResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via RequestedTxID) in current block
	return false, nil
}

func (bbRes *BuyBackResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (bbRes *BuyBackResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (bbRes *BuyBackResponse) Hash() *common.Hash {
	record := bbRes.RequestedTxID.String()
	record += string(bbRes.MetadataBase.Hash()[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
