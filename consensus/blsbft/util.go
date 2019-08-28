package blsbft

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
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
	fmt.Println("\n\nSleep for", e.Chain.GetMinBlkInterval()-timeSinceLastBlk, "\n\n")

	time.Sleep(e.Chain.GetMinBlkInterval() - timeSinceLastBlk)
}

func (e *BLSBFT) setState(state string) {
	e.RoundData.State = state
}

func (e *BLSBFT) getCurrentRound() int {
	round := int(e.getTimeSinceLastBlock().Seconds() / TIMEOUT.Seconds())
	if round == 0 {
		return 1
	}
	return round
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
	earlyVote, ok := e.EarlyVotes[getRoundKey(e.RoundData.NextHeight, e.RoundData.Round)]
	if ok {
		for validator, vote := range earlyVote {
			validatorIdx := common.IndexOfStr(validator, e.RoundData.CommitteeBLS.StringList)
			if err := validateSingleBLSSig(e.RoundData.Block.Hash(), vote.BLS, validatorIdx, e.RoundData.CommitteeBLS.ByteList); err != nil {
				e.logger.Error(err)
				continue
			}
			e.RoundData.Votes[validator] = vote
		}
		delete(e.EarlyVotes, getRoundKey(e.RoundData.NextHeight, e.RoundData.Round))
	}
	size := len(e.RoundData.Committee)
	if len(e.RoundData.Votes) > 2*size/3 {
		return true
	}
	return false
}

func getRoundKey(nextHeight uint64, round int) string {
	return fmt.Sprint(nextHeight, "_", round)
}

func (e *BLSBFT) ExtractBridgeValidationData(block common.BlockInterface) ([][]byte, []int, error) {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return nil, nil, err
	}
	return valData.BridgeSig, valData.ValidatiorsIdx, nil
}
