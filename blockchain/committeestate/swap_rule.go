package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/instruction"
)

type SwapRuleProcessor interface {
	Process(
		shardID byte,
		committees, substitutes []string,
		minCommitteeSize, maxCommitteeSize, typeIns, numberOfFixedValidators int,
		penalty map[string]signaturecounter.Penalty,
	) (
		*instruction.SwapShardInstruction, []string, []string, []string, []string) // instruction, newCommitteees, newSubstitutes, slashingCommittees, normalSwapCommittees
	CalculateAssignOffset(lenSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int
	Version() int
}

func cloneSwapRuleByVersion(swapRule SwapRuleProcessor) SwapRuleProcessor {
	var res SwapRuleProcessor
	if swapRule != nil {
		switch swapRule.Version() {
		case swapRuleSlashingVersion:
			res = swapRule.(*swapRuleV2).clone()
		case swapRuleDCSVersion:
			res = swapRule.(*swapRuleV3).clone()
		case swapRuleTestVersion:
			res = swapRule
		default:
			panic("Not implement this version yet")
		}
	}
	return res
}

func SwapRuleByEnv(env *BeaconCommitteeStateEnvironment) SwapRuleProcessor {
	var swapRule SwapRuleProcessor
	if env.BeaconHeight >= env.StakingV3Height {
		swapRule = NewSwapRuleV3()
	} else {
		swapRule = NewSwapRuleV2()
	}
	return swapRule
}
