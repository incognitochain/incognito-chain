package blockchain

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ninjadotorg/constant/blockchain/btc/btcapi"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
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
func (self *BlkTmplGenerator) NewBlockBeacon(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey) (*BeaconBlock, error) {
	beaconBlock := &BeaconBlock{}
	beaconBestState := BestStateBeacon{}
	// lock blockchain
	self.chain.chainLock.Lock()
	beaconBestState, err := self.chain.GetMaybeAcceptBeaconBestState(self.chain.highestBeaconBlock)
	if err != nil {
		tempMarshal, err := json.Marshal(self.chain.BestState.Beacon)
		if err != nil {
			return nil, NewBlockChainError(UnmashallJsonBlockError, err)
		}
		json.Unmarshal(tempMarshal, beaconBestState)
	}
	if reflect.DeepEqual(beaconBestState, BestStateBeacon{}) {
		panic(NewBlockChainError(BeaconError, errors.New("Can't create beacon block beacause no beststate found")))
	}
	// unlock blockchain
	self.chain.chainLock.Unlock()

	//==========Create header
	beaconBlock.Header.Producer = base58.Base58Check{}.Encode(payToAddress.Pk, byte(0x00))
	beaconBlock.Header.Version = VERSION
	beaconBlock.Header.Height = beaconBestState.BeaconHeight + 1
	beaconBlock.Header.Epoch = beaconBestState.BeaconEpoch
	if beaconBlock.Header.Height%200 == 0 {
		beaconBlock.Header.Epoch++
	}
	beaconBlock.Header.Timestamp = time.Now().Unix()
	beaconBlock.Header.PrevBlockHash = beaconBestState.BestBlockHash
	tempShardState, staker, swap := self.GetShardState()
	tempInstruction := beaconBestState.GenerateInstruction(beaconBlock, staker, swap)
	//==========Create Body
	beaconBlock.Body.Instructions = tempInstruction
	beaconBlock.Body.ShardState = tempShardState
	// Process new block with beststate
	beaconBestState.Update(beaconBlock)
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
	// TODO: Shard validator root
	// Shard state hash
	tempShardStateArr := []common.Hash{}
	for _, hashes := range tempShardState {
		tempShardStateArr = append(tempShardStateArr, hashes...)
	}
	tempShardStateHash, err := GenerateHashFromHashArray(tempShardStateArr)
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

	return beaconBlock, nil
}

func (self *BlkTmplGenerator) GetShardState() (map[byte][]common.Hash, map[byte]interface{}, map[byte]interface{}) {
	shardState := make(map[byte][]common.Hash)
	staker := make(map[byte]interface{})
	swap := make(map[byte]interface{})
	shardsBlocks := self.shardToBeaconPool.GetFinalBlock()
	for shardID, shardBlocks := range shardsBlocks {
		for _, shardBlock := range shardBlocks {
			//TODO: Need to define hash to append
			shardState[shardID] = append(shardState[shardID], shardBlock.Header.Hash())
			//TODO: Get staker from shard block
			//TODO: Get Swap validator from shard block
		}
	}
	return shardState, staker, swap
}

/*
	- set instruction
	- del instruction
	- swap instruction -> ok
	+ format
	+ ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
	+ ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
	- random instruction -> ok
	- assign instruction -> ok
*/
func (self *BestStateBeacon) GenerateInstruction(block *BeaconBlock, staker map[byte]interface{}, swap map[byte]interface{}) [][]string {
	instructions := [][]string{}
	//=======Swap
	// Shard Swap: both abnormal or normal swap
	for _, swapInstruction := range swap {
		instructions = append(instructions, swapInstruction.([]string))
		//TODO: detect swap and change
		// - ShardCommittee map[byte][]string
		// - ShardPendingValidator map[byte][]string
		// Build ShardValidatorsRoot
	}
	// TODO: beacon unexpeted swap
	// Beacon normal swap
	if block.Header.Height%EPOCH == EPOCH-1 {
		swapBeaconInstructions := []string{}
		swappedValidator := []string{}
		beaconNextCommittee := []string{}
		_, _, swappedValidator, beaconNextCommittee, _ = SwapValidator(self.BeaconPendingValidator, self.BeaconCommittee, OFFSET)
		swapBeaconInstructions = append(swapBeaconInstructions, "swap")
		swapBeaconInstructions = append(swapBeaconInstructions, beaconNextCommittee...)
		swapBeaconInstructions = append(swapBeaconInstructions, swappedValidator...)
		swapBeaconInstructions = append(swapBeaconInstructions, "beacon")
		instructions = append(instructions, swapBeaconInstructions)
	}

	//=======Assign
	// ["assign", "pubkey.....", "shard" or "beacon"]
	// beaconStaker := []string{}
	// shardStaker := []string{}
	for _, assignInstruction := range staker {
		instructions = append(instructions, assignInstruction.([]string))
		// assignInstructionTemp := assignInstruction.([]string)
		// if assignInstructionTemp[0] == "assign" && assignInstructionTemp[2] == "beacon" {
		// 	beaconStaker = append(beaconStaker, strings.Split(assignInstructionTemp[1], ",")...)
		// }
		// if assignInstructionTemp[0] == "assign" && assignInstructionTemp[2] == "shard" {
		// 	shardStaker = append(shardStaker, strings.Split(assignInstructionTemp[1], ",")...)
		// }
	}

	//=======Random
	// Time to get random number and no block in this epoch get it
	if block.Header.Height%200 >= RANDOM_TIME && self.IsGetRandomNUmber == false {
		chainTimeStamp, err := btcapi.GetCurrentChainTimeStamp()
		if err != nil {
			panic(err)
		}
		if chainTimeStamp > self.CurrentRandomTimeStamp {
			randomInstruction := GenerateRandomInstruction(self.CurrentRandomTimeStamp)
			instructions = append(instructions, randomInstruction)
			Logger.log.Infof("RandomNumber %+v", randomInstruction)

			// beaconAssingInstruction := []string{"assign"}
			// beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(beaconStaker, ","))
			// beaconAssingInstruction = append(beaconAssingInstruction, "beacon")

			// shardAssingInstruction := []string{"assign"}
			// shardAssingInstruction = append(shardAssingInstruction, strings.Join(shardStaker, ","))
			// shardAssingInstruction = append(shardAssingInstruction, "shard")
		}
	}
	return instructions
}

// ["random" "{blockheight}" "{bitcointimestamp}" "{nonce}" "{timestamp}"]
func GenerateRandomInstruction(timestamp int64) []string {
	msg := make(chan string)

	go btcapi.GenerateRandomNumber(timestamp, msg)
	res := <-msg
	reses := strings.Split(res, (","))
	strs := []string{}
	strs = append(strs, "random")
	strs = append(strs, reses...)
	strs = append(strs, strconv.Itoa(int(timestamp)))
	return strs
}
