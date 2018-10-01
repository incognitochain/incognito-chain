package blockchain

import "fmt"

/*
maybeAcceptBlock potentially accepts a block into the block chain and, if
accepted, returns whether or not it is on the main chain.  It performs
several validation checks which depend on its position within the block chain
before adding it.  The block is expected to have already gone through
ProcessBlock before calling this function with it.

The flags are also passed to checkBlockContext and connectBestChain.  See
their documentation for how the flags modify their behavior.

This function MUST be called with the chain state lock held (for writes).
*/
// #1 - block - candidate block which be needed to check
func (self *BlockChain) maybeAcceptBlock(block *Block) (bool, error) {
	// TODO
	// The height of this block is one more than the referenced previous
	// block.
	prevHash := &block.Header.PrevBlockHash
	prevNode, err := self.GetBlockByBlockHash(prevHash)
	if prevNode == nil {
		str := fmt.Sprintf("previous block %s is unknown", prevHash)
		return false, ruleError(ErrPreviousBlockUnknown, str)
	}

	//blockHeight := prevNode.Height + 1
	//block.Height = blockHeight

	// Insert the block into the database if it's not already there.  Even
	// though it is possible the block will ultimately fail to connect, it
	// has already passed all proof-of-work and validity tests which means
	// it would be prohibitively expensive for an attacker to fill up the
	// disk with a bunch of blocks that fail to connect.  This is necessary
	// since it allows block download to be decoupled from the much more
	// expensive connection logic.  It also has some other nice properties
	// such as making blocks that never become part of the main chain or
	// blocks that fail to connect available for further analysis.
	err = self.StoreBlock(block)
	if err != nil {
		return false, err
	}

	// fetch nullifiers and commitments(utxo) from block and save
	isMainChain, err := self.connectBestChain(block)
	if err != nil {
		return false, err
	}

	return isMainChain, nil
}
