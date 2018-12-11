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
	Beacon *BestStateBeacon         //Beacon node & shard node allow access this
	Shards map[byte]*BestStateShard //Only shard node allow to access this
}

type BestStateBeacon struct {
	BestBlockHash common.Hash // The hash of the block.
	BestBlock     *BlockV2    // The block.
	// PendingValidator []string    //pending validators pubkey in base58
}
type BestStateShard struct {
	BestBlockHash common.Hash // The hash of the block.
	BestBlock     *BlockV2    // The block.

	NumTxns   uint64 // The number of txns in the block.
	TotalTxns uint64 // The total number of txns in the chain.
	// PendingValidator []string //pending validators pubkey in base58
}

/*
Init create a beststate data from block and commitment tree
*/
// #1 - block
// #2 - commitment merkle tree
func (self *BestStateNew) Init(shardsBlock map[byte]*BlockV2, beaconBlock *BlockV2) {
	//shards beststate
	for shardID, block := range shardsBlock {
		bestShardBlock := &BestStateShard{
			BestBlockHash: block.Hash(),
			BestBlock:     block,
			TotalTxns:     uint64(len(block.Body.(*BlockBodyShard).Transactions)),
			NumTxns:       uint64(len(block.Body.(*BlockBodyShard).Transactions)),
		}
		self.Shards[shardID] = bestShardBlock
	}

	//beacon beststate
	self.Beacon = &BestStateBeacon{
		BestBlockHash: beaconBlock.Hash(),
		BestBlock:     beaconBlock,
	}

}

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

/*func (self *BestState) UpdateLoanIDs(block *Block) error {
	for _, blockTx := range block.Transactions {
		if blockTx.GetType() == common.TxLoanRequest {
			tx, ok := blockTx.(*transaction.TxLoanRequest)
			if ok == false {
				return NewBlockChainError(UnExpectedError, fmt.Errorf("Transaction in block not valid, expected TxLoanRequest"))
			}

			self.LoanIDs = append(self.LoanIDs, tx.LoanID)
		}
	}
	return nil
}*/
