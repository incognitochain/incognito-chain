package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type remoteRPCClient struct {
	Endpoint string
}
type ErrMsg struct {
	Code       int
	Message    string
	StackTrace string
}

func (r *remoteRPCClient) sendRequest(requestBody []byte) ([]byte, error) {
	resp, err := http.Post(r.Endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *remoteRPCClient) GetShardCommitteeByBeaconHash(sid int, beaconHash common.Hash) (res []incognitokey.CommitteePublicKey, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getshardcommitteestatebybeaconhash",
		"params":  []interface{}{sid, beaconHash},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, err
	}

	resp := struct {
		Result []incognitokey.CommitteePublicKey
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res, err
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, nil
}

func (r *remoteRPCClient) GetLatestBackup() (res BackupProcess, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getlatestbackup",
		"params":  []interface{}{},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, err
	}

	resp := struct {
		Result BackupProcess
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res, err
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, nil
}

func (r *remoteRPCClient) GetStateDB(checkpoint string, cid int, dbType int, offset uint64, f func([]byte)) error {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbootstrapstatedb",
		"params":  []interface{}{checkpoint, cid, dbType, offset},
		"id":      1,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(r.Endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for {
		b := make([]byte, 8)
		n, err := resp.Body.Read(b)
		if err != nil && err.Error() != "EOF" {
			return err
		}
		if n == 0 {
			f(nil)
			return nil
		}
		var dataLen uint64

		err = binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &dataLen)
		if err != nil || n == 0 {
			return err
		}
		dataByte := make([]byte, dataLen)
		dataRead := uint64(0)
		for {
			n, _ := resp.Body.Read(dataByte[dataRead:])
			dataRead += uint64(n)
			if dataRead == dataLen {
				break
			}
		}
		f(dataByte)
	}
}

func (r *remoteRPCClient) GetBlocksFromHeight(shardID int, from uint64, num int) (res interface{}, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getblocksfromheight",
		"params":  []interface{}{shardID, from, num},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)

	if err != nil {
		return res, err
	}

	if shardID == -1 {
		resp := struct {
			Result []types.BeaconBlock
			Error  *ErrMsg
		}{}
		err = json.Unmarshal(body, &resp)
		if resp.Error != nil && resp.Error.StackTrace != "" {
			return res, errors.New(resp.Error.StackTrace)
		}
		if err != nil {
			return res, err
		}
		return resp.Result, nil
	} else {
		resp := struct {
			Result []types.ShardBlock
			Error  *ErrMsg
		}{}
		err = json.Unmarshal(body, &resp)
		if resp.Error != nil && resp.Error.StackTrace != "" {
			return res, errors.New(resp.Error.StackTrace)
		}
		if err != nil {
			return res, err
		}
		return resp.Result, nil
	}
}
func (s *remoteRPCClient) OnShardBlock(sid int, fromBlk uint64, toBlk uint64, f func(block types.ShardBlock)) {
	shardCh := make(chan types.ShardBlock, 500)
	go func() {
		for {
			data, err := s.GetBlocksFromHeight(sid, uint64(fromBlk), 50)
			if err != nil || len(data.([]types.ShardBlock)) == 0 {
				fmt.Println(err)
				time.Sleep(time.Minute)
				continue
			}
			for _, blk := range data.([]types.ShardBlock) {
				shardCh <- blk
				fromBlk = blk.GetHeight() + 1
				if blk.GetHeight() == toBlk {
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case blk := <-shardCh:
				f(blk)
				if blk.GetHeight() == toBlk {
					return
				}
			}
		}
	}()
}

func (s *remoteRPCClient) OnBeaconBlock(fromBlk uint64, toBlock uint64, f func(block types.BeaconBlock)) {
	beaconCh := make(chan types.BeaconBlock, 500)
	go func() {
		for {
			data, err := s.GetBlocksFromHeight(-1, uint64(fromBlk), 50)
			if err != nil || len(data.([]types.BeaconBlock)) == 0 {
				fmt.Println(err)
				time.Sleep(time.Minute)
				continue
			}
			for _, blk := range data.([]types.BeaconBlock) {
				beaconCh <- blk
				fromBlk = blk.GetHeight() + 1
				if blk.GetHeight() == toBlock {
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case blk := <-beaconCh:
				f(blk)
				if blk.GetHeight() == toBlock {
					return
				}
			}
		}
	}()
}

type BootstrapManager struct {
	hosts      []string
	blockchain *BlockChain
}

func NewBootstrapManager(hosts []string, bc *BlockChain) *BootstrapManager {
	return &BootstrapManager{hosts, bc}
}

var CommitteeFromBlockBootStrapCache = map[byte]*lru.Cache{}

func (s *BootstrapManager) BootstrapBeacon() {
	host := s.hosts[0]
	rpcClient := remoteRPCClient{host}
	latestBackup, _ := rpcClient.GetLatestBackup()
	fmt.Println("latestBackup", latestBackup)
	if latestBackup.CheckpointName == "" {
		return
	}
	lastCrossShardState := map[byte]map[byte]uint64{}
	wg := sync.WaitGroup{}

	//retrieve beacon stateDB
	flushDB := func(data []byte) {
		batch := s.blockchain.GetBeaconChainDatabase().NewBatch()
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		m := []StateDBData{}
		if err := dec.Decode(&m); err != nil {
			panic(err)
		}
		for _, stateData := range m {
			//fmt.Printf("%v - %v\n", stateData.K, len(stateData.V))
			err := batch.Put(stateData.K, stateData.V)
			if err != nil {
				panic(err)
			}

		}
		err := batch.Write()
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second)
	}

	recheckDB := func(roothash common.Hash) {
		txDB, err := statedb.NewWithPrefixTrie(roothash, statedb.NewDatabaseAccessWarper(s.blockchain.GetBeaconChainDatabase()))
		if err != nil {
			panic(fmt.Sprintf("Something wrong when init txDB"))
		}
		if err := txDB.Recheck(); err != nil {
			fmt.Println(err)
			fmt.Println("Recheck roothash fail!", roothash.String())
			panic("recheck fail!")
		}
	}

	beaconView := NewBeaconBestState()
	beaconView.CloneBeaconBestStateFrom(s.blockchain.GetBeaconBestState())
	wg.Add(5)

	rpcClient.OnBeaconBlock(2, latestBackup.ChainInfo[-1].Height, func(beaconBlock types.BeaconBlock) {
		batch := s.blockchain.GetBeaconChainDatabase().NewBatch()

		blockHash := beaconBlock.Hash()
		if err := rawdbv2.StoreBeaconBlockByHash(batch, *blockHash, beaconBlock); err != nil {
			panic(err)
		}

		for shardID, shardStates := range beaconBlock.Body.ShardState {
			for _, shardState := range shardStates {
				err := rawdbv2.StoreBeaconConfirmInstantFinalityShardBlock(batch, shardID, shardState.Height, shardState.Hash)
				if err != nil {
					panic(err)
				}
			}
		}

		if err := rawdbv2.StoreFinalizedBeaconBlockHashByIndex(batch, beaconBlock.GetHeight(), *beaconBlock.Hash()); err != nil {
			panic(err)
		}

		if err := processBeaconForConfirmmingCrossShard(s.blockchain, &beaconBlock, lastCrossShardState); err != nil {
			panic(err)
		}
		batch.Write()

		beaconView.BeaconHeight = beaconBlock.GetHeight()
		beaconView.BestBlock = beaconBlock
		beaconView.BestBlockHash = *beaconBlock.Hash()
		//update trigger feature
		for _, instr := range beaconBlock.Body.Instructions {
			if instr[0] == instruction.ENABLE_FEATURE {
				enableFeatures, err := instruction.ValidateAndImportEnableFeatureInstructionFromString(instr)
				if err != nil {
					panic(err)
				}
				if beaconView.TriggeredFeature == nil {
					beaconView.TriggeredFeature = make(map[string]uint64)
				}
				for _, feature := range enableFeatures.Features {
					beaconView.TriggeredFeature[feature] = beaconBlock.GetHeight()
				}

			}
		}

		//update bestshardhash, bestshardheight
		for sid, shardstate := range beaconBlock.Body.ShardState {
			beaconView.BestShardHash[sid] = shardstate[len(shardstate)-1].Hash
			beaconView.BestShardHeight[sid] = shardstate[len(shardstate)-1].Height
		}
		//update lastcrossshardstate
		beaconView.LastCrossShardState = lastCrossShardState

		//update random number
		if s.blockchain.IsFirstBeaconHeightInEpoch(beaconBlock.GetHeight()) {
			beaconView.IsGetRandomNumber = false
		} else if s.blockchain.IsEqualToRandomTime(beaconBlock.GetHeight()) {
			beaconView.CurrentRandomTimeStamp = beaconBlock.Header.Timestamp
		}

		for _, instr := range beaconBlock.Body.Instructions {
			if instr[0] == instruction.RANDOM_ACTION {
				randomInstruction, err := instruction.ValidateAndImportRandomInstructionFromString(instr)
				if err != nil {
					panic(err)
				}
				beaconView.CurrentRandomNumber = randomInstruction.RandomNumber()
				beaconView.IsGetRandomNumber = true
			}
		}
		//update number of shard block
		for shardID, shardStates := range beaconBlock.Body.ShardState {
			beaconView.NumberOfShardBlock[shardID] = beaconView.NumberOfShardBlock[shardID] + uint(len(shardStates))
		}
		if s.blockchain.IsFirstBeaconHeightInEpoch(beaconBlock.GetHeight()) {
			beaconView.NumberOfShardBlock = make(map[byte]uint)
			for i := 0; i < beaconView.ActiveShards; i++ {
				shardID := byte(i)
				beaconView.NumberOfShardBlock[shardID] = 0
			}
		}

		// update max committee size
		newMaxCommitteeSize := GetMaxCommitteeSize(beaconView.MaxShardCommitteeSize, beaconView.TriggeredFeature, beaconBlock.Header.Height)
		if newMaxCommitteeSize != beaconView.MaxShardCommitteeSize {
			beaconView.MaxShardCommitteeSize = newMaxCommitteeSize
		}
		beaconView.tryUpgradeConsensusRule()

		//update LastBlockProcessBridge
		if beaconBlock.GetVersion() >= types.INSTANT_FINALITY_VERSION {
			if beaconBlock.Header.ProcessBridgeFromBlock != nil && *beaconBlock.Header.ProcessBridgeFromBlock != 0 {
				beaconView.LastBlockProcessBridge = beaconBlock.GetHeight() - 1
			}
		} else {
			beaconView.LastBlockProcessBridge = beaconBlock.GetHeight()
		}

		if beaconBlock.GetHeight() == latestBackup.ChainInfo[-1].Height {
			//get all required committeefromblock (for rebuild missing signature counter)
			firstBeaconHeightOfEpoch := s.blockchain.GetFirstBeaconHeightInEpoch(beaconView.Epoch)

			tempBeaconBlock := &beaconBlock
			tempBeaconHeight := beaconBlock.Header.Height
			committeeFromBlock := map[byte]map[common.Hash]bool{}
			log.Println("firstBeaconHeightOfEpoch", tempBeaconHeight, firstBeaconHeightOfEpoch)
			for tempBeaconHeight >= firstBeaconHeightOfEpoch {
				log.Println("checking beacon height", tempBeaconHeight)
				for sid, shardStates := range tempBeaconBlock.Body.ShardState {
					if _, ok := committeeFromBlock[sid]; !ok {
						committeeFromBlock[sid] = map[common.Hash]bool{}
					}
					for _, state := range shardStates {
						committeeFromBlock[sid][state.CommitteeFromBlock] = true
					}
				}
				if tempBeaconHeight == 1 {
					break
				}
				previousBeaconBlock, _, err := s.blockchain.GetBeaconBlockByHash(tempBeaconBlock.Header.PreviousBlockHash)
				if err != nil {
					panic(err)
				}
				tempBeaconBlock = previousBeaconBlock
				tempBeaconHeight--
			}
			log.Println("committeeFromBlock", committeeFromBlock)
			for sid, committeeFromBlockHash := range committeeFromBlock {
				if _, ok := CommitteeFromBlockCache[sid]; !ok {
					CommitteeFromBlockBootStrapCache[sid], _ = lru.New(5)
				}
				for hash, _ := range committeeFromBlockHash {
					//stream committee from block and set to cache
					log.Println("stream", sid, hash.String())
					res, err := rpcClient.GetShardCommitteeByBeaconHash(int(sid), hash)
					if err != nil {
						panic(err)
					}
					CommitteeFromBlockBootStrapCache[sid].Add(hash.String(), res)
				}
			}

			wg.Done()
		}
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconConsensus, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(latestBackup.ChainInfo[-1].ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(data)
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconReward, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(latestBackup.ChainInfo[-1].ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(data)
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconFeature, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(latestBackup.ChainInfo[-1].ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(data)
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconSlash, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(latestBackup.ChainInfo[-1].ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(data)
	})
	wg.Wait()

	beaconView.ConsensusStateDBRootHash = latestBackup.ChainInfo[-1].ConsensusStateDBRootHash
	beaconView.FeatureStateDBRootHash = latestBackup.ChainInfo[-1].FeatureStateDBRootHash
	beaconView.RewardStateDBRootHash = latestBackup.ChainInfo[-1].RewardStateDBRootHash
	beaconView.SlashStateDBRootHash = latestBackup.ChainInfo[-1].SlashStateDBRootHash

	allViews := []*BeaconBestState{beaconView}
	b, _ := json.Marshal(allViews)
	rawdbv2.StoreBeaconViews(s.blockchain.GetBeaconChainDatabase(), b)
	s.blockchain.RestoreBeaconViews()
}

func (s *BootstrapManager) BootstrapShard(sid int) {
	host := s.hosts[0]
	rpcClient := remoteRPCClient{host}
	latestBackup, _ := rpcClient.GetLatestBackup()
	fmt.Println("latestBackup", latestBackup)
	if latestBackup.CheckpointName == "" {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(5)
	//retrieve beacon block -> backup height
	bestView := NewShardBestStateWithShardID(byte(sid))
	bestView.cloneShardBestStateFrom(s.blockchain.GetBestStateShard(byte(sid)))

	rpcClient.OnShardBlock(sid, 2, latestBackup.ChainInfo[sid].Height, func(shardBlock types.ShardBlock) {
		batch := s.blockchain.GetShardChainDatabase(byte(sid)).NewBatch()

		for sid, crossTx := range shardBlock.Body.CrossTransactions {
			bestView.BestCrossShard[sid] = crossTx[len(crossTx)-1].BlockHeight
		}
		for index, tx := range shardBlock.Body.Transactions {
			if err := rawdbv2.StoreTransactionIndex(batch, *tx.Hash(), shardBlock.Header.Hash(), index); err != nil {
				panic(err)
			}
		}

		if err := rawdbv2.StoreShardBlock(batch, *shardBlock.Hash(), shardBlock); err != nil {
			panic(err)
		}

		if err := rawdbv2.StoreFinalizedShardBlockHashByIndex(batch, byte(sid), shardBlock.GetHeight(), *shardBlock.Hash()); err != nil {
			panic(err)
		}

		batch.Write()
		if shardBlock.GetHeight() == latestBackup.ChainInfo[sid].Height {
			bestView.BestBlockHash = *shardBlock.Hash()
			bestView.BestBeaconHash = shardBlock.Header.BeaconHash
			bestView.BeaconHeight = shardBlock.Header.BeaconHeight
			bestView.Epoch = shardBlock.GetCurrentEpoch()
			bestView.ShardHeight = shardBlock.GetHeight()
			bestView.NumTxns = uint64(len(shardBlock.Body.Transactions))
			bestView.TotalTxns += uint64(len(shardBlock.Body.Transactions))
			bestView.BestBlock = &shardBlock
			wg.Done()
		}
		//fmt.Println("save block", shardBlock.GetHeight())
	})

	//retrieve beacon stateDB
	flushDB := func(sid int, data []byte) {
		batch := s.blockchain.GetShardChainDatabase(byte(sid)).NewBatch()
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		m := []StateDBData{}
		if err := dec.Decode(&m); err != nil {
			panic(err)
		}
		for _, stateData := range m {
			//fmt.Println("write", stateData.K, len(stateData.V))
			err := batch.Put(stateData.K, stateData.V)
			if err != nil {
				panic(err)
			}

		}
		err := batch.Write()
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second)
	}

	recheckDB := func(sid int, roothash common.Hash) {
		txDB, err := statedb.NewWithPrefixTrie(roothash, statedb.NewDatabaseAccessWarper(s.blockchain.GetShardChainDatabase(byte(sid))))
		if err != nil {
			fmt.Println(bestView.ConsensusStateDBRootHash.String())
			fmt.Println("check", roothash.String(), err)
			panic(fmt.Sprintf("Something wrong when init txDB"))
		}
		if err := txDB.Recheck(); err != nil {
			fmt.Println(err)
			fmt.Println("Recheck roothash fail!", roothash.String())
			panic("recheck fail!")
		}
	}

	rpcClient.GetStateDB(latestBackup.CheckpointName, sid, ShardConsensus, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(sid, latestBackup.ChainInfo[sid].ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(sid, data)
	})
	rpcClient.GetStateDB(latestBackup.CheckpointName, sid, ShardTransacton, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(sid, latestBackup.ChainInfo[sid].TransactionStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(sid, data)
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, sid, ShardReward, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(sid, latestBackup.ChainInfo[sid].RewardStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(sid, data)
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, sid, ShardFeature, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(sid, latestBackup.ChainInfo[sid].FeatureStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(sid, data)
	})

	//post processing
	wg.Wait()
	bestView.ConsensusStateDBRootHash = latestBackup.ChainInfo[sid].ConsensusStateDBRootHash
	bestView.TransactionStateDBRootHash = latestBackup.ChainInfo[sid].TransactionStateDBRootHash
	bestView.RewardStateDBRootHash = latestBackup.ChainInfo[sid].RewardStateDBRootHash
	bestView.FeatureStateDBRootHash = latestBackup.ChainInfo[sid].FeatureStateDBRootHash
	allViews := []*ShardBestState{}
	allViews = append(allViews, bestView)
	if err := rawdbv2.StoreShardBestState(s.blockchain.GetShardChainDatabase(byte(sid)), byte(sid), allViews); err != nil {
		panic(err)
	}
	//s.blockchain.
	s.blockchain.RestoreShardViews(0)
}
