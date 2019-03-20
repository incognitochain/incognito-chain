package component

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
)

type VoteProposalData struct {
	ProposalTxID      common.Hash
	ConstitutionIndex uint32
	VoterPayment      privacy.PaymentAddress
}

func (voteProposalData VoteProposalData) ToBytes() []byte {
	b := voteProposalData.ProposalTxID.GetBytes()
	b = append(b, common.Uint32ToBytes(voteProposalData.ConstitutionIndex)...)
	b = append(b, voteProposalData.VoterPayment.Bytes()...)
	return b
}
