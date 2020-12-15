package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/instruction"
)

//swapRuleV3 ...
type swapRuleV3 struct {
	versionName string
}

func NewSwapRuleV3() *swapRuleV3 {
	return &swapRuleV3{}
}

//@tin
func (s *swapRuleV3) GenInstructions(
	shardID byte,
	committees, substitutes []string,
	minCommitteeSize, maxCommitteeSize, typeIns, numberOfFixedValidators, dcsMaxCommitteeSize, dcsMinCommitteeSize int,
	penalty map[string]signaturecounter.Penalty,
) (*instruction.SwapShardInstruction, []string, []string, []string, []string) {

	//get slashed nodes
	newCommittees, slashingCommittees := s.slashingSwapOut(committees, penalty, minCommitteeSize, numberOfFixedValidators, MAX_SLASH_PERCENT)
	lenSlashedCommittees := len(slashingCommittees)
	//get normal swap out nodes
	normalSwapOutCommittees := s.normalSwapOut(
		newCommittees, substitutes, len(committees), lenSlashedCommittees, numberOfFixedValidators,
		MAX_SWAP_OUT_PERCENT, dcsMaxCommitteeSize, dcsMinCommitteeSize)
	swappedOutCommittees := append(slashingCommittees, normalSwapOutCommittees...)
	//get committees list after swap out
	newCommittees = newCommittees[:len(newCommittees)-len(normalSwapOutCommittees)]

	newCommittees, newSubstitutes, swapInCommittees :=
		s.swapInAfterSwapOut(newCommittees, substitutes, numberOfFixedValidators,
			MAX_SWAP_IN_PERCENT, dcsMaxCommitteeSize, dcsMinCommitteeSize)

	if len(swapInCommittees) == 0 && len(swappedOutCommittees) == 0 {
		return instruction.NewSwapShardInstruction(), newCommittees, newSubstitutes, slashingCommittees, normalSwapOutCommittees
	}

	swapShardInstruction := instruction.NewSwapShardInstructionWithValue(
		swapInCommittees,
		swappedOutCommittees,
		int(shardID),
		typeIns,
	)

	return swapShardInstruction, newCommittees, newSubstitutes, slashingCommittees, normalSwapOutCommittees
}

