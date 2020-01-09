package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

/*
// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
*/
func (blockchain *BlockChain) initChainStateV2() error {
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool
	blockchain.Chains = make(map[string]ChainInterface)
	blockchain.BestState = &BestState{
		Beacon: nil,
		Shard:  make(map[byte]*ShardBestState),
	}

	bestStateBeaconBytes, err := rawdbv2.FetchBeaconBestState(blockchain.GetDatabase())
	if err == nil {
		beacon := &BeaconBestState{}
		err = json.Unmarshal(bestStateBeaconBytes, beacon)
		//update singleton object
		SetBeaconBestState(beacon)
		//update beacon field in blockchain Beststate
		blockchain.BestState.Beacon = GetBeaconBestState()
		errStateDB := blockchain.BestState.Beacon.InitStateRootHash(blockchain.GetDatabase())
		if errStateDB != nil {
			return errStateDB
		}
		if err != nil {
			initialized = false
		} else {
			initialized = true
		}
	} else {
		initialized = false
	}
	if !initialized {
		// At this point the database has not already been initialized, so
		// initialize both it and the chain state to the genesis block.
		err := blockchain.initBeaconStateV2()
		if err != nil {
			return err
		}
	}
	beaconChain := BeaconChain{
		BestState:  GetBeaconBestState(),
		BlockGen:   blockchain.config.BlockGen,
		ChainName:  common.BeaconChainKey,
		Blockchain: blockchain,
	}
	blockchain.Chains[common.BeaconChainKey] = &beaconChain
	for shard := 1; shard <= blockchain.BestState.Beacon.ActiveShards; shard++ {
		shardID := byte(shard - 1)
		bestStateBytes, err := rawdbv2.FetchShardBestState(blockchain.config.DataBase, shardID)
		if err == nil {
			shardBestState := &ShardBestState{}
			err = json.Unmarshal(bestStateBytes, shardBestState)
			//update singleton object
			SetBestStateShard(shardID, shardBestState)
			//update Shard field in blockchain Beststate
			blockchain.BestState.Shard[shardID] = GetBestStateShard(shardID)
			errStateDB := blockchain.BestState.Shard[shardID].InitStateRootHash(blockchain.GetDatabase())
			if errStateDB != nil {
				return errStateDB
			}
			if err != nil {
				initialized = false
			} else {
				initialized = true
			}
		} else {
			initialized = false
		}

		if !initialized {
			// At this point the database has not already been initialized, so
			// initialize both it and the chain state to the genesis block.
			err := blockchain.initShardStateV2(shardID)
			if err != nil {
				return err
			}
		}
		shardChain := ShardChain{
			BestState:  GetBestStateShard(shardID),
			BlockGen:   blockchain.config.BlockGen,
			ChainName:  common.GetShardChainKey(shardID),
			Blockchain: blockchain,
		}
		blockchain.Chains[shardChain.ChainName] = &shardChain
	}

	return nil
}

