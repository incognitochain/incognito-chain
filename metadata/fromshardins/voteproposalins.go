package fromshardins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"strconv"
)

type SealedLv1Or2VoteProposalIns struct {
	BoardType common.BoardType
	Lv3TxID   common.Hash
}

func (sealedLv1Or2VoteProposalIns SealedLv1Or2VoteProposalIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(sealedLv1Or2VoteProposalIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.SealedLv1Or2VoteProposalIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewSealedLv1Or2VoteProposalIns(boardType common.BoardType, lv3TxID common.Hash) *SealedLv1Or2VoteProposalIns {
	return &SealedLv1Or2VoteProposalIns{BoardType: boardType, Lv3TxID: lv3TxID}
}

func NewSealedLv1Or2VoteProposalInsFromStr(inst string) (*SealedLv1Or2VoteProposalIns, error) {
	sealedLv1Or2VoteProposalIns := &SealedLv1Or2VoteProposalIns{}
	err := json.Unmarshal([]byte(inst), sealedLv1Or2VoteProposalIns)
	if err != nil {
		return nil, err
	}
	return sealedLv1Or2VoteProposalIns, nil
}

type SealedLv3VoteProposalIns struct {
	BoardType common.BoardType
	Lv3TxID   common.Hash
}

func (sealedLv3VoteProposalIns SealedLv3VoteProposalIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(sealedLv3VoteProposalIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.SealedLv3VoteProposalIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewSealedLv3VoteProposalIns(boardType common.BoardType, lv3TxID common.Hash) *SealedLv3VoteProposalIns {
	return &SealedLv3VoteProposalIns{BoardType: boardType, Lv3TxID: lv3TxID}
}

func NewSealedLv3VoteProposalInsFromStr(inst string) (*SealedLv3VoteProposalIns, error) {
	sealedLv3VoteProposalIns := &SealedLv3VoteProposalIns{}
	err := json.Unmarshal([]byte(inst), sealedLv3VoteProposalIns)
	if err != nil {
		return nil, err
	}
	return sealedLv3VoteProposalIns, nil
}

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