func (s *swapRuleV3) AssignOffset(lenShardSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int {
	assignOffset := lenCommittees / MAX_ASSIGN_PERCENT
	if assignOffset == 0 && lenCommittees < MAX_ASSIGN_PERCENT {
		assignOffset = 1
	}
	return assignOffset
}

func (s *swapRuleV3) swapInAfterSwapOut(
	committees, substitutes []string,
	maxSwapInPercent, numberOfFixedValidators,
	dcsMaxCommitteeSize, dcsMinCommitteeSize int,
) (
	[]string, []string, []string,
) {
	resCommittees := committees
	resSubstitutes := substitutes
	resSwapInCommittees := []string{}
	swapInOffset := s.getSwapInOffset(len(committees), len(substitutes), maxSwapInPercent, numberOfFixedValidators, dcsMaxCommitteeSize, dcsMinCommitteeSize)

	resSwapInCommittees = append(resSwapInCommittees, substitutes[:swapInOffset]...)
	resSubstitutes = resSubstitutes[swapInOffset:]
	resCommittees = append(resCommittees, resSwapInCommittees...)

	return resCommittees, resSubstitutes, resSwapInCommittees
}

//@tin
func (s *swapRuleV3) getSwapInOffset(
	lenCommitteesAfterSwapOut, lenSubstitutes int,
	maxSwapInPercent, numberOfFixedValidators,
	dcsMaxCommitteeSize, dcsMinCommitteeSize int,
) int {
	swapInOffset := 0

	if lenSubstitutes < lenCommitteesAfterSwapOut {
		if lenCommitteesAfterSwapOut > dcsMinCommitteeSize {
			return 0
		} else {
			swapInOffset = lenCommitteesAfterSwapOut / maxSwapInPercent
		}
	} else {
		swapInOffset = lenCommitteesAfterSwapOut / maxSwapInPercent
	}

	if swapInOffset+lenCommitteesAfterSwapOut > dcsMaxCommitteeSize {
		swapInOffset = dcsMaxCommitteeSize - lenCommitteesAfterSwapOut
	}
	return swapInOffset
}

func (s *swapRuleV3) normalSwapOut(committees, substitutes []string,
	lenBeforeSlashedCommittees, lenSlashedCommittees, maxSwapOutPercent, numberOfFixedValidators,
	dcsMaxCommitteeSize, dcsMinCommitteeSize int,
) []string {
	resNormalSwapOut := []string{}
	normalSwapOutOffset := s.getNormalSwapOutOffset(
		lenBeforeSlashedCommittees, len(substitutes),
		lenSlashedCommittees, maxSwapOutPercent, numberOfFixedValidators,
		dcsMaxCommitteeSize, dcsMinCommitteeSize)

	resNormalSwapOut = committees[numberOfFixedValidators : numberOfFixedValidators+normalSwapOutOffset]

	return resNormalSwapOut
}

func (s *swapRuleV3) getNormalSwapOutOffset(
	lenCommitteesBeforeSlash, lenSubstitutes,
	lenSlashedCommittees, maxSwapOutPercent, numberOfFixedValidators,
	dcsMaxCommitteeSize, dcsMinCommitteeSize int,
) int {
	normalSwapOutOffset := 0
	if lenSlashedCommittees < lenCommitteesBeforeSlash/maxSwapOutPercent {
		if lenSubstitutes >= 4*lenCommitteesBeforeSlash {
			if lenCommitteesBeforeSlash >= dcsMaxCommitteeSize {
				normalSwapOutOffset = lenCommitteesBeforeSlash/maxSwapOutPercent - lenSlashedCommittees
			} else {
				normalSwapOutOffset = 0
			}
		} else {
			if lenCommitteesBeforeSlash < dcsMinCommitteeSize {
				normalSwapOutOffset = 0
			} else {
				normalSwapOutOffset = lenCommitteesBeforeSlash/maxSwapOutPercent - lenSlashedCommittees
				if lenCommitteesBeforeSlash-lenSlashedCommittees-normalSwapOutOffset < dcsMinCommitteeSize {
					normalSwapOutOffset = lenCommitteesBeforeSlash - lenSlashedCommittees - dcsMinCommitteeSize
				}
			}
		}
	}

	if normalSwapOutOffset < 0 {
		normalSwapOutOffset = 0
	}

	return normalSwapOutOffset
}

func (s *swapRuleV3) slashingSwapOut(
	committees []string,
	penalty map[string]signaturecounter.Penalty,
	minCommitteeSize, numberOfFixedValidators, maxSlashOutPercent int,
) (
	[]string,
	[]string,
) {

	fixedCommittees := make([]string, len(committees[:numberOfFixedValidators]))
	copy(fixedCommittees, committees[:numberOfFixedValidators])
	flexCommittees := make([]string, len(committees[numberOfFixedValidators:]))
	copy(flexCommittees, committees[numberOfFixedValidators:])
	flexAfterSlashingCommittees := []string{}
	slashingCommittees := []string{}

	slashingOffset := s.getSlashingOffset(len(committees), minCommitteeSize, numberOfFixedValidators, maxSlashOutPercent)
	for _, flexCommittee := range flexCommittees {
		if _, ok := penalty[flexCommittee]; ok && slashingOffset > 0 {
			slashingCommittees = append(slashingCommittees, flexCommittee)
			slashingOffset--
		} else {
			flexAfterSlashingCommittees = append(flexAfterSlashingCommittees, flexCommittee)
		}
	}

	newCommittees := append(fixedCommittees, flexAfterSlashingCommittees...)
	return newCommittees, slashingCommittees
}

func (s *swapRuleV3) getSlashingOffset(
	lenCommittees, minCommitteeSize, numberOfFixedValidators, maxSlashOutPercent int,
) int {
	if lenCommittees == minCommitteeSize {
		return 0
	}
	return lenCommittees / maxSlashOutPercent
}

func (s *swapRuleV3) clone() SwapRule {
	return &swapRuleV3{}
}

func (s *swapRuleV3) Version() int {
	return swapRuleDCSVersion
}
