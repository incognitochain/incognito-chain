package blockchain

import (
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

	// Store block data
	if self.config.LightMode {
		//
		// **** IN Light Mode running ***
		//
		// if block contain output of account in local wallet: store full block data
		if self.BlockContainAccountLocalWallet(block) {
			// this support database contain a full block data
			err := self.StoreBlock(block)
			if err != nil {
				return NewBlockChainError(UnExpectedError, err)
			}
		} else {
			// else: only store block header
			// because db only store block header
			// -> when we have blockhash, we only get block struct with only blockheader, not body or any more data
			err := self.StoreBlockHeader(block)
			if err != nil {
				return NewBlockChainError(UnExpectedError, err)
			}
		}
	} else {
		// store full data of block
		err := self.StoreBlock(block)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}
	}

	// store full data of tx tracking(with tx hash,  block hash and index in block)
	// in light mode running, any blocks not contain data of account in local wallet (which should only store block header)
	// will not contain any tx in db -> can not get tx by tx hash
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

// BlockContainAccountLocalWallet - checking block which contain any data of account in local wallet
func (self *BlockChain) BlockContainAccountLocalWallet(block *Block) (bool) {
	if self.config.Wallet == nil {
		return false
	}
	if len(block.Transactions) > 0 {
		for _, tx := range block.Transactions {
			switch tx.GetType() {
			case common.TxNormalType, common.TxSalaryType:
				{
					txNormal := tx.(*transaction.Tx)
					if txNormal.Proof != nil {
						for _, out := range txNormal.Proof.OutputCoins {
							if self.config.Wallet.ContainPubKey(out.CoinDetails.PublicKey.Compress()) {
								return true
							}
						}
					}
				}
			case common.TxCustomTokenType:
				txCustomToken := tx.(*transaction.TxCustomToken)
				if txCustomToken.Proof != nil {
					for _, out := range txCustomToken.Proof.OutputCoins {
						if self.config.Wallet.ContainPubKey(out.CoinDetails.PublicKey.Compress()) {
							return true
						}
					}
				}
				if txCustomToken.TxTokenData.Vouts != nil {
					for _, out := range txCustomToken.TxTokenData.Vouts {
						if self.config.Wallet.ContainPubKey(out.PaymentAddress.Pk) {
							return true
						}
					}
				}
			case common.TxCustomTokenPrivacyType:
				{
					txCustomTokenPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
					if txCustomTokenPrivacy.Proof != nil {
						for _, out := range txCustomTokenPrivacy.Proof.OutputCoins {
							if self.config.Wallet.ContainPubKey(out.CoinDetails.PublicKey.Compress()) {
								return true
							}
						}
					}
					if txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.Proof != nil {
						for _, out := range txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.Proof.OutputCoins {
							if self.config.Wallet.ContainPubKey(out.CoinDetails.PublicKey.Compress()) {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}
