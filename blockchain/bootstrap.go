package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
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
	blockchain       *BlockChain
	lastBootStrap    *bootstrapProcess
	runningBootStrap *bootstrapProcess
}

type StateDBData struct {
	K []byte
	v []byte
}

func NewBootStrapManager(bc *BlockChain) *BootstrapManager {
	return &BootstrapManager{bc, nil, nil}
}
func (s *BootstrapManager) Start() {
	shardBestView := map[int]*ShardBestState{}
	beaconBestView := s.blockchain.GetBeaconBestState()
	checkPoint := time.Now().Format(time.RFC3339)
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
	fmt.Println("Backup beacon")
	cfg := config.LoadConfig()
	s.backupBeacon(path.Join(cfg.DataDir, cfg.DatabaseDir, checkPoint), beaconBestView)
	for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
		fmt.Println("Backup shard", i)
		s.backupShard(path.Join(cfg.DataDir, cfg.DatabaseDir, checkPoint), shardBestView[i])
	}

	//update final status
	s.lastBootStrap = &bootstrapProcess{
		beaconBestView.BeaconHeight,
	}
}

func (s *BootstrapManager) backupShard(name string, bestView *ShardBestState) {
	consensusDB := bestView.GetCopiedConsensusStateDB()
	txDB := bestView.GetCopiedTransactionStateDB()
	featureDB := bestView.GetCopiedFeatureStateDB()
	rewardDB := bestView.GetShardRewardStateDB()

	consensusFF, _ := flatfile.NewFlatFile(path.Join(name, "shard", fmt.Sprint(bestView.ShardID), "consensus"), 5000)
	featureFF, _ := flatfile.NewFlatFile(path.Join(name, "shard", fmt.Sprint(bestView.ShardID), "feature"), 5000)
	txFF, _ := flatfile.NewFlatFile(path.Join(name, "shard", fmt.Sprint(bestView.ShardID), "tx"), 5000)
	rewardFF, _ := flatfile.NewFlatFile(path.Join(name, "shard", fmt.Sprint(bestView.ShardID), "reward"), 5000)

	wg := sync.WaitGroup{}
	wg.Add(4)

	go backupStateDB(consensusDB, consensusFF, &wg)
	go backupStateDB(featureDB, featureFF, &wg)
	go backupStateDB(txDB, txFF, &wg)
	go backupStateDB(rewardDB, rewardFF, &wg)
	wg.Wait()
}

func (s *BootstrapManager) backupBeacon(name string, bestView *BeaconBestState) {
	consensusDB := bestView.GetBeaconConsensusStateDB()
	featureDB := bestView.GetBeaconFeatureStateDB()
	rewardDB := bestView.GetBeaconRewardStateDB()
	slashDB := bestView.GetBeaconSlashStateDB()
	fmt.Println(path.Join(name, "beacon", "consensus"))
	consensusFF, _ := flatfile.NewFlatFile(path.Join(name, "beacon", "consensus"), 5000)
	featureFF, _ := flatfile.NewFlatFile(path.Join(name, "beacon", "feature"), 5000)
	rewardFF, _ := flatfile.NewFlatFile(path.Join(name, "beacon", "reward"), 5000)
	slashFF, _ := flatfile.NewFlatFile(path.Join(name, "beacon", "slash"), 5000)

	wg := sync.WaitGroup{}
	wg.Add(4)

	go backupStateDB(consensusDB, consensusFF, &wg)
	go backupStateDB(featureDB, featureFF, &wg)
	go backupStateDB(rewardDB, rewardFF, &wg)
	go backupStateDB(slashDB, slashFF, &wg)
	wg.Wait()
}

func backupStateDB(stateDB *statedb.StateDB, ff *flatfile.FlatFileManager, wg *sync.WaitGroup) {
	defer wg.Done()
	it := stateDB.GetIterator()
	batchData := []StateDBData{}
	totalLen := 0
	for it.Next(false, true, true) {
		key := make([]byte, len(it.Key))
		value := make([]byte, len(it.Value))
		copy(key, it.Key)
		copy(value, it.Value)
		data := StateDBData{key, value}
		batchData = append(batchData, data)
		if len(batchData) == 1000 {
			totalLen += 1000

			buf := new(bytes.Buffer)
			enc := gob.NewEncoder(buf)
			err := enc.Encode(batchData)
			if err != nil {
				panic(err)
			}
			x, err := ff.Append(buf.Bytes())
			if err != nil {
				panic(err)
			}
			fmt.Println("write to batch", totalLen, len(buf.Bytes()), x)
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
