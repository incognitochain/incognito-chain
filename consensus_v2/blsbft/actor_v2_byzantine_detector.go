package blsbft

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"sync"
	"time"
)

var (
	ErrDuplicateVoteInOneTimeSlot = errors.New("duplicate vote in one timeslot")
	ErrVoteForHigherTimeSlot      = errors.New("vote for block with same height but higher timeslot")
)

var ByzantineDetectorObject *ByzantineDetector

var defaultBlackListTTL = 30 * 24 * time.Hour
var defaultHeightTTL = uint64(100)
var defaultTimeSlotTTL = int64(100)

type VoteMessageHandler func(bftVote *BFTVote) error

func NewBlackListValidator(reason error) *rawdb_consensus.BlackListValidator {
	return &rawdb_consensus.BlackListValidator{
		Error:     reason.Error(),
		StartTime: time.Now(),
		TTL:       defaultBlackListTTL,
	}
}

type ByzantineDetector struct {
	fixedNodes                   map[string]bool                                // fixed nodes
	blackList                    map[string]*rawdb_consensus.BlackListValidator // validator => reason for blacklist
	voteInTimeSlot               map[string]map[int64]*BFTVote                  // validator => timeslot => vote
	smallestBlockProduceTimeSlot map[string]map[uint64]int64                    // validator => height => timeslot
	logger                       common.Logger
	mu                           *sync.RWMutex
}

func (b *ByzantineDetector) SetFixedNodes(fixedNodes []incognitokey.CommitteePublicKey) {
	blsKeys := make(map[string]bool)
	for _, k := range fixedNodes {
		blsKey := k.GetMiningKeyBase58(common.BlsConsensus)
		blsKeys[blsKey] = true
	}

	b.fixedNodes = blsKeys
}

func NewByzantineDetector(logger common.Logger) *ByzantineDetector {

	blackListValidators, err := rawdb_consensus.GetAllBlackListValidator(rawdb_consensus.GetConsensusDatabase())
	if err != nil {
		logger.Error(err)
	}

	return &ByzantineDetector{
		logger:                       logger,
		blackList:                    blackListValidators,
		voteInTimeSlot:               make(map[string]map[int64]*BFTVote),
		smallestBlockProduceTimeSlot: make(map[string]map[uint64]int64),
		mu:                           new(sync.RWMutex),
	}
}

func (b ByzantineDetector) GetByzantineDetectorInfo() map[string]interface{} {

	b.mu.RLock()
	defer b.mu.RUnlock()

	blackList := make(map[string]*rawdb_consensus.BlackListValidator)
	voteInTimeSlot := make(map[string]map[int64]*BFTVote)
	smallestBlockProduceTimeSlot := make(map[string]map[uint64]int64)
	for k, v := range b.blackList {
		blackList[k] = v
	}
	for k, v := range b.voteInTimeSlot {
		voteInTimeSlot[k] = v
	}
	for k, v := range b.smallestBlockProduceTimeSlot {
		smallestBlockProduceTimeSlot[k] = v
	}
	m := map[string]interface{}{
		"BlackList":                 blackList,
		"VoteInTimeSlot":            voteInTimeSlot,
		"BlockWithSmallestTimeSlot": smallestBlockProduceTimeSlot,
	}

	return m
}

func (b *ByzantineDetector) Validate(bestViewHeight uint64, vote *BFTVote) error {

	b.mu.Lock()
	defer b.mu.Unlock()

	var err error

	handlers := []VoteMessageHandler{
		b.voteMoreThanOneTimesInATimeSlot,
		b.voteForHigherTimeSlotSameHeight,
	}

	if config.Param().ConsensusParam.ByzantineDetectorHeight >= bestViewHeight {
		if err := b.checkBlackListValidator(vote); err != nil {
			return err
		}
	}

	for _, handler := range handlers {
		err = handler(vote)
		if err != nil {
			b.logger.Error("Byzantine Detector Error", err)
			break
		}
	}

	b.addNewVote(rawdb_consensus.GetConsensusDatabase(), vote, err)

	if config.Param().ConsensusParam.ByzantineDetectorHeight >= bestViewHeight {
		return err
	}

	return nil
}

