package frombeaconins

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type AcceptProposalIns struct {
	boardType byte
	txID      common.Hash
	voter     metadata.Voter
}

func NewAcceptProposalIns(boardType byte, txID common.Hash, voter metadata.Voter) *AcceptProposalIns {
	return &AcceptProposalIns{boardType: boardType, txID: txID, voter: voter}
}

func (AcceptProposalIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func (acceptProposalIns AcceptProposalIns) BuildTransaction(minerPrivateKey *privacy.SpendingKey, db database.DatabaseInterface) (metadata.Transaction, error) {
	txId := acceptProposalIns.txID
	voter := acceptProposalIns.voter
	var meta metadata.Metadata
	if acceptProposalIns.boardType == common.DCBBoard {
		meta = metadata.NewAcceptDCBProposalMetadata(txId, voter)
	} else {
		meta = metadata.NewAcceptGOVProposalMetadata(txId, voter)
	}
	acceptIns := transaction.NewEmptyTx(minerPrivateKey, db, meta)
	return acceptIns, nil
}
