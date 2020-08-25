package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
	"reflect"
	"sort"
)

// createSwapInstructionV2 create swap instruction and new substitutes list
// return params
// #1: swap instruction
// #2: new substitute list
// #3: error
func createSwapInstructionV2(
	shardID byte,
	substitutes []string,
	committees []string,
	maxSwapOffset int,
	numberOfRound map[string]int,
) (*instruction.SwapInstruction, []string, error) {
	newSubstitutes, _, swappedOutCommittees, swappInCommittees, err := swapV2(
		substitutes,
		committees,
		maxSwapOffset,
		numberOfRound,
	)
	if err != nil {
		return &instruction.SwapInstruction{}, []string{}, err
	}
	swapInstruction := instruction.NewSwapInstructionWithValue(
		swappInCommittees,
		swappedOutCommittees,
		int(shardID),
	)
	return swapInstruction, newSubstitutes, nil
}

// removeValidatorV2 remove validator and return removed list
// return: #param1: validator list after remove
// in parameter: #param1: list of full validator
// in parameter: #param2: list of removed validator
// removed validators list must be a subset of full validator list and it must be first in the list
func removeValidatorV2(validators []string, removedValidators []string) ([]string, error) {
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	remainingValidators := []string{}
	if len(removedValidators) > len(validators) {
		return remainingValidators, fmt.Errorf("removed validator length %+v, bigger than current validator length %+v", removedValidators, validators)
	}
	if !reflect.DeepEqual(validators[:len(removedValidators)], removedValidators) {
		return remainingValidators, fmt.Errorf("current validator %+v and removed validator %+v is not compatible", validators, removedValidators)
	}
	remainingValidators = validators[len(removedValidators):]
	return remainingValidators, nil
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
	maxSwapOffSet int,
	numberOfRound map[string]int,
) ([]string, []string, []string, []string, error) {
	// if swap offset = 0 then do nothing
	swapOffset := (len(substitutes) + len(committees)) / MAX_SWAP_OR_ASSIGN_PERCENT
	Logger.log.Info("Swap Rule V2, Swap Offset ", swapOffset)
	if swapOffset == 0 {
		// return pendingValidators, currentGoodProducers, currentBadProducers, []string{}, errors.New("no pending validator for swapping")
		return committees, substitutes, []string{}, []string{}, nil
	}
	// swap offset must be less than or equal to maxSwapOffSet
	// maxSwapOffSet mainnet is 10 => swapOffset is <= 10
	if swapOffset > maxSwapOffSet {
		swapOffset = maxSwapOffSet
	}
	// swapOffset must be less than or equal to substitutes length
	if swapOffset > len(substitutes) {
		swapOffset = len(substitutes)
	}
	vacantSlot := maxSwapOffSet - len(committees)
	// vacantSlot is greater than number of substitutes
	if vacantSlot >= swapOffset {
		swappedInCommittees := substitutes[:swapOffset]
		swappedOutCommittees := []string{}
		committees = append(committees, swappedInCommittees...)
		substitutes = substitutes[swapOffset:]
		return committees, substitutes, swappedOutCommittees, swappedInCommittees, nil
	} else {
		// number of substitutes is greater than vacantSlot
		// push substitutes into vacant slot in committee list until full
		swappedInCommittees := substitutes[:vacantSlot]
		substitutes = substitutes[vacantSlot:]
		committees = append(committees, swappedInCommittees...)

		swapOffsetAfterFillVacantSlot := swapOffset - vacantSlot
		// swapped out committees: record swapped out committees
		tryToSwappedOutCommittees := committees[:swapOffsetAfterFillVacantSlot]
		swappedOutCommittees := []string{}
		backToSubstitutes := []string{}
		for _, tryToSwappedOutCommittee := range tryToSwappedOutCommittees {
			if numberOfRound[tryToSwappedOutCommittee] >= MAX_NUMBER_OF_ROUND {
				swappedOutCommittees = append(swappedOutCommittees, tryToSwappedOutCommittee)
			} else {
				backToSubstitutes = append(backToSubstitutes, tryToSwappedOutCommittee)
			}
		}
		// un-queue committees:  start from index 0 to swapOffset - 1
		committees = committees[swapOffset:]
		// swapped in: (continue) to un-queue substitute from index from 0 to swapOffsetAfterFillVacantSlot -1
		swappedInCommittees = append(swappedInCommittees, substitutes[:swapOffsetAfterFillVacantSlot]...)
		// en-queue new validator: from substitute list to committee list
		committees = append(committees, substitutes[:swapOffsetAfterFillVacantSlot]...)
		// un-queue substitutes: start from index 0 to swapOffsetAfterFillVacantSlot - 1
		substitutes = substitutes[swapOffsetAfterFillVacantSlot:]
		// en-queue some swapped out committees (if satisfy condition above)
		substitutes = append(substitutes, backToSubstitutes...)
		return substitutes, committees, swappedOutCommittees, swappedInCommittees, nil
	}
}

// assignShardCandidateV2 assign candidates into shard pool with random number
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
