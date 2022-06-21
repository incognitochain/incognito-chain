package syncker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/metrics/monitor"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/peerv2"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	configpkg "github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/wire"
)

const MAX_S2B_BLOCK = 90
const MAX_CROSSX_BLOCK = 10

type SynckerManagerConfig struct {
	Network    Network
	Blockchain *blockchain.BlockChain
	Consensus  peerv2.ConsensusData
	MiningKey  string
	CQuit      chan struct{}
}

type SynckerManager struct {
	isEnabled             bool //0 > stop, 1: running
	config                *SynckerManagerConfig
	BeaconSyncProcess     *BeaconSyncProcess
	ShardSyncProcess      map[int]*ShardSyncProcess
	CrossShardSyncProcess map[int]*CrossShardSyncProcess
	Blockchain            *blockchain.BlockChain
	beaconPool            *BlkPool
	shardPool             map[int]*BlkPool
	crossShardPool        map[int]*BlkPool
}

func NewSynckerManager() *SynckerManager {
	s := &SynckerManager{
		ShardSyncProcess:      make(map[int]*ShardSyncProcess),
		shardPool:             make(map[int]*BlkPool),
		CrossShardSyncProcess: make(map[int]*CrossShardSyncProcess),
		crossShardPool:        make(map[int]*BlkPool),
	}
	return s
}

func (synckerManager *SynckerManager) Init(config *SynckerManagerConfig) {
	synckerManager.config = config

	//check preload beacon
	preloadAddr := configpkg.Config().PreloadAddress
	if preloadAddr != "" {
		if err := preloadDatabase(-1, int(config.Blockchain.BeaconChain.GetEpoch()), preloadAddr, config.Blockchain.GetBeaconChainDatabase(), config.Blockchain.GetBTCHeaderChain()); err != nil {
			fmt.Println(err)
			Logger.Infof("Preload beacon fail!")
		} else {
			config.Blockchain.RestoreBeaconViews()
		}
	}

	if os.Getenv("FULLNODE") == "1" {
		synckerManager.config.Network.SetSyncMode("fullnode")
	}

	//init beacon sync process
	synckerManager.BeaconSyncProcess = NewBeaconSyncProcess(
		synckerManager.config.Network,
		synckerManager.config.Blockchain,
		synckerManager.config.Blockchain.BeaconChain,
		config.CQuit,
	)
	synckerManager.beaconPool = synckerManager.BeaconSyncProcess.beaconPool

	//init shard sync process
	for _, chain := range synckerManager.config.Blockchain.ShardChain {
		sid := chain.GetShardID()
		synckerManager.ShardSyncProcess[sid] = NewShardSyncProcess(
			sid, synckerManager.config.Network,
			synckerManager.config.Blockchain,
			synckerManager.config.Blockchain.BeaconChain,
			chain, synckerManager.config.Consensus, config.CQuit,
		)
		synckerManager.shardPool[sid] = synckerManager.ShardSyncProcess[sid].shardPool
		synckerManager.CrossShardSyncProcess[sid] = synckerManager.ShardSyncProcess[sid].crossShardSyncProcess
		synckerManager.crossShardPool[sid] = synckerManager.CrossShardSyncProcess[sid].crossShardPool
	}
	synckerManager.Blockchain = config.Blockchain

	//watch commitee change
	go synckerManager.manageSyncProcess()
}

func (synckerManager *SynckerManager) Start() {
	synckerManager.isEnabled = true
}

func (synckerManager *SynckerManager) Stop() {
	synckerManager.isEnabled = false
	synckerManager.BeaconSyncProcess.stop()
	for _, chain := range synckerManager.ShardSyncProcess {
		chain.stop()
	}
}

func (s *SynckerManager) InsertCrossShardBlock(blk *types.CrossShardBlock) {
	s.CrossShardSyncProcess[int(blk.ToShardID)].InsertCrossShardBlock(blk)
}

