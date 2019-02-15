package blockchain

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
)

// BestState houses information about the current best block and other info
// related to the state of the main chain as it exists from the point of view of
// the current best block.
//
// The BestSnapshot method can be used to obtain access to this information
// in a concurrent safe manner and the data will not be changed out from under
// the caller when chain state changes occur as the function name implies.
// However, the returned snapshot must be treated as immutable since it is
// shared by all callers.

type BestStateShard struct {
	PrevShardBlockHash    common.Hash `json:"PrevShardBlockHash,omitempty"`
	BestShardBlockHash    common.Hash `json:"BestShardBlockHash,omitempty"` // hash of block.
	BestBeaconHash        common.Hash `json:"BestBeaconHash,omitempty"`
	BestShardBlock        *ShardBlock `json:"BestShardBlock,omitempty"` // block data
	ShardHeight           uint64      `json:"ShardHeight,omitempty"`
	BeaconHeight          uint64      `json:"BeaconHeight,omitempty"`
	Epoch                 uint64      `json:"Epoch,omitempty"`
	ShardCommittee        []string    `json:"ShardCommittee,omitempty"`
	ShardPendingValidator []string    `json:"ShardPendingValidator,omitempty"`
	ShardProposerIdx      int         `json:"ShardProposerIdx,omitempty"`
	// Best cross shard block by height
	BestCrossShard map[byte]uint64 `json:"BestCrossShard,omitempty"`
	//TODO: verify if these information are needed or not
	NumTxns   uint64 `json:"NumTxns,omitempty"`   // The number of txns in the block.
	TotalTxns uint64 `json:"TotalTxns,omitempty"` // The total number of txns in the chain.
}

// Get role of a public key base on best state shard
func (bestStateShard *BestStateShard) Hash() common.Hash {
	res := []byte{}
	res = append(res, bestStateShard.PrevShardBlockHash.GetBytes()...)
	res = append(res, bestStateShard.BestShardBlockHash.GetBytes()...)
	res = append(res, bestStateShard.BestBeaconHash.GetBytes()...)
	res = append(res, bestStateShard.BestShardBlock.Hash().GetBytes()...)
	res = append(res, byte(bestStateShard.ShardHeight))
	res = append(res, byte(bestStateShard.BeaconHeight))
	for _, value := range bestStateShard.ShardCommittee {
		res = append(res, []byte(value)...)
	}
	for _, value := range bestStateShard.ShardPendingValidator {
		res = append(res, []byte(value)...)
	}
	res = append(res, byte(bestStateShard.ShardProposerIdx))
	res = append(res, byte(bestStateShard.NumTxns))
	res = append(res, byte(bestStateShard.TotalTxns))
	return common.DoubleHashH(res)
}
func (bestStateShard *BestStateShard) GetPubkeyRole(pubkey string) string {
	fmt.Println("Shard BestState/ BEST STATE", bestStateShard)
	found := common.IndexOfStr(pubkey, bestStateShard.ShardCommittee)
	fmt.Println("Shard BestState/ Get Public Key Role, Found IN Shard COMMITTEES", found)
	if found > -1 {
		tmpID := (bestStateShard.ShardProposerIdx + 1) % len(bestStateShard.ShardCommittee)
		if found == tmpID {
			fmt.Println("Shard BestState/ Get Public Key Role, ROLE", common.SHARD_PROPOSER_ROLE)
			return common.SHARD_PROPOSER_ROLE
		} else {
			fmt.Println("Shard BestState/ Get Public Key Role, ROLE", common.SHARD_VALIDATOR_ROLE)
			return common.SHARD_VALIDATOR_ROLE
		}

	}

	found = common.IndexOfStr(pubkey, bestStateShard.ShardPendingValidator)
	if found > -1 {
		fmt.Println("Shard BestState/ Get Public Key Role, ROLE", common.SHARD_PENDING_ROLE)
		return common.SHARD_PENDING_ROLE
	}

	return common.EmptyString
}
