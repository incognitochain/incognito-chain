package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"strings"
)

func (blockchain *BlockChain) ValidateBlockWithPreviousShardBestStateV2(shardBlock *ShardBlock) error {
	prevBST, err := rawdbv2.GetPreviousShardBestState(blockchain.GetDatabase(), shardBlock.Header.ShardID)
	if err != nil {
		return err
	}
	shardBestState := ShardBestState{}
	if err := json.Unmarshal(prevBST, &shardBestState); err != nil {
		return err
	}
	producerPk := shardBlock.Header.Producer
	producerPosition := (shardBestState.ShardProposerIdx) % len(shardBestState.ShardCommittee)
	tempProducer := shardBestState.ShardCommittee[producerPosition].GetMiningKeyBase58(shardBestState.ConsensusAlgorithm)
	if strings.Compare(tempProducer, producerPk) != 0 {
		return NewBlockChainError(ValidateBlockWithPreviousShardBestStateError, errors.New("Producer should be should be :"+tempProducer))
	}
	// Verify parent hash exist or not
	previousBlockHash := shardBlock.Header.PreviousBlockHash
	parentBlockBytes, err := rawdbv2.GetShardBlockByHash(blockchain.GetDatabase(), previousBlockHash)
	if err != nil {
		return NewBlockChainError(ValidateBlockWithPreviousShardBestStateError, err)
	}
	parentBlock := ShardBlock{}
	err = json.Unmarshal(parentBlockBytes, &parentBlock)
	if err != nil {
		return NewBlockChainError(ValidateBlockWithPreviousShardBestStateError, err)
	}
	// Verify shardBlock height with parent shardBlock
	if parentBlock.Header.Height+1 != shardBlock.Header.Height {
		return NewBlockChainError(ValidateBlockWithPreviousShardBestStateError, fmt.Errorf("ShardBlock height of new shardBlock should be: %+v", shardBlock.Header.Height+1))
	}
	return nil
}

//This only happen if user is a shard committee member.
func (blockchain *BlockChain) RevertShardStateV2(shardID byte) error {
	//Steps:
	// 1. Restore current beststate to previous beststate
	// 2. Set pool shardstate
	// 3. Delete newly inserted block
	// 4. Remove incoming crossShardBlks
	// 5. Delete txs and its related stuff (ex: txview) belong to block

	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	return blockchain.revertShardStateV2(shardID)
}

func (blockchain *BlockChain) revertShardStateV2(shardID byte) error {
	//Steps:
	// 1. Restore current beststate to previous beststate
	// 2. Set pool shardstate
	// 3. Delete newly inserted block
	// 4. Remove incoming crossShardBlks
	// 5. Delete txs and its related stuff (ex: txview) belong to block
	var currentBestState ShardBestState
	err := currentBestState.cloneShardBestStateFrom(blockchain.BestState.Shard[shardID])
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	revertedBestShardBlock := currentBestState.BestBlock
	for _, tx := range revertedBestShardBlock.Body.Transactions {
		if err := rawdbv2.DeleteTransactionIndex(blockchain.GetDatabase(), *tx.Hash()); err != nil {
			return NewBlockChainError(RevertStateError, err)
		}
	}
	err = rawdbv2.DeleteShardBlock(blockchain.GetDatabase(), shardID, revertedBestShardBlock.Header.Height, revertedBestShardBlock.Header.Hash())
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if err := rawdbv2.DeleteShardConsensusRootHash(blockchain.GetDatabase(), shardID, revertedBestShardBlock.Header.Height); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if err := rawdbv2.DeleteShardTransactionRootHash(blockchain.GetDatabase(), shardID, revertedBestShardBlock.Header.Height); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if err := rawdbv2.DeleteShardFeatureRootHash(blockchain.GetDatabase(), shardID, revertedBestShardBlock.Header.Height); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if err := rawdbv2.DeleteShardCommitteeRewardRootHash(blockchain.GetDatabase(), shardID, revertedBestShardBlock.Header.Height); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if err := rawdbv2.DeleteShardSlashRootHash(blockchain.GetDatabase(), shardID, revertedBestShardBlock.Header.Height); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	// Revert current shard best state to previous shard best state
	err = blockchain.revertShardBestStateV2(shardID)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if err := blockchain.StoreShardBestStateV2(shardID, nil); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	blockchain.config.ShardPool[shardID].RevertShardPool(blockchain.BestState.Shard[shardID].ShardHeight)
	for sid, height := range blockchain.BestState.Shard[shardID].BestCrossShard {
		blockchain.config.CrossShardPool[sid].RevertCrossShardPool(height)
	}
	Logger.log.Criticalf("REVERT SHARD SUCCESS FROM height %+v to %+v", revertedBestShardBlock.Header.Height, blockchain.BestState.Shard[shardID].ShardHeight)
	return nil
}

func (blockchain *BlockChain) revertShardBestStateV2(shardID byte) error {
	previousShardBestStateBytes, err := rawdbv2.GetPreviousShardBestState(blockchain.GetDatabase(), shardID)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousShardBestState := ShardBestState{}
	if err := json.Unmarshal(previousShardBestStateBytes, &previousShardBestState); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if previousShardBestState.ShardHeight == blockchain.BestState.Shard[shardID].ShardHeight {
		return NewBlockChainError(RevertStateError, fmt.Errorf("can't revert same best state, best shard height %+v", previousShardBestState.ShardHeight))
	}
	consensusRootHash, err := blockchain.GetShardConsensusRootHash(blockchain.GetDatabase(), shardID, previousShardBestState.ShardHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousShardBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(consensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	transactionRootHash, err := blockchain.GetShardTransactionRootHash(blockchain.GetDatabase(), shardID, previousShardBestState.ShardHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousShardBestState.transactionStateDB, err = statedb.NewWithPrefixTrie(transactionRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	featureRootHash, err := blockchain.GetShardFeatureRootHash(blockchain.GetDatabase(), shardID, previousShardBestState.ShardHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousShardBestState.featureStateDB, err = statedb.NewWithPrefixTrie(featureRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	rewardRootHash, err := blockchain.GetShardCommitteeRewardRootHash(blockchain.GetDatabase(), shardID, previousShardBestState.ShardHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousShardBestState.rewardStateDB, err = statedb.NewWithPrefixTrie(rewardRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	slashRootHash, err := blockchain.GetShardSlashRootHash(blockchain.GetDatabase(), shardID, previousShardBestState.ShardHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousShardBestState.slashStateDB, err = statedb.NewWithPrefixTrie(slashRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	SetBestStateShard(shardID, &previousShardBestState)
	return nil
}
