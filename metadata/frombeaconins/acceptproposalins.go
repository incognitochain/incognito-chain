package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
)

type AcceptProposalIns struct {
	BoardType common.BoardType
	TxID      common.Hash
	Voters    []privacy.PaymentAddress
	ShardID   byte
}

func NewAcceptProposalIns(
	boardType common.BoardType,
	txID common.Hash,
	voters []privacy.PaymentAddress,
	shardID byte,
) *AcceptProposalIns {
	return &AcceptProposalIns{BoardType: boardType, TxID: txID, Voters: voters, ShardID: shardID}
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
