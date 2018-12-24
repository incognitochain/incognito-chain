package blockchain

import (
	"github.com/ninjadotorg/constant/common"
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

	if self.config.LightMode {
		// only store block header
		err := self.StoreBlockHeader(block)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}
		// store tx which relate to account in local wallet
		/*if len(block.Transactions) < 1 {
			Logger.log.Infof("No transaction in this block")
		} else {
			Logger.log.Infof("Number of transaction in this block %+v", len(block.Transactions))
		}
		for _, account := range self.config.Wallet.MasterAccount.Child {
			for index, tx := range block.Transactions {
				err := self.StoreTransactionLightMode(&account.Key.KeySet.PrivateKey, block.Header.ChainID, block.Header.Height, index, tx)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
			}
		}*/
	} else {
		// save full data of block
		err := self.StoreBlock(block)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}

		// store full data of tx tracking(which block and index in block)
		if len(block.Transactions) < 1 {
			Logger.log.Infof("No transaction in this block")
		} else {
			Logger.log.Infof("Number of transaction in this block %+v", len(block.Transactions))
		}
		for index, tx := range block.Transactions {
			err := self.StoreTransactionIndex(tx.Hash(), block.Hash(), index)
			if err != nil {
				Logger.log.Error("ERROR", err, "Transaction in block with hash", blockHash, "and index", index, ":", tx)
				return NewBlockChainError(UnExpectedError, err)
			}
			Logger.log.Infof("Transaction in block with hash", blockHash, "and index", index, ":", tx)
		}
	}
	// TODO: @0xankylosaurus optimize for loop once instead of multiple times ; metadata.process
	// save index of block
	err := self.StoreBlockIndex(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	// fetch data in each tx of block and save into db
	// data: commitments, serialnumbers, snderivator, outputcoins...
	// need to use lightmode flag to check data
	err = self.CreateAndSaveTxViewPointFromBlock(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	// Save loan txs
	err = self.ProcessLoanForBlock(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	// Update utxo reward for dividends
	err = self.UpdateDividendPayout(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	//Update database for vote board
	err = self.UpdateVoteCountBoard(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	//Update amount of token of each holder
	err = self.UpdateVoteTokenHolder(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	// Update database for vote proposal
	err = self.ProcessVoteProposal(block)

	// Process crowdsale tx
	err = self.ProcessCrowdsaleTxs(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
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
		return false, NewBlockChainError(UnExpectedError, err)
	} else {
		return result, nil
	}
}
