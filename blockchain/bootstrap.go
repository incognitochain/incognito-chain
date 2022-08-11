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
	"io/ioutil"
	"log"
	mathrand "math/rand"
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

func (r *remoteRPCClient) GetLatestBackup() (res BackupProcessInfo, err error) {
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
		Result BackupProcessInfo
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
		panic(err)
		return err
	}
	defer resp.Body.Close()
	for {
		b := make([]byte, 8)
		n, err := resp.Body.Read(b)
		if err != nil && err.Error() != "EOF" {
			panic(err)
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

func (s *BootstrapManager) BootstrapBeacon() {
	randServer := mathrand.Intn(len(s.hosts))
	host := fmt.Sprintf("http://%v", s.hosts[randServer])
	rpcClient := remoteRPCClient{host}
	latestBackup, _ := rpcClient.GetLatestBackup()

	log.Println("host", host)
	tmp, _ := json.Marshal(latestBackup)
	log.Println("latestBackup", string(tmp))

	if latestBackup.CheckpointName == "" {
		return
	}
	//only backup if back 500k block
	if latestBackup.BeaconView.BeaconHeight < s.blockchain.GetBeaconBestState().BeaconHeight+500*1000 {
		return
	}

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
		//time.Sleep(time.Second)
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

	bestView := latestBackup.BeaconView
	wg.Add(5)

	committeeFromBlock := map[byte]map[common.Hash]bool{}
	Logger.log.Info("Start bootstrap beacon from host", host)
	Logger.log.Infof("Stream block beacon from %v", latestBackup.MinBeaconHeight)

	rpcClient.OnBeaconBlock(latestBackup.MinBeaconHeight, bestView.BeaconHeight, func(beaconBlock types.BeaconBlock) {
		batch := s.blockchain.GetBeaconChainDatabase().NewBatch()

		if err := s.blockchain.BeaconChain.BlockStorage.StoreBlock(&beaconBlock); err != nil {
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

		batch.Write()

		for sid, shardStates := range beaconBlock.Body.ShardState {
			if _, ok := committeeFromBlock[sid]; !ok {
				committeeFromBlock[sid] = map[common.Hash]bool{}
			}
			for _, state := range shardStates {
				committeeFromBlock[sid][state.CommitteeFromBlock] = true
			}
		}

		if beaconBlock.GetHeight() == bestView.GetBeaconHeight() {
			bestView.BestBlock = beaconBlock
			for sid, committeeFromBlockHash := range committeeFromBlock {
				if cache, ok := CommitteeFromBlockBootStrapCache[sid]; !ok || cache == nil {
					CommitteeFromBlockBootStrapCache[sid], _ = lru.New(10)
				}
				for hash, _ := range committeeFromBlockHash {
					//stream committee from block and set to cache
					log.Println("stream", sid, hash.String())
					res, err := rpcClient.GetShardCommitteeByBeaconHash(int(sid), hash)
					if err != nil {
						panic(err)
					}
					err = rawdbv2.StoreCacheCommitteeFromBlock(s.blockchain.GetBeaconChainDatabase(), hash, int(sid), res)
					if err != nil {
						panic(err)
					}
					CommitteeFromBlockBootStrapCache[sid].Add(hash.String(), res)
				}
			}
			wg.Done()
		}
	})

	Logger.log.Info("Stream beacon consensus state ...")
	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconConsensus, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(bestView.ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(data)
	})

	Logger.log.Info("Stream beacon reward state ...")
	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconReward, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(bestView.ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(data)
	})

	Logger.log.Info("Stream beacon feature state ...")
	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconFeature, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(bestView.ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(data)
	})

	Logger.log.Info("Stream beacon slashing state ...")
	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconSlash, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(bestView.ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(data)
	})
	wg.Wait()

	Logger.log.Info("Finish sync ... post processing ...")

	allViews := []*BeaconBestState{bestView}
	b, _ := json.Marshal(allViews)
	err := rawdbv2.StoreBeaconViews(s.blockchain.GetBeaconChainDatabase(), b)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = s.blockchain.RestoreBeaconViews()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	Logger.log.Info("Bootstrap beacon finish!")
}

