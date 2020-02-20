package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func (blockchain *BlockChain) BackupCurrentBeaconState(block *BeaconBlock) error {
	beaconBestStateBytes, err := json.Marshal(blockchain.BestState.Beacon)
	if err != nil {
		return NewBlockChainError(BackupCurrentBeaconStateError, err)
	}
	if err := rawdbv2.StorePreviousBeaconBestState(blockchain.GetDatabase(), beaconBestStateBytes); err != nil {
		return NewBlockChainError(BackupCurrentBeaconStateError, err)
	}
	return nil
}

func (blockchain *BlockChain) BackupCurrentShardState(shardBlock *ShardBlock) error {
	shardBestStateBytes, err := json.Marshal(blockchain.BestState.Shard[shardBlock.Header.ShardID])
	if err != nil {
		return NewBlockChainError(BackUpShardStateError, err)
	}
	if err := rawdbv2.StorePreviousShardBestState(blockchain.GetDatabase(), shardBlock.Header.ShardID, shardBestStateBytes); err != nil {
		return NewBlockChainError(BackUpShardStateError, err)
	}
	return nil
}

func (blockchain *BlockChain) ValidateBlockWithPreviousBeaconBestState(beaconBlock *BeaconBlock) error {
	previousBeaconBestStateBytes, err := rawdbv2.GetPreviousBeaconBestState(blockchain.GetDatabase())
	if err != nil {
		return NewBlockChainError(ValidateBlockWithPreviousBeaconBestStateError, err)
	}
	previousBeaconBestState := BeaconBestState{}
	if err := json.Unmarshal(previousBeaconBestStateBytes, &previousBeaconBestState); err != nil {
		return NewBlockChainError(ValidateBlockWithPreviousBeaconBestStateError, err)
	}
	producerPk := beaconBlock.Header.Producer
	producerPosition := (previousBeaconBestState.BeaconProposerIndex + beaconBlock.Header.Round) % len(previousBeaconBestState.BeaconCommittee)
	tempProducer := previousBeaconBestState.BeaconCommittee[producerPosition].GetMiningKeyBase58(previousBeaconBestState.ConsensusAlgorithm)
	if strings.Compare(tempProducer, producerPk) != 0 {
		return NewBlockChainError(ValidateBlockWithPreviousBeaconBestStateError, fmt.Errorf("Producer should be should be: %+v", tempProducer))
	}
	//verify version
	if beaconBlock.Header.Version != BEACON_BLOCK_VERSION {
		return NewBlockChainError(ValidateBlockWithPreviousBeaconBestStateError, fmt.Errorf("Version should be: %+v", strconv.Itoa(BEACON_BLOCK_VERSION)))
	}
	prevBlockHash := beaconBlock.Header.PreviousBlockHash
	// Verify parent hash exist or not
	parentBlockBytes, err := rawdbv2.GetBeaconBlockByHash(blockchain.GetDatabase(), prevBlockHash)
	if err != nil {
		return NewBlockChainError(DatabaseError, err)
	}
	parentBlock := NewBeaconBlock()
	err = json.Unmarshal(parentBlockBytes, parentBlock)
	if err != nil {
		return NewBlockChainError(ValidateBlockWithPreviousBeaconBestStateError, err)
	}
	// Verify beaconBlock height with parent beaconBlock
	if parentBlock.Header.Height+1 != beaconBlock.Header.Height {
		return NewBlockChainError(ValidateBlockWithPreviousBeaconBestStateError, fmt.Errorf("beaconBlock height of new beaconBlock should be: %+v", strconv.Itoa(int(beaconBlock.Header.Height+1))))
	}
	return nil
}

func (blockchain *BlockChain) ValidateBlockWithPreviousShardBestState(shardBlock *ShardBlock) error {
	prevBST, err := rawdbv2.GetPreviousShardBestState(blockchain.GetDatabase(), shardBlock.Header.ShardID)
	if err != nil {
		return err
	}
	shardBestState := ShardBestState{}
	if err := json.Unmarshal(prevBST, &shardBestState); err != nil {
		return err
	}
	producerPk := shardBlock.Header.Producer
	producerPosition := (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)
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

// RevertShardState only happen if user is a shard committee member.
func (blockchain *BlockChain) RevertShardState(shardID byte) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	return blockchain.revertShardState(shardID)
}

