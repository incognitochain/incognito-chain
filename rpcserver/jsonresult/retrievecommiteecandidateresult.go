package jsonresult

import "github.com/ninjadotorg/constant/blockchain"

type RetrieveCommitteecCandidateResult struct {
	Value     uint64 `json:"H"`
	Timestamp int64  `json:"Timestamp"`
	ShardID   byte   `json:"ShardID"`
}

func (self *RetrieveCommitteecCandidateResult) Init(obj *blockchain.CommitteeCandidateInfo) {
	// self.ShardID = obj.ShardID
	// self.Timestamp = obj.Timestamp
}
