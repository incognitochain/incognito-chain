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
func (self *BestStateShard) Hash() common.Hash {
	res := []byte{}
	res = append(res, self.PrevShardBlockHash.GetBytes()...)
	res = append(res, self.BestShardBlockHash.GetBytes()...)
	res = append(res, self.BestBeaconHash.GetBytes()...)
	res = append(res, self.BestShardBlock.Hash().GetBytes()...)
	res = append(res, byte(self.ShardHeight))
	res = append(res, byte(self.BeaconHeight))
	for _, value := range self.ShardCommittee {
		res = append(res, []byte(value)...)
	}
	for _, value := range self.ShardPendingValidator {
		res = append(res, []byte(value)...)
	}
	res = append(res, byte(self.ShardProposerIdx))
	res = append(res, byte(self.NumTxns))
	res = append(res, byte(self.TotalTxns))
	return common.DoubleHashH(res)
}
func (self *BestStateShard) GetPubkeyRole(pubkey string) string {
	fmt.Println("Shard BestState/ BEST STATE", self)
	found := common.IndexOfStr(pubkey, self.ShardCommittee)
	fmt.Println("Shard BestState/ Get Public Key Role, Found IN Shard COMMITTEES", found)
	if found > -1 {
		tmpID := (self.ShardProposerIdx + 1) % len(self.ShardCommittee)
		if found == tmpID {
			fmt.Println("Shard BestState/ Get Public Key Role, ROLE", common.SHARD_PROPOSER_ROLE)
			return common.SHARD_PROPOSER_ROLE
		} else {
			fmt.Println("Shard BestState/ Get Public Key Role, ROLE", common.SHARD_VALIDATOR_ROLE)
			return common.SHARD_VALIDATOR_ROLE
		}

	}

	found = common.IndexOfStr(pubkey, self.ShardPendingValidator)
	if found > -1 {
		fmt.Println("Shard BestState/ Get Public Key Role, ROLE", common.SHARD_PENDING_ROLE)
		return common.SHARD_PENDING_ROLE
	}

	return common.EmptyString
}
