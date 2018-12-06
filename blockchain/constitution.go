package blockchain

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/transaction"
)

type ConstitutionInfo struct {
	StartedBlockHeight int32
	ExecuteDuration    int32
	ProposalTXID       *common.Hash
}

type GOVConstitution struct {
	ConstitutionInfo
	CurrentGOVNationalWelfare int32
	GOVParams                 GOVParams
}

type DCBConstitution struct {
	ConstitutionInfo
	CurrentDCBNationalWelfare int32
	DCBParams                 params.DCBParams
}

type DCBConstitutionHelper struct{}
type GOVConstitutionHelper struct{}

func (DCBConstitutionHelper) GetStartedBlockHeight(blockgen *BlkTmplGenerator, chainID byte) int32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastDCBConstitution := BestBlock.Header.DCBConstitution
	return lastDCBConstitution.StartedBlockHeight
}

func (DCBConstitutionHelper) CheckSubmitProposalType(tx metadata.Transaction) bool {
	return tx.GetType() == common.TxSubmitDCBProposal
}

func (DCBConstitutionHelper) CheckVotingProposalType(tx metadata.Transaction) bool {
	return tx.GetType() == common.TxVoteDCBProposal
}

func (DCBConstitutionHelper) GetAmountVoteToken(tx metadata.Transaction) uint32 {
	return tx.(*transaction.TxVoteDCBProposal).VoteDCBProposalData.AmountVoteToken
}

func (GOVConstitutionHelper) GetStartedBlockHeight(blockgen *BlkTmplGenerator, chainID byte) int32 {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastGOVConstitution := BestBlock.Header.GOVConstitution
	return lastGOVConstitution.StartedBlockHeight
}

func (GOVConstitutionHelper) CheckSubmitProposalType(tx metadata.Transaction) bool {
	return tx.GetType() == common.TxSubmitGOVProposal
}

func (GOVConstitutionHelper) CheckVotingProposalType(tx metadata.Transaction) bool {
	return tx.GetType() == common.TxVoteGOVProposal
}

func (GOVConstitutionHelper) GetAmountVoteToken(tx metadata.Transaction) uint32 {
	return tx.(*transaction.TxVoteGOVProposal).VoteGOVProposalData.AmountVoteToken
}

func (DCBConstitutionHelper) TxAcceptProposal(originTx metadata.Transaction) metadata.Transaction {
	SubmitTx := originTx.(*transaction.TxSubmitDCBProposal)
	AcceptTx := transaction.TxAcceptDCBProposal{
		DCBProposalTXID: SubmitTx.GetTxID(),
	}
	AcceptTx.Type = common.TxAcceptDCBProposal
	return AcceptTx
}

func (GOVConstitutionHelper) TxAcceptProposal(originTx metadata.Transaction) metadata.Transaction {
	SubmitTx := originTx.(*transaction.TxSubmitGOVProposal)
	AcceptTx := transaction.TxAcceptGOVProposal{
		GOVProposalTXID: SubmitTx.GetTxID(),
	}
	AcceptTx.Type = common.TxAcceptGOVProposal
	return AcceptTx
}
