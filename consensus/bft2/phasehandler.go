package bft2

// create new block (sequence number)
func (e *BFTEngine) handleProposePhase() {
	//initiate current view
	e.CurrentState.View = View{}
	//create new block from current view
	block := e.createNewBlockFromCurrentView(e.CurrentState.View)
	//TODO: broadcast block
	e.CurrentState.Block = block
	e.nextState(PREPARE)
}

//listen for block
func (e *BFTEngine) handleListenPhase() {

}

//send prepare message (signature of that message & sequence number) and wait for > 2/3 signature of nodes
//block for the message and sequence number
func (e *BFTEngine) handlePreparePhase() {

	//initiate current view
	e.CurrentState.View = View{}
	//validate node with current view
	isValid := false
	if e.CurrentState.Block != nil {
		isValid = e.validateBlockWithCurrentView(e.CurrentState.Block, e.CurrentState.View)
	}

	e.PrepareMsgCh <- PrepareMsg{isValid}
	e.CurrentState.isOk = isValid
	//TODO: create signature and broadcast
}

//broadcast handleCommitPhase for a block
//for those who dont know which state it is/or he de-sync from network, 2/3 handleCommitPhase message will show him
func (e *BFTEngine) handleCommitPhase() {

	//There are replicas (non-faulty or otherwise) that didn't receive enough (i.e. 2f+1) PREPARE messages, either due to lossy network or being offline for a while. For them, they can't reach PREPARED stage. But! But when they heard from 2f+1 replicas broadcasting COMMIT message, they could be certain to handleCommitPhase on (m,v,n,i)
	//TODO: broadcast

	//TODO: if certain that block is handleCommitPhase , proceed to next state

	//TODO: if block is commit, then insert block to chain and broadcast block
}

func (e *BFTEngine) handleNewRoundPhase() {
	if !e.IsReady {
		return
	}

	//wait for min blk time
	e.waitForNextRound()

}
