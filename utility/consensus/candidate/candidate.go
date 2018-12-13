package candidate

import (
	"crypto/sha256"
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/blockchain"
)

func UpdateBeaconBestState(beaconBestState *blockchain.BestStateBeacon, newBlock *blockchain.BlockV2) (*blockchain.BestStateBeacon, error) {
	// TODO:
	// update BestShardHash, BestBlock, BestBlockHash
	// unassign -> remove out the candidate

	if beaconBestState.BeaconHeight%200 == 1 {
		newBeaconNode, newShardNode := GetStakingCandidate(newBlock)
		beaconBestState.UnassignBeaconCandidate = append(beaconBestState.UnassignBeaconCandidate, newBeaconNode...)
		beaconBestState.UnassignShardCandidate = append(beaconBestState.UnassignShardCandidate, newShardNode...)
		//TODO: assign unAssignCandidate to assignCandidate	& clear UnassignShardCandidate
		// ShardCandidate         map[byte][]string
		// ShardPendingCandidate  map[byte][]string
		// BeaconCandidate        []string
		// BeaconPendingCandidate []string

		/// Shuffle candidate for shard
		// assign UnassignShardCandidate to ShardPendingCandidate with CurrentRandom this shard
		err := AssignValidatorShard(beaconBestState.ShardPendingCandidate, beaconBestState.UnassignShardCandidate, beaconBestState.CurrentRandomNumber)
		// reset beaconBestState.UnassignShardCandidate
		beaconBestState.UnassignShardCandidate = []string{}

		if err != nil {
			return beaconBestState, err
		}

		// swap 3 validators each time
		const offset = int(3)
		for i := 0; i < 256; i++ {
			shardID := byte(i)
			//swap validator for each shard
			beaconBestState.ShardPendingCandidate[shardID], beaconBestState.ShardCandidate[shardID], err = SwapValidator(beaconBestState.ShardPendingCandidate[shardID], beaconBestState.ShardCandidate[shardID], offset)
			if err != nil {
				return beaconBestState, err
			}
		}
		// ShuffleCandidate
		shuffleBeaconCandidate, err := ShuffleCandidate(beaconBestState.UnassignBeaconCandidate, beaconBestState.CurrentRandomNumber)
		if err != nil {
			return beaconBestState, err
		}
		// append new candidate to pending validator in beacon
		beaconBestState.BeaconPendingCandidate = append(beaconBestState.BeaconPendingCandidate, shuffleBeaconCandidate...)
		// reset UnassignBeaconCandidate
		beaconBestState.UnassignBeaconCandidate = []string{}
		//swap validator in beacon
		beaconBestState.BeaconPendingCandidate, beaconBestState.BeaconCandidate, err = SwapValidator(beaconBestState.BeaconPendingCandidate, beaconBestState.BeaconCandidate, offset)
		if err != nil {
			return beaconBestState, err
		}
		// update random number for new epoch
		beaconBestState.CurrentRandomNumber = beaconBestState.NextRandomNumber
	} else {
		// GetStakingCandidate -> UnassignCandidate
		newBeaconNode, newShardNode := GetStakingCandidate(newBlock)
		beaconBestState.UnassignBeaconCandidate = append(beaconBestState.UnassignBeaconCandidate, newBeaconNode...)
		beaconBestState.UnassignShardCandidate = append(beaconBestState.UnassignShardCandidate, newShardNode...)
	}
	// TODO: Param "set" "del"

	return beaconBestState, nil
}

func GetStakingCandidate(beaconBlock *blockchain.BlockV2) (beacon []string, shard []string) {
	if beaconBlock.Type == "beacon" {
		beaconBlockBody := beaconBlock.Body.(*blockchain.BeaconBlockBody)
		for _, v := range beaconBlockBody.Instructions {
			if v[0] == "assign" && v[2] == "beacon" {
				beacon = strings.Split(v[1], ",")
			}
			if v[0] == "assign" && v[2] == "shard" {
				shard = strings.Split(v[1], ",")
			}
		}
	} else {
		panic("GetStakingCandidate not from beacon block")
	}
	return beacon, shard
}

// Assumption:
// validator and candidate public key encode as base58 string
// assume that candidates are already been checked
// Check validation of candidate in transaction
func AssignValidator(candidates []string, rand int64) (map[byte][]string, error) {
	pendingValidators := make(map[byte][]string)
	for _, candidate := range candidates {
		shardID := calculateHash(candidate, rand)
		pendingValidators[shardID] = append(pendingValidators[shardID], candidate)
	}
	return pendingValidators, nil
}

// AssignValidatorShard, param for better convenice than AssignValidator
func AssignValidatorShard(currentCandidates map[byte][]string, shardCandidates []string, rand int64) error {
	for _, candidate := range shardCandidates {
		shardID := calculateHash(candidate, rand)
		currentCandidates[shardID] = append(currentCandidates[shardID], candidate)
	}
	return nil
}

func VerifyValidator(candidate string, rand int64, shardID byte) (bool, error) {
	res := calculateHash(candidate, rand)
	if shardID == res {
		return true, nil
	} else {
		return false, nil
	}
}

// Formula ShardID: LSB[hash(candidatePubKey+randomNumber)]
// Last byte of hash(candidatePubKey+randomNumber)
func calculateHash(candidate string, rand int64) (shardID byte) {
	seed := candidate + strconv.Itoa(int(rand))
	hash := sha256.Sum256([]byte(seed))
	// fmt.Println("Candidate public key", candidate)
	// fmt.Println("Hash of candidate serialized pubkey and random number", hash)
	// fmt.Printf("\"%d\",\n", hash[len(hash)-1])
	// fmt.Println("Shard to be assign", hash[len(hash)-1])
	shardID = hash[len(hash)-1]
	return shardID
}

// consider these list as queue structure
// unqueue a number of validator out of currentValidators list
// enqueue a number of validator into currentValidators list <=> unqueue a number of validator out of pendingValidators list
// return value: #1 remaining pendingValidators, #2 new currentValidators
func SwapValidator(pendingValidators []string, currentValidators []string, offset int) ([]string, []string, error) {
	if offset == 0 {
		return pendingValidators, currentValidators, errors.New("Can't not swap 0 validator")
	}
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if offset > len(pendingValidators) {
		offset = len(pendingValidators)
	}
	// do nothing
	if offset == 0 {
		return pendingValidators, currentValidators, errors.New("No pending validator for swapping")
	}
	if offset > len(currentValidators) {
		return pendingValidators, currentValidators, errors.New("Trying to swap too many validator")
	}
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	currentValidators = currentValidators[offset:]
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	tempValidators := pendingValidators[:offset]
	// save new pending validators list
	pendingValidators = pendingValidators[offset:]

	// enqueue new validator to the remaning of current validators list
	currentValidators = append(currentValidators, tempValidators...)
	return pendingValidators, currentValidators, nil
}

func ShuffleCandidate(candidates []string, rand int64) ([]string, error) {
	hashes := []string{}
	m := make(map[string]string)
	sortedCandidate := []string{}
	for _, candidate := range candidates {
		seed := candidate + strconv.Itoa(int(rand))
		hash := sha256.Sum256([]byte(seed))
		hashes = append(hashes, string(hash[:32]))
		m[string(hash[:32])] = candidate
	}
	sort.Strings(hashes)
	for _, candidate := range hashes {
		sortedCandidate = append(sortedCandidate, candidate)
	}
	return hashes, nil
}
