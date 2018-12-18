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
type BestStateNew struct {
	Beacon *BestStateBeacon
	Shard  map[byte]*BestStateShard
}

type BestStateBeacon struct {
	BestBlockHash common.Hash // The hash of the block.
	BestBlock     *BlockV2    // The block.
	BestShardHash []common.Hash

	BeaconHeight           uint64
	BeaconCandidate        []string
	BeaconPendingCandidate []string
	ShardValidator         map[byte][]string
	ShardPendingValidator  map[byte][]string

	UnassignBeaconCandidate []string
	UnassignShardCandidate  []string

	CurrentRandomNumber int64
	NextRandomNumber    int64

	Params map[string]string
}

type BestStateShard struct {
	BestBlockHash common.Hash // The hash of the block.
	BestBlock     *BlockV2    // The block.

	NumTxns   uint64 // The number of txns in the block.
	TotalTxns uint64 // The total number of txns in the chain.
}

/*
Init create a beststate data from block and commitment tree
*/
// #1 - block
// #2 - commitment merkle tree
func (self *BestStateBeacon) Init(block *BlockV2) {
	self.BestBlock = block
	self.BestBlockHash = *block.Hash()
}

func (self *BestStateShard) Init(block *BlockV2) {

	self.BestBlock = block
	self.BestBlockHash = *block.Hash()

	// self.  += uint64(len(block.Transactions))
	self.NumTxns = uint64(len(block.Body.(*BlockBodyShard).Transactions))
	self.TotalTxns = self.NumTxns
}

func (self *BestStateBeacon) Update(block *BlockV2) error {

	self.BestBlock = block
	self.BestBlockHash = *block.Hash()

	return nil
}

func (self *BestStateShard) Update(block *BlockV2) error {

	self.BestBlock = block
	self.BestBlockHash = *block.Hash()

	self.TotalTxns += uint64(len(block.Body.(*BlockBodyShard).Transactions))
	self.NumTxns = uint64(len(block.Body.(*BlockBodyShard).Transactions))
	return nil
}
