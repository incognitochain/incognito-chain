package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
)

//processUnstakeInstruction : process unstake instruction from beacon block

func (b BeaconCommitteeStateV1) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {

	for _, committeePublicKey := range unstakeInstruction.PublicKeys {
		if common.IndexOfStr(committeePublicKey, env.allCandidateSubstituteCommittee) == -1 {
			// if not found then delete auto staking data for this public key if present
			if _, ok := b.autoStake[committeePublicKey]; ok {
				delete(b.autoStake, committeePublicKey)
			}
		} else {
			// if found in committee list then turn off auto staking
			if _, ok := b.autoStake[committeePublicKey]; ok {
				b.autoStake[committeePublicKey] = false
				committeeChange.Unstake = append(committeeChange.Unstake, committeePublicKey)
			}
		}
	}

	return nil, nil
}
