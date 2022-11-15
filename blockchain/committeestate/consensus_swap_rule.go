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
	ProcessBeacon(
		committees, substitutes []string,
		minCommitteeSize, maxCommitteeSize, numberOfFixedValidators int,
		reputation map[string]uint64,
		performance map[string]uint64,
	) (
		newCommittees []string,
		newSubstitutes []string,
		swapOutList []string,
		slashedList []string,
	)
	CalculateAssignOffset(lenSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int
	Version() int
}

func GetSwapRuleVersion(beaconHeight, stakingFlowV3Height uint64) SwapRuleProcessor {

	var swapRule SwapRuleProcessor

	if beaconHeight >= stakingFlowV3Height {
		Logger.log.Infof("Beacon Height %+v, using Swap Rule V3", beaconHeight)
		swapRule = NewSwapRuleV3()
	} else {
		Logger.log.Infof("Beacon Height %+v, using Swap Rule V2", beaconHeight)
		swapRule = NewSwapRuleV2()
	}

	return swapRule
}
