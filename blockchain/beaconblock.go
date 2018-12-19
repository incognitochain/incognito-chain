package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/ninjadotorg/constant/common"
)

type BeaconBody struct {
	ShardState   [][]common.Hash
	Instructions [][]string // Random here
}

type BeaconHeader struct {
	Version    int         `json:"Version"`
	ParentHash common.Hash `json:"ParentBlockHash"`
	Height     uint64      `json:"Height"`
	//epoch length should be config in consensus
	Epoch         uint64      `json:"Epoch"`
	Timestamp     int64       `json:"Timestamp"`
	PrevBlockHash common.Hash `json:"PrevBlockHash"`

	//Validator list will be store in database/memory (locally)
	ValidatorsRoot common.Hash `json:"CurrentValidatorRootHash"`
	//Candidate = unassigned_validator list will be store in database/memory (locally)
	// infer from history
	CandidateRoot common.Hash `json:"CandidateListRootHash"`
	// Store these two list make sure all node process the same data

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
	Producer      string `json:"Producer"`

	Body   BeaconBody
	Header BeaconHeader
}

func (self *BeaconBlock) Hash() *common.Hash {
	record := common.EmptyString
	record += self.Header.Hash().String() + self.ProducerSig + self.AggregatedSig + common.IntArrayToString(self.ValidatorsIdx, ",")
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

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

func (self *BeaconBody) UnmarshalJSON(data []byte) error {
	blkBody := &BeaconBody{}

	err := json.Unmarshal(data, blkBody)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self = blkBody
	return nil
}

func (self *BeaconHeader) toString() string {
	res := ""
	res += fmt.Sprintf("%v", self.Version)
	res += fmt.Sprintf("%v", self.Height)
	res += fmt.Sprintf("%v", self.Timestamp)
	res += self.PrevBlockHash.String()
	res += self.ShardStateHash.String()
	res += self.InstructionHash.String()
	return res
}

func (self *BeaconHeader) Hash() common.Hash {
	return common.DoubleHashH([]byte(self.toString()))
}

func (self *BeaconHeader) UnmarshalJSON(data []byte) error {
	blkHeader := &BeaconHeader{}
	err := json.Unmarshal(data, blkHeader)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self = blkHeader
	return nil
}

func (self *BeaconHeader) GetHeight() uint64 {
	return self.Height
}
