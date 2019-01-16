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
	record := common.EmptyString
	for _, i := range acceptDCBBoardMetadata.DCBBoardPaymentAddress {
		record += string(i.Bytes())
	}
	record += string(acceptDCBBoardMetadata.StartAmountDCBToken)
	record += string(acceptDCBBoardMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return common.TrueValue, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptDCBBoardMetadata.DCBBoardPaymentAddress) != bcr.GetNumberOfDCBGovernors() {
		return common.TrueValue, common.FalseValue, nil
	}
	return common.TrueValue, common.TrueValue, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateMetadataByItself() bool {
	return common.TrueValue
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
	record := common.EmptyString
	for _, i := range acceptGOVBoardMetadata.GOVBoardPaymentAddress {
		record += string(i.Bytes())
	}
	record += string(acceptGOVBoardMetadata.StartAmountGOVToken)
	record += string(acceptGOVBoardMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return common.TrueValue, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptGOVBoardMetadata.GOVBoardPaymentAddress) != bcr.GetNumberOfGOVGovernors() {
		return common.TrueValue, common.FalseValue, nil
	}
	return common.TrueValue, common.TrueValue, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateMetadataByItself() bool {
	return common.TrueValue
}
