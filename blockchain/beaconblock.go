package blockchain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

const (
	COMMITEES     = 3
	OFFSET        = 1
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
	R             string  `json:"r"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`

	Body   BeaconBody
	Header BeaconHeader
}

func NewBeaconBlock() BeaconBlock {
	return BeaconBlock{}
}
func (self *BeaconBlock) Hash() *common.Hash {
	hash := self.Header.Hash()
	return &hash
}

func (self *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string  `json:"AggregatedSig"`
		ValidatorsIdx [][]int `json:"ValidatorsIdx"`
		ProducerSig   string  `json:"ProducerSig"`
		R             string  `json:"r"`
		Header        BeaconHeader
		Body          BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBlk)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self.AggregatedSig = tempBlk.AggregatedSig
	self.R = tempBlk.R
	self.ValidatorsIdx = tempBlk.ValidatorsIdx
	self.ProducerSig = tempBlk.ProducerSig
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

	self.Header = tempBlk.Header

	self.Body = tempBlk.Body
	return nil
}

func (self *BeaconBody) toString() string {
	res := ""
	for _, l := range self.ShardState {
		for _, r := range l {
			res += strconv.Itoa(int(r.Height))
			res += r.Hash.String()
			crossShard, _ := json.Marshal(r.CrossShard)
			res += string(crossShard)

		}
	}

	for _, l := range self.Instructions {
		for _, r := range l {
			res += r
		}
	}
	return res
}

func (self *BeaconBody) Hash() common.Hash {
	return common.DoubleHashH([]byte(self.toString()))
}

// func (self *BeaconBody) UnmarshalJSON(data []byte) error {
// 	type BodyAlias BeaconBody
// 	blkBody := &BodyAlias{}

// 	err := json.Unmarshal(data, blkBody)
// 	if err != nil {
// 		return NewBlockChainError(UnmashallJsonBlockError, err)
// 	}
// 	self.Instructions = blkBody.Instructions
// 	self.ShardState = blkBody.ShardState
// 	return nil
// }

func (self *BeaconHeader) toString() string {
	res := ""
	res += fmt.Sprintf("%v", self.Version)
	res += fmt.Sprintf("%v", self.Height)
	res += fmt.Sprintf("%v", self.Timestamp)
	res += self.PrevBlockHash.String()
	res += self.ShardStateHash.String()
	res += self.InstructionHash.String()
	res += self.Producer
	return res
}

func (self *BeaconHeader) Hash() common.Hash {
	return common.DoubleHashH([]byte(self.toString()))
}

// func (self *BeaconHeader) UnmarshalJSON(data []byte) error {
// 	type HeaderAlias BeaconHeader
// 	blkHeader := &HeaderAlias{}
// 	err := json.Unmarshal(data, blkHeader)
// 	if err != nil {
// 		return NewBlockChainError(UnmashallJsonBlockError, err)
// 	}
// 	self.Height = blkHeader.Height
// 	self.InstructionHash = blkHeader.InstructionHash
// 	self.PrevBlockHash = blkHeader.PrevBlockHash
// 	self.Producer = blkHeader.Producer
// 	self.ShardCandidateRoot = blkHeader.ShardCandidateRoot
// 	self.ShardStateHash = blkHeader.ShardStateHash
// 	self.ShardValidatorsRoot = blkHeader.ShardValidatorsRoot
// 	self.Timestamp = blkHeader.Timestamp
// 	self.Epoch = blkHeader.Epoch
// 	self.ValidatorsRoot = blkHeader.ValidatorsRoot
// 	self.Version = blkHeader.Version
// 	self.BeaconCandidateRoot = blkHeader.BeaconCandidateRoot
// 	return nil
// }
