package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
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
	//TODO: change from dbv1 => dbv2 for shard
	for shard := 1; shard <= blockchain.BestState.Beacon.ActiveShards; shard++ {
		shardID := byte(shard - 1)
		bestStateBytes, err := rawdb.FetchShardBestState(blockchain.config.DataBase, shardID)
		if err == nil {
			shardBestState := &ShardBestState{}
			err = json.Unmarshal(bestStateBytes, shardBestState)
			//update singleton object
			SetBestStateShard(shardID, shardBestState)
			//update Shard field in blockchain Beststate
			blockchain.BestState.Shard[shardID] = GetBestStateShard(shardID)
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
			err := blockchain.initShardState(shardID)
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
	if err := statedb.StoreAllShardCommittee(tempBeaconBestState.consensusStateDB, tempBeaconBestState.GetShardCommittee(), tempBeaconBestState.GetRewardReceiver(), tempBeaconBestState.GetAutoStaking()); err != nil {
		return err
	}
	if err := statedb.StoreBeaconCommittee(tempBeaconBestState.consensusStateDB, tempBeaconBestState.GetBeaconCommittee(), tempBeaconBestState.GetRewardReceiver(), tempBeaconBestState.GetAutoStaking()); err != nil {
		return err
	}
	if err := rawdbv2.StoreBeaconBlockIndex(blockchain.GetDatabase(), initBlockHash, initBlockHeight); err != nil {
		return err
	}
	return nil
}
