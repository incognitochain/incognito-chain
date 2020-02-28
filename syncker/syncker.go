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
	GetChainParam() *blockchain.Params
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

type BeaconChainInterface interface {
	Chain
	GetShardBestViewHash() map[byte]common.Hash
	GetShardBestViewHeight() map[byte]uint64
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
	IsEnabled         bool //0 > stop, 1: running
	config            *SynckerConfig
	BeaconSyncProcess *BeaconSyncProcess
	S2BSyncProcess    *S2BSyncProcess
	ShardSyncProcess  map[int]*ShardSyncProcess
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
		ShardSyncProcess: make(map[int]*ShardSyncProcess),
	}

	return s
}

func (s *Syncker) Init(config *SynckerConfig) {
	s.config = config
	//init beacon sync process
	s.BeaconSyncProcess = NewBeaconSyncProcess(s.config.Node, s.config.Blockchain.Chains["beacon"].(BeaconChainInterface))
	s.S2BSyncProcess = s.BeaconSyncProcess.S2BSyncProcess

	//init shard sync process
	for chainName, chain := range s.config.Blockchain.Chains {
		if chainName != "beacon" {
			sid := chain.GetShardID()
			s.ShardSyncProcess[sid] = NewShardSyncProcess(sid, s.config.Node, s.config.Blockchain.Chains["beacon"], chain)
			//s.Cr[sid] = NewShardSyncProcess(sid, s.config.Node, s.config.Blockchain.Chains["beacon"], chain)
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

func (s *Syncker) ReceiveBlock(blk interface{}, peerID string) {
	switch blk.(type) {
	case *blockchain.BeaconBlock:
		beaconBlk := blk.(*blockchain.BeaconBlock)
		s.BeaconSyncProcess.BeaconPool.AddBlock(beaconBlk)
		//create fake s2b pool peerstate
		s.BeaconSyncProcess.BeaconPeerStateCh <- &wire.MessagePeerState{
			Beacon: wire.ChainState{
				Timestamp: beaconBlk.Header.Timestamp,
				BlockHash: *beaconBlk.Hash(),
				Height:    beaconBlk.GetHeight(),
			},
		}

	case *blockchain.ShardBlock:
		shardBlk := blk.(*blockchain.ShardBlock)
		s.ShardSyncProcess[shardBlk.GetShardID()].ShardPool.AddBlock(shardBlk)

	case *blockchain.ShardToBeaconBlock:
		s2bBlk := blk.(*blockchain.ShardToBeaconBlock)
		s.S2BSyncProcess.S2BPool.AddBlock(s2bBlk)
		//create fake s2b pool peerstate
		s.S2BSyncProcess.S2BPeerStateCh <- &wire.MessagePeerState{
			SenderID:          time.Now().String(),
			ShardToBeaconPool: map[byte][]uint64{s2bBlk.Header.ShardID: []uint64{1, s2bBlk.GetHeight()}},
			Timestamp:         time.Now().Unix(),
		}
	}

}

func (s *Syncker) ReceivePeerState(peerState *wire.MessagePeerState) {
	//b, _ := json.Marshal(peerState)
	//fmt.Println("SYNCKER: receive peer state", string(b))
	//beacon
	if peerState.Beacon.Height != 0 {
		s.BeaconSyncProcess.BeaconPeerStateCh <- peerState
	}
	//s2b
	if len(peerState.ShardToBeaconPool) != 0 {
		s.S2BSyncProcess.S2BPeerStateCh <- peerState
	}
	//shard
	//fmt.Printf("SYNCKER %+v\n", peerState)
	for sid, _ := range peerState.Shards {
		s.ShardSyncProcess[int(sid)].ShardPeerStateCh <- peerState
	}
	//crossshard
	//if len(peerState.CrossShardPool) != 0 {
	//	for sid, peerShardState := range peerState.CrossShardPool {
	//		crossShardState := make(map[byte]uint64)
	//		for sid, v := range peerShardState {
	//			crossShardState[sid] = v[len(v)-1]
	//		}
	//		s.ShardSyncProcess[int(sid)].lock.Lock()
	//		s.CrossShardPeerState[int(sid)][peerState.SenderID] = CrossShardPeerState{
	//			Timestamp: peerState.Timestamp,
	//			Height:    crossShardState,
	//		}
	//		s.ShardSyncProcess[int(sid)].lock.Unlock()
	//	}
	//}
}

func (s *Syncker) Start() {
	s.IsEnabled = true
}

func (s *Syncker) Stop() {
	s.IsEnabled = false
	s.BeaconSyncProcess.Stop()

	for _, chain := range s.ShardSyncProcess {
		chain.Stop()
	}
}
