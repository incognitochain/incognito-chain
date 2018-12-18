package blocktype

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
)

type ShardBody struct {
	RefBlocks    []BlockRef
	Transactions []metadata.Transaction
}
type BlockRef struct {
	ShardID byte
	Block   common.Hash
}

type ShardHeader struct {
	Version    int         `json:"Version"`
	ParentHash common.Hash `json:"ParentBlockHash"`
	Height     uint64      `json:"Height"`
	//epoch length should be config in consensus
	Epoch     uint64 `json:"Epoch"`
	Timestamp int64  `json:"Timestamp"`

	//Validator list will be store in database/memory (locally)
	ValidatorsRoot common.Hash `json:"CurrentValidatorRootHash"`
	//Candidate = unassigned_validator list will be store in database/memory (locally)
	// infer from history
	PendingValidatorRoot common.Hash `json:"PendingValidatorRoot"`
	// Store these two list make sure all node process the same data

	MerkleRoot      common.Hash
	MerkleRootShard common.Hash
	Actions         []interface{}
	ShardID         byte
}

type ShardBlock struct {
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"BlockProducerSignature"`
	Producer      string `json:"Producer"`

	Body   ShardBody
	Header ShardHeader
}

type ShardToBeaconBlock struct {
	Header        ShardHeader
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"BlockProducerSignature"`
	Producer      string `json:"Producer"`
}

type ShardToShardBlock struct {
	///
}

// Hash
// Marshall
// UnMarshall
