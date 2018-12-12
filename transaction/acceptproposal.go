package transaction

/*import "github.com/ninjadotorg/constant/common"

type TxAcceptGOVProposal struct {
	*TxNormal
	GOVProposalTXID *common.Hash
}

type TxAcceptDCBProposal struct {
	*TxNormal
	DCBProposalTXID *common.Hash
}

func (thisTx TxAcceptDCBProposal) Hash() *common.Hash {
	record := string(common.ToBytes(thisTx.TxNormal.Hash()))
	record += string(common.ToBytes(thisTx.DCBProposalTXID))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxAcceptGOVProposal) Hash() *common.Hash {
	record := string(common.ToBytes(thisTx.TxNormal.Hash()))
	record += string(common.ToBytes(thisTx.GOVProposalTXID))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func ValidateDCBTXID(DCBProposalTXID *common.Hash) bool {
	return true
	// xxx check if this TXID point to some DCB Proposal
}

func ValidateGOVTXID(GOVProposalTXID *common.Hash) bool {
	return true
	// xxx check if this TXID point to some GOV Proposal
}

func (thisTx TxAcceptDCBProposal) ValidateTransaction() bool {
	return thisTx.TxNormal.ValidateTransaction() && ValidateDCBTXID(thisTx.DCBProposalTXID)
}

func (thisTx TxAcceptGOVProposal) ValidateTransaction() bool {
	return thisTx.TxNormal.ValidateTransaction() && ValidateGOVTXID(thisTx.GOVProposalTXID)
}*/
