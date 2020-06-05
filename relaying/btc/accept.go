package btcrelaying

import (
	"fmt"

	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btcutil"
)

func (b *BlockChain) maybeAcceptBlockV2(block *btcutil.Block, flags BehaviorFlags) (bool, error) {
	// The height of this block is one more than the referenced previous
	// block.
	prevHash := &block.MsgBlock().Header.PrevBlock
	prevNode := b.index.LookupNode(prevHash)
	if prevNode == nil {
		str := fmt.Sprintf("previous block %s is unknown", prevHash)
		return false, ruleError(ErrPreviousBlockUnknown, str)
	} else if b.index.NodeStatus(prevNode).KnownInvalid() {
		str := fmt.Sprintf("previous block %s is known to be invalid", prevHash)
		return false, ruleError(ErrInvalidAncestorBlock, str)
	}

	blockHeight := prevNode.height + 1
	block.SetHeight(blockHeight)

	// The block must pass all of the validation rules which depend on the
	// position of the block within the block chain.
	err := b.checkBlockContext(block, prevNode, flags)
	if err != nil {
		return false, err
	}

	// Insert the block into the database if it's not already there.  Even
	// though it is possible the block will ultimately fail to connect, it
	// has already passed all proof-of-work and validity tests which means
	// it would be prohibitively expensive for an attacker to fill up the
	// disk with a bunch of blocks that fail to connect.  This is necessary
	// since it allows block download to be decoupled from the much more
	// expensive connection logic.  It also has some other nice properties
	// such as making blocks that never become part of the main chain or
	// blocks that fail to connect available for further analysis.
	err = b.db.Update(func(dbTx database.Tx) error {
		return dbStoreBlock(dbTx, block)
	})
	if err != nil {
		return false, err
	}

	// Create a new block node for the block and add it to the node index. Even
	// if the block ultimately gets connected to the main chain, it starts out
	// on a side chain.
	blockHeader := &block.MsgBlock().Header
	newNode := newBlockNode(blockHeader, prevNode, b.genesisBlkHeight)
	newNode.status = statusDataStored

	b.index.AddNode(newNode)
	err = b.index.flushToDB()
	if err != nil {
		return false, err
	}

	// Connect the passed block to the chain while respecting proper chain
	// selection according to the chain with the most proof of work.  This
	// also handles validation of the transaction scripts.
	isMainChain, err := b.connectBestChainV2(newNode, block, flags)
	if err != nil {
		return false, err
	}

	// Notify the caller that the new block was accepted into the block
	// chain.  The caller would typically want to react by relaying the
	// inventory to other peers.
	b.chainLock.Unlock()
	b.sendNotification(NTBlockAccepted, block)
	b.chainLock.Lock()

	return isMainChain, nil
}
