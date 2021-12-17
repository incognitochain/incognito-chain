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
	ErrVoteForSmallerBlockHeight  = errors.New("vote for block smaller block height but voted for higher block height")
)

var ByzantineDetectorObject *ByzantineDetector

var defaultBlackListTTL = 1 * time.Second
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
	validRecentVote              map[string]*BFTVote
	smallestBlockProduceTimeSlot map[string]map[uint64]*BFTVote // validator => height => timeslot
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
	defaultBlackListTTL = time.Duration(config.Param().EpochParam.NumberOfBlockInEpoch) * time.Duration(common.TIMESLOT) * time.Second
	blackListValidators, err := rawdb_consensus.GetAllBlackListValidator(rawdb_consensus.GetConsensusDatabase())
	if err != nil {
		logger.Error(err)
	}

	return &ByzantineDetector{
		logger:                       logger,
		blackList:                    blackListValidators,
		voteInTimeSlot:               make(map[string]map[int64]*BFTVote),
		smallestBlockProduceTimeSlot: make(map[string]map[uint64]*BFTVote),
		validRecentVote:              make(map[string]*BFTVote),
		mu:                           new(sync.RWMutex),
	}
}

func (b ByzantineDetector) GetByzantineDetectorInfo() map[string]interface{} {

	b.mu.RLock()
	defer b.mu.RUnlock()

	blackList := make(map[string]*rawdb_consensus.BlackListValidator)
	voteInTimeSlot := make(map[string]map[int64]*BFTVote)
	validRecentVote := make(map[string]*BFTVote)
	smallestBlockProduceTimeSlot := make(map[string]map[uint64]*BFTVote)

	for k, v := range b.blackList {
		blackList[k] = v
	}
	for k, v := range b.voteInTimeSlot {
		voteInTimeSlot[k] = v
	}
	for k, v := range b.smallestBlockProduceTimeSlot {
		smallestBlockProduceTimeSlot[k] = v
	}
	for k, v := range b.validRecentVote {
		validRecentVote[k] = v
	}

	m := map[string]interface{}{
		"BlackList":                 blackList,
		"VoteInTimeSlot":            voteInTimeSlot,
		"BlockWithSmallestTimeSlot": smallestBlockProduceTimeSlot,
		"ValidRecentVote":           validRecentVote,
	}

	return m
}