func (s *BootstrapManager) BootstrapShard(sid int) {
	randServer := mathrand.Intn(len(s.hosts))
	host := fmt.Sprintf("http://%v", s.hosts[randServer])
	rpcClient := remoteRPCClient{host}
	latestBackup, _ := rpcClient.GetLatestBackup()
	fmt.Println("latestBackup", latestBackup)
	if latestBackup.CheckpointName == "" {
		return
	}

	//only backup if back 500k block
	if latestBackup.ShardView[sid].ShardHeight < s.blockchain.GetBestStateShard(byte(sid)).ShardHeight+500*1000 {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(5)
	//retrieve beacon block -> backup height
	bestView := latestBackup.ShardView[sid]

	Logger.log.Infof("Start bootstrap shard %v from host %v", sid, host)
	Logger.log.Infof("Stream block shard %v from %v to %v", sid, s.blockchain.GetBestStateShard(byte(sid)).ShardHeight+1, bestView.ShardHeight)

	rpcClient.OnShardBlock(sid, s.blockchain.GetBestStateShard(byte(sid)).ShardHeight+1, bestView.ShardHeight, func(shardBlock types.ShardBlock) {
		batch := s.blockchain.GetShardChainDatabase(byte(sid)).NewBatch()
		if shardBlock.GetHeight()%1000 == 0 {
			log.Println("shard", sid, "save block ", shardBlock.GetHeight())
		}
		if err := s.blockchain.ShardChain[sid].BlockStorage.StoreTXIndex(&shardBlock); err != nil {
			panic(err)
		}

		if err := s.blockchain.ShardChain[sid].BlockStorage.StoreBlock(&shardBlock); err != nil {
			panic(err)
		}

		if err := s.blockchain.ShardChain[sid].BlockStorage.StoreFinalizedShardBlock(byte(sid), shardBlock.GetHeight(), *shardBlock.Hash()); err != nil {
			panic(err)
		}

		batch.Write()

		if shardBlock.GetHeight() == bestView.ShardHeight {
			bestView.BestBlock = &shardBlock
			wg.Done()
		}
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
	}

	recheckDB := func(sid int, roothash common.Hash) {
		txDB, err := statedb.NewWithPrefixTrie(roothash, statedb.NewDatabaseAccessWarper(s.blockchain.GetShardChainDatabase(byte(sid))))
		if err != nil {
			fmt.Println("check", sid, roothash.String(), err)
			panic(fmt.Sprintf("Something wrong when init txDB"))
		}
		if err := txDB.Recheck(); err != nil {
			fmt.Println(sid, err)
			fmt.Println("Recheck roothash fail!", roothash.String())
			panic("recheck fail!")
		}
	}
	Logger.log.Infof("Stream shard %v consensus state ...", sid)
	rpcClient.GetStateDB(latestBackup.CheckpointName, sid, ShardConsensus, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(sid, bestView.ConsensusStateDBRootHash)
			wg.Done()
			return
		}
		fmt.Println("sid", sid, len(data))
		flushDB(sid, data)
	})
	Logger.log.Infof("Stream shard %v transaction state ...", sid)
	rpcClient.GetStateDB(latestBackup.CheckpointName, sid, ShardTransacton, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(sid, bestView.TransactionStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(sid, data)
	})
	Logger.log.Infof("Stream shard %v reward state ...", sid)
	rpcClient.GetStateDB(latestBackup.CheckpointName, sid, ShardReward, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(sid, bestView.RewardStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(sid, data)
	})
	Logger.log.Infof("Stream shard %v feature state ...", sid)

	rpcClient.GetStateDB(latestBackup.CheckpointName, sid, ShardFeature, 0, func(data []byte) {
		if len(data) == 0 {
			recheckDB(sid, bestView.FeatureStateDBRootHash)
			wg.Done()
			return
		}
		flushDB(sid, data)
	})
	Logger.log.Info("Finish sync ... post processing ...")
	//post processing
	wg.Wait()
	allViews := []*ShardBestState{}
	allViews = append(allViews, bestView)
	if err := rawdbv2.StoreShardBestState(s.blockchain.GetShardChainDatabase(byte(sid)), byte(sid), allViews); err != nil {
		panic(err)
	}
	//s.blockchain.
	s.blockchain.RestoreShardViews(byte(sid))
	Logger.log.Infof("Bootstrap shard %v finish!", sid)
}
