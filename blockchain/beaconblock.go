package blockchain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type BeaconBlock struct {
	ValidationData string `json:"ValidationData"`
	Body           BeaconBody
	Header         BeaconHeader
}

type BeaconBody struct {
	// Shard State extract from shard to beacon block
	// Store all shard state == store content of all shard to beacon block
	ShardState   map[byte][]ShardState
	Instructions [][]string
}

type BeaconHeader struct {
	Version           int         `json:"Version"`
	Height            uint64      `json:"Height"`
	Epoch             uint64      `json:"Epoch"`
	Round             int         `json:"Round"`
	Timestamp         int64       `json:"Timestamp"`
	PreviousBlockHash common.Hash `json:"PreviousBlockHash"`
	InstructionHash   common.Hash `json:"InstructionHash"` // hash of all parameters == hash of instruction
	ShardStateHash    common.Hash `json:"ShardStateHash"`  // each shard will have a list of blockHash, shardRoot is hash of all list
	// Merkle root of all instructions (using Keccak256 hash func) to relay to Ethreum
	// This obsoletes InstructionHash but for simplicity, we keep it for now
	InstructionMerkleRoot           common.Hash `json:"InstructionMerkleRoot"`
	BeaconCommitteeAndValidatorRoot common.Hash `json:"BeaconCommitteeAndValidatorRoot"` //Build from two list: BeaconCommittee + BeaconPendingValidator
	BeaconCandidateRoot             common.Hash `json:"BeaconCandidateRoot"`             // CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
	ShardCandidateRoot              common.Hash `json:"ShardCandidateRoot"`              // CandidateShardWaitingForCurrentRandom + CandidateShardWaitingForNextRandom
	ShardCommitteeAndValidatorRoot  common.Hash `json:"ShardCommitteeAndValidatorRoot"`
	AutoStakingRoot                 common.Hash `json:"AutoStakingRoot"`
	ConsensusType                   string      `json:"ConsensusType"`
	Producer                        string      `json:"Producer"`
	ProducerPubKeyStr               string      `json:"ProducerPubKeyStr"`

	//for version 2
	Proposer    string `json:"Proposer"`
	ProposeTime int64  `json:"ProposeTime"`
}

type ShardState struct {
	Height     uint64
	Hash       common.Hash
	CrossShard []byte //In this state, shard i send cross shard tx to which shard
}

func (beaconBlock *BeaconBlock) GetVersion() int {
	return beaconBlock.Header.Version
}

func (beaconBlock *BeaconBlock) GetPrevHash() common.Hash {
	return beaconBlock.Header.PreviousBlockHash
}

func NewBeaconBlock() *BeaconBlock {
	return &BeaconBlock{}
}

func (beaconBlock *BeaconBlock) GetProposer() string {
	return beaconBlock.Header.Proposer
}

func (beaconBlock *BeaconBlock) GetProposeTime() int64 {
	return beaconBlock.Header.ProposeTime
}

func (beaconBlock *BeaconBlock) GetProduceTime() int64 {
	return beaconBlock.Header.Timestamp
}

func (beaconBlock BeaconBlock) Hash() *common.Hash {
	hash := beaconBlock.Header.Hash()
	return &hash
}

func (beaconBlock BeaconBlock) GetCurrentEpoch() uint64 {
	return beaconBlock.Header.Epoch
}

func (beaconBlock BeaconBlock) GetHeight() uint64 {
	return beaconBlock.Header.Height
}
func (beaconBlock BeaconBlock) GetShardID() int {
	return -1
}

// func (beaconBlock *BeaconBlock) GetProducerPubKey() string {
// 	return string(beaconBlock.Header.ProducerAddress.Pk)
// }

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBeaconBlock := &struct {
		ValidationData string `json:"ValidationData"`
		Header         BeaconHeader
		Body           BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBeaconBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	// beaconBlock.AggregatedSig = tempBlk.AggregatedSig
	// beaconBlock.R = tempBlk.R
	// beaconBlock.ValidatorsIdx = tempBlk.ValidatorsIdx
	// beaconBlock.ProducerSig = tempBlk.ProducerSig
	beaconBlock.ValidationData = tempBeaconBlock.ValidationData
	beaconBlock.Header = tempBeaconBlock.Header
	beaconBlock.Body = tempBeaconBlock.Body
	return nil
}

func (beaconBlock *BeaconBlock) AddValidationField(validationData string) error {
	beaconBlock.ValidationData = validationData
	return nil
}
func (beaconBlock BeaconBlock) GetValidationField() string {
	return beaconBlock.ValidationData
}

func (beaconBlock BeaconBlock) GetRound() int {
	return beaconBlock.Header.Round
}
func (beaconBlock BeaconBlock) GetRoundKey() string {
	return fmt.Sprint(beaconBlock.Header.Height, "_", beaconBlock.Header.Round)
}

func (beaconBlock BeaconBlock) GetInstructions() [][]string {
	return beaconBlock.Body.Instructions
}

func (beaconBlock BeaconBlock) GetProducer() string {
	return beaconBlock.Header.Producer
}

func (beaconBlock BeaconBlock) GetProducerPubKeyStr() string {
	return beaconBlock.Header.ProducerPubKeyStr
}

