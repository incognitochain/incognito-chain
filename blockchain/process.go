package blockchain

import (
	"github.com/ninjadotorg/cash/common"
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
func (self *BlockChain) ConnectBlock(block *Block) error {
	self.chainLock.Lock()
	defer self.chainLock.Unlock()

	blockHash := block.Hash().String()
	Logger.log.Infof("Processing block %+v", blockHash)

	// Insert the block into the database if it's not already there.  Even
	// though it is possible the block will ultimately fail to connect, it
	// has already passed all proof-of-work and validity tests which means
	// it would be prohibitively expensive for an attacker to fill up the
	// disk with a bunch of blocks that fail to connect.  This is necessary
	// since it allows block download to be decoupled from the much more
	// expensive connection logic.  It also has some other nice properties
	// such as making blocks that never become part of the main chain or
	// blocks that fail to connect available for further analysis.
	err := self.StoreBlock(block)
	if err != nil {
		return err
	}
	// save index of block
	err = self.StoreBlockIndex(block)
	if err != nil {
		return err
	}
	// fetch nullifiers and commitments(utxo) from block and save
	err = self.CreateTxViewPoint(block)
	if err != nil {
		return err
	}

	Logger.log.Infof("Accepted block %s", blockHash)

	return nil
}

// blockExists determines whether a block with the given hash exists either in
// the main chain or any side chains.
//
// This function is safe for concurrent access.
func (self *BlockChain) BlockExists(hash *common.Hash) (bool, error) {
	result, err := self.config.DataBase.HasBlock(hash)
	if err != nil {
		return false, err
	}
	return result, err
}
