package committeestate

import "github.com/incognitochain/incognito-chain/instruction"

//processUnstakeInstruction : ....
func (b BeaconCommitteeStateV1) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (map[string]bool, *CommitteeChange, error) {
	return nil, nil, nil
}
