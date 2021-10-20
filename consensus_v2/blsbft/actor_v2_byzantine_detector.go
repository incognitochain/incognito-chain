package blsbft

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"
	"reflect"
	"time"
)

var (
	ErrInvalidSignature           = errors.New("vote owner invalid signature")
	ErrInvalidVoteOwner           = errors.New("vote owner is not in committee list")
	ErrDuplicateVoteInOneTimeSlot = errors.New("duplicate vote in one timeslot")
	ErrVoteForHigherTimeSlot      = errors.New("vote for block with same height but higher timeslot")
)

type VoteMessageHandler func(bftVote *BFTVote) error

type IByzantineDetector interface {
	validate(vote *BFTVote) error
	updateState(finalHeight uint64, finalTimeSlot int64)
}

func getByzantineDetectorRule(detector IByzantineDetector, height uint64, handler CommitteeChainHandler, logger common.Logger) IByzantineDetector {
	if height < config.Param().ConsensusParam.ByzantineDetectorHeight {
		if reflect.TypeOf(detector) != reflect.TypeOf(new(NilByzantineDetector)) {
			detector = NewNilByzantineDetector()
		}
	} else {
		if reflect.TypeOf(detector) != reflect.TypeOf(new(ByzantineDetector)) {
			detector = NewByzantineDetector(handler, logger)
		}
	}

	return detector
}

type NilByzantineDetector struct{}

func NewNilByzantineDetector() *NilByzantineDetector {
	return &NilByzantineDetector{}
}

func (n NilByzantineDetector) validate(vote *BFTVote) error {
	return nil
}

func (n NilByzantineDetector) updateState(finalHeight uint64, finalTimeSlot int64) {
	return
}

var defaultBlackListTTL = 30 * 24 * time.Hour

func NewBlackListValidator(reason error) *rawdb_consensus.BlackListValidator {
	return &rawdb_consensus.BlackListValidator{
		Reason:    reason,
		StartTime: time.Now(),
		TTL:       defaultBlackListTTL,
	}
}

type ByzantineDetector struct {
	blackList               map[string]*rawdb_consensus.BlackListValidator // validator => reason for blacklist
	voteInTimeSlot          map[string]map[int64]*BFTVote                  // validator => timeslot => vote
	smallestProduceTimeSlot map[string]map[uint64]int64                    // validator => height => timeslot
	committeeHandler        CommitteeChainHandler
	logger                  common.Logger
}

func NewByzantineDetector(committeeHandler CommitteeChainHandler, logger common.Logger) *ByzantineDetector {

	blackListValidators, err := rawdb_consensus.GetAllBlackListValidator(rawdb_consensus.GetConsensusDatabase())
	if err != nil {
		logger.Error(err)
	}

	return &ByzantineDetector{
		committeeHandler: committeeHandler,
		logger:           logger,
		blackList:        blackListValidators,
	}
}

func (b ByzantineDetector) validate(vote *BFTVote) error {

	var err error

	handlers := []VoteMessageHandler{
		b.voteOwnerSignature,
		b.voteMoreThanOneTimesInATimeSlot,
		b.voteForHigherTimeSlotSameHeight,
	}

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

	for validator, blacklist := range b.blackList {
		if time.Now().Unix() > blacklist.StartTime.Add(blacklist.TTL).Unix() {
			err := rawdb_consensus.DeleteBlackListValidator(
				rawdb_consensus.GetConsensusDatabase(),
				validator,
			)
			if err != nil {
				b.logger.Error("Fail to delete long life-time black list validator", err)
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
		b.blackList = make(map[string]*rawdb_consensus.BlackListValidator)
	}
	if validatorErr != nil {
		blackListValidator := NewBlackListValidator(validatorErr)
		b.blackList[bftVote.Validator] = blackListValidator
		err := rawdb_consensus.StoreBlackListValidator(
			rawdb_consensus.GetConsensusDatabase(),
			bftVote.Validator,
			blackListValidator,
		)
		if err != nil {
			b.logger.Error("Store Black List Validator Error", err)
		}
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