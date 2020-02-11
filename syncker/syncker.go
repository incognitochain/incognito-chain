package syncker

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

type Server interface {
	//Request block from "peerID" of shard "fromSID" with param currentFinalHeight and currentBestHash
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	RequestBlock(peerID string, fromSID int, currentFinalHeight uint64, currentBestHash string) (blockCh chan interface{}, stopCh chan int)

	//Request cross block from "peerID" for shard "toShardID" with param latestCrossShardBlockHeight in current pool
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	RequestCrossShardBlockPool(peerID string, toShardID int, latestCrossShardBlockHeight uint64) (blockCh chan interface{}, stopCh chan int)

	//Request s2b block from "peerID" of shard "fromSID" with param latestS2BHeight in current pool
	//Function return channel of each block, and a stop channel to tell sender side to stop send block
	RequestS2BBlockPool(peerID string, fromSID int, latestS2BHeight uint64) (blockCh chan interface{}, stopCh chan int)

	GetCrossShardPool(sid byte) Pool
	GetS2BPool(sid byte) Pool

	PublishNodeState(userLayer string, shardID int) bool
}

type Pool interface {
	GetLatestCrossShardFinalHeight(byte) uint64
	GetLatestFinalHeight() uint64
	AddBlock(block interface{}) error
}

type ViewInterface interface {
	GetHeight() uint64
	GetHash() string
}

type Chain interface {
	GetBestView() ViewInterface
	GetFinalView() ViewInterface
	InsertBlock(block interface{}) error
}

type PeerStateMsg struct {
	PeerID string
}

type Syncker struct {
	PeerStateCh chan wire.MessagePeerState
	UserPk      incognitokey.CommitteePublicKey

	BeaconSyncProcess *BeaconSyncProcess
	ShardSyncProcess  map[byte]*ShardSyncProcess

	BeaconPeerState map[string]BeaconPeerState         //sender -> state
	ShardPeerState  map[byte]map[string]ShardPeerState //sid -> sender -> state

	S2BPeerState map[string]S2BPeerState //sender -> state

	CrossShardPeerState map[byte]map[string]CrossShardPeerState //toShardID -> fromShardID-> state
}

// Everytime beacon block is inserted after sync finish, we update shard committee (from beacon view)
func (s *Syncker) UpdateCurrentCommittee(shardCommittee map[byte][]incognitokey.CommitteePublicKey, shardPendingCommittee map[byte][]incognitokey.CommitteePublicKey) {
	userBlsPubKey := s.UserPk.GetMiningKeyBase58("bls")

	{ //check shard
		shardID := byte(0)
		for sid, committees := range shardCommittee {
			syncShard := false
			shardID = sid
			for _, v := range committees {
				if userBlsPubKey == v.GetMiningKeyBase58("bls") {
					syncShard = true
					break
				}
			}
			for _, v := range shardPendingCommittee[shardID] {
				if userBlsPubKey == v.GetMiningKeyBase58("bls") {
					syncShard = true
					break
				}
			}
			if syncShard {
				s.ShardSyncProcess[shardID].Start()
			} else {
				s.ShardSyncProcess[shardID].Stop()
			}
		}
	}
}

func (s *Syncker) UpdatePeerState() {
	for {
		select {
		case peerState := <-s.PeerStateCh:
			//beacon
			if peerState.Beacon.Height != 0 {
				s.BeaconPeerState[peerState.SenderID] = BeaconPeerState{
					Timestamp:      peerState.Timestamp,
					BestViewHash:   peerState.Beacon.BlockHash.String(),
					BestViewHeight: peerState.Beacon.Height,
				}
			}
			//s2b
			if len(peerState.ShardToBeaconPool) != 0 {
				s2bState := make(map[byte]uint64)
				for sid, v := range peerState.ShardToBeaconPool {
					s2bState[sid] = v[len(v)-1]
				}
				s.S2BPeerState[peerState.SenderID] = S2BPeerState{
					Timestamp: peerState.Timestamp,
					Height:    s2bState,
				}
			}
			//shard
			for sid, peerShardState := range peerState.Shards {
				s.ShardPeerState[sid][peerState.SenderID] = ShardPeerState{
					Timestamp:      peerState.Timestamp,
					BestViewHash:   peerShardState.BlockHash.String(),
					BestViewHeight: peerShardState.Height,
				}
			}
			//crossshard
			if len(peerState.CrossShardPool) != 0 {
				for sid, peerShardState := range peerState.CrossShardPool {
					crossShardState := make(map[byte]uint64)
					for sid, v := range peerShardState {
						crossShardState[sid] = v[len(v)-1]
					}
					s.CrossShardPeerState[sid][peerState.SenderID] = CrossShardPeerState{
						Timestamp: peerState.Timestamp,
						Height:    crossShardState,
					}
				}
			}
		}
	}
}

func NewSyncker(userPk incognitokey.CommitteePublicKey, server Server, beaconChain Chain, shardChain map[byte]Chain) *Syncker {
	s := &Syncker{
		PeerStateCh:      make(chan wire.MessagePeerState),
		UserPk:           userPk,
		ShardSyncProcess: make(map[byte]*ShardSyncProcess),
	}
	go s.UpdatePeerState()
	s.BeaconSyncProcess = NewBeaconSyncProcess(server, beaconChain)
	s.BeaconSyncProcess.Start()

	for i, v := range shardChain {
		s.ShardSyncProcess[i] = NewShardSyncProcess(i, server, v)
	}

	//TODO: broadcast node state
	return s
}
