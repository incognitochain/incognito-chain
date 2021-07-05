package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
	"math/big"
)

//swapRuleV3 ...
type swapRuleV3 struct {
}

func NewSwapRuleV3() *swapRuleV3 {
	return &swapRuleV3{}
}

//GenInstructions generate instructions for swap rule v3
func (s *swapRuleV3) Process(
	shardID byte,
	committees, substitutes []string,
	minCommitteeSize, maxCommitteeSize, typeIns, numberOfFixedValidators int,
	penalty map[string]signaturecounter.Penalty,
) (*instruction.SwapShardInstruction, []string, []string, []string, []string) {

	// @NOTICE: hack code to reduce code complexity
	// All running network need to maintain numberOfFixedValidators equal to minCommitteeSize
	// if numberOfFixedValidators = 0, code execution may go wrong
	minCommitteeSize = numberOfFixedValidators
	//get slashed nodes
	newCommittees, slashingCommittees :=
		s.slashingSwapOut(committees, penalty, numberOfFixedValidators)
	lenSlashedCommittees := len(slashingCommittees)
	//get normal swap out nodes
	newCommittees, normalSwapOutCommittees :=
		s.normalSwapOut(newCommittees, substitutes, len(committees), lenSlashedCommittees, numberOfFixedValidators, maxCommitteeSize)
	swappedOutCommittees := append(slashingCommittees, normalSwapOutCommittees...)

	newCommittees, newSubstitutes, swapInCommittees :=
		s.swapInAfterSwapOut(newCommittees, substitutes, maxCommitteeSize, len(slashingCommittees), len(committees))

	if len(swapInCommittees) == 0 && len(swappedOutCommittees) == 0 {
		return instruction.NewSwapShardInstructionWithShardID(int(shardID)), newCommittees, newSubstitutes, slashingCommittees, normalSwapOutCommittees
	}

	swapShardInstruction := instruction.NewSwapShardInstructionWithValue(
		swapInCommittees,
		swappedOutCommittees,
		int(shardID),
		typeIns,
	)

	return swapShardInstruction, newCommittees, newSubstitutes, slashingCommittees, normalSwapOutCommittees
}

