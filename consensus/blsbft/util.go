package blsbft

import (
	"time"

	"github.com/incognitochain/incognito-chain/consensus/multisigschemes/bls"
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

func (e *BLSBFT) getMajorityVote(votes map[string]SigStatus) int {
	size := e.Chain.GetCommitteeSize()
	approve := 0
	reject := 0
	for k, v := range votes {
		if !v.Verified && bls.ValidateSingleSig(e.RoundData.Block.Hash(), v.SigContent, k) != nil {
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

func (e *BLSBFT) sendVote() {
	sig, _ := e.UserKeySet.SignData(e.RoundData.Block.Hash())
	MakeBFTVoteMsg(e.UserKeySet, e.ChainKey, sig, getRoundKey(e.RoundData.NextHeight, e.RoundData.Round))
	// go e.Node.PushMessageToChain(msg)
	e.RoundData.NotYetSendVote = false
}

// func (e *BLSBFT) getTimeout() time.Duration {
// 	return
// }
