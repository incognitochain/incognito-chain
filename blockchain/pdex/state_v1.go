package pdex

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/metadata"
)

func InitStateV1() *stateV1 {
	return nil
}

type stateV1 struct {
	stateBase
	producer  stateProducerV1
	processor stateProcessorV1
}

func (s *stateV1) Version() uint {
	return BasicVersion
}

func (s *stateV1) Clone() State {
	res := &stateV1{}
	res.stateBase = *s.stateBase.Clone().(*stateBase)
	res.producer = s.producer
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
	tempInstructions, err := s.producer.buildInstructionsForFeeWithdrawal(
		env.FeeWithdrawalActions(),
		env.BeaconHeight(),
		s.tradingFees,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle trade
	tempInstructions, err = s.producer.buildInstructionsForTrade(
		env.TradeActions(),
		env.BeaconHeight(),
		s.poolPairs,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle cross pool trade
	tempInstructions, err = s.producer.buildInstructionsForCrossPoolTrade(
		env.CrossPoolTradeActions(),
		env.BeaconHeight(),
		s.poolPairs,
		s.shares,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle withdrawal
	tempInstructions, err = s.producer.buildInstructionsForWithdrawal(
		env.WithdrawalActions(),
		env.BeaconHeight(),
		s.poolPairs,
		s.shares,
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle contribution
	tempInstructions, err = s.producer.buildInstructionsForContribution(
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
	instructions = append(instructions, tempInstructions...)

	// handle prv required contribution
	tempInstructions, err = s.producer.buildInstructionsForContribution(
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
	instructions = append(instructions, tempInstructions...)

	return instructions, nil
}

func (s *stateV1) Upgrade(env StateEnvironment) State {
	var state State
	return state
}
