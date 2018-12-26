package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/voting"
)

type AcceptDCBProposalMetadata struct {
	DCBProposalTXID common.Hash
	Voter           voting.Voter
	MetadataBase
}

func NewAcceptDCBProposalMetadata(DCBProposalTXID common.Hash, voter voting.Voter) *AcceptDCBProposalMetadata {
	return &AcceptDCBProposalMetadata{
		DCBProposalTXID: DCBProposalTXID,
		Voter:           voter,
		MetadataBase:    *NewMetadataBase(AcceptDCBProposalMeta),
	}
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	_, _, _, tx, err := bcr.GetTransactionByHash(&acceptDCBProposalMetadata.DCBProposalTXID)
	if err != nil {
		return false, err
	}
	if tx == nil {
		return false, nil
	}
	return true, nil
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) Hash() *common.Hash {
	record := string(acceptDCBProposalMetadata.DCBProposalTXID.GetBytes())
	record += string(acceptDCBProposalMetadata.Voter.Hash().GetBytes())
	record += string(acceptDCBProposalMetadata.MetadataBase.Hash().GetBytes())
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
	GOVProposalTXID common.Hash
	Voter           voting.Voter
	MetadataBase
}

func NewAcceptGOVProposalMetadata(GOVProposalTXID common.Hash, voter voting.Voter) *AcceptGOVProposalMetadata {
	return &AcceptGOVProposalMetadata{
		GOVProposalTXID: GOVProposalTXID,
		Voter:           voter,
		MetadataBase:    *NewMetadataBase(AcceptGOVProposalMeta),
	}
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	_, _, _, tx, err := bcr.GetTransactionByHash(&acceptGOVProposalMetadata.GOVProposalTXID)
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
	record := string(acceptGOVProposalMetadata.GOVProposalTXID.GetBytes())
	record += string(acceptGOVProposalMetadata.Hash().GetBytes())
	record += string(acceptGOVProposalMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateMetadataByItself() bool {
	return true
}
