package bft2

import (
	"log"
	"time"
)

func (e *BFTEngine) getTimeSinceLastBlock() time.Duration {
	return time.Since(time.Unix(int64(e.Chain.GetLastBlockTimeStamp()), 0))
}

func (e *BFTEngine) waitForNextRound() {
	timeSinceLastBlk := e.getTimeSinceLastBlock()
	if timeSinceLastBlk > e.Chain.GetBlkMinTime() {
		return
	}
	time.Sleep(e.Chain.GetBlkMinTime() - timeSinceLastBlk)
}

func (e *BFTEngine) setState(state string) {
	e.State = state
}

/*
Return the round is calculated since the latest block time
*/
func (e *BFTEngine) getCurrentRound() uint64 {
	return uint64(e.getTimeSinceLastBlock().Seconds() / TIMEOUT.Seconds())
}

func (e *BFTEngine) viewIsInTimeFrame() bool {
	if e.Chain.GetHeight()+1 != e.NextHeight {
		return false
	}

	if e.getTimeSinceLastBlock() > TIMEOUT && e.getCurrentRound() != e.Round {
		return false
	}
	return true
}

func (e *BFTEngine) getMajorityVote(s map[string]bool) int {
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
		//e.debug("Approve", approve)
		return 1
	}
	if reject > 2*size/3 {
		return -1
	}
	return 0
}

func (e *BFTEngine) debug(s ...interface{}) {
	//if e.PeerID == "1" {
	s = append([]interface{}{"Peer " + e.PeerID + ": "}, s...)
	log.Println(s...)
	//}

}
