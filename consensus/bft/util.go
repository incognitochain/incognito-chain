package bft

import (
	"time"
)

func (e *BFTCore) getTimeSinceLastBlock() time.Duration {
	return time.Since(time.Unix(int64(e.Chain.GetLastBlockTimeStamp()), 0))
}

func (e *BFTCore) waitForNextRound() {
	timeSinceLastBlk := e.getTimeSinceLastBlock()
	if timeSinceLastBlk > e.Chain.GetBlkMinTime() {
		return
	}
	//TODO: chunk time sleep into small time chunk -> if change view during sleep => break it
	time.Sleep(e.Chain.GetBlkMinTime() - timeSinceLastBlk)
}

func (e *BFTCore) setState(state string) {
	e.State = state
}

func (e *BFTCore) getCurrentRound() uint64 {
	return uint64(e.getTimeSinceLastBlock().Seconds() / TIMEOUT.Seconds())
}

func (e *BFTCore) isInTimeFrame() bool {
	if e.Chain.GetHeight()+1 != e.NextHeight {
		return false
	}
	if e.getTimeSinceLastBlock() > TIMEOUT && e.getCurrentRound() != e.Round {
		return false
	}
	return true
}

func (e *BFTCore) getMajorityVote(s map[string]bool) int {
	size := e.Chain.GetCommitteeSize()
	approve := 0
	reject := 0
	for _, v := range s {
		if v {
			approve++
		} else {
			reject++
		}
	}
	if approve > 2*size/3 {
		return 1
	}
	if reject > 2*size/3 {
		return -1
	}
	return 0
}
