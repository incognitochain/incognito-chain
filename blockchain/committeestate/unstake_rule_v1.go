package committeestate

import "github.com/incognitochain/incognito-chain/instruction"

//processUnstakeInstruction : process unstake instruction from beacon block
func (b BeaconCommitteeStateV1) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	return nil, nil
}
