package blockchain

import (
	"fmt"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
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
	if self.config.Light {
		Logger.log.Infof("Storing Block Header of Block %+v", blockHash)
		err := self.StoreBlockHeader(block)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}

		Logger.log.Infof("Fetch Block %+v to get unspent tx of all accoutns in wallet", blockHash)
		nullifiersInDb := make([][]byte, 0)
		chainId := block.Header.ChainID
		txViewPoint, err := self.FetchTxViewPoint(chainId)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}
		nullifiersInDb = append(nullifiersInDb, txViewPoint.listNullifiers...)
		for _, account := range self.config.Wallet.MasterAccount.Child {
			unspentTxs, err1 := self.GetListUnspentTxByKeysetInBlock(&account.Key.KeySet, block, nullifiersInDb, true)
			if err1 != nil {
				return NewBlockChainError(UnExpectedError, err1)
			}

			for chainId, txs := range unspentTxs {
				for _, unspent := range txs {
					var txIndex = -1
					// Iterate to get Tx index of transaction in a block
					for i, _ := range block.Transactions {
						txHash := unspent.Hash().String()
						blockTxHash := block.Transactions[i].(*transaction.Tx).Hash().String()
						if strings.Compare(txHash, blockTxHash) == 0 {
							txIndex = i
							fmt.Println("Found Transaction i", unspent.Hash(), i)
							break
						}
					}
					if txIndex == -1 {
						return NewBlockChainError(UnExpectedError, err)
					}
					err := self.StoreUnspentTransactionLightMode(&account.Key.KeySet.PrivateKey, chainId, block.Header.Height, txIndex, &unspent)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err)
					}
				}
			}
		}
	} else {
		err := self.StoreBlock(block)
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
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}
	}
	// TODO: @0xankylosaurus optimize for loop once instead of multiple times
	// save index of block
	err := self.StoreBlockIndex(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	// fetch nullifiers and commitments(utxo) from block and save
	err = self.CreateAndSaveTxViewPoint(block)
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

	// Update vote DCB Vote Count
	err = self.UpdateVoteCountBoard(block)
	//Update database for vote board
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	err = self.UpdateVoteCountBoard(block)
	//Update database for vote board
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

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
