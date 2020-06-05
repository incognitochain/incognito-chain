package blockchain

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

//===================================Util for Beacon=============================
func GetStakingCandidate(beaconBlock BeaconBlock) ([]string, []string) {
	beacon := []string{}
	shard := []string{}
	beaconBlockBody := beaconBlock.Body
	for _, v := range beaconBlockBody.Instructions {
		if len(v) < 1 {
			continue
		}
		if v[0] == StakeAction && v[2] == "beacon" {
			beacon = strings.Split(v[1], ",")
		}
		if v[0] == StakeAction && v[2] == "shard" {
			shard = strings.Split(v[1], ",")
		}
	}

	return beacon, shard
}

// Assumption:
// validator and candidate public key encode as base58 string
// assume that candidates are already been checked
// Check validation of candidate in transaction
func AssignValidator(candidates []incognitokey.CommitteePublicKey, rand int64, activeShards int) (map[byte][]incognitokey.CommitteePublicKey, error) {
	pendingValidators := make(map[byte][]incognitokey.CommitteePublicKey)
	for _, candidate := range candidates {
		candidateStr, _ := candidate.ToBase58()
		shardID := calculateCandidateShardID(candidateStr, rand, activeShards)
		pendingValidators[shardID] = append(pendingValidators[shardID], candidate)
	}
	return pendingValidators, nil
}

// AssignValidatorShard, param for better convenience than AssignValidator
func AssignValidatorShard(currentShardPendingValidator map[byte][]incognitokey.CommitteePublicKey, shardCandidates []incognitokey.CommitteePublicKey, rand int64, activeShards int) {
	for _, candidate := range shardCandidates {
		candidateStr, _ := candidate.ToBase58()
		shardID := calculateCandidateShardID(candidateStr, rand, activeShards)
		currentShardPendingValidator[shardID] = append(currentShardPendingValidator[shardID], candidate)
	}
}

func VerifyValidator(candidate string, rand int64, shardID byte, activeShards int) (bool, error) {
	res := calculateCandidateShardID(candidate, rand, activeShards)
	if shardID == res {
		return true, nil
	} else {
		return false, nil
	}
}

/*
	Assign Candidates Into Shard Pending Validator List
	Each Shard Pending Validator List has a limit
	If a candidate is assigned into shard which Pending Validator List has reach its limit then candidate will get back into candidate list
	Otherwise, candidate will be converted to shard pending validator
	- return param #1: remain shard candidate (not assign yet)
	- return param #2: assigned candidate
*/
func assignShardCandidate(candidates []string, numberOfPendingValidator map[byte]int, rand int64, testnetAssignOffset int, activeShards int) ([]string, map[byte][]string) {
	assignedCandidates := make(map[byte][]string)
	remainShardCandidates := []string{}
	shuffledCandidate := shuffleShardCandidate(candidates, rand)
	for _, candidate := range shuffledCandidate {
		shardID := calculateCandidateShardID(candidate, rand, activeShards)
		if numberOfPendingValidator[shardID]+1 > testnetAssignOffset {
			remainShardCandidates = append(remainShardCandidates, candidate)
			continue
		} else {
			assignedCandidates[shardID] = append(assignedCandidates[shardID], candidate)
			numberOfPendingValidator[shardID] += 1
		}
	}
	return remainShardCandidates, assignedCandidates
}

/*
	Shuffle Position Of Shard Candidates in List with Random Number
*/
func shuffleShardCandidate(candidates []string, rand int64) []string {
	m := make(map[string]string)
	temp := []string{}
	shuffledCandidates := []string{}
	for _, candidate := range candidates {
		seed := strconv.Itoa(int(rand)) + candidate
		hash := common.HashH([]byte(seed)).String()
		m[hash] = candidate
		temp = append(temp, hash)
	}
	if len(m) != len(temp) {
		fmt.Println(candidates)
		panic("Failed To Shuffle Shard Candidate Before Assign to Shard")
	}
	sort.Strings(temp)
	for _, key := range temp {
		shuffledCandidates = append(shuffledCandidates, m[key])
	}
	if len(shuffledCandidates) != len(candidates) {
		panic("Failed To Shuffle Shard Candidate Before Assign to Shard")
	}
	return shuffledCandidates
}

