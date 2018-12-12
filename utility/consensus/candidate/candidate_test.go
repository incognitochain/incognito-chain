package candidate

import (
	"crypto/sha256"
	"errors"
	"strconv"
)

//Assumption:
// validator and candidate public key encode as base58 string
func AssignValidator(candidates []string, rand int64) (map[byte][]string, error) {
	pendingValidators := make(map[byte][]string)
	for _, candidate := range candidates {
		seed := candidate + strconv.Itoa(int(rand))
		hash := sha256.Sum256([]byte(seed))
		pendingValidators[hash[len(hash)-1]] = append(pendingValidators[hash[len(hash)-1]], candidate)
	}
	return pendingValidators, nil
}

// consider these list as queue structure
// unqueue a number of validator out of currentValidators list
// enqueue a number of validator into currentValidators list <=> unqueue a number of validator out of pendingValidators list
// return value: #1 remaining pendingValidators, #2 new currentValidators
func SwapValidator(pendingValidators []string, currentValidators []string, offset int) ([]string, []string, error) {
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if offset > len(pendingValidators) {
		offset = len(pendingValidators)
	}
	// do nothing
	if offset == 0 {
		return pendingValidators, currentValidators, errors.New("No pending validator for swapping")
	}

	// unqueue validator with index from 0 to offset-1 from currentValidators list
	currentValidators = currentValidators[offset:]
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	tempValidators := pendingValidators[offset:]
	// save new pending validators list
	pendingValidators = pendingValidators[offset:]

	// enqueue new validator to the remaning of current validators list
	currentValidators = append(currentValidators, tempValidators...)
	return pendingValidators, currentValidators, nil
}
