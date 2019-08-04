package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconHeader struct {
	ProducerAddress privacy.PaymentAddress
	Version         int    `json:"Version"`
	Height          uint64 `json:"Height"`
	//epoch length should be config in consensus
	Epoch             uint64      `json:"Epoch"`
	Round             int         `json:"Round"`
	Timestamp         int64       `json:"Timestamp"`
	PreviousBlockHash common.Hash `json:"PreviousBlockHash"`
	//Validator list will be store in database/memory (locally)
	//Build from two list: BeaconCommittee + BeaconPendingValidator
	ValidatorsRoot common.Hash `json:"CurrentValidatorRootHash"`
	//Candidate = unassigned_validator list will be store in database/memory (locally)
	//Build from two list: CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
	// infer from history
	// Candidate public key for beacon chain
	BeaconCandidateRoot common.Hash `json:"BeaconCandidateRoot"` // Candidate public key for all shard
	ShardCandidateRoot  common.Hash `json:"ShardCandidateRoot"`  // Shard validator build from ShardCommittee and ShardPendingValidator
	ShardValidatorsRoot common.Hash `json:"ShardValidatorRoot"`
	ShardStateHash      common.Hash `json:"ShardListRootHash"` // each shard will have a list of blockHash, shardRoot is hash of all list
	InstructionHash     common.Hash `json:"InstructionHash"`   // hash of all parameters == hash of instruction
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
	res += beaconHeader.ValidatorsRoot.String()
	res += beaconHeader.BeaconCandidateRoot.String()
	res += beaconHeader.ShardCandidateRoot.String()
	res += beaconHeader.ShardValidatorsRoot.String()
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
