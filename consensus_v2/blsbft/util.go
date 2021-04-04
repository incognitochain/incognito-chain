package blsbft

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/metrics/monitor"

	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/common"
)

func GetProposerIndexByRound(lastId, round, committeeSize int) int {
	//return (lastId + round) % committeeSize
	return 0
}

func (actorV1 *actorV1) getTimeSinceLastBlock() time.Duration {
	return time.Since(time.Unix(int64(actorV1.chain.GetLastBlockTimeStamp()), 0))
}

func (actorV1 *actorV1) waitForNextRound() bool {
	timeSinceLastBlk := actorV1.getTimeSinceLastBlock()
	if timeSinceLastBlk >= actorV1.chain.GetMinBlkInterval() {
		return false
	} else {
		//fmt.Println("\n\nWait for", e.Chain.GetMinBlkInterval()-timeSinceLastBlk, "\n\n")
		return true
	}
}

func (actorV1 *actorV1) setState(state string) {
	actorV1.roundData.state = state
}

func (actorV1 *actorV1) getCurrentRound() int {
	round := int((actorV1.getTimeSinceLastBlock().Seconds() - float64(actorV1.chain.GetMinBlkInterval().Seconds())) / timeout.Seconds())
	if round < 0 {
		return 1
	}

	return round + 1
}

func (actorV1 *actorV1) isInTimeFrame() bool {
	if actorV1.chain.CurrentHeight()+1 != actorV1.roundData.nextHeight {
		return false
	}

	if actorV1.getCurrentRound() != actorV1.roundData.round {
		return false
	}

	return true
}

func (actorV1 *actorV1) isHasMajorityVotes() bool {
	// e.RoundData.lockVotes.Lock()
	// defer e.RoundData.lockVotes.Unlock()
	actorV1.lockEarlyVotes.Lock()
	defer actorV1.lockEarlyVotes.Unlock()
	roundKey := getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round)
	earlyVote, ok := actorV1.earlyVotes[roundKey]
	committeeSize := len(actorV1.roundData.committee)
	if ok {
		wg := sync.WaitGroup{}
		blockHashBytes := actorV1.roundData.blockHash.GetBytes()
		for k, v := range earlyVote {
			wg.Add(1)
			go func(validatorKey string, voteData vote) {
				defer wg.Done()
				validatorIdx := common.IndexOfStr(validatorKey, actorV1.roundData.committeeBLS.stringList)
				if err := actorV1.preValidateVote(blockHashBytes, &voteData, actorV1.roundData.committee[validatorIdx].MiningPubKey[common.BridgeConsensus]); err == nil {
					actorV1.roundData.lockVotes.Lock()
					actorV1.roundData.votes[validatorKey] = voteData
					actorV1.roundData.lockVotes.Unlock()
				} else {
					actorV1.logger.Error(err)
				}
			}(k, v)
		}
		wg.Wait()
		if len(actorV1.roundData.votes) > 2*committeeSize/3 {
			delete(actorV1.earlyVotes, roundKey)
		}
	}
	monitor.SetGlobalParam("NVote", len(actorV1.roundData.votes))
	if len(actorV1.roundData.votes) > 2*committeeSize/3 {
		return true
	}
	return false
}

func getRoundKey(nextHeight uint64, round int) string {
	return fmt.Sprint(nextHeight, "_", round)
}

func parseRoundKey(roundKey string) (uint64, int) {
	stringArray := strings.Split(roundKey, "_")
	if len(stringArray) != 2 {
		return 0, 0
	}
	height, err := strconv.Atoi(stringArray[0])
	if err != nil {
		return 0, 0
	}
	round, err := strconv.Atoi(stringArray[1])
	if err != nil {
		return 0, 0
	}
	return uint64(height), round
}

func ExtractBridgeValidationData(block types.BlockInterface) ([][]byte, []int, error) {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return nil, nil, NewConsensusError(UnExpectedError, err)
	}
	return valData.BridgeSig, valData.ValidatiorsIdx, nil
}

func (actorV1 *actorV1) UpdateCommitteeBLSList() {
	committee := actorV1.chain.GetCommittee()
	if !reflect.DeepEqual(actorV1.roundData.committee, committee) {
		actorV1.roundData.committee = committee
		actorV1.roundData.committeeBLS.byteList = []blsmultisig.PublicKey{}
		actorV1.roundData.committeeBLS.stringList = []string{}
		for _, member := range actorV1.roundData.committee {
			actorV1.roundData.committeeBLS.byteList = append(actorV1.roundData.committeeBLS.byteList, member.MiningPubKey[consensusName])
		}
		committeeBLSString, err := incognitokey.ExtractPublickeysFromCommitteeKeyList(actorV1.roundData.committee, consensusName)
		if err != nil {
			actorV1.logger.Error(err)
			return
		}
		actorV1.roundData.committeeBLS.stringList = committeeBLSString
	}
}

func (actorV1 *actorV1) initRoundData() {
	roundKey := getRoundKey(actorV1.roundData.nextHeight, actorV1.roundData.round)
	if _, ok := actorV1.blocks[roundKey]; ok {
		delete(actorV1.blocks, roundKey)
	}
	actorV1.roundData.nextHeight = actorV1.chain.CurrentHeight() + 1
	actorV1.roundData.round = actorV1.getCurrentRound()
	actorV1.roundData.votes = make(map[string]vote)
	actorV1.roundData.block = nil
	actorV1.roundData.blockHash = common.Hash{}
	actorV1.roundData.notYetSendVote = true
	actorV1.roundData.timeStart = time.Now()
	actorV1.roundData.lastProposerIndex = actorV1.chain.GetLastProposerIndex()
	actorV1.UpdateCommitteeBLSList()
	actorV1.setState(newround)
}

func NewActorWithValue(
	chain, committeeChain blockchain.Chain,
	version, blockVersion int,
	chainID int, chainName string,
	node NodeInterface, logger common.Logger,
) Actor {
	var res Actor
	switch version {
	case BftVersion:
		res = NewActorV1WithValue(chain, chainName, chainID, node, logger)
	case MultiViewsVersion:
		res = NewActorV2WithValue(chain, committeeChain, chainName, chainID, blockVersion, node, logger)
	case SlashingVersion:
	case MultiSubsetsVersion:
	default:
		panic("Bft version is not valid")
	}
	return res
}
