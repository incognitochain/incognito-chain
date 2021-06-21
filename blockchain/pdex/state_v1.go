package pdex

import (
	"github.com/incognitochain/incognito-chain/metadata"
)

func InitStateV1() *stateV1 {
	return nil
}

type stateV1 struct {
	stateBase
	producer stateProducerV1
}

func (s *stateV1) Version() uint {
	return BasicVersion
}

func (s *stateV1) Clone() State {
	var state State
	return state
}

func (s *stateV1) Process(env StateEnvironment) error {
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
