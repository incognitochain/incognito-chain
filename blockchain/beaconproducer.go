package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
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
func (blockGenerator *BlockGenerator) NewBlockBeacon(producerAddress *privacy.PaymentAddress, round int, shardsToBeaconLimit map[byte]uint64) (*BeaconBlock, error) {
	// lock blockchain
	blockGenerator.chain.chainLock.Lock()
	defer blockGenerator.chain.chainLock.Unlock()
	Logger.log.Infof("⛏ Creating Beacon Block %+v", blockGenerator.chain.BestState.Beacon.BeaconHeight+1)
	//============Init Variable============
	beaconBlock := NewBeaconBlock()
	beaconBestState := NewBeaconBestState()
	var err error
	var epoch uint64
	// produce new block with current beststate
	err = beaconBestState.cloneBeaconBestState(blockGenerator.chain.BestState.Beacon)
	if err != nil {
		return nil, err
	}
	//if !reflect.DeepEqual(*blockGenerator.chain.BestState.Beacon, *beaconBestState) {
	//	panic("abc")
	//}
	beaconBestState.InitRandomClient(blockGenerator.chain.config.RandomClient)
	//======Build Header Essential Data=======
	rewardByEpochInstruction := [][]string{}
	if (beaconBestState.BeaconHeight+1)%uint64(common.EPOCH) == 1 {
		rewardByEpochInstruction, err = blockGenerator.chain.BuildRewardInstructionByEpoch(beaconBestState.Epoch)
		if err != nil {
			return nil, NewBlockChainError(BuildRewardInstructionError, err)
		}
		epoch = beaconBestState.Epoch + 1
	} else {
		epoch = beaconBestState.Epoch
	}
	beaconBlock.Header.ProducerAddress = *producerAddress
	beaconBlock.Header.Version = BEACON_BLOCK_VERSION
	beaconBlock.Header.Height = beaconBestState.BeaconHeight + 1
	beaconBlock.Header.Epoch = epoch
	beaconBlock.Header.Round = round
	beaconBlock.Header.PreviousBlockHash = beaconBestState.BestBlockHash
	BLogger.log.Infof("Producing block: %d, %d", beaconBlock.Header.Height, beaconBlock.Header.Epoch)
	//=====END Build Header Essential Data=====
	//============Build body===================
	tempShardState, staker, swap, bridgeInstructions, acceptedRewardInstructions := blockGenerator.GetShardState(beaconBestState, shardsToBeaconLimit)
	tempInstruction := beaconBestState.GenerateInstruction(beaconBlock, staker, swap, beaconBestState.CandidateShardWaitingForCurrentRandom, bridgeInstructions, acceptedRewardInstructions)
	if len(rewardByEpochInstruction) != 0 {
		tempInstruction = append(tempInstruction, rewardByEpochInstruction...)
	}
	beaconBlock.Body.Instructions = tempInstruction
	beaconBlock.Body.ShardState = tempShardState
	if len(beaconBlock.Body.Instructions) != 0 {
		Logger.log.Info("Beacon Produce: Beacon Instruction", beaconBlock.Body.Instructions)
	}
	//============End Build Body================
	//============Build Header Hash=============
	// Process new block with beststate
	err = beaconBestState.updateBeaconBestState(beaconBlock)
	if err != nil {
		return nil, err
	}
	// calculate hash
	// BeaconValidator root: beacon committee + beacon pending committee
	validatorArr := append(beaconBestState.BeaconCommittee, beaconBestState.BeaconPendingValidator...)
	tempBeaconCommitteeAndValidatorRoot, err := GenerateHashFromStringArray(validatorArr)
	if err != nil {
		return nil, NewBlockChainError(GenerateBeaconCommitteeAndValidatorRootError, err)
	}
	// BeaconCandidate root: beacon current candidate + beacon next candidate
	beaconCandidateArr := append(beaconBestState.CandidateBeaconWaitingForCurrentRandom, beaconBestState.CandidateBeaconWaitingForNextRandom...)
	tempBeaconCandidateRoot, err := GenerateHashFromStringArray(beaconCandidateArr)
	if err != nil {
		return nil, NewBlockChainError(GenerateBeaconCandidateRootError, err)
	}
	// Shard candidate root: shard current candidate + shard next candidate
	shardCandidateArr := append(beaconBestState.CandidateShardWaitingForCurrentRandom, beaconBestState.CandidateShardWaitingForNextRandom...)
	tempShardCandidateRoot, err := GenerateHashFromStringArray(shardCandidateArr)
	if err != nil {
		return nil, NewBlockChainError(GenerateShardCandidateRootError, err)
	}
	// Shard Validator root
	tempShardCommitteeAndValidatorRoot, err := GenerateHashFromMapByteString(beaconBestState.GetShardPendingValidator(), beaconBestState.GetShardCommittee())
	if err != nil {
		return nil, NewBlockChainError(GenerateShardCommitteeAndValidatorRootError, err)
	}
	// Shard state hash
	tempShardStateHash, err := GenerateHashFromShardState(tempShardState)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewBlockChainError(GenerateShardStateError, err)
	}
	// Instruction Hash
	tempInstructionArr := []string{}
	for _, strs := range tempInstruction {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	tempInstructionHash, err := GenerateHashFromStringArray(tempInstructionArr)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewBlockChainError(GenerateInstructionError, err)
	}
	// Instruction merkle root
	flattenInsts, err := FlattenAndConvertStringInst(tempInstruction)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, err)
	}
	// add hash to header
	beaconBlock.Header.BeaconCommitteeAndValidatorRoot = tempBeaconCommitteeAndValidatorRoot
	beaconBlock.Header.BeaconCandidateRoot = tempBeaconCandidateRoot
	beaconBlock.Header.ShardCandidateRoot = tempShardCandidateRoot
	beaconBlock.Header.ShardCommitteeAndValidatorRoot = tempShardCommitteeAndValidatorRoot
	beaconBlock.Header.ShardStateHash = tempShardStateHash
	beaconBlock.Header.InstructionHash = tempInstructionHash
	copy(beaconBlock.Header.InstructionMerkleRoot[:], GetKeccak256MerkleRoot(flattenInsts))
	//============END Build Header Hash=========
	return beaconBlock, nil
}

