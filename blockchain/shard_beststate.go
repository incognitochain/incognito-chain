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
	BestBlockHash common.Hash // The hash of the block.
	BestBlock     *ShardBlock // The block.

	Height uint64

	ShardCommittee        []string
	ShardPendingValidator []string
	ShardProposerIdx      int

	//@Hung: these information are no needs!!!
	NumTxns   uint64 // The number of txns in the block.
	TotalTxns uint64 // The total number of txns in the chain.
}

// func (self *BestStateShard) Init(block *ShardBlock) {

// 	self.BestBlock = block
// 	self.BestBlockHash = *block.Hash()

// 	// self.  += uint64(len(block.Transactions))
// 	self.NumTxns = uint64(len(block.Body.Transactions))
// 	self.TotalTxns = self.NumTxns
// }

func (self *BestStateShard) Update(block *ShardBlock) error {
	self.BestBlock = block
	self.BestBlockHash = *block.Hash()
	self.Height = block.Header.Height
	self.TotalTxns += uint64(len(block.Body.Transactions))
	self.NumTxns = uint64(len(block.Body.Transactions))
	return nil
}

// Get role of a public key base on best state shard
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
