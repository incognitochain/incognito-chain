package pdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type stateV1 struct {
	stateBase
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
			s.waitingContributions,
				s.deletedWaitingContributions,
				s.poolPairs,
				s.shares,
				err = s.processor.contribution(
				env.StateDB(),
				env.PrevBeaconHeight(),
				inst,
				s.waitingContributions,
				s.deletedWaitingContributions,
				s.poolPairs,
				s.shares,
			)
		case metadata.PDEPRVRequiredContributionRequestMeta:
			s.waitingContributions,
				s.deletedWaitingContributions,
				s.poolPairs,
				s.shares,
				err = s.processor.contribution(
				env.StateDB(),
				env.PrevBeaconHeight(),
				inst,
				s.waitingContributions,
				s.deletedWaitingContributions,
				s.poolPairs,
				s.shares,
			)
		case metadata.PDETradeRequestMeta:
			s.poolPairs, err = s.processor.trade(
				env.StateDB(),
				env.PrevBeaconHeight(),
				inst,
				s.poolPairs,
			)
		case metadata.PDECrossPoolTradeRequestMeta:
			s.poolPairs, err = s.processor.crossPoolTrade(
				env.StateDB(),
				env.PrevBeaconHeight(),
				inst,
				s.poolPairs,
			)
		case metadata.PDEWithdrawalRequestMeta:
			s.poolPairs, s.shares, err = s.processor.withdrawal(
				env.StateDB(),
				env.PrevBeaconHeight(),
				inst,
				s.poolPairs,
				s.shares,
			)
		case metadata.PDEFeeWithdrawalRequestMeta:
			s.tradingFees, err = s.processor.feeWithdrawal(
				env.StateDB(),
				env.PrevBeaconHeight(),
				inst,
				s.tradingFees,
			)
		case metadata.PDETradingFeesDistributionMeta:
			s.tradingFees, err = s.processor.tradingFeesDistribution(
				env.StateDB(),
				env.PrevBeaconHeight(),
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
	var err error
	feeWithdrawalInstructions := [][]string{}
	tradeInstructions := [][]string{}
	crossPoolTradeInstructions := [][]string{}
	withdrawalInstructions := [][]string{}
	contributionInstructions := [][]string{}

	// handle fee withdrawal
	feeWithdrawalInstructions, s.tradingFees, err = s.producer.feeWithdrawal(
		env.FeeWithdrawalActions(),
		env.PrevBeaconHeight(),
		s.tradingFees,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, feeWithdrawalInstructions...)

	// handle trade
	tradeInstructions, s.poolPairs, err = s.producer.trade(
		env.TradeActions(),
		env.PrevBeaconHeight(),
		s.poolPairs,
		env.BCHeightBreakPointPrivacyV2(),
		env.Pdexv3BreakPoint(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tradeInstructions...)

	// handle cross pool trade
	crossPoolTradeInstructions, s.poolPairs, s.shares, err = s.producer.crossPoolTrade(
		env.CrossPoolTradeActions(),
		env.PrevBeaconHeight(),
		s.poolPairs,
		s.shares,
		env.Pdexv3BreakPoint(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, crossPoolTradeInstructions...)

	// handle withdrawal
	withdrawalInstructions, s.poolPairs, s.shares, err = s.producer.withdrawal(
		env.WithdrawalActions(),
		env.PrevBeaconHeight(),
		s.poolPairs,
		s.shares,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, withdrawalInstructions...)

	// handle contribution
	contributionInstructions, s.waitingContributions, s.poolPairs, s.shares, err = s.producer.contribution(
		env.ContributionActions(),
		env.PrevBeaconHeight(),
		false,
		metadata.PDEContributionMeta,
		s.waitingContributions,
		s.poolPairs,
		s.shares,
		env.BCHeightBreakPointPrivacyV2(),
		env.Pdexv3BreakPoint(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, contributionInstructions...)

	// handle prv required contribution
	contributionInstructions, s.waitingContributions, s.poolPairs, s.shares, err = s.producer.contribution(
		env.PRVRequiredContributionActions(),
		env.PrevBeaconHeight(),
		true,
		metadata.PDEPRVRequiredContributionRequestMeta,
		s.waitingContributions,
		s.poolPairs,
		s.shares,
		env.BCHeightBreakPointPrivacyV2(),
		env.Pdexv3BreakPoint(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, contributionInstructions...)

	return instructions, nil
}

func (s *stateV1) StoreToDB(env StateEnvironment, stateChange *v2utils.StateChange) error {
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
	*s = *newState
}

func (s *stateV1) ClearCache() {
	s.deletedWaitingContributions = make(map[string]*rawdbv2.PDEContribution)
}

func (s *stateV1) GetDiff(compareState State, stateChange *v2utils.StateChange) (State, *v2utils.StateChange, error) {
	if compareState == nil {
		return nil, stateChange, errors.New("compareState is nil")
	}

	res := newStateV1()
	compareStateV1, ok := compareState.(*stateV1)
	if !ok {
		return nil, stateChange, errors.New("compareState is not stateV1")
	}

	//transform the height in pdeState to be the same as in compareStateV1 (support look up key)
	for key, _ := range compareStateV1.waitingContributions {
		keySplit := strings.Split(key, "-")
		height, _ := math.ParseUint64(keySplit[1])
		s.TransformKeyWithNewBeaconHeight(height)
		break
	}

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
	return res, stateChange, nil
}

func (s *stateV1) WaitingContributions() []byte {
	temp := make(map[string]*rawdbv2.PDEContribution, len(s.waitingContributions))
	for k, v := range s.waitingContributions {
		temp[k] = new(rawdbv2.PDEContribution)
		*temp[k] = *v
	}
	data, _ := json.Marshal(temp)
	return data
}

func (s *stateV1) PoolPairs() []byte {
	temp := make(map[string]*rawdbv2.PDEPoolForPair, len(s.poolPairs))
	for k, v := range s.poolPairs {
		temp[k] = new(rawdbv2.PDEPoolForPair)
		*temp[k] = *v
	}
	data, _ := json.Marshal(temp)
	return data
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

func (s *stateV1) Reader() StateReader {
	return s
}

func (s *stateV1) Validator() StateValidator {
	return s
}

func (s *stateV1) IsValidPoolPairID(poolPairID string) error {
	poolPair, found := s.poolPairs[poolPairID]
	if !found || poolPair == nil {
		return fmt.Errorf("%v pool pair ID can not be found", poolPairID)
	}
	return nil
}
