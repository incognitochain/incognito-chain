package transaction

import "github.com/ninjadotorg/constant/common"

type TxVoteDCBBoard struct {
	TxCustomToken
	VoteDCBBoardData VoteDCBBoardData
}

type VoteDCBBoardData struct {
	CandidatePubKey string
}

type TxVoteGOVBoard struct {
	TxCustomToken
	VoteGOVBoardData VoteGOVBoardData
}

type VoteGOVBoardData struct {
	CandidatePubKey string
}

func (thisTx TxVoteDCBBoard) Hash() *common.Hash {
	record := string(common.ToBytes(*thisTx.TxCustomToken.Hash()))
	record += string(common.ToBytes(*thisTx.VoteDCBBoardData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (VoteDCBBoardData VoteDCBBoardData) Hash() *common.Hash {
	record := VoteDCBBoardData.CandidatePubKey
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (thisTx TxVoteGOVBoard) Hash() *common.Hash {
	record := string(common.ToBytes(*thisTx.TxCustomToken.Hash()))
	record += string(common.ToBytes(*thisTx.VoteGOVBoardData.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (VoteGOVBoardData VoteGOVBoardData) Hash() *common.Hash {
	record := VoteGOVBoardData.CandidatePubKey
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

//xxx
func (TxVoteGOVBoard) Validate() bool {
	return true
}
