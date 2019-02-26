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
