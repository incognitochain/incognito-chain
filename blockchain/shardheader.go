package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

/*
	-TxRoot and MerkleRootShard: make from transaction
	-Validator Root is root hash of current committee in beststate
	-PendingValidator Root is root hash of pending validator in beststate
*/
type ShardHeader struct {
	ProducerAddress privacy.PaymentAddress
	ShardID         byte
	Version         int
	PrevBlockHash   common.Hash
	Height          uint64
	Round           int
	Epoch           uint64
	Timestamp       int64

	TxRoot               common.Hash //Transaction root created from transaction in shard
	ShardTxRoot          common.Hash //Output root created for other shard
	CrossTransactionRoot common.Hash //Transaction root created from transaction of micro shard to shard block (from other shard)
	InstructionsRoot     common.Hash //Actions root created from Instructions and Metadata of transaction
	CommitteeRoot        common.Hash
	PendingValidatorRoot common.Hash

	CrossShards []byte // CrossShards for beacon

	BeaconHeight uint64 //Beacon check point
	BeaconHash   common.Hash

	TotalTxsFee uint64

	// Merkle root of all instructions (using Keccak256 hash func) to relay to Ethreum
	// This obsoletes InstructionRoot but for simplicity, we keep it for now
	InstructionMerkleRoot common.Hash
}

func (shardHeader *ShardHeader) String() string {
	res := common.EmptyString
	res += shardHeader.ProducerAddress.String()
	res += string(shardHeader.ShardID)
	res += fmt.Sprintf("%v", shardHeader.Version)
	res += shardHeader.PrevBlockHash.String()
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
	res += fmt.Sprintf("%v", shardHeader.BeaconHeight)
	res += fmt.Sprintf("%v", shardHeader.TotalTxsFee)
	for _, value := range shardHeader.CrossShards {
		res += string(value)
	}
	return res
}

func (shardHeader *ShardHeader) MetaHash() common.Hash {
	return common.Keccak256([]byte(shardHeader.String()))
}

func (shardHeader *ShardHeader) Hash() common.Hash {
	// TODO(@0xbunyip): modify only bridge shard
	// Block header of bridge uses Keccak256 as a hash func to check on Ethereum when relaying blocks
	blkMetaHash := shardHeader.MetaHash()
	blkInstHash := shardHeader.InstructionMerkleRoot
	combined := append(blkMetaHash[:], blkInstHash[:]...)
	return common.Keccak256(combined)
}
