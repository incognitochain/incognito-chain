package blockchain

import (
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

/*
	-MerkleRoot and MerkleRootShard: make from transaction
	-Validator Root is root hash of current committee in beststate
	-PendingValidator Root is root hash of pending validator in beststate
*/
type ShardHeader struct {
	Producer      string
	ShardID       byte
	Version       int
	PrevBlockHash common.Hash
	Height        uint64
	Epoch         uint64
	Timestamp     int64
	SalaryFund    uint64
	//Transaction root created from transaction in shard
	TxRoot common.Hash
	//Transaction root created from transaction of micro shard to shard block (from other shard)
	ShardTxRoot common.Hash
	//Output root created for other shard
	CrossOutputCoinRoot common.Hash
	//Actions root created from Instructions and Metadata of transaction
	ActionsRoot          common.Hash `description: verify when update`
	CommitteeRoot        common.Hash `description: verify post processing`
	PendingValidatorRoot common.Hash `description: verify post processing`
	// CrossShards for beacon
	CrossShards []byte
	//Beacon check point
	BeaconHeight uint64
	BeaconHash   common.Hash
}

func (self ShardHeader) Hash() common.Hash {
	record := common.EmptyString

	// add data from header
	record += strconv.FormatInt(self.Timestamp, 10) +
		string(self.ShardID) +
		self.PrevBlockHash.String() + self.Producer
	//TODO: add ALL information from header
	return common.DoubleHashH([]byte(record))
}
