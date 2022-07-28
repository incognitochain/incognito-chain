package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
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

const (
	BeaconConsensus = 1
	BeaconFeature   = 2
	BeaconReward    = 3
	BeaconSlash     = 4
	ShardConsensus  = 5
	ShardTransacton = 6
	ShardFeature    = 7
	ShardReward     = 8
)

type CheckpointInfo struct {
	Hash   string
	Height int64
}

func (s *BootstrapManager) getBackupReader(dbType int, cid int) (CheckpointInfo, *flatfile.FlatFileManager) {
	dbLoc := ""
	infoLoc := ""
	switch dbType {
	case BeaconConsensus:
		dbLoc = path.Join(dbLoc, "beacon", "consensus")
		infoLoc = path.Join(dbLoc, "beacon", "info")
	case BeaconFeature:
		dbLoc = path.Join(dbLoc, "beacon", "feature")
		infoLoc = path.Join(dbLoc, "beacon", "info")
	case BeaconReward:
		dbLoc = path.Join(dbLoc, "beacon", "reward")
		infoLoc = path.Join(dbLoc, "beacon", "info")
	case BeaconSlash:
		dbLoc = path.Join(dbLoc, "beacon", "slash")
		infoLoc = path.Join(dbLoc, "beacon", "info")
	case ShardConsensus:
		dbLoc = path.Join(dbLoc, "shard", fmt.Sprint(cid), "consensus")
		infoLoc = path.Join(dbLoc, "shard", fmt.Sprint(cid), "info")
	case ShardTransacton:
		dbLoc = path.Join(dbLoc, "shard", fmt.Sprint(cid), "transaction")
		infoLoc = path.Join(dbLoc, "shard", fmt.Sprint(cid), "info")
	case ShardFeature:
		dbLoc = path.Join(dbLoc, "shard", fmt.Sprint(cid), "feature")
		infoLoc = path.Join(dbLoc, "shard", fmt.Sprint(cid), "info")
	case ShardReward:
		dbLoc = path.Join(dbLoc, "shard", fmt.Sprint(cid), "reward")
		infoLoc = path.Join(dbLoc, "shard", fmt.Sprint(cid), "info")
	}
	infoFD, _ := os.Open(infoLoc)
	infoB, _ := ioutil.ReadAll(infoFD)
	info := CheckpointInfo{}
	if len(infoB) > 0 {
		height := strings.Split(string(infoB), "-")[0]
		hash := strings.Split(string(infoB), "-")[1]
		info.Hash = hash
		info.Height, _ = strconv.ParseInt(height, 10, 64)
	} else {
		return CheckpointInfo{}, nil
	}

	ff, _ := flatfile.NewFlatFile(dbLoc, 5000)
	return info, ff
}

func (s *BootstrapManager) backupShard(name string, bestView *ShardBestState) {
	consensusDB := bestView.GetCopiedConsensusStateDB()
	txDB := bestView.GetCopiedTransactionStateDB()
	featureDB := bestView.GetCopiedFeatureStateDB()
	rewardDB := bestView.GetShardRewardStateDB()

	fd, _ := os.OpenFile(path.Join(name, "beacon", "shard", fmt.Sprint(bestView.ShardID), "info"), os.O_RDWR, 0666)
	fd.WriteString(fmt.Sprintf("%v-%v", bestView.ShardHeight, bestView.Hash().String()))

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
	fd, _ := os.OpenFile(path.Join(name, "beacon", "info"), os.O_RDWR, 0666)
	fd.WriteString(fmt.Sprintf("%v-%v", bestView.BeaconHeight, bestView.Hash().String()))

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
