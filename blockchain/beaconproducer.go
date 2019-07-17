package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
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
func (blkTmplGenerator *BlkTmplGenerator) NewBlockBeacon(producerAddress *privacy.PaymentAddress, round int, shardsToBeacon map[byte]uint64) (*BeaconBlock, error) {
	beaconBlock := &BeaconBlock{}
	beaconBestState := BestStateBeacon{}
	// lock blockchain
	blkTmplGenerator.chain.chainLock.Lock()
	// fmt.Printf("Beacon Produce: BeaconBestState Original %+v \n", blkTmplGenerator.chain.BestState.Beacon)
	// produce new block with current beststate
	tempMarshal, err := blkTmplGenerator.chain.BestState.Beacon.MarshalJSON()
	if err != nil {
		blkTmplGenerator.chain.chainLock.Unlock()
		return nil, NewBlockChainError(MashallJsonError, err)
	}
	err = json.Unmarshal(tempMarshal, &beaconBestState)
	if err != nil {
		blkTmplGenerator.chain.chainLock.Unlock()
		return nil, NewBlockChainError(UnmashallJsonBlockError, err)
	}
	beaconBestState.CandidateShardWaitingForCurrentRandom = blkTmplGenerator.chain.BestState.Beacon.CandidateShardWaitingForCurrentRandom
	beaconBestState.CandidateShardWaitingForNextRandom = blkTmplGenerator.chain.BestState.Beacon.CandidateShardWaitingForNextRandom
	beaconBestState.CandidateBeaconWaitingForCurrentRandom = blkTmplGenerator.chain.BestState.Beacon.CandidateBeaconWaitingForCurrentRandom
	beaconBestState.CandidateBeaconWaitingForNextRandom = blkTmplGenerator.chain.BestState.Beacon.CandidateBeaconWaitingForNextRandom

	if reflect.DeepEqual(beaconBestState, BestStateBeacon{}) {
		blkTmplGenerator.chain.chainLock.Unlock()
		panic(NewBlockChainError(BeaconError, errors.New("problem with beststate in producing new block")))
	}

	// unlock blockchain
	blkTmplGenerator.chain.chainLock.Unlock()

	//==========Create header
	beaconBlock.Header.ProducerAddress = *producerAddress
	beaconBlock.Header.Version = VERSION
	beaconBlock.Header.Height = beaconBestState.BeaconHeight + 1
	beaconBlock.Header.Epoch = beaconBestState.Epoch
	beaconBlock.Header.Round = round
	fmt.Printf("[db] producing block: %d\n", beaconBlock.Header.Height)
	// Eg: Epoch is 200 blocks then increase epoch at block 201, 401, 601
	rewardByEpochInstruction := [][]string{}
	if beaconBlock.Header.Height%common.EPOCH == 1 {
		rewardByEpochInstruction, err = blkTmplGenerator.chain.BuildRewardInstructionByEpoch(beaconBlock.Header.Epoch)
		if err != nil {
			fmt.Printf("[ndh]-[ERROR] -- --- -- --- %+v\n", err)
			return nil, err
		}
		beaconBlock.Header.Epoch++
	}
	beaconBlock.Header.PrevBlockHash = beaconBestState.BestBlockHash
	//fmt.Println("[db] NewBlockBeacon GetShardState")
	tempShardState, staker, swap, stabilityInstructions, acceptedRewardInstructions := blkTmplGenerator.GetShardState(&beaconBestState, shardsToBeacon)
	bestStateBeacon.InitRandomClient(blkTmplGenerator.chain.config.RandomClient)
	tempInstruction := beaconBestState.GenerateInstruction(beaconBlock, staker, swap, beaconBestState.CandidateShardWaitingForCurrentRandom, stabilityInstructions, acceptedRewardInstructions)
	if len(rewardByEpochInstruction) != 0 {
		tempInstruction = append(tempInstruction, rewardByEpochInstruction...)
	}
	//fmt.Println("BeaconProducer/tempInstruction", tempInstruction)
	//==========Create Body
	beaconBlock.Body.Instructions = tempInstruction
	beaconBlock.Body.ShardState = tempShardState
	//==========End Create Body
	//============Process new block with beststate
	fmt.Println("Beacon Candidate", beaconBestState.CandidateBeaconWaitingForCurrentRandom)
	if len(beaconBlock.Body.Instructions) != 0 {
		Logger.log.Critical("Beacon Produce: Beacon Instruction", beaconBlock.Body.Instructions)
	}
	beaconBestState.Update(beaconBlock, blkTmplGenerator.chain)
	//============End Process new block with beststate
	//==========Create Hash in Header
	// BeaconValidator root: beacon committee + beacon pending committee
	validatorArr := append(beaconBestState.BeaconCommittee, beaconBestState.BeaconPendingValidator...)
	beaconBlock.Header.ValidatorsRoot, err = GenerateHashFromStringArray(validatorArr)
	// fmt.Printf("Beacon Produce/AfterUpdate: Beacon Pending Validator %+v , Beacon Committee %+v, Beacon Validator Root %+v \n", beaconBestState.BeaconPendingValidator, beaconBestState.BeaconCommittee, beaconBlock.Header.ValidatorsRoot)
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
	// Shard Validator root
	beaconBlock.Header.ShardValidatorsRoot, err = GenerateHashFromMapByteString(beaconBestState.GetShardPendingValidator(), beaconBestState.GetShardCommittee())
	// fmt.Printf("Beacon Produce/AfterUpdate: Shard Pending Validator %+v , ShardCommitee %+v, Shard Validator Root %+v \n", beaconBestState.ShardPendingValidator, beaconBestState.ShardCommittee, beaconBlock.Header.ShardValidatorsRoot)
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

	// Instruction merkle root
	flattenInsts := FlattenAndConvertStringInst(tempInstruction)
	copy(beaconBlock.Header.InstructionMerkleRoot[:], GetKeccak256MerkleRoot(flattenInsts))

	//===============End Create Header
	return beaconBlock, nil
}

