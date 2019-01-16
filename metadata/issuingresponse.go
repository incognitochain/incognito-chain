package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type IssuingResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

func NewIssuingResponse(requestedTxID common.Hash, metaType int) *IssuingResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes *IssuingResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return common.TrueValue
}

func (iRes *IssuingResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via RequestedTxID) in current block
	return common.FalseValue, nil
}

func (iRes *IssuingResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return common.FalseValue, common.TrueValue, nil
}

func (iRes *IssuingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning common.TrueValue here
	return common.TrueValue
}

func (iRes *IssuingResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += string(iRes.MetadataBase.Hash()[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
