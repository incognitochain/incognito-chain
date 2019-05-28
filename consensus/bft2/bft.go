package bft2

import (
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"time"
)

/*
	Sequence Number: blockheight + round
*/

const (
	PROPOSE    = "PROPOSE"
	LISTEN     = "LISTEN"
	PREPREPARE = "PREPREPARE"
	PREPARE    = "PREPARE"
	COMMIT     = "COMMIT"
	NEWROUND   = "NEWROUND"
)

const (
	PROPOSE_TIMEOUT     = 30 * time.Second
	LISTEN_TIMEOUT      = 30 * time.Second
	PREPARE_TIMEOUT     = 30 * time.Second
	PRE_PREPARE_TIMEOUT = 30 * time.Second
	COMMIT_TIMEOUT      = 30 * time.Second
)

const (
	BEACON_TIME_MIN = 10 * time.Second
	SHARD_TIME_MIN  = 5 * time.Second
)

type serverInterface interface {
	// list functions callback which are assigned from Server struct
	GetPeerIDsFromPublicKey(string) []libp2p.ID
	PushMessageToAll(wire.Message) error
	PushMessageToPeer(wire.Message, libp2p.ID) error
	PushMessageToShard(wire.Message, byte) error
	PushMessageToBeacon(wire.Message) error
	PushMessageToPbk(wire.Message, string) error
	//UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string)
}

type Block interface {
	//validate() bool
}

type Role struct {
	nodeType string //beacon|shard
	role     string //pending|validator|proposer
	shardID  byte   //shardID
}

type View struct {
	PeerID        string
	PubKey        string
	Timestamp     int
	Role          Role //role of node
	Round         int
	BeaconHeight  uint64
	ShardHeight   map[byte]uint64
	CommitteeSize struct {
		Beacon int
		Shard  map[byte]int
	}
}

type NextState struct {
	state     *BFTState
	nextState string
}

type PrePrepareMsg struct {
	isOk bool
}

type PrepareMsg struct {
	isOk bool
}

type CommitMsg struct {
	isOk bool
}

type BFTState struct {
	State         string //state
	View          View   //view when creating/listening block
	Block         Block  //message
	isOk          bool   //vote for this message
	PrePrepareMsg []PrePrepareMsg
	PrepareMsg    []PrePrepareMsg
	Commit        []PrePrepareMsg
}

type BFTEngine struct {
	CurrentState    *BFTState
	ValidatorsView  map[string]View
	UserKeySet      *cashec.KeySet
	Server          serverInterface
	newStateCh      chan NextState
	IsReady         bool
	ProposedBlockCh chan Block
	PrePrepareMsgCh chan PrePrepareMsg
	PrepareMsgCh    chan PrepareMsg
	CommitMsgCh     chan CommitMsg
	ViewMsgCh       chan View
}

func (e *BFTEngine) Start(
	UserKeySet *cashec.KeySet,
	Server serverInterface,
) {
	//var stateCache = make(map[uint64]BFTState)
	//var curentState = BFTState{}

	broadcastViewTicker := time.Tick(5 * time.Second)
	checkReadyTicker := time.Tick(1 * time.Second)
	go func() {
		for _ = range broadcastViewTicker {
			_ = e.createCurrentView()
			//TODO: broadcast view

		}
	}()

	for { //data flow
		select {
		case s := <-e.newStateCh:
			e.nextState(s.state, s.nextState)
		case <-e.ProposedBlockCh:
		case <-e.PrePrepareMsgCh:
		case <-e.PrepareMsgCh:
		case <-e.CommitMsgCh:

		case view := <-e.ViewMsgCh: //update view of other nodes
			nodeType, shardID := blockchain.GetBestStateBeacon().GetPubkeyNodeRole(view.PubKey)
			curView := e.createCurrentView()
			if nodeType == curView.Role.nodeType && shardID == curView.Role.shardID {
				e.ValidatorsView[view.PubKey] = view
			}
		case <-checkReadyTicker:
			curView := e.createCurrentView()
			if curView.Role.nodeType == common.BEACON_ROLE {
				//if beacon chain -> beacon must be max
				maxHeight := uint64(0)
				for _, v := range e.ValidatorsView {
					nodeType, shardID := blockchain.GetBestStateBeacon().GetPubkeyNodeRole(v.PubKey)
					if nodeType != curView.Role.nodeType && shardID != curView.Role.shardID {
						continue
					}
					if v.BeaconHeight > maxHeight {
						maxHeight = v.BeaconHeight
					}
				}
				//check if get max
				if curView.BeaconHeight == maxHeight && curView.Role.role != common.PENDING_ROLE {
					e.IsReady = true
				} else {
					e.IsReady = false
				}
			}

			if curView.Role.nodeType == common.SHARD_ROLE {
				//if shard chain -> shard height must be max
				maxHeight := uint64(0)
				for _, v := range e.ValidatorsView {
					nodeType, shardID := blockchain.GetBestStateBeacon().GetPubkeyNodeRole(v.PubKey)
					if nodeType != curView.Role.nodeType && shardID != curView.Role.shardID {
						continue
					}
					if v.ShardHeight[curView.Role.shardID] > maxHeight {
						maxHeight = v.BeaconHeight
					}
				}
				//check if get max
				if curView.ShardHeight[curView.Role.shardID] == maxHeight && curView.Role.role != common.PENDING_ROLE {
					e.IsReady = true
				} else {
					e.IsReady = false
				}
			}

		}
	}

}

func (e *BFTEngine) nextState(s *BFTState, nextState string) {
	if s.State == nextState {
		return //already transition
	}

	switch nextState {
	case PROPOSE:
		s.State = PROPOSE
		e.handleProposePhase(s)
	case LISTEN:
		s.State = LISTEN
		e.handleListenPhase(s)
	case PREPREPARE:
		s.State = PREPREPARE
		e.handlePrePreparePhase(s)
	case PREPARE:
		s.State = PREPARE
		e.handlePreparePhase(s)
	case COMMIT:
		s.State = COMMIT
		e.handleCommitPhase(s)
	case NEWROUND:
		s.State = NEWROUND
		e.handleNewRoundPhase(s)
	}
}
