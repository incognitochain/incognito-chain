package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bradfitz/slice"
	"github.com/ninjadotorg/constant/cashec"

	"github.com/ninjadotorg/constant/blockchain/btc/btcapi"
	"github.com/ninjadotorg/constant/common/base58"
	privacy "github.com/ninjadotorg/constant/privacy"
)

/*
	Load beststate + block of current block from cache to create new block
	Because final beststate height should behind highest block 1
	For example: current height block: 91, final beststate should be 90, new block height is 92

	Create Block (Body and Header)
	* Header:
		1. Create Producer: public key of child 0 or from config
		2. Create Version: load current version
		3. Create Height: prev block height + 1
		4. Create Epoch: Epoch ++ if height % epoch 0
		5. Create timestamp: now
		6. Attach previous block hash

	* Header & Body
		7. Create Shard State:
			- Shard State Vulue from beaconblockpool
			- Shard State Hash
			- Get new staker from shard(beacon or pool) -> help to create Instruction
			- Swap validator from shard -> help to create Instruction
		8. Create Instruction:
			- Instruction value -> body
			- Instruction Hash -> Header
		9. Process Instruction with best state:
			- Create Validator Root -> Header
			- Create BeaconCandidate Root -> Header
	Sign:
		Sign block and update validator index, agg sig
*/
func (self *BlkTmplGenerator) NewBlockBeacon(payToAddress *privacy.PaymentAddress, privateKey *privacy.SpendingKey) (*BeaconBlock, error) {
	beaconBlock := &BeaconBlock{}
	beaconBestState := BestStateBeacon{}
	// lock blockchain
	self.chain.chainLock.Lock()

	// produce new block with current beststate
	tempMarshal, err := json.Marshal(self.chain.BestState.Beacon)
	if err != nil {
		return nil, NewBlockChainError(UnmashallJsonBlockError, err)
	}
	err = json.Unmarshal(tempMarshal, &beaconBestState)
	if err != nil {
		return nil, NewBlockChainError(UnmashallJsonBlockError, err)
	}

	if reflect.DeepEqual(beaconBestState, BestStateBeacon{}) {
		panic(NewBlockChainError(BeaconError, errors.New("Problem with beststate in producing new block")))
	}
	// unlock blockchain
	self.chain.chainLock.Unlock()

	//==========Create header
	beaconBlock.Header.Producer = base58.Base58Check{}.Encode(payToAddress.Pk, byte(0x00))
	beaconBlock.Header.Version = VERSION
	beaconBlock.Header.Height = beaconBestState.BeaconHeight + 1
	beaconBlock.Header.Epoch = beaconBestState.BeaconEpoch
	// Eg: Epoch is 200 blocks then increase epoch at block 201, 401, 601
	if beaconBlock.Header.Height%EPOCH == 1 {
		beaconBlock.Header.Epoch++
	}
	beaconBlock.Header.Timestamp = time.Now().Unix()
	beaconBlock.Header.PrevBlockHash = beaconBestState.BestBlockHash
	tempShardState, staker, swap := self.GetShardState(&beaconBestState)
	tempInstruction := beaconBestState.GenerateInstruction(beaconBlock, staker, swap, self.chain.BestState.Beacon.CandidateShardWaitingForCurrentRandom)
	//==========Create Body
	beaconBlock.Body.Instructions = tempInstruction
	beaconBlock.Body.ShardState = tempShardState
	//==========End Create Body
	//============Process new block with beststate
	beaconBestState.Update(beaconBlock)
	//============End Process new block with beststate
	//==========Create Hash in Header
	// BeaconValidator root: beacon committee + beacon pending committee
	validatorArr := append(beaconBestState.BeaconCommittee, beaconBestState.BeaconPendingValidator...)
	beaconBlock.Header.ValidatorsRoot, err = GenerateHashFromStringArray(validatorArr)
	if err != nil {
		panic(err)
	}
	// BeaconCandidate root: beacon current candidate + beacon next candidate
	beaconCandidateArr := append(beaconBestState.CandidateBeaconWaitingForCurrentRandom, beaconBestState.CandidateBeaconWaitingForNextRandom...)
	beaconBlock.Header.BeaconCandidateRoot, err = GenerateHashFromStringArray(beaconCandidateArr)
	if err != nil {
		panic(err)
	}
	// Shard candidate root: shard current candidate + shard next candidate
	shardCandidateArr := append(beaconBestState.CandidateShardWaitingForCurrentRandom, beaconBestState.CandidateShardWaitingForNextRandom...)
	beaconBlock.Header.ShardCandidateRoot, err = GenerateHashFromStringArray(shardCandidateArr)
	if err != nil {
		panic(err)
	}
	beaconBlock.Header.ShardValidatorsRoot, err = GenerateHashFromMapByteString(beaconBestState.ShardPendingValidator, beaconBestState.ShardCommittee)
	if err != nil {
		panic(err)
	}
	// Shard state hash
	tempShardStateHash, err := GenerateHashFromShardState(tempShardState)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	beaconBlock.Header.ShardStateHash = tempShardStateHash
	// Instruction Hash
	tempInstructionArr := []string{}
	for _, strs := range tempInstruction {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	tempInstructionHash, err := GenerateHashFromStringArray(tempInstructionArr)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	beaconBlock.Header.InstructionHash = tempInstructionHash
	//===============End Create Header
	//===============Generate Signature
	// Signature of producer, sign on hash of header
	blockHash := beaconBlock.Header.Hash()
	keySet := &cashec.KeySet{}
	keySet.ImportFromPrivateKey(privateKey)
	producerSig, err := keySet.SignDataB58(blockHash.GetBytes())
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	beaconBlock.ProducerSig = producerSig
	//================End Generate Signature
	return beaconBlock, nil
}

// return param:
// #1: shard state
// #2: valid stakers
// #3: swap validator => map[byte][][]string
func (self *BlkTmplGenerator) GetShardState(beaconBestState *BestStateBeacon) (map[byte][]ShardState, [][]string, map[byte][][]string) {
	shardStates := make(map[byte][]ShardState)
	stakers := [][]string{}
	swaps := [][]string{}
	validStakers := [][]string{}
	validSwap := make(map[byte][][]string)
	//Get shard to beacon block from pool
	shardsBlocks := self.shardToBeaconPool.GetFinalBlock()

	//Shard block is a map ShardId -> array of shard block
	for shardID, shardBlocks := range shardsBlocks {
		// Only accept block in one epoch
		tempShardBlocks := make([]ShardToBeaconBlock, len(shardBlocks))
		copy(tempShardBlocks, shardBlocks)
		totalBlock := 0
		slice.Sort(tempShardBlocks[:], func(i, j int) bool {
			return tempShardBlocks[i].Header.Height < tempShardBlocks[j].Header.Height
		})
		if !reflect.DeepEqual(tempShardBlocks, shardBlocks) {
			panic("Shard To Beacon block not in right format of increasing height")
		}
		for index, shardBlock := range shardBlocks {
			currentCommittee := beaconBestState.ShardCommittee[shardID]
			hash := shardBlock.Header.Hash()
			err := ValidateAggSignature(shardBlock.ValidatorsIdx, currentCommittee, shardBlock.AggregatedSig, shardBlock.R, &hash)
			if err != nil {
				totalBlock = index
				break
			}
			for _, l := range shardBlock.Instructions {
				if l[0] == "swap" {
					if l[3] != "shard" || l[4] != strconv.Itoa(int(shardID)) {
						panic("Swap instruction is invalid")
					} else {
						validSwap[shardID] = append(validSwap[shardID], l)
					}
				}
			}
			if index != 0 && err != nil {
				totalBlock = index
				break
			}
		}
		for _, shardBlock := range shardBlocks[:totalBlock+1] {
			// for each shard block, create a corresponding shard state
			instructions := shardBlock.Instructions
			shardState := ShardState{}
			shardState.CrossShard = make([]byte, len(shardBlock.Header.CrossShards))
			copy(shardState.CrossShard, shardBlock.Header.CrossShards)
			shardState.Hash = shardBlock.Header.Hash()
			shardState.Height = shardBlock.Header.Height
			shardStates[shardID] = append(shardStates[shardID], shardState)

			for _, l := range instructions {
				if l[0] == "stake" {
					stakers = append(stakers, l)
				} else if l[0] == "swap" {
					swaps = append(swaps, l)
				}
			}
			// ["stake" "pubkey1,pubkey2,..." "shard"]
			// ["stake" "pubkey1,pubkey2,..." "beacon"]
			for _, staker := range stakers {
				tempStaker := []string{}
				newBeaconCandidate, newShardCandidate := GetStakeValidatorArrayString(staker)
				assignShard := true
				if !reflect.DeepEqual(newBeaconCandidate, []string{}) {
					tempStaker = make([]string, len(newBeaconCandidate))
					copy(tempStaker, newBeaconCandidate[:])
					assignShard = false
				} else {
					tempStaker = make([]string, len(newShardCandidate))
					copy(tempStaker, newShardCandidate[:])
				}
				tempStaker = self.chain.BestState.Beacon.GetValidStakers(tempStaker)
				if assignShard {
					validStakers = append(validStakers, []string{"stake", strings.Join(tempStaker, ","), "shard"})
				} else {
					validStakers = append(validStakers, []string{"stake", strings.Join(tempStaker, ","), "beacon"})
				}
			}
			// format
			// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,...") "shard" "shardID"]
			// ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,...") "beacon"]
			// Validate swap instruction => extract only valid swap instruction
			//TODO: define error handler scheme
			for _, swap := range swaps {
				if swap[2] == "beacon" {
					continue
				} else if swap[2] == "shard" {
					temp, err := strconv.Atoi(swap[3])
					if err != nil {
						continue
					}
					swapShardID := byte(temp)
					if swapShardID != shardID {
						continue
					}
					validSwap[shardID] = append(validSwap[shardID], swap)
				} else {
					continue
				}
			}
		}
	}
	return shardStates, validStakers, validSwap
}

/*
	- set instruction
	- del instruction
	- swap instruction -> ok
	+ format
	+ ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
	+ ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
	- random instruction -> ok
	- stake instruction -> ok
*/
func (self *BestStateBeacon) GenerateInstruction(block *BeaconBlock, stakers [][]string, swap map[byte][][]string, shardCandidates []string) [][]string {
	instructions := [][]string{}
	//=======Swap
	// Shard Swap: both abnormal or normal swap
	for _, swapInstruction := range swap {
		instructions = append(instructions, swapInstruction...)
	}
	// TODO: beacon unexpeted swap -> pbft
	// Beacon normal swap
	if block.Header.Height%EPOCH == EPOCH-1 {
		swapBeaconInstructions := []string{}
		swappedValidator := []string{}
		beaconNextCommittee := []string{}
		_, _, swappedValidator, beaconNextCommittee, _ = SwapValidator(self.BeaconPendingValidator, self.BeaconCommittee, COMMITEES, OFFSET)
		swapBeaconInstructions = append(swapBeaconInstructions, "swap")
		swapBeaconInstructions = append(swapBeaconInstructions, beaconNextCommittee...)
		swapBeaconInstructions = append(swapBeaconInstructions, swappedValidator...)
		swapBeaconInstructions = append(swapBeaconInstructions, "beacon")
		instructions = append(instructions, swapBeaconInstructions)
	}
	//=======Stake
	// ["stake", "pubkey.....", "shard" or "beacon"]
	// beaconStaker := []string{}
	// shardStaker := []string{}
	for _, stakeInstruction := range stakers {
		instructions = append(instructions, stakeInstruction)
	}
	//=======Random and Assign if random number is detected
	// Time to get random number and no block in this epoch get it
	fmt.Printf("RandomTimestamp %+v \n", self.CurrentRandomTimeStamp)
	fmt.Printf("============height epoch: %+v, RANDOM TIME: %+v \n", block.Header.Height%EPOCH, RANDOM_TIME)
	fmt.Printf("============IsGetRandomNumber %+v \n", self.IsGetRandomNumber)
	if block.Header.Height%EPOCH > RANDOM_TIME && self.IsGetRandomNumber == false {
		chainTimeStamp, err := btcapi.GetCurrentChainTimeStamp()
		fmt.Printf("============chainTimeStamp %+v \n", chainTimeStamp)
		if err != nil {
			panic(err)
		}
		var wg sync.WaitGroup
		assignedCandidates := make(map[byte][]string)
		if chainTimeStamp > self.CurrentRandomTimeStamp {
			wg.Add(1)
			randomInstruction, rand := GenerateRandomInstruction(self.CurrentRandomTimeStamp, &wg)
			wg.Wait()
			instructions = append(instructions, randomInstruction)
			Logger.log.Infof("RandomNumber %+v", randomInstruction)
			for _, candidate := range shardCandidates {
				shardID := calculateHash(candidate, rand)
				assignedCandidates[shardID] = append(assignedCandidates[shardID], candidate)
			}
			for shardId, candidates := range assignedCandidates {
				shardAssingInstruction := []string{"assign"}
				shardAssingInstruction = append(shardAssingInstruction, strings.Join(candidates, ","))
				shardAssingInstruction = append(shardAssingInstruction, "shard")
				shardAssingInstruction = append(shardAssingInstruction, strconv.Itoa(int(shardId)))
				instructions = append(instructions, shardAssingInstruction)
			}
		}
	}
	return instructions
}
func (self *BestStateBeacon) GetValidStakers(tempStaker []string) []string {
	for _, committees := range self.ShardCommittee {
		tempStaker = GetValidStaker(committees, tempStaker)
	}
	for _, validators := range self.ShardPendingValidator {
		tempStaker = GetValidStaker(validators, tempStaker)
	}
	tempStaker = GetValidStaker(self.BeaconCommittee, tempStaker)
	tempStaker = GetValidStaker(self.BeaconPendingValidator, tempStaker)
	tempStaker = GetValidStaker(self.CandidateBeaconWaitingForCurrentRandom, tempStaker)
	tempStaker = GetValidStaker(self.CandidateBeaconWaitingForNextRandom, tempStaker)
	tempStaker = GetValidStaker(self.CandidateShardWaitingForCurrentRandom, tempStaker)
	tempStaker = GetValidStaker(self.CandidateBeaconWaitingForNextRandom, tempStaker)
	return tempStaker
}

//===================================Util for Beacon=============================

// ["random" "{blockheight}" "{bitcointimestamp}" "{nonce}" "{timestamp}"]
func GenerateRandomInstruction(timestamp int64, wg *sync.WaitGroup) ([]string, int64) {
	msg := make(chan string)
	go btcapi.GenerateRandomNumber(timestamp, msg)
	res := <-msg
	reses := strings.Split(res, (","))
	strs := []string{}
	strs = append(strs, "random")
	strs = append(strs, reses...)
	strs = append(strs, strconv.Itoa(int(timestamp)))
	nonce, _ := strconv.Atoi(reses[2])
	wg.Done()
	return strs, int64(nonce)
}

func GetValidStaker(committees []string, stakers []string) []string {
	validStaker := []string{}
	for _, staker := range stakers {
		flag := false
		for _, committee := range committees {
			if strings.Compare(staker, committee) == 0 {
				flag = true
				break
			}
		}
		if !flag {
			validStaker = append(validStaker, staker)
		}
	}
	return validStaker
}

func GetStakeValidatorArrayString(v []string) ([]string, []string) {
	beacon := []string{}
	shard := []string{}
	if v[0] == "stake" && v[2] == "beacon" {
		beacon = strings.Split(v[1], ",")
	}
	if v[0] == "stake" && v[2] == "shard" {
		shard = strings.Split(v[1], ",")
	}
	return beacon, shard
}
