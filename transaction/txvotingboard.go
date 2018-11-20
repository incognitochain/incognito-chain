package transaction

import "github.com/ninjadotorg/constant/common"

type TxVoteDCBBoard struct {
	Tx
	VoteDCBBoardData VoteDCBBoardData
}

type VoteDCBBoardData struct {
	CandidatePubKey string
	AmountDCBToken uint32
}

type TxVoteGOVBoard struct {
	Tx
	VoteGOVBoardData VoteGOVBoardData
}

type VoteGOVBoardData struct {
	CandidatePubKey string
	AmountGOVToken uint32
}

func (thisTx TxVoteDCBBoard) Hash() *common.Hash{
	record := string(common.ToBytes(thisTx.Hash()))
	record += string(common.ToBytes(thisTx.VoteDCBBoardData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (VoteDCBBoardData VoteDCBBoardData) Hash() *common.Hash {
	record := string(VoteDCBBoardData.AmountDCBToken) + VoteDCBBoardData.CandidatePubKey
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxVoteGOVBoard) Hash() *common.Hash{
	record := string(common.ToBytes(thisTx.Hash()))
	record += string(common.ToBytes(thisTx.VoteGOVBoardData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (VoteGOVBoardData VoteGOVBoardData) Hash() *common.Hash {
	record := string(VoteGOVBoardData.AmountGOVToken) + VoteGOVBoardData.CandidatePubKey
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

//xxx
func (TxVoteDCBBoard) Validate() bool {
	return true
}
func (TxVoteGOVBoard) Validate() bool {
	return true
}