func (beaconBlock BeaconBlock) GetConsensusType() string {
	return beaconBlock.Header.ConsensusType
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

func (beaconBody BeaconBody) Hash() common.Hash {
	return common.HashH([]byte(beaconBody.toString()))
}

func (beaconHeader *BeaconHeader) toString() string {
	res := ""
	// res += beaconHeader.ProducerAddress.String()
	res += fmt.Sprintf("%v", beaconHeader.Version)
	res += fmt.Sprintf("%v", beaconHeader.Height)
	res += fmt.Sprintf("%v", beaconHeader.Epoch)
	res += fmt.Sprintf("%v", beaconHeader.Round)
	res += fmt.Sprintf("%v", beaconHeader.Timestamp)
	res += beaconHeader.PreviousBlockHash.String()
	res += beaconHeader.BeaconCommitteeAndValidatorRoot.String()
	res += beaconHeader.BeaconCandidateRoot.String()
	res += beaconHeader.ShardCandidateRoot.String()
	res += beaconHeader.ShardCommitteeAndValidatorRoot.String()
	res += beaconHeader.AutoStakingRoot.String()
	res += beaconHeader.ShardStateHash.String()
	res += beaconHeader.InstructionHash.String()

	if beaconHeader.Version == 2 {
		res += beaconHeader.Proposer
		res += fmt.Sprintf("%v", beaconHeader.ProposeTime)
	}
	return res
}

func (beaconBlock *BeaconHeader) MetaHash() common.Hash {
	return common.Keccak256([]byte(beaconBlock.toString()))
}

func (beaconBlock *BeaconHeader) Hash() common.Hash {
	// Block header of beacon uses Keccak256 as a hash func to check on Ethereum when relaying blocks
	blkMetaHash := beaconBlock.MetaHash()
	blkInstHash := beaconBlock.InstructionMerkleRoot
	combined := append(blkMetaHash[:], blkInstHash[:]...)
	return common.Keccak256(combined)
}

func CreateGenesisBeaconBlock(
	version int,
	net uint16,
	genesisBlockTime string,
	genesisParams *GenesisParams,
) *BeaconBlock {
	inst := [][]string{}
	shardAutoStaking := []string{}
	beaconAutoStaking := []string{}
	for i := 0; i < len(genesisParams.PreSelectShardNodeSerializedPubkey); i++ {
		shardAutoStaking = append(shardAutoStaking, "false")
	}
	for i := 0; i < len(genesisParams.PreSelectBeaconNodeSerializedPubkey); i++ {
		beaconAutoStaking = append(beaconAutoStaking, "false")
	}
	// build validator beacon
	// test generate public key in utility/generateKeys
	beaconAssignInstruction := []string{StakeAction}
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPubkey[:], ","))
	beaconAssignInstruction = append(beaconAssignInstruction, "beacon")
	beaconAssignInstruction = append(beaconAssignInstruction, []string{""}...)
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPaymentAddress[:], ","))
	beaconAssignInstruction = append(beaconAssignInstruction, strings.Join(beaconAutoStaking[:], ","))

	shardAssignInstruction := []string{StakeAction}
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPubkey[:], ","))
	shardAssignInstruction = append(shardAssignInstruction, "shard")
	shardAssignInstruction = append(shardAssignInstruction, []string{""}...)
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPaymentAddress[:], ","))
	shardAssignInstruction = append(shardAssignInstruction, strings.Join(shardAutoStaking[:], ","))

	inst = append(inst, beaconAssignInstruction)
	inst = append(inst, shardAssignInstruction)

	// init network param
	inst = append(inst, []string{SetAction, "randomnumber", strconv.Itoa(int(0))})

	layout := "2006-01-02T15:04:05.000Z"
	str := genesisBlockTime
	genesisTime, err := time.Parse(layout, str)

	if err != nil {
		fmt.Println(err)
	}
	body := BeaconBody{ShardState: nil, Instructions: inst}
	header := BeaconHeader{
		Timestamp:                       genesisTime.Unix(),
		Version:                         version,
		Epoch:                           1,
		Height:                          1,
		Round:                           1,
		PreviousBlockHash:               common.Hash{},
		BeaconCommitteeAndValidatorRoot: common.Hash{},
		BeaconCandidateRoot:             common.Hash{},
		ShardCandidateRoot:              common.Hash{},
		ShardCommitteeAndValidatorRoot:  common.Hash{},
		ShardStateHash:                  common.Hash{},
		InstructionHash:                 common.Hash{},
	}

	block := &BeaconBlock{
		Body:   body,
		Header: header,
	}

	return block
}

func GetBeaconSwapInstructionKeyListV2(genesisParams *GenesisParams, epoch uint64) ([]string, []string) {
	newCommittees := genesisParams.SelectBeaconNodeSerializedPubkeyV2[epoch]
	newRewardReceivers := genesisParams.SelectBeaconNodeSerializedPaymentAddressV2[epoch]
	oldCommittees := genesisParams.PreSelectBeaconNodeSerializedPubkey
	beaconSwapInstructionKeyListV2 := []string{SwapAction, strings.Join(newCommittees, ","), strings.Join(oldCommittees, ","), "beacon", "", "", strings.Join(newRewardReceivers, ",")}
	return beaconSwapInstructionKeyListV2, newCommittees
}
