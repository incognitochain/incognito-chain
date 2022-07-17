package blsbft

import (
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
)

type IConsensusValidator interface {
	FilterValidProposeBlockInfo(bestViewHash common.Hash, bestViewHeight uint64, finalViewHeight uint64, currentTimeSlot int64, proposeBlockInfos map[string]*ProposeBlockInfo) ([]*ProposeBlockInfo, []*ProposeBlockInfo, []string)
	ValidateBlock(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) (bool, error)
	ValidateConsensusRules(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) bool
}

type ConsensusValidatorLemma2 struct {
	logger common.Logger
	chain  Chain
}

func NewConsensusValidatorLemma2(logger common.Logger, chain Chain) *ConsensusValidatorLemma2 {
	return &ConsensusValidatorLemma2{logger: logger, chain: chain}
}

// FilterValidProposeBlockInfo validate received proposed block and return valid proposed block
// Special case: in case block is already inserted, try to send vote (avoid slashing)
// 1. by pass nil block
// 2. just validate recently
// 3. not in current time slot
// 4. not connect to best view
func (c ConsensusValidatorLemma2) FilterValidProposeBlockInfo(bestViewProposeHash common.Hash, bestViewHeight uint64, finalViewHeight uint64, currentTimeSlot int64, proposeBlockInfos map[string]*ProposeBlockInfo) ([]*ProposeBlockInfo, []*ProposeBlockInfo, []string) {
	//Check for valid block to vote
	validProposeBlock := []*ProposeBlockInfo{}
	tryReVoteInsertedBlock := []*ProposeBlockInfo{}
	invalidProposeBlock := []string{}
	//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
	for h, proposeBlockInfo := range proposeBlockInfos {
		if proposeBlockInfo.block == nil {
			continue
		}

		previousBlockHash := proposeBlockInfo.block.GetPrevHash()
		previousView := c.chain.GetMultiView().GetViewByHash(previousBlockHash)
		if previousView == nil {
			continue
		}

		//must link to bestview, we expect all node having same block data having same bestview
		if previousBlockHash.String() != c.chain.GetBestViewHash() {
			continue
		}

		//special case: if we insert block too quick, before voting
		//=> vote for this block (within TS,but block is inserted into bestview)
		//this special case by pass validate with consensus rules
		if proposeBlockInfo.block.GetHeight() == bestViewHeight &&
			proposeBlockInfo.block.ProposeHash().IsEqual(&bestViewProposeHash) &&
			!proposeBlockInfo.IsVoted {
			tryReVoteInsertedBlock = append(tryReVoteInsertedBlock, proposeBlockInfo)
			continue
		}

		//not validate if we do it recently
		if time.Since(proposeBlockInfo.LastValidateTime).Seconds() < 1 {
			continue
		}

		// check if propose block in within TS
		if previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) != currentTimeSlot {
			continue
		}

		//if the block height is not next height or current height
		if proposeBlockInfo.block.GetHeight() != bestViewHeight+1 {
			continue
		}

		// check if producer time > proposer time
		if previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime()) > currentTimeSlot {
			continue
		}

		//finality check
		if proposeBlockInfo.block.GetFinalityHeight() != 0 {
			if !proposeBlockInfo.IsValidLemma2Proof {
				c.logger.Errorf("Reject block %+v as invalid lemma2 block, but finality height is set", proposeBlockInfo.block.FullHashString())
				continue
			}

			if previousView.GetBlock().GetVersion() >= types.INSTANT_FINALITY_VERSION {
				previousProposeTimeSlot := previousView.CalculateTimeSlot(previousView.GetBlock().GetProposeTime())
				previousProduceTimeSlot := previousView.CalculateTimeSlot(previousView.GetBlock().GetProduceTime())
				if previousView.GetBlock().GetFinalityHeight() == 0 && previousProposeTimeSlot != previousProduceTimeSlot {
					c.logger.Errorf("Reject block %+v as previous block finality height not set (%+v) or produce/propose not the same (%+v)", proposeBlockInfo.block.FullHashString(), previousView.GetBlock().GetFinalityHeight(), previousProposeTimeSlot, previousProduceTimeSlot)
					continue
				}
			}
		}

		if proposeBlockInfo.block.GetFinalityHeight() == 0 {
			if previousView.GetBlock().GetVersion() >= types.INSTANT_FINALITY_VERSION {
				previousProposeTimeSlot := previousView.CalculateTimeSlot(previousView.GetBlock().GetProposeTime())
				previousProduceTimeSlot := previousView.CalculateTimeSlot(previousView.GetBlock().GetProduceTime())
				if proposeBlockInfo.IsValidLemma2Proof && (previousView.GetBlock().GetFinalityHeight() != 0 || previousProposeTimeSlot == previousProduceTimeSlot) {
					c.logger.Errorf("Reject block %+v as this block should set finality height", proposeBlockInfo.block.FullHashString())
					continue
				}
			}
		}

		if proposeBlockInfo.block.GetHeight() < finalViewHeight {
			invalidProposeBlock = append(invalidProposeBlock, h)
			continue
		}

		validProposeBlock = append(validProposeBlock, proposeBlockInfo)
	}
	//rule 1: get history of vote for this height, vote if (round is lower than the vote before)
	// or (round is equal but new proposer) or (there is no vote for this height yet)
	sort.Slice(validProposeBlock, func(i, j int) bool {
		return validProposeBlock[i].block.GetProduceTime() < validProposeBlock[j].block.GetProduceTime()
	})

	return validProposeBlock, tryReVoteInsertedBlock, invalidProposeBlock
}

