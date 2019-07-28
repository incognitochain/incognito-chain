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
	if timeSinceLastBlk > e.Chain.GetBlkMinTime() {
		return
	}
	//TODO: chunk time sleep into small time chunk -> if change view during sleep => break it
	time.Sleep(e.Chain.GetBlkMinTime() - timeSinceLastBlk)
}

func (e *BLSBFT) setState(state string) {
	e.State = state
}

func (e *BLSBFT) getCurrentRound() int {
	return int(e.getTimeSinceLastBlock().Seconds() / TIMEOUT.Seconds())
}

func (e *BLSBFT) isInTimeFrame() bool {
	if e.Chain.GetHeight()+1 != e.NextHeight {
		return false
	}
	if e.getTimeSinceLastBlock() > TIMEOUT && e.getCurrentRound() != e.Round {
		return false
	}
	return true
}

func (e *BLSBFT) getMajorityVote(votes map[string]SigStatus) int {
	size := e.Chain.GetCommitteeSize()
	approve := 0
	reject := 0
	for k, v := range votes {

		if !v.Verified && e.MultiSigScheme.ValidateSingleSig(e.Block.Hash(), v.SigContent, k) != nil {
			delete(votes, k)
			continue
		}
		v.Verified = true

		if v.IsOk {
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

func (e *BLSBFT) validateAndSendVote() {
	if e.Chain.ValidateBlock(e.Block) == nil {
		msg, _ := MakeBFTPrepareMsg(true, e.ChainKey, e.Block.Hash().String(), fmt.Sprint(e.NextHeight, "_", e.Round), e.UserKeySet)
		go e.Chain.PushMessageToValidator(msg)
	} else {
		msg, _ := MakeBFTPrepareMsg(false, e.ChainKey, e.Block.Hash().String(), fmt.Sprint(e.NextHeight, "_", e.Round), e.UserKeySet)
		go e.Chain.PushMessageToValidator(msg)
	}
}
