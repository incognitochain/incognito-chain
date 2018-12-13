package metadata

import "github.com/ninjadotorg/constant/common"

type AcceptDCBProposalMetadata struct {
	DCBProposalTXID *common.Hash

	MetadataBase
}

func NewAcceptDCBProposalMetadata(voteDCBBoardMetadata map[string]interface{}) *AcceptDCBProposalMetadata {
	return &AcceptDCBProposalMetadata{
		DCBProposalTXID: voteDCBBoardMetadata["DCBProposalTXID"].(*common.Hash),
	}
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	_, _, _, tx, err := bcr.GetTransactionByHash(acceptDCBProposalMetadata.DCBProposalTXID)
	if err != nil {
		return false, err
	}
	if tx == nil {
		return false, nil
	}
	return true, nil
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) GetType() int {
	return AcceptDCBProposalMeta
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) Hash() *common.Hash {
	record := string(common.ToBytes(acceptDCBProposalMetadata.DCBProposalTXID))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ValidateMetadataByItself() bool {
	return true
}

type AcceptGOVProposalMetadata struct {
	GOVProposalTXID *common.Hash

	MetadataBase
}

func NewAcceptGOVProposalMetadata(voteGOVBoardMetadata map[string]interface{}) *AcceptGOVProposalMetadata {
	return &AcceptGOVProposalMetadata{
		GOVProposalTXID: voteGOVBoardMetadata["GOVProposalTXID"].(*common.Hash),
	}
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	_, _, _, tx, err := bcr.GetTransactionByHash(acceptGOVProposalMetadata.GOVProposalTXID)
	if err != nil {
		return false, err
	}
	if tx == nil {
		return false, nil
	}
	return true, nil
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) GetType() int {
	return AcceptGOVProposalMeta
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) Hash() *common.Hash {
	record := string(common.ToBytes(acceptGOVProposalMetadata.GOVProposalTXID))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateMetadataByItself() bool {
	return true
}
