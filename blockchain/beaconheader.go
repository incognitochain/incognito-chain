package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconHeader struct {
	ProducerAddress                  privacy.PaymentAddress
	Version                          int         `json:"Version"`
	Height                           uint64      `json:"Height"`
	Epoch                            uint64      `json:"Epoch"`
	Round                            int         `json:"Round"`
	Timestamp                        int64       `json:"Timestamp"`
	PreviousBlockHash                common.Hash `json:"PreviousBlockHash"`
	BeaconCommitteeAndValidatorsRoot common.Hash `json:"BeaconCommitteeAndValidatorsRoot"` //Build from two list: BeaconCommittee + BeaconPendingValidator
	BeaconCandidateRoot              common.Hash `json:"BeaconCandidateRoot"`              // CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
	ShardCandidateRoot               common.Hash `json:"ShardCandidateRoot"`               // CandidateShardWaitingForCurrentRandom + CandidateShardWaitingForNextRandom
	ShardCommitteeAndValidatorsRoot  common.Hash `json:"ShardCommitteeAndValidatorsRoot"`
	ShardStateHash                   common.Hash `json:"ShardStateHash"`  // each shard will have a list of blockHash, shardRoot is hash of all list
	InstructionHash                  common.Hash `json:"InstructionHash"` // hash of all parameters == hash of instruction
	// Merkle root of all instructions (using Keccak256 hash func) to relay to Ethreum
	// This obsoletes InstructionHash but for simplicity, we keep it for now
	InstructionMerkleRoot common.Hash
}

func (beaconHeader *BeaconHeader) toString() string {
	res := ""
	res += beaconHeader.ProducerAddress.String()
	res += fmt.Sprintf("%v", beaconHeader.Version)
	res += fmt.Sprintf("%v", beaconHeader.Height)
	res += fmt.Sprintf("%v", beaconHeader.Epoch)
	res += fmt.Sprintf("%v", beaconHeader.Round)
	res += fmt.Sprintf("%v", beaconHeader.Timestamp)
	res += beaconHeader.PreviousBlockHash.String()
	res += beaconHeader.BeaconCommitteeAndValidatorsRoot.String()
	res += beaconHeader.BeaconCandidateRoot.String()
	res += beaconHeader.ShardCandidateRoot.String()
	res += beaconHeader.ShardCommitteeAndValidatorsRoot.String()
	res += beaconHeader.ShardStateHash.String()
	res += beaconHeader.InstructionHash.String()
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
