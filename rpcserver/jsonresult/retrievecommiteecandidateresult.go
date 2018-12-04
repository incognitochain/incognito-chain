package jsonresult

import "github.com/ninjadotorg/constant/blockchain"

type RetrieveCommitteecCandidateResult struct {
	Value     uint64 `json:"H"`
	Timestamp int64  `json:"Timestamp"`
	ChainID   byte   `json:"ChainID"`
}

func (self *RetrieveCommitteecCandidateResult) Init(obj *blockchain.CommitteeCandidateInfo) {
	self.ChainID = obj.ChainID
	self.Timestamp = obj.Timestamp
	self.ChainID = obj.ChainID
}