func (c ConsensusValidatorLemma2) ValidateBlock(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) (bool, error) {

	isValid := c.ValidateConsensusRules(lastVotedBlock, isVoteNextHeight, proposeBlockInfo)
	if !isValid {
		return isValid, nil
	}

	if proposeBlockInfo.IsVoted {
		return true, nil
	}

	if !proposeBlockInfo.IsValid {
		c.logger.Infof("validate block: %+v \n", proposeBlockInfo.block.FullHashString())
		if err := c.chain.ValidatePreSignBlock(proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, proposeBlockInfo.Committees); err != nil {
			c.logger.Error(err)
			return false, err
		}
	}

	return true, nil
}

func (c ConsensusValidatorLemma2) ValidateConsensusRules(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) bool {

	if !isVoteNextHeight {
		c.logger.Infof("Block %+v is valid with because no block in height is voted yet %+v",
			proposeBlockInfo.block.FullHashString(),
			proposeBlockInfo.block.GetHeight())
		return true
	}

	previousView := c.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())

	blockProduceTimeSlot := previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())
	lastBlockProduceTimeSlot := previousView.CalculateTimeSlot(lastVotedBlock.GetProduceTime())
	if blockProduceTimeSlot < lastBlockProduceTimeSlot {
		// blockProduceTimeSlot is smaller than voted block => vote for this block
		c.logger.Infof("Block %+v is valid with rule 1, Block Produce Time %+v, < Last Block Produce Time %+v",
			proposeBlockInfo.block.FullHashString(), blockProduceTimeSlot, lastBlockProduceTimeSlot)
		return true
	} else if blockProduceTimeSlot == lastBlockProduceTimeSlot &&
		previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) > previousView.CalculateTimeSlot(lastVotedBlock.GetProposeTime()) {
		c.logger.Infof("Block %+v is valid with rule 2, Block Propose Time %+v, < Last Block Propose Time %+v",
			proposeBlockInfo.block.FullHashString(),
			previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()),
			previousView.CalculateTimeSlot(lastVotedBlock.GetProposeTime()))
		// block is old block (same round), but new proposer(larger timeslot) => vote again
		return true
	} else if proposeBlockInfo.block.CommitteeFromBlock().String() != lastVotedBlock.CommitteeFromBlock().String() {
		c.logger.Infof("Block %+v is valid with rule 3, Block Produce Time %+v, < Last Block Produce Time %+v",
			proposeBlockInfo.block.FullHashString(),
			blockProduceTimeSlot, lastBlockProduceTimeSlot)
		// blockProduceTimeSlot is larger or equal than voted block
		return true
	} // if not swap committees => do nothing

	c.logger.Infof("ValidateConsensusRules failed, hash %+v, height %+v | "+
		"blockProduceTs %+v, lastBlockProduceTs %+v |"+
		"blockProposeTs %+v, lastBlockProposeTs %+v | "+
		"isSameCommittee %+v",
		proposeBlockInfo.block.FullHashString(),
		proposeBlockInfo.block.GetHeight(),
		blockProduceTimeSlot,
		lastBlockProduceTimeSlot,
		previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()),
		previousView.CalculateTimeSlot(lastVotedBlock.GetProposeTime()),
		proposeBlockInfo.block.CommitteeFromBlock().String() == lastVotedBlock.CommitteeFromBlock().String(),
	)

	return false
}

