package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"

	"github.com/incognitochain/incognito-chain/instruction"
)

// swapRuleV2 ...
type swapRuleV2 struct {
}

func NewSwapRuleV2() *swapRuleV2 {
	return &swapRuleV2{}
}

func (s *swapRuleV2) ProcessBeacon(
	committees, substitutes []string,
	minCommitteeSize, maxCommitteeSize, numberOfFixedValidators int,
	reputation map[string]uint64,
	performance map[string]uint64,
) (
	newCommittees []string,
	newSubstitutes []string,
	swapOutList []string,
	slashedList []string,
) {
	return nil, nil, nil, nil
}

// genInstructions
func (s *swapRuleV2) Process(
	shardID byte,
	committees, substitutes []string,
	minCommitteeSize, maxCommitteeSize, typeIns, numberOfFixedValidators int,
	penalty map[string]signaturecounter.Penalty,
) (*instruction.SwapShardInstruction, []string, []string, []string, []string) {

	committees, slashingCommittees, normalSwapCommittees :=
		s.slashingSwapOut(committees, substitutes, penalty, minCommitteeSize, numberOfFixedValidators)

	swappedOutCommittees := append(slashingCommittees, normalSwapCommittees...)

	newCommittees, newSubstitutes, swapInCommittees :=
		s.swapInAfterSwapOut(committees, substitutes, maxCommitteeSize)

	if len(swapInCommittees) == 0 && len(swappedOutCommittees) == 0 {
		return instruction.NewSwapShardInstruction(), newCommittees, newSubstitutes, slashingCommittees, normalSwapCommittees
	}

	swapShardInstruction := instruction.NewSwapShardInstructionWithValue(
		swapInCommittees,
		swappedOutCommittees,
		int(shardID),
		typeIns,
	)

	return swapShardInstruction, newCommittees, newSubstitutes, slashingCommittees, normalSwapCommittees
}

// getSwapOutOffset assumes that numberOfFixedValidator <= minCommitteeSize and won't replace fixed nodes
// CONDITION:
// #1 swapOutOffset <= floor(numberOfCommittees/6)
// #2 swapOutOffset <=  numberOfCommittees - numberOfFixedValidator
// #3 committees length after both swap out and swap in must remain >= minCommitteeSize
// #4 swap operation must begin from start position (which is fixed node validator position) as a queue
// #5 number of swap out nodes >= number of swap in nodes
func (s *swapRuleV2) getSwapOutOffset(numberOfSubstitutes, numberOfCommittees, numberOfFixedValidator, minCommitteeSize int) int {

	swapOffset := numberOfCommittees / MAX_SWAP_OR_ASSIGN_PERCENT_V2
	if swapOffset == 0 {
		return 0
	}

	if swapOffset > numberOfCommittees-numberOfFixedValidator {
		swapOffset = numberOfCommittees - numberOfFixedValidator
	}

	noReplaceOffset := 0
	for swapOffset > 0 && numberOfCommittees > minCommitteeSize {
		swapOffset--
		noReplaceOffset++
		numberOfCommittees--
	}

	replaceSwapOffset := swapOffset
	if numberOfSubstitutes < swapOffset {
		replaceSwapOffset = numberOfSubstitutes
	}

	return noReplaceOffset + replaceSwapOffset
}

// slashingSwapOut swap node out of committee
// because of penalty or end of epoch
func (s *swapRuleV2) slashingSwapOut(
	committees, substitutes []string,
	penalty map[string]signaturecounter.Penalty,
	minCommitteeSize int,
	numberOfFixedValidator int,
) (
	[]string,
	[]string,
	[]string,
) {
	if len(committees) == numberOfFixedValidator {
		return committees, []string{}, []string{}
	}

	fixedCommittees := make([]string, len(committees[:numberOfFixedValidator]))
	copy(fixedCommittees, committees[:numberOfFixedValidator])
	flexCommittees := make([]string, len(committees[numberOfFixedValidator:]))
	copy(flexCommittees, committees[numberOfFixedValidator:])
	flexAfterSlashingCommittees := []string{}
	slashingCommittees := []string{}

	swapOutOffset := s.getSwapOutOffset(len(substitutes), len(committees), numberOfFixedValidator, minCommitteeSize)

	for _, flexCommittee := range flexCommittees {
		if _, ok := penalty[flexCommittee]; ok && swapOutOffset > 0 {
			slashingCommittees = append(slashingCommittees, flexCommittee)
			swapOutOffset--
		} else {
			flexAfterSlashingCommittees = append(flexAfterSlashingCommittees, flexCommittee)
		}
	}

	normalSwapOutCommittees := make([]string, len(flexAfterSlashingCommittees[:swapOutOffset]))
	copy(normalSwapOutCommittees, flexAfterSlashingCommittees[:swapOutOffset])

	flexAfterSlashingCommittees = flexAfterSlashingCommittees[swapOutOffset:]

	committees = append(fixedCommittees, flexAfterSlashingCommittees...)
	return committees, slashingCommittees, normalSwapOutCommittees
}

// swapInAfterSwapOut must be perform after normalSwapOut function is executed
// swap in as many as possible
// output:
// #1 new committee list
// #2 new substitutes list
// #3 swapped in committee list (from substitutes)
func (s *swapRuleV2) swapInAfterSwapOut(committees, substitutes []string, maxCommitteeSize int) (
	[]string,
	[]string,
	[]string,
) {
	vacantSlot := maxCommitteeSize - len(committees)
	if vacantSlot > len(substitutes) {
		vacantSlot = len(substitutes)
	}
	newCommittees := substitutes[:vacantSlot]
	committees = append(committees, newCommittees...)
	substitutes = substitutes[vacantSlot:]
	return committees, substitutes, newCommittees
}

func (s *swapRuleV2) CalculateAssignOffset(lenShardSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int {
	assignPerShard := s.getSwapOutOffset(
		lenShardSubstitute,
		lenCommittees,
		numberOfFixedValidators,
		minCommitteeSize,
	)

	if assignPerShard == 0 {
		if lenCommittees < MAX_SWAP_OR_ASSIGN_PERCENT_V2 || lenCommittees-numberOfFixedValidators == 0 {
			assignPerShard = 1
		}
	}
	return assignPerShard
}

func (s *swapRuleV2) clone() SwapRuleProcessor {
	return &swapRuleV2{}
}

func (s *swapRuleV2) Version() int {
	return swapRuleSlashingVersion
}
