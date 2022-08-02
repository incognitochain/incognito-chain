package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"io/ioutil"
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

func (s *BootstrapManager) BootstrapBeacon() {
	host := s.hosts[0]
	rpcClient := remoteRPCClient{host}
	latestBackup, _ := rpcClient.GetLatestBackup()
	fmt.Println("latestBackup", latestBackup)
	if latestBackup.CheckpointName == "" {
		return
	}

	//retrieve beacon block -> backup height
	lastCrossShardState := map[byte]map[byte]uint64{}
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
	})

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

	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconConsensus, 0, func(data []byte) {
		fmt.Println("data 0", len(data))
		flushDB(data)
		fmt.Println("recheck ", latestBackup.ChainInfo[-1].ConsensusStateDBRootHash)
		recheckDB(latestBackup.ChainInfo[-1].ConsensusStateDBRootHash)
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconReward, 0, func(data []byte) {
		fmt.Println("data 1", len(data))
		flushDB(data)
		recheckDB(latestBackup.ChainInfo[-1].RewardStateDBRootHash)
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconFeature, 0, func(data []byte) {
		fmt.Println("data 2", len(data))
		flushDB(data)
		recheckDB(latestBackup.ChainInfo[-1].FeatureStateDBRootHash)
	})

	rpcClient.GetStateDB(latestBackup.CheckpointName, -1, BeaconSlash, 0, func(data []byte) {
		fmt.Println("data 3", len(data))
		flushDB(data)
		recheckDB(latestBackup.ChainInfo[-1].SlashStateDBRootHash)
	})

	//post processing

	panic(1)
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
