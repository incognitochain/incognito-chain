package blockchain

import (
	"errors"
)

/*
	Verify CrossShard Block
	- Agg Signature
	- MerklePath
*/
func (crossShardBlock *CrossShardBlock) VerifyCrossShardBlock(committees []string) error {
	if err := ValidateAggSignature(crossShardBlock.ValidatorsIdx, committees, crossShardBlock.AggregatedSig, crossShardBlock.R, crossShardBlock.Hash()); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := VerifyCrossShardBlockUTXO(crossShardBlock, crossShardBlock.MerklePathShard); !ok {
		return NewBlockChainError(HashError, errors.New("verify Merkle Path Shard"))
	}
	return nil
}

func (self *CrossShardBlock) ShouldStoreBlock() bool {
	// verify block aggregation
	// verify with best cross shard

	return true
}

// get next cross shard height from current cross shard height (of cross shard from fromShardID to toShardID)
func GetNextCrossShardHeight(fromShardID, toShardID byte, currentCrossShardHeight uint64) uint64 {
	// asking database for the next cross shard height
	// e.g at height 30 of shard 1, there are cross shard to shard 2, next at height 33 of shard 1, there are cross shard to shard 2
	// ask db: cross-f-1-t-2-30 : from shard 1 to shard 2 with current cross shard height = 30
	// should return value (33 + hash)
	return 0
}
