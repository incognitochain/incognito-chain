package blockchain

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
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
func (blockGenerator *BlockGenerator) NewBlockBeacon(round int, shardsToBeacon map[byte]uint64) (*BeaconBlock, error) {
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
	BLogger.log.Infof("Producing block: %d (epoch %d)", beaconBlock.Header.Height, beaconBlock.Header.Epoch)
	//=====END Build Header Essential Data=====
	//============Build body===================
	tempShardState, staker, swap, bridgeInstructions, acceptedRewardInstructions := blockGenerator.GetShardState(beaconBestState, shardsToBeaconLimit)
	tempInstruction := beaconBestState.GenerateInstruction(beaconBlock.Header.Height, staker, swap, beaconBestState.CandidateShardWaitingForCurrentRandom, bridgeInstructions, acceptedRewardInstructions)
	if len(rewardByEpochInstruction) != 0 {
		tempInstruction = append(tempInstruction, rewardByEpochInstruction...)
	}
	beaconBlock.Body.Instructions = tempInstruction
	beaconBlock.Body.ShardState = tempShardState
	if len(beaconBlock.Body.Instructions) != 0 {
		Logger.log.Info("Beacon Produce: Beacon Instruction", beaconBlock.Body.Instructions)
	}
	if len(bridgeInstructions) > 0 {
		BLogger.log.Infof("Producer instructions: %+v", tempInstruction)
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
	tempBeaconCommitteeAndValidatorRoot, err := generateHashFromStringArray(validatorArr)
	if err != nil {
		return nil, NewBlockChainError(GenerateBeaconCommitteeAndValidatorRootError, err)
	}
	// BeaconCandidate root: beacon current candidate + beacon next candidate
	beaconCandidateArr := append(beaconBestState.CandidateBeaconWaitingForCurrentRandom, beaconBestState.CandidateBeaconWaitingForNextRandom...)
	tempBeaconCandidateRoot, err := generateHashFromStringArray(beaconCandidateArr)
	if err != nil {
		return nil, NewBlockChainError(GenerateBeaconCandidateRootError, err)
	}
	// Shard candidate root: shard current candidate + shard next candidate
	shardCandidateArr := append(beaconBestState.CandidateShardWaitingForCurrentRandom, beaconBestState.CandidateShardWaitingForNextRandom...)
	tempShardCandidateRoot, err := generateHashFromStringArray(shardCandidateArr)
	if err != nil {
		return nil, NewBlockChainError(GenerateShardCandidateRootError, err)
	}
	// Shard Validator root
	tempShardCommitteeAndValidatorRoot, err := generateHashFromMapByteString(beaconBestState.GetShardPendingValidator(), beaconBestState.GetShardCommittee())
	if err != nil {
		return nil, NewBlockChainError(GenerateShardCommitteeAndValidatorRootError, err)
	}
	// Shard state hash
	tempShardStateHash, err := generateHashFromShardState(tempShardState)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewBlockChainError(GenerateShardStateError, err)
	}
	// Instruction Hash
	tempInstructionArr := []string{}
	for _, strs := range tempInstruction {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	tempInstructionHash, err := generateHashFromStringArray(tempInstructionArr)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewBlockChainError(GenerateInstructionHashError, err)
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

// func (blockGenerator *BlockGenerator) FinalizeBeaconBlock(blk *BeaconBlock, producerKeyset *incognitokey.KeySet) error {
// 	// Signature of producer, sign on hash of header
// 	blk.Header.Timestamp = time.Now().Unix()
// 	blockHash := blk.Header.Hash()
// 	producerSig, err := producerKeyset.SignDataInBase58CheckEncode(blockHash.GetBytes())
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return err
// 	}
// 	blk.ProducerSig = producerSig
// 	//================End Generate Signature
// 	return nil
// }

// return param:
// #1: shard state
// #2: valid stakers
// #3: swap validator => map[byte][][]string
func (blockGenerator *BlockGenerator) GetShardState(beaconBestState *BeaconBestState, shardsToBeacon map[byte]uint64) (
	map[byte][]ShardState,
	[][]string,
	map[byte][][]string,
	[][]string,
	[][]string,
) {

	shardStates := make(map[byte][]ShardState)
	validStakeInstructions := [][]string{}
	validSwapInstructions := make(map[byte][][]string)
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
		Logger.log.Infof("Beacon Producer Got %+v Shard Block from shard %+v: ", len(shardBlocks), shardID)
		for _, shardBlocks := range shardBlocks {
			Logger.log.Infof(" %+v ", shardBlocks.Header.Height)
		}
		//=======
		for index, shardBlock := range shardBlocks {
			currentCommittee := beaconBestState.GetAShardCommittee(shardID)
			// hash := shardBlock.Header.Hash()
			err1 := blockGenerator.chain.config.ConsensusEngine.ValidateBlockCommitteSig(shardBlock.Hash(), currentCommittee, shardBlock.ValidationData, beaconBestState.ShardConsensusAlgorithm[shardID])
			Logger.log.Infof("Beacon Producer/ Validate Agg Signature for shard %+v, block height %+v, err %+v", shardID, shardBlock.Header.Height, err1 == nil)
			if err != nil {
				break
			}
			totalBlock = index
		}
		Logger.log.Infof("Beacon Producer/ AFTER FILTER, Shard %+v ONLY GET %+v block", shardID, totalBlock+1)
		if totalBlock > MAX_S2B_BLOCK {
			totalBlock = MAX_S2B_BLOCK
		}
		for _, shardBlock := range shardBlocks[:totalBlock+1] {
			shardState, validStakeInstruction, validSwapInstruction, bridgeInstruction, acceptedRewardInstruction := blockGenerator.chain.GetShardStateFromBlock(beaconBestState.BeaconHeight+1, shardBlock, shardID)
			shardStates[shardID] = append(shardStates[shardID], shardState[shardID])
			validStakeInstructions = append(validStakeInstructions, validStakeInstruction...)
			validSwapInstructions[shardID] = append(validSwapInstructions[shardID], validSwapInstruction[shardID]...)
			bridgeInstructions = append(bridgeInstructions, bridgeInstruction...)
			acceptedRewardInstructions = append(acceptedRewardInstructions, acceptedRewardInstruction)
		}
	}
	return shardStates, validStakeInstructions, validSwapInstructions, bridgeInstructions, acceptedRewardInstructions
}

/*
	- set instruction
	- del instruction
	- swap instruction
	+ format
	+ ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
	+ ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
	- random instruction
	- stake instruction
*/
func (beaconBestState *BeaconBestState) GenerateInstruction(
	newBeaconHeight uint64,
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
	if newBeaconHeight%uint64(common.EPOCH) == 0 {
		swapBeaconInstructions := []string{}
		_, currentValidators, swappedValidator, beaconNextCommittee, _ := SwapValidator(beaconBestState.BeaconPendingValidator, beaconBestState.BeaconCommittee, beaconBestState.MaxBeaconCommitteeSize, common.OFFSET)
		if len(swappedValidator) > 0 || len(beaconNextCommittee) > 0 {
			swapBeaconInstructions = append(swapBeaconInstructions, "swap")
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(beaconNextCommittee, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(swappedValidator, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, "beacon")
			instructions = append(instructions, swapBeaconInstructions)
			// Generate instruction storing validators pubkey and send to bridge
			beaconRootInst := buildBeaconSwapConfirmInstruction(currentValidators, newBeaconHeight)
			instructions = append(instructions, beaconRootInst)
		}
	}
	//=======Stake
	// ["stake", "pubkey.....", "shard" or "beacon"]
	instructions = append(instructions, stakers...)
	if newBeaconHeight%uint64(common.EPOCH) > uint64(common.RANDOM_TIME) && !beaconBestState.IsGetRandomNumber {
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
			Logger.log.Infof("Beacon Producer found Random Instruction at Block Height %+v", randomInstruction, newBeaconHeight)
			for _, candidate := range shardCandidates {
				shardID := calculateCandidateShardID(candidate, rand, beaconBestState.ActiveShards)
				assignedCandidates[shardID] = append(assignedCandidates[shardID], candidate)
			}
			for shardId, candidates := range assignedCandidates {
				Logger.log.Infof("Assign Candidate at Shard %+v: %+v", shardId, candidates)
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
	newBeaconHeight uint64,
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
	stakeInstructions := [][]string{}
	swapInstructions := make(map[byte][][]string)
	stakeInstructionFromShardBlock := [][]string{}
	swapInstructionFromShardBlock := [][]string{}
	bridgeInstructions := [][]string{}
	stakeBeacon := []string{}
	stakeShard := []string{}
	stakeBeaconTx := []string{}
	stakeShardTx := []string{}
	acceptedBlockRewardInfo := metadata.NewAcceptedBlockRewardInfo(shardID, shardBlock.Header.TotalTxsFee, shardBlock.Header.Height)
	acceptedRewardInstructions, err := acceptedBlockRewardInfo.GetStringFormat()
	if err != nil {
		panic("Can't create acceptedRewardInstructions")
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
	// extract instructions
	for _, l := range instructions {
		if len(l) > 0 {
			if l[0] == StakeAction {
				stakeInstructionFromShardBlock = append(stakeInstructionFromShardBlock, l)
			}
			if l[0] == SwapAction {
				swapInstructionFromShardBlock = append(swapInstructionFromShardBlock, l)
			}
		}
	}
	if len(stakeInstructionFromShardBlock) != 0 {
		Logger.log.Info("Beacon Producer/ Process Stakers List ", stakeInstructionFromShardBlock)
	}
	if len(swapInstructionFromShardBlock) != 0 {
		Logger.log.Info("Beacon Producer/ Process Stakers List ", swapInstructionFromShardBlock)
	}
	// Process Stake Instruction form Shard Block
	// Validate stake instruction => extract only valid stake instruction
	for _, stakePublicKey := range stakeInstructionFromShardBlock {
		var tempStakePublicKey []string
		newBeaconCandidate, newShardCandidate := getStakeValidatorArrayString(stakePublicKey)
		assignShard := true
		if !reflect.DeepEqual(newBeaconCandidate, []string{}) {
			tempStakePublicKey = make([]string, len(newBeaconCandidate))
			copy(tempStakePublicKey, newBeaconCandidate[:])
			assignShard = false
		} else {
			tempStakePublicKey = make([]string, len(newShardCandidate))
			copy(tempStakePublicKey, newShardCandidate[:])
		}
		tempStakePublicKey = blockChain.BestState.Beacon.GetValidStakers(tempStakePublicKey)
		tempStakePublicKey = metadata.GetValidStaker(stakeShard, tempStakePublicKey)
		tempStakePublicKey = metadata.GetValidStaker(stakeBeacon, tempStakePublicKey)
		if len(tempStakePublicKey) > 0 {
			if assignShard {
				stakeShard = append(stakeShard, tempStakePublicKey...)
				for i, v := range strings.Split(stakePublicKey[1], ",") {
					if common.IndexOfStr(v, stakeShard) > -1 {
						stakeShardTx = append(stakeShardTx, strings.Split(stakePublicKey[3], ",")[i])
					}
				}
			} else {
				stakeBeacon = append(stakeBeacon, tempStakePublicKey...)
				for i, v := range strings.Split(stakePublicKey[1], ",") {
					if common.IndexOfStr(v, stakeBeacon) > -1 {
						stakeBeaconTx = append(stakeBeaconTx, strings.Split(stakePublicKey[3], ",")[i])
					}
				}
			}
		}
	}
	if len(stakeShard) > 0 {
		stakeInstructions = append(stakeInstructions, []string{StakeAction, strings.Join(stakeShard, ","), "shard", strings.Join(stakeShardTx, ",")})
	}
	if len(stakeBeacon) > 0 {
		stakeInstructions = append(stakeInstructions, []string{StakeAction, strings.Join(stakeBeacon, ","), "beacon", strings.Join(stakeBeaconTx, ",")})
	}
	// Process Swap Instruction from Shard Block
	// Validate swap instruction => extract only valid swap instruction
	for _, swap := range swapInstructionFromShardBlock {
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
			swapInstructions[shardID] = append(swapInstructions[shardID], swap)
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
		newBeaconHeight,
		//beaconBestState,
		blockChain.config.DataBase,
	)
	if err != nil {
		BLogger.log.Errorf("Build bridge instructions failed: %s", err.Error())
	}
	// Pick instruction with shard committee's pubkeys to save to beacon block
	confirmInsts := pickBridgeSwapConfirmInst(shardBlock)
	if len(confirmInsts) > 0 {
		bridgeInstructionForBlock = append(bridgeInstructionForBlock, confirmInsts...)
		BLogger.log.Infof("Found bridge swap confirm inst in shard block %d: %s", shardBlock.Header.Height, confirmInsts)
	}
	bridgeInstructions = append(bridgeInstructions, bridgeInstructionForBlock...)
	Logger.log.Infof("Becon Produce: Got Shard Block %+v Shard %+v \n", shardBlock.Header.Height, shardID)
	return shardStates, stakeInstructions, swapInstructions, bridgeInstructions, acceptedRewardInstructions
}

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