// revertShardState steps
// 1. Delete transaction
// 2. Delete reverted shard block
// 3. Delete root hash of reverted shard block
// 4. Revert Shard Best State to previous shard best state
// 5. Update Cross Shard Pool and Shard Pool
func (blockchain *BlockChain) revertShardState(shardID byte) error {

	var revertedBestState ShardBestState
	err := revertedBestState.cloneShardBestStateFrom(blockchain.BestState.Shard[shardID])
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	revertedBestShardBlock := revertedBestState.BestBlock
	// Revert current shard best state to previous shard best state
	err = blockchain.revertShardBestState(shardID)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if err := blockchain.StoreShardBestState(shardID); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	for _, tx := range revertedBestShardBlock.Body.Transactions {
		if err := rawdbv2.DeleteTransactionIndex(blockchain.GetDatabase(), *tx.Hash()); err != nil {
			return NewBlockChainError(RevertStateError, err)
		}
	}
	if err = rawdbv2.DeleteShardBlock(blockchain.GetDatabase(), shardID, revertedBestShardBlock.Header.Height, revertedBestShardBlock.Header.Hash()); err != nil {
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
	Logger.log.Criticalf("REVERT SHARD SUCCESS FROM height %+v to %+v", revertedBestShardBlock.Header.Height, blockchain.BestState.Shard[shardID].ShardHeight)
	return nil
}

func (blockchain *BlockChain) revertShardBestState(shardID byte) error {
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
	blockchain.config.ShardPool[shardID].RevertShardPool(blockchain.BestState.Shard[shardID].ShardHeight)
	for sid, height := range blockchain.BestState.Shard[shardID].BestCrossShard {
		blockchain.config.CrossShardPool[sid].RevertCrossShardPool(height)
	}
	return nil
}

//This only happen if user is a beacon committee member.
func (blockchain *BlockChain) RevertBeaconState() error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	return blockchain.revertBeaconState()
}

// revertBeaconState
// 1. Restore current beststate to previous beststate
// 2. Set beacon/shardtobeacon pool state
// 3. Delete reverted block
// 4. Restore cross shard state
func (blockchain *BlockChain) revertBeaconState() error {
	var currentBestState BeaconBestState
	err := currentBestState.CloneBeaconBestStateFrom(blockchain.BestState.Beacon)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	currentBestBeaconBlock := currentBestState.BestBlock
	err = blockchain.revertBeaconBestState()
	if err != nil {
		return err
	}
	lastCrossShardState := beaconBestState.LastCrossShardState
	for fromShard, toShards := range lastCrossShardState {
		for toShard, height := range toShards {
			err := rawdbv2.RestoreCrossShardNextHeights(blockchain.GetDatabase(), fromShard, toShard, height)
			if err != nil {
				return NewBlockChainError(RevertStateError, err)
			}
		}
		blockchain.config.CrossShardPool[fromShard].UpdatePool()
	}
	err = rawdbv2.DeleteBeaconBlock(blockchain.GetDatabase(), currentBestBeaconBlock.Header.Height, currentBestBeaconBlock.Header.Hash())
	if err != nil {
		return err
	}
	if err := blockchain.StoreBeaconBestState(); err != nil {
		return err
	}
	Logger.log.Criticalf("REVERT BEACON SUCCESS from %+v to %+v", currentBestBeaconBlock.Header.Height, blockchain.BestState.Beacon.BeaconHeight)
	return nil
}

func (blockchain *BlockChain) revertBeaconBestState() error {
	previousBeaconBestStatebytes, err := rawdbv2.GetPreviousBeaconBestState(blockchain.GetDatabase())
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousBeaconBestState := BeaconBestState{}
	if err := json.Unmarshal(previousBeaconBestStatebytes, &previousBeaconBestState); err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	if previousBeaconBestState.BeaconHeight == blockchain.BestState.Beacon.BeaconHeight {
		return NewBlockChainError(RevertStateError, errors.New("can't revert same beststate"))
	}
	consensusRootHash, err := blockchain.GetBeaconConsensusRootHash(blockchain.GetDatabase(), previousBeaconBestState.BeaconHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousBeaconBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(consensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	featureRootHash, err := blockchain.GetBeaconFeatureRootHash(blockchain.GetDatabase(), previousBeaconBestState.BeaconHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousBeaconBestState.featureStateDB, err = statedb.NewWithPrefixTrie(featureRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	rewardRootHash, err := blockchain.GetBeaconRewardRootHash(blockchain.GetDatabase(), previousBeaconBestState.BeaconHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousBeaconBestState.rewardStateDB, err = statedb.NewWithPrefixTrie(rewardRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	slashRootHash, err := blockchain.GetBeaconSlashRootHash(blockchain.GetDatabase(), previousBeaconBestState.BeaconHeight)
	if err != nil {
		return NewBlockChainError(RevertStateError, err)
	}
	previousBeaconBestState.slashStateDB, err = statedb.NewWithPrefixTrie(slashRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	SetBeaconBestState(&previousBeaconBestState)
	blockchain.config.BeaconPool.RevertBeconPool(previousBeaconBestState.BeaconHeight)
	for sid, height := range blockchain.BestState.Beacon.GetBestShardHeight() {
		blockchain.config.ShardToBeaconPool.RevertShardToBeaconPool(sid, height)
	}
	return nil
}
