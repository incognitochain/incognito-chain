package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type AcceptDCBBoardMetadata struct {
	DCBBoardPaymentAddress []privacy.PaymentAddress
	StartAmountDCBToken    uint64

	MetadataBase
}

func NewAcceptDCBBoardMetadata(DCBBoardPaymentAddress []privacy.PaymentAddress, startAmountDCBToken uint64) *AcceptDCBBoardMetadata {
	return &AcceptDCBBoardMetadata{
		DCBBoardPaymentAddress: DCBBoardPaymentAddress,
		StartAmountDCBToken:    startAmountDCBToken,
		MetadataBase:           *NewMetadataBase(AcceptDCBBoardMeta),
	}
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) Hash() *common.Hash {
	record := ""
	for _, i := range acceptDCBBoardMetadata.DCBBoardPaymentAddress {
		record += i.String()
	}
	record += string(acceptDCBBoardMetadata.StartAmountDCBToken)
	record += acceptDCBBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptDCBBoardMetadata.DCBBoardPaymentAddress) != bcr.GetNumberOfDCBGovernors() {
		return true, false, nil
	}
	return true, true, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

type AcceptGOVBoardMetadata struct {
	GOVBoardPaymentAddress []privacy.PaymentAddress
	StartAmountGOVToken    uint64

	MetadataBase
}

func NewAcceptGOVBoardMetadata(GOVBoardPaymentAddress []privacy.PaymentAddress, startAmountGOVToken uint64) *AcceptGOVBoardMetadata {
	return &AcceptGOVBoardMetadata{
		GOVBoardPaymentAddress: GOVBoardPaymentAddress,
		StartAmountGOVToken:    startAmountGOVToken,
		MetadataBase:           *NewMetadataBase(AcceptGOVBoardMeta),
	}
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) Hash() *common.Hash {
	record := ""
	for _, i := range acceptGOVBoardMetadata.GOVBoardPaymentAddress {
		record += i.String()
	}
	record += string(acceptGOVBoardMetadata.StartAmountGOVToken)
	record += acceptGOVBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptGOVBoardMetadata.GOVBoardPaymentAddress) != bcr.GetNumberOfGOVGovernors() {
		return true, false, nil
	}
	return true, true, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) CalculateSize() uint64 {
	return calculateSize(acceptGOVBoardMetadata)
}
