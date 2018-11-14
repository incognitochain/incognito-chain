package blockchain

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
)

type ConstitutionInfo struct {
	StartedBlockHeight int32
	ExecuteDuration int32
}

type GovConstitution struct{
	ConstitutionInfo
	CurrentGovNationalWelfare int
	GovParameters GovParameters
}

type DCBConstitution struct{
	ConstitutionInfo
	CurrentDCBNationalWelfare int
	DCBParameters DCBParameters
}

type DCBConstitutionHelper struct{}
type GovConstitutionHelper struct{}
func (DCBConstitutionHelper) GetStartedBlockHeight(blockgen *BlkTmplGenerator, chainID byte) (int32){
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastDCBConstitution := BestBlock.DCBConstitution
	return lastDCBConstitution.StartedBlockHeight
}

func (DCBConstitutionHelper) CheckSubmitProposalType(tx transaction.Transaction) (bool) {
	return tx.GetType() == common.TxSubmitDCBProposal
}

func (DCBConstitutionHelper) CheckVotingProposalType(tx transaction.Transaction) (bool){
	return tx.GetType() == common.TxVotingDCBProposal
}

func (DCBConstitutionHelper) GetAmountVoteToken(tx transaction.Transaction) (uint32) {
	return tx.(*transaction.TxVoteDCBProposal).TxVoteDCBProposalData.AmountVoteToken
}

func (GovConstitutionHelper) GetStartedBlockHeight(blockgen *BlkTmplGenerator, chainID byte) (int32){
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastGovConstitution := BestBlock.GovConstitution
	return lastGovConstitution.StartedBlockHeight
}

func (GovConstitutionHelper) CheckSubmitProposalType(tx transaction.Transaction) (bool) {
	return tx.GetType() == common.TxSubmitGovProposal
}

func (GovConstitutionHelper) CheckVotingProposalType(tx transaction.Transaction) (bool){
	return tx.GetType() == common.TxVotingGovProposal
}

func (GovConstitutionHelper) GetAmountVoteToken(tx transaction.Transaction) (uint32) {
	return tx.(*transaction.TxVoteGovProposal).TxVoteGovProposalData.AmountVoteToken
}
