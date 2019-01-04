package blockchain

import (
	"encoding/json"
	"errors"
	"github.com/ninjadotorg/constant/metadata"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy"
)

//@tamle
func (self *BlockChain) MayBeAcceptShardBlock(block *ShardBlock) error {
	err := self.ValidateShardBlockSignature(block)
	if err != nil {
		return err
	}
	//TODO: push to tmp database

	return nil
}

func (self *BlockChain) ValidateShardBlockSignature(block *ShardBlock) error {
	// get shard id
	shardID := block.Header.ShardID
	// get best state shard committee corresponding to shardID
	bestStateShardCommittee := self.BestState.Shard[shardID].ShardCommittee

	pubKeys := []*privacy.PublicKey{}
	for _, index := range block.ValidatorsIdx {
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(bestStateShardCommittee[index])
		if err != nil {
			return errors.New("Error in convert Public key from string to byte")
		}
		pubKey := privacy.PublicKey{}
		pubKey = pubkeyBytes
		pubKeys = append(pubKeys, &pubKey)
	}

	aggSig, _, err := base58.Base58Check{}.Decode(block.AggregatedSig)
	if err != nil {
		return errors.New("Error in convert aggregated signature from string to byte")
	}
	schnMultiSig := &privacy.SchnMultiSig{}
	schnMultiSig.SetBytes(aggSig)
	blockHash := block.Header.Hash()
	if schnMultiSig.VerifyMultiSig(blockHash.GetBytes(), pubKeys, nil, nil) == false {
		return errors.New("Invalid Agg signature")
	}

	return nil
}

func (self *BlockChain) ProcessShardBlockListenner(block *ShardBlock) error {
	//TODO: get next block height
	//TODO: get block (next block height)from tmp database
	//TODO: chose 1 valid block (ValidateShardBlockSignature)
	//TODO: InsertShardBlock
	return nil
}

func (self *BlockChain) InsertShardBlock(block *ShardBlock) error {
	blockHash := block.Hash().String()
	Logger.log.Infof("Processing block %+v", blockHash)

	err := self.ValidateShardBlockSignature(block)
	if err != nil {
		return err
	}
	self.BestState.Shard[block.Header.ShardID].Update(block)
	self.StoreShardBlockIndex(block)

	// Process transaction db
	if len(block.Body.Transactions) < 1 {
		Logger.log.Infof("No transaction in this block")
	} else {
		Logger.log.Infof("Number of transaction in this block %d", len(block.Body.Transactions))
	}

	for index, tx := range block.Body.Transactions {
		if tx.GetType() == common.TxCustomTokenPrivacyType {
			_ = 1
			//TODO: do what???
		}

		err := self.StoreTransactionIndex(tx.Hash(), block.Hash(), index)
		if err != nil {
			Logger.log.Error("ERROR", err, "Transaction in block with hash", blockHash, "and index", index, ":", tx)
			return NewBlockChainError(UnExpectedError, err)
		}
		Logger.log.Infof("Transaction in block with hash", blockHash, "and index", index, ":", tx)
	}

	return nil
}

