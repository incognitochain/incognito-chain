package candidate

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/big0t/constant-chain/blockchain"
	"github.com/big0t/constant-chain/common"
)

type ByShardIDAndBlockHeight []blockchain.BlockV2

func (a ByShardIDAndBlockHeight) Len() int { return len(a) }
func (a ByShardIDAndBlockHeight) Less(i, j int) bool {
	shardIDi := a[i].Header.(*blockchain.BlockHeaderShard).ShardID
	shardIDj := a[j].Header.(*blockchain.BlockHeaderShard).ShardID
	heightI := a[i].Header.(*blockchain.BlockHeaderShard).Height
	heightJ := a[j].Header.(*blockchain.BlockHeaderShard).Height

	if shardIDi < shardIDj {
		return true
	} else if shardIDi > shardIDj {
		return false
	} else {
		if heightI <= heightJ {
			return true
		}
		return false
	}
}
func (a ByShardIDAndBlockHeight) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func BuildBeaconBlock(beaconBestState *blockchain.BestStateBeacon, newShardBlock []blockchain.BlockV2) *blockchain.BlockV2 {
	// Create new unsigned beacon block (UBB)
	var newUnsignedBeaconBlock = &blockchain.BlockV2{
		Type: "beacon",
	}

	sort.Sort(ByShardIDAndBlockHeight(newShardBlock))

	var tempBeaconShardState [][]common.Hash
	var blocksInShardWithIdx map[int][]blockchain.BlockV2

	for shardBlkIdx, shardBlk := range newShardBlock {
		shardID := shardBlk.Header.(*blockchain.BlockHeaderShard).ShardID
		blocksInShardWithIdx[int(shardID)] = append(blocksInShardWithIdx[int(shardID)], shardBlk)
	}

	for i1, v := range blocksInShardWithIdx {
		bestShardHash := beaconBestState.BestShardHash[i1]

		if v[0].Header.(*blockchain.BlockHeaderShard).PrevBlockHash != bestShardHash {
			continue
		}

		for i2, blk := range v {
			if i2 == len(v)-1 {
				break
			}
			if *blk.HashFinal() != v[i2+1].Header.(*blockchain.BlockHeaderShard).PrevBlockHash {
				blocksInShardWithIdx[i1] = blocksInShardWithIdx[i1][:i2+1]
				break
			}
		}
	}

	for idx, v := range blocksInShardWithIdx {
		for _, u := range v {
			tempBeaconShardState[idx] = append(tempBeaconShardState[idx], *u.HashFinal())
		}
	}

	newUnsignedBeaconBlock.Body.(*blockchain.BlockBodyBeacon).ShardState = tempBeaconShardState

	return newUnsignedBeaconBlock
}