// Formula ShardID: LSB[hash(candidatePubKey+randomNumber)]
// Last byte of hash(candidatePubKey+randomNumber)
func calculateCandidateShardID(candidate string, rand int64, activeShards int) (shardID byte) {
	seed := candidate + strconv.Itoa(int(rand))
	hash := common.HashB([]byte(seed))
	// fmt.Println("Candidate public key", candidate)
	// fmt.Println("Hash of candidate serialized pubkey and random number", hash)
	// fmt.Printf("\"%d\",\n", hash[len(hash)-1])
	// fmt.Println("Shard to be assign", hash[len(hash)-1])
	shardID = byte(int(hash[len(hash)-1]) % activeShards)
	Logger.log.Critical("calculateCandidateShardID/shardID", shardID)
	return shardID
}

func filterValidators(
	validators []string,
	producersBlackList map[string]uint8,
	isExistenceIncluded bool,
) []string {
	resultingValidators := []string{}
	for _, pv := range validators {
		_, found := producersBlackList[pv]
		if (found && isExistenceIncluded) || (!found && !isExistenceIncluded) {
			resultingValidators = append(resultingValidators, pv)
		}
	}
	return resultingValidators
}

// consider these list as queue structure
// unqueue a number of validator out of currentValidators list
// enqueue a number of validator into currentValidators list <=> unqueue a number of validator out of pendingValidators list
// return value: #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator #5 error
// func SwapValidator(pendingValidators []string, currentValidators []string, maxCommittee int, offset int) ([]string, []string, []string, []string, error) {
// 	// filter for pending validator, , it should not already exist in current validator list
// 	//filterPendingValidators := make([]string, len(pendingValidators))
// 	//copy(filterPendingValidators, pendingValidators)
// 	//for i, v := range filterPendingValidators {
// 	//	ok, _ := common.SliceExists(currentValidators, v) // item in pending list already exist in current list
// 	//	if ok {
// 	//		filterPendingValidators = append(filterPendingValidators[:i], filterPendingValidators[i+1:]...)
// 	//	}
// 	//}
// 	//if len(filterPendingValidators) < len(pendingValidators) {
// 	//	pendingValidators = filterPendingValidators
// 	//}
// 	// end

// 	if maxCommittee < 0 || offset < 0 {
// 		panic("committee can't be zero")
// 	}
// 	if offset == 0 {
// 		return []string{}, pendingValidators, currentValidators, []string{}, errors.New("can't not swap 0 validator")
// 	}

// }

func isBadProducer(badProducers []string, producer string) bool {
	for _, badProducer := range badProducers {
		if badProducer == producer {
			return true
		}
	}
	return false
}

func swap(
	badPendingValidators []string,
	goodPendingValidators []string,
	currentGoodProducers []string,
	currentBadProducers []string,
	maxCommittee int,
	offset int,
) ([]string, []string, []string, []string, error) {
	// if swap offset = 0 then do nothing
	if offset == 0 {
		// return pendingValidators, currentGoodProducers, currentBadProducers, []string{}, errors.New("no pending validator for swapping")
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, currentBadProducers, []string{}, nil
	}
	if offset > maxCommittee {
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, currentBadProducers, []string{}, errors.New("try to swap too many validators")
	}
	tempValidators := []string{}
	swapValidator := currentBadProducers
	diff := maxCommittee - len(currentGoodProducers)
	if diff >= offset {
		tempValidators = append(tempValidators, goodPendingValidators[:offset]...)
		currentGoodProducers = append(currentGoodProducers, tempValidators...)
		goodPendingValidators = goodPendingValidators[offset:]
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, swapValidator, tempValidators, nil
	}
	offset -= diff
	tempValidators = append(tempValidators, goodPendingValidators[:diff]...)
	goodPendingValidators = goodPendingValidators[diff:]
	currentGoodProducers = append(currentGoodProducers, tempValidators...)

	// out pubkey: swapped out validator
	swapValidator = append(swapValidator, currentGoodProducers[:offset]...)
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	currentGoodProducers = currentGoodProducers[offset:]
	// in pubkey: unqueue validator with index from 0 to offset-1 from pendingValidators list
	tempValidators = append(tempValidators, goodPendingValidators[:offset]...)
	// enqueue new validator to the remaning of current validators list
	currentGoodProducers = append(currentGoodProducers, goodPendingValidators[:offset]...)
	// save new pending validators list
	goodPendingValidators = goodPendingValidators[offset:]
	return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, swapValidator, tempValidators, nil
}

