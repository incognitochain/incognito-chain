package metadata

import (
	"encoding/json"
	"errors"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
)

type ProposalVote struct {
	TxId common.Hash
	// AmountOfVote int64
	NumberOfVote uint32
}

func (proposalVote ProposalVote) Greater(proposalVote2 ProposalVote) bool {
	return (proposalVote.NumberOfVote > proposalVote2.NumberOfVote) ||
		(proposalVote.NumberOfVote == proposalVote2.NumberOfVote && string(proposalVote.TxId.GetBytes()) > string(proposalVote2.TxId.GetBytes()))
}

type AcceptProposalMetadata struct {
	ProposalTXID common.Hash
	Voter        component.Voter
}

func (acceptProposalMetadata *AcceptProposalMetadata) ToBytes() []byte {
	record := string(acceptProposalMetadata.ProposalTXID.GetBytes())
	record += string(acceptProposalMetadata.Voter.ToBytes())
	return []byte(record)
}

func NewAcceptProposalMetadata(DCBProposalTXID common.Hash, voter component.Voter) *AcceptProposalMetadata {
	return &AcceptProposalMetadata{ProposalTXID: DCBProposalTXID, Voter: voter}
}

type AcceptDCBProposalMetadata struct {
	AcceptProposalMetadata AcceptProposalMetadata
	MetadataBase
}

type ConstitutionInterface interface {
	GetConstitutionIndex() uint32
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	_, _, _, tx, err := bcr.GetTransactionByHash(&acceptDCBProposalMetadata.AcceptProposalMetadata.ProposalTXID)
	if err != nil {
		return false, err
	}
	if tx == nil {
		return false, nil
	}
	return true, nil
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) Hash() *common.Hash {
	record := string(acceptDCBProposalMetadata.AcceptProposalMetadata.ToBytes())
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

func getSaleDataActionValue(meta *AcceptDCBProposalMetadata, bcr BlockchainRetriever) (string, error) {
	_, _, _, txProposal, err := bcr.GetTransactionByHash(&meta.AcceptProposalMetadata.ProposalTXID)
	if err != nil {
		return "", err
	}
	metaProposal, ok := txProposal.GetMetadata().(*SubmitDCBProposalMetadata)
	if !ok {
		return "", errors.New("Error parsing proposal metadata")
	}
	value, err := json.Marshal(metaProposal.DCBParams)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func ParseAcceptDCBProposalMetadataActionValue(values string) (*component.DCBParams, error) {
	params := &component.DCBParams{}
	err := json.Unmarshal([]byte(values), params)
	if err != nil {
		return nil, err
	}
	return params, nil
}

type AcceptGOVProposalMetadata struct {
	AcceptProposalMetadata AcceptProposalMetadata
	MetadataBase
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	_, _, _, tx, err := bcr.GetTransactionByHash(&acceptGOVProposalMetadata.AcceptProposalMetadata.ProposalTXID)
	if err != nil {
		return false, err
	}
	if tx == nil {
		return false, nil
	}
	return true, nil
}

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) Hash() *common.Hash {
	record := string(acceptGOVProposalMetadata.AcceptProposalMetadata.ToBytes())
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

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) CalculateSize() uint64 {
	return calculateSize(acceptGOVProposalMetadata)
}
