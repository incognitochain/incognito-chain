package committeestate

import (
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
)

// createSwapShardInstructionV2 create swap instruction and new substitutes list
// return params
// #1: swap instruction
// #2: new substitute list
// #3: error
func createSwapShardInstructionV2(
	shardID byte,
	substitutes, committees []string,
	maxCommitteeSize int,
	numberOfRound map[string]int,
	typeIns int,
	height uint64,
	numberOfFixedValidators uint64,
) (*instruction.SwapShardInstruction, []string, error) {
	_, newSubstitutes, swappedOutCommittees, swapInCommittees, err := swapCommitteesV2(
		committees,
		substitutes,
		maxCommitteeSize,
		numberOfFixedValidators,
	)

	if err != nil {
		return &instruction.SwapShardInstruction{}, []string{}, err
	}
	swapShardInstruction := instruction.NewSwapShardInstructionWithValue(
		swapInCommittees,
		swappedOutCommittees,
		int(shardID),
		typeIns,
		height,
	)

	return swapShardInstruction, newSubstitutes, nil
}

// removeValidatorV2 remove validator and return removed list
// return validator list after remove
// parameter:
// #1: list of full validator
// #2: list of removed validator
func removeValidatorV2(validators []string, removedValidators []string) ([]string, error) {
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	for _, removedValidator := range removedValidators {
		found := false
		index := 0
		for i, validator := range validators {
			if validator == removedValidator {
				found = true
				index = i
				break
			}
		}
		if found {
			validators = append(validators[:index], validators[index+1:]...)
		} else {
			return []string{}, fmt.Errorf("Try to remove validator %+v but not found in list %+v", removedValidator, validators)
		}
	}
	return validators, nil
}

//swapCommitteesV2
// Input:
// committees list, subtitutes list, max committee size, number of fixed validators
// Output:
// #1 new committees list
// #2 remained substitutes list
// #3 swapped out committees list (removed from committees list
// #4 swapped in committees list (new committees from substitutes list)
func swapCommitteesV2(
	committees []string,
	substitutes []string,
	maxCommitteeSize int,
	numberOfFixedValidators uint64,
) ([]string, []string, []string, []string, error) {
	swappedInCommittees := []string{}
	swappedOutCommittees := []string{}
	// if swap offset = 0 then do nothing

	rootCommittees := committees
	committees = committees[numberOfFixedValidators:]
	swapOffset := (len(substitutes) + len(committees)) / MAX_SWAP_OR_ASSIGN_PERCENT

	Logger.log.Info("Swap Rule V2, Swap Offset ", swapOffset)
	if swapOffset == 0 {
		return rootCommittees, substitutes, swappedOutCommittees, swappedInCommittees, nil
	}

	// swap offset must be less than or equal to maxCommitteeSize
	// maxCommitteeSize mainnet is 10 => swapOffset is <= 10
	if swapOffset > maxCommitteeSize {
		swapOffset = maxCommitteeSize
	}
	// swapOffset must be less than or equal to substitutes length
	if swapOffset > len(substitutes) {
		swapOffset = len(substitutes)
	}
	// vacantSlot must be equal to or greater than 0
	vacantSlot := maxCommitteeSize - len(committees)
	fixedValidators := committees[:numberOfFixedValidators]

	if vacantSlot >= swapOffset {
		// vacantSlot is greater than number of swap offset
		swappedInCommittees = append(swappedInCommittees, substitutes[:swapOffset]...)
		committees = append(committees, swappedInCommittees...)
		substitutes = substitutes[swapOffset:]
	} else {
		// vacantSlot is less than number of swap offset
		// get new committee from substitute list for push in only
		swappedInCommittees = append(swappedInCommittees, substitutes[:vacantSlot]...)
		// un-queue substitutes if vacant slot > 0
		substitutes = substitutes[vacantSlot:]

		swapOffsetAfterFillVacantSlot := swapOffset - vacantSlot

		// swapped out committees: record swapped out committees
		swappedOutCommittees = append(swappedOutCommittees, committees[:swapOffsetAfterFillVacantSlot]...)
		// un-queue committees:  start from index 0 to swapOffsetAfterFillVacantSlot - 1
		committees = committees[swapOffsetAfterFillVacantSlot:]
		// swapped in: (continue) to un-queue substitute from index from 0 to swapOffsetAfterFillVacantSlot -1
		swappedInCommittees = append(swappedInCommittees, substitutes[:swapOffsetAfterFillVacantSlot]...)
		// en-queue new validator: from substitute list to committee list
		committees = append(committees, swappedInCommittees...)
		// un-queue substitutes: start from index 0 to swapOffsetAfterFillVacantSlot - 1
		substitutes = substitutes[swapOffsetAfterFillVacantSlot:]
	}
	committees = append(fixedValidators, committees...)

	return committees, substitutes, swappedOutCommittees, swappedInCommittees, nil
}

