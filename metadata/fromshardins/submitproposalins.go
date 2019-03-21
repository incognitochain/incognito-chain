package fromshardins

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"

	"github.com/constant-money/constant-chain/common"
)

type SubmitProposalIns struct {
	BoardType      common.BoardType
	SubmitProposal component.SubmitProposalData
}

func (submitProposalIns SubmitProposalIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(submitProposalIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.VoteProposalIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewSubmitProposalIns(boardType common.BoardType, submitProposal component.SubmitProposalData) *SubmitProposalIns {
	return &SubmitProposalIns{BoardType: boardType, SubmitProposal: submitProposal}
}

func NewSubmitProposalInsFromStr(inst string) (*SubmitProposalIns, error) {
	Ins := &SubmitProposalIns{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, err
	}
	return Ins, nil
}
