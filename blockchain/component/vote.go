package component

import "github.com/big0t/constant-chain/common"

type VoteProposalData struct {
	ProposalTxID      common.Hash
	AmountOfVote      int32
	ConstitutionIndex uint32
}

func (voteProposalData VoteProposalData) ToBytes() []byte {
	b := voteProposalData.ProposalTxID.GetBytes()
	b = append(b, common.Int32ToBytes(voteProposalData.AmountOfVote)...)
	b = append(b, common.Uint32ToBytes(voteProposalData.ConstitutionIndex)...)
	return b
}
