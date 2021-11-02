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
func (c ConsensusValidatorLemma2) FilterValidProposeBlockInfo(bestViewHash common.Hash, bestViewHeight uint64, finalViewHeight uint64, currentTimeSlot int64, proposeBlockInfos map[string]*ProposeBlockInfo) ([]*ProposeBlockInfo, []*ProposeBlockInfo, []string) {
	//Check for valid block to vote
	validProposeBlock := []*ProposeBlockInfo{}
	tryReVoteInsertedBlock := []*ProposeBlockInfo{}
	invalidProposeBlock := []string{}
	//get all block that has height = bestview height  + 1(rule 2 & rule 3) (
	for h, proposeBlockInfo := range proposeBlockInfos {
		if proposeBlockInfo.Block == nil {
			continue
		}

		//// check if this time slot has been voted
		//if a.votedTimeslot[common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime())] {
		//	continue
		//}

		//special case: if we insert block too quick, before voting
		//=> vote for this block (within TS,but block is inserted into bestview)
		//this special case by pass validate with consensus rules
		if proposeBlockInfo.Block.GetHeight() == bestViewHeight &&
			proposeBlockInfo.Block.Hash().IsEqual(&bestViewHash) &&
			!proposeBlockInfo.IsVoted {
			tryReVoteInsertedBlock = append(tryReVoteInsertedBlock, proposeBlockInfo)
			continue
		}

		//not validate if we do it recently
		if time.Since(proposeBlockInfo.LastValidateTime).Seconds() < 1 {
			continue
		}

		// check if propose block in within TS
		if common.CalculateTimeSlot(proposeBlockInfo.Block.GetProposeTime()) != currentTimeSlot {
			continue
		}

		//if the block height is not next height or current height
		if proposeBlockInfo.Block.GetHeight() != bestViewHeight+1 {
			continue
		}

		// check if producer time > proposer time
		if common.CalculateTimeSlot(proposeBlockInfo.Block.GetProduceTime()) > currentTimeSlot {
			continue
		}

		// lemma 2
		if proposeBlockInfo.IsValidLemma2Proof {
			if proposeBlockInfo.Block.GetFinalityHeight() != proposeBlockInfo.Block.GetHeight()-1 {
				c.logger.Errorf("Block %+v %+v, is valid for lemma 2, expect finality height %+v, got %+v",
					proposeBlockInfo.Block.GetHeight(), proposeBlockInfo.Block.Hash().String(),
					proposeBlockInfo.Block.GetHeight(), proposeBlockInfo.Block.GetFinalityHeight())
				continue
			}
		}
		if !proposeBlockInfo.IsValidLemma2Proof {
			if proposeBlockInfo.Block.GetFinalityHeight() != 0 {
				c.logger.Errorf("Block %+v %+v, root hash %+v, previous block hash %+v, is invalid for lemma 2, expect finality height %+v, got %+v",
					proposeBlockInfo.Block.GetHeight(), proposeBlockInfo.Block.Hash().String(),
					proposeBlockInfo.Block.GetAggregateRootHash(), proposeBlockInfo.Block.GetPrevHash().String(),
					0, proposeBlockInfo.Block.GetFinalityHeight())
				continue
			}
		}

		if proposeBlockInfo.Block.GetHeight() < finalViewHeight {
			invalidProposeBlock = append(invalidProposeBlock, h)
			continue
		}

		validProposeBlock = append(validProposeBlock, proposeBlockInfo)
	}
	//rule 1: get history of vote for this height, vote if (round is lower than the vote before)
	// or (round is equal but new proposer) or (there is no vote for this height yet)
	sort.Slice(validProposeBlock, func(i, j int) bool {
		return validProposeBlock[i].Block.GetProduceTime() < validProposeBlock[j].Block.GetProduceTime()
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
		c.logger.Infof("validate block: %+v \n", proposeBlockInfo.Block.Hash().String())
		if err := c.chain.ValidatePreSignBlock(proposeBlockInfo.Block, proposeBlockInfo.SigningCommittees, proposeBlockInfo.Committees); err != nil {
			c.logger.Error(err)
			return false, err
		}
	}

	return true, nil
}

func (c ConsensusValidatorLemma2) ValidateConsensusRules(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) bool {

	if !isVoteNextHeight {
		return true
	}

	blockProduceTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.Block.GetProduceTime())
	lastBlockProduceTimeSlot := common.CalculateTimeSlot(lastVotedBlock.GetProduceTime())
	if blockProduceTimeSlot < lastBlockProduceTimeSlot {
		// blockProduceTimeSlot is smaller than voted block => vote for this block
		c.logger.Debug("Block Produce Time %+v, < Last Block Produce Time %+v", blockProduceTimeSlot, lastBlockProduceTimeSlot)
		return true
	} else if blockProduceTimeSlot == lastBlockProduceTimeSlot &&
		common.CalculateTimeSlot(proposeBlockInfo.Block.GetProposeTime()) > common.CalculateTimeSlot(lastVotedBlock.GetProposeTime()) {
		c.logger.Debug("Block Propose Time %+v, < Last Block Propose Time %+v",
			common.CalculateTimeSlot(proposeBlockInfo.Block.GetProposeTime()),
			common.CalculateTimeSlot(lastVotedBlock.GetProposeTime()))
		// block is old block (same round), but new proposer(larger timeslot) => vote again
		return true
	} else if proposeBlockInfo.Block.CommitteeFromBlock().String() != lastVotedBlock.CommitteeFromBlock().String() {
		// blockProduceTimeSlot is larger or equal than voted block
		return true
	} // if not swap committees => do nothing

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
		if proposeBlockInfo.Block == nil {
			continue
		}

		//// check if this time slot has been voted
		//if a.votedTimeslot[common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime())] {
		//	continue
		//}

		//special case: if we insert block too quick, before voting
		//=> vote for this block (within TS,but block is inserted into bestview)
		//this special case by pass validate with consensus rules
		if proposeBlockInfo.Block.GetHeight() == bestViewHeight &&
			proposeBlockInfo.Block.Hash().IsEqual(&bestViewHash) &&
			!proposeBlockInfo.IsVoted {
			tryReVoteInsertedBlock = append(tryReVoteInsertedBlock, proposeBlockInfo)
			continue
		}

		//not validate if we do it recently
		if time.Since(proposeBlockInfo.LastValidateTime).Seconds() < 1 {
			continue
		}

		// check if propose block in within TS
		if common.CalculateTimeSlot(proposeBlockInfo.Block.GetProposeTime()) != currentTimeSlot {
			continue
		}

		//if the block height is not next height or current height
		if proposeBlockInfo.Block.GetHeight() != bestViewHeight+1 {
			continue
		}

		// check if producer time > proposer time
		if common.CalculateTimeSlot(proposeBlockInfo.Block.GetProduceTime()) > currentTimeSlot {
			continue
		}

		if proposeBlockInfo.Block.GetHeight() < finalViewHeight {
			invalidProposeBlock = append(invalidProposeBlock, h)
			continue
		}

		validProposeBlock = append(validProposeBlock, proposeBlockInfo)
	}
	//rule 1: get history of vote for this height, vote if (round is lower than the vote before)
	// or (round is equal but new proposer) or (there is no vote for this height yet)
	sort.Slice(validProposeBlock, func(i, j int) bool {
		return validProposeBlock[i].Block.GetProduceTime() < validProposeBlock[j].Block.GetProduceTime()
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
		c.logger.Infof("validate block: %+v \n", proposeBlockInfo.Block.Hash().String())
		if err := c.chain.ValidatePreSignBlock(proposeBlockInfo.Block, proposeBlockInfo.SigningCommittees, proposeBlockInfo.Committees); err != nil {
			c.logger.Error(err)
			return false, err
		}
	}

	return true, nil
}

func (c ConsensusValidatorLemma1) ValidateConsensusRules(lastVotedBlock types.BlockInterface, isVoteNextHeight bool, proposeBlockInfo *ProposeBlockInfo) bool {

	if !isVoteNextHeight {
		return true
	}

	blockProduceTimeSlot := common.CalculateTimeSlot(proposeBlockInfo.Block.GetProduceTime())

	if blockProduceTimeSlot < common.CalculateTimeSlot(lastVotedBlock.GetProduceTime()) {
		// blockProduceTimeSlot is smaller than voted block => vote for this block
		return true
	} else if blockProduceTimeSlot == common.CalculateTimeSlot(lastVotedBlock.GetProduceTime()) &&
		common.CalculateTimeSlot(proposeBlockInfo.Block.GetProposeTime()) > common.CalculateTimeSlot(lastVotedBlock.GetProposeTime()) {
		// block is old block (same round), but new proposer(larger timeslot) => vote again
		return true
	} else if proposeBlockInfo.Block.CommitteeFromBlock().String() != lastVotedBlock.CommitteeFromBlock().String() {
		// blockProduceTimeSlot is larger or equal than voted block
		return true
	} // if not swap committees => do nothing

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
