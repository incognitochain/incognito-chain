package bft2

import (
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"math"
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
	Role  Role //role of node
	Round int
}

type NextStateCh struct {
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
	State string //state

	View  View  //view when creating/listening block
	Block Block //message

	isOk bool //vote for this message

	PrePrepareMsg []PrePrepareMsg
	PrepareMsg    []PrePrepareMsg
	Commit        []PrePrepareMsg
}

type BFTEngine struct {
	UserKeySet      *cashec.KeySet
	Server          serverInterface
	newStateCh      chan NextStateCh
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

	ticker := time.Tick(5 * time.Second)
	go func() {
		for _ = range ticker {
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
		case <-e.ViewMsgCh: //update view of other nodes
			//TODO: check for node role and start propose or handleListenPhase
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

// create new block
func (e *BFTEngine) handleProposePhase(s *BFTState) {
	time.AfterFunc(PROPOSE_TIMEOUT, func() {
		e.newStateCh <- NextStateCh{s, PREPREPARE}
	})

	//initiate current view
	s.View = View{}
	//create new block from current view
	block := e.createNewBlockFromCurrentView(s.View)
	//TODO: broadcast block
	s.Block = block
	e.newStateCh <- NextStateCh{s, PREPREPARE}
}

//listen for block
func (e *BFTEngine) handleListenPhase(s *BFTState) {
	time.AfterFunc(LISTEN_TIMEOUT, func() {
		if s.State == LISTEN {
			e.newStateCh <- NextStateCh{s, PREPREPARE}
		}
	})
}

// send pre-prepare message (sequence number)
// wait for more than > 2/3 nodes
func (e *BFTEngine) handlePrePreparePhase(s *BFTState) {
	time.AfterFunc(PRE_PREPARE_TIMEOUT, func() {
		e.newStateCh <- NextStateCh{s, PREPARE}
	})

	//initiate current view
	s.View = View{}
	//validate node with current view
	isValid := false
	if s.Block != nil {
		isValid = e.validateBlockWithCurrentView(s.Block, s.View)
	}

	e.PrePrepareMsgCh <- PrePrepareMsg{isValid}
	s.isOk = isValid
	//TODO: broadcast
}

//send prepare message (signature of that message & sequence number) and wait for > 2/3 signature of nodes
//block for the message and sequence number
func (e *BFTEngine) handlePreparePhase(s *BFTState) {
	time.AfterFunc(PREPARE_TIMEOUT, func() {
		e.newStateCh <- NextStateCh{s, COMMIT}
	})
	//TODO: create signature and broadcast
}

//broadcast handleCommitPhase for a block
//for those who dont know which state it is/or he de-sync from network, 2/3 handleCommitPhase message will show him
func (e *BFTEngine) handleCommitPhase(s *BFTState) {
	time.AfterFunc(COMMIT_TIMEOUT, func() {
		e.newStateCh <- NextStateCh{s, NEWROUND}
	})
	//There are replicas (non-faulty or otherwise) that didn't receive enough (i.e. 2f+1) PREPARE messages, either due to lossy network or being offline for a while. For them, they can't reach PREPARED stage. But! But when they heard from 2f+1 replicas broadcasting COMMIT message, they could be certain to handleCommitPhase on (m,v,n,i)
	//TODO: broadcast

	//TODO: if certain that block is handleCommitPhase , proceed to next state

	//TODO: if block is commit, then insert block to chain and broadcast block
}

func (e *BFTEngine) handleNewRoundPhase(s *BFTState) {
	//wait for min block time
}

func (e *BFTEngine) validateBlockWithCurrentView(b Block, v View) bool {
	return true
}

func (e *BFTEngine) createNewBlockFromCurrentView(v View) Block {
	return "sd"
}

func (e *BFTEngine) createCurrentView() View {
	view := View{}
	nodeType, shardID := blockchain.GetBestStateBeacon().GetPubkeyNodeRole(e.UserKeySet.GetPublicKeyB58())
	round := e.predictCurrentRound(nodeType, shardID)
	role := blockchain.GetBestStateBeacon().GetPubkeyChainRole(e.UserKeySet.GetPublicKeyB58(), round)
	view.Role = Role{nodeType, role, shardID}
	view.Round = round
	return view
}

func (e *BFTEngine) predictCurrentRound(nodeType string, shardID byte) int {
	timeSinceLastBlk := time.Since(time.Unix(0, 0))
	if nodeType == common.BEACON_ROLE {
		timeSinceLastBlk = time.Since(time.Unix(blockchain.GetBestStateBeacon().BestBlock.Header.Timestamp, 0))
	} else if nodeType == common.SHARD_ROLE {
		timeSinceLastBlk = time.Since(time.Unix(blockchain.GetBestStateShard(shardID).BestBlock.Header.Timestamp, 0))
	} else {
		return 0
	}

	roundTime := LISTEN_TIMEOUT + PRE_PREPARE_TIMEOUT + PREPARE_TIMEOUT + COMMIT_TIMEOUT
	if timeSinceLastBlk > roundTime {
		return int(math.Floor(timeSinceLastBlk.Seconds() / roundTime.Seconds()))
	}

	return 0
}
