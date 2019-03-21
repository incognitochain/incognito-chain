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

type SubmitProposalData struct {
	ProposalTxID      common.Hash
	ConstitutionIndex uint32
	SubmitterPayment  privacy.PaymentAddress
}

func (submitProposalData SubmitProposalData) ToBytes() []byte {
	res := make([]byte, common.HashSize+4+common.PaymentAddressLength)
	i := 0
	i += copy(res[i:], submitProposalData.ProposalTxID.GetBytes())
	i += copy(res[i:], common.Uint32ToBytes(submitProposalData.ConstitutionIndex))
	i += copy(res[i:], submitProposalData.SubmitterPayment.Bytes())
	return res
}
