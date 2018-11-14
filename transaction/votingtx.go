package transaction

type TxVoteDCBProposal struct {
	Tx
	TxVoteDCBProposalData TxVoteDCBProposalData
}

type TxVoteGovProposal struct {
	Tx
	TxVoteGovProposalData TxVoteGovProposalData
}

type TxVoteGovProposalData struct {
	AmountVoteToken uint32
}

type TxVoteDCBProposalData struct{
	AmountVoteToken uint32
}
