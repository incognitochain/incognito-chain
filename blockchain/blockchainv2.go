package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"io"
	"log"
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

	bestStateBeaconBytes, err := rawdbv2.GetBeaconBestState(blockchain.GetDatabase())
	if err == nil {
		beacon := &BeaconBestState{}
		err = json.Unmarshal(bestStateBeaconBytes, beacon)
		//update singleton object
		SetBeaconBestState(beacon)
		//update beacon field in blockchain Beststate
		blockchain.BestState.Beacon = GetBeaconBestState()
		errStateDB := blockchain.BestState.Beacon.InitStateRootHash(blockchain)
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
		bestStateBytes, err := rawdbv2.GetShardBestState(blockchain.config.DataBase, shardID)
		if err == nil {
			shardBestState := &ShardBestState{}
			err = json.Unmarshal(bestStateBytes, shardBestState)
			//update singleton object
			SetBestStateShard(shardID, shardBestState)
			//update Shard field in blockchain Beststate
			blockchain.BestState.Shard[shardID] = GetBestStateShard(shardID)
			errStateDB := blockchain.BestState.Shard[shardID].InitStateRootHash(blockchain.GetDatabase(), blockchain)
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
	err := blockchain.BestState.Beacon.initBeaconBestStateV2(initBlock, blockchain.GetDatabase())
	if err != nil {
		return err
	}
	tempBeaconBestState := blockchain.BestState.Beacon
	beaconBestStateBytes, err := json.Marshal(tempBeaconBestState)
	if err != nil {
		return NewBlockChainError(InitBeaconStateError, err)
	}
	blockchain.BestState.Beacon.lock.Lock()
	defer blockchain.BestState.Beacon.lock.Unlock()
	initBlockHash := tempBeaconBestState.BestBlock.Header.Hash()
	initBlockHeight := tempBeaconBestState.BestBlock.Header.Height
	// Insert new block into beacon chain
	if err := statedb.StoreAllShardCommittee(tempBeaconBestState.consensusStateDB, tempBeaconBestState.ShardCommittee, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking); err != nil {
		return err
	}
	if err := statedb.StoreBeaconCommittee(tempBeaconBestState.consensusStateDB, tempBeaconBestState.BeaconCommittee, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking); err != nil {
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
	if err := rawdbv2.StoreBeaconBestState(blockchain.GetDatabase(), beaconBestStateBytes); err != nil {
		Logger.log.Error("Error Store best state for block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return NewBlockChainError(InitBeaconStateError, err)
	}
	if err := rawdbv2.StoreBeaconBlock(blockchain.GetDatabase(), initBlockHeight, initBlockHash, &tempBeaconBestState.BestBlock); err != nil {
		Logger.log.Error("Error store beacon block", tempBeaconBestState.BestBlockHash, "in beacon chain")
		return err
	}
	if err := rawdbv2.StoreBeaconBlockIndex(blockchain.GetDatabase(), initBlockHeight, initBlockHash); err != nil {
		return err
	}
	// State Root Hash
	if err := rawdbv2.StoreConsensusStateRootHash(blockchain.GetDatabase(), initBlockHeight, consensusRootHash); err != nil {
		return err
	}
	if err := rawdbv2.StoreRewardStateRootHash(blockchain.GetDatabase(), initBlockHeight, common.EmptyRoot); err != nil {
		return err
	}
	if err := rawdbv2.StoreFeatureStateRootHash(blockchain.GetDatabase(), initBlockHeight, common.EmptyRoot); err != nil {
		return err
	}
	if err := rawdbv2.StoreSlashStateRootHash(blockchain.GetDatabase(), initBlockHeight, common.EmptyRoot); err != nil {
		return err
	}
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
	err = blockchain.BestState.Shard[shardID].initShardBestStateV2(blockchain, blockchain.GetDatabase(), &initShardBlock, genesisBeaconBlock)
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
	rootHash, err := blockchain.GetBeaconFeatureRootHash(blockchain.GetDatabase(), height)
	if err != nil {
		return nil, fmt.Errorf("Beacon Feature State DB not found, height %+v, error %+v", height, err)
	}
	return statedb.NewWithPrefixTrie(rootHash, statedb.NewDatabaseAccessWarper(db))
}

func (blockchain *BlockChain) GetBeaconSlashStateDB() *statedb.StateDB {
	return blockchain.BestState.Beacon.slashStateDB
}

func (blockchain *BlockChain) GetBeaconRewardStateDB() *statedb.StateDB {
	return blockchain.BestState.Beacon.rewardStateDB
}

func (blockchain *BlockChain) GetTransactionByHashV2(txHash common.Hash) (byte, common.Hash, int, metadata.Transaction, error) {
	blockHash, index, err := rawdbv2.GetTransactionByHash(blockchain.config.DataBase, txHash)
	if err != nil {
		return byte(255), common.Hash{}, -1, nil, NewBlockChainError(GetTransactionFromDatabaseError, err)
	}
	shardBlock, _, err := blockchain.GetShardBlockByHashV2(blockHash)
	if err != nil {
		return byte(255), common.Hash{}, -1, nil, NewBlockChainError(GetTransactionFromDatabaseError, err)
	}
	return shardBlock.Header.ShardID, blockHash, index, shardBlock.Body.Transactions[index], nil
}

// GetTransactionHashByReceiver - return list tx id which receiver get from any sender
// this feature only apply on full node, because full node get all data from all shard
func (blockchain *BlockChain) GetTransactionHashByReceiverV2(keySet *incognitokey.KeySet) (map[byte][]common.Hash, error) {
	result := make(map[byte][]common.Hash)
	var err error
	result, err = rawdbv2.GetTxByPublicKey(blockchain.config.DataBase, keySet.PaymentAddress.Pk)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	return result, nil
}

func (blockchain *BlockChain) StoreShardBestStateV2(shardID byte) error {
	return rawdbv2.StoreShardBestState(blockchain.GetDatabase(), shardID, blockchain.BestState.Shard[shardID])
}

func (blockchain *BlockChain) StoreBeaconBestStateV2() error {
	beaconBestStateBytes, err := json.Marshal(blockchain.BestState.Beacon)
	if err != nil {
		return err
	}
	return rawdbv2.StoreBeaconBestState(blockchain.config.DataBase, beaconBestStateBytes)
}

func CalculateNumberOfByteToRead(amountBytes int) []byte {
	var result = make([]byte, 8)
	binary.LittleEndian.PutUint32(result, uint32(amountBytes))
	return result
}
func GetNumberOfByteToRead(value []byte) (int, error) {
	var result uint32
	err := binary.Read(bytes.NewBuffer(value), binary.LittleEndian, &result)
	if err != nil {
		return -1, err
	}
	return int(result), nil
}
func (blockchain *BlockChain) BackupShardChain(writer io.Writer, shardID byte) error {
	bestStateBytes, err := rawdbv2.GetShardBestState(blockchain.config.DataBase, shardID)
	if err != nil {
		return err
	}
	shardBestState := &ShardBestState{}
	err = json.Unmarshal(bestStateBytes, shardBestState)
	bestShardHeight := shardBestState.ShardHeight
	var i uint64
	for i = 1; i < bestShardHeight; i++ {
		shardBlocks, err := blockchain.GetShardBlockByHeightV2(i, shardID)
		if err != nil {
			return err
		}
		var shardBlock *ShardBlock
		for _, v := range shardBlocks {
			shardBlock = v
		}
		data, err := json.Marshal(shardBlocks)
		if err != nil {
			return err
		}
		_, err = writer.Write(CalculateNumberOfByteToRead(len(data)))
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		if err != nil {
			return err
		}
		if i%100 == 0 {
			log.Printf("Backup Shard %+v Block %+v", shardBlock.Header.ShardID, i)
		}
		if i == bestShardHeight-1 {
			log.Printf("Finish Backup Shard %+v with Block %+v", shardBlock.Header.ShardID, i)
		}
	}
	return nil
}
func (blockchain *BlockChain) BackupBeaconChain(writer io.Writer) error {
	bestStateBytes, err := rawdbv2.GetBeaconBestState(blockchain.GetDatabase())
	if err != nil {
		return err
	}
	beaconBestState := &BeaconBestState{}
	err = json.Unmarshal(bestStateBytes, beaconBestState)
	bestBeaconHeight := beaconBestState.BeaconHeight
	var i uint64
	for i = 1; i < bestBeaconHeight; i++ {
		beaconBlocks, err := blockchain.GetBeaconBlockByHeightV2(i)
		if err != nil {
			return err
		}
		beaconBlock := beaconBlocks[0]
		data, err := json.Marshal(beaconBlock)
		if err != nil {
			return err
		}
		numOfByteToRead := CalculateNumberOfByteToRead(len(data))
		_, err = writer.Write(numOfByteToRead)
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		if err != nil {
			return err
		}
		if i%100 == 0 {
			log.Printf("Backup Beacon Block %+v", i)
		}
		if i == bestBeaconHeight-1 {
			log.Printf("Finish Backup Beacon with Block %+v", i)
		}
	}
	return nil
}