func (blkTmplGenerator *BlkTmplGenerator) FinalizeBeaconBlock(blk *BeaconBlock, producerKeyset *incognitokey.KeySet) error {
	// Signature of producer, sign on hash of header
	blk.Header.Timestamp = time.Now().Unix()
	blockHash := blk.Header.Hash()
	producerSig, err := producerKeyset.SignDataB58(blockHash.GetBytes())
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	blk.ProducerSig = producerSig
	//================End Generate Signature
	return nil
}

// return param:
// #1: shard state
// #2: valid stakers
// #3: swap validator => map[byte][][]string
func (blkTmplGenerator *BlkTmplGenerator) GetShardState(
	beaconBestState *BestStateBeacon,
	shardsToBeacon map[byte]uint64,
) (
	map[byte][]ShardState,
	[][]string,
	map[byte][][]string,
	[][]string,
	[][]string,
) {

	shardStates := make(map[byte][]ShardState)
	validStakers := [][]string{}
	validSwappers := make(map[byte][][]string)
	//Get shard to beacon block from pool
	allShardBlocks := blkTmplGenerator.shardToBeaconPool.GetValidBlock(shardsToBeacon)
	//Shard block is a map ShardId -> array of shard block
	stabilityInstructions := [][]string{}
	acceptedRewardInstructions := [][]string{}
	var keys []int
	for k := range allShardBlocks {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	//fmt.Printf("[db] in GetShardState\n")
	for _, value := range keys {
		shardID := byte(value)
		shardBlocks := allShardBlocks[shardID]
		// Only accept block in one epoch
		totalBlock := 0
		//UNCOMMENT FOR TESTING
		fmt.Println("Beacon Producer/ Got These Block from pool")
		for _, shardBlocks := range shardBlocks {
			fmt.Printf(" %+v ", shardBlocks.Header.Height)
		}
		//=======
		for index, shardBlock := range shardBlocks {
			currentCommittee := beaconBestState.GetAShardCommittee(shardID)
			hash := shardBlock.Header.Hash()
			err1 := ValidateAggSignature(shardBlock.ValidatorsIdx, currentCommittee, shardBlock.AggregatedSig, shardBlock.R, &hash)
			fmt.Println("Beacon Producer/ Validate Agg Signature for shard", shardID, err1 == nil)
			if err1 != nil {
				break
			}
			if index != 0 && err1 != nil {
				break
			}
			totalBlock = index
		}
		fmt.Printf("Beacon Producer/ AFTER FILTER, ONLY GET %+v block \n", totalBlock)
		fmt.Println("Beacon Producer/ FILTER and ONLY GET These Block from pool")
		if totalBlock > 49 {
			totalBlock = 49
		}
		for _, shardBlock := range shardBlocks[:totalBlock+1] {
			shardState, validStaker, validSwapper, stabilityInstruction, acceptedRewardInstruction := blkTmplGenerator.chain.GetShardStateFromBlock(beaconBestState, shardBlock, shardID)
			shardStates[shardID] = append(shardStates[shardID], shardState[shardID])
			validStakers = append(validStakers, validStaker...)
			validSwappers[shardID] = append(validSwappers[shardID], validSwapper[shardID]...)
			stabilityInstructions = append(stabilityInstructions, stabilityInstruction...)
			acceptedRewardInstructions = append(acceptedRewardInstructions, acceptedRewardInstruction)
		}
	}
	return shardStates, validStakers, validSwappers, stabilityInstructions, acceptedRewardInstructions
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
func (bestStateBeacon *BestStateBeacon) GenerateInstruction(
	block *BeaconBlock,
	stakers [][]string,
	swap map[byte][][]string,
	shardCandidates []string,
	stabilityInstructions [][]string,
	acceptedRewardInstructions [][]string,
) [][]string {
	instructions := [][]string{}
	instructions = append(instructions, stabilityInstructions...)
	instructions = append(instructions, acceptedRewardInstructions...)
	//=======Swap
	// Shard Swap: both abnormal or normal swap
	var keys []int
	for k := range swap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		instructions = append(instructions, swap[byte(shardID)]...)
	}
	// Beacon normal swap
	if block.Header.Height%common.EPOCH == 0 {
		swapBeaconInstructions := []string{}
		_, currentValidators, swappedValidator, beaconNextCommittee, _ := SwapValidator(bestStateBeacon.BeaconPendingValidator, bestStateBeacon.BeaconCommittee, bestStateBeacon.BeaconCommitteeSize, common.OFFSET)
		if len(swappedValidator) > 0 || len(beaconNextCommittee) > 0 {
			swapBeaconInstructions = append(swapBeaconInstructions, "swap")
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(beaconNextCommittee, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(swappedValidator, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, "beacon")
			instructions = append(instructions, swapBeaconInstructions)
			// Generate instruction storing merkle root of validators pubkey and send to bridge
			beaconRootInst := buildBeaconSwapConfirmInstruction(currentValidators, block.Header.Height+1)
			instructions = append(instructions, beaconRootInst)
		}
	}
	//=======Stake
	// ["stake", "pubkey.....", "shard" or "beacon"]
	instructions = append(instructions, stakers...)
	if block.Header.Height%common.EPOCH > common.RANDOM_TIME && !bestStateBeacon.IsGetRandomNumber {
		//=================================
		// COMMENT FOR TESTING
		//var err error
		//chainTimeStamp, err := bestStateBeacon.randomClient.GetCurrentChainTimeStamp()
		// UNCOMMENT FOR TESTING
		chainTimeStamp := bestStateBeacon.CurrentRandomTimeStamp + 1
		//==================================
		assignedCandidates := make(map[byte][]string)
		if chainTimeStamp > bestStateBeacon.CurrentRandomTimeStamp {
			randomInstruction, rand := bestStateBeacon.generateRandomInstruction(bestStateBeacon.CurrentRandomTimeStamp)
			instructions = append(instructions, randomInstruction)
			Logger.log.Critical("RandomNumber", randomInstruction)
			for _, candidate := range shardCandidates {
				shardID := calculateCandidateShardID(candidate, rand, bestStateBeacon.ActiveShards)
				assignedCandidates[shardID] = append(assignedCandidates[shardID], candidate)
			}
			Logger.log.Criticalf("assignedCandidates %+v", assignedCandidates)
			for shardId, candidates := range assignedCandidates {
				shardAssingInstruction := []string{"assign"}
				shardAssingInstruction = append(shardAssingInstruction, strings.Join(candidates, ","))
				shardAssingInstruction = append(shardAssingInstruction, "shard")
				shardAssingInstruction = append(shardAssingInstruction, fmt.Sprintf("%v", shardId))
				instructions = append(instructions, shardAssingInstruction)
			}
		}
	}
	return instructions
}

func (bestStateBeacon *BestStateBeacon) GetValidStakers(tempStaker []string) []string {
	for _, committees := range bestStateBeacon.GetShardCommittee() {
		tempStaker = metadata.GetValidStaker(committees, tempStaker)
	}
	for _, validators := range bestStateBeacon.GetShardPendingValidator() {
		tempStaker = metadata.GetValidStaker(validators, tempStaker)
	}
	tempStaker = metadata.GetValidStaker(bestStateBeacon.BeaconCommittee, tempStaker)
	tempStaker = metadata.GetValidStaker(bestStateBeacon.BeaconPendingValidator, tempStaker)
	tempStaker = metadata.GetValidStaker(bestStateBeacon.CandidateBeaconWaitingForCurrentRandom, tempStaker)
	tempStaker = metadata.GetValidStaker(bestStateBeacon.CandidateBeaconWaitingForNextRandom, tempStaker)
	tempStaker = metadata.GetValidStaker(bestStateBeacon.CandidateShardWaitingForCurrentRandom, tempStaker)
	tempStaker = metadata.GetValidStaker(bestStateBeacon.CandidateShardWaitingForNextRandom, tempStaker)
	tempStaker = metadata.GetValidStaker(bestStateBeacon.CandidateShardWaitingForNextRandom, tempStaker)
	return tempStaker
}

/*
	Swap format:
	- ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
	- ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
	Stake format:
	- ["stake" "pubkey1,pubkey2,..." "shard"]
	- ["stake" "pubkey1,pubkey2,..." "beacon"]

*/
func (blockChain *BlockChain) GetShardStateFromBlock(
	beaconBestState *BestStateBeacon,
	shardBlock *ShardToBeaconBlock,
	shardID byte,
) (
	map[byte]ShardState,
	[][]string,
	map[byte][][]string,
	[][]string,
	[]string,
) {
	//Variable Declaration
	shardStates := make(map[byte]ShardState)
	validStakers := [][]string{}
	validSwap := make(map[byte][][]string)
	stakers := [][]string{}
	swapers := [][]string{}
	stabilityInstructions := [][]string{}
	acceptedBlockRewardInfo := metadata.NewAcceptedBlockRewardInfo(shardID, shardBlock.Header.TotalTxsFee, shardBlock.Header.Height)
	// str, _ := acceptedBlockRewardInfo.GetStringFormat()
	// fmt.Printf("[ndh] - - - - - - aaaaaaaaaaaaaaa\n\n\n")
	// for key, value := range shardBlock.Header.TotalTxsFee {
	// 	fmt.Printf("[ndh] ======================= %+v %+v \n", key, value)
	// }
	acceptedRewardInstructions, err := acceptedBlockRewardInfo.GetStringFormat()
	// fmt.Printf("[ndh] ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ %+v\n", acceptedRewardInstructions)
	if err != nil {
		panic("[ndh] Cant create acceptedRewardInstructions")
	}
	//Get Shard State from Block
	shardState := ShardState{}
	shardState.CrossShard = make([]byte, len(shardBlock.Header.CrossShards))
	copy(shardState.CrossShard, shardBlock.Header.CrossShards)
	shardState.Hash = shardBlock.Header.Hash()
	shardState.Height = shardBlock.Header.Height
	shardStates[shardID] = shardState

	instructions := shardBlock.Instructions
	Logger.log.Critical(instructions)
	// Validate swap instruction => for testing
	for _, l := range shardBlock.Instructions {
		if len(l) > 0 {
			if l[0] == SwapAction {
				if l[3] != "shard" || l[4] != strconv.Itoa(int(shardID)) {
					panic("Swap instruction is invalid")
				}
			}
		}
	}

	if len(instructions) != 0 {
		Logger.log.Criticalf("Instruction in shardBlock %+v, %+v \n", shardBlock.Header.Height, instructions)
	}
	for _, l := range instructions {
		// fmt.Printf("[ndh]-[INFO] - - Instruction from shard: %+v\n", l)
		if len(l) > 0 {
			if l[0] == StakeAction {
				stakers = append(stakers, l)
			}
			if l[0] == SwapAction {
				swapers = append(swapers, l)
			}
		}
	}

	stakeBeacon := []string{}
	stakeShard := []string{}
	stakeBeaconTx := []string{}
	stakeShardTx := []string{}
	if len(stakers) != 0 {
		Logger.log.Critical("Beacon Producer/ Process Stakers List", stakers)
	}
	if len(swapers) != 0 {
		Logger.log.Critical("Beacon Producer/ Process Stakers List", swapers)
	}
	// Validate stake instruction => extract only valid stake instruction
	for _, staker := range stakers {
		var tempStaker []string
		newBeaconCandidate, newShardCandidate := getStakeValidatorArrayString(staker)
		assignShard := true
		if !reflect.DeepEqual(newBeaconCandidate, []string{}) {
			tempStaker = make([]string, len(newBeaconCandidate))
			copy(tempStaker, newBeaconCandidate[:])
			assignShard = false
		} else {
			tempStaker = make([]string, len(newShardCandidate))
			copy(tempStaker, newShardCandidate[:])
		}
		tempStaker = blockChain.BestState.Beacon.GetValidStakers(tempStaker)
		tempStaker = metadata.GetValidStaker(stakeShard, tempStaker)
		tempStaker = metadata.GetValidStaker(stakeBeacon, tempStaker)

		if len(tempStaker) > 0 {
			if assignShard {
				stakeShard = append(stakeShard, tempStaker...)
				for i, v := range strings.Split(staker[1], ",") {
					if common.IndexOfStr(v, stakeShard) > -1 {
						stakeShardTx = append(stakeShardTx, strings.Split(staker[3], ",")[i])
					}
				}
			} else {
				stakeBeacon = append(stakeBeacon, tempStaker...)
				for i, v := range strings.Split(staker[1], ",") {
					if common.IndexOfStr(v, stakeBeacon) > -1 {
						stakeBeaconTx = append(stakeBeaconTx, strings.Split(staker[3], ",")[i])
					}
				}
			}
		}
	}

	if len(stakeShard) > 0 {
		validStakers = append(validStakers, []string{StakeAction, strings.Join(stakeShard, ","), "shard", strings.Join(stakeShardTx, ",")})
	}
	if len(stakeBeacon) > 0 {
		validStakers = append(validStakers, []string{StakeAction, strings.Join(stakeBeacon, ","), "beacon", strings.Join(stakeBeaconTx, ",")})
	}
	// Validate swap instruction => extract only valid swap instruction
	for _, swap := range swapers {
		if swap[3] == "beacon" {
			continue
		} else if swap[3] == "shard" {
			temp, err := strconv.Atoi(swap[4])
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
	// Create stability instruction
	fmt.Printf("[db] included shardID %d, block %d, insts: %s\n", shardID, shardBlock.Header.Height, shardBlock.Instructions)
	stabilityInstructionsPerBlock, err := blockChain.buildStabilityInstructions(
		shardID,
		shardBlock.Instructions,
		beaconBestState,
	)
	if err != nil {
		Logger.log.Errorf("Build stability instructions failed: %s \n", err.Error())
	}

	// Pick instruction with merkle root of shard committee's pubkeys and save to beacon block
	commPubkeyInst := pickBridgePubkeyRootInstruction(shardBlock)
	if len(commPubkeyInst) > 0 {
		stabilityInstructionsPerBlock = append(instructions, commPubkeyInst...)
		fmt.Printf("[db] found bridge pubkey root inst: %s\n", commPubkeyInst)
	}

	stabilityInstructions = append(stabilityInstructions, stabilityInstructionsPerBlock...)
	Logger.log.Infof("Becon Produce: Got Shard Block %+v Shard %+v \n", shardBlock.Header.Height, shardID)
	return shardStates, validStakers, validSwap, stabilityInstructions, acceptedRewardInstructions
}

//===================================Util for Beacon=============================

// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
func (bestStateBeacon *BestStateBeacon) generateRandomInstruction(timestamp int64) ([]string, int64) {
	//COMMENT FOR TESTING
	//var (
	//	blockHeight int
	//	chainTimestamp int64
	//	nonce int64
	//  strs []string
	//	err error
	//)
	//for {
	//	blockHeight, chainTimestamp, nonce, err = bestStateBeacon.randomClient.GetNonceByTimestamp(timestamp)
	//	if err == nil {
	//		break
	//	}
	//}
	//strs = append(strs, "random")
	//strs = append(strs, strconv.Itoa(int(nonce)))
	//strs = append(strs, strconv.Itoa(blockHeight))
	//strs = append(strs, strconv.Itoa(int(timestamp)))
	//strs = append(strs, strconv.Itoa(int(chainTimestamp)))
	//@NOTICE: Hard Code for testing
	var strs []string
	reses := []string{"1000", strconv.Itoa(int(timestamp)), strconv.Itoa(int(timestamp) + 1)}
	strs = append(strs, RandomAction)
	strs = append(strs, reses...)
	strs = append(strs, strconv.Itoa(int(timestamp)))
	return strs, int64(1000)
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
