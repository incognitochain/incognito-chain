package bft2

// create new block (sequence number)
func (e *BFTEngine) handleProposePhase() {
	if !e.viewIsInTimeFrame() || e.State == PROPOSE {
		return //not in right time frame or already in propose phase
	}
	e.setState(PROPOSE)

	block := e.Chain.CreateNewBlock()
	//TODO: broadcast block
	e.Block = block
	e.debug("start propose block", block)

	//TODO: broadcast block

	e.nextState(PREPARE)

}

//listen for block
func (e *BFTEngine) handleListenPhase() {
	if !e.viewIsInTimeFrame() || e.State == LISTEN {
		return //not in right time frame or already in listen phase
	}
	e.setState(LISTEN)
	e.debug("start listen block")
}

//send prepare message (signature of that message & sequence number) and wait for > 2/3 signature of nodes
//block for the message and sequence number
func (e *BFTEngine) handlePreparePhase() {
	if !e.viewIsInTimeFrame() || e.State == PREPARE {
		return //not in right time frame or already in prepare phase
	}
	e.setState(PREPARE)

	e.debug("start prepare phase")

	//initiate current view
	//e.View = View{}
	////validate node with current view
	//isValid := false
	//if e.Block != nil {
	//	isValid = e.validateBlockWithCurrentView(e.Block, e.View)
	//}

	e.PrepareMsgCh <- PrepareMsg{true}
	//TODO: create signature and broadcast
}

//broadcast handleCommitPhase for a block
//for those who dont know which state it is/or he de-sync from network, 2/3 handleCommitPhase message will show him
func (e *BFTEngine) handleCommitPhase() {
	if !e.viewIsInTimeFrame() || e.State == COMMIT {
		return //not in right time frame or already in commit phase
	}
	e.setState(COMMIT)

	//There are replicas (non-faulty or otherwise) that didn't receive enough (i.e. 2f+1) PREPARE messages, either due to lossy network or being offline for a while. For them, they can't reach PREPARED stage. But! But when they heard from 2f+1 replicas broadcasting COMMIT message, they could be certain to handleCommitPhase on (m,v,n,i)
	//TODO: broadcast

	//TODO: if certain that block is handleCommitPhase , proceed to next state

	//TODO: if block is commit, then insert block to chain and broadcast block
}

func (e *BFTEngine) handleNewRoundPhase() {
	//if chain is not ready
	if !e.Chain.IsReady() {
		return
	}

	//if already running a round for current timeframe
	if e.viewIsInTimeFrame() && e.State != NEWROUND {
		return
	}
	e.setState(NEWROUND)

	//wait for min blk time
	e.waitForNextRound()

	//move to next phase

	//create new round
	e.Round = e.getCurrentRound()
	e.NextHeight = e.Chain.GetHeight() + 1
	if e.Chain.GetNodePubKeyIndex() == (e.Round % e.Chain.GetCommitteeSize()) {
		e.nextState(PROPOSE)
	} else {
		e.nextState(LISTEN)
	}

}
