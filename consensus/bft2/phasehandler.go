package bft2

import "time"

// create new block
func (e *BFTEngine) handleProposePhase(s *BFTState) {
	time.AfterFunc(PROPOSE_TIMEOUT, func() {
		e.newStateCh <- NextState{s, PREPREPARE}
	})

	//initiate current view
	s.View = View{}
	//create new block from current view
	block := e.createNewBlockFromCurrentView(s.View)
	//TODO: broadcast block
	s.Block = block
	e.newStateCh <- NextState{s, PREPREPARE}
}

//listen for block
func (e *BFTEngine) handleListenPhase(s *BFTState) {
	time.AfterFunc(LISTEN_TIMEOUT, func() {
		if s.State == LISTEN {
			e.newStateCh <- NextState{s, PREPREPARE}
		}
	})
}

// send pre-prepare message (sequence number)
// wait for more than > 2/3 nodes
func (e *BFTEngine) handlePrePreparePhase(s *BFTState) {
	time.AfterFunc(PRE_PREPARE_TIMEOUT, func() {
		e.newStateCh <- NextState{s, PREPARE}
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
		e.newStateCh <- NextState{s, COMMIT}
	})
	//TODO: create signature and broadcast
}

//broadcast handleCommitPhase for a block
//for those who dont know which state it is/or he de-sync from network, 2/3 handleCommitPhase message will show him
func (e *BFTEngine) handleCommitPhase(s *BFTState) {
	time.AfterFunc(COMMIT_TIMEOUT, func() {
		e.newStateCh <- NextState{s, NEWROUND}
	})
	//There are replicas (non-faulty or otherwise) that didn't receive enough (i.e. 2f+1) PREPARE messages, either due to lossy network or being offline for a while. For them, they can't reach PREPARED stage. But! But when they heard from 2f+1 replicas broadcasting COMMIT message, they could be certain to handleCommitPhase on (m,v,n,i)
	//TODO: broadcast

	//TODO: if certain that block is handleCommitPhase , proceed to next state

	//TODO: if block is commit, then insert block to chain and broadcast block
}

func (e *BFTEngine) handleNewRoundPhase(s *BFTState) {
	//wait for min block time
}
