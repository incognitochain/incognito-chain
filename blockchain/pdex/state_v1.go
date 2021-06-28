package pdex

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type stateV1 struct {
	waitingContributions        map[string]*rawdbv2.PDEContribution
	deletedWaitingContributions map[string]*rawdbv2.PDEContribution
	poolPairs                   map[string]*rawdbv2.PDEPoolForPair
	shares                      map[string]uint64
	tradingFees                 map[string]uint64
	producer                    stateProducerV1
	processor                   stateProcessorV1
}

func newStateV1() *stateV1 {
	res := &stateV1{
		waitingContributions:        make(map[string]*rawdbv2.PDEContribution),
		deletedWaitingContributions: make(map[string]*rawdbv2.PDEContribution),
		poolPairs:                   make(map[string]*rawdbv2.PDEPoolForPair),
		shares:                      make(map[string]uint64),
		tradingFees:                 make(map[string]uint64),
	}
	return res
}

func newStateV1WithValue(
	waitingContributions map[string]*rawdbv2.PDEContribution,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
	tradingFees map[string]uint64,
) *stateV1 {
	return &stateV1{
		waitingContributions:        waitingContributions,
		deletedWaitingContributions: make(map[string]*rawdbv2.PDEContribution),
		poolPairs:                   poolPairs,
		shares:                      shares,
		tradingFees:                 tradingFees,
	}
}

