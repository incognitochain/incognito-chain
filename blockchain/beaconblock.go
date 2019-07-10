package blockchain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
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
	// Shard State extract from shard to beacon block
	// Store all shard state == store content of all shard to beacon block
	ShardState   map[byte][]ShardState
	Instructions [][]string // Random here
}

type BeaconHeader struct {
	ProducerAddress privacy.PaymentAddress
	Version         int    `json:"Version"`
	Height          uint64 `json:"Height"`
	//epoch length should be config in consensus
	Epoch         uint64      `json:"Epoch"`
	Round         int         `json:"Round"`
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
	InstructionHash common.Hash `json:"InstructionHash"`
}

type BeaconBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"R"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`

	ValidationData string `json:"ValidationData"`

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

func (beaconBlock *BeaconBlock) GetHeight() uint64 {
	return beaconBlock.Header.Height
}

func (beaconBlock *BeaconBlock) GetProducerPubKey() string {
	return string(beaconBlock.Header.ProducerAddress.Pk)
}

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string  `json:"AggregatedSig"`
		ValidatorsIdx [][]int `json:"ValidatorsIdx"`
		ProducerSig   string  `json:"ProducerSig"`
		R             string  `json:"R"`

		ValidationData string `json:"ValidationData"`

		Header BeaconHeader
		Body   BeaconBody
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

func (beaconBody *BeaconBody) Hash() common.Hash {
	return common.HashH([]byte(beaconBody.toString()))
}

func (beaconHeader *BeaconHeader) toString() string {
	res := ""
	res += beaconHeader.ProducerAddress.String()
	res += fmt.Sprintf("%v", beaconHeader.Version)
	res += fmt.Sprintf("%v", beaconHeader.Height)
	res += fmt.Sprintf("%v", beaconHeader.Epoch)
	res += fmt.Sprintf("%v", beaconHeader.Round)
	res += fmt.Sprintf("%v", beaconHeader.Timestamp)
	res += beaconHeader.PrevBlockHash.String()
	res += beaconHeader.ValidatorsRoot.String()
	res += beaconHeader.BeaconCandidateRoot.String()
	res += beaconHeader.ShardCandidateRoot.String()
	res += beaconHeader.ShardValidatorsRoot.String()
	res += beaconHeader.ShardStateHash.String()
	res += beaconHeader.InstructionHash.String()
	return res
}

func (beaconBlock *BeaconHeader) Hash() common.Hash {
	return common.HashH([]byte(beaconBlock.toString()))
}

func (beaconBlock *BeaconBlock) AddValidationField(validateData string) error {
	beaconBlock.ValidationData = validateData
	return nil
}
func (beaconBlock *BeaconBlock) GetValidationField() string {
	return beaconBlock.ValidationData
}
