package blsbft

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

var (
	ErrInvalidSignature           = errors.New("vote owner invalid signature")
	ErrInvalidVoteOwner           = errors.New("vote owner is not in committee list")
	ErrDuplicateVoteInOneTimeSlot = errors.New("duplicate vote in one timeslot")
	ErrVoteForHigherTimeSlot      = errors.New("vote for block with same height but higher timeslot")
)

type VoteMessageHandler func(bftVote *BFTVote) error

type IByzantineDetector interface {
	Validate(vote *BFTVote, handler ...VoteMessageHandler) error
}

type ByzantineDetector struct {
	blackList               map[string]error              // validator => reason for blacklist
	voteInTimeSlot          map[string]map[int64]*BFTVote // validator => timeslot => vote
	smallestProduceTimeSlot map[string]map[uint64]int64   // validator => height => timeslot
	committeeHandler        CommitteeChainHandler
}

func NewByzantineDetector(committeeHandler CommitteeChainHandler) *ByzantineDetector {
	return &ByzantineDetector{committeeHandler: committeeHandler}
}

func (b ByzantineDetector) Validate(vote *BFTVote, handlers ...VoteMessageHandler) error {

	var err error

	if err := b.checkBlackListValidator(vote); err != nil {
		return err
	}

	for _, handler := range handlers {
		err = handler(vote)
		if err != nil {
			break
		}
	}

	b.addNewVote(vote, err)

	return err
}

func (b *ByzantineDetector) updateState(finalHeight uint64, finalTimeSlot int64) {

	for _, voteInTimeSlot := range b.voteInTimeSlot {
		for timeSlot, _ := range voteInTimeSlot {
			if timeSlot < finalTimeSlot {
				delete(voteInTimeSlot, timeSlot)
			}
		}
	}

	for _, smallestTimeSlot := range b.smallestProduceTimeSlot {
		for height, _ := range smallestTimeSlot {
			if height < finalHeight {
				delete(smallestTimeSlot, height)
			}
		}
	}

}

func (b ByzantineDetector) checkBlackListValidator(bftVote *BFTVote) error {

	if err, ok := b.blackList[bftVote.Validator]; ok {
		return fmt.Errorf("validator in black list %+v, due to %+v", bftVote.Validator, err)
	}

	return nil
}

func (b ByzantineDetector) voteOwnerSignature(bftVote *BFTVote) error {

	committees, err := b.committeeHandler.CommitteesFromViewHashForShard(bftVote.CommitteeFromBlock, byte(bftVote.ChainID))
	if err != nil {
		return err
	}

	found := false
	for _, v := range committees {
		if v.GetMiningKeyBase58(common.BlsConsensus) == bftVote.Validator {
			found = true
			err := bftVote.validateVoteOwner(v.MiningPubKey[common.BridgeConsensus])
			if err != nil {
				return fmt.Errorf("%+v, %+v", ErrInvalidSignature, err)
			}
		}
	}

	if !found {
		return ErrInvalidVoteOwner
	}

	return nil
}

func (b ByzantineDetector) voteMoreThanOneTimesInATimeSlot(bftVote *BFTVote) error {

	voteInTimeSlot, ok := b.voteInTimeSlot[bftVote.Validator]
	if !ok {
		return nil
	}

	if vote, ok := voteInTimeSlot[bftVote.ProposeTimeSlot]; ok {
		// allow receiving same vote multiple times
		if !reflect.DeepEqual(vote, bftVote) {
			return ErrDuplicateVoteInOneTimeSlot
		}
	}

	return nil
}

func (b ByzantineDetector) voteForHigherTimeSlotSameHeight(bftVote *BFTVote) error {

	smallestTimeSlotBlock, ok := b.smallestProduceTimeSlot[bftVote.Validator]
	if !ok {
		return nil
	}

	blockTimeSlot, ok := smallestTimeSlotBlock[bftVote.BlockHeight]
	if !ok {
		return nil
	}

	if bftVote.ProduceTimeSlot > blockTimeSlot {
		return ErrVoteForHigherTimeSlot
	}

	return nil
}

func (b *ByzantineDetector) addNewVote(bftVote *BFTVote, validatorErr error) {

	if b.blackList == nil {
		b.blackList = make(map[string]error)
	}
	if validatorErr != nil {
		b.blackList[bftVote.Validator] = validatorErr
		return
	}

	if b.voteInTimeSlot == nil {
		b.voteInTimeSlot = make(map[string]map[int64]*BFTVote)
	}
	_, ok := b.voteInTimeSlot[bftVote.Validator]
	if !ok {
		b.voteInTimeSlot[bftVote.Validator] = make(map[int64]*BFTVote)
	}
	b.voteInTimeSlot[bftVote.Validator][bftVote.ProposeTimeSlot] = bftVote

	if b.smallestProduceTimeSlot == nil {
		b.smallestProduceTimeSlot = make(map[string]map[uint64]int64)
	}
	_, ok2 := b.smallestProduceTimeSlot[bftVote.Validator]
	if !ok2 {
		b.smallestProduceTimeSlot[bftVote.Validator] = make(map[uint64]int64)
	}
	if res, ok := b.smallestProduceTimeSlot[bftVote.Validator][bftVote.BlockHeight]; !ok || (ok && bftVote.ProduceTimeSlot < res) {
		b.smallestProduceTimeSlot[bftVote.Validator][bftVote.BlockHeight] = bftVote.ProduceTimeSlot
	}
}
