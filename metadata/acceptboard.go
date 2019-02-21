package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type AcceptBoardMetadata struct {
	BoardPaymentAddress []privacy.PaymentAddress
	StartAmountToken    uint64
}

func (acceptBoardMetadata AcceptBoardMetadata) ToBytes() []byte {
	record := ""
	for _, i := range acceptBoardMetadata.BoardPaymentAddress {
		record += i.String()
	}
	record += string(acceptBoardMetadata.StartAmountToken)
	return []byte(record)
}

func NewAcceptBoardMetadata(boardPaymentAddress []privacy.PaymentAddress, startAmountToken uint64) *AcceptBoardMetadata {
	return &AcceptBoardMetadata{BoardPaymentAddress: boardPaymentAddress, StartAmountToken: startAmountToken}
}

type AcceptDCBBoardMetadata struct {
	AcceptBoardMetadata AcceptBoardMetadata
	MetadataBase
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	err := bcr.UpdateDCBBoard(tx)
	if err != nil {
		return err
	}
	return nil
}

func NewAcceptDCBBoardMetadata(DCBBoardPaymentAddress []privacy.PaymentAddress, startAmountDCBToken uint64) *AcceptDCBBoardMetadata {
	return &AcceptDCBBoardMetadata{
		AcceptBoardMetadata: *NewAcceptBoardMetadata(
			DCBBoardPaymentAddress,
			startAmountDCBToken,
		),
		MetadataBase: *NewMetadataBase(AcceptDCBBoardMeta),
	}
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) Hash() *common.Hash {
	record := string(acceptDCBBoardMetadata.AcceptBoardMetadata.ToBytes())
	record += acceptDCBBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptDCBBoardMetadata.AcceptBoardMetadata.BoardPaymentAddress) != bcr.GetNumberOfDCBGovernors() {
		return true, false, nil
	}
	return true, true, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

type AcceptGOVBoardMetadata struct {
	AcceptBoardMetadata AcceptBoardMetadata

	MetadataBase
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	err := bcr.UpdateGOVBoard(tx)
	if err != nil {
		return err
	}
	return nil
}

func NewAcceptGOVBoardMetadata(GOVBoardPaymentAddress []privacy.PaymentAddress, startAmountGOVToken uint64) *AcceptGOVBoardMetadata {
	return &AcceptGOVBoardMetadata{
		AcceptBoardMetadata: *NewAcceptBoardMetadata(
			GOVBoardPaymentAddress,
			startAmountGOVToken,
		),
		MetadataBase: *NewMetadataBase(AcceptGOVBoardMeta),
	}
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) Hash() *common.Hash {
	record := string(acceptGOVBoardMetadata.AcceptBoardMetadata.ToBytes())
	record += acceptGOVBoardMetadata.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptGOVBoardMetadata.AcceptBoardMetadata.BoardPaymentAddress) != bcr.GetNumberOfGOVGovernors() {
		return true, false, nil
	}
	return true, true, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}