// periodically check user commmittee status, enable shard sync process if needed (beacon always start)
func (synckerManager *SynckerManager) manageSyncProcess() {
	defer time.AfterFunc(time.Second*5, synckerManager.manageSyncProcess)

	syncStat := synckerManager.GetSyncStats()
	monitor.SetGlobalParam("SYNC_STAT", syncStat)

	//check if enable
	if !synckerManager.isEnabled || synckerManager.config == nil {
		return
	}

	chainValidator := synckerManager.config.Consensus.GetOneValidatorForEachConsensusProcess()

	if beaconChain, ok := chainValidator[-1]; ok {
		synckerManager.BeaconSyncProcess.isCommittee = beaconChain.State.Role == common.CommitteeRole
	}

	preloadAddr := configpkg.Config().PreloadAddress
	synckerManager.BeaconSyncProcess.start()

	if time.Now().Unix()-synckerManager.Blockchain.GetBeaconBestState().BestBlock.GetProduceTime() > 4*60*60 {
		lastInsertTime := synckerManager.BeaconSyncProcess.lastInsert
		if lastInsertTime == "" {
			lastInsertTime = "N/A (node restart)"
		}
		Logger.Infof("Beacon is syncing ... last time insert was %v", lastInsertTime)
	}

	wg := sync.WaitGroup{}
	wantedShard := synckerManager.config.Blockchain.GetWantedShard(synckerManager.BeaconSyncProcess.isCommittee)
	for chainID := range chainValidator {
		wantedShard[byte(chainID)] = struct{}{}
	}
	for sid, syncProc := range synckerManager.ShardSyncProcess {
		wg.Add(1)
		go func(sid int, syncProc *ShardSyncProcess) {
			defer wg.Done()

			//only start shard when beacon seem to be almost finish (sync up to block of 4 hours ago) and not in relay shards
			if time.Now().Unix()-synckerManager.Blockchain.GetBeaconBestState().BestBlock.GetProduceTime() > 4*60*60 {
				shouldRelay := false
				for _, relaySID := range synckerManager.config.Blockchain.GetConfig().RelayShards {
					if sid == int(relaySID) {
						shouldRelay = true
					}
				}
				if !shouldRelay {
					return
				}
			}

			if _, ok := wantedShard[byte(sid)]; ok {
				//check preload shard
				if preloadAddr != "" {
					if syncProc.status != RUNNING_SYNC { //run only when start
						if err := preloadDatabase(sid, int(syncProc.Chain.GetEpoch()), preloadAddr, synckerManager.config.Blockchain.GetShardChainDatabase(byte(sid)), nil); err != nil {
							fmt.Println(err)
							Logger.Infof("Preload shard %v fail!", sid)
						} else {
							synckerManager.config.Blockchain.RestoreShardViews(byte(sid))
						}
					}
				}
				syncProc.start()
			} else {
				syncProc.stop()
			}
			if chain, ok := chainValidator[sid]; ok {
				syncProc.isCommittee = chain.State.Role == common.CommitteeRole || chain.State.Role == common.PendingRole
			}

		}(sid, syncProc)
	}
	wg.Wait()

}

//Process incomming broadcast block
func (synckerManager *SynckerManager) ReceiveBlock(block interface{}, previousValidationData string, peerID string) {
	switch block.(type) {
	case *types.BeaconBlock:
		beaconBlock := block.(*types.BeaconBlock)
		Logger.Infof("syncker: receive beacon block %d \n", beaconBlock.GetHeight())
		//create fake s2b pool peerstate
		if synckerManager.BeaconSyncProcess != nil {
			synckerManager.beaconPool.AddBlock(beaconBlock)
			synckerManager.BeaconSyncProcess.beaconPeerStateCh <- &wire.MessagePeerState{
				Beacon: wire.ChainState{
					Timestamp: beaconBlock.Header.Timestamp,
					BlockHash: *beaconBlock.Hash(),
					Height:    beaconBlock.GetHeight(),
				},
				SenderID:  peerID,
				Timestamp: time.Now().Unix(),
			}
		}

	case *types.ShardBlock:

		shardBlock := block.(*types.ShardBlock)
		Logger.Infof("syncker: receive shard block %d \n", shardBlock.GetHeight())
		if synckerManager.shardPool[shardBlock.GetShardID()] != nil {
			synckerManager.shardPool[shardBlock.GetShardID()].AddBlock(shardBlock)
			synckerManager.shardPool[shardBlock.GetShardID()].AddPreviousValidationData(shardBlock.GetPrevHash(), previousValidationData)
			if synckerManager.ShardSyncProcess[shardBlock.GetShardID()] != nil {
				synckerManager.ShardSyncProcess[shardBlock.GetShardID()].shardPeerStateCh <- &wire.MessagePeerState{
					Shards: map[byte]wire.ChainState{
						byte(shardBlock.GetShardID()): {
							Timestamp: shardBlock.Header.Timestamp,
							BlockHash: *shardBlock.Hash(),
							Height:    shardBlock.GetHeight(),
						},
					},
					SenderID:  peerID,
					Timestamp: time.Now().Unix(),
				}
			}
		}

	case *types.CrossShardBlock:
		crossShardBlock := block.(*types.CrossShardBlock)
		if synckerManager.CrossShardSyncProcess[int(crossShardBlock.ToShardID)] != nil {
			Logger.Infof("crossdebug: receive block from %d to %d (%synckerManager)\n", crossShardBlock.Header.ShardID, crossShardBlock.ToShardID, crossShardBlock.Hash().String())
			synckerManager.crossShardPool[int(crossShardBlock.ToShardID)].AddBlock(crossShardBlock)
		}
	}
}

