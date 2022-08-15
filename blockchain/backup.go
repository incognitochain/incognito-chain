package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"io/ioutil"
	"log"
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
	blockchain    *BlockChain
	lastBackup    *BackupProcessInfo
	runningBackup *BackupProcessInfo
	lock          *sync.Mutex
}

type StateDBData struct {
	K []byte
	V []byte
}

func NewBackupManager(bc *BlockChain) *BackupManager {
	//read bootstrap dir and load lastBootstrap
	cfg := config.Config()
	fd, err := os.OpenFile(path.Join(path.Join(cfg.DataDir, cfg.DatabaseDir), "backupinfo"), os.O_RDONLY, 0666)
	if err != nil {
		return &BackupManager{bc, nil, nil, new(sync.Mutex)}
	}
	jsonStr, err := ioutil.ReadAll(fd)
	if err != nil {
		panic(err)
	}
	lastBackup := &BackupProcessInfo{}
	err = json.Unmarshal(jsonStr, lastBackup)
	if err != nil {
		panic(err)
	}
	log.Println(string(jsonStr))
	return &BackupManager{bc, lastBackup, nil, new(sync.Mutex)}
}

func (s *BackupManager) GetLastestBootstrap() BackupProcessInfo {
	return *s.lastBackup
}

const BackupInterval = 350 * 6 * 3 // days

func (s *BackupManager) Backup(backupHeight uint64) {
	s.lock.Lock()
	if s.runningBackup != nil {
		s.lock.Unlock()
		return
	}

	s.runningBackup = &BackupProcessInfo{}
	s.lock.Unlock()
	defer func() {
		s.runningBackup = nil
	}()

	//backup condition period
	if backupHeight < BackupInterval {
		return
	}
	if s.lastBackup != nil && s.lastBackup.BeaconView != nil && s.lastBackup.BeaconView.BeaconHeight+BackupInterval > backupHeight {
		return
	}
	bestState := s.blockchain.GetBeaconBestState()
	for sid, shardChain := range s.blockchain.ShardChain {
		if bestState.BestShardHeight[byte(sid)] > shardChain.GetFinalViewHeight() {
			log.Println("Not backup as shard not sync up")
			return
		}
	}

	cfg := config.Config()

	beaconFinalView := NewBeaconBestState()
	beaconFinalView.cloneBeaconBestStateFrom(s.blockchain.BeaconChain.FinalView().(*BeaconBestState))

	//update current status
	checkPoint := time.Now().Format(time.RFC3339)
	bootstrapInfo := &BackupProcessInfo{
		CheckpointName: checkPoint,
		BeaconView:     beaconFinalView,
		ShardView:      map[int]*ShardBestState{},
	}
	s.runningBackup = bootstrapInfo
	defer func() {
		s.runningBackup = nil
	}()

	//backup beacon then shard
	log.Println("backup beacon")
	backUpPath := path.Join(cfg.DataDir, cfg.DatabaseDir, checkPoint)
	s.backupBeacon(backUpPath, beaconFinalView)
	beaconFinalView.BestBlock = types.BeaconBlock{}

	//backup shard
	log.Println("backup shard")
	shardFinalView := map[int]*ShardBestState{}
	for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
	WAITING:
		finalView := s.blockchain.ShardChain[i].multiView.GetFinalView().(*ShardBestState)
		if finalView.BeaconHeight <= beaconFinalView.BeaconHeight {
			Logger.log.Infof("Waiting for confirm more beacon blocks ... shard confirm beacon: %v, beacon: %v", finalView.BeaconHeight, beaconFinalView.BeaconHeight)
			time.Sleep(time.Minute)
			goto WAITING
		}
		shardFinalView[i] = NewShardBestState()
		shardFinalView[i].cloneShardBestStateFrom(s.blockchain.ShardChain[i].multiView.GetFinalView().(*ShardBestState))
	}

	shardWG := sync.WaitGroup{}
	for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
		shardWG.Add(1)
		go func(sid int) {
			s.backupShard(backUpPath, shardFinalView[sid])
			bootstrapInfo.ShardView[sid] = shardFinalView[sid]
			shardWG.Done()
		}(i)
	}
	shardWG.Wait()
	//get smallest beacon height of committeefromblock
	bootstrapInfo.MinBeaconHeight = 1e9
	for _, view := range bootstrapInfo.ShardView {
		beaconHash := view.BestBlock.CommitteeFromBlock()
		view.BestBlock = nil
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
	bootstrapInfo.MinBeaconHeight-- //for get previous block

	//backup beacon block
	s.backupBeaconBlock(backUpPath, bootstrapInfo.MinBeaconHeight, beaconFinalView)

	//update final status
	s.lastBackup = bootstrapInfo
	fd, err := os.OpenFile(path.Join(backUpPath, "backupinfo"), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	jsonStr, err := json.Marshal(bootstrapInfo)
	if err != nil {
		panic(err)
	}
	fd.Truncate(0)
	n, _ := fd.Write(jsonStr)
	fmt.Println("update lastBackup", n)
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
	cfg := config.Config()
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
		dbLoc = path.Join(dbLoc, fmt.Sprintf("shard%v", cid), "tx")
	case ShardFeature:
		dbLoc = path.Join(dbLoc, fmt.Sprintf("shard%v", cid), "feature")
	case ShardReward:
		dbLoc = path.Join(dbLoc, fmt.Sprintf("shard%v", cid), "reward")
	}
	fmt.Println("GetBackupReader", dbLoc)
	ff, _ := flatfile.NewFlatFile(dbLoc, 5000)
	return ff
}

