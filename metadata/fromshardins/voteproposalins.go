package fromshardins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"strconv"
)

type NormalVoteProposalIns struct {
	BoardType    common.BoardType
	VoteProposal component.VoteProposalData
}

func (normalVoteProposalIns NormalVoteProposalIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(normalVoteProposalIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.NormalVoteProposalIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewNormalVoteProposalIns(boardType common.BoardType, voteProposal component.VoteProposalData) *NormalVoteProposalIns {
	return &NormalVoteProposalIns{BoardType: boardType,  VoteProposal: voteProposal}
}

func NewNormalVoteProposalInsFromStr(inst string) (*NormalVoteProposalIns, error) {
	Ins := &NormalVoteProposalIns{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, err
	}
	return Ins, nil
}

type PunishDeryptIns struct {
	BoardType common.BoardType
}

func (punishDeryptIns PunishDeryptIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(punishDeryptIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.PunishDecryptIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewPunishDeryptIns(boardType common.BoardType) *PunishDeryptIns {
	return &PunishDeryptIns{BoardType: boardType}
}
