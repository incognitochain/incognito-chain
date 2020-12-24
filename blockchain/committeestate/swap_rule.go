package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/instruction"
)

type SwapRule interface {
	GenInstructions(
		shardID byte,
		committees, substitutes []string,
		minCommitteeSize, maxCommitteeSize, typeIns, numberOfFixedValidators, dcsMaxCommitteeSize, dcsMinCommitteeSize int,
		penalty map[string]signaturecounter.Penalty,
	) (
		*instruction.SwapShardInstruction, []string, []string, []string, []string) // instruction, newCommitteees, newSubstitutes, slashingCommittees, normalSwapCommittees
	AssignOffset(lenSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int
	Version() int
}

func cloneSwapRuleByVersion(swapRule SwapRule) SwapRule {
	var res SwapRule
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

func SwapRuleByEnv(env *BeaconCommitteeStateEnvironment) SwapRule {
	var swapRule SwapRule
	if env.Epoch >= env.SwapRuleV3Epoch {
		swapRule = NewSwapRuleV3()
	} else {
		if env.Epoch >= env.SwapRuleV2Epoch {
			swapRule = NewSwapRuleV2()
		}
	}
	return swapRule
}