func (s *swapRuleV3) CalculateAssignOffset(lenShardSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int {
	assignOffset := lenCommittees / MAX_ASSIGN_PERCENT_V3
	if assignOffset == 0 && lenCommittees < MAX_ASSIGN_PERCENT_V3 {
		assignOffset = 1
	}
	return assignOffset
}

func (s *swapRuleV3) swapInAfterSwapOut(lenCommitteesAfterSwapOut []string, substitutes []string, maxCommitteeSize int, numberOfSlashingValidators int, lenCommitteesBeforeSwapOut int) ([]string, []string, []string) {
	swapInOffset := s.getSwapInOffset(len(lenCommitteesAfterSwapOut), len(substitutes), maxCommitteeSize, numberOfSlashingValidators, lenCommitteesBeforeSwapOut)

	newCommittees := common.DeepCopyString(lenCommitteesAfterSwapOut)
	swapInCommittees := common.DeepCopyString(substitutes[:swapInOffset])
	newSubstitutes := common.DeepCopyString(substitutes[swapInOffset:])
	newCommittees = append(newCommittees, swapInCommittees...)

	return newCommittees, newSubstitutes, swapInCommittees
}

//getSwapInOffset calculate based on lenCommitteesAfterSwapOut
// swap_in = min(lenSubstitute, lenCommitteesAfterSwapOut/8) but no more than maxCommitteeSize
// Special case: when committees size reach max and no slashing, swap in equal to normal swap out
func (s *swapRuleV3) getSwapInOffset(lenCommitteesAfterSwapOut, lenSubstitutes, maxCommitteeSize, numberOfSlashingValidators, lenCommitteesBeforeSwapOut int) int {
	var offset int
	// special case: no slashing node && committee reach max committee size => normal swap out == swap in
	if numberOfSlashingValidators == 0 && lenCommitteesBeforeSwapOut == maxCommitteeSize {
		offset = lenCommitteesBeforeSwapOut / MAX_SWAP_IN_PERCENT_V3
	} else {
		// normal case
		offset = lenCommitteesAfterSwapOut / MAX_SWAP_IN_PERCENT_V3
	}

	// if committee size after swap out below than maxSwapInPercent => no swap in
	// try to swap in at least one
	if offset == 0 && lenCommitteesAfterSwapOut < MAX_SWAP_IN_PERCENT_V3 {
		offset = 1
	}

	if lenSubstitutes < offset {
		offset = lenSubstitutes
	}

	if lenCommitteesAfterSwapOut+offset > maxCommitteeSize {
		offset = maxCommitteeSize - lenCommitteesAfterSwapOut
	}

	return offset
}

func (s *swapRuleV3) normalSwapOut(committeesAfterSlashing, substitutes []string, lenCommitteesBeforeSlash, lenSlashedCommittees, numberOfFixedValidators, maxCommitteeSize int) ([]string, []string) {

	resNormalSwapOut := []string{}
	tempCommittees := make([]string, len(committeesAfterSlashing))
	copy(tempCommittees, committeesAfterSlashing)

	normalSwapOutOffset := s.getNormalSwapOutOffset(lenCommitteesBeforeSlash, len(substitutes), lenSlashedCommittees, numberOfFixedValidators, maxCommitteeSize)

	resCommittees := append(tempCommittees[:numberOfFixedValidators], tempCommittees[(numberOfFixedValidators+normalSwapOutOffset):]...)
	resNormalSwapOut = committeesAfterSlashing[numberOfFixedValidators : numberOfFixedValidators+normalSwapOutOffset]

	return resCommittees, resNormalSwapOut
}

//getNormalSwapOutOffset calculate normal swapout offset
func (s *swapRuleV3) getNormalSwapOutOffset(lenCommitteesBeforeSlash, lenSubstitutes, lenSlashedCommittees, numberOfFixedValidators, maxCommitteeSize int) int {
	if lenSubstitutes == 0 {
		return 0
	}

	if maxCommitteeSize != lenCommitteesBeforeSlash {
		return 0
	}

	maxSlashingOffset := s.getMaxSlashingOffset(lenCommitteesBeforeSlash, numberOfFixedValidators)
	if lenSlashedCommittees == maxSlashingOffset {
		return 0
	}

	maxNormalSwapOutOffset := lenCommitteesBeforeSlash / MAX_SWAP_OUT_PERCENT_V3
	if maxNormalSwapOutOffset > lenCommitteesBeforeSlash-numberOfFixedValidators {
		maxNormalSwapOutOffset = lenCommitteesBeforeSlash - numberOfFixedValidators
	}

	if lenSlashedCommittees >= maxNormalSwapOutOffset {
		return 0
	}

	offset := maxNormalSwapOutOffset - lenSlashedCommittees

	if offset > lenSubstitutes {
		offset = lenSubstitutes
	}

	lenCommitteesAfterSlash := lenCommitteesBeforeSlash - lenSlashedCommittees
	if offset > lenCommitteesAfterSlash-numberOfFixedValidators {
		offset = lenCommitteesAfterSlash - numberOfFixedValidators
	}

	return offset
}

//slashingSwapOut only consider all penalties type as one type
func (s *swapRuleV3) slashingSwapOut(committees []string, penalty map[string]signaturecounter.Penalty, numberOfFixedValidators int) ([]string, []string) {
	fixedCommittees := common.DeepCopyString(committees[:numberOfFixedValidators])
	flexCommittees := common.DeepCopyString(committees[numberOfFixedValidators:])
	flexAfterSlashingCommittees := []string{}
	slashingCommittees := []string{}

	maxSlashingOffset := s.getMaxSlashingOffset(len(committees), numberOfFixedValidators)

	for _, flexCommittee := range flexCommittees {
		if _, ok := penalty[flexCommittee]; ok && maxSlashingOffset > 0 {
			slashingCommittees = append(slashingCommittees, flexCommittee)
			maxSlashingOffset--
		} else {
			flexAfterSlashingCommittees = append(flexAfterSlashingCommittees, flexCommittee)
		}
	}

	newCommittees := append(fixedCommittees, flexAfterSlashingCommittees...)
	return newCommittees, slashingCommittees
}

// getMaxSlashingOffset calculate maximum slashing offset, fixed nodes must be spare
// max_slashing_offset = 1/3 committee length
func (s *swapRuleV3) getMaxSlashingOffset(lenCommittees, numberOfFixedValidators int) int {
	if lenCommittees == numberOfFixedValidators {
		return 0
	}

	offset := lenCommittees / MAX_SLASH_PERCENT_V3
	if offset > lenCommittees-numberOfFixedValidators {
		offset = lenCommittees - numberOfFixedValidators
	}

	return offset
}

// calculateNewSubstitutePosition calculate reverse shardID for candidate
func calculateNewSubstitutePosition(candidate string, rand int64, total int) (pos int) {

	seed := candidate + fmt.Sprintf("%v", rand)
	hash := common.HashB([]byte(seed))

	// Using big.Int to convert a random Hash value to an integer
	temp := new(big.Int)
	temp.SetBytes(hash[:])
	data := int(temp.Int64())
	if data < 0 {
		data *= -1
	}

	pos = data % total

	return pos
}

func (s *swapRuleV3) Version() int {
	return swapRuleDCSVersion
}