func (blockGenerator *BlockGenerator) FinalizeBeaconBlock(blk *BeaconBlock, producerKeyset *incognitokey.KeySet) error {
	// Signature of producer, sign on hash of header
	blk.Header.Timestamp = time.Now().Unix()
	blockHash := blk.Header.Hash()
	producerSig, err := producerKeyset.SignDataInBase58CheckEncode(blockHash.GetBytes())
	if err != nil {
		Logger.log.Error(err)
		return NewBlockChainError(ProduceSignatureError, err)
	}
	blk.ProducerSig = producerSig
	//================End Generate Signature
	return nil
}

// return param:
// #1: shard state
// #2: valid stakers
// #3: swap validator => map[byte][][]string
func (blockGenerator *BlockGenerator) GetShardState(
	beaconBestState *BeaconBestState,
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
	allShardBlocks := blockGenerator.shardToBeaconPool.GetValidBlock(shardsToBeacon)
	//Shard block is a map ShardId -> array of shard block
	bridgeInstructions := [][]string{}
	acceptedRewardInstructions := [][]string{}
	var keys []int
	for k := range allShardBlocks {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, value := range keys {
		shardID := byte(value)
		shardBlocks := allShardBlocks[shardID]
		// Only accept block in one epoch
		totalBlock := 0
		//UNCOMMENT FOR TESTING
		Logger.log.Info("Beacon Producer Got These Block from pool", shardID)
		for _, shardBlocks := range shardBlocks {
			Logger.log.Infof(" %+v ", shardBlocks.Header.Height)
		}
		//=======
		for index, shardBlock := range shardBlocks {
			currentCommittee := beaconBestState.GetAShardCommittee(shardID)
			hash := shardBlock.Header.Hash()
			err1 := ValidateAggSignature(shardBlock.ValidatorsIdx, currentCommittee, shardBlock.AggregatedSig, shardBlock.R, &hash)
			Logger.log.Infof("Beacon Producer/ Validate Agg Signature for shard %+v, block height %+v, err %+v", shardID, shardBlock.Header.Height, err1 == nil)
			if index != 0 && err1 != nil {
				break
			}
			if err1 != nil {
				break
			}
			totalBlock = index
		}
		Logger.log.Infof("Beacon Producer/ AFTER FILTER, Shard %+v ONLY GET %+v block", shardID, totalBlock+1)
		if totalBlock > MAX_S2B_BLOCK {
			totalBlock = MAX_S2B_BLOCK
		}
		for _, shardBlock := range shardBlocks[:totalBlock+1] {
			shardState, validStaker, validSwapper, bridgeInstruction, acceptedRewardInstruction := blockGenerator.chain.GetShardStateFromBlock(beaconBestState, shardBlock, shardID)
			shardStates[shardID] = append(shardStates[shardID], shardState[shardID])
			validStakers = append(validStakers, validStaker...)
			validSwappers[shardID] = append(validSwappers[shardID], validSwapper[shardID]...)
			bridgeInstructions = append(bridgeInstructions, bridgeInstruction...)
			acceptedRewardInstructions = append(acceptedRewardInstructions, acceptedRewardInstruction)
		}
	}
	return shardStates, validStakers, validSwappers, bridgeInstructions, acceptedRewardInstructions
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
func (beaconBestState *BeaconBestState) GenerateInstruction(
	block *BeaconBlock,
	stakers [][]string,
	swap map[byte][][]string,
	shardCandidates []string,
	bridgeInstructions [][]string,
	acceptedRewardInstructions [][]string,
) [][]string {
	instructions := [][]string{}
	instructions = append(instructions, bridgeInstructions...)
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
		_, currentValidators, swappedValidator, beaconNextCommittee, _ := SwapValidator(beaconBestState.BeaconPendingValidator, beaconBestState.BeaconCommittee, beaconBestState.MaxBeaconCommitteeSize, common.OFFSET)
		if len(swappedValidator) > 0 || len(beaconNextCommittee) > 0 {
			swapBeaconInstructions = append(swapBeaconInstructions, "swap")
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(beaconNextCommittee, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(swappedValidator, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, "beacon")
			instructions = append(instructions, swapBeaconInstructions)

			// Generate instruction storing validators pubkey and send to bridge
			beaconRootInst := buildBeaconSwapConfirmInstruction(currentValidators, block.Header.Height+1)
			instructions = append(instructions, beaconRootInst)
		}
	}
	//=======Stake
	// ["stake", "pubkey.....", "shard" or "beacon"]
	instructions = append(instructions, stakers...)
	if block.Header.Height%common.EPOCH > common.RANDOM_TIME && !beaconBestState.IsGetRandomNumber {
		//=================================
		// COMMENT FOR TESTING
		//var err error
		//chainTimeStamp, err := beaconBestState.randomClient.GetCurrentChainTimeStamp()
		// UNCOMMENT FOR TESTING
		chainTimeStamp := beaconBestState.CurrentRandomTimeStamp + 1
		//==================================
		assignedCandidates := make(map[byte][]string)
		if chainTimeStamp > beaconBestState.CurrentRandomTimeStamp {
			randomInstruction, rand := beaconBestState.generateRandomInstruction(beaconBestState.CurrentRandomTimeStamp)
			instructions = append(instructions, randomInstruction)
			Logger.log.Info("RandomNumber", randomInstruction)
			for _, candidate := range shardCandidates {
				shardID := calculateCandidateShardID(candidate, rand, beaconBestState.ActiveShards)
				assignedCandidates[shardID] = append(assignedCandidates[shardID], candidate)
			}
			Logger.log.Infof("assignedCandidates %+v", assignedCandidates)
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

func (beaconBestState *BeaconBestState) GetValidStakers(tempStaker []string) []string {
	for _, committees := range beaconBestState.GetShardCommittee() {
		tempStaker = metadata.GetValidStaker(committees, tempStaker)
	}
	for _, validators := range beaconBestState.GetShardPendingValidator() {
		tempStaker = metadata.GetValidStaker(validators, tempStaker)
	}
	tempStaker = metadata.GetValidStaker(beaconBestState.BeaconCommittee, tempStaker)
	tempStaker = metadata.GetValidStaker(beaconBestState.BeaconPendingValidator, tempStaker)
	tempStaker = metadata.GetValidStaker(beaconBestState.CandidateBeaconWaitingForCurrentRandom, tempStaker)
	tempStaker = metadata.GetValidStaker(beaconBestState.CandidateBeaconWaitingForNextRandom, tempStaker)
	tempStaker = metadata.GetValidStaker(beaconBestState.CandidateShardWaitingForCurrentRandom, tempStaker)
	tempStaker = metadata.GetValidStaker(beaconBestState.CandidateShardWaitingForNextRandom, tempStaker)
	tempStaker = metadata.GetValidStaker(beaconBestState.CandidateShardWaitingForNextRandom, tempStaker)
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
	beaconBestState *BeaconBestState,
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
	bridgeInstructions := [][]string{}
	acceptedBlockRewardInfo := metadata.NewAcceptedBlockRewardInfo(shardID, shardBlock.Header.TotalTxsFee, shardBlock.Header.Height)
	acceptedRewardInstructions, err := acceptedBlockRewardInfo.GetStringFormat()
	if err != nil {
		panic("[ndh] Cant create acceptedRewardInstructions")
	}
	//Get Shard State from Block
	shardState := ShardState{}
	shardState.CrossShard = make([]byte, len(shardBlock.Header.CrossShardBitMap))
	copy(shardState.CrossShard, shardBlock.Header.CrossShardBitMap)
	shardState.Hash = shardBlock.Header.Hash()
	shardState.Height = shardBlock.Header.Height
	shardStates[shardID] = shardState

	instructions := shardBlock.Instructions
	Logger.log.Info(instructions)
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
		Logger.log.Infof("Instruction in shardBlock %+v, %+v \n", shardBlock.Header.Height, instructions)
	}
	for _, l := range instructions {
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
		Logger.log.Info("Beacon Producer/ Process Stakers List", stakers)
	}
	if len(swapers) != 0 {
		Logger.log.Info("Beacon Producer/ Process Stakers List", swapers)
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
	// Create bridge instruction
	if len(shardBlock.Instructions) > 0 || shardBlock.Header.Height%10 == 0 {
		BLogger.log.Debugf("Included shardID %d, block %d, insts: %s", shardID, shardBlock.Header.Height, shardBlock.Instructions)
	}
	bridgeInstructionForBlock, err := blockChain.buildBridgeInstructions(
		shardID,
		shardBlock.Instructions,
		beaconBestState,
		blockChain.config.DataBase,
	)
	if err != nil {
		BLogger.log.Errorf("Build bridge instructions failed: %s", err.Error())
	}

	// Pick instruction with shard committee's pubkeys to save to beacon block
	confirmInsts := pickBridgeSwapConfirmInst(shardBlock)
	if len(confirmInsts) > 0 {
		bridgeInstructionForBlock = append(bridgeInstructionForBlock, confirmInsts...)
		BLogger.log.Infof("Found bridge swap confirm inst: %s", confirmInsts)
	}

	bridgeInstructions = append(bridgeInstructions, bridgeInstructionForBlock...)
	Logger.log.Infof("Becon Produce: Got Shard Block %+v Shard %+v \n", shardBlock.Header.Height, shardID)
	return shardStates, validStakers, validSwap, bridgeInstructions, acceptedRewardInstructions
}

//===================================Util for Beacon=============================

// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
func (beaconBestState *BeaconBestState) generateRandomInstruction(timestamp int64) ([]string, int64) {
	//COMMENT FOR TESTING
	//var (
	//	blockHeight int
	//	chainTimestamp int64
	//	nonce int64
	//  strs []string
	//	err error
	//)
	//for {
	//	blockHeight, chainTimestamp, nonce, err = beaconBestState.randomClient.GetNonceByTimestamp(timestamp)
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
