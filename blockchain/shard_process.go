package blockchain

import (
	"encoding/json"
	"errors"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"math"
	"strconv"

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

//TODO: get #shard from param

//Receive a shard block, produce a map of CrossShard block with shardID as key
func GenerateAllCrossShardBlock(block *ShardBlock) map[byte]*CrossShardBlock {
	allCrossShard := make(map[byte]*CrossShardBlock)
	for i := 0; i < 256; i++ {
		crossShard, err := GenerateCrossShard(block, byte(i))
		if crossShard != nil && err == nil {
			allCrossShard[byte(i)] = crossShard
		}
	}
	return allCrossShard
}

//Receive a shard block and shard ID, produce CrossShard block for that shardID
func GenerateCrossShard(block *ShardBlock, shardID byte) (*CrossShardBlock, error) {
	crossShard := &CrossShardBlock{}
	utxoList := getOutCoinCrossShard(block.Body.Transactions, shardID)
	if len(utxoList) == 0 {
		return nil, nil
	}
	merklePathShard, merkleShardRoot := GetMerklePathCrossShard(block.Body.Transactions, shardID)

	if merkleShardRoot != block.Header.MerkleRootShard {
		return crossShard, NewBlockChainError(CrossShardBlockError, errors.New("MerkleRootShard mismatch"))
	}

	//Copy signature and header
	crossShard.AggregatedSig = block.AggregatedSig
	copy(crossShard.ValidatorsIdx, block.ValidatorsIdx)
	crossShard.ProducerSig = block.ProducerSig
	crossShard.Header = block.Header
	crossShard.MerklePathShard = merklePathShard
	crossShard.UTXOList = utxoList
	return crossShard, nil
}

//Receive tx list from shard block body and shard ID, produce merkle path for the UTXO CrossShard
func GetMerklePathCrossShard(txList []metadata.Transaction, shardID byte) (merklePathShard []common.Hash, merkleShardRoot common.Hash) {
	//calculate output coin hash for each shard
	outputCoinHash := getOutCoinHashEachShard(txList)
	// calculate merkel path for a shardID
	// step 1: calculate merkle data : 1 2 3 4 12 34 1234
	merkleData := outputCoinHash
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

	// step 2: get merkle path
	cursor = 0
	lastCursor := 0
	sid := int(shardID)
	i := sid
	for {
		if cursor >= len(merkleData)-2 {
			break
		}
		if i%2 == 0 {
			merklePathShard = append(merklePathShard, merkleData[cursor+i+1])
		} else {
			merklePathShard = append(merklePathShard, merkleData[cursor+i-1])
		}
		i = int(math.Floor(float64(i / 2)))

		if cursor == 0 {
			cursor += len(outputCoinHash)
		} else {
			tmp := cursor
			cursor += int(math.Floor(float64((cursor - lastCursor) / 2)))
			lastCursor = tmp
		}
	}
	merkleShardRoot = merkleData[len(merkleData)-1]
	return merklePathShard, merkleShardRoot
}

//Receive a cross shard block and merkle path, verify whether the UTXO list is valid or not
func VerifyCrossShardBlockUTXO(block *CrossShardBlock, merklePathShard []common.Hash) bool {
	outCoins := block.UTXOList
	tmpByte := []byte{}
	for _, coin := range outCoins {
		tmpByte = append(tmpByte, coin.Bytes()...)
	}
	finalHash := common.HashH(tmpByte)
	for _, hash := range merklePathShard {
		finalHash = common.HashH(append(finalHash.GetBytes(), hash.GetBytes()...))
	}

	merkleShardRoot := block.Header.MerkleRootShard
	if finalHash != merkleShardRoot {
		return false
	} else {
		return true
	}
}

// helper function to group OutputCoin into shard and get the hash of each group
func getOutCoinHashEachShard(txList []metadata.Transaction) []common.Hash {
	// group transaction by shardID
	outCoinEachShard := make([][]*privacy.OutputCoin, 256)
	for _, tx := range txList {
		for _, outCoin := range tx.GetProof().OutputCoins {
			lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
			outCoinEachShard[lastByte] = append(outCoinEachShard[lastByte], outCoin)
		}
	}
	//calcualte hash for each shard
	outputCoinHash := make([]common.Hash, 256)
	for i := 0; i < 256; i++ {
		if len(outCoinEachShard[i]) == 0 {
			outputCoinHash[i] = common.HashH([]byte(""))
		} else {
			tmpByte := []byte{}
			for _, coin := range outCoinEachShard[i] {
				tmpByte = append(tmpByte, coin.Bytes()...)
			}
			outputCoinHash[i] = common.HashH(tmpByte)
		}
	}
	return outputCoinHash
}

// helper function to get the hash of OutputCoins (send to a shard) from list of transaction
func getOutCoinCrossShard(txList []metadata.Transaction, shardID byte) []privacy.OutputCoin {
	coinList := []privacy.OutputCoin{}
	for _, tx := range txList {
		for _, outCoin := range tx.GetProof().OutputCoins {
			lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
			if lastByte == shardID {
				coinList = append(coinList, *outCoin)
			}
		}
	}
	return coinList
}
