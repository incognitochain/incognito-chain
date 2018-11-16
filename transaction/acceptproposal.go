package transaction

import "github.com/ninjadotorg/constant/common"

type TxAcceptGovProposal struct {
	Tx
	GovProposalTXID *common.Hash
}

type TxAcceptDCBProposal struct {
	Tx
	DCBProposalTXID *common.Hash
}

func (thisTx TxAcceptDCBProposal) Hash() *common.Hash {
	record := string(common.ToBytes(thisTx.Tx.Hash()))
	record += string(common.ToBytes(thisTx.DCBProposalTXID))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxAcceptGovProposal) Hash() *common.Hash {
	record := string(common.ToBytes(thisTx.Tx.Hash()))
	record += string(common.ToBytes(thisTx.GovProposalTXID))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func ValidateDCBTXID(DCBProposalTXID *common.Hash) bool {
	return true
	// xxx check if this TXID point to some DCB Proposal
}

func ValidateGovTXID(GovProposalTXID *common.Hash) bool {
	return true
	// xxx check if this TXID point to some Gov Proposal
}

func (thisTx TxAcceptDCBProposal) Validate() bool {
	return thisTx.Tx.ValidateTransaction() && ValidateDCBTXID(thisTx.DCBProposalTXID)
}

func (thisTx TxAcceptGovProposal) Validate() bool {
	return thisTx.Tx.ValidateTransaction() && ValidateGovTXID(thisTx.GovProposalTXID)
}
