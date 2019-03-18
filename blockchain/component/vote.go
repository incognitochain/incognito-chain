package component

import "github.com/constant-money/constant-chain/common"

type VoteProposalData struct {
	ProposalTxID      common.Hash
	ConstitutionIndex uint32
}

func (voteProposalData VoteProposalData) ToBytes() []byte {
	b := voteProposalData.ProposalTxID.GetBytes()
	b = append(b, common.Uint32ToBytes(voteProposalData.ConstitutionIndex)...)
	return b
}

// func NewVoteProposalDataFromByte(b []byte) VoteProposalData {
// 	return nil
// }
