package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

type BackupProcessInfo struct {
	CheckpointName  string
	BeaconView      *BeaconBestState
	ShardView       map[int]*ShardBestState
	MinBeaconHeight uint64
}

type BackupManager struct {
	blockchain       *BlockChain
	lastBootStrap    *BackupProcessInfo
	runningBootStrap *BackupProcessInfo
}

type StateDBData struct {
	K []byte
	V []byte
}

func NewBackupManager(bc *BlockChain) *BackupManager {
	//read bootstrap dir and load lastBootstrap
	cfg := config.LoadConfig()
	fd, err := os.OpenFile(path.Join(path.Join(cfg.DataDir, cfg.DatabaseDir), "backupinfo"), os.O_RDONLY, 0666)
	if err != nil {
		return &BackupManager{bc, nil, nil}
	}
	jsonStr, err := ioutil.ReadAll(fd)
	fmt.Println(string(jsonStr))
	if err != nil {
		panic(err)
	}
	lastBackup := &BackupProcessInfo{}
	err = json.Unmarshal(jsonStr, lastBackup)
	if err != nil {
		panic(err)
	}
	fmt.Println(lastBackup)
	return &BackupManager{bc, lastBackup, nil}
}

func (s *BackupManager) GetLastestBootstrap() BackupProcessInfo {
	return *s.lastBootStrap
}

const BackupInterval = 20

