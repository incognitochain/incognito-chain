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

	//get slashed nodes
	newCommittees, slashingCommittees := s.slashingSwapOut(committees, penalty, numberOfFixedValidators)
	lenSlashedCommittees := len(slashingCommittees)
	//get normal swap out nodes
	newCommittees, normalSwapOutCommittees := s.normalSwapOut(newCommittees, substitutes, len(committees), lenSlashedCommittees, numberOfFixedValidators, minCommitteeSize, maxCommitteeSize)
	swappedOutCommittees := append(slashingCommittees, normalSwapOutCommittees...)

	newCommittees, newSubstitutes, swapInCommittees :=
		s.swapInAfterSwapOut(newCommittees, substitutes, maxCommitteeSize, MAX_SWAP_IN_PERCENT_V3, len(slashingCommittees), len(committees))

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

func (s *swapRuleV3) CalculateAssignOffset(lenShardSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int {
	assignOffset := lenCommittees / MAX_ASSIGN_PERCENT_V3
	if assignOffset == 0 && lenCommittees < MAX_ASSIGN_PERCENT_V3 {
		assignOffset = 1
	}
	return assignOffset
}

//TODO: @hung remove unused parameter numberOfFixedValidators
func (s *swapRuleV3) swapInAfterSwapOut(
	committees []string,
	substitutes []string,
	maxCommitteeSize int,
	maxSwapInPercent int,
	numberOfSlashingValidators int,
	numberOfOldCommittees int,
) ([]string, []string, []string) {
	swapInOffset := s.getSwapInOffset(len(committees), len(substitutes), maxSwapInPercent, maxCommitteeSize, numberOfSlashingValidators, numberOfOldCommittees)

	newCommittees := common.DeepCopyString(committees)
	swapInCommittees := common.DeepCopyString(substitutes[:swapInOffset])
	newSubstitutes := common.DeepCopyString(substitutes[swapInOffset:])
	newCommittees = append(newCommittees, swapInCommittees...)

	return newCommittees, newSubstitutes, swapInCommittees
}

//TODO: @tin calculate based on lenCommitteesAfterSwapOut or lenCommitteesBeforeSwapOut?
// In document, vacant_slot = min(max_committee_size/8, 1/8*len(shardCommittee), but getSwapInOffset doesn't stick to this formula
func (s *swapRuleV3) getSwapInOffset(
	lenCommitteesAfterSwapOut,
	lenSubstitutes,
	maxSwapInPercent,
	maxCommitteeSize,
	numberOfSlashingValidators,
	numberOfOldCommittees int,
) int {
	var offset int
	// special case: no slashing node && committee reach max committee size => normal swap out == swap in
	if numberOfSlashingValidators == 0 && numberOfOldCommittees == maxCommitteeSize {
		offset = maxCommitteeSize / maxSwapInPercent
	} else {
		// normal case
		offset = lenCommitteesAfterSwapOut / maxSwapInPercent
	}

	if lenSubstitutes < offset {
		offset = lenSubstitutes
	}

	// hack case: many fixed nodes in committee
	if lenCommitteesAfterSwapOut+offset > maxCommitteeSize {
		offset = maxCommitteeSize - lenCommitteesAfterSwapOut
	}

	if offset == 0 && lenCommitteesAfterSwapOut < maxSwapInPercent && lenSubstitutes > 0 {
		//TODO: @tin offset + currentCommitteeSize maybe > maxCommitteeSize
		offset = 1
	}

	return offset
}

func (s *swapRuleV3) normalSwapOut(committees, substitutes []string, lenBeforeSlashedCommittees, lenSlashedCommittees, numberOfFixedValidators, minCommitteeSize, maxCommitteeSize int) ([]string, []string) {

	resNormalSwapOut := []string{}
	tempCommittees := make([]string, len(committees))
	copy(tempCommittees, committees)

	normalSwapOutOffset := s.getNormalSwapOutOffset(lenBeforeSlashedCommittees, len(substitutes), lenSlashedCommittees, numberOfFixedValidators, minCommitteeSize, maxCommitteeSize)

	resCommittees := append(tempCommittees[:numberOfFixedValidators], tempCommittees[(numberOfFixedValidators+normalSwapOutOffset):]...)
	resNormalSwapOut = committees[numberOfFixedValidators : numberOfFixedValidators+normalSwapOutOffset]

	return resCommittees, resNormalSwapOut
}

//getNormalSwapOutOffset calculate normal swapout offset
// max_Normal_Swap_Out_offset = min(C/8 - SL, C - numberOfFixedValidators - SL)
func (s *swapRuleV3) getNormalSwapOutOffset(lenCommitteesBeforeSlash, lenSubstitutes, lenSlashedCommittees, numberOfFixedValidators, minCommitteeSize, maxCommitteeSize int) int {
	if lenSubstitutes == 0 {
		return 0
	}

	if lenCommitteesBeforeSlash < maxCommitteeSize {
		return 0
	}

	offset := lenCommitteesBeforeSlash / MAX_SWAP_OUT_PERCENT_V3
	if lenSlashedCommittees >= offset {
		if lenSlashedCommittees == offset {
			if offset == 0 {
				if lenCommitteesBeforeSlash < MAX_SWAP_OUT_PERCENT_V3 && lenSubstitutes > 0 {
					return 1
				}
			}
		}
		return 0
	}
	if lenCommitteesBeforeSlash < minCommitteeSize {
		return 0
	}

	offset = offset - lenSlashedCommittees
	if offset > lenSubstitutes {
		offset = lenSubstitutes
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

	//TODO: @tin don't change hash to int like this, because the maximum value is only 255*32
	//for _, v := range hash {
	//	data += int(v)
	//}

	// Using big.Int to convert a random Hash value to an integer
	temp := new(big.Int)
	temp.SetBytes(hash[:])
	data := int(temp.Int64())
	if data < 0 {
		data *= -1
	}

	//TODO: @tin what if total == 0
	// pos < total => never insert at the end of list?
	pos = data % total
	if pos == 0 {
		//TODO: @tin why set pos = 1 when it's equal to 0, because total might equal to 0
		pos = 1
	}

	return pos
}

func (s *swapRuleV3) clone() SwapRuleProcessor {
	return &swapRuleV3{}
}

func (s *swapRuleV3) Version() int {
	return swapRuleDCSVersion
}