// swapV2 swap substitute into committee
// return params
// #2 remained substitutes list
// #1 new committees list
// #3 swapped out committees list (removed from committees list
// #4 swapped in committees list (new committees from substitutes list)
func swapV2(
	substitutes []string,
	committees []string,
	maxCommitteeSize int,
) ([]string, []string, []string, []string, error) {
	// if swap offset = 0 then do nothing
	swapOffset := (len(substitutes) + len(committees)) / MAX_SWAP_OR_ASSIGN_PERCENT
	Logger.log.Info("Swap Rule V2, Swap Offset ", swapOffset)
	if swapOffset == 0 {
		// return pendingValidators, currentGoodProducers, currentBadProducers, []string{}, errors.New("no pending validator for swapping")
		return committees, substitutes, []string{}, []string{}, nil
	}
	// swap offset must be less than or equal to maxCommitteeSize
	// maxCommitteeSize mainnet is 10 => swapOffset is <= 10
	if swapOffset > maxCommitteeSize {
		swapOffset = maxCommitteeSize
	}
	// swapOffset must be less than or equal to substitutes length
	if swapOffset > len(substitutes) {
		swapOffset = len(substitutes)
	}
	swappedInCommittees := []string{}
	swappedOutCommittees := []string{}
	// vacantSlot must be equal to or greater than 0
	vacantSlot := maxCommitteeSize - len(committees)
	if vacantSlot >= swapOffset {
		// vacantSlot is greater than number of swap offset
		swappedInCommittees = append(swappedInCommittees, substitutes[:swapOffset]...)
		committees = append(committees, swappedInCommittees...)
		substitutes = substitutes[swapOffset:]
		return committees, substitutes, swappedOutCommittees, swappedInCommittees, nil
	} else {
		// vacantSlot is less than number of swap offset
		// get new committee from substitute list for push in only
		swappedInCommittees = append(swappedInCommittees, substitutes[:vacantSlot]...)
		// un-queue substitutes if vacant slot > 0
		substitutes = substitutes[vacantSlot:]

		swapOffsetAfterFillVacantSlot := swapOffset - vacantSlot

		// swapped out committees: record swapped out committees
		swappedOutCommittees = append(swappedOutCommittees, committees[:swapOffsetAfterFillVacantSlot]...)
		// un-queue committees:  start from index 0 to swapOffsetAfterFillVacantSlot - 1
		committees = committees[swapOffsetAfterFillVacantSlot:]
		// swapped in: (continue) to un-queue substitute from index from 0 to swapOffsetAfterFillVacantSlot -1
		swappedInCommittees = append(swappedInCommittees, substitutes[:swapOffsetAfterFillVacantSlot]...)
		// en-queue new validator: from substitute list to committee list
		committees = append(committees, swappedInCommittees...)
		// un-queue substitutes: start from index 0 to swapOffsetAfterFillVacantSlot - 1
		substitutes = substitutes[swapOffsetAfterFillVacantSlot:]

		return substitutes, committees, swappedOutCommittees, swappedInCommittees, nil
	}
	//TODO: @hung Return in 2 branches is different variables (substitutes, committees)
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
	for candidate, randomShardID := range candidateRandomShardID {
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

//getLatestHeightByShardsState get highest height from list shard states
// return 0 for empty list states
// return uint > 0 heights for fork cases
func getLatestHeightByShardsState(states []types.ShardState) uint64 {
	if len(states) == 0 {
		return 0
	}
	max := states[0].Height
	for _, v := range states {
		if v.Height > max {
			max = v.Height
		}
	}
	return max
}
