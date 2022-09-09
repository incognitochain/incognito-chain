package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"golang.org/x/sync/semaphore"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
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

	lock            *sync.Mutex
	downloading     map[string]int
	donwloadingLock *sync.Mutex
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
		return &BackupManager{bc, nil, nil, new(sync.Mutex), make(map[string]int), new(sync.Mutex)}
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
	return &BackupManager{bc, lastBackup, nil, new(sync.Mutex), make(map[string]int), new(sync.Mutex)}
}

func (s *BackupManager) GetLastestBootstrap() BackupProcessInfo {
	return *s.lastBackup
}

func (s *BackupManager) StartDownload(checkpoint string) {
	s.donwloadingLock.Lock()
	defer s.donwloadingLock.Unlock()
	s.downloading[checkpoint]++
}
func (s *BackupManager) StopDownload(checkpoint string) {
	s.donwloadingLock.Lock()
	defer s.donwloadingLock.Unlock()
	s.downloading[checkpoint]--
}

func (s *BackupManager) Backup(backupHeight uint64) {
	s.lock.Lock()
	if s.runningBackup != nil {
		s.lock.Unlock()
		return
	}
	cfg := config.Config()

	//remove old backup
	filepath.Walk(path.Join(cfg.DataDir, cfg.DatabaseDir), func(dirPath string, info fs.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if _, err := time.Parse(time.RFC3339, info.Name()); err == nil {
			if _, err := os.Stat(path.Join(dirPath, "done")); err != nil {
				t1 := time.Now()
				os.RemoveAll(dirPath)
				log.Println("remove unfinished backup folder", dirPath, time.Since(t1).Seconds())
			} else {
				s.donwloadingLock.Lock()
				if s.lastBackup.CheckpointName != info.Name() && s.downloading[info.Name()] == 0 {
					log.Println("remove old backup folder", dirPath)
					os.RemoveAll(dirPath)
				}
				s.donwloadingLock.Unlock()
			}
		}
		return nil
	})

	s.runningBackup = &BackupProcessInfo{}
	s.lock.Unlock()
	defer func() {
		s.runningBackup = nil
	}()

	BackupInterval := uint64(60 / s.blockchain.BeaconChain.GetBestView().GetCurrentTimeSlot() * 60 * 24 * 3)
	if config.Config().BackupInterval != 0 {
		BackupInterval = uint64(60 / s.blockchain.BeaconChain.GetBestView().GetCurrentTimeSlot() * 60 * config.Config().BackupInterval)
	}

	//backup condition period
	if s.lastBackup != nil && s.lastBackup.BeaconView != nil && s.lastBackup.BeaconView.BeaconHeight+BackupInterval > backupHeight {
		log.Println("view not satisfy to backup")
		return
	}
	bestState := s.blockchain.GetBeaconBestState()
	for sid, shardChain := range s.blockchain.ShardChain {
		if bestState.BestShardHeight[byte(sid)]-3 > shardChain.GetFinalViewHeight() {
			log.Println("Not backup as shard not sync up")
			return
		}
	}

	beaconFinalView := NewBeaconBestState()
	beaconFinalView.cloneBeaconBestStateFrom(s.blockchain.BeaconChain.FinalView().(*BeaconBestState))

	//update current status
	checkPoint := time.Now().Format(time.RFC3339)
	backupInfo := &BackupProcessInfo{
		CheckpointName: checkPoint,
		BeaconView:     beaconFinalView,
		ShardView:      map[int]*ShardBestState{},
	}
	s.runningBackup = backupInfo
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
	sem := semaphore.NewWeighted(4)
	for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
		shardWG.Add(1)
		sem.Acquire(context.Background(), 1)
		go func(sid int) {
			s.backupShard(backUpPath, shardFinalView[sid])
			backupInfo.ShardView[sid] = shardFinalView[sid]
			shardWG.Done()
			sem.Release(1)
		}(i)
	}
	shardWG.Wait()
	//get smallest beacon height of committeefromblock
	backupInfo.MinBeaconHeight = 1e9
	for _, view := range backupInfo.ShardView {
		beaconHash := view.BestBlock.CommitteeFromBlock()
		view.BestBlock = nil
		if beaconHash.IsEqual(&common.Hash{}) {
			continue
		}
		blk, _, err := s.blockchain.GetBeaconBlockByHash(beaconHash)
		if err != nil {
			panic(err)
		}
		if backupInfo.MinBeaconHeight > blk.GetHeight() {
			backupInfo.MinBeaconHeight = blk.GetHeight()
		}
	}
	backupInfo.MinBeaconHeight-- //for get previous block

	//backup beacon block
	s.backupBeaconBlock(backUpPath, backupInfo.MinBeaconHeight, beaconFinalView)

	//create done file
	fd, err := os.OpenFile(path.Join(backUpPath, "done"), os.O_CREATE|os.O_RDWR, 0666)
	fd.Close()

	//update final status
	s.lastBackup = backupInfo
	fd, err = os.OpenFile(path.Join(cfg.DataDir, cfg.DatabaseDir, "backupinfo"), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	jsonStr, err := json.Marshal(backupInfo)
	if err != nil {
		panic(err)
	}
	fd.Truncate(0)
	n, _ := fd.Write(jsonStr)
	fmt.Println("update lastBackup", n)
	fd.Close()
}

