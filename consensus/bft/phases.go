package bft

import (
	"encoding/json"
	"fmt"
)

// create new block (sequence number)
func (e *BFTCore) enterProposePhase() {
	if !e.isInTimeFrame() || e.State == PROPOSE {
		return //not in right time frame or already in propose phase
	}
	e.setState(PROPOSE)

	block := e.Chain.CreateNewBlock(int(e.Round))
	e.Block = block
	e.debug("start propose block", block)

	blockData, _ := json.Marshal(e.Block)
	msg, _ := MakeBFTProposeMsg(string(blockData), e.ChainKey, fmt.Sprint(e.NextHeight, "_", e.Round), e.UserKeySet)
	go e.Chain.PushMessageToValidator(msg)
	e.enterPreparePhase()

}

//listen for block
func (e *BFTCore) enterListenPhase() {
	if !e.isInTimeFrame() || e.State == LISTEN {
		return //not in right time frame or already in listen phase
	}
	e.setState(LISTEN)
	//e.debug("start listen block")
}

//send prepare message (signature of that message & sequence number) and wait for > 2/3 signature of nodes
//block for the message and sequence number
func (e *BFTCore) enterPreparePhase() {
	if !e.isInTimeFrame() || e.State == PREPARE {
		return //not in right time frame or already in prepare phase
	}
	e.setState(PREPARE)
	//e.debug("start prepare phase")
	//TODO: validate block isOK???

	msg, _ := MakeBFTPrepareMsg(true, e.ChainKey, e.Block.Hash().String(), fmt.Sprint(e.NextHeight, "_", e.Round), e.UserKeySet)
	go e.Chain.PushMessageToValidator(msg)
}

func (e *BFTCore) enterNewRound() {
	//if chain is not ready, return
	if !e.Chain.IsReady() {
		return
	}

	//if already running a round for current timeframe
	if e.isInTimeFrame() && e.State != NEWROUND {
		return
	}

	e.setState(NEWROUND)

	//wait for min blk time
	e.waitForNextRound()

	//move to next phase

	//create new round
	e.Round = e.getCurrentRound()
	e.NextHeight = e.Chain.GetHeight() + 1
	e.Block = nil

	if e.Chain.GetNodePubKeyIndex() == (e.Chain.GetLastProposerIndex()+1+int(e.Round))%e.Chain.GetCommitteeSize() {
		e.enterProposePhase()
	} else {
		e.enterListenPhase()
	}

}
