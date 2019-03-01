package frombeaconins

import (
	"encoding/json"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"strconv"
)

type AcceptProposalIns struct {
	boardType metadata.BoardType
	txID      common.Hash
	voter     metadata.Voter
	shardID   byte
}

func NewAcceptProposalIns(boardType metadata.BoardType, txID common.Hash, voter metadata.Voter, shardID byte) *AcceptProposalIns {
	return &AcceptProposalIns{boardType: boardType, txID: txID, voter: voter, shardID: shardID}
}

func (acceptProposalIns AcceptProposalIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(acceptProposalIns)
	if err != nil {
		return nil, err
	}
	var t int
	if acceptProposalIns.boardType == metadata.DCBBoard {
		t = metadata.AcceptDCBProposalIns
	} else {
		t = metadata.AcceptGOVProposalIns
	}
	return []string{
		strconv.Itoa(t),
		strconv.Itoa(int(acceptProposalIns.shardID)),
		string(content),
	}, nil
}
