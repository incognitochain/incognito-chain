package transaction

/*
import "github.com/ninjadotorg/constant/common"

type TxVoteDCBProposal struct {
	TxNormal
	VoteDCBProposalData VoteDCBProposalData
}

type TxVoteGOVProposal struct {
	TxNormal
	VoteGOVProposalData VoteGOVProposalData
}

type VoteGOVProposalData struct {
	GOVProposalTXID *common.Hash
	AmountVoteToken uint32
}

type VoteDCBProposalData struct {
	DCBProposalTXID *common.Hash
	AmountVoteToken uint32
}

func (VoteDCBProposalData VoteDCBProposalData) Validate() bool {
	return true
	//xxx check if TXID exist

	//xxx check if AmountVoteToken less than current amount of token this user has
}

func (VoteGOVProposalData VoteGOVProposalData) Validate() bool {
	return true
	//xxx check if TXID exist

	//xxx check if AmountVoteToken less than current amount of token this user has
}

func (VoteDCBProposalData VoteDCBProposalData) Hash() *common.Hash {
	record := string(common.ToBytes(VoteDCBProposalData))
	record += string(VoteDCBProposalData.AmountVoteToken)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (VoteGOVProposalData VoteGOVProposalData) Hash() *common.Hash {
	record := string(common.ToBytes(VoteGOVProposalData))
	record += string(VoteGOVProposalData.AmountVoteToken)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxVoteDCBProposal) Hash() *common.Hash {
	record := string(common.ToBytes(thisTx.TxNormal.Hash()))
	record += string(common.ToBytes(thisTx.VoteDCBProposalData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxVoteGOVProposal) Hash() *common.Hash {
	record := string(common.ToBytes(thisTx.TxNormal.Hash()))
	record += string(common.ToBytes(thisTx.VoteGOVProposalData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxVoteDCBProposal) Validate() bool {
	return thisTx.TxNormal.ValidateTransaction() && thisTx.VoteDCBProposalData.Validate()
}

func (thisTx TxVoteGOVProposal) Validate() bool {
	return thisTx.TxNormal.ValidateTransaction() && thisTx.VoteGOVProposalData.Validate()
}
*/
