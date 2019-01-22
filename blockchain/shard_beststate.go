package blockchain

import (
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
	PrevShardBlockHash    common.Hash
	BestShardBlockHash    common.Hash // hash of block.
	BestBeaconHash        common.Hash
	BestShardBlock        *ShardBlock // block data
	ShardHeight           uint64
	BeaconHeight          uint64
	Epoch                 uint64
	ShardCommittee        []string
	ShardPendingValidator []string
	ShardProposerIdx      int
	//TODO: verify if these information are needed or not
	NumTxns   uint64 // The number of txns in the block.
	TotalTxns uint64 // The total number of txns in the chain.
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

	found := common.IndexOfStr(pubkey, self.ShardCommittee)
	if found > -1 {
		tmpID := (self.ShardProposerIdx + 1) % len(self.ShardCommittee)
		if found == tmpID {
			return "shard-proposer"
		} else {
			return "shard-validator"
		}

	}

	found = common.IndexOfStr(pubkey, self.ShardPendingValidator)
	if found > -1 {
		return "shard-pending"
	}

	return ""
}
