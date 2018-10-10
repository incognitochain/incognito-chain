package blockchain

import (
	"github.com/ninjadotorg/cash-prototype/common"
)

// ProcessBlock is the main workhorse for handling insertion of new blocks into
// the block chain.  It includes functionality such as rejecting duplicate
// blocks, ensuring blocks follow all rules, orphan handling, and insertion into
// the block chain along with best chain selection and reorganization.
//
// When no errors occurred during processing, the first return value indicates
// whether or not the block is on the main chain and the second indicates
// whether or not the block is an orphan.
//
// This function is safe for concurrent access.
// Return
// isMainChain
// isOrphan
// err
func (self *BlockChain) ProcessBlock(block *Block) (bool, bool, error) {
	self.chainLock.Lock()
	defer self.chainLock.Unlock()

	blockHash := block.Hash()
	Logger.log.Infof("Processing block %v", blockHash)

	// The block must not already exist in the main chain or side chains.
	exists, err := self.BlockExists(blockHash)
	//if err != nil {
	//	return false, false, err
	//}
	if exists {
		Logger.log.Infof("already have block %v", blockHash)
		return false, false, nil
	}
	// TODO: privacy consensus checks:
	// - Only first tx is coinbase (only one desc and reward >= 0)
	// - All js desc's proofs are valid
	// - Coinbase reward == block reward + sum(fee of all tx)

	// TODO check more
	// check orphan blocks
	// check checkpoint if using checkpoint to prevent sidechain
	// Perform preliminary sanity checks on the block and its transactions.
	// ... TODO TODO TODO by consensus

	// The block has passed all context independent checks and appears sane
	// enough to potentially accept it into the block chain.
	isMainChain, err := self.maybeAcceptBlock(block)
	if err != nil {
		return false, false, err
	}

	Logger.log.Infof("Accepted block %s", blockHash.String())

	return isMainChain, false, nil
}

// blockExists determines whether a block with the given hash exists either in
// the main chain or any side chains.
//
// This function is safe for concurrent access.
func (self *BlockChain) blockExists(hash *common.Hash) (bool, error) {
	result, err := self.config.DataBase.HasBlock(hash)
	if err != nil {
		return false, err
	}
	return result, err
}