func (s *BackupManager) backupShard(name string, finalView *ShardBestState) {
	consensusDB := finalView.GetCopiedConsensusStateDB()
	txDB := finalView.GetCopiedTransactionStateDB()
	featureDB := finalView.GetCopiedFeatureStateDB()
	rewardDB := finalView.GetShardRewardStateDB()

	shardStateDB, _ := incdb.Open("leveldb", path.Join(name, fmt.Sprintf("shard%v", finalView.ShardID)))

	wg := sync.WaitGroup{}
	wg.Add(5)
	go s.backupShardBlock(name, finalView, &wg)
	go backupStateDB(consensusDB, shardStateDB, &wg)
	go backupStateDB(featureDB, shardStateDB, &wg)
	go backupStateDB(txDB, shardStateDB, &wg)
	go backupStateDB(rewardDB, shardStateDB, &wg)
	wg.Wait()
}

func (s *BackupManager) backupShardBlock(name string, finalView *ShardBestState, wg *sync.WaitGroup) {
	defer wg.Done()
	sid := finalView.GetShardID()
	blockStorage := NewBlockStorage(s.blockchain.GetShardChainDatabase(sid), path.Join(name, fmt.Sprintf("shard%v", sid), "blockstorage"), int(sid), true)
	for blkHeight := uint64(1); blkHeight <= finalView.ShardHeight; blkHeight++ {
		shardBlock, err := s.blockchain.GetShardBlockByHeightV1(finalView.GetShardHeight(), sid)
		if err != nil {
			panic(err)
		}
		if shardBlock.GetHeight()%1000 == 0 {
			log.Println("shard", sid, "save block ", shardBlock.GetHeight())
		}
		if err := blockStorage.StoreTXIndex(shardBlock); err != nil {
			panic(err)
		}

		if err := blockStorage.StoreBlock(shardBlock); err != nil {
			panic(err)
		}

		if err := blockStorage.StoreFinalizedShardBlock(shardBlock.GetHeight(), *shardBlock.Hash()); err != nil {
			panic(err)
		}
	}
}

