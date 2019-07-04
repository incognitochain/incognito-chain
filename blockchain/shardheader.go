package blockchain

import (
	"fmt"
	"sort"

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

	TotalTxsFee map[common.Hash]uint64
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
	for _, value := range shardHeader.CrossShards {
		res += string(value)
	}
	return res
}

func (shardHeader *ShardHeader) Hash() common.Hash {
	return common.HashH([]byte(shardHeader.String()))
}
