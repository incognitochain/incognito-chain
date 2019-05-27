package bft2

import "time"

const (
	PROPOSE    = "PROPOSE"
	LISTEN     = "LISTEN"
	PREPREPARE = "PREPREPARE"
	PREPARE    = "PREPARE"
	COMMIT     = "COMMIT"
	REPLY      = "REPLY"
)

const (
	PROPOSE_TIMEOUT     = 30 * time.Second
	LISTEN_TIMEOUT      = 30 * time.Second
	PREPARE_TIMEOUT     = 30 * time.Second
	PRE_PREPARE_TIMEOUT = 30 * time.Second
	COMMIT_TIMEOUT      = 60 * time.Second
)

type Block interface {
	//validate() bool
}

type View struct {
	Role  string //role of node
	Round uint64
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

	BlockHeight uint64 //sequence number
	View        View   //view when creating/listening block
	Block       Block  //message
	NodeID      string //node id

	isOk bool //vote for this message

	PrePrepareMsg []PrePrepareMsg
	PrepareMsg    []PrePrepareMsg
	Commit        []PrePrepareMsg
}

var newStateCh chan BFTState
var proposedBlockCh chan Block
var prePrepareMsgCh chan PrePrepareMsg
var prepareMsgCh chan PrepareMsg
var commitMsgCh chan CommitMsg

func Start() {
	//var stateCache = make(map[uint64]BFTState)
	//var curentState = BFTState{}

	//TODO: check for node role and start propose or handleListenPhase

	for { //data flow
		select {
		case <-proposedBlockCh:
		case <-prePrepareMsgCh:
		case <-prepareMsgCh:
		case <-commitMsgCh:
		}
	}

}

func nextState(curState string, s *BFTState) {
	if s.State != curState {
		return // not in the same state (avoid change state multiple time)
	}

	switch s.State {
	case PROPOSE:
		s.State = PREPREPARE
		handlePrePreparePhase(s)
	case LISTEN:
		s.State = PREPREPARE
		handlePrePreparePhase(s)
	case PREPREPARE:
		s.State = PREPARE
		handlePreparePhase(s)
	case PREPARE:
		s.State = COMMIT
		handleCommitPhase(s)
	case COMMIT:
		//TODO: check view for transition to PROPOSE or LISTEN
		handleListenPhase(s)
	case REPLY:
		//TODO: check view for transition to PROPOSE or LISTEN
		handleReplyPhase(s)
	}
}

// create new block
func handleProposePhase(s *BFTState) {
	time.AfterFunc(PROPOSE_TIMEOUT, func() {
		nextState(PROPOSE, s)
	})

	//initiate current view
	s.View = View{}
	//create new block from current view
	block := createNewBlockFromCurrentView(s.View)
	//TODO: broadcast block
	s.Block = block
	nextState(PROPOSE, s)
}

//listen for block
func handleListenPhase(s *BFTState) {
	time.AfterFunc(LISTEN_TIMEOUT, func() {
		if s.State == LISTEN {
			nextState(LISTEN, s)
		}
	})
}

// send pre-prepare message (sequence number)
// wait for more than > 2/3 nodes
func handlePrePreparePhase(s *BFTState) {
	time.AfterFunc(PRE_PREPARE_TIMEOUT, func() {
		nextState(PREPREPARE, s)
	})

	//initiate current view
	s.View = View{}
	//validate node with current view
	isValid := false
	if s.Block != nil {
		isValid = validateBlockWithCurrentView(s.Block, s.View)
	}

	prePrepareMsgCh <- PrePrepareMsg{isValid}
	s.isOk = isValid
	//TODO: broadcast
}

//send prepare message (signature of that message & sequence number) and wait for > 2/3 signature of nodes
//block for the message and sequence number
func handlePreparePhase(s *BFTState) {
	time.AfterFunc(PREPARE_TIMEOUT, func() {
		nextState(PREPREPARE, s)
	})
	//TODO: create signature and broadcast
}

//broadcast handleCommitPhase for a block
//for those who dont know which state it is/or he de-sync from network, 2/3 handleCommitPhase message will show him
func handleCommitPhase(s *BFTState) {
	time.AfterFunc(COMMIT_TIMEOUT, func() {
		nextState(COMMIT, s)
	})
	//There are replicas (non-faulty or otherwise) that didn't receive enough (i.e. 2f+1) PREPARE messages, either due to lossy network or being offline for a while. For them, they can't reach PREPARED stage. But! But when they heard from 2f+1 replicas broadcasting COMMIT message, they could be certain to handleCommitPhase on (m,v,n,i)
	//TODO: broadcast

	//TODO: if certain that block is handleCommitPhase , proceed to next state
}

func handleReplyPhase(s *BFTState) {
	//TODO: if block is commit, then insert block to chain and broadcast block
}

func validateBlockWithCurrentView(b Block, v View) bool {
	return true
}

func createNewBlockFromCurrentView(v View) Block {
	return "sd"
}
