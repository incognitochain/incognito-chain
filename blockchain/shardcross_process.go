package blockchain

import (
	"errors"
)

/*
	Verify CrossShard Block
	- Agg Signature
	- MerklePath
*/
func (block *CrossShardBlock) VerifyCrossShardBlock(committees []string) error {
	if err := ValidateAggSignature(block.ValidatorsIdx, committees, block.AggregatedSig, block.R, block.Hash()); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := VerifyCrossShardBlockUTXO(block, block.MerklePathShard); !ok {
		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
	}
	return nil
}
