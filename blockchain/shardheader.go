package blockchain

import (
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

type ShardHeader struct {
	Version int    `json:"Version"`
	Height  uint64 `json:"Height"`
	//epoch length should be config in consensus
	Epoch         uint64      `json:"Epoch"`
	Timestamp     int64       `json:"Timestamp"`
	PrevBlockHash common.Hash `json:"PrevBlockHash"`
	SalaryFund    uint64
	//Validator list will be store in database/memory (locally)
	ValidatorsRoot common.Hash `json:"CurrentValidatorRootHash"`
	//Candidate = unassigned_validator list will be store in database/memory (locally)
	// infer from history
	PendingValidatorRoot common.Hash `json:"PendingValidatorRoot"`
	// Store these two list make sure all node process the same data

	MerkleRoot      common.Hash
	MerkleRootShard common.Hash
	RefBlocksHash   common.Hash
	Actions         []interface{}
	ShardID         byte
}

func (self ShardHeader) Hash() common.Hash {
	record := common.EmptyString

	// add data from header
	record += strconv.FormatInt(self.Timestamp, 10) +
		string(self.ShardID) +
		self.MerkleRoot.String() +
		self.PrevBlockHash.String()

	return common.DoubleHashH([]byte(record))
}
