package blockchain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

const (
	VERSION       = 1
	RANDOM_NUMBER = 3
)

type ShardState struct {
	Height uint64
	Hash   common.Hash
	//In this state, shard i send cross shard tx to which shard
	CrossShard []byte
}
type BeaconBody struct {
	ShardState   map[byte][]ShardState
	Instructions [][]string // Random here
}

type BeaconHeader struct {
	Producer string `json:"Producer"`
	Version  int    `json:"Version"`
	Height   uint64 `json:"Height"`
	//epoch length should be config in consensus
	Epoch         uint64      `json:"Epoch"`
	Timestamp     int64       `json:"Timestamp"`
	PrevBlockHash common.Hash `json:"PrevBlockHash"`

	//Validator list will be store in database/memory (locally)
	//Build from two list: BeaconCommittee + BeaconPendingValidator
	ValidatorsRoot common.Hash `json:"CurrentValidatorRootHash"`
	//Candidate = unassigned_validator list will be store in database/memory (locally)
	//Build from two list: CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
	// infer from history
	// Candidate public key for beacon chain
	BeaconCandidateRoot common.Hash `json:"BeaconCandidateRoot"`

	// Candidate public key for all shard
	ShardCandidateRoot common.Hash `json:"ShardCandidateRoot"`

	// Shard validator build from ShardCommittee and ShardPendingValidator
	ShardValidatorsRoot common.Hash `json:"ShardValidatorRoot"`

	// each shard will have a list of blockHash
	// shardRoot is hash of all list
	ShardStateHash common.Hash `json:"ShardListRootHash"`
	// hash of all parameters == hash of instruction
	InstructionHash common.Hash `json:"ParameterHash"`
}

type BeaconBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"R"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`

	Body   BeaconBody
	Header BeaconHeader
}

func NewBeaconBlock() BeaconBlock {
	return BeaconBlock{}
}
func (beaconBlock *BeaconBlock) Hash() *common.Hash {
	hash := beaconBlock.Header.Hash()
	return &hash
}

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string  `json:"AggregatedSig"`
		ValidatorsIdx [][]int `json:"ValidatorsIdx"`
		ProducerSig   string  `json:"ProducerSig"`
		R             string  `json:"R"`
		Header        BeaconHeader
		Body          BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBlk)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	beaconBlock.AggregatedSig = tempBlk.AggregatedSig
	beaconBlock.R = tempBlk.R
	beaconBlock.ValidatorsIdx = tempBlk.ValidatorsIdx
	beaconBlock.ProducerSig = tempBlk.ProducerSig
	// blkBody := BeaconBody{}
	// err = blkBody.UnmarshalJSON(tempBlk.Body)
	// if err != nil {
	// 	return NewBlockChainError(UnmashallJsonBlockError, err)
	// }
	// blkHeader := BeaconHeader{}
	// err = blkBody.UnmarshalJSON(tempBlk.Header)
	// if err != nil {
	// 	return NewBlockChainError(UnmashallJsonBlockError, err)
	// }

	beaconBlock.Header = tempBlk.Header

	beaconBlock.Body = tempBlk.Body
	return nil
}

func (beaconBlock *BeaconBody) toString() string {
	res := ""
	for _, l := range beaconBlock.ShardState {
		for _, r := range l {
			res += strconv.Itoa(int(r.Height))
			res += r.Hash.String()
			crossShard, _ := json.Marshal(r.CrossShard)
			res += string(crossShard)

		}
	}

	for _, l := range beaconBlock.Instructions {
		for _, r := range l {
			res += r
		}
	}
	return res
}

func (beaconBlock *BeaconBody) Hash() common.Hash {
	return common.DoubleHashH([]byte(beaconBlock.toString()))
}

// func (beaconBlock *BeaconBody) UnmarshalJSON(data []byte) error {
// 	type BodyAlias BeaconBody
// 	blkBody := &BodyAlias{}

// 	err := json.Unmarshal(data, blkBody)
// 	if err != nil {
// 		return NewBlockChainError(UnmashallJsonBlockError, err)
// 	}
// 	beaconBlock.Instructions = blkBody.Instructions
// 	beaconBlock.ShardState = blkBody.ShardState
// 	return nil
// }

func (beaconBlock *BeaconHeader) toString() string {
	res := ""
	res += fmt.Sprintf("%v", beaconBlock.Version)
	res += fmt.Sprintf("%v", beaconBlock.Height)
	res += fmt.Sprintf("%v", beaconBlock.Timestamp)
	res += beaconBlock.PrevBlockHash.String()
	res += beaconBlock.ShardStateHash.String()
	res += beaconBlock.InstructionHash.String()
	res += beaconBlock.Producer
	return res
}

func (beaconBlock *BeaconHeader) Hash() common.Hash {
	return common.DoubleHashH([]byte(beaconBlock.toString()))
}

// func (beaconBlock *BeaconHeader) UnmarshalJSON(data []byte) error {
// 	type HeaderAlias BeaconHeader
// 	blkHeader := &HeaderAlias{}
// 	err := json.Unmarshal(data, blkHeader)
// 	if err != nil {
// 		return NewBlockChainError(UnmashallJsonBlockError, err)
// 	}
// 	beaconBlock.Height = blkHeader.Height
// 	beaconBlock.InstructionHash = blkHeader.InstructionHash
// 	beaconBlock.PrevBlockHash = blkHeader.PrevBlockHash
// 	beaconBlock.Producer = blkHeader.Producer
// 	beaconBlock.ShardCandidateRoot = blkHeader.ShardCandidateRoot
// 	beaconBlock.ShardStateHash = blkHeader.ShardStateHash
// 	beaconBlock.ShardValidatorsRoot = blkHeader.ShardValidatorsRoot
// 	beaconBlock.Timestamp = blkHeader.Timestamp
// 	beaconBlock.Epoch = blkHeader.Epoch
// 	beaconBlock.ValidatorsRoot = blkHeader.ValidatorsRoot
// 	beaconBlock.Version = blkHeader.Version
// 	beaconBlock.BeaconCandidateRoot = blkHeader.BeaconCandidateRoot
// 	return nil
// }
