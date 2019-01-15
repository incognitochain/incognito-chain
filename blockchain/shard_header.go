package blockchain

import (
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

/*
	-MerkleRoot and MerkleRootShard: make from transaction
	-Validator Root is root hash of current committee in beststate
	-PendingValidator Root is root hash of pending validator in beststate
*/
type ShardHeader struct {
	ShardID              byte
	Producer             string      `json:"Producer"`
	Version              int         `json:"Version"`
	Height               uint64      `json:"Height"`
	Epoch                uint64      `json:"Epoch"`
	Timestamp            int64       `json:"Timestamp"`
	PrevBlockHash        common.Hash `json:"PrevBlockHash"`
	SalaryFund           uint64      `json:"SalaryFund"`
	MerkleRoot           common.Hash `json:"TransactionRoot"`
	MerkleRootShard      common.Hash `json:"ShardTransactionRoot"`
	ActionsRoot          common.Hash `json:"ActionsRootHash"`
	CommitteeRoot        common.Hash `json:"CurrentValidatorRootHash"`
	PendingValidatorRoot common.Hash `json:"PendingValidatorRoot"`
	//@Hung: These field contains real data, should not be included in header
	//@Hung: How to make sure cross shard byte map contain all data from other
	//@Hung: CrossShardByteMap should have hash
	CrossShardByteMap []byte
	Actions           [][]string
}

func (self ShardHeader) Hash() common.Hash {
	record := common.EmptyString

	// add data from header
	record += strconv.FormatInt(self.Timestamp, 10) +
		string(self.ShardID) +
		self.MerkleRoot.String() +
		self.PrevBlockHash.String() + self.Producer

	return common.DoubleHashH([]byte(record))
}