//Process incomming broadcast peerstate
func (synckerManager *SynckerManager) ReceivePeerState(peerState *wire.MessagePeerState) {
	//beacon
	if peerState.Beacon.Height != 0 && synckerManager.BeaconSyncProcess != nil {
		synckerManager.BeaconSyncProcess.beaconPeerStateCh <- peerState
	}
	//shard
	for sid := range peerState.Shards {
		if synckerManager.ShardSyncProcess[int(sid)] != nil {
			// b, _ := json.Marshal(peerState)
			// fmt.Println("[debugshard]: receive peer state", string(b))
			synckerManager.ShardSyncProcess[int(sid)].shardPeerStateCh <- peerState
		}

	}
}

//Get Crossshard Block for creating shardblock block
func (synckerManager *SynckerManager) GetCrossShardBlocksForShardProducer(curView *blockchain.ShardBestState, limit map[byte][]uint64) map[byte][]interface{} {
	//get last confirm crossshard -> process request until retrieve info
	res := make(map[byte][]interface{})
	toShard := curView.ShardID
	lastRequestCrossShard := make(map[byte]uint64)
	for index, key := range curView.BestCrossShard {
		lastRequestCrossShard[index] = key
	}

	bc := synckerManager.config.Blockchain
	for i := 0; i < synckerManager.config.Blockchain.GetActiveShardNumber(); i++ {
		for {
			if i == int(toShard) {
				break
			}

			//if limit has 0 length, we should break now
			if limit != nil && len(res[byte(i)]) >= len(limit[byte(i)]) {
				break
			}

			requestHeight := lastRequestCrossShard[byte(i)]
			nextCrossShardInfo := synckerManager.config.Blockchain.FetchNextCrossShard(i, int(toShard), requestHeight)
			if nextCrossShardInfo == nil {
				break
			}
			if requestHeight == nextCrossShardInfo.NextCrossShardHeight {
				break
			}

			Logger.Info("nextCrossShardInfo.NextCrossShardHeight", i, toShard, requestHeight, nextCrossShardInfo)

			beaconHash, _ := common.Hash{}.NewHashFromStr(nextCrossShardInfo.ConfirmBeaconHash)
			beaconBlock, _, err := synckerManager.Blockchain.GetBeaconBlockByHash(*beaconHash)
			if err != nil {
				Logger.Errorf("Get beacon block by hash %v failed\n", beaconHash.String())
				break
			}

			beaconFinalView := bc.BeaconChain.FinalView().(*blockchain.BeaconBestState)

			Logger.Infof("beaconFinalView: %v\n", beaconFinalView.Hash().String())
			for _, shardState := range beaconBlock.Body.ShardState[byte(i)] {
				if shardState.Height == nextCrossShardInfo.NextCrossShardHeight {
					if synckerManager.crossShardPool[int(toShard)].HasHash(shardState.Hash) {
						//validate crossShardBlock before add to result
						blkXShard := synckerManager.crossShardPool[int(toShard)].GetBlock(shardState.Hash)
						isValid := types.VerifyCrossShardBlockUTXO(blkXShard.(*types.CrossShardBlock))
						if !isValid {
							Logger.Error("Validate Crossshard block body fail", blkXShard.GetHeight(), blkXShard.Hash())
							return nil
						}
						// TODO: @committees
						//For releasing beacon nodes and re verify cross shard blocks from beacon
						//Use committeeFromBlock field for getting committees
						if beaconFinalView.CommitteeStateVersion() == committeestate.SELF_SWAP_SHARD_VERSION {
							beaconConsensusRootHash, err := bc.GetBeaconConsensusRootHash(bc.GetBeaconBestState(), beaconBlock.GetHeight()-1)
							if err != nil {
								Logger.Error("Cannot get beacon consensus root hash from block ", beaconBlock.GetHeight()-1)
								return nil
							}
							beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(bc.GetBeaconChainDatabase()))
							committee := statedb.GetOneShardCommittee(beaconConsensusStateDB, byte(i))
							err = bc.ShardChain[byte(i)].ValidateBlockSignatures(blkXShard.(types.BlockInterface), committee)
							if err != nil {
								Logger.Error("Validate crossshard block fail", blkXShard.GetHeight(), blkXShard.Hash())
								return nil
							}
						}
						//add to result list
						res[byte(i)] = append(res[byte(i)], blkXShard)
						//has block in pool, update request pointer
						lastRequestCrossShard[byte(i)] = nextCrossShardInfo.NextCrossShardHeight
					}
					break
				}
			}

			//cannot append crossshard for a shard (no block in pool, validate error) => break process for this shard
			if requestHeight == lastRequestCrossShard[byte(i)] {
				break
			}

			if len(res[byte(i)]) >= MAX_CROSSX_BLOCK {
				break
			}
		}
	}
	return res
}