func (self *BlockChain) ConnectBlock(block *ShardBlock) error {
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
	// if self.config.Light {
	/*Logger.log.Infof("Storing Block Header of Block %+v", blockHash)
	err := self.StoreShardBlockHeader(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	Logger.log.Infof("Fetch Block %+v to get unspent tx of all accoutns in wallet", blockHash)
	for _, account := range self.config.Wallet.MasterAccount.Child {
		unspentTxs, err1 := self.GetListUnspentTxByKeysetInBlock(&account.Key.KeySet, block.Header.shardID, block.Transactions, true)
		if err1 != nil {
			return NewBlockChainError(UnExpectedError, err1)
		}

		for shardID, txs := range unspentTxs {
			for _, unspent := range txs {
				var txIndex = -1
				// Iterate to get TxNormal index of transaction in a block
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
				err := self.StoreUnspentTransactionLightMode(&account.Key.KeySet.PrivateKey, shardID, block.Header.Height, txIndex, &unspent)
				if err != nil {
					return NewBlockChainError(UnExpectedError, err)
				}
			}
		}
	}*/
	// } else {
	err := self.StoreShardBlock(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	if len(block.Body.Transactions) < 1 {
		Logger.log.Infof("No transaction in this block")
	} else {
		Logger.log.Infof("Number of transaction in this block %d", len(block.Body.Transactions))
	}

	for index, tx := range block.Body.Transactions {
		err := self.StoreTransactionIndex(tx.Hash(), block.Hash(), index)
		if err != nil {
			Logger.log.Error("ERROR", err, "Transaction in block with hash", blockHash, "and index", index, ":", tx)
			return NewBlockChainError(UnExpectedError, err)
		}

		if len(block.Body.Transactions) < 1 {
			Logger.log.Infof("No transaction in this block")
		} else {
			Logger.log.Infof("Number of transaction in this block %+v", len(block.Body.Transactions))
		}

		for index, tx := range block.Body.Transactions {
			if tx.GetType() == common.TxCustomTokenPrivacyType {
				_ = 1
			}
			err := self.StoreTransactionIndex(tx.Hash(), block.Hash(), index)
			if err != nil {
				Logger.log.Error("ERROR", err, "Transaction in block with hash", blockHash, "and index", index, ":", tx)
				return NewBlockChainError(UnExpectedError, err)
			}
			Logger.log.Infof("Transaction in block with hash", blockHash, "and index", index, ":", tx)
		}
	}

	err = self.BestState.Shard[block.Header.ShardID].Update(block)
	if err != nil {
		Logger.log.Error("Error update best state for block", block, "in shard", block.Header.ShardID)
		return NewBlockChainError(UnExpectedError, err)
	}
	// }
	// TODO: @0xankylosaurus optimize for loop once instead of multiple times ; metadata.process
	// save index of block
	// err = self.StoreShardBlockIndex(block)
	// if err != nil {
	// 	return NewBlockChainError(UnExpectedError, err)
	// }
	// // fetch serialNumbers and commitments(utxo) from block and save
	// err = self.CreateAndSaveTxViewPointFromBlock(block)
	// if err != nil {
	// 	return NewBlockChainError(UnExpectedError, err)
	// }

	// // Save loan txs
	// err = self.ProcessLoanForBlock(block)
	// if err != nil {
	// 	return NewBlockChainError(UnExpectedError, err)
	// }

	// // Update utxo reward for dividends
	// err = self.UpdateDividendPayout(block)
	// if err != nil {
	// 	return NewBlockChainError(UnExpectedError, err)
	// }

	// //Update database for vote board
	// err = self.UpdateVoteCountBoard(block)
	// if err != nil {
	// 	return NewBlockChainError(UnExpectedError, err)
	// }

	// //Update amount of token of each holder
	// err = self.UpdateVoteTokenHolder(block)
	// if err != nil {
	// 	return NewBlockChainError(UnExpectedError, err)
	// }

	// // Update database for vote proposal
	// err = self.ProcessVoteProposal(block)

	// // Process crowdsale tx
	// err = self.ProcessCrowdsaleTxs(block)
	// if err != nil {
	// 	return NewBlockChainError(UnExpectedError, err)
	// }

	Logger.log.Infof("Accepted block %s", blockHash)

	return nil
}

func (self *BlockChain) VerifyPreProcessingShardBlock(block *ShardBlock) error {
	if block.Header.Version != VERSION {
		return NewBlockChainError(VersionError, errors.New("Version should be :"+strconv.Itoa(VERSION)))
	}
	prevBlockHash := block.Header.PrevBlockHash
	// Verify parent hash exist or not
	parentBlockData, err := self.config.DataBase.FetchBlock(&prevBlockHash)
	if err != nil {
		return NewBlockChainError(DBError, err)
	}
	parentBlock := ShardBlock{}
	json.Unmarshal(parentBlockData, &parentBlock)
	// Verify block height with parent block
	if parentBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("Block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}
	// Verify epoch with parent block
	if block.Header.Height%EPOCH == 0 && parentBlock.Header.Epoch != block.Header.Epoch-1 {
		return NewBlockChainError(EpochError, errors.New("Block height and Epoch is not compatiable"))
	}
	// Verify timestamp with parent block
	if block.Header.Timestamp <= parentBlock.Header.Timestamp {
		return NewBlockChainError(TimestampError, errors.New("Timestamp of new block can't equal to parent block"))
	}

	return nil
}

func (self *BlockChain) VerifyPostProcessingShardBlock(block *ShardBlock) error {
	return nil
}

func CreateMerkleRootShard(txList []metadata.Transaction) common.Hash {
	//calculate output coin hash for each shard
	outputCoinHash := getOutCoinHashEachShard(txList)
	// calculate merkle data : 1 2 3 4 12 34 1234
	merkleData := outputCoinHash
	if len(merkleData)%2 == 1 {
		merkleData = append(merkleData, common.HashH([]byte{}))
	}

	cursor := 0
	for {
		v1 := merkleData[cursor]
		v2 := merkleData[cursor+1]
		merkleData = append(merkleData, common.HashH(append(v1.GetBytes(), v2.GetBytes()...)))
		cursor += 2
		if cursor >= len(merkleData)-1 {
			break
		}
	}
	merkleShardRoot := merkleData[len(merkleData)-1]
	return merkleShardRoot
}
