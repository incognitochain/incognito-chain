package frombeaconins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"strconv"
)

type AcceptProposalIns struct {
	BoardType common.BoardType
	TxID      common.Hash
	Voter     component.Voter
	ShardID   byte
}

func NewAcceptProposalIns(
	boardType common.BoardType,
	txID common.Hash,
	voter component.Voter,
	shardID byte,
) *AcceptProposalIns {
	return &AcceptProposalIns{BoardType: boardType, TxID: txID, Voter: voter, ShardID: shardID}
}

func (acceptProposalIns AcceptProposalIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(acceptProposalIns)
	if err != nil {
		return nil, err
	}
	var t int
	if acceptProposalIns.BoardType == common.DCBBoard {
		t = component.AcceptDCBProposalIns
	} else {
		t = component.AcceptGOVProposalIns
	}
	return []string{
		strconv.Itoa(t),
		strconv.Itoa(int(acceptProposalIns.ShardID)),
		string(content),
	}, nil
}