type CheckpointInfo struct {
	Hash   string
	Height int64
}

func (s *BackupManager) GetFileID(cid int, blockheight uint64) uint64 {
	switch cid {
	case -1:
		return blockheight / s.blockchain.BeaconChain.BlockStorage.flatfile.FileSize()
	default:
		return blockheight / s.blockchain.ShardChain[cid].BlockStorage.flatfile.FileSize()
	}
}

func (s *BackupManager) backupShard(name string, finalView *ShardBestState) {
	consensusDB := finalView.GetCopiedConsensusStateDB()
	txDB := finalView.GetCopiedTransactionStateDB()
	featureDB := finalView.GetCopiedFeatureStateDB()
	rewardDB := finalView.GetShardRewardStateDB()

	shardKeyValueDB, _ := incdb.Open("leveldb", path.Join(name, fmt.Sprintf("shard%v", finalView.ShardID)))

	wg := sync.WaitGroup{}
	wg.Add(5)
	go s.backupShardBlock(name, finalView, &wg)
	go backupStateDB(consensusDB, shardKeyValueDB, &wg)
	go backupStateDB(featureDB, shardKeyValueDB, &wg)
	go backupStateDB(txDB, shardKeyValueDB, &wg)
	go backupStateDB(rewardDB, shardKeyValueDB, &wg)
	wg.Wait()
}

func (s *BackupManager) GetBackupReader(checkpoint string, cid int) string {
	cfg := config.Config()
	dbLoc := path.Join(cfg.DataDir, cfg.DatabaseDir, checkpoint)
	switch cid {
	case -1:
		dbLoc = path.Join(dbLoc, "beacon")
	default:
		dbLoc = path.Join(dbLoc, fmt.Sprintf("shard%v", cid))
	}
	return dbLoc
}

func (s *BackupManager) backupShardBlock(name string, finalView *ShardBestState, wg *sync.WaitGroup) {
	defer wg.Done()
	sid := finalView.GetShardID()
	blockStorage := NewBlockStorage(nil, path.Join(name, fmt.Sprintf("shard%v", sid), "blockstorage"), int(sid), true)
	for blkHeight := uint64(1); blkHeight <= finalView.ShardHeight; blkHeight++ {
		shardBlock, err := s.blockchain.GetShardBlockByHeightV1(blkHeight, sid)
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
	blockStorage := NewBlockStorage(nil, path.Join(name, "beacon", "blockstorage"), -1, true)
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

func recheck(stateDB incdb.Database, roothash common.Hash) {
	txDB, err := statedb.NewWithPrefixTrie(roothash, statedb.NewDatabaseAccessWarper(stateDB))
	if err != nil {
		panic(fmt.Sprintf("Something wrong when init txDB"))
	}
	if err := txDB.Recheck(); err != nil {
		fmt.Println("Recheck roothash fail!", roothash.String())
		panic("recheck fail!")
	}
}

func backupStateDB(stateDB *statedb.StateDB, kvDB incdb.Database, wg *sync.WaitGroup) {
	defer wg.Done()
	it := stateDB.GetIterator()
	batchData := kvDB.NewBatch()
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
		if batchData.ValueSize() > 5*1024*1024 {
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

	//recheck stateDB
	rootHash, err := stateDB.Commit(false)
	if err != nil {
		panic(err)
	}
	recheck(kvDB, rootHash)
}