type ConsensusValidatorLemma1 struct {
	logger common.Logger
	chain  Chain
}

func NewConsensusValidatorLemma1(logger common.Logger, chain Chain) *ConsensusValidatorLemma1 {
	return &ConsensusValidatorLemma1{logger: logger, chain: chain}
}

func (c ConsensusValidatorLemma1) FilterValidProposeBlockInfo(bestViewHash common.Hash, bestViewHeight uint64, finalViewHeight uint64, currentTimeSlot int64, proposeBlockInfos map[string]*ProposeBlockInfo) ([]*ProposeBlockInfo, []*ProposeBlockInfo, []string) {
	//Check for valid block to vote
	validProposeBlock := []*ProposeBlockInfo{}
	tryReVoteInsertedBlock := []*ProposeBlockInfo{}
	invalidProposeBlock := []string{}

	//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
	for h, proposeBlockInfo := range proposeBlockInfos {
		if proposeBlockInfo.block == nil {
			continue
		}
		previousView := c.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())
		if previousView == nil {
			continue
		}

		//// check if this time slot has been voted
		//if a.votedTimeslot[c.chain.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime())] {
		//	continue
		//}

		//special case: if we insert block too quick, before voting
		//=> vote for this block (within TS,but block is inserted into bestview)
		//this special case by pass validate with consensus rules
		if proposeBlockInfo.block.GetHeight() == bestViewHeight &&
			proposeBlockInfo.block.Hash().IsEqual(&bestViewHash) &&
			!proposeBlockInfo.IsVoted {
			tryReVoteInsertedBlock = append(tryReVoteInsertedBlock, proposeBlockInfo)
			continue
		}

		//not validate if we do it recently
		if time.Since(proposeBlockInfo.LastValidateTime).Seconds() < 1 {
			continue
		}

		// check if propose block in within TS
		if previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) != currentTimeSlot {
			continue
		}

		//if the block height is not next height or current height
		if proposeBlockInfo.block.GetHeight() != bestViewHeight+1 {
			continue
		}

		// check if producer time > proposer time
		if previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime()) > currentTimeSlot {
			continue
		}

		if proposeBlockInfo.block.GetHeight() < finalViewHeight {
			invalidProposeBlock = append(invalidProposeBlock, h)
			continue
		}

		validProposeBlock = append(validProposeBlock, proposeBlockInfo)
	}
	//rule 1: get history of vote for this height, vote if (round is lower than the vote before)
	// or (round is equal but new proposer) or (there is no vote for this height yet)
	sort.Slice(validProposeBlock, func(i, j int) bool {
		return validProposeBlock[i].block.GetProduceTime() < validProposeBlock[j].block.GetProduceTime()
	})

	return validProposeBlock, tryReVoteInsertedBlock, invalidProposeBlock
}

func (c ConsensusValidatorLemma1) ValidateBlock(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) (bool, error) {

	isValid := c.ValidateConsensusRules(lastVotedBlock, isVoteNextHeight, proposeBlockInfo)
	if !isValid {
		return isValid, nil
	}

	if proposeBlockInfo.IsVoted {
		return true, nil
	}

	if !proposeBlockInfo.IsValid {
		c.logger.Infof("validate block: %+v \n", proposeBlockInfo.block.FullHashString())
		if err := c.chain.ValidatePreSignBlock(proposeBlockInfo.block, proposeBlockInfo.SigningCommittees, proposeBlockInfo.Committees); err != nil {
			c.logger.Error(err)
			return false, err
		}
	}

	return true, nil
}