func (s *BackupManager) backupBeacon(name string, finalView *BeaconBestState) {
	consensusDB := finalView.GetBeaconConsensusStateDB()
	featureDB := finalView.GetBeaconFeatureStateDB()
	rewardDB := finalView.GetBeaconRewardStateDB()
	slashDB := finalView.GetBeaconSlashStateDB()

	beaconStateDB, _ := incdb.Open("leveldb", path.Join(name, "beacon"))

	wg := sync.WaitGroup{}
	wg.Add(4)

	go backupStateDB(consensusDB, beaconStateDB, &wg)
	go backupStateDB(featureDB, beaconStateDB, &wg)
	go backupStateDB(rewardDB, beaconStateDB, &wg)
	go backupStateDB(slashDB, beaconStateDB, &wg)

	//store beacon finalview
	allViews := []*BeaconBestState{finalView}
	b, _ := json.Marshal(allViews)
	err := rawdbv2.StoreBeaconViews(beaconStateDB, b)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	wg.Wait()
}

func (s *BackupManager) backupBeaconBlock(name string, fromBlock uint64, finalView *BeaconBestState) {
	blockStorage := NewBlockStorage(s.blockchain.GetBeaconChainDatabase(), path.Join(name, "beacon", "blockstorage"), -1, true)
	committeeFromBlock := map[byte]map[common.Hash]bool{}
	for blkHeight := fromBlock; blkHeight <= finalView.BeaconHeight; blkHeight++ {
		beaconBlock, err := s.blockchain.GetBeaconBlockByHeightV1(blkHeight)
		if err != nil {
			panic(err)
		}
		if err := blockStorage.StoreBlock(beaconBlock); err != nil {
			panic(err)
		}

		for shardID, shardStates := range beaconBlock.Body.ShardState {
			for _, shardState := range shardStates {
				err := blockStorage.StoreBeaconConfirmShardBlockByHeight(shardID, shardState.Height, shardState.Hash)
				if err != nil {
					panic(err)
				}
			}
		}

		if err := blockStorage.StoreFinalizedBeaconBlock(beaconBlock.GetHeight(), *beaconBlock.Hash()); err != nil {
			panic(err)
		}

		for sid, shardStates := range beaconBlock.Body.ShardState {
			if _, ok := committeeFromBlock[sid]; !ok {
				committeeFromBlock[sid] = map[common.Hash]bool{}
			}
			for _, state := range shardStates {
				committeeFromBlock[sid][state.CommitteeFromBlock] = true
			}
		}
	}

	for sid, committeeFromBlockHash := range committeeFromBlock {
		for hash, _ := range committeeFromBlockHash {
			//stream committee from block and set to cache
			log.Println("stream", sid, hash.String())
			committees, err := s.blockchain.BeaconChain.CommitteesFromViewHashForShard(hash, sid)
			if err != nil {
				panic(err)
			}
			err = rawdbv2.StoreCacheCommitteeFromBlock(blockStorage.blockStorageDB, hash, int(sid), committees)
			if err != nil {
				panic(err)
			}
		}
	}
}

func backupStateDB(stateDB *statedb.StateDB, beaconStateDB incdb.Database, wg *sync.WaitGroup) {
	defer wg.Done()
	it := stateDB.GetIterator()
	batchData := beaconStateDB.NewBatch()
	if stateDB == nil {
		return
	}
	for it.Next(false, true, true) {
		diskvalue, err := stateDB.Database().TrieDB().DiskDB().Get(it.Key)
		if err != nil {
			continue
		}
		key := make([]byte, len(it.Key))
		copy(key, it.Key)
		batchData.Put(key, diskvalue)
		if batchData.ValueSize() > 10*1024*1024 {
			batchData.Write()
			batchData.Reset()
		}
	}
	if batchData.ValueSize() > 0 {
		batchData.Write()
		batchData.Reset()
	}
	if it.Err != nil {
		fmt.Println(it.Err)
		panic(it.Err)
	}
}
