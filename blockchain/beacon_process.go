package blockchain

import (
	"crypto/sha256"
	"errors"
	"sort"
	"strconv"
	"strings"
)

/*
Insert new block into beaconchain
1. Verify Block
2. Update: Process block
	2.1 Process BestStateBeacon
	2.2 Store BestStateBeacon
3. Store Block
*/
func (self *BlockChain) ConnectBlockBeacon(block *BeaconBlock) error {
	self.chainLock.Lock()
	defer self.chainLock.Unlock()
	blockHash := block.Hash().String()

	Logger.log.Infof("Insert block %+v to Blockchain", blockHash)

	//===================Verify============================
	Logger.log.Infof("Verify Pre-Process block %+v to Blockchain", blockHash)

	err := self.VerifyPreProcessingBlockBeacon(block)
	if err != nil {
		Logger.log.Error("Error update best state for block", block, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}

	//===================Process============================
	Logger.log.Infof("Process block %+v", blockHash)

	Logger.log.Infof("Process BeaconBestState block %+v", blockHash)
	// Process best state or not and store beststate
	err = self.BestState.Beacon.Update(block)
	if err != nil {
		Logger.log.Error("Error update best state for block", block, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}

	Logger.log.Infof("Verify Post-Process block %+v to Blockchain", blockHash)

	err = self.VerifyPostProcessingBlockBeacon(block)
	if err != nil {
		Logger.log.Error("Error update best state for block", block, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}

	Logger.log.Infof("Store BeaconBestState block %+v", blockHash)
	err = self.StoreBeaconBestState()

	if err != nil {
		Logger.log.Error("Error update best state for block", block, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}

	//===================Store============================
	Logger.log.Infof("Store Beacon block %+v", blockHash)
	err = self.config.DataBase.StoreBeaconBlock(block)
	if err != nil {
		return err
	}

	if err != nil {
		Logger.log.Error("Error update best state for block", block, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}
	Logger.log.Infof("Accepted block %+v", blockHash)

	return nil
}

func (self *BlockChain) VerifyPreProcessingBlockBeacon(block *BeaconBlock) error {
	return nil
	/* Verify Pre-prosessing data
	- Signature
	- version
	- parent hash
	- Height = parent hash + 1
	- Epoch
	- Timestamp can not excess some limit
	- Instruction hash
	- ShardStateHash
	- Random number
	- Sanity
	*/
	return nil
}

func (self *BlockChain) VerifyPostProcessingBlockBeacon(block *BeaconBlock) error {
	return nil
	/* Verify Post-processing data
	- Validator root
	- Candidate root

	*/
	return nil
}

func (self *BestStateBeacon) Update(newBlock *BeaconBlock) error {
	if newBlock == nil {
		return errors.New("Null pointer")
	}
	// signal of random parameter from beacon block
	randomFlag := false
	// update BestShardHash, BestBlock, BestBlockHash
	self.BestBlockHash = *newBlock.Hash()
	self.BestBlock = newBlock
	shardState := newBlock.Body.ShardState
	for idx, l := range shardState {
		self.BestShardHash[idx] = l[len(l)-1]
	}

	// update param
	instructions := newBlock.Body.Instructions

	for _, l := range instructions {
		if l[0] == "set" {
			self.Params[l[1]] = l[2]
		}
		if l[0] == "del" {
			delete(self.Params, l[1])
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
					Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				shardID := byte(temp)
				// delete in public key out of sharding pending validator list
				self.ShardPendingValidator[shardID], err = RemoveValidator(self.ShardPendingValidator[shardID], inPubkeys)
				if err != nil {
					Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				// delete out public key out of current committees
				self.ShardCommittee[shardID], err = RemoveValidator(self.ShardPendingValidator[shardID], outPubkeys)
				if err != nil {
					Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				// append in public key to committees
				self.ShardCommittee[shardID] = append(self.ShardCommittee[shardID], inPubkeys...)

				// TODO: Check new list with root hash received from block
			} else if l[3] == "beacon" {
				var err error
				self.BeaconPendingValidator, err = RemoveValidator(self.BeaconPendingValidator, inPubkeys)
				if err != nil {
					Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				self.BeaconCommittee, err = RemoveValidator(self.BeaconCommittee, outPubkeys)
				if err != nil {
					Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				self.BeaconCommittee = append(self.BeaconCommittee, inPubkeys...)
				// TODO: Check new list with root hash received from block
			}
		}
		// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
		if l[0] == "random" {
			//TODO: Verify nonce is from a right block
			temp, err := strconv.Atoi(l[4])
			if err != nil {
				Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			self.CurrentRandomNumber = int64(temp)
			randomFlag = true
		}
	}
	// get staking candidate list and store
	newBeaconCandidate, newShardCandidate := GetStakingCandidate(*newBlock)
	// store new staking candidate
	self.CandidateBeaconWaitingForNextRandom = append(self.CandidateBeaconWaitingForNextRandom, newBeaconCandidate...)
	self.CandidateShardWaitingForNextRandom = append(self.CandidateShardWaitingForNextRandom, newShardCandidate...)
	if self.BeaconHeight%EPOCH == 0 && self.BeaconHeight != 0 {
		// Begin of each epoch
	} else if self.BeaconHeight%EPOCH < RANDOM_TIME {
		// Before get random from bitcoin

	} else if self.BeaconHeight%EPOCH >= RANDOM_TIME {
		// After get random from bitcoin
		if self.BeaconHeight%EPOCH == RANDOM_TIME {
			// snapshot candidate list
			self.CandidateShardWaitingForCurrentRandom = self.CandidateShardWaitingForNextRandom
			self.CandidateBeaconWaitingForCurrentRandom = self.CandidateBeaconWaitingForNextRandom

			// reset candidate list
			self.CandidateShardWaitingForNextRandom = []string{}
			self.CandidateBeaconWaitingForNextRandom = []string{}
		}
		// if get new random number???
		// Assign candidate to shard
		// assign CandidateShardWaitingForCurrentRandom to ShardPendingValidator with CurrentRandom this shard
		if randomFlag {
			err := AssignValidatorShard(self.ShardPendingValidator, self.CandidateShardWaitingForCurrentRandom, self.CurrentRandomNumber)
			if err != nil {
				Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			// delete CandidateShardWaitingForCurrentRandom list
			self.CandidateShardWaitingForCurrentRandom = []string{}

			/// Shuffle candidate
			// shuffle CandidateBeaconWaitingForCurrentRandom with current random number
			newBeaconPendingValidator, err := ShuffleCandidate(self.CandidateBeaconWaitingForCurrentRandom, self.CurrentRandomNumber)
			if err != nil {
				Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			self.CandidateBeaconWaitingForCurrentRandom = []string{}
			self.BeaconPendingValidator = append(self.BeaconPendingValidator, newBeaconPendingValidator...)
		}
	} else if self.BeaconHeight%EPOCH == EPOCH-1 {
		// At the end of each epoch, eg: block 199, 399, 599 with epoch is 200
		// Swap pending validator in committees, pop some of public key in committees out
		// ONLY SWAP FOR BEACON
		// SHARD WILL SWAP ITSELF
		var (
			beaconSwapedCommittees []string
			err                    error
		)
		self.BeaconPendingValidator, self.BeaconCommittee, beaconSwapedCommittees, err = SwapValidator(self.BeaconPendingValidator, self.BeaconCommittee, OFFSET)
		Logger.log.Infof("Swaped out committees %+v", beaconSwapedCommittees)
		if err != nil {
			Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
			return NewBlockChainError(UnExpectedError, err)
		}
	}
	return nil
}
func GetStakingCandidate(beaconBlock BeaconBlock) (beacon []string, shard []string) {

	beaconBlockBody := beaconBlock.Body
	for _, v := range beaconBlockBody.Instructions {
		if v[0] == "assign" && v[2] == "beacon" {
			beacon = strings.Split(v[1], ",")
		}
		if v[0] == "assign" && v[2] == "shard" {
			shard = strings.Split(v[1], ",")
		}
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