func (c ConsensusValidatorLemma1) ValidateConsensusRules(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) bool {

	if !isVoteNextHeight {
		c.logger.Infof("Block %+v is valid with because no block in height is voted yet %+v",
			proposeBlockInfo.block.FullHashString(),
			proposeBlockInfo.block.GetHeight())
		return true
	}

	previousView := c.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())
	blockProduceTimeSlot := previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProduceTime())
	lastBlockProduceTimeSlot := previousView.CalculateTimeSlot(lastVotedBlock.GetProduceTime())

	if blockProduceTimeSlot < previousView.CalculateTimeSlot(lastVotedBlock.GetProduceTime()) {
		c.logger.Infof("Block %+v is valid with rule 1, Block Produce Time %+v, < Last Block Produce Time %+v",
			proposeBlockInfo.block.FullHashString(), blockProduceTimeSlot, lastBlockProduceTimeSlot)
		// blockProduceTimeSlot is smaller than voted block => vote for this block
		return true
	} else if blockProduceTimeSlot == previousView.CalculateTimeSlot(lastVotedBlock.GetProduceTime()) &&
		previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) > previousView.CalculateTimeSlot(lastVotedBlock.GetProposeTime()) {
		c.logger.Infof("Block %+v is valid with rule 2, Block Propose Time %+v, < Last Block Propose Time %+v",
			proposeBlockInfo.block.FullHashString(),
			previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()),
			previousView.CalculateTimeSlot(lastVotedBlock.GetProposeTime()))
		// block is old block (same round), but new proposer(larger timeslot) => vote again
		return true
	} else if proposeBlockInfo.block.CommitteeFromBlock().String() != lastVotedBlock.CommitteeFromBlock().String() {
		// blockProduceTimeSlot is larger or equal than voted block
		c.logger.Infof("Block %+v is valid with rule 3, Block Produce Time %+v, < Last Block Produce Time %+v",
			proposeBlockInfo.block.FullHashString(),
			blockProduceTimeSlot, lastBlockProduceTimeSlot)
		return true
	} // if not swap committees => do nothing

	c.logger.Infof("ValidateConsensusRules failed, hash %+v, height %+v | "+
		"blockProduceTs %+v, lastBlockProduceTs %+v |"+
		"blockProposeTs %+v, lastBlockProposeTs %+v | "+
		"isSameCommittee %+v",
		proposeBlockInfo.block.FullHashString(),
		proposeBlockInfo.block.GetHeight(),
		blockProduceTimeSlot,
		lastBlockProduceTimeSlot,
		previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()),
		previousView.CalculateTimeSlot(lastVotedBlock.GetProposeTime()),
		proposeBlockInfo.block.CommitteeFromBlock().String() == lastVotedBlock.CommitteeFromBlock().String(),
	)

	return false
}

type ConsensusValidatorNoValidate struct {
}

func NewConsensusValidatorNoValidate() *ConsensusValidatorNoValidate {
	return &ConsensusValidatorNoValidate{}
}

func (c ConsensusValidatorNoValidate) FilterValidProposeBlockInfo(bestViewHash common.Hash, bestViewHeight uint64, finalViewHeight uint64, currentTimeSlot int64, proposeBlockInfos map[string]*ProposeBlockInfo) ([]*ProposeBlockInfo, []*ProposeBlockInfo, []string) {
	validProposeBlockInfo := []*ProposeBlockInfo{}
	for _, v := range proposeBlockInfos {
		validProposeBlockInfo = append(validProposeBlockInfo, v)
	}
	return validProposeBlockInfo, []*ProposeBlockInfo{}, []string{}
}

func (c ConsensusValidatorNoValidate) ValidateBlock(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) (bool, error) {
	return true, nil
}

func (c ConsensusValidatorNoValidate) ValidateConsensusRules(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) bool {
	return true
}
