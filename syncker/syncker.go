package syncker

import (
	"github.com/incognitochain/incognito-chain/common"
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

	PublishNodeState(userLayer string, shardID int) error

	GetBeaconBestState() Chain
	GetAllShardBestState() map[byte]Chain
}

type Pool interface {
	GetLatestCrossShardFinalHeight(byte) uint64
	GetLatestFinalHeight() uint64
	AddBlock(block interface{}) error
}

type Chain interface {
	GetBestViewHeight() uint64
	GetFinalViewHeight() uint64

	GetBestViewHash() string
	GetFinalViewHash() string
}

type Syncker struct {
	PeerStateCh chan *wire.MessagePeerState
	UserPk      incognitokey.CommitteePublicKey

	BeaconSyncProcess *BeaconSyncProcess
	ShardSyncProcess  map[byte]*ShardSyncProcess

	BeaconPeerStates    map[string]BeaconPeerState              //sender -> state
	ShardPeerStates     map[byte]map[string]ShardPeerState      //sid -> sender -> state
	S2BPeerState        map[string]S2BPeerState                 //sender -> state
	CrossShardPeerState map[byte]map[string]CrossShardPeerState //toShardID -> fromShardID-> state
}

// Everytime beacon block is inserted after sync finish, we update shard committee (from beacon view)
func (s *Syncker) UpdateCurrentCommittee(relayShards []byte, beaconCommittee, beaconPendingCommittee []incognitokey.CommitteePublicKey, shardCommittee, shardPendingCommittee map[byte][]incognitokey.CommitteePublicKey) {
	userBlsPubKey := s.UserPk.GetMiningKeyBase58("bls")

	{ //check beacon
		isCommittee := false
		for _, v := range beaconCommittee {
			if userBlsPubKey == v.GetMiningKeyBase58("bls") {
				isCommittee = true
				break
			}
		}
		for _, v := range beaconPendingCommittee {
			if userBlsPubKey == v.GetMiningKeyBase58("bls") {
				isCommittee = true
				break
			}
		}
		s.BeaconSyncProcess.IsCommittee = isCommittee
	}

	{ //check shard
		shardID := byte(0)
		for sid, committees := range shardCommittee {
			IsCommitee := false
			shardID = sid
			for _, v := range committees {
				if userBlsPubKey == v.GetMiningKeyBase58("bls") {
					IsCommitee = true
					break
				}
			}
			for _, v := range shardPendingCommittee[shardID] {
				if userBlsPubKey == v.GetMiningKeyBase58("bls") {
					IsCommitee = true
					break
				}
			}
			s.ShardSyncProcess[shardID].IsCommittee = IsCommitee
			if IsCommitee || common.IndexOfByte(shardID, relayShards) > -1 {
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
				s.BeaconPeerStates[peerState.SenderID] = BeaconPeerState{
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
			//fmt.Printf("SYNCKER %+v\n", peerState)
			for sid, peerShardState := range peerState.Shards {

				if s.ShardPeerStates[sid] == nil {
					s.ShardPeerStates[sid] = make(map[string]ShardPeerState)
				}
				s.ShardPeerStates[sid][peerState.SenderID] = ShardPeerState{
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
					if s.CrossShardPeerState[sid] == nil {
						s.CrossShardPeerState[sid] = make(map[string]CrossShardPeerState)
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

func NewSyncker(userPk incognitokey.CommitteePublicKey, server Server) *Syncker {
	s := &Syncker{
		PeerStateCh:      make(chan *wire.MessagePeerState),
		UserPk:           userPk,
		ShardSyncProcess: make(map[byte]*ShardSyncProcess),

		BeaconPeerStates:    make(map[string]BeaconPeerState),
		ShardPeerStates:     make(map[byte]map[string]ShardPeerState),
		S2BPeerState:        make(map[string]S2BPeerState),
		CrossShardPeerState: make(map[byte]map[string]CrossShardPeerState),
	}
	go s.UpdatePeerState()
	//init beacon sync process
	s.BeaconSyncProcess = NewBeaconSyncProcess(server, server.GetBeaconBestState())
	s.BeaconSyncProcess.BeaconPeerStates = s.BeaconPeerStates
	s.BeaconSyncProcess.S2BPeerState = s.S2BPeerState
	s.BeaconSyncProcess.Start()

	//init shard sync process
	for sid, chain := range server.GetAllShardBestState() {
		s.ShardSyncProcess[sid] = NewShardSyncProcess(sid, server, chain)
		s.ShardSyncProcess[sid].ShardPeerState = s.ShardPeerStates[sid]
		s.ShardSyncProcess[sid].CrossShardPeerState = s.CrossShardPeerState[sid]
	}

	//TODO: broadcast node state
	return s
}
