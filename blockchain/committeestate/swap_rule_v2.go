package committeestate

import (
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
)

//swapRuleV2 ...
type swapRuleV2 struct {
}

func NewSwapRuleV2() *swapRuleV2 {
	return &swapRuleV2{}
}

//genInstructions
func (s *swapRuleV2) GenInstructions(
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

	swapOffset := numberOfCommittees / MAX_SWAP_OR_ASSIGN_PERCENT
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

// assignShardCandidateV2 assign unassignedCommonPool into shard pool with random number
func assignShardCandidateV2(candidates []string, numberOfValidators []int, rand int64) map[byte][]string {
	total := 0
	for _, v := range numberOfValidators {
		total += v
	}
	n := byte(len(numberOfValidators))
	sortedShardIDs := sortShardIDByIncreaseOrder(numberOfValidators)
	m := getShardIDPositionFromArray(sortedShardIDs)
	assignedCandidates := make(map[byte][]string)
	candidateRandomShardID := make(map[string]byte)
	for _, candidate := range candidates {
		randomPosition := calculateCandidatePosition(candidate, rand, total)
		shardID := 0
		tempPosition := numberOfValidators[shardID]
		for randomPosition > tempPosition {
			shardID++
			tempPosition += numberOfValidators[shardID]
		}
		candidateRandomShardID[candidate] = byte(shardID)
	}
	for _, candidate := range candidates {
		randomShardID := candidateRandomShardID[candidate]
		assignShardID := sortedShardIDs[n-1-m[randomShardID]]
		assignedCandidates[byte(assignShardID)] = append(assignedCandidates[byte(assignShardID)], candidate)
	}
	return assignedCandidates
}

// calculateCandidatePosition calculate reverse shardID for candidate
// randomPosition = sum(hash(candidate+rand)) % total, if randomPosition == 0 then randomPosition = 1
// randomPosition in range (1, total)
func calculateCandidatePosition(candidate string, rand int64, total int) (pos int) {
	seed := candidate + fmt.Sprintf("%v", rand)
	hash := common.HashB([]byte(seed))
	data := 0
	for _, v := range hash {
		data += int(v)
	}
	pos = data % total
	if pos == 0 {
		pos = 1
	}
	return pos
}

// sortShardIDByIncreaseOrder take an array and sort array, return sorted index of array
func sortShardIDByIncreaseOrder(arr []int) []byte {
	sortedIndex := []byte{}
	tempArr := []struct {
		shardID byte
		value   int
	}{}
	for i, v := range arr {
		tempArr = append(tempArr, struct {
			shardID byte
			value   int
		}{byte(i), v})
	}
	sort.Slice(tempArr, func(i, j int) bool {
		return tempArr[i].value < tempArr[j].value
	})
	for _, v := range tempArr {
		sortedIndex = append(sortedIndex, v.shardID)
	}
	return sortedIndex
}

func getShardIDPositionFromArray(arr []byte) map[byte]byte {
	m := make(map[byte]byte)
	for i, v := range arr {
		m[v] = byte(i)
	}
	return m
}

func (s *swapRuleV2) AssignOffset(lenShardSubstitute, lenCommittees, numberOfFixedValidators, minCommitteeSize int) int {
	assignPerShard := s.getSwapOutOffset(
		lenShardSubstitute,
		lenCommittees,
		numberOfFixedValidators,
		minCommitteeSize,
	)

	if assignPerShard == 0 && lenCommittees < MAX_SWAP_OR_ASSIGN_PERCENT {
		assignPerShard = 1
	}
	return assignPerShard
}

func (s *swapRuleV2) clone() SwapRule {
	return &swapRuleV2{}
}

func (s *swapRuleV2) Version() int {
	return swapRuleSlashingVersion
}
