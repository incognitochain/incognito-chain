package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy"
)

type VoteBoardMetadata struct {
	CandidatePaymentAddress privacy.PaymentAddress
	Amount                  int64
}

func NewVoteBoardMetadata(candidatePaymentAddress privacy.PaymentAddress, amount int64) *VoteBoardMetadata {
	return &VoteBoardMetadata{
		CandidatePaymentAddress: candidatePaymentAddress,
		Amount:                  amount,
	}
}

func (voteBoardMetadata *VoteBoardMetadata) GetBytes() []byte {
	return append(voteBoardMetadata.CandidatePaymentAddress.Bytes(), common.Int64ToBytes(voteBoardMetadata.Amount)...)
}