func (b *ByzantineDetector) UpdateState(finalHeight uint64, finalTimeSlot int64) {

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, voteInTimeSlot := range b.voteInTimeSlot {
		for timeSlot, _ := range voteInTimeSlot {
			if timeSlot+defaultTimeSlotTTL < finalTimeSlot {
				delete(voteInTimeSlot, timeSlot)
			}
		}
	}

	for _, smallestTimeSlot := range b.smallestBlockProduceTimeSlot {
		for height, _ := range smallestTimeSlot {
			if height+defaultHeightTTL < finalHeight {
				delete(smallestTimeSlot, height)
			}
		}
	}
}

func (b *ByzantineDetector) Loop() {

	ticker := time.Tick(1 * time.Hour)

	for _ = range ticker {
		b.mu.Lock()
		b.removeBlackListValidator()
		b.mu.Unlock()
	}
}

func (b *ByzantineDetector) removeBlackListValidator() {
	for validator, blacklist := range b.blackList {
		if time.Now().Unix() > blacklist.StartTime.Add(blacklist.TTL).Unix() {
			err := rawdb_consensus.DeleteBlackListValidator(
				rawdb_consensus.GetConsensusDatabase(),
				validator,
			)
			if err != nil {
				b.logger.Error("Fail to delete long life-time black list validator", err)
			}

			delete(b.blackList, validator)
		}
	}
}

func (b ByzantineDetector) checkFixedNodes(validator string) bool {
	return b.fixedNodes[validator]
}

func (b ByzantineDetector) checkBlackListValidator(bftVote *BFTVote) error {

	if b.checkFixedNodes(bftVote.Validator) {
		return nil
	}

	if err, ok := b.blackList[bftVote.Validator]; ok {
		return fmt.Errorf("validator in black list %+v, due to %+v", bftVote.Validator, err)
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
			return fmt.Errorf("error name: %+v, "+
				"first bftvote %+v, latter bftvote %+v",
				ErrVoteForHigherTimeSlot, vote, bftVote)
		}
	}

	return nil
}

func (b ByzantineDetector) voteForHigherTimeSlotSameHeight(bftVote *BFTVote) error {

	smallestTimeSlotBlock, ok := b.smallestBlockProduceTimeSlot[bftVote.Validator]
	if !ok {
		return nil
	}

	blockTimeSlot, ok := smallestTimeSlotBlock[bftVote.BlockHeight]
	if !ok {
		return nil
	}

	if bftVote.ProduceTimeSlot > blockTimeSlot {
		return fmt.Errorf("error name: %+v, "+
			"block height %+v, bigger vote block produce timeslot %+v, smallest vote block produce timeSlot %+v",
			ErrVoteForHigherTimeSlot, bftVote.BlockHeight, bftVote.ProduceTimeSlot, blockTimeSlot)
	}

	return nil
}

func (b *ByzantineDetector) addNewVote(database incdb.Database, bftVote *BFTVote, validatorErr error) {

	if b.blackList == nil {
		b.blackList = make(map[string]*rawdb_consensus.BlackListValidator)
	}
	if validatorErr != nil {
		blackListValidator := NewBlackListValidator(validatorErr)
		b.blackList[bftVote.Validator] = blackListValidator
		err := rawdb_consensus.StoreBlackListValidator(
			database,
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

	if b.smallestBlockProduceTimeSlot == nil {
		b.smallestBlockProduceTimeSlot = make(map[string]map[uint64]int64)
	}
	_, ok2 := b.smallestBlockProduceTimeSlot[bftVote.Validator]
	if !ok2 {
		b.smallestBlockProduceTimeSlot[bftVote.Validator] = make(map[uint64]int64)
	}
	if res, ok := b.smallestBlockProduceTimeSlot[bftVote.Validator][bftVote.BlockHeight]; !ok || (ok && bftVote.ProduceTimeSlot < res) {
		b.smallestBlockProduceTimeSlot[bftVote.Validator][bftVote.BlockHeight] = bftVote.ProduceTimeSlot
	}
}
