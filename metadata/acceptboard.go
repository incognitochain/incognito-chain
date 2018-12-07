package metadata

import (
	"github.com/ninjadotorg/constant/common"
)

type AcceptDCBBoardMetadata struct {
	DCBBoardPubKeys     [][]byte
	StartAmountDCBToken uint64

	MetadataBase
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) GetType() int {
	return AcceptDCBProposalMeta
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) Hash() *common.Hash {
	record := ""
	for _, i := range acceptDCBBoardMetadata.DCBBoardPubKeys {
		record += string(i)
	}
	record += string(acceptDCBBoardMetadata.StartAmountDCBToken)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte) (bool, error) {
	return true, nil
}

func (acceptDCBBoardMetadata *AcceptDCBBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptDCBBoardMetadata.DCBBoardPubKeys) != bcr.GetNumberOfDCBGovernors() {
		return true, false, nil
	}
	for _, i := range acceptDCBBoardMetadata.DCBBoardPubKeys {
		if len(i) != common.HashSize {
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

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) GetType() int {
	return AcceptGOVProposalMeta
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) Hash() *common.Hash {
	record := ""
	for _, i := range acceptGOVBoardMetadata.GOVBoardPubKeys {
		record += string(i)
	}
	record += string(acceptGOVBoardMetadata.StartAmountGOVToken)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateTxWithBlockChain(Transaction, BlockchainRetriever, byte) (bool, error) {
	return true, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	if len(acceptGOVBoardMetadata.GOVBoardPubKeys) != bcr.GetNumberOfGOVGovernors() {
		return true, false, nil
	}
	for _, i := range acceptGOVBoardMetadata.GOVBoardPubKeys {
		if len(i) != common.HashSize {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (acceptGOVBoardMetadata *AcceptGOVBoardMetadata) ValidateMetadataByItself() bool {
	return true
}
