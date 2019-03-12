package fromshardins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"strconv"
)

type NormalVoteProposalFromSealerIns struct {
	BoardType    common.BoardType
	Lv3TxID      common.Hash
	VoteProposal component.VoteProposalData
}

func (normalVoteProposalFromSealerIns NormalVoteProposalFromSealerIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(normalVoteProposalFromSealerIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.NormalVoteProposalFromSealerIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewNormalVoteProposalFromSealerIns(boardType common.BoardType, lv3TxID common.Hash, voteProposal component.VoteProposalData) *NormalVoteProposalFromSealerIns {
	return &NormalVoteProposalFromSealerIns{BoardType: boardType, Lv3TxID: lv3TxID, VoteProposal: voteProposal}
}

func NewNormalVoteProposalFromSealerInsFromStr(inst string) (*NormalVoteProposalFromSealerIns, error) {
	Ins := &NormalVoteProposalFromSealerIns{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, err
	}
	return Ins, nil
}

type NormalVoteProposalFromOwnerIns struct {
	BoardType    common.BoardType
	Lv3TxID      common.Hash
	VoteProposal component.VoteProposalData
}

func (normalVoteProposalFromOwnerIns NormalVoteProposalFromOwnerIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(normalVoteProposalFromOwnerIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.NormalVoteProposalFromOwnerIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewNormalVoteProposalFromOwnerIns(boardType common.BoardType, lv3TxID common.Hash, voteProposal component.VoteProposalData) *NormalVoteProposalFromOwnerIns {
	return &NormalVoteProposalFromOwnerIns{BoardType: boardType, Lv3TxID: lv3TxID, VoteProposal: voteProposal}
}

func NewNormalVoteProposalFromOwnerInsFromStr(inst string) (*NormalVoteProposalFromOwnerIns, error) {
	Ins := &NormalVoteProposalFromOwnerIns{}
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
