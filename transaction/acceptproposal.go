package transaction

import "github.com/ninjadotorg/constant/common"

type TxlAcceptGovProposal struct {
	Tx
	GovProposalTXID *common.Hash
}

type TxlAcceptDCBProposal struct {
	Tx
	DCBProposalTXID *common.Hash
}

func (thisTx TxlAcceptDCBProposal) Hash() *common.Hash {
	record := string(common.ToBytes(thisTx.Tx.Hash()))
	record += string(common.ToBytes(thisTx.DCBProposalTXID))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxlAcceptGovProposal) Hash() *common.Hash {
	record := string(common.ToBytes(thisTx.Tx.Hash()))
	record += string(common.ToBytes(thisTx.GovProposalTXID))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
