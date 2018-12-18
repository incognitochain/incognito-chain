package blockchain

import "github.com/ninjadotorg/constant/common"

type BeaconBody struct {
	ShardState   [][]common.Hash
	Instructions [][]string // Random here
}

type BeaconHeader struct {
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
	CandidateRoot common.Hash `json:"CandidateListRootHash"`
	// Store these two list make sure all node process the same data

	// each shard will have a list of blockHash
	// shardRoot is hash of all list
	ShardMerkleRoot common.Hash `json:"ShardListRootHash"`
	// hash of all parameters == hash of instruction
	InstructionMerkleRoot common.Hash `json:"ParameterHash"`
}

type BeaconBlock struct {
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"BlockProducerSignature"`
	Producer      string `json:"Producer"`

	Body   BeaconBody
	Header BeaconHeader
}

// Hash
// Marshall
// UnMarshall
