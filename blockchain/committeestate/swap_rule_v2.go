package committeestate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
)

//TODO: @hung
func createSwapInstructionV2(
	substitutes []string,
	committees []string,
	maxCommitteeSize int,
	percent int,
) ([]string, error) {
	return []string{}, nil
}

// removeValidatorV2 remove validator and return removed list
// return: #param1: validator list after remove
// in parameter: #param1: list of full validator
// in parameter: #param2: list of removed validator
// removed validators list must be a subset of full validator list and it must be first in the list
func removeValidatorV2(validators []string, removedValidators []string) ([]string, error) {
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if len(removedValidators) > len(validators) {
		return validators, fmt.Errorf("removed validator length %+v, bigger than current validator length %+v", removedValidators, validators)
	}
	remainingValidators := []string{}
	for _, validator := range validators {
		isRemoved := false
		for _, removedValidator := range removedValidators {
			if strings.Compare(validator, removedValidator) == 0 {
				isRemoved = true
			}
		}
		if !isRemoved {
			remainingValidators = append(remainingValidators, validator)
		}
	}
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
	maxCommitteeSize int,
	percent int,
	numberOfRound map[string]int,
) ([]string, []string, []string, []string, error) {
	// if swap offset = 0 then do nothing
	swapOffset := (len(substitutes) + len(committees)) * 100 / percent
	Logger.log.Info("Swap Rule V2, Swap Offset ", swapOffset)
	if swapOffset == 0 {
		// return pendingValidators, currentGoodProducers, currentBadProducers, []string{}, errors.New("no pending validator for swapping")
		return committees, substitutes, []string{}, []string{}, nil
	}
	// swap offset must be less than or equal to maxCommitteeSize
	if swapOffset > maxCommitteeSize {
		swapOffset = maxCommitteeSize
	}
	// swapOffset must be less than or equal to substitutes length
	if swapOffset > len(substitutes) {
		swapOffset = len(substitutes)
	}
	vacantSlot := maxCommitteeSize - len(committees)
	if vacantSlot >= swapOffset {
		swappedInCommittees := substitutes[:swapOffset]
		swappedOutCommittees := []string{}
		committees = append(committees, swappedInCommittees...)
		substitutes = substitutes[swapOffset:]
		return committees, substitutes, swappedOutCommittees, swappedInCommittees, nil
	} else {
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