// consider these list as queue structure
// unqueue a number of validator out of currentValidators list
// enqueue a number of validator into currentValidators list <=> unqueue a number of validator out of pendingValidators list
// return value: #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator #5 error
func SwapValidator(
	pendingValidators []string,
	currentValidators []string,
	maxCommittee int,
	minCommittee int,
	offset int,
	producersBlackList map[string]uint8,
	swapOffset int,
) ([]string, []string, []string, []string, error) {
	goodPendingValidators := filterValidators(pendingValidators, producersBlackList, false)
	badPendingValidators := filterValidators(pendingValidators, producersBlackList, true)
	currentBadProducers := filterValidators(currentValidators, producersBlackList, true)
	currentGoodProducers := filterValidators(currentValidators, producersBlackList, false)
	goodPendingValidatorsLen := len(goodPendingValidators)
	currentGoodProducersLen := len(currentGoodProducers)

	if currentGoodProducersLen >= minCommittee {
		if currentGoodProducersLen == maxCommittee {
			offset = swapOffset
		}
		if offset > goodPendingValidatorsLen {
			offset = goodPendingValidatorsLen
		}
		return swap(badPendingValidators, goodPendingValidators, currentGoodProducers, currentBadProducers, maxCommittee, offset)
	}

	minProducersNeeded := minCommittee - currentGoodProducersLen
	if len(pendingValidators) >= minProducersNeeded {
		if offset < minProducersNeeded {
			offset = minProducersNeeded
		} else if offset > goodPendingValidatorsLen {
			offset = goodPendingValidatorsLen
		}
		return swap(badPendingValidators, goodPendingValidators, currentGoodProducers, currentBadProducers, maxCommittee, offset)
	}

	producersNumCouldBeSwapped := len(goodPendingValidators) + len(currentValidators) - minCommittee
	swappedProducers := []string{}
	remainingProducers := []string{}
	for _, producer := range currentValidators {
		if isBadProducer(currentBadProducers, producer) && len(swappedProducers) < producersNumCouldBeSwapped {
			swappedProducers = append(swappedProducers, producer)
			continue
		}
		remainingProducers = append(remainingProducers, producer)
	}
	newProducers := append(remainingProducers, goodPendingValidators...)
	return badPendingValidators, newProducers, swappedProducers, goodPendingValidators, nil
}

