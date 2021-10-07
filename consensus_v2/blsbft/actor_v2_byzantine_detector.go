package blsbft

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

var (
	ErrInvalidSignature           = errors.New("vote owner invalid signature")
	ErrInvalidVoteOwner           = errors.New("vote owner is not in committee list")
	ErrDuplicateVoteInOneTimeSlot = errors.New("duplicate vote in one timeslot")
	ErrVoteForHigherTimeSlot      = errors.New("vote for block with same height but higher timeslot")
	ErrVoteForSmallerHeight       = errors.New("vote for block with smaller height")
)

type VoteMessageHandler func(bftVote *BFTVote) error

type ByzantineDetector struct {
	blackList         map[string]error              // validator => reason for blacklist
	voteInTimeSlot    map[string]map[int64]*BFTVote // validator => timeslot => vote
	smallestTimeSlot  map[string]map[uint64]int64   // validator => height => timeslot
	latestBlockHeight map[string]uint64
	committeeHandler  CommitteeChainHandler
}

func (b ByzantineDetector) validate(vote *BFTVote, handlers ...VoteMessageHandler) error {

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

	b.addNewVote(vote)

	return err
}

func (b *ByzantineDetector) cleanMem(finalHeight uint64, finalTimeSlot int64) {

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

	if _, ok := voteInTimeSlot[bftVote.TimeSlot]; ok {
		return ErrDuplicateVoteInOneTimeSlot
	}

	return nil
}

func (b ByzantineDetector) voteForSmallerTimeSlotSameHeight(bftVote *BFTVote) error {

	smallestTimeSlotBlock, ok := b.smallestTimeSlot[bftVote.Validator]
	if !ok {
		return nil
	}

	blockTimeSlot, ok := smallestTimeSlotBlock[bftVote.BlockHeight]
	if !ok {
		return nil
	}

	if blockTimeSlot < bftVote.TimeSlot {
		return ErrVoteForHigherTimeSlot
	}

	return nil
}

func (b ByzantineDetector) voteForSmallerBlockHeight(bftVote *BFTVote) error {

	latestHeight, ok := b.latestBlockHeight[bftVote.Validator]
	if !ok {
		return nil
	}

	if latestHeight < bftVote.BlockHeight {
		return ErrVoteForSmallerHeight
	}

	return nil
}

func (b *ByzantineDetector) addNewVote(bftVote *BFTVote) {

}
