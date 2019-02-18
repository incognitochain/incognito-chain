package metadata

import (
	"github.com/ninjadotorg/constant/privacy"
)

type VoteBoardMetadata struct {
	CandidatePaymentAddress privacy.PaymentAddress
}

func NewVoteBoardMetadata(candidatePaymentAddress privacy.PaymentAddress) *VoteBoardMetadata {
	return &VoteBoardMetadata{
		CandidatePaymentAddress: candidatePaymentAddress,
	}
}

func (voteBoardMetadata *VoteBoardMetadata) GetBytes() []byte {
	return voteBoardMetadata.CandidatePaymentAddress.Bytes()
}
