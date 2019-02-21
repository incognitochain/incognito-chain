package metadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/blockchain/params"
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
		(voter.AmountOfVote == voter2.AmountOfVote && bytes.Compare(voter.PaymentAddress.Bytes(), voter2.PaymentAddress.Bytes()) > 0)
}

func (voter *Voter) ToBytes() []byte {
	record := voter.PaymentAddress.String()
	record += string(voter.AmountOfVote)
	return []byte(record)
}

type ProposalVote struct {
	TxId         common.Hash
	AmountOfVote int64
	NumberOfVote uint32
}

func (proposalVote ProposalVote) Greater(proposalVote2 ProposalVote) bool {
	return proposalVote.AmountOfVote > proposalVote2.AmountOfVote ||
		(proposalVote.AmountOfVote == proposalVote2.AmountOfVote && proposalVote.NumberOfVote > proposalVote2.NumberOfVote) ||
		(proposalVote.AmountOfVote == proposalVote2.AmountOfVote && proposalVote.NumberOfVote == proposalVote2.NumberOfVote && string(proposalVote.TxId.GetBytes()) > string(proposalVote2.TxId.GetBytes()))
}

type AcceptProposalMetadata struct {
	ProposalTXID common.Hash
	Voter        Voter
}

func (acceptProposalMetadata *AcceptProposalMetadata) ToBytes() []byte {
	record := string(acceptProposalMetadata.ProposalTXID.GetBytes())
	record += string(acceptProposalMetadata.Voter.ToBytes())
	return []byte(record)
}

func NewAcceptProposalMetadata(DCBProposalTXID common.Hash, voter Voter) *AcceptProposalMetadata {
	return &AcceptProposalMetadata{ProposalTXID: DCBProposalTXID, Voter: voter}
}

type AcceptDCBProposalMetadata struct {
	AcceptProposalMetadata AcceptProposalMetadata
	MetadataBase
}

type ConstitutionInterface interface {
	GetConstitutionIndex() uint32
}

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := common.DCBBoard
	consitution := bcr.GetConstitution(boardType)
	nextConstitutionIndex := consitution.GetConstitutionIndex() + 1
	err := bcr.UpdateConstitution(tx, boardType)
	if err != nil {
		return err
	}
	err1 := bcr.GetDatabase().TakeVoteTokenFromWinner(boardType, nextConstitutionIndex, acceptDCBProposalMetadata.AcceptProposalMetadata.Voter.PaymentAddress, acceptDCBProposalMetadata.AcceptProposalMetadata.Voter.AmountOfVote)
	if err1 != nil {
		return err1
	}
	err2 := bcr.GetDatabase().SetNewProposalWinningVoter(boardType, nextConstitutionIndex, acceptDCBProposalMetadata.AcceptProposalMetadata.Voter.PaymentAddress)
	if err2 != nil {
		return err2
	}
	return nil
}

func NewAcceptDCBProposalMetadata(DCBProposalTXID common.Hash, voter Voter) *AcceptDCBProposalMetadata {
	return &AcceptDCBProposalMetadata{
		AcceptProposalMetadata: *NewAcceptProposalMetadata(
			DCBProposalTXID,
			voter,
		),
		MetadataBase: *NewMetadataBase(AcceptDCBProposalMeta),
	}
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

func (acceptDCBProposalMetadata *AcceptDCBProposalMetadata) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionValue, err := getSaleDataActionValue(acceptDCBProposalMetadata, bcr)
	if err != nil {
		return nil, err
	}
	action := []string{strconv.Itoa(AcceptDCBProposalMeta), actionValue}
	return [][]string{action}, nil
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

func ParseAcceptDCBProposalMetadataActionValue(values string) (*params.DCBParams, error) {
	params := &params.DCBParams{}
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

func (acceptGOVProposalMetadata *AcceptGOVProposalMetadata) ProcessWhenInsertBlockShard(tx Transaction, bcr BlockchainRetriever) error {
	boardType := common.GOVBoard
	consitution := bcr.GetConstitution(boardType)
	nextConstitutionIndex := consitution.GetConstitutionIndex() + 1
	err := bcr.UpdateConstitution(tx, boardType)
	if err != nil {
		return err
	}
	underlieMetadata := tx.GetMetadata().(*AcceptGOVProposalMetadata)
	err1 := bcr.GetDatabase().TakeVoteTokenFromWinner(boardType, nextConstitutionIndex, underlieMetadata.AcceptProposalMetadata.Voter.PaymentAddress, underlieMetadata.AcceptProposalMetadata.Voter.AmountOfVote)
	if err1 != nil {
		return err1
	}
	err2 := bcr.GetDatabase().SetNewProposalWinningVoter(boardType, nextConstitutionIndex, underlieMetadata.AcceptProposalMetadata.Voter.PaymentAddress)
	if err2 != nil {
		return err2
	}
	return nil

}

func NewAcceptGOVProposalMetadata(GOVProposalTXID common.Hash, voter Voter) *AcceptGOVProposalMetadata {
	return &AcceptGOVProposalMetadata{
		AcceptProposalMetadata: *NewAcceptProposalMetadata(
			GOVProposalTXID,
			voter,
		),
		MetadataBase: *NewMetadataBase(AcceptGOVProposalMeta),
	}
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
