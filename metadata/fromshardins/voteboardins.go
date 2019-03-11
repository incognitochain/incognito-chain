package fromshardins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
)

type VoteBoardIns struct {
	BoardType               common.BoardType
	CandidatePaymentAddress privacy.PaymentAddress
	VoterPaymentAddress     privacy.PaymentAddress
	BoardIndex              uint32
	AmountOfVote            uint64
}

func NewVoteBoardIns(
	boardType common.BoardType,
	candidatePaymentAddress privacy.PaymentAddress,
	voterPaymentAddress privacy.PaymentAddress,
	boardIndex uint32,
	amountOfVote uint64,
) *VoteBoardIns {
	return &VoteBoardIns{BoardType: boardType, CandidatePaymentAddress: candidatePaymentAddress, VoterPaymentAddress: voterPaymentAddress, BoardIndex: boardIndex, AmountOfVote: amountOfVote}
}

func (VoteBoardIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func NewVoteDCBBoardInsFromStr(inst string) (*VoteBoardIns, error) {
	voteDCBBoardIns := &VoteBoardIns{}
	err := json.Unmarshal([]byte(inst), voteDCBBoardIns)
	voteDCBBoardIns.BoardType = common.DCBBoard
	if err != nil {
		return nil, err
	}
	return voteDCBBoardIns, nil
}

func NewVoteGOVBoardInsFromStr(inst string) (*VoteBoardIns, error) {
	voteGOVBoardIns := &VoteBoardIns{}
	err := json.Unmarshal([]byte(inst), voteGOVBoardIns)
	voteGOVBoardIns.BoardType = common.GOVBoard
	if err != nil {
		return nil, err
	}
	return voteGOVBoardIns, nil
}
