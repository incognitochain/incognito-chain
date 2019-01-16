package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
)

type Voter struct {
	PaymentAddress privacy.PaymentAddress
	AmountOfVote   int32
}

func (voter *Voter) Greater(voter2 Voter) bool {
	return voter.AmountOfVote > voter2.AmountOfVote ||
		(voter.AmountOfVote == voter2.AmountOfVote && string(voter.PaymentAddress.Bytes()) > string(voter2.PaymentAddress.Bytes()))
}

func (voter *Voter) Hash() *common.Hash {
	record := string(voter.PaymentAddress.Bytes())
	record += string(voter.AmountOfVote)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

type ProposalVote struct {
	TxId         common.Hash
	AmountOfVote int64
	NumberOfVote uint32
}

func (proposalVote ProposalVote) Greater(proposalVote2 ProposalVote) bool {
	return proposalVote.AmountOfVote > proposalVote2.AmountOfVote ||
		(proposalVote.AmountOfVote == proposalVote2.AmountOfVote || proposalVote.NumberOfVote > proposalVote2.NumberOfVote) ||
		(proposalVote.AmountOfVote == proposalVote2.AmountOfVote || proposalVote.NumberOfVote == proposalVote2.NumberOfVote || string(proposalVote.TxId.GetBytes()) > string(proposalVote2.TxId.GetBytes()))
}

type AcceptDCBProposalMetadata struct {
	DCBProposalTXID common.Hash
	Voter           Voter
	MetadataBase
}

func NewAcceptDCBProposalMetadata(DCBProposalTXID common.Hash, voter Voter) *AcceptDCBProposalMetadata {
	return &AcceptDCBProposalMetadata{
		DCBProposalTXID: DCBProposalTXID,
		Voter:           voter,
		MetadataBase:    *NewMetadataBase(AcceptDCBProposalMeta),
	}
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	_, _, _, tx, err := bcr.GetTransactionByHash(&acceptDCBProposalMetadata.DCBProposalTXID)
	if err != nil {
		return common.FalseValue, err
	}
	if tx == nil {
		return common.FalseValue, nil
	}
	return common.TrueValue, nil
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) Hash() *common.Hash {
	record := string(acceptDCBProposalMetadata.DCBProposalTXID.GetBytes())
	record += string(acceptDCBProposalMetadata.Voter.Hash().GetBytes())
	record += string(acceptDCBProposalMetadata.MetadataBase.Hash().GetBytes())
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ValidateSanityData(bcr BlockchainRetriever, tx Transaction) (bool, bool, error) {
	return common.TrueValue, common.TrueValue, nil
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ValidateMetadataByItself() bool {
	return common.TrueValue
}

type AcceptGOVProposalMetadata struct {
	GOVProposalTXID common.Hash
	Voter           Voter
	MetadataBase
}

func NewAcceptGOVProposalMetadata(GOVProposalTXID common.Hash, voter Voter) *AcceptGOVProposalMetadata {
	return &AcceptGOVProposalMetadata{
		GOVProposalTXID: GOVProposalTXID,
		Voter:           voter,
		MetadataBase:    *NewMetadataBase(AcceptGOVProposalMeta),
	}
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	_, _, _, tx, err := bcr.GetTransactionByHash(&acceptGOVProposalMetadata.GOVProposalTXID)
	if err != nil {
		return common.FalseValue, err
	}
	if tx == nil {
		return common.FalseValue, nil
	}
	return common.TrueValue, nil
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
	return common.TrueValue, common.TrueValue, nil
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateMetadataByItself() bool {
	return common.TrueValue
}
