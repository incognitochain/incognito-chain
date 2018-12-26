package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type AcceptDCBBoardMetadata struct {
	DCBBoardPubKeys     [][]byte
	StartAmountDCBToken uint64

	MetadataBase
}

func NewAcceptDCBBoardMetadata(DCBBoardPubKeys [][]byte, startAmountDCBToken uint64) *AcceptDCBBoardMetadata {
	return &AcceptDCBBoardMetadata{
		DCBBoardPubKeys:     DCBBoardPubKeys,
		StartAmountDCBToken: startAmountDCBToken,
		MetadataBase:        *NewMetadataBase(AcceptDCBBoardMeta),
	}
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) Hash() *common.Hash {
	record := ""
	for _, i := range acceptDCBBoardMetadata.DCBBoardPubKeys {
		record += string(i)
	}
	record += string(acceptDCBBoardMetadata.StartAmountDCBToken)
	record += string(acceptDCBBoardMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptDCBBoardMetadata.DCBBoardPubKeys) != bcr.GetNumberOfDCBGovernors() {
		return true, false, nil
	}
	for _, i := range acceptDCBBoardMetadata.DCBBoardPubKeys {
		if len(i) != common.PubKeyLength {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateMetadataByItself() bool {
	return true
}

type AcceptGOVBoardMetadata struct {
	GOVBoardPubKeys     [][]byte
	StartAmountGOVToken uint64

	MetadataBase
}

func NewAcceptGOVBoardMetadata(GOVBoardPubKeys [][]byte, startAmountGOVToken uint64) *AcceptGOVBoardMetadata {
	return &AcceptGOVBoardMetadata{
		GOVBoardPubKeys:     GOVBoardPubKeys,
		StartAmountGOVToken: startAmountGOVToken,
		MetadataBase:        *NewMetadataBase(AcceptGOVBoardMeta),
	}
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) Hash() *common.Hash {
	record := ""
	for _, i := range acceptGOVBoardMetadata.GOVBoardPubKeys {
		record += string(i)
	}
	record += string(acceptGOVBoardMetadata.StartAmountGOVToken)
	record += string(acceptGOVBoardMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte, database.DatabaseInterface) (bool, error) {
	return true, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptGOVBoardMetadata.GOVBoardPubKeys) != bcr.GetNumberOfGOVGovernors() {
		return true, false, nil
	}
	for _, i := range acceptGOVBoardMetadata.GOVBoardPubKeys {
		if len(i) != common.PubKeyLength {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}