func UpdateBeaconBestState(beaconBestState *blockchain.BestStateBeacon, newBlock *blockchain.BlockV2) (*blockchain.BestStateBeacon, error) {
	//variable
	// swap 3 validators each time
	const offset = int(3)
	// shardSwapValidator := make(map[byte][]string)
	beaconSwapValidator := []string{}
	// TODO:
	// update BestShardHash, BestBlock, BestBlockHash
	beaconBestState.BestBlockHash = newBlock.Hash()
	beaconBestState.BestBlock = newBlock
	shardState := newBlock.Body.(*blockchain.BeaconBlockBody).ShardState
	for idx, l := range shardState {
		beaconBestState.BestShardHash[idx] = l[len(l)-1]
	}

	// Assign Validator
	// Shuffle candidate + validator for beacon
	if beaconBestState.BeaconHeight%200 == 1 {
		newBeaconNode, newShardNode := GetStakingCandidate(newBlock)
		beaconBestState.UnassignBeaconCandidate = append(beaconBestState.UnassignBeaconCandidate, newBeaconNode...)
		beaconBestState.UnassignShardCandidate = append(beaconBestState.UnassignShardCandidate, newShardNode...)
		//TODO: assign unAssignCandidate to assignCandidate	& clear UnassignShardCandidate

		/// Shuffle candidate for shard
		// assign UnassignShardCandidate to ShardPendingValidator with CurrentRandom this shard
		err := AssignValidatorShard(beaconBestState.ShardPendingValidator, beaconBestState.UnassignShardCandidate, beaconBestState.CurrentRandomNumber)
		// reset beaconBestState.UnassignShardCandidate
		beaconBestState.UnassignShardCandidate = []string{}

		if err != nil {
			return beaconBestState, err
		}
		// for i := 0; i < 256; i++ {
		// 	shardID := byte(i)
		// 	//swap validator for each shard
		// 	beaconBestState.ShardPendingValidator[shardID], beaconBestState.ShardValidator[shardID], shardSwapValidator[shardID], err = SwapValidator(beaconBestState.ShardPendingValidator[shardID], beaconBestState.ShardValidator[shardID], offset)
		// 	if err != nil {
		// 		return beaconBestState, err
		// 	}
		// }
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
		beaconBestState.BeaconPendingCandidate, beaconBestState.BeaconCandidate, beaconSwapValidator, err = SwapValidator(beaconBestState.BeaconPendingCandidate, beaconBestState.BeaconCandidate, offset)
		if err != nil {
			return beaconBestState, err
		}
		// update random number for new epoch
		fmt.Println(beaconSwapValidator)
		beaconBestState.CurrentRandomNumber = beaconBestState.NextRandomNumber
	} else {
		// GetStakingCandidate -> UnassignCandidate
		newBeaconNode, newShardNode := GetStakingCandidate(newBlock)
		beaconBestState.UnassignBeaconCandidate = append(beaconBestState.UnassignBeaconCandidate, newBeaconNode...)
		beaconBestState.UnassignShardCandidate = append(beaconBestState.UnassignShardCandidate, newShardNode...)
	}

	return beaconBestState, nil
	// update param
	instructions := newBlock.Body.(*blockchain.BeaconBlockBody).Instructions

	for _, l := range instructions {
		if l[0] == "set" {
			beaconBestState.Params[l[1]] = l[2]
		}
		if l[0] == "del" {
			delete(beaconBestState.Params, l[1])
		}
		if l[0] == "swap" {
			//TODO: remove from candidate list
			// format
			// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,...") "shard" "shardID"]
			// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,...") "beacon"]
			inPubkeys := strings.Split(l[1], ",")
			outPubkeys := strings.Split(l[2], ",")
			if l[3] == "shard" {
				temp, err := strconv.Atoi(l[4])
				if err != nil {
					return beaconBestState, nil
				}
				shardID := byte(temp)
				beaconBestState.ShardPendingValidator[shardID], err = RemoveValidator(beaconBestState.ShardPendingValidator[shardID], inPubkeys)
				if err != nil {
					return beaconBestState, nil
				}
				beaconBestState.ShardValidator[shardID], err = RemoveValidator(beaconBestState.ShardPendingValidator[shardID], outPubkeys)
				if err != nil {
					return beaconBestState, nil
				}
				beaconBestState.ShardValidator[shardID] = append(beaconBestState.ShardValidator[shardID], inPubkeys...)
			} else if l[3] == "beacon" {
				var err error
				beaconBestState.BeaconPendingCandidate, err = RemoveValidator(beaconBestState.BeaconPendingCandidate, inPubkeys)
				if err != nil {
					return beaconBestState, nil
				}
				beaconBestState.BeaconCandidate, err = RemoveValidator(beaconBestState.BeaconPendingCandidate, outPubkeys)
				if err != nil {
					return beaconBestState, nil
				}
				beaconBestState.BeaconCandidate = append(beaconBestState.BeaconCandidate, inPubkeys...)
			}
		}
	}

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
// return value: #1 remaining pendingValidators, #2 new currentValidators # swap validator
func SwapValidator(pendingValidators []string, currentValidators []string, offset int) ([]string, []string, []string, error) {
	if offset == 0 {
		return pendingValidators, currentValidators, nil, errors.New("Can't not swap 0 validator")
	}
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if offset > len(pendingValidators) {
		offset = len(pendingValidators)
	}
	// do nothing
	if offset == 0 {
		return pendingValidators, currentValidators, nil, errors.New("No pending validator for swapping")
	}
	if offset > len(currentValidators) {
		return pendingValidators, currentValidators, nil, errors.New("Trying to swap too many validator")
	}
	swapValidator := currentValidators[:offset]
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	currentValidators = currentValidators[offset:]
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	tempValidators := pendingValidators[:offset]
	// save new pending validators list
	pendingValidators = pendingValidators[offset:]

	// enqueue new validator to the remaning of current validators list
	currentValidators = append(currentValidators, tempValidators...)
	return pendingValidators, currentValidators, swapValidator, nil
}

// return: #param1: validator list after remove
// in parameter: #param1: list of full validator
// in parameter: #param2: list of removed validator
// removed validators list must be a subset of full validator list and it must be first in the list
func RemoveValidator(validators []string, removedValidators []string) ([]string, error) {
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if len(removedValidators) > len(validators) {
		return validators, errors.New("Trying to remove too many validators")
	}

	for index, validator := range removedValidators {
		if strings.Compare(validators[index], validator) == 0 {
			validators = validators[1:]
		} else {
			return validators, errors.New("Remove Validator with Wrong Format")
		}
	}
	return validators, nil
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