//Get Crossshard Block for validating shardblock block
func (synckerManager *SynckerManager) GetCrossShardBlocksForShardValidator(curView *blockchain.ShardBestState, list map[byte][]uint64) (map[byte][]interface{}, error) {
	toShard := curView.ShardID
	crossShardPoolLists := synckerManager.GetCrossShardBlocksForShardProducer(curView, list)

	missingBlocks := compareListsByHeight(crossShardPoolLists, list)
	// synckerManager.config.Server.
	if len(missingBlocks) > 0 {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		synckerManager.StreamMissingCrossShardBlock(ctx, toShard, missingBlocks)
		//Logger.Info("debug finish stream missing crossX block")

		crossShardPoolLists = synckerManager.GetCrossShardBlocksForShardProducer(curView, list)
		//Logger.Info("get crosshshard block for shard producer", crossShardPoolLists)
		missingBlocks = compareListsByHeight(crossShardPoolLists, list)

		if len(missingBlocks) > 0 {
			return nil, errors.New("Unable to sync required block in time")
		}
	}

	for sid, heights := range list {
		if len(crossShardPoolLists[sid]) != len(heights) {
			return nil, fmt.Errorf("CrossShard list not match sid:%v pool:%v producer:%v", sid, len(crossShardPoolLists[sid]), len(heights))
		}
	}

	return crossShardPoolLists, nil
}

//Stream Missing CrossShard Block
func (synckerManager *SynckerManager) StreamMissingCrossShardBlock(ctx context.Context, toShard byte, missingBlock map[byte][]uint64) {
	for fromShard, missingHeight := range missingBlock {
		//fmt.Println("debug stream missing crossshard block", int(fromShard), int(toShard), missingHeight)
		ch, err := synckerManager.config.Network.RequestCrossShardBlocksViaStream(ctx, "", int(fromShard), int(toShard), missingHeight)
		if err != nil {
			fmt.Println("Syncker: create channel fail")
			return
		}
		//receive
		for {
			select {
			case blk := <-ch:
				if !isNil(blk) {
					Logger.Infof("Receive crosshard block from shard %v ->  %v, hash %v", fromShard, toShard, blk.(types.BlockPoolInterface).Hash().String())
					synckerManager.crossShardPool[int(toShard)].AddBlock(blk.(types.BlockPoolInterface))
				} else {
					//Logger.Info("Block is nil, break stream")
					return
				}
			}
		}
	}
}

