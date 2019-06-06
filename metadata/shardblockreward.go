package metadata

import (
	// "errors"

	"github.com/constant-money/constant-chain/database"
)

type ShardBlockRewardMeta struct {
	MetadataBase
}

func NewShardBlockRewardMeta() *ShardBlockRewardMeta {
	metadataBase := MetadataBase{
		Type: ShardBlockReward,
	}
	return &ShardBlockRewardMeta{
		MetadataBase: metadataBase,
	}
}

func (shardBlockRewardMeta *ShardBlockRewardMeta) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with request tx (via RequestedTxID) in current block
	return false, nil
}

func (shardBlockRewardMeta *ShardBlockRewardMeta) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (shardBlockRewardMeta *ShardBlockRewardMeta) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	// The validation just need to check at tx level, so returning true here
	return true, true, nil
}
