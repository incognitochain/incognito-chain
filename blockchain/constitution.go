package blockchain

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/voting"
)

type ConstitutionInfo struct {
	StartedBlockHeight int32
	ExecuteDuration int32
	CurrentGovNationalWelfare int
	ProposalTXID *common.Hash
}

type GovConstitution struct{
	ConstitutionInfo
	GovParams voting.GovParams
}

type DCBConstitution struct{
	ConstitutionInfo
	DCBParameters voting.DCBParams
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
	return tx.(*transaction.TxVoteDCBProposal).VoteDCBProposalData.AmountVoteToken
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
	return tx.(*transaction.TxVoteGovProposal).VoteGovProposalData.AmountVoteToken
}

//xxx
func (DCBConstitutionHelper) TxAcceptProposal(originTx transaction.Transaction) (transaction.TxAcceptDCBProposal){
	SubmitTx := originTx.(*transaction.TxSubmitDCBProposal)
	AcceptTx := transaction.TxAcceptDCBProposal{

	}
	//tx := originTx.(tran{
	//	originTx
	//}
}
