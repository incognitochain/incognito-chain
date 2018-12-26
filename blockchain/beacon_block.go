package blockchain

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
)

const (
	EPOCH       = 200
	RANDOM_TIME = 100
	OFFSET      = 3
	VERSION     = 1
)

type BeaconBody struct {
	ShardState   map[byte][]common.Hash
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
	ShardCandidateRoot common.Hash `json:"BeaconCandidateRoot"`

	// Shard validator build from ShardCommittee and ShardPendingValidator
	ShardValidatorsRoot common.Hash `json:"ShardValidatorRoot"`

	// each shard will have a list of blockHash
	// shardRoot is hash of all list
	ShardStateHash common.Hash `json:"ShardListRootHash"`
	// hash of all parameters == hash of instruction
	InstructionHash common.Hash `json:"ParameterHash"`
}

type BeaconBlock struct {
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"BlockProducerSignature"`

	Body   BeaconBody
	Header BeaconHeader
}

func NewBeaconBlock() BeaconBlock {
	return BeaconBlock{}
}
func (self *BeaconBlock) Hash() *common.Hash {
	record := common.EmptyString
	record += self.Header.Hash().String() + self.AggregatedSig + common.IntArrayToString(self.ValidatorsIdx, ",")
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// func (self *BeaconBlock) UnmarshalJSON(data []byte) error {
// 	tempBlk := &struct {
// 		AggregatedSig string
// 		ValidatorsIdx []int
// 		Header        BeaconHeader
// 		Body          *json.RawMessage
// 	}{}
// 	err := json.Unmarshal(data, &tempBlk)
// 	if err != nil {
// 		return NewBlockChainError(UnmashallJsonBlockError, err)
// 	}
// 	self.AggregatedSig = tempBlk.AggregatedSig
// 	self.ValidatorsIdx = tempBlk.ValidatorsIdx

// 	blkBody := BeaconBody{}
// 	err = blkBody.UnmarshalJSON(*tempBlk.Body)
// 	if err != nil {
// 		return NewBlockChainError(UnmashallJsonBlockError, err)
// 	}
// 	self.Header = tempBlk.Header

// 	self.Body = blkBody
// 	return nil
// }

func (self *BeaconBody) toString() string {
	res := ""
	for _, l := range self.ShardState {
		for _, r := range l {
			res += r.String()
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
// 	blkHeader := &BeaconHeader{}
// 	err := json.Unmarshal(data, blkHeader)
// 	if err != nil {
// 		return NewBlockChainError(UnmashallJsonBlockError, err)
// 	}
// 	self = blkHeader
// 	return nil
// }
