package committeestate

import (
	"fmt"
	"strings"
)

type StakeInstruction struct {
	Action          int
	InPublicKey     string
	OutPublicKey    string
	Chain           int
	RewardReceiver  string
	AutoStakingFlag bool
}

type ImportFromString
// validate stake instruction sanity
func validateStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 6 {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid length, %+v", instruction))
	}
	if instruction[0] != stakeAction {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid swap action, %+v", instruction))
	}
	if instruction[2] != shardInst && instruction[2] != beaconInst {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid swap action, %+v", instruction))
	}
	publicKeys := strings.Split(instruction[1], splitter)
	txStakes := strings.Split(instruction[3], splitter)
	rewardReceivers := strings.Split(instruction[4], splitter)
	autoStakings := strings.Split(instruction[5], splitter)
	if len(publicKeys) != len(txStakes) {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid public key & tx stake length, %+v", instruction))
	}
	if len(rewardReceivers) != len(txStakes) {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid reward receivers & tx stake length, %+v", instruction))
	}
	if len(rewardReceivers) != len(autoStakings) {
		return NewCommitteeStateError(ErrStakeInstructionSanity, fmt.Errorf("invalid reward receivers & tx auto staking length, %+v", instruction))
	}
	return nil
}
