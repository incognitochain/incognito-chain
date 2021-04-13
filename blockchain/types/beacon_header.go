package types

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
)

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

func NewBeaconHeader(version int, height uint64, epoch uint64, round int, timestamp int64, previousBlockHash common.Hash, consensusType string, producer string, producerPubKeyStr string) BeaconHeader {
	return BeaconHeader{Version: version, Height: height, Epoch: epoch, Round: round, Timestamp: timestamp, PreviousBlockHash: previousBlockHash, ConsensusType: consensusType, Producer: producer, ProducerPubKeyStr: producerPubKeyStr}
}

func (beaconHeader *BeaconHeader) AddBeaconHeaderHash(instructionHash common.Hash, shardStateHash common.Hash, instructionMerkleRoot []byte, beaconCommitteeAndValidatorRoot common.Hash, beaconCandidateRoot common.Hash, shardCandidateRoot common.Hash, shardCommitteeAndValidatorRoot common.Hash, autoStakingRoot common.Hash) {
	beaconHeader.InstructionHash = instructionHash
	beaconHeader.ShardStateHash = shardStateHash
	copy(beaconHeader.InstructionMerkleRoot[:], instructionMerkleRoot)
	beaconHeader.BeaconCommitteeAndValidatorRoot = beaconCommitteeAndValidatorRoot
	beaconHeader.BeaconCandidateRoot = beaconCandidateRoot
	beaconHeader.ShardCandidateRoot = shardCandidateRoot
	beaconHeader.ShardCommitteeAndValidatorRoot = shardCommitteeAndValidatorRoot
	beaconHeader.AutoStakingRoot = autoStakingRoot

}

func (header *BeaconHeader) toString() string {
	res := ""
	// res += beaconHeader.ProducerAddress.String()
	res += fmt.Sprintf("%v", header.Version)
	res += fmt.Sprintf("%v", header.Height)
	res += fmt.Sprintf("%v", header.Epoch)
	res += fmt.Sprintf("%v", header.Round)
	res += fmt.Sprintf("%v", header.Timestamp)
	res += header.PreviousBlockHash.String()
	res += header.BeaconCommitteeAndValidatorRoot.String()
	res += header.BeaconCandidateRoot.String()
	res += header.ShardCandidateRoot.String()
	res += header.ShardCommitteeAndValidatorRoot.String()
	res += header.AutoStakingRoot.String()
	res += header.ShardStateHash.String()
	res += header.InstructionHash.String()

	if header.Version == 2 {
		res += header.Proposer
		res += fmt.Sprintf("%v", header.ProposeTime)
	}
	return res
}

func (header *BeaconHeader) MetaHash() common.Hash {
	return common.Keccak256([]byte(header.toString()))
}

func (header *BeaconHeader) Hash() common.Hash {
	// Block header of beacon uses Keccak256 as a hash func to check on Ethereum when relaying blocks
	blkMetaHash := header.MetaHash()
	blkInstHash := header.InstructionMerkleRoot
	combined := append(blkMetaHash[:], blkInstHash[:]...)
	return common.Keccak256(combined)
}
