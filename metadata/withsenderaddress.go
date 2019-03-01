package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type WithSenderAddress struct {
	MetadataBase
	SenderAddress privacy.PaymentAddress
}

func NewWithSenderAddress(senderAddress privacy.PaymentAddress, metaType int) *WithSenderAddress {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &WithSenderAddress{
		SenderAddress: senderAddress,
		MetadataBase:  metadataBase,
	}
}

func (wsa *WithSenderAddress) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via RequestedTxID) in current block
	return false, nil
}

func (wsa *WithSenderAddress) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(wsa.SenderAddress.Pk) == 0 {
		return false, false, nil
	}
	if len(wsa.SenderAddress.Tk) == 0 {
		return false, false, nil
	}
	return true, true, nil
}

func (wsa *WithSenderAddress) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (wsa *WithSenderAddress) Hash() *common.Hash {
	record := wsa.SenderAddress.String()
	// final hash
	record += wsa.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (wsa *WithSenderAddress) CalculateSize() uint64 {
	return calculateSize(wsa)
}