////Get Crossshard Block for validating shardblock block
//func (synckerManager *SynckerManager) GetCrossShardBlocksForShardValidatorByHash(toShard byte, list map[byte][]common.Hash) (map[byte][]interface{}, error) {
//	crossShardPoolLists := synckerManager.GetCrossShardBlocksForShardProducer(toShard)
//
//	missingBlocks := compareLists(crossShardPoolLists, list)
//	// synckerManager.config.Server.
//	if len(missingBlocks) > 0 {
//		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
//		synckerManager.StreamMissingCrossShardBlockByHash(ctx, toShard, missingBlocks)
//		fmt.Println("debug finish stream missing s2b block")
//
//		crossShardPoolLists = synckerManager.GetCrossShardBlocksForShardProducer(toShard)
//		missingBlocks = compareLists(crossShardPoolLists, list)
//
//		if len(missingBlocks) > 0 {
//			return nil, errors.New("Unable to sync required block in time")
//		}
//	}
//	return crossShardPoolLists, nil
//}
//
////Stream Missing CrossShard Block
//func (synckerManager *SynckerManager) StreamMissingCrossShardBlockByHash(ctx context.Context, toShard byte, missingBlock map[byte][]common.Hash) {
//	fmt.Println("debug stream missing crossshard block", missingBlock)
//	wg := sync.WaitGroup{}
//	for i, v := range missingBlock {
//		wg.Add(1)
//		go func(sid byte, list []common.Hash) {
//			defer wg.Done()
//			hashes := [][]byte{}
//			for _, h := range list {
//				hashes = append(hashes, h.Bytes())
//			}
//			ch, err := synckerManager.config.Node.RequestCrossShardBlocksByHashViaStream(ctx, "", int(sid), int(toShard), hashes)
//			if err != nil {
//				fmt.Println("Syncker: create channel fail")
//				return
//			}
//			//receive
//			for {
//				select {
//				case blk := <-ch:
//					if !isNil(blk) {
//						synckerManager.crossShardPool[int(toShard)].AddBlock(blk.(common.BlockPoolInterface))
//					} else {
//						return
//					}
//				}
//			}
//		}(i, v)
//	}
//	wg.Wait()
//}

//Sync missing beacon block  from a hash to our final view (skip if we already have)
func (synckerManager *SynckerManager) SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash) {
	requestHash := fromHash
	for {
		ch, err := synckerManager.config.Network.RequestBeaconBlocksByHashViaStream(ctx, peerID, [][]byte{requestHash.Bytes()})
		if err != nil {
			fmt.Println("[Monitor] Syncker: create channel fail")
			return
		}
		blk := <-ch
		if !isNil(blk) {
			if blk.(*types.BeaconBlock).GetHeight() <= synckerManager.config.Blockchain.BeaconChain.GetFinalViewHeight() {
				return
			}
			synckerManager.beaconPool.AddBlock(blk.(types.BlockPoolInterface))
			prevHash := blk.(*types.BeaconBlock).GetPrevHash()
			if v := synckerManager.config.Blockchain.BeaconChain.GetViewByHash(prevHash); v == nil {
				requestHash = prevHash
				continue
			}
		}
		return
	}
}

//Sync back missing shard block from a hash to our final views (skip if we already have)
func (synckerManager *SynckerManager) SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash) {
	requestHash := fromHash
	for {
		ch, err := synckerManager.config.Network.RequestShardBlocksByHashViaStream(ctx, peerID, int(sid), [][]byte{requestHash.Bytes()})
		if err != nil {
			fmt.Println("Syncker: create channel fail")
			return
		}
		blk := <-ch
		if !isNil(blk) {
			if blk.(*types.ShardBlock).GetHeight() <= synckerManager.config.Blockchain.ShardChain[sid].GetFinalViewHeight() {
				return
			}
			synckerManager.shardPool[int(sid)].AddBlock(blk.(types.BlockPoolInterface))
			prevHash := blk.(*types.ShardBlock).GetPrevHash()
			if v := synckerManager.config.Blockchain.ShardChain[sid].GetViewByHash(prevHash); v == nil {
				requestHash = prevHash
				continue
			}
		}
		return

	}
}

//Get Status Function
type SyncInfo struct {
	IsSync      bool
	LastInsert  string
	BlockHeight uint64
	BlockTime   string
	BlockHash   string
}

type SynckerStats struct {
	Beacon SyncInfo
	Shard  map[int]*SyncInfo
}

func (synckerManager *SynckerManager) GetSyncStats() SynckerStats {
	info := SynckerStats{}
	info.Beacon = SyncInfo{
		IsSync:      synckerManager.BeaconSyncProcess.status == RUNNING_SYNC,
		LastInsert:  synckerManager.BeaconSyncProcess.lastInsert,
		BlockHeight: synckerManager.BeaconSyncProcess.chain.GetBestViewHeight(),
		BlockHash:   synckerManager.BeaconSyncProcess.chain.GetBestViewHash(),
		BlockTime:   time.Unix(synckerManager.config.Blockchain.GetBeaconBestState().GetBlockTime(), 0).Format("2006-01-02T15:04:05-0700"),
	}
	info.Shard = make(map[int]*SyncInfo)
	for k, v := range synckerManager.ShardSyncProcess {
		info.Shard[k] = &SyncInfo{
			IsSync:      v.status == RUNNING_SYNC,
			LastInsert:  v.lastInsert,
			BlockHeight: v.Chain.GetBestViewHeight(),
			BlockHash:   v.Chain.GetBestViewHash(),
			BlockTime:   time.Unix(synckerManager.config.Blockchain.GetBestStateShard(byte(k)).GetBlockTime(), 0).Format("2006-01-02T15:04:05-0700"),
		}
	}
	return info
}

