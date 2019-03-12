package fromshardins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
	"strconv"
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

func (voteBoardIns VoteBoardIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(voteBoardIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.VoteBoardIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewVoteBoardInsFromStr(inst string) (*VoteBoardIns, error) {
	voteBoardIns := &VoteBoardIns{}
	err := json.Unmarshal([]byte(inst), voteBoardIns)
	if err != nil {
		return nil, err
	}
	return voteBoardIns, nil
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