func (blockchain *BlockChain) initBeaconStateV2() error {
	blockchain.BestState.Beacon = NewBeaconBestStateWithConfig(blockchain.config.ChainParams)
	initBlock := blockchain.config.ChainParams.GenesisBeaconBlock
	err := blockchain.BestState.Beacon.initBeaconBestState(initBlock, blockchain.GetDatabase())
	if err != nil {
		return err
	}
	tempBeaconBestState := blockchain.BestState.Beacon
	initBlockHash := tempBeaconBestState.BestBlock.Header.Hash()
	initBlockHeight := tempBeaconBestState.BestBlock.Header.Height
	// Insert new block into beacon chain
	if err := rawdbv2.StoreBeaconBestState(blockchain.GetDatabase(), tempBeaconBestState); err != nil {
		Logger.log.Error("Error Store best state for block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}
	if err := rawdbv2.StoreBeaconBlock(blockchain.GetDatabase(), initBlockHeight, initBlockHash, &tempBeaconBestState.BestBlock); err != nil {
		Logger.log.Error("Error store beacon block", tempBeaconBestState.BestBlockHash, "in beacon chain")
		return err
	}
	if err := rawdbv2.StoreBeaconBlockIndex(blockchain.GetDatabase(), initBlockHash, initBlockHeight); err != nil {
		return err
	}
	if err := statedb.StoreAllShardCommittee(tempBeaconBestState.consensusStateDB, tempBeaconBestState.GetShardCommittee(), tempBeaconBestState.GetRewardReceiver(), tempBeaconBestState.GetAutoStaking()); err != nil {
		return err
	}
	if err := statedb.StoreBeaconCommittee(tempBeaconBestState.consensusStateDB, tempBeaconBestState.GetBeaconCommittee(), tempBeaconBestState.GetRewardReceiver(), tempBeaconBestState.GetAutoStaking()); err != nil {
		return err
	}
	consensusRootHash, err := beaconBestState.consensusStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = beaconBestState.consensusStateDB.Database().TrieDB().Commit(consensusRootHash, false)
	if err != nil {
		return err
	}
	beaconBestState.consensusStateDB.ClearObjects()
	tempBeaconBestState.ConsensusStateRootHash[initBlockHeight] = consensusRootHash
	return nil
}

/*
// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
*/
func (blockchain *BlockChain) initShardStateV2(shardID byte) error {
	blockchain.BestState.Shard[shardID] = NewBestStateShardWithConfig(shardID, blockchain.config.ChainParams)
	// Create a new block from genesis block and set it as best block of chain
	initShardBlock := ShardBlock{}
	initShardBlock = *blockchain.config.ChainParams.GenesisShardBlock
	initShardBlock.Header.ShardID = shardID
	initShardBlockHeight := initShardBlock.Header.Height
	_, newShardCandidate := GetStakingCandidate(*blockchain.config.ChainParams.GenesisBeaconBlock)
	newShardCandidateStructs := []incognitokey.CommitteePublicKey{}
	for _, candidate := range newShardCandidate {
		key := incognitokey.CommitteePublicKey{}
		err := key.FromBase58(candidate)
		if err != nil {
			return err
		}
		newShardCandidateStructs = append(newShardCandidateStructs, key)
	}
	blockchain.BestState.Shard[shardID].ShardCommittee = append(blockchain.BestState.Shard[shardID].ShardCommittee, newShardCandidateStructs[int(shardID)*blockchain.config.ChainParams.MinShardCommitteeSize:(int(shardID)*blockchain.config.ChainParams.MinShardCommitteeSize)+blockchain.config.ChainParams.MinShardCommitteeSize]...)
	tempShardBestState := blockchain.BestState.Shard[shardID]
	beaconBlocks, err := blockchain.GetBeaconBlockByHeightV2(initShardBlockHeight)
	genesisBeaconBlock := beaconBlocks[0]
	if err != nil {
		return NewBlockChainError(FetchBeaconBlockError, err)
	}
	err = blockchain.BestState.Shard[shardID].initShardBestState(blockchain, blockchain.GetDatabase(), &initShardBlock, genesisBeaconBlock)
	if err != nil {
		return err
	}
	committeeChange := newCommitteeChange()
	committeeChange.shardCommitteeAdded[shardID] = tempShardBestState.GetShardCommittee()
	err = blockchain.processStoreShardBlockV2(&initShardBlock, committeeChange)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) GetShardRewardStateDB(shardID byte) *statedb.StateDB {
	return blockchain.BestState.Shard[shardID].rewardStateDB
}

func (blockchain *BlockChain) GetTransactionStateDB(shardID byte) *statedb.StateDB {
	return blockchain.BestState.Shard[shardID].transactionStateDB
}

func (blockchain *BlockChain) GetShardFeatureStateDB(shardID byte) *statedb.StateDB {
	return blockchain.BestState.Shard[shardID].featureStateDB
}

func (blockchain *BlockChain) GetBeaconFeatureStateDB() *statedb.StateDB {
	return blockchain.BestState.Beacon.featureStateDB
}

func (blockchain *BlockChain) GetBeaconFeatureStateDBByHeight(height uint64, db incdb.Database) (*statedb.StateDB, error) {
	rootHash, ok := blockchain.BestState.Beacon.FeatureStateRootHash[height]
	if !ok {
		return nil, fmt.Errorf("Beacon Feature State DB not found, height %+v", height)
	}
	return statedb.NewWithPrefixTrie(rootHash, statedb.NewDatabaseAccessWarper(db))
}

func (blockchain *BlockChain) GetBeaconSlashStateDB() *statedb.StateDB {
	return blockchain.BestState.Beacon.slashStateDB
}

func (blockchain *BlockChain) GetBeaconRewardStateDB() *statedb.StateDB {
	return blockchain.BestState.Beacon.rewardStateDB
}
