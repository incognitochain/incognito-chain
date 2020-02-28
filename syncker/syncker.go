package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
)

type Server interface {
	GetUserMiningState() (role string, chainID int)

	//Request block from "peerID" of shard "fromSID" with param currentFinalHeight and currentBestHash
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	RequestBlocksViaStream(ctx context.Context, peerID string, fromSID int, _type proto.BlkType, fromBlockHeight uint64, finalBlockHeight uint64, toBlockheight uint64, toBlockHashString string) (blockCh chan common.BlockInterface, err error)

	//Request cross block from "peerID" for shard "toShardID" with param latestCrossShardBlockHeight in current pool
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	//RequestCrossShardBlock(peerID string, toShardID int, latestCrossShardBlockHeight uint64) (blockCh chan interface{}, stopCh chan int)

	//Request s2b block from "peerID" of shard "fromSID" with param latestS2BHeight in current pool
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	//RequestS2BBlock(peerID string, fromSID int, latestS2BHeight uint64) (blockCh chan interface{}, stopCh chan int)

	//GetCrossShardPool(sid byte) Pool
	//GetS2BPool(sid byte) Pool

	PublishNodeState(userLayer string, shardID int) error

	//GetBeaconBestState() Chain
	//GetAllShardBestState() map[byte]Chain
}

type Pool interface {
	GetLatestCrossShardFinalHeight(byte) uint64
	GetLatestFinalHeight() uint64
	AddBlock(block interface{}) error
}