// return: #param1: validator list after remove
// in parameter: #param1: list of full validator
// in parameter: #param2: list of removed validator
// removed validators list must be a subset of full validator list and it must be first in the list
func RemoveValidator(validators []string, removedValidators []string) ([]string, error) {
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if len(removedValidators) > len(validators) {
		return validators, errors.New("trying to remove too many validators")
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

/*
	Shuffle Candidate:
		Candidate Value Concatenate with Random Number
		Then Hash and Obtain Hash Value
		Sort Hash Value Then Re-arrange Candidate corresponding to Hash Value
*/
func ShuffleCandidate(candidates []incognitokey.CommitteePublicKey, rand int64) ([]incognitokey.CommitteePublicKey, error) {
	Logger.log.Debug("Beacon Process/Shuffle Candidate: Candidate Before Sort ", candidates)
	hashes := []string{}
	m := make(map[string]incognitokey.CommitteePublicKey)
	sortedCandidate := []incognitokey.CommitteePublicKey{}
	for _, candidate := range candidates {
		candidateStr, _ := candidate.ToBase58()
		seed := candidateStr + strconv.Itoa(int(rand))
		hash := common.HashB([]byte(seed))
		hashes = append(hashes, string(hash[:32]))
		m[string(hash[:32])] = candidate
	}
	sort.Strings(hashes)
	for _, hash := range hashes {
		sortedCandidate = append(sortedCandidate, m[hash])
	}
	Logger.log.Debug("Beacon Process/Shuffle Candidate: Candidate After Sort ", sortedCandidate)
	return sortedCandidate, nil
}

/*
	Kick a list of candidate out of current validators list
	Candidates will be eliminated as the list order (from 0 index to last index)
	A candidate will be click out of list if it match those condition:
		- candidate pubkey found in current validators list
		- size of current validator list is greater or equal to min committess size
	Return params:
	#1 kickedValidator, #2 remain candidates (not kick yet), #3 new current validator list
*/
func kickValidatorByPubkeyList(candidates []string, currentValidators []string, minCommitteeSize int) ([]string, []string, []string) {
	removedCandidates := []string{}
	remainedCandidates := []string{}
	remainedIndex := 0
	for index, candidate := range candidates {
		remainedIndex = index
		if len(currentValidators) == minCommitteeSize {
			break
		}
		if index := common.IndexOfStr(candidate, currentValidators); index < 0 {
			remainedCandidates = append(remainedCandidates, candidate)
			continue
		} else {
			removedCandidates = append(removedCandidates, candidate)
			currentValidators = append(currentValidators[:index], currentValidators[index+1:]...)
		}
	}
	if remainedIndex < len(candidates)-1 {
		remainedCandidates = append(remainedCandidates, candidates[remainedIndex:]...)
	}
	return removedCandidates, remainedCandidates, currentValidators
}
func kickValidatorByPubkey(candidate string, currentValidators []string, minCommitteeSize int) (bool, []string) {
	if index := common.IndexOfStr(candidate, currentValidators); index < 0 {
		return false, currentValidators
	} else {
		currentValidators = append(currentValidators[:index], currentValidators[index+1:]...)
		return true, currentValidators
	}
}
func getStakeValidatorArrayString(v []string) ([]string, []string) {
	beacon := []string{}
	shard := []string{}
	if len(v) > 0 {
		if v[0] == StakeAction && v[2] == "beacon" {
			beacon = strings.Split(v[1], ",")
		}
		if v[0] == StakeAction && v[2] == "shard" {
			shard = strings.Split(v[1], ",")
		}
	}
	return beacon, shard
}
func snapshotCommittee(beaconCommittee []incognitokey.CommitteePublicKey, allShardCommittee map[byte][]incognitokey.CommitteePublicKey) ([]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, error) {
	snapshotBeaconCommittee := []incognitokey.CommitteePublicKey{}
	snapshotAllShardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for _, committee := range beaconCommittee {
		snapshotBeaconCommittee = append(snapshotBeaconCommittee, committee)
	}
	for shardID, shardCommittee := range allShardCommittee {
		clonedShardCommittee := []incognitokey.CommitteePublicKey{}
		snapshotAllShardCommittee[shardID] = []incognitokey.CommitteePublicKey{}
		for _, committee := range shardCommittee {
			clonedShardCommittee = append(clonedShardCommittee, committee)
		}
		snapshotAllShardCommittee[shardID] = clonedShardCommittee
	}
	if !reflect.DeepEqual(beaconCommittee, snapshotBeaconCommittee) {
		return []incognitokey.CommitteePublicKey{}, nil, fmt.Errorf("Failed To Clone Beacon Committee, expect %+v but get %+v", beaconCommittee, snapshotBeaconCommittee)
	}
	if !reflect.DeepEqual(allShardCommittee, snapshotAllShardCommittee) {
		return []incognitokey.CommitteePublicKey{}, nil, fmt.Errorf("Failed To Clone Beacon Committee, expect %+v but get %+v", allShardCommittee, snapshotAllShardCommittee)
	}
	return snapshotBeaconCommittee, snapshotAllShardCommittee, nil
}
func snapshotRewardReceiver(rewardReceiver map[string]string) (map[string]string, error) {
	snapshotRewardReceiver := make(map[string]string)
	for k, v := range rewardReceiver {
		snapshotRewardReceiver[k] = v
	}
	if !reflect.DeepEqual(snapshotRewardReceiver, rewardReceiver) {
		return snapshotRewardReceiver, fmt.Errorf("Failed to Clone Reward Rewards, expect %+v but get %+v", rewardReceiver, snapshotRewardReceiver)
	}
	return snapshotRewardReceiver, nil
}
