package committeestate

import (
	"errors"

	"github.com/incognitochain/incognito-chain/instruction"
)

//BuildIncurredInstructions : BuildIncurredInstruction from instructions
func (engine BeaconCommitteeEngine) BuildIncurredInstructions(
	env *BeaconCommitteeStateEnvironment) (
	[][]string, error) {
	newB := NewBeaconCommitteeStateV1()
	engine.beaconCommitteeStateV1.clone(newB)
	committeeChange := NewCommitteeChange()

	incurredInstructions := [][]string{}
	if env == nil {
		return incurredInstructions, errors.New("Environment Variable Is Null")
	}
	if len(env.BeaconInstructions) == 0 {
		return incurredInstructions, nil
	}
	var err error

	env.substituteCandidates, err = newB.getSubstituteCandidates()
	if err != nil {
		return incurredInstructions, err
	}
	env.allSubstituteCommittees, err = newB.getAllSubstituteCommittees()
	if err != nil {
		return incurredInstructions, err
	}
	for _, inst := range env.BeaconInstructions {
		switch inst[0] {
		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP unstake instruction %+v, error %+v", inst, err)
				return incurredInstructions, err
			}
			_, incurredInsFromUnstake, err :=
				newB.processUnstakeInstruction(unstakeInstruction, env, committeeChange)
			if err != nil {
				return incurredInstructions, NewCommitteeStateError(ErrBuildIncurredInstruction, err)
			}
			if incurredInsFromUnstake != nil {
				incurredInstructions = append(incurredInstructions, incurredInsFromUnstake...)
			}
		}
	}

	return incurredInstructions, nil
}