func (synckerManager *SynckerManager) IsChainReady(chainID int) bool {
	if chainID == -1 {
		return synckerManager.BeaconSyncProcess.isCatchUp
	} else if chainID >= 0 {
		return synckerManager.ShardSyncProcess[chainID].isCatchUp
	}
	return false
}

type TmpBlock struct {
	Height  uint64
	BlkHash *common.Hash
	PreHash common.Hash
	ShardID int
	Round   int
}

func (blk *TmpBlock) GetHeight() uint64 {
	return blk.Height
}

func (blk *TmpBlock) Hash() *common.Hash {
	return blk.BlkHash
}

func (blk *TmpBlock) GetPrevHash() common.Hash {
	return blk.PreHash
}

func (blk *TmpBlock) FullHashString() string {
	return blk.Hash().String()
}

func (blk *TmpBlock) GetShardID() int {
	return blk.ShardID
}
func (blk *TmpBlock) GetRound() int {
	return 1
}

func (synckerManager *SynckerManager) GetPoolInfo(poolType byte, sID int) []types.BlockPoolInterface {
	switch poolType {
	case BeaconPoolType:
		if synckerManager.BeaconSyncProcess != nil {
			if synckerManager.BeaconSyncProcess.beaconPool != nil {
				return synckerManager.BeaconSyncProcess.beaconPool.GetPoolInfo()
			}
		}
	case ShardPoolType:
		if syncProcess, ok := synckerManager.ShardSyncProcess[sID]; ok {
			if syncProcess.shardPool != nil {
				return syncProcess.shardPool.GetPoolInfo()
			}
		}
	case CrossShardPoolType:
		if syncProcess, ok := synckerManager.ShardSyncProcess[sID]; ok {
			if syncProcess.shardPool != nil {
				res := []types.BlockPoolInterface{}
				for fromSID, blksPool := range synckerManager.crossShardPool {
					for _, blk := range blksPool.GetBlockList() {
						res = append(res, &TmpBlock{
							Height:  blk.GetHeight(),
							BlkHash: blk.Hash(),
							PreHash: common.Hash{},
							ShardID: fromSID,
						})
					}
				}
				return res
			}
		}
	}
	return []types.BlockPoolInterface{}
}

func (synckerManager *SynckerManager) GetPoolLatestHeight(poolType byte, bestHash string, sID int) uint64 {
	switch poolType {
	case BeaconPoolType:
		if synckerManager.BeaconSyncProcess != nil {
			if synckerManager.BeaconSyncProcess.beaconPool != nil {
				return synckerManager.BeaconSyncProcess.beaconPool.GetLatestHeight(bestHash)
			}
		}
	case ShardPoolType:
		if syncProcess, ok := synckerManager.ShardSyncProcess[sID]; ok {
			if syncProcess.shardPool != nil {
				return syncProcess.shardPool.GetLatestHeight(bestHash)
			}
		}
	case CrossShardPoolType:
		//TODO
		return 0
	}
	return 0
}

func (synckerManager *SynckerManager) GetAllViewByHash(poolType byte, bestHash string, sID int) []types.BlockPoolInterface {
	switch poolType {
	case BeaconPoolType:
		if synckerManager.BeaconSyncProcess != nil {
			if synckerManager.BeaconSyncProcess.beaconPool != nil {
				return synckerManager.BeaconSyncProcess.beaconPool.GetAllViewByHash(bestHash)
			}
		}
	case ShardPoolType:
		if syncProcess, ok := synckerManager.ShardSyncProcess[sID]; ok {
			if syncProcess.shardPool != nil {
				return syncProcess.shardPool.GetAllViewByHash(bestHash)
			}
		}
	default:
		//TODO
		return []types.BlockPoolInterface{}
	}
	return []types.BlockPoolInterface{}
}
