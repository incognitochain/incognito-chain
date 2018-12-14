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
	BestBlockHash common.Hash // The hash of the block.
	BestBlock     *BlockV2    // The hash of the block.

	Height           int32    // The height of the block.
	NumTxns          uint64   // The number of txns in the block.
	TotalTxns        uint64   // The total number of txns in the chain.
	PendingValidator []string //pending validators pubkey in base58
}

type BestStateBeacon struct {
	BestBlockHash common.Hash // The hash of the block.
	BestBlock     *BlockV2    // The block.
	BestShardHash []common.Hash

	BeaconHeight           uint64
	BeaconCandidate        []string
	BeaconPendingCandidate []string
	ShardCandidate         map[byte][]string
	ShardPendingCandidate  map[byte][]string

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
// func (self *BestStateNew) Init(block *BlockV2) {
// 	bestBlockHash := block.Hash()
// 	self.BestBlock = block
// 	self.BestBlockHash = bestBlockHash

// 	// self.  += uint64(len(block.Transactions))
// 	self.NumTxns = uint64(len(block.Transactions))
// 	self.Height = block.Header.Height
// 	if self.Candidates == nil {
// 		self.Candidates = make(map[string]CommitteeCandidateInfo)
// 	}

// 	if self.LoanIDs == nil {
// 		self.LoanIDs = make([][]byte, 0)
// 	}
// }

// func (self *BestStateNew) Update(block *BlockV2) error {
// 	//tree := self.CmTree
// 	//err := UpdateMerkleTreeForBlock(tree, block)
// 	//if err != nil {
// 	//	return NewBlockChainError(UnExpectedError, err)
// 	//}
// 	bestBlockHash := block.Hash()
// 	self.BestBlock = block
// 	self.BestBlockHash = bestBlockHash
// 	//self.CmTree = tree

// 	self.TotalTxns += uint64(len(block.Transactions))
// 	self.NumTxns = uint64(len(block.Transactions))
// 	self.Height = block.Header.Height
// 	if self.Candidates == nil {
// 		self.Candidates = make(map[string]CommitteeCandidateInfo)
// 	}

// 	// Update list of loan ids
// 	// TODO
// 	/*err = self.UpdateLoanIDs(block)
// 	if err != nil {
// 		return NewBlockChainError(UnExpectedError, err)
// 	}*/
// 	return nil
// }

// func (self *BestStateNew) RemoveCandidate(producerPbk string) {
// 	_, ok := self.Candidates[producerPbk]
// 	if ok {
// 		delete(self.Candidates, producerPbk)
// 	}
// }
