package instruction

import (
	"fmt"
	"strings"
)

type StakeInstruction struct {
	Action          string
	PublicKeys      []string
	Chain           string
	TxStakes        []string
	RewardReceivers []string
	AutoStakingFlag []bool
}

func NewStakeInstructionWithValue(publicKeys []string, chain string, txStakes []string, rewardReceivers []string, autoStakingFlag []bool) *StakeInstruction {
	return &StakeInstruction{Action: STAKE_ACTION, PublicKeys: publicKeys, Chain: chain, TxStakes: txStakes, RewardReceivers: rewardReceivers, AutoStakingFlag: autoStakingFlag}
}

func NewStakeInstruction() *StakeInstruction {
	return &StakeInstruction{Action: STAKE_ACTION}
}

func importStakeInstructionFromString(instruction []string) (*StakeInstruction, error) {
	if err := validateStakeInstructionSanityData(instruction); err != nil {
		return nil, err
	}
	stakeInstruction := NewStakeInstruction()
	stakeInstruction.PublicKeys = strings.Split(instruction[1], SPLITTER)
	stakeInstruction.TxStakes = strings.Split(instruction[3], SPLITTER)
	stakeInstruction.RewardReceivers = strings.Split(instruction[4], SPLITTER)
	tempAutoStakings := strings.Split(instruction[5], SPLITTER)
	autoStakeFlags := []bool{}
	for _, v := range tempAutoStakings {
		if v == TRUE {
			autoStakeFlags = append(autoStakeFlags, true)
		} else {
			autoStakeFlags = append(autoStakeFlags, false)
		}
	}
	stakeInstruction.AutoStakingFlag = autoStakeFlags
	stakeInstruction.Chain = instruction[2]
	return stakeInstruction, nil
}

func (s *StakeInstruction) ToString() []string {
	stakeInstructionStr := []string{STAKE_ACTION}
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(s.PublicKeys, SPLITTER))
	stakeInstructionStr = append(stakeInstructionStr, s.Chain)
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(s.TxStakes, SPLITTER))
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(s.RewardReceivers, SPLITTER))
	tempStopAutoStakeFlag := []string{}
	for _, v := range s.AutoStakingFlag {
		if v == true {
			tempStopAutoStakeFlag = append(tempStopAutoStakeFlag, TRUE)
		} else {
			tempStopAutoStakeFlag = append(tempStopAutoStakeFlag, FALSE)
		}
	}
	stakeInstructionStr = append(stakeInstructionStr, strings.Join(tempStopAutoStakeFlag, SPLITTER))
	return stakeInstructionStr
}

// validate stake instruction sanity
// beaconprocess.go: 1122 - 1165
// beaconproducer.go: 386
func validateStakeInstructionSanityData(instruction []string) error {
	if len(instruction) != 6 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != STAKE_ACTION {
		return fmt.Errorf("invalid stake action, %+v", instruction)
	}
	if instruction[2] != SHARD_INST && instruction[2] != BEACON_INST {
		return fmt.Errorf("invalid chain id, %+v", instruction)
	}
	publicKeys := strings.Split(instruction[1], SPLITTER)
	txStakes := strings.Split(instruction[3], SPLITTER)
	rewardReceivers := strings.Split(instruction[4], SPLITTER)
	autoStakings := strings.Split(instruction[5], SPLITTER)
	if len(publicKeys) != len(txStakes) {
		return fmt.Errorf("invalid public key & tx stake length, %+v", instruction)
	}
	if len(rewardReceivers) != len(txStakes) {
		return fmt.Errorf("invalid reward receivers & tx stake length, %+v", instruction)
	}
	if len(rewardReceivers) != len(autoStakings) {
		return fmt.Errorf("invalid reward receivers & tx auto staking length, %+v", instruction)
	}
	return nil
}
