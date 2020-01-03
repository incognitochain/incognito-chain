package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processStoreShardBlockV2(shardBlock *ShardBlock) error {
	shardID := shardBlock.Header.ShardID
	blockHeight := shardBlock.Header.Height
	blockHash := shardBlock.Header.Hash()
	tempShardBestState := blockchain.BestState.Shard[shardID]
	tempBeaconBestState := blockchain.BestState.Beacon
	Logger.log.Infof("SHARD %+v | Process store block height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, *shardBlock.Hash())
	if err := rawdbv2.StoreShardBlock(blockchain.GetDatabase(), shardID, blockHeight, blockHash, shardBlock); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardBlockIndex(blockchain.GetDatabase(), shardID, blockHeight, blockHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardBestState(blockchain.GetDatabase(), shardID, tempShardBestState); err != nil {
		return NewBlockChainError(StoreBestStateError, err)
	}
	if len(shardBlock.Body.CrossTransactions) != 0 {
		Logger.log.Critical("processStoreShardBlockAndUpdateDatabase/CrossTransactions	", shardBlock.Body.CrossTransactions)
	}
	if err := blockchain.CreateAndSaveTxViewPointFromBlockV2(shardBlock, tempShardBestState.transactionStateDB, tempBeaconBestState.featureStateDB); err != nil {
		return NewBlockChainError(FetchAndStoreTransactionError, err)
	}

	for index, tx := range shardBlock.Body.Transactions {
		if err := rawdbv2.StoreTransactionIndex(blockchain.GetDatabase(), *tx.Hash(), shardBlock.Header.Hash(), index); err != nil {
			return NewBlockChainError(FetchAndStoreTransactionError, err)
		}
		// Process Transaction Metadata
		metaType := tx.GetMetadataType()
		if metaType == metadata.WithDrawRewardResponseMeta {
			_, publicKey, amountRes, coinID := tx.GetTransferData()
			err := statedb.RemoveCommitteeReward(tempBeaconBestState.rewardStateDB, publicKey, amountRes, *coinID)
			if err != nil {
				return NewBlockChainError(RemoveCommitteeRewardError, err)
			}
		}
		Logger.log.Debugf("Transaction in block with hash", blockHash, "and index", index)
	}
	// Store Incomming Cross Shard
	if err := blockchain.CreateAndSaveCrossTransactionCoinViewPointFromBlockV2(shardBlock, tempShardBestState.transactionStateDB); err != nil {
		return NewBlockChainError(FetchAndStoreCrossTransactionError, err)
	}
	// Save result of BurningConfirm instruction to get proof later
	err := blockchain.storeBurningConfirmV2(tempShardBestState.featureStateDB, shardBlock)
	if err != nil {
		return NewBlockChainError(StoreBurningConfirmError, err)
	}
	// Update bridge issuance request status
	err = blockchain.updateBridgeIssuanceStatusV2(tempShardBestState.featureStateDB, shardBlock)
	if err != nil {
		return NewBlockChainError(UpdateBridgeIssuanceStatusError, err)
	}
	// call FeeEstimator for processing
	if feeEstimator, ok := blockchain.config.FeeEstimator[shardBlock.Header.ShardID]; ok {
		err := feeEstimator.RegisterBlock(shardBlock)
		if err != nil {
			Logger.log.Warn(NewBlockChainError(RegisterEstimatorFeeError, err))
		}
	}
	Logger.log.Infof("SHARD %+v | 🔎 %d transactions in block height %+v \n", shardBlock.Header.ShardID, len(shardBlock.Body.Transactions), shardBlock.Header.Height)
	return nil
}