func initStateV1(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*stateV1, error) {
	waitingContributions, err := statedb.GetWaitingPDEContributions(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	poolPairs, err := statedb.GetPDEPoolPair(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	shares, err := statedb.GetPDEShares(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	tradingFees, err := statedb.GetPDETradingFees(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	return newStateV1WithValue(
		waitingContributions,
		poolPairs,
		shares,
		tradingFees,
	), nil
}

func (s *stateV1) Version() uint {
	return BasicVersion
}

func (s *stateV1) Clone() State {
	res := newStateV1()
	res.waitingContributions = make(map[string]*rawdbv2.PDEContribution, len(s.waitingContributions))
	res.deletedWaitingContributions = make(map[string]*rawdbv2.PDEContribution, len(s.deletedWaitingContributions))
	res.poolPairs = make(map[string]*rawdbv2.PDEPoolForPair, len(s.poolPairs))
	res.shares = make(map[string]uint64, len(s.shares))
	res.tradingFees = make(map[string]uint64, len(s.tradingFees))

	for k, v := range s.waitingContributions {
		res.waitingContributions[k] = new(rawdbv2.PDEContribution)
		*res.waitingContributions[k] = *v
	}

	for k, v := range s.deletedWaitingContributions {
		res.deletedWaitingContributions[k] = new(rawdbv2.PDEContribution)
		*res.deletedWaitingContributions[k] = *v
	}

	for k, v := range s.poolPairs {
		res.poolPairs[k] = new(rawdbv2.PDEPoolForPair)
		*res.poolPairs[k] = *v
	}

	for k, v := range s.shares {
		res.shares[k] = v
	}

	for k, v := range s.tradingFees {
		res.tradingFees[k] = v
	}
	res.producer = s.producer
	res.processor = s.processor

	return res
}

func (s *stateV1) Process(env StateEnvironment) error {
	for _, inst := range env.BeaconInstructions() {
		if len(inst) < 2 {
			continue // Not error, just not PDE instructions
		}
		metadataType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue // Not error, just not PDE instructions
		}
		if !metadata.IsPDEType(metadataType) {
			continue // Not error, just not PDE instructions
		}
		switch metadataType {
		case metadata.PDEContributionMeta:
			if env.BeaconHeight() >= env.BCHeightBreakPointPrivacyV2() {
				continue
			}
			err = s.processor.processContribution(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.waitingContributions,
				s.deletedWaitingContributions,
				s.poolPairs,
				s.shares,
			)
		case metadata.PDEPRVRequiredContributionRequestMeta:
			err = s.processor.processContribution(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.waitingContributions,
				s.deletedWaitingContributions,
				s.poolPairs,
				s.shares,
			)
		case metadata.PDETradeRequestMeta:
			if env.BeaconHeight() >= env.BCHeightBreakPointPrivacyV2() {
				continue
			}
			err = s.processor.processTrade(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.poolPairs,
			)
		case metadata.PDECrossPoolTradeRequestMeta:
			err = s.processor.processCrossPoolTrade(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.poolPairs,
			)
		case metadata.PDEWithdrawalRequestMeta:
			err = s.processor.processWithdrawal(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.poolPairs,
				s.shares,
			)
		case metadata.PDEFeeWithdrawalRequestMeta:
			err = s.processor.processFeeWithdrawal(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.tradingFees,
			)
		case metadata.PDETradingFeesDistributionMeta:
			err = s.processor.processTradingFeesDistribution(
				env.StateDB(),
				env.BeaconHeight(),
				inst,
				s.tradingFees,
			)
		default:
			Logger.log.Debug("Can not process this metadata")
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *stateV1) BuildInstructions(env StateEnvironment) ([][]string, error) {
	instructions := [][]string{}

	// handle fee withdrawal
	feeWithdrawalInstructions, err := s.producer.buildInstructionsForFeeWithdrawal(
		env.FeeWithdrawalActions(),
		env.BeaconHeight(),
		s.tradingFees,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, feeWithdrawalInstructions...)

	if env.BeaconHeight() < env.BCHeightBreakPointPrivacyV2() {
		// handle trade
		tradeInstructions, err := s.producer.buildInstructionsForTrade(
			env.TradeActions(),
			env.BeaconHeight(),
			s.poolPairs,
		)
		if err != nil {
			return instructions, err
		}
		instructions = append(instructions, tradeInstructions...)
	}

	// handle cross pool trade
	crossPoolTradeInstructions, err := s.producer.buildInstructionsForCrossPoolTrade(
		env.CrossPoolTradeActions(),
		env.BeaconHeight(),
		s.poolPairs,
		s.shares,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, crossPoolTradeInstructions...)

	// handle withdrawal
	withdrawalInstructions, err := s.producer.buildInstructionsForWithdrawal(
		env.WithdrawalActions(),
		env.BeaconHeight(),
		s.poolPairs,
		s.shares,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawalInstructions...)

	if env.BeaconHeight() < env.BCHeightBreakPointPrivacyV2() {
		// handle contribution
		contributionInstructions, err := s.producer.buildInstructionsForContribution(
			env.ContributionActions(),
			env.BeaconHeight(),
			false,
			metadata.PDEContributionMeta,
			s.waitingContributions,
			s.poolPairs,
			s.shares,
		)
		if err != nil {
			return instructions, err
		}
		instructions = append(instructions, contributionInstructions...)
	}

	// handle prv required contribution
	contributionInstructions, err := s.producer.buildInstructionsForContribution(
		env.PRVRequiredContributionActions(),
		env.BeaconHeight(),
		true,
		metadata.PDEPRVRequiredContributionRequestMeta,
		s.waitingContributions,
		s.poolPairs,
		s.shares,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, contributionInstructions...)

	return instructions, nil
}

func (s *stateV1) Upgrade(env StateEnvironment) State {
	var state State
	return state
}

func (s *stateV1) StoreToDB(env StateEnvironment) error {
	var err error
	statedb.DeleteWaitingPDEContributions(
		env.StateDB(),
		s.deletedWaitingContributions,
	)
	err = statedb.StoreWaitingPDEContributions(
		env.StateDB(),
		s.waitingContributions,
	)
	if err != nil {
		return err
	}
	err = statedb.StorePDEPoolPairs(
		env.StateDB(),
		s.poolPairs,
	)
	if err != nil {
		return err
	}
	err = statedb.StorePDEShares(
		env.StateDB(),
		s.shares,
	)
	if err != nil {
		return err
	}
	err = statedb.StorePDETradingFees(
		env.StateDB(),
		s.tradingFees,
	)
	if err != nil {
		return err
	}
	return err
}

func (s *stateV1) TransformKeyWithNewBeaconHeight(beaconHeight uint64) {
	time1 := time.Now()
	sameHeight := false
	//transform pdex key prefix-<beaconheight>-id1-id2 (if same height, no transform)
	transformKey := func(key string, beaconHeight uint64) string {
		if sameHeight {
			return key
		}
		keySplit := strings.Split(key, "-")
		if keySplit[1] == strconv.Itoa(int(beaconHeight)) {
			sameHeight = true
		}
		keySplit[1] = strconv.Itoa(int(beaconHeight))
		return strings.Join(keySplit, "-")
	}

	newState := newStateV1()

	for k, v := range s.waitingContributions {
		newState.waitingContributions[transformKey(k, beaconHeight)] = v
		if sameHeight {
			s = newState
			return
		}
	}
	for k, v := range s.deletedWaitingContributions {
		newState.deletedWaitingContributions[transformKey(k, beaconHeight)] = v
	}
	for k, v := range s.poolPairs {
		newState.poolPairs[transformKey(k, beaconHeight)] = v
	}
	for k, v := range s.shares {
		newState.shares[transformKey(k, beaconHeight)] = v
	}
	for k, v := range s.tradingFees {
		newState.tradingFees[transformKey(k, beaconHeight)] = v
	}
	Logger.log.Infof("Time spent for transforming keys: %f", time.Since(time1).Seconds())
	*s = *newState
}

func (s *stateV1) ClearCache() {
	s.deletedWaitingContributions = make(map[string]*rawdbv2.PDEContribution)
}

func (s *stateV1) GetDiff(compareState State) (State, error) {
	if compareState == nil {
		return nil, errors.New("compareState is nil")
	}

	res := newStateV1()
	compareStateV1 := compareState.(*stateV1)

	for k, v := range s.waitingContributions {
		if m, ok := compareStateV1.waitingContributions[k]; !ok || !reflect.DeepEqual(m, v) {
			res.waitingContributions[k] = new(rawdbv2.PDEContribution)
			*res.waitingContributions[k] = *v
		}
	}
	for k, v := range s.deletedWaitingContributions {
		if m, ok := compareStateV1.deletedWaitingContributions[k]; !ok || !reflect.DeepEqual(m, v) {
			res.deletedWaitingContributions[k] = new(rawdbv2.PDEContribution)
			*res.deletedWaitingContributions[k] = *v
		}
	}
	for k, v := range s.poolPairs {
		if m, ok := compareStateV1.poolPairs[k]; !ok || !reflect.DeepEqual(m, v) {
			res.poolPairs[k] = new(rawdbv2.PDEPoolForPair)
			*res.poolPairs[k] = *v
		}
	}
	for k, v := range s.shares {
		if m, ok := compareStateV1.shares[k]; !ok || !reflect.DeepEqual(m, v) {
			res.shares[k] = v
		}
	}
	for k, v := range s.tradingFees {
		if m, ok := compareStateV1.tradingFees[k]; !ok || !reflect.DeepEqual(m, v) {
			res.tradingFees[k] = v
		}
	}
	return res, nil
}

func (s *stateV1) WaitingContributions() map[string]*rawdbv2.PDEContribution {
	res := make(map[string]*rawdbv2.PDEContribution, len(s.waitingContributions))
	for k, v := range s.waitingContributions {
		res[k] = new(rawdbv2.PDEContribution)
		*res[k] = *v
	}
	return res
}

func (s *stateV1) DeletedWaitingContributions() map[string]*rawdbv2.PDEContribution {
	res := make(map[string]*rawdbv2.PDEContribution, len(s.deletedWaitingContributions))
	for k, v := range s.deletedWaitingContributions {
		res[k] = new(rawdbv2.PDEContribution)
		*res[k] = *v
	}
	return res
}

func (s *stateV1) PoolPairs() map[string]*rawdbv2.PDEPoolForPair {
	res := make(map[string]*rawdbv2.PDEPoolForPair, len(s.poolPairs))
	for k, v := range s.poolPairs {
		res[k] = new(rawdbv2.PDEPoolForPair)
		*res[k] = *v
	}
	return res
}

func (s *stateV1) Shares() map[string]uint64 {
	res := make(map[string]uint64, len(s.shares))
	for k, v := range s.shares {
		res[k] = v
	}
	return res
}

func (s *stateV1) TradingFees() map[string]uint64 {
	res := make(map[string]uint64, len(s.tradingFees))
	for k, v := range s.tradingFees {
		res[k] = v
	}
	return res
}
