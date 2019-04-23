package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
)

type KeepOldProposalIns struct {
	BoardType common.BoardType
}

func NewKeepOldProposalIns(
	boardType common.BoardType,
) *KeepOldProposalIns {
	return &KeepOldProposalIns{BoardType: boardType}
}

func (keepOldProposalIns KeepOldProposalIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(keepOldProposalIns)
	if err != nil {
		return nil, err
	}
	var t int
	if keepOldProposalIns.BoardType == common.DCBBoard {
		t = component.KeepOldDCBProposalIns
	} else {
		t = component.KeepOldGOVProposalIns
	}
	return []string{
		strconv.Itoa(t),
		strconv.Itoa(component.AllShards),
		string(content),
	}, nil
}

func NewKeepOldProposalInsFromStr(inst []string) (*KeepOldProposalIns, error) {
	keepOldProposalIns := &KeepOldProposalIns{}
	err := json.Unmarshal([]byte(inst[2]), keepOldProposalIns)
	if err != nil {
		return nil, err
	}
	return keepOldProposalIns, nil
}
