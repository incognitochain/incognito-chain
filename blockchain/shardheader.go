package blockchain

import (
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
)

/*
	-TxRoot and MerkleRootShard: make from transaction
	-Validator Root is root hash of current committee in beststate
	-PendingValidator Root is root hash of pending validator in beststate
*/
type ShardHeader struct {
	Producer              string                 `json:"Producer"`
	ProducerPubKeyStr     string                 `json:"ProducerPubKeyStr"`
	ShardID               byte                   `json:"ShardID"`               // shard ID which block belong to
	Version               int                    `json:"Version"`               // version of block structure
	PreviousBlockHash     common.Hash            `json:"PreviousBlockHash"`     // previous block hash or Parent block hash
	Height                uint64                 `json:"Height"`                // block height
	Round                 int                    `json:"Round"`                 // bpft consensus round
	Epoch                 uint64                 `json:"Epoch"`                 // epoch of block (according to current beacon height)
	CrossShardBitMap      []byte                 `json:"CrossShardBitMap"`      // crossShards bitmap for beacon
	BeaconHeight          uint64                 `json:"BeaconHeight"`          // beacon check point height
	BeaconHash            common.Hash            `json:"BeaconHash"`            // beacon check point hash
	TotalTxsFee           map[common.Hash]uint64 `json:"TotalTxsFee"`           // fee of all txs in block
	ConsensusType         string                 `json:"ConsensusType"`         // consensus type, by which this block is produced
	Timestamp             int64                  `json:"Timestamp"`             // timestamp of block
	TxRoot                common.Hash            `json:"TxRoot"`                // Transaction root created from transaction in shard
	ShardTxRoot           common.Hash            `json:"ShardTxRoot"`           // output root created for other shard
	CrossTransactionRoot  common.Hash            `json:"CrossTransactionRoot"`  // transaction root created from transaction of micro shard to shard block (from other shard)
	InstructionsRoot      common.Hash            `json:"InstructionsRoot"`      // actions root created from Instructions and Metadata of transaction
	CommitteeRoot         common.Hash            `json:"CommitteeRoot"`         // hash from public key list of all committees designated to create this block
	PendingValidatorRoot  common.Hash            `json:"PendingValidatorRoot"`  // hash from public key list of all pending validators designated to this ShardID
	StakingTxRoot         common.Hash            `json:"StakingTxRoot"`         // hash from staking transaction map in shard best state
	InstructionMerkleRoot common.Hash            `json:"InstructionMerkleRoot"` // Merkle root of all instructions (using Keccak256 hash func) to relay to Ethreum
	// This obsoletes InstructionMerkleRoot but for simplicity, we keep it for now

	//for version 2
	Proposer    string
	ProposeTime int64
}

func (shardHeader *ShardHeader) String() string {
	res := common.EmptyString
	// res += shardHeader.ProducerAddress.String()
	res += string(shardHeader.ShardID)
	res += fmt.Sprintf("%v", shardHeader.Version)
	res += shardHeader.PreviousBlockHash.String()
	res += fmt.Sprintf("%v", shardHeader.Height)
	res += fmt.Sprintf("%v", shardHeader.Round)
	res += fmt.Sprintf("%v", shardHeader.Epoch)
	res += fmt.Sprintf("%v", shardHeader.Timestamp)
	res += shardHeader.TxRoot.String()
	res += shardHeader.ShardTxRoot.String()
	res += shardHeader.CrossTransactionRoot.String()
	res += shardHeader.InstructionsRoot.String()
	res += shardHeader.CommitteeRoot.String()
	res += shardHeader.PendingValidatorRoot.String()
	res += shardHeader.BeaconHash.String()
	res += shardHeader.StakingTxRoot.String()
	res += fmt.Sprintf("%v", shardHeader.BeaconHeight)
	tokenIDs := make([]common.Hash, 0)
	for tokenID, _ := range shardHeader.TotalTxsFee {
		tokenIDs = append(tokenIDs, tokenID)
	}
	sort.Slice(tokenIDs, func(i int, j int) bool {
		res, _ := tokenIDs[i].Cmp(&tokenIDs[j])
		return res == -1
	})

	for _, tokenID := range tokenIDs {
		res += fmt.Sprintf("%v~%v", tokenID.String(), shardHeader.TotalTxsFee[tokenID])
	}
	for _, value := range shardHeader.CrossShardBitMap {
		res += string(value)
	}

	if shardHeader.Version == 2 {
		res += shardHeader.Proposer
		res += fmt.Sprintf("%v", shardHeader.ProposeTime)
	}
	return res
}

func (shardHeader *ShardHeader) MetaHash() common.Hash {
	return common.Keccak256([]byte(shardHeader.String()))
}

func (shardHeader *ShardHeader) Hash() common.Hash {
	// Block header of bridge uses Keccak256 as a hash func to check on Ethereum when relaying blocks
	blkMetaHash := shardHeader.MetaHash()
	blkInstHash := shardHeader.InstructionMerkleRoot
	combined := append(blkMetaHash[:], blkInstHash[:]...)
	return common.Keccak256(combined)
}