func (b *ByzantineDetector) Validate(bestViewHeight uint64, vote *BFTVote) error {

	if vote.isEmptyDataForByzantineDetector() {
		b.logger.Debug("Empty data to validate Byzantine vote")
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	var err error

	handlers := []VoteMessageHandler{
		b.voteMoreThanOneTimesInATimeSlot,
		b.voteForHigherTimeSlotSameHeight,
		b.voteForSmallerBlockHeight,
	}

	if config.Param().ConsensusParam.ByzantineDetectorHeight < bestViewHeight {
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

	if config.Param().ConsensusParam.ByzantineDetectorHeight < bestViewHeight {
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
		b.intervalRemoveBlackListValidator()
		b.mu.Unlock()
	}
}

func (b *ByzantineDetector) intervalRemoveBlackListValidator() {
	for validator, blacklist := range b.blackList {
		if time.Now().Unix() > blacklist.StartTime.Add(blacklist.TTL).Unix() {
			b.removeBlackListValidator(validator)
		}
	}
}

func (b *ByzantineDetector) RemoveBlackListValidator(validator string) error {

	b.mu.Lock()
	err := b.removeBlackListValidator(validator)
	b.mu.Unlock()

	return err
}

func (b *ByzantineDetector) removeBlackListValidator(validator string) error {
	err := rawdb_consensus.DeleteBlackListValidator(
		rawdb_consensus.GetConsensusDatabase(),
		validator,
	)

	if err != nil {
		b.logger.Error("Fail to delete black list validator", err)
		return err
	}
	delete(b.blackList, validator)

	return nil
}

func (b ByzantineDetector) checkFixedNodes(validator string) bool {
	return b.fixedNodes[validator]
}

func (b ByzantineDetector) checkBlackListValidator(bftVote *BFTVote) error {

	if b.checkFixedNodes(bftVote.Validator) {
		b.logger.Debugf("Validator %+v in byzantine detector black list but allow vote because of it's fixed node", bftVote.Validator)
		return nil
	}

	if err, ok := b.blackList[bftVote.Validator]; ok {
		return fmt.Errorf("validator in black list %+v, due to %+v", bftVote.Validator, err)
	}

	return nil
}

func (b ByzantineDetector) voteMoreThanOneTimesInATimeSlot(newVote *BFTVote) error {

	voteInTimeSlot, ok := b.voteInTimeSlot[newVote.Validator]
	if !ok {
		return nil
	}

	if vote, ok := voteInTimeSlot[newVote.ProposeTimeSlot]; ok {
		// allow receiving same vote multiple times
		if !vote.CommitteeFromBlock.IsEqual(&newVote.CommitteeFromBlock) {
			return nil
		}
		if !reflect.DeepEqual(vote, newVote) {
			return fmt.Errorf("error name: %+v,"+
				"first bftvote: %+v, latter bftvote: %+v",
				ErrVoteForHigherTimeSlot, vote, newVote)
		}
	}

	return nil
}

func (b ByzantineDetector) voteForHigherTimeSlotSameHeight(newVote *BFTVote) error {

	smallestTimeSlotBlock, ok := b.smallestBlockProduceTimeSlot[newVote.Validator]
	if !ok {
		return nil
	}

	smallestTimeSlotVote, ok := smallestTimeSlotBlock[newVote.BlockHeight]
	if !ok {
		return nil
	}

	if newVote.CommitteeFromBlock != smallestTimeSlotVote.CommitteeFromBlock {
		return nil
	}

	if newVote.ProduceTimeSlot > smallestTimeSlotVote.ProduceTimeSlot {
		return fmt.Errorf("error name: %+v,"+
			"block height: %+v, bigger vote: %+v, smallest vote: %+v",
			ErrVoteForHigherTimeSlot, newVote.BlockHeight, newVote, smallestTimeSlotBlock)
	}

	return nil
}

func (b ByzantineDetector) voteForSmallerBlockHeight(newVote *BFTVote) error {

	recentVote, ok := b.validRecentVote[newVote.Validator]

	if !ok {
		return nil
	}

	if newVote.ChainID != recentVote.ChainID {
		return nil
	}

	if newVote.CommitteeFromBlock != recentVote.CommitteeFromBlock {
		return nil
	}

	if recentVote.BlockHeight > newVote.BlockHeight {
		return fmt.Errorf("error name: %+v,"+
			"recent vote %+v, new vote %+v",
			ErrVoteForSmallerBlockHeight, recentVote, newVote)
	}

	return nil
}

func (b *ByzantineDetector) addNewVote(database incdb.Database, newVote *BFTVote, validatorErr error) {

	if b.blackList == nil {
		b.blackList = make(map[string]*rawdb_consensus.BlackListValidator)
	}
	if validatorErr != nil {
		blackListValidator := NewBlackListValidator(validatorErr)
		b.blackList[newVote.Validator] = blackListValidator
		err := rawdb_consensus.StoreBlackListValidator(
			database,
			newVote.Validator,
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
	_, ok := b.voteInTimeSlot[newVote.Validator]
	if !ok {
		b.voteInTimeSlot[newVote.Validator] = make(map[int64]*BFTVote)
	}
	b.voteInTimeSlot[newVote.Validator][newVote.ProposeTimeSlot] = newVote

	if b.smallestBlockProduceTimeSlot == nil {
		b.smallestBlockProduceTimeSlot = make(map[string]map[uint64]*BFTVote)
	}
	_, ok2 := b.smallestBlockProduceTimeSlot[newVote.Validator]
	if !ok2 {
		b.smallestBlockProduceTimeSlot[newVote.Validator] = make(map[uint64]*BFTVote)
	}
	if oldVote, ok := b.smallestBlockProduceTimeSlot[newVote.Validator][newVote.BlockHeight]; !ok ||
		(ok && newVote.ProduceTimeSlot < oldVote.ProduceTimeSlot) ||
		(ok && newVote.CommitteeFromBlock != oldVote.CommitteeFromBlock) {
		b.smallestBlockProduceTimeSlot[newVote.Validator][newVote.BlockHeight] = newVote
	}

	b.validRecentVote[newVote.Validator] = newVote
}
