package blockchain

import (
	"encoding/json"
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

type BlockHeaderShard struct {
	BlockHeaderGeneric
	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot      common.Hash
	MerkleRootShard common.Hash
	Actions         []interface{}
	ShardID         byte
}

func (self BlockHeaderShard) Hash() common.Hash {
	record := common.EmptyString

	// add data from header
	record += strconv.FormatInt(self.Timestamp, 10) +
		string(self.ShardID) +
		self.MerkleRoot.String() +
		self.PrevBlockHash.String()

	return common.DoubleHashH([]byte(record))
}

func (self *BlockHeaderShard) UnmarshalJSON([]byte) error {
	type AliasHeader BlockHeaderShard
	temp := &AliasHeader{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self = temp
	return nil
}
