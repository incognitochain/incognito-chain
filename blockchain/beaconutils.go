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

// AssignValidatorShard, param for better convenice than AssignValidator
func AssignValidatorShard(currentCandidates map[byte][]incognitokey.CommitteePublicKey, shardCandidates []incognitokey.CommitteePublicKey, rand int64, activeShards int) error {
	for _, candidate := range shardCandidates {
		candidateStr, _ := candidate.ToBase58()
		shardID := calculateCandidateShardID(candidateStr, rand, activeShards)
		currentCandidates[shardID] = append(currentCandidates[shardID], candidate)
	}
	return nil
}

func VerifyValidator(candidate string, rand int64, shardID byte, activeShards int) (bool, error) {
	res := calculateCandidateShardID(candidate, rand, activeShards)
	if shardID == res {
		return true, nil
	} else {
		return false, nil
	}
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

// consider these list as queue structure
// unqueue a number of validator out of currentValidators list
// enqueue a number of validator into currentValidators list <=> unqueue a number of validator out of pendingValidators list
// return value: #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator #5 error
func SwapValidator(pendingValidators []string, currentValidators []string, maxCommittee int, offset int) ([]string, []string, []string, []string, error) {
	if maxCommittee < 0 || offset < 0 {
		panic("committee can't be zero")
	}
	if offset == 0 {
		return []string{}, pendingValidators, currentValidators, []string{}, errors.New("can't not swap 0 validator")
	}
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if offset > len(pendingValidators) {
		offset = len(pendingValidators)
	}
	// if swap offset = 0 then do nothing
	if offset == 0 {
		return pendingValidators, currentValidators, []string{}, []string{}, errors.New("no pending validator for swapping")
	}
	if offset > maxCommittee {
		return pendingValidators, currentValidators, []string{}, []string{}, errors.New("trying to swap too many validator")
	}
	tempValidators := []string{}
	swapValidator := []string{}
	// if len(currentValidator) < maxCommittee then push validator until it is full
	if len(currentValidators) < maxCommittee {
		diff := maxCommittee - len(currentValidators)
		if diff >= offset {
			tempValidators = append(tempValidators, pendingValidators[:offset]...)
			currentValidators = append(currentValidators, tempValidators...)
			pendingValidators = pendingValidators[offset:]
			return pendingValidators, currentValidators, swapValidator, tempValidators, nil
		} else {
			offset -= diff
			tempValidators := append(tempValidators, pendingValidators[:diff]...)
			pendingValidators = pendingValidators[diff:]
			currentValidators = append(currentValidators, tempValidators...)
		}
	}
	// out pubkey: swapped out validator
	swapValidator = append(swapValidator, currentValidators[:offset]...)
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	currentValidators = currentValidators[offset:]
	// in pubkey: unqueue validator with index from 0 to offset-1 from pendingValidators list
	tempValidators = append(tempValidators, pendingValidators[:offset]...)
	// enqueue new validator to the remaning of current validators list
	currentValidators = append(currentValidators, pendingValidators[:offset]...)
	// save new pending validators list
	pendingValidators = pendingValidators[offset:]
	if len(currentValidators) > maxCommittee {
		panic("Length of current validator greater than max committee in Swap validator ")
	}
	return pendingValidators, currentValidators, swapValidator, tempValidators, nil
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
	for index, validator := range removedValidators {
		if strings.Compare(validators[index], validator) == 0 {
			validators = validators[1:]
		} else {
			// not found wanted validator
			return validators, errors.New("remove Validator with Wrong Format")
		}
	}
	return validators, nil
}

/*
	Shuffle Candidate:
		Candidate Value Concatenate with Random Number
		Then Hash and Obtain Hash Value
		Sort Hash Value Then Re-arrange Candidate corresponding to Hash Value
*/
func ShuffleCandidate(candidates []incognitokey.CommitteePublicKey, rand int64) ([]incognitokey.CommitteePublicKey, error) {
	fmt.Println("Beacon Process/Shuffle Candidate: Candidate Before Sort ", candidates)
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
	fmt.Println("Beacon Process/Shuffle Candidate: Candidate After Sort ", sortedCandidate)
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
		return snapshotRewardReceiver, fmt.Errorf("Failed to Clone Reward Receivers, expect %+v but get %+v", rewardReceiver, snapshotRewardReceiver)
	}
	return snapshotRewardReceiver, nil
}
