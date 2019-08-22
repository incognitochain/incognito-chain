package blsbft

import (
	"fmt"
	"time"
)

func (e *BLSBFT) getTimeSinceLastBlock() time.Duration {
	return time.Since(time.Unix(int64(e.Chain.GetLastBlockTimeStamp()), 0))
}

func (e *BLSBFT) waitForNextRound() {
	timeSinceLastBlk := e.getTimeSinceLastBlock()
	if timeSinceLastBlk > e.Chain.GetMinBlkInterval() {
		return
	}
	//TODO: chunk time sleep into small time chunk -> if change view during sleep => break it
	time.Sleep(e.Chain.GetMinBlkInterval() - timeSinceLastBlk)
}

func (e *BLSBFT) setState(state string) {
	e.RoundData.State = state
}

func (e *BLSBFT) getCurrentRound() int {
	return int(e.getTimeSinceLastBlock().Seconds() / TIMEOUT.Seconds())
}

func (e *BLSBFT) isInTimeFrame() bool {
	if e.Chain.CurrentHeight()+1 != e.RoundData.NextHeight {
		return false
	}
	if e.getTimeSinceLastBlock() > TIMEOUT && e.getCurrentRound() != e.RoundData.Round {
		return false
	}
	return true
}

func (e *BLSBFT) isHasMajorityVotes() bool {
	size := e.Chain.GetCommitteeSize()
	if len(e.RoundData.Votes) >= 2*size/3 {
		return true
	}
	return false
}

func getRoundKey(nextHeight uint64, round int) string {
	return fmt.Sprint(nextHeight, "_", round)
}