type Chain interface {
	GetBestViewHeight() uint64
	GetFinalViewHeight() uint64
	SetReady(bool)
	IsReady() bool
	GetBestViewHash() string
	GetFinalViewHash() string

	GetEpoch() uint64
	ValidateBlockSignatures(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error
	GetCommittee() []incognitokey.CommitteePublicKey
	CurrentHeight() uint64
	InsertBlk(block common.BlockInterface) error
}

type SynckerConfig struct {
	Node       Server
	Blockchain *blockchain.BlockChain
}

type Syncker struct {
	IsEnabled bool //0 > stop, 1: running
	config    *SynckerConfig

	PeerStateCh chan *wire.MessagePeerState

	BeaconBlockCh        chan common.BlockPoolInterface
	ShardBlockCh         chan common.BlockPoolInterface
	ShardToBeaconBlockCh chan common.BlockPoolInterface

	BeaconSyncProcess *BeaconSyncProcess
	ShardSyncProcess  map[int]*ShardSyncProcess

	BeaconPeerStates    map[string]BeaconPeerState             //sender -> state
	ShardPeerStates     map[int]map[string]ShardPeerState      //sid -> sender -> state
	S2BPeerState        map[string]S2BPeerState                //sender -> state
	CrossShardPeerState map[int]map[string]CrossShardPeerState //toShardID -> fromShardID-> state

}

// Everytime beacon block is inserted after sync finish, we update shard committee (from beacon view)
func (s *Syncker) WatchCommitteeChange() {
	defer func() {
		time.AfterFunc(time.Second*3, s.WatchCommitteeChange)
	}()

	//check if enable
	if !s.IsEnabled || s.config == nil {
		fmt.Println("SYNCKER: enable", s.IsEnabled, s.config == nil)
		return
	}
	role, chainID := s.config.Node.GetUserMiningState()
	s.BeaconSyncProcess.Start(chainID)

	if role == common.CommitteeRole || role == common.PendingRole {
		if chainID == -1 {
			s.BeaconSyncProcess.IsCommittee = true
		} else {
			for sid, syncProc := range s.ShardSyncProcess {
				if int(sid) == chainID {
					syncProc.IsCommittee = true
					syncProc.Start()
				} else {
					syncProc.IsCommittee = false
					syncProc.Stop()
				}
			}
		}
	}

	if chainID == -1 {
		s.config.Node.PublishNodeState(common.BeaconRole, chainID)
	} else if chainID >= 0 {
		s.config.Node.PublishNodeState(common.ShardRole, chainID)
	} else {
		s.config.Node.PublishNodeState("", -2)
	}

}

func NewSyncker() *Syncker {
	s := &Syncker{
		PeerStateCh:          make(chan *wire.MessagePeerState),
		BeaconBlockCh:        make(chan common.BlockPoolInterface),
		ShardBlockCh:         make(chan common.BlockPoolInterface),
		ShardToBeaconBlockCh: make(chan common.BlockPoolInterface),

		ShardSyncProcess: make(map[int]*ShardSyncProcess),

		BeaconPeerStates:    make(map[string]BeaconPeerState),
		ShardPeerStates:     make(map[int]map[string]ShardPeerState),
		S2BPeerState:        make(map[string]S2BPeerState),
		CrossShardPeerState: make(map[int]map[string]CrossShardPeerState),
	}
	return s
}

func (s *Syncker) Init(config *SynckerConfig) {
	s.config = config
	//init beacon sync process
	s.BeaconSyncProcess = NewBeaconSyncProcess(s.config.Node, s.config.Blockchain.Chains["beacon"])
	s.BeaconSyncProcess.BeaconPeerStates = s.BeaconPeerStates
	s.BeaconSyncProcess.S2BPeerState = s.S2BPeerState

	//init shard sync process
	for chainName, chain := range s.config.Blockchain.Chains {
		if chainName != "beacon" {
			sid := chain.GetShardID()
			s.ShardSyncProcess[sid] = NewShardSyncProcess(sid, s.config.Node, s.config.Blockchain.Chains["beacon"], chain)
			s.ShardPeerStates[int(sid)] = make(map[string]ShardPeerState)
			s.CrossShardPeerState[int(sid)] = make(map[string]CrossShardPeerState)

			s.ShardSyncProcess[sid].ShardPeerState = s.ShardPeerStates[sid]
			s.ShardSyncProcess[sid].CrossShardPeerState = s.CrossShardPeerState[sid]
		}
	}

	//watch commitee change
	go s.WatchCommitteeChange()

	//Publish node state to other peer
	go func() {
		t := time.NewTicker(time.Second * 3)
		for _ = range t.C {
			_, chainID := s.config.Node.GetUserMiningState()
			if chainID == -1 {
				_ = s.config.Node.PublishNodeState("beacon", chainID)
			}
			if chainID >= 0 {
				_ = s.config.Node.PublishNodeState("shard", chainID)
			}
		}
	}()
}

func (s *Syncker) ReceiveS2BBlock(blk common.BlockPoolInterface, peerID string) {
	s.ShardToBeaconBlockCh <- blk
	s.BeaconSyncProcess.lock.Lock()
	s2bState := make(map[int]uint64)
	s2bState[blk.GetShardID()] = blk.GetHeight()
	if peerID == "" {
		peerID = time.Now().String()
	}
	s.S2BPeerState[peerID] = S2BPeerState{
		Timestamp: time.Now().Unix(),
		Height:    s2bState,
	}
	s.BeaconSyncProcess.lock.Unlock()

}

func (s *Syncker) Start() {
	s.IsEnabled = true

	//TODO: clear outdate peerstate
	for {
		if !s.IsEnabled {
			break
		}
		select {
		case block := <-s.BeaconBlockCh:
			s.BeaconSyncProcess.BeaconPool.AddBlock(block)
		case block := <-s.ShardBlockCh:
			s.ShardSyncProcess[block.GetShardID()].ShardPool.AddBlock(block)
		case block := <-s.ShardToBeaconBlockCh:
			s.BeaconSyncProcess.S2BPool.AddBlock(block)
		case peerState := <-s.PeerStateCh:
			//b, _ := json.Marshal(peerState)
			//fmt.Println("SYNCKER: receive peer state", string(b))
			//beacon
			if peerState.Beacon.Height != 0 {
				s.BeaconSyncProcess.lock.Lock()
				s.BeaconPeerStates[peerState.SenderID] = BeaconPeerState{
					Timestamp:      peerState.Timestamp,
					BestViewHash:   peerState.Beacon.BlockHash.String(),
					BestViewHeight: peerState.Beacon.Height,
				}
				s.BeaconSyncProcess.lock.Unlock()
			}
			//s2b
			if len(peerState.ShardToBeaconPool) != 0 {
				s.BeaconSyncProcess.lock.Lock()
				s2bState := make(map[int]uint64)
				for sid, v := range peerState.ShardToBeaconPool {
					s2bState[int(sid)] = v[len(v)-1]
				}
				s.S2BPeerState[peerState.SenderID] = S2BPeerState{
					Timestamp: peerState.Timestamp,
					Height:    s2bState,
				}
				s.BeaconSyncProcess.lock.Unlock()
			}
			//shard
			//fmt.Printf("SYNCKER %+v\n", peerState)
			for sid, peerShardState := range peerState.Shards {
				s.ShardSyncProcess[int(sid)].lock.Lock()
				s.ShardPeerStates[int(sid)][peerState.SenderID] = ShardPeerState{
					Timestamp:      peerState.Timestamp,
					BestViewHash:   peerShardState.BlockHash.String(),
					BestViewHeight: peerShardState.Height,
				}
				s.ShardSyncProcess[int(sid)].lock.Unlock()
			}
			//crossshard
			if len(peerState.CrossShardPool) != 0 {
				for sid, peerShardState := range peerState.CrossShardPool {
					crossShardState := make(map[byte]uint64)
					for sid, v := range peerShardState {
						crossShardState[sid] = v[len(v)-1]
					}
					s.ShardSyncProcess[int(sid)].lock.Lock()
					s.CrossShardPeerState[int(sid)][peerState.SenderID] = CrossShardPeerState{
						Timestamp: peerState.Timestamp,
						Height:    crossShardState,
					}
					s.ShardSyncProcess[int(sid)].lock.Unlock()
				}

			}

		}
	}
}

func (s *Syncker) Stop() {
	s.IsEnabled = false
	s.BeaconSyncProcess.Stop()
	for _, chain := range s.ShardSyncProcess {
		chain.Stop()
	}
}
