package fromshardins

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy"
)

type VoteBoardIns struct {
	BoardType               common.BoardType
	CandidatePaymentAddress privacy.PaymentAddress
	BoardIndex              uint32
}

func (VoteBoardIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func NewVoteBoardIns(boardType common.BoardType, candidatePaymentAddress privacy.PaymentAddress, boardIndex uint32) *VoteBoardIns {
	return &VoteBoardIns{BoardType: boardType, CandidatePaymentAddress: candidatePaymentAddress, BoardIndex: boardIndex}
}
