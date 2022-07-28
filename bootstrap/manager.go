package bootstrap

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"path"
	"sync"
	"time"
)

type bootstrapProcess struct {
	checkPointHeight uint64
}

type BootstrapManager struct {
	blockchain       *blockchain.BlockChain
	lastBootStrap    *bootstrapProcess
	runningBootStrap *bootstrapProcess
}

type StateDBData struct {
	k []byte
	v []byte
}

func NewBootStrapManager(bc *blockchain.BlockChain) *BootstrapManager {
	return &BootstrapManager{bc, nil, nil}
}
func (s *BootstrapManager) Start() {
	shardBestView := map[int]*blockchain.ShardBestState{}
	beaconBestView := s.blockchain.GetBeaconBestState()
	checkPoint := time.Now().String()
	defer func() {
		s.runningBootStrap = nil
	}()
	for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
		shardBestView[i] = s.blockchain.GetBestStateShard(byte(i))
	}

	//update current status
	s.runningBootStrap = &bootstrapProcess{
		beaconBestView.BeaconHeight,
	}

	//backup beacon then shard
	s.backupBeacon(checkPoint, beaconBestView)
	for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
		s.backupShard(checkPoint, shardBestView[0])
	}

	//update final status
	s.lastBootStrap = &bootstrapProcess{
		beaconBestView.BeaconHeight,
	}
}

func (s *BootstrapManager) backupShard(name string, bestView *blockchain.ShardBestState) {
	consensusDB := bestView.GetCopiedConsensusStateDB()
	txDB := bestView.GetCopiedTransactionStateDB()
	featureDB := bestView.GetCopiedFeatureStateDB()
	rewardDB := bestView.GetShardRewardStateDB()

	consensusFF, _ := flatfile.NewFlatFile(path.Join(config.LoadConfig().DatabaseDir, name, "shard", fmt.Sprint(bestView.ShardID), "consensus"), 5000)
	featureFF, _ := flatfile.NewFlatFile(path.Join(config.LoadConfig().DatabaseDir, name, "shard", fmt.Sprint(bestView.ShardID), "feature"), 5000)
	txFF, _ := flatfile.NewFlatFile(path.Join(config.LoadConfig().DatabaseDir, name, "shard", fmt.Sprint(bestView.ShardID), "tx"), 5000)
	rewardFF, _ := flatfile.NewFlatFile(path.Join(config.LoadConfig().DatabaseDir, name, "shard", fmt.Sprint(bestView.ShardID), "reward"), 5000)

	wg := sync.WaitGroup{}
	wg.Add(4)

	go backupStateDB(consensusDB, consensusFF, wg)
	go backupStateDB(featureDB, featureFF, wg)
	go backupStateDB(txDB, txFF, wg)
	go backupStateDB(rewardDB, rewardFF, wg)
	wg.Wait()
}

func (s *BootstrapManager) backupBeacon(name string, bestView *blockchain.BeaconBestState) {
	consensusDB := bestView.GetBeaconConsensusStateDB()
	featureDB := bestView.GetBeaconFeatureStateDB()
	rewardDB := bestView.GetBeaconRewardStateDB()
	slashDB := bestView.GetBeaconSlashStateDB()

	consensusFF, _ := flatfile.NewFlatFile(path.Join(config.LoadConfig().DatabaseDir, name, "beacon", "consensus"), 5000)
	featureFF, _ := flatfile.NewFlatFile(path.Join(config.LoadConfig().DatabaseDir, name, "beacon", "feature"), 5000)
	rewardFF, _ := flatfile.NewFlatFile(path.Join(config.LoadConfig().DatabaseDir, name, "beacon", "reward"), 5000)
	slashFF, _ := flatfile.NewFlatFile(path.Join(config.LoadConfig().DatabaseDir, name, "beacon", "slash"), 5000)

	wg := sync.WaitGroup{}
	wg.Add(4)

	go backupStateDB(consensusDB, consensusFF, wg)
	go backupStateDB(featureDB, featureFF, wg)
	go backupStateDB(rewardDB, rewardFF, wg)
	go backupStateDB(slashDB, slashFF, wg)
	wg.Wait()
}

func backupStateDB(stateDB *statedb.StateDB, ff *flatfile.FlatFileManager, wg sync.WaitGroup) {
	defer wg.Done()
	it := stateDB.GetIterator()
	batchData := []StateDBData{}
	for it.Next(false, true, true) {
		data := StateDBData{it.Key, it.Value}
		batchData = append(batchData, data)
		if len(batchData) == 500 {
			buf := new(bytes.Buffer)
			enc := gob.NewEncoder(buf)
			enc.Encode(batchData)
			ff.Append(buf.Bytes())
			batchData = []StateDBData{}
		}
	}
	if len(batchData) > 0 {
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		enc.Encode(batchData)
		ff.Append(buf.Bytes())
	}
}
