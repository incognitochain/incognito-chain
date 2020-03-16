package syncker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
)

type SynckerManagerConfig struct {
	Node       Server
	Blockchain *blockchain.BlockChain
}

type SynckerManager struct {
	isEnabled             bool //0 > stop, 1: running
	config                *SynckerManagerConfig
	BeaconSyncProcess     *BeaconSyncProcess
	S2BSyncProcess        *S2BSyncProcess
	ShardSyncProcess      map[int]*ShardSyncProcess
	CrossShardSyncProcess map[int]*CrossShardSyncProcess
	beaconPool            *BlkPool
	shardPool             map[int]*BlkPool
	s2bPool               *BlkPool
	crossShardPool        map[int]*CrossShardBlkPool
}

func NewSynckerManager() *SynckerManager {
	s := &SynckerManager{
		ShardSyncProcess:      make(map[int]*ShardSyncProcess),
		shardPool:             make(map[int]*BlkPool),
		CrossShardSyncProcess: make(map[int]*CrossShardSyncProcess),
		crossShardPool:        make(map[int]*CrossShardBlkPool),
	}
	return s
}

func (synckerManager *SynckerManager) Init(config *SynckerManagerConfig) {
	synckerManager.config = config
	//init beacon sync process
	synckerManager.BeaconSyncProcess = NewBeaconSyncProcess(synckerManager.config.Node, synckerManager.config.Blockchain.BeaconChain)
	synckerManager.S2BSyncProcess = synckerManager.BeaconSyncProcess.s2bSyncProcess
	synckerManager.beaconPool = synckerManager.BeaconSyncProcess.beaconPool
	synckerManager.s2bPool = synckerManager.S2BSyncProcess.s2bPool

	//init shard sync process
	for _, chain := range synckerManager.config.Blockchain.ShardChain {
		sid := chain.GetShardID()
		synckerManager.ShardSyncProcess[sid] = NewShardSyncProcess(sid, synckerManager.config.Node, synckerManager.config.Blockchain.BeaconChain, chain)
		synckerManager.shardPool[sid] = synckerManager.ShardSyncProcess[sid].shardPool
		synckerManager.CrossShardSyncProcess[sid] = synckerManager.ShardSyncProcess[sid].crossShardSyncProcess
		synckerManager.crossShardPool[sid] = synckerManager.CrossShardSyncProcess[sid].crossShardPool

	}

	//watch commitee change
	go synckerManager.manageSyncProcess()

	//Publish node state to other peer
	go func() {
		t := time.NewTicker(time.Second * 3)
		for _ = range t.C {
			_, chainID := synckerManager.config.Node.GetUserMiningState()
			if chainID == -1 {
				_ = synckerManager.config.Node.PublishNodeState("beacon", chainID)
			}
			if chainID >= 0 {
				_ = synckerManager.config.Node.PublishNodeState("shard", chainID)
			}
		}
	}()
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

// periodically check user commmittee status, enable shard sync process if needed (beacon always start)
func (synckerManager *SynckerManager) manageSyncProcess() {
	defer time.AfterFunc(time.Second*5, synckerManager.manageSyncProcess)

	//check if enable
	if !synckerManager.isEnabled || synckerManager.config == nil {
		return
	}
	role, chainID := synckerManager.config.Node.GetUserMiningState()
	synckerManager.BeaconSyncProcess.start(chainID == -1)

	if role == common.CommitteeRole || role == common.PendingRole {
		if chainID == -1 {
			synckerManager.BeaconSyncProcess.isCommittee = true
		} else {
			for sid, syncProc := range synckerManager.ShardSyncProcess {
				if int(sid) == chainID {
					syncProc.isCommittee = true
					syncProc.start()
				} else {
					syncProc.isCommittee = false
					syncProc.stop()
				}
			}
		}
	}

	if chainID == -1 {
		synckerManager.config.Node.PublishNodeState(common.BeaconRole, chainID)
	} else if chainID >= 0 {
		synckerManager.config.Node.PublishNodeState(common.ShardRole, chainID)
	} else {
		synckerManager.config.Node.PublishNodeState("", -2)
	}

}

//Process incomming broadcast block
func (synckerManager *SynckerManager) ReceiveBlock(blk interface{}, peerID string) {
	switch blk.(type) {
	case *blockchain.BeaconBlock:
		beaconBlk := blk.(*blockchain.BeaconBlock)
		synckerManager.beaconPool.AddBlock(beaconBlk)
		fmt.Printf("syncker: receive beacon block %d \n", beaconBlk.GetHeight())
		//create fake s2b pool peerstate
		synckerManager.BeaconSyncProcess.beaconPeerStateCh <- &wire.MessagePeerState{
			Beacon: wire.ChainState{
				Timestamp: beaconBlk.Header.Timestamp,
				BlockHash: *beaconBlk.Hash(),
				Height:    beaconBlk.GetHeight(),
			},
			SenderID:  time.Now().String(),
			Timestamp: time.Now().Unix(),
		}

	case *blockchain.ShardBlock:

		shardBlk := blk.(*blockchain.ShardBlock)
		//fmt.Printf("syncker: receive shard block %d \n", shardBlk.GetHeight())
		synckerManager.shardPool[shardBlk.GetShardID()].AddBlock(shardBlk)

	case *blockchain.ShardToBeaconBlock:
		s2bBlk := blk.(*blockchain.ShardToBeaconBlock)
		synckerManager.s2bPool.AddBlock(s2bBlk)
		//fmt.Println("syncker AddBlock S2B", s2bBlk.Header.ShardID, s2bBlk.Header.Height)
		//create fake s2b pool peerstate
		synckerManager.S2BSyncProcess.s2bPeerStateCh <- &wire.MessagePeerState{
			SenderID:          time.Now().String(),
			ShardToBeaconPool: map[byte][]uint64{s2bBlk.Header.ShardID: []uint64{1, s2bBlk.GetHeight()}},
			Timestamp:         time.Now().Unix(),
		}
	case *blockchain.CrossShardBlock:
		csBlk := blk.(*blockchain.CrossShardBlock)
		fmt.Printf("crossdebug: receive block from %d to %d (%synckerManager)\n", csBlk.Header.ShardID, csBlk.ToShardID, csBlk.Hash().String())
		synckerManager.crossShardPool[int(csBlk.ToShardID)].AddBlock(csBlk)
	}

}

//Process incomming broadcast peerstate
func (synckerManager *SynckerManager) ReceivePeerState(peerState *wire.MessagePeerState) {
	//b, _ := json.Marshal(peerState)
	//fmt.Println("SYNCKER: receive peer state", string(b))
	//beacon
	if peerState.Beacon.Height != 0 {
		synckerManager.BeaconSyncProcess.beaconPeerStateCh <- peerState
	}
	//s2b
	if len(peerState.ShardToBeaconPool) != 0 {
		synckerManager.S2BSyncProcess.s2bPeerStateCh <- peerState
	}
	//shard
	for sid, _ := range peerState.Shards {
		synckerManager.ShardSyncProcess[int(sid)].shardPeerStateCh <- peerState
	}
}

//Get S2B Block for creating beacon block
func (synckerManager *SynckerManager) GetS2BBlocksForBeaconProducer() map[byte][]interface{} {
	bestViewShardHash := synckerManager.config.Blockchain.BeaconChain.GetShardBestViewHash()
	res := make(map[byte][]interface{})

	for i := 0; i < synckerManager.config.Node.GetChainParam().ActiveShards; i++ {
		v := bestViewShardHash[byte(i)]
		//beacon beststate dont have shard hash  => create one
		if (&v).IsEqual(&common.Hash{}) {
			blk := *synckerManager.config.Node.GetChainParam().GenesisShardBlock
			blk.Header.ShardID = byte(i)
			v = *blk.Hash()
		}
		for _, v := range synckerManager.s2bPool.GetFinalBlockFromBlockHash(v.String()) {
			res[byte(i)] = append(res[byte(i)], v)
			fmt.Println("syncker: get block ", i, v.GetHeight(), v.Hash().String())
		}
	}
	return res
}

//Get Crossshard Block for creating shardblock block
func (synckerManager *SynckerManager) GetCrossShardBlocksForShardProducer(toShard byte) map[byte][]interface{} {
	//get last confirm crossshard -> process request until retrieve info
	res := make(map[byte][]interface{})
	db := synckerManager.config.Node.GetIncDatabase()
	lastRequestCrossShard := synckerManager.ShardSyncProcess[int(toShard)].Chain.GetCrossShardState()
	for i := 0; i < synckerManager.config.Node.GetChainParam().ActiveShards; i++ {
		for {
			if i == int(toShard) {
				break
			}
			requestHeight := lastRequestCrossShard[byte(i)]
			nextCrossShardInfo := synckerManager.config.Node.FetchNextCrossShard(i, int(toShard), requestHeight)
			if nextCrossShardInfo == nil {
				break
			}
			beaconHash, _ := common.Hash{}.NewHashFromStr(nextCrossShardInfo.ConfirmBeaconHash)
			beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(db, *beaconHash)
			if err != nil {
				break
			}

			beaconBlock := new(blockchain.BeaconBlock)
			json.Unmarshal(beaconBlockBytes, beaconBlock)
			for _, shardState := range beaconBlock.Body.ShardState[byte(i)] {
				if shardState.Height == nextCrossShardInfo.NextCrossShardHeight {
					if synckerManager.crossShardPool[int(toShard)].HasBlock(shardState.Hash) {
						res[byte(i)] = append(res[byte(i)], synckerManager.crossShardPool[int(toShard)].GetBlock(shardState.Hash))
					}
					lastRequestCrossShard[byte(i)] = nextCrossShardInfo.NextCrossShardHeight
					break
				}
			}
		}
	}
	return res
}

//Get S2B Block for validating beacon block
func (synckerManager *SynckerManager) GetS2BBlocksForBeaconValidator(list map[byte][]common.Hash) (map[byte][]interface{}, error) {
	s2bPoolLists := synckerManager.GetS2BBlocksForBeaconProducer()

	missingBlocks := compareLists(s2bPoolLists, list)
	// synckerManager.config.Server.
	if len(missingBlocks) > 0 {
		ticker := time.NewTicker(5 * time.Second)
		<-ticker.C

		s2bPoolLists = synckerManager.GetS2BBlocksForBeaconProducer()
		missingBlocks = compareLists(s2bPoolLists, list)
		if len(missingBlocks) > 0 {
			return nil, errors.New("Unable to sync required block in time")
		}
	}

	return s2bPoolLists, nil
}

//Get Crossshard Block for validating shardblock block
func (synckerManager *SynckerManager) GetCrossShardBlocksForShardValidator(toShard byte, list map[byte][]common.Hash) (map[byte][]interface{}, error) {
	crossShardPoolLists := synckerManager.GetCrossShardBlocksForShardProducer(toShard)

	missingBlocks := compareLists(crossShardPoolLists, list)
	// synckerManager.config.Server.
	if len(missingBlocks) > 0 {
		ticker := time.NewTicker(5 * time.Second)
		<-ticker.C

		crossShardPoolLists = synckerManager.GetCrossShardBlocksForShardProducer(toShard)
		missingBlocks = compareLists(crossShardPoolLists, list)
		if len(missingBlocks) > 0 {
			return nil, errors.New("Unable to sync required block in time")
		}
	}

	return crossShardPoolLists, nil
}

//Get Status Function
type syncInfo struct {
	IsSync     bool
	IsLatest   bool
	PoolLength int
}

type SynckerStatusInfo struct {
	Beacon     syncInfo
	S2B        syncInfo
	Shard      map[int]*syncInfo
	Crossshard map[int]*syncInfo
}

func (synckerManager *SynckerManager) GetSyncStatus(includePool bool) SynckerStatusInfo {
	info := SynckerStatusInfo{}
	info.Beacon = syncInfo{
		IsSync:   synckerManager.BeaconSyncProcess.status == RUNNING_SYNC,
		IsLatest: synckerManager.BeaconSyncProcess.isCatchUp,
	}
	info.S2B = syncInfo{
		IsSync:   synckerManager.S2BSyncProcess.status == RUNNING_SYNC,
		IsLatest: false,
	}

	info.Shard = make(map[int]*syncInfo)
	for k, v := range synckerManager.ShardSyncProcess {
		info.Shard[k] = &syncInfo{
			IsSync:   v.status == RUNNING_SYNC,
			IsLatest: v.isCatchUp,
		}
	}

	info.Crossshard = make(map[int]*syncInfo)
	for k, v := range synckerManager.CrossShardSyncProcess {
		info.Crossshard[k] = &syncInfo{
			IsSync:   v.status == RUNNING_SYNC,
			IsLatest: false,
		}
	}

	if includePool {
		info.Beacon.PoolLength = synckerManager.beaconPool.GetPoolLength()
		info.S2B.PoolLength = synckerManager.s2bPool.GetPoolLength()
		for k, _ := range synckerManager.ShardSyncProcess {
			info.Shard[k].PoolLength = synckerManager.shardPool[k].GetPoolLength()
		}
		for k, _ := range synckerManager.CrossShardSyncProcess {
			info.Crossshard[k].PoolLength = synckerManager.crossShardPool[k].GetPoolLength()
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
