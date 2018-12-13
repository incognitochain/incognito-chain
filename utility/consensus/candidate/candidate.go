package candidate

import (
	"crypto/sha256"
	"errors"
	"github.com/ninjadotorg/constant/blockchain"
	"strconv"
	"strings"
)

func UpdateBeaconBestState(beaconBestState *blockchain.BestStateBeacon, newBlock *blockchain.BlockV2) *blockchain.BestStateBeacon {
	// TODO:
	// update BestShardHash, BestBlock, BestBlockHash
	// unassign -> remove out the candidate

	if beaconBestState.BeaconHeight%200 == 1 {
		newBeaconNode, newShardNode := GetStakingCandidate(newBlock)
		beaconBestState.UnassignBeaconCandidate = append(beaconBestState.UnassignBeaconCandidate, newBeaconNode...)
		beaconBestState.UnassignShardCandidate = append(beaconBestState.UnassignShardCandidate, newShardNode...)
		//TODO: assign unAssignCandidate to assignCandidate	& clear UnassignShardCandidate

		// update random number for new epoch
		beaconBestState.CurrentRandomNumber = beaconBestState.NextRandomNumber
	} else {
		// GetStakingCandidate -> UnassignCandidate
		newBeaconNode, newShardNode := GetStakingCandidate(newBlock)
		beaconBestState.UnassignBeaconCandidate = append(beaconBestState.UnassignBeaconCandidate, newBeaconNode...)
		beaconBestState.UnassignShardCandidate = append(beaconBestState.UnassignShardCandidate, newShardNode...)
	}
	// update param
	instructions := newBlock.Body.(*blockchain.BeaconBlockBody).Instructions
	for _, l := range instructions {
		if l[0] == "set" {
			beaconBestState.Params[l[1]] = l[2]
		}
		if l[0] == "del" {
			delete(beaconBestState.Params, l[1])
		}
	}

	return beaconBestState
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
