package bft2

import (
	"fmt"
	"github.com/constant-money/constant-chain/wire"
	"time"
)

/*
	Sequence Number: blockheight + round
*/

const (
	PROPOSE  = "PROPOSE"
	LISTEN   = "LISTEN"
	PREPARE  = "PREPARE"
	COMMIT   = "COMMIT"
	NEWROUND = "NEWROUND"
)

const (
	TIMEOUT = 30 * time.Second
)

type ChainInterface interface {
	// list functions callback which are assigned from Server struct
	//GetPeerIDsFromPublicKey(string) []libp2p.ID
	//PushMessageToAll(wire.Message) error
	//PushMessageToPeer(wire.Message, libp2p.ID) error
	//PushMessageToShard(wire.Message, byte) error
	//PushMessageToBeacon(wire.Message) error
	//PushMessageToPbk(wire.Message, string) error
	//UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string)
	PushMessageToValidator(wire.Message) error
	GetLastBlockTimeStamp() int64
	GetBlkMinTime() time.Duration
	IsReady() bool
	GetRole() Role
	GetHeight() uint64
	GetCommitteeSize() int
}

type BlockInterface interface {
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
	Height        uint64
	CommitteeSize int
}

type ProposeMsg struct {
	isOk bool
}

type PrepareMsg struct {
	isOk bool
}

type CommitMsg struct {
	isOk bool
}

type BFTState struct {
	State      string         //state
	View       View           //view when creating/listening block
	Block      BlockInterface //message
	isOk       bool           //vote for this message
	PrepareMsg []PrepareMsg
	Commit     []CommitMsg
}

type BFTEngine struct {
	PeerID         string
	CurrentState   *BFTState
	ValidatorsView map[string]View
	NextStateCh    chan string
	Chain          ChainInterface
	IsReady        bool
	ProposeMsgCh   chan ProposeMsg
	PrepareMsgCh   chan PrepareMsg
	CommitMsgCh    chan CommitMsg
	ViewMsgCh      chan View
}

func (e *BFTEngine) Start() {
	e.ViewMsgCh = make(chan View)
	e.ProposeMsgCh = make(chan ProposeMsg)
	e.PrepareMsgCh = make(chan PrepareMsg)
	e.CommitMsgCh = make(chan CommitMsg)
	e.NextStateCh = make(chan string)

	//var stateCache = make(map[uint64]BFTState)
	//var curentState = BFTState{}

	broadcastViewTicker := time.Tick(5 * time.Second)
	ticker := time.Tick(100 * time.Millisecond)

	// goroutine to update view
	go func() {
		for _ = range broadcastViewTicker {
			view := e.createCurrentView()
			if err := e.Chain.PushMessageToValidator(&view); err != nil {
				fmt.Println(err)
			}
		}
	}()

	// goroutine to update view
	go func() {
		for {
			<-ticker
			// check if chain is ready
			e.IsReady = e.Chain.IsReady()
			// check if round is timeout
			if e.IsReady && e.getTimeSinceLastBlock() > TIMEOUT && e.getCurrentRound() != e.CurrentState.View.Round {
				e.NextStateCh <- NEWROUND
			}

		}
	}()

	go func() {
		for { //action react pattern
			select {
			case s := <-e.NextStateCh:
				e.nextState(s)
			case <-e.ProposeMsgCh:
			case <-e.PrepareMsgCh:
			case <-e.CommitMsgCh:

			case view := <-e.ViewMsgCh: //update view of other nodes
				e.ValidatorsView[view.PeerID] = view

			}
		}
	}()

}

func (e *BFTEngine) nextState(nextState string) {
	if e.CurrentState.State == nextState {
		return //already transition
	}

	switch nextState {
	case PROPOSE:
		e.CurrentState.State = PROPOSE
		e.handleProposePhase()
	case LISTEN:
		e.CurrentState.State = LISTEN
		e.handleListenPhase()
	case PREPARE:
		e.CurrentState.State = PREPARE
		e.handlePreparePhase()
	case COMMIT:
		e.CurrentState.State = COMMIT
		e.handleCommitPhase()
	case NEWROUND:
		e.CurrentState.State = NEWROUND
		e.handleNewRoundPhase()
	}
}
