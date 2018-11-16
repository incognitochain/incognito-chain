package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/voting"
)

type TxSubmitGovProposal struct {
	Tx
	GovProposalData voting.GovProposalData
}

type TxSubmitDCBProposal struct {
	Tx
	DCBProposalData voting.DCBProposalData
}

func (thisTx TxSubmitDCBProposal) Hash() *common.Hash{
	record := string(common.ToBytes(thisTx.Tx.Hash()))
	record += string(common.ToBytes(thisTx.DCBProposalData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxSubmitGovProposal) Hash() *common.Hash{
	record := string(common.ToBytes(thisTx.Tx.Hash()))
	record += string(common.ToBytes(thisTx.GovProposalData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
