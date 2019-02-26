package blockchain

import (
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy"
)

/*
	-MerkleRoot and MerkleRootShard: make from transaction
	-Validator Root is root hash of current committee in beststate
	-PendingValidator Root is root hash of pending validator in beststate
*/
type ShardHeader struct {
	ProducerAddress *privacy.PaymentAddress
	Producer        string
	ShardID         byte
	Version         int
	PrevBlockHash   common.Hash
	Height          uint64
	Round           int
	Epoch           uint64
	Timestamp       int64
	SalaryFund      uint64
	//Transaction root created from transaction in shard
	TxRoot common.Hash
	//Output root created for other shard
	ShardTxRoot common.Hash
	//Transaction root created from transaction of micro shard to shard block (from other shard)
	CrossOutputCoinRoot common.Hash
	//Actions root created from Instructions and Metadata of transaction
	InstructionsRoot     common.Hash
	CommitteeRoot        common.Hash `description: verify post processing`
	PendingValidatorRoot common.Hash `description: verify post processing`
	// CrossShards for beacon
	CrossShards []byte
	//Beacon check point
	BeaconHeight uint64
	BeaconHash   common.Hash
}

func (shardHeader ShardHeader) Hash() common.Hash {
	record := common.EmptyString
	// crossShardHash, _ := common.Hash{}.NewHash(shardHeader.CrossShards)
	// add data from header
	record += strconv.FormatInt(shardHeader.Timestamp, 10) +
		shardHeader.Producer +
		string(shardHeader.ShardID) +
		strconv.Itoa(shardHeader.Version)
		// TODO: Uncomment this when finish genesis shard block
		// shardHeader.PrevBlockHash.String() +
		// strconv.Itoa(int(shardHeader.Height)) +
		// strconv.Itoa(int(shardHeader.Epoch)) +
		// strconv.Itoa(int(shardHeader.Timestamp)) +
		// strconv.Itoa(int(shardHeader.SalaryFund)) +
		// shardHeader.TxRoot.String() +
		// shardHeader.ShardTxRoot.String() +
		// shardHeader.CrossOutputCoinRoot.String() +
		// shardHeader.ActionsRoot.String() +
		// shardHeader.CommitteeRoot.String() +
		// shardHeader.PendingValidatorRoot.String() +
		// shardHeader.BeaconHash.String() +
		// crossShardHash.String() +
		// strconv.Itoa(int(shardHeader.BeaconHeight)) +
		// shardHeader.ProducerAddress.String()
	return common.DoubleHashH([]byte(record))
}