func (s *BackupManager) Backup(backupHeight uint64) {
	//backup condition period
	if backupHeight < BackupInterval {
		return
	}
	if s.lastBootStrap.BeaconView != nil && s.lastBootStrap.BeaconView.GetHeight()+BackupInterval > backupHeight {
		return
	}

	shardBestView := map[int]*ShardBestState{}
	beaconBestView := NewBeaconBestState()
	beaconBestView.cloneBeaconBestStateFrom(s.blockchain.GetBeaconBestState())
	beaconBestView.BestBlock = types.BeaconBlock{}

	checkPoint := time.Now().Format(time.RFC3339)
	defer func() {
		s.runningBootStrap = nil
	}()
	for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
		shardBestView[i] = NewShardBestState()
		shardBestView[i].cloneShardBestStateFrom(s.blockchain.GetBestStateShard(byte(i)))
		shardBestView[i].BestBlock = nil
	}

	//update current status
	bootstrapInfo := &BackupProcessInfo{
		CheckpointName: checkPoint,
		BeaconView:     beaconBestView,
		ShardView:      map[int]*ShardBestState{},
	}
	s.runningBootStrap = bootstrapInfo

	//backup beacon then shard
	cfg := config.LoadConfig()
	s.backupBeacon(path.Join(cfg.DataDir, cfg.DatabaseDir, checkPoint), beaconBestView)
	for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
		s.backupShard(path.Join(cfg.DataDir, cfg.DatabaseDir, checkPoint), shardBestView[i])
		bootstrapInfo.ShardView[i] = shardBestView[i]
	}

	//update final status
	s.lastBootStrap = bootstrapInfo
	fd, err := os.OpenFile(path.Join(path.Join(cfg.DataDir, cfg.DatabaseDir), "backupinfo"), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	//get smallest beacon height of committeefromblock
	bootstrapInfo.MinBeaconHeight = 1e9
	for _, view := range bootstrapInfo.ShardView {
		beaconHash := view.BestBlock.CommitteeFromBlock()
		if beaconHash.IsEqual(&common.Hash{}) {
			continue
		}
		blk, _, err := s.blockchain.GetBeaconBlockByHash(beaconHash)
		if err != nil {
			panic(err)
		}
		if bootstrapInfo.MinBeaconHeight > blk.GetHeight() {
			bootstrapInfo.MinBeaconHeight = blk.GetHeight()
		}
	}

	jsonStr, err := json.Marshal(bootstrapInfo)
	if err != nil {
		panic(err)
	}
	n, _ := fd.Write(jsonStr)
	fmt.Println("update lastBootStrap", n)
	fd.Close()

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

func (s *BackupManager) GetBackupReader(checkpoint string, cid int, dbType int) *flatfile.FlatFileManager {
	cfg := config.LoadConfig()
	dbLoc := path.Join(cfg.DataDir, cfg.DatabaseDir, checkpoint)
	switch dbType {
	case BeaconConsensus:
		dbLoc = path.Join(dbLoc, "beacon", "consensus")
	case BeaconFeature:
		dbLoc = path.Join(dbLoc, "beacon", "feature")
	case BeaconReward:
		dbLoc = path.Join(dbLoc, "beacon", "reward")
	case BeaconSlash:
		dbLoc = path.Join(dbLoc, "beacon", "slash")
	case ShardConsensus:
		dbLoc = path.Join(dbLoc, fmt.Sprintf("shard%v", cid), "consensus")
	case ShardTransacton:
		dbLoc = path.Join(dbLoc, fmt.Sprintf("shard%v", cid), "transaction")
	case ShardFeature:
		dbLoc = path.Join(dbLoc, fmt.Sprintf("shard%v", cid), "feature")
	case ShardReward:
		dbLoc = path.Join(dbLoc, fmt.Sprintf("shard%v", cid), "reward")
	}
	fmt.Println("GetBackupReader", dbLoc)
	ff, _ := flatfile.NewFlatFile(dbLoc, 5000)
	return ff
}

func (s *BackupManager) backupShard(name string, bestView *ShardBestState) {
	consensusDB := bestView.GetCopiedConsensusStateDB()
	txDB := bestView.GetCopiedTransactionStateDB()
	featureDB := bestView.GetCopiedFeatureStateDB()
	rewardDB := bestView.GetShardRewardStateDB()

	consensusFF, _ := flatfile.NewFlatFile(path.Join(name, fmt.Sprintf("shard%v", bestView.ShardID), "consensus"), 5000)
	featureFF, _ := flatfile.NewFlatFile(path.Join(name, fmt.Sprintf("shard%v", bestView.ShardID), "feature"), 5000)
	txFF, _ := flatfile.NewFlatFile(path.Join(name, fmt.Sprintf("shard%v", bestView.ShardID), "tx"), 5000)
	rewardFF, _ := flatfile.NewFlatFile(path.Join(name, fmt.Sprintf("shard%v", bestView.ShardID), "reward"), 5000)

	wg := sync.WaitGroup{}
	wg.Add(4)

	go backupStateDB(consensusDB, consensusFF, &wg)
	go backupStateDB(featureDB, featureFF, &wg)
	go backupStateDB(txDB, txFF, &wg)
	go backupStateDB(rewardDB, rewardFF, &wg)
	wg.Wait()
}

func (s *BackupManager) backupBeacon(name string, bestView *BeaconBestState) {
	consensusDB := bestView.GetBeaconConsensusStateDB()
	featureDB := bestView.GetBeaconFeatureStateDB()
	rewardDB := bestView.GetBeaconRewardStateDB()
	slashDB := bestView.GetBeaconSlashStateDB()

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
	if stateDB == nil {
		return
	}
	for it.Next(false, true, true) {
		diskvalue, err := stateDB.Database().TrieDB().DiskDB().Get(it.Key)
		if err != nil {
			continue
		}
		//fmt.Println(it.Key, len(diskvalue))
		key := make([]byte, len(it.Key))
		copy(key, it.Key)
		data := StateDBData{key, diskvalue}
		batchData = append(batchData, data)
		if len(batchData) == 1000 {
			totalLen += 1000
			buf := new(bytes.Buffer)
			enc := gob.NewEncoder(buf)
			err := enc.Encode(batchData)
			if err != nil {
				panic(err)
			}
			_, err = ff.Append(buf.Bytes())
			if err != nil {
				panic(err)
			}
			batchData = []StateDBData{}
		}
	}
	if len(batchData) > 0 {
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		enc.Encode(batchData)
		_, err := ff.Append(buf.Bytes())
		if err != nil {
			panic(err)
		}
	}
}
