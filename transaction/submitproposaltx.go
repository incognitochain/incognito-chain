package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/voting"
)

type TxSubmitGOVProposal struct {
	Tx
	GOVProposalData voting.GOVProposalData
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

func (thisTx TxSubmitGOVProposal) Hash() *common.Hash{
	record := string(common.ToBytes(thisTx.Tx.Hash()))
	record += string(common.ToBytes(thisTx.GOVProposalData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxSubmitDCBProposal) ValidateTransaction() bool {
	return thisTx.Tx.ValidateTransaction() && thisTx.DCBProposalData.Validate()
}

