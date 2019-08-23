package blockchain

import (
	"fmt"
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
	I. Block Produce:
	1. Clone Current Best State
	2. Build Essential Header Data:
		a. Producer: Get Producer Address value from input parameters
		b. Version: Get Proper version value
		c. Epoch: Increase Epoch if next height mod epoch is 1 (begin of new epoch), otherwise use current epoch value
		d. Height: Previous block height + 1
		e. Round: Get Round Value from consensus
		f. Previous Block Hash: Get Current Best Block Hash
	3. Build Body:
		a. Build Reward Instruction:
			- These instruction will only be built at the begining of each epoch (for previous committee)
		b. Get Shard State and Instruction:
			- These information will be extracted from all shard block, which got from shard to beacon pool
		c. Create Instruction:
			- Instruction created from beacon data
			- Instruction created from shard instructions
	4. Update Cloned Beacon Best State to Build Root Hash for Header
		+ Beacon Root Hash will be calculated from new beacon best state (beacon best state after process by this new block)
		+ Some data may changed if beacon best state is updated:
			+ Beacon Committee, Pending Validator, Candidate List
			+ Shard Committee, Pending Validator, Candidate List
	5. Build Root Hash in Header
		a. Beacon Committee and Validator Root Hash: Hash from Beacon Committee and Pending Validator
		b. Beacon Caiddate Root Hash: Hash from Beacon candidate list
		c. Shard Committee and Validator Root Hash: Hash from Shard Committee and Pending Validator
		d. Shard Caiddate Root Hash: Hash from Shard candidate list
		+ These Root Hash will be used to verify that, either Two arbitray Nodes have the same data
			after they update beacon best state by new block.
		e. ShardStateHash: shard states from blocks of all shard
		f. InstructionHash: from instructions in beacon block body
		g. InstructionMerkleRoot
	II. Block Finalize:
	1. Add Block Timestamp
	2. Calculate block Producer Signature
		+ Block Producer Signature is calculated from hash block header
		+ Block Producer Signature is not included in block header
*/
func (blockGenerator *BlockGenerator) NewBlockBeacon(round int, shardsToBeaconLimit map[byte]uint64) (*BeaconBlock, error) {
	// lock blockchain
	blockGenerator.chain.chainLock.Lock()
	defer blockGenerator.chain.chainLock.Unlock()
	Logger.log.Infof("⛏ Creating Beacon Block %+v", blockGenerator.chain.BestState.Beacon.BeaconHeight+1)
	//============Init Variable============
	var err error
	var epoch uint64
	beaconBlock := NewBeaconBlock()
	beaconBestState := NewBeaconBestState()
	rewardByEpochInstruction := [][]string{}
	// produce new block with current beststate
	err = beaconBestState.cloneBeaconBestState(blockGenerator.chain.BestState.Beacon)
	if err != nil {
		return nil, err
	}
	beaconBestState.InitRandomClient(blockGenerator.chain.config.RandomClient)
	//======Build Header Essential Data=======
	// beaconBlock.Header.ProducerAddress = *producerAddress
	beaconBlock.Header.Version = BEACON_BLOCK_VERSION
	beaconBlock.Header.Height = beaconBestState.BeaconHeight + 1
	if (beaconBestState.BeaconHeight+1)%blockGenerator.chain.config.ChainParams.Epoch == 1 {
		epoch = beaconBestState.Epoch + 1
	} else {
		epoch = beaconBestState.Epoch
	}
	committee := blockGenerator.chain.BestState.Beacon.GetBeaconCommittee()
	producerPosition := (blockGenerator.chain.BestState.Beacon.BeaconProposerIndex + round) % len(beaconBestState.BeaconCommittee)
	beaconBlock.Header.ConsensusType = beaconBestState.ConsensusAlgorithm
	beaconBlock.Header.Producer = committee[producerPosition].GetMiningKeyBase58(beaconBestState.ConsensusAlgorithm)
	beaconBlock.Header.Version = BEACON_BLOCK_VERSION
	beaconBlock.Header.Height = beaconBestState.BeaconHeight + 1
	beaconBlock.Header.Epoch = epoch
	beaconBlock.Header.Round = round
	beaconBlock.Header.PreviousBlockHash = beaconBestState.BestBlockHash
	BLogger.log.Infof("Producing block: %d (epoch %d)", beaconBlock.Header.Height, beaconBlock.Header.Epoch)
	//=====END Build Header Essential Data=====
	//============Build body===================
	if (beaconBestState.BeaconHeight+1)%blockGenerator.chain.config.ChainParams.Epoch == 1 {
		rewardByEpochInstruction, err = blockGenerator.chain.BuildRewardInstructionByEpoch(beaconBestState.Epoch)
		if err != nil {
			return nil, NewBlockChainError(BuildRewardInstructionError, err)
		}
	}
	tempShardState, staker, swap, bridgeInstructions, acceptedRewardInstructions := blockGenerator.GetShardState(beaconBestState, shardsToBeaconLimit)
	tempInstruction := beaconBestState.GenerateInstruction(beaconBlock.Header.Height, staker, swap, beaconBestState.CandidateShardWaitingForCurrentRandom,
		bridgeInstructions, acceptedRewardInstructions, blockGenerator.chain.config.ChainParams.Epoch, blockGenerator.chain.config.ChainParams.RandomTime)
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
	err = beaconBestState.updateBeaconBestState(beaconBlock, blockGenerator.chain.config.ChainParams.Epoch, blockGenerator.chain.config.ChainParams.RandomTime)
	if err != nil {
		return nil, err
	}
	// calculate hash
	// BeaconValidator root: beacon committee + beacon pending committee
	validatorArr := append(beaconBestState.BeaconCommittee, beaconBestState.BeaconPendingValidator...)
	tempBeaconCommitteeAndValidatorRoot, err := generateHashFromStringArray(incognitokey.CommitteeKeyListToString(validatorArr))
	if err != nil {
		return nil, NewBlockChainError(GenerateBeaconCommitteeAndValidatorRootError, err)
	}
	// BeaconCandidate root: beacon current candidate + beacon next candidate
	beaconCandidateArr := append(beaconBestState.CandidateBeaconWaitingForCurrentRandom, beaconBestState.CandidateBeaconWaitingForNextRandom...)
	tempBeaconCandidateRoot, err := generateHashFromStringArray(incognitokey.CommitteeKeyListToString(beaconCandidateArr))
	if err != nil {
		return nil, NewBlockChainError(GenerateBeaconCandidateRootError, err)
	}
	// Shard candidate root: shard current candidate + shard next candidate
	shardCandidateArr := append(beaconBestState.CandidateShardWaitingForCurrentRandom, beaconBestState.CandidateShardWaitingForNextRandom...)
	tempShardCandidateRoot, err := generateHashFromStringArray(incognitokey.CommitteeKeyListToString(shardCandidateArr))
	if err != nil {
		return nil, NewBlockChainError(GenerateShardCandidateRootError, err)
	}
	// Shard Validator root
	shardPendingValidator := make(map[byte][]string)
	for shardID, keys := range beaconBestState.ShardPendingValidator {
		shardPendingValidator[shardID] = incognitokey.CommitteeKeyListToString(keys)
	}

	shardCommittee := make(map[byte][]string)
	for shardID, keys := range beaconBestState.ShardCommittee {
		shardCommittee[shardID] = incognitokey.CommitteeKeyListToString(keys)
	}

	tempShardCommitteeAndValidatorRoot, err := generateHashFromMapByteString(shardPendingValidator, shardCommittee)
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
	beaconBlock.Header.Timestamp = time.Now().Unix()
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
// #2: valid stake instruction
// #3: valid swap instruction
// #4: bridge instructions
// #5: accepted reward instructions
/*
	Get Shard To Beacon Block Rule:
	1. Shard To Beacon Blocks will be get from Shard To Beacon Pool (only valid block)
	2. Process shards independently, for each shard:
		a. Shard To Beacon Block List must be compatible with current shard state in beacon best state:
			+ Increased continuosly in height (10, 11, 12,...)
				Ex: Shard state in beacon best state has height 11 then shard to beacon block list must have first block in list with height 12
			+ Shard To Beacon Block List must have incremental height in list (10, 11, 12,... NOT 10, 12,...)
			+ Shard To Beacon Block List can be verify with and only with current shard committee in beacon best state
			+ DO NOT accept Shard To Beacon Block List that can have two arbitrary blocks that can be verify with two different committee set
			+ If in Shard To Beacon Block List have one block with Swap Instruction, then this block must be the last block in this list (or only block in this list)
*/
func (blockGenerator *BlockGenerator) GetShardState(beaconBestState *BeaconBestState, shardsToBeacon map[byte]uint64) (map[byte][]ShardState, [][]string, map[byte][][]string, [][]string, [][]string) {
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
		currentCommittee := beaconBestState.GetAShardCommittee(shardID)
		for index, shardBlock := range shardBlocks {
			// hash := shardBlock.Header.Hash()
			err := blockGenerator.chain.config.ConsensusEngine.ValidateBlockCommitteSig(shardBlock, currentCommittee, beaconBestState.ShardConsensusAlgorithm[shardID])
			Logger.log.Infof("Beacon Producer/ Validate Agg Signature for shard %+v, block height %+v, err %+v", shardID, shardBlock.Header.Height, err == nil)
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
	shardCandidates []incognitokey.CommitteePublicKey,
	bridgeInstructions [][]string,
	acceptedRewardInstructions [][]string,
	chainParamEpoch uint64,
	randomTime uint64,
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

	if newBeaconHeight%uint64(chainParamEpoch) == 0 {
		swapBeaconInstructions := []string{}
		_, currentValidators, swappedValidator, beaconNextCommittee, _ := SwapValidator(incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator), incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee), beaconBestState.MaxBeaconCommitteeSize, common.OFFSET)
		if len(swappedValidator) > 0 || len(beaconNextCommittee) > 0 {
			swapBeaconInstructions = append(swapBeaconInstructions, "swap")
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(beaconNextCommittee, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(
				swappedValidator, ","))
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
	if newBeaconHeight%uint64(chainParamEpoch) > randomTime && !beaconBestState.IsGetRandomNumber {
		//=================================
		// COMMENT FOR TESTING
		//var err error
		//chainTimeStamp, err := beaconBestState.randomClient.GetCurrentChainTimeStamp()
		// UNCOMMENT FOR TESTING
		chainTimeStamp := beaconBestState.CurrentRandomTimeStamp + 1
		//==================================
		assignedCandidates := make(map[byte][]incognitokey.CommitteePublicKey)
		if chainTimeStamp > beaconBestState.CurrentRandomTimeStamp {
			randomInstruction, rand := beaconBestState.generateRandomInstruction(beaconBestState.CurrentRandomTimeStamp)
			instructions = append(instructions, randomInstruction)
			Logger.log.Infof("Beacon Producer found Random Instruction at Block Height %+v", randomInstruction, newBeaconHeight)
			for _, candidate := range shardCandidates {
				candidateStr, _ := candidate.ToBase58()
				shardID := calculateCandidateShardID(candidateStr, rand, beaconBestState.ActiveShards)
				assignedCandidates[shardID] = append(assignedCandidates[shardID], candidate)
			}
			for shardId, candidates := range assignedCandidates {
				candidatesStr := incognitokey.CommitteeKeyListToString(candidates)
				Logger.log.Infof("Assign Candidate at Shard %+v: %+v", shardId, candidatesStr)
				shardAssingInstruction := []string{"assign"}
				shardAssingInstruction = append(shardAssingInstruction, strings.Join(candidatesStr, ","))
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
		tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(committees), tempStaker)
	}
	for _, validators := range beaconBestState.GetShardPendingValidator() {
		tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(validators), tempStaker)
	}
	tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee), tempStaker)
	tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator), tempStaker)
	tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForCurrentRandom), tempStaker)
	tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForNextRandom), tempStaker)
	tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForCurrentRandom), tempStaker)
	tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForNextRandom), tempStaker)
	tempStaker = common.GetValidStaker(incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForNextRandom), tempStaker)
	return tempStaker
}

/*
	Swap format:
	- ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
	- ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
	Stake format:
	- ["stake" "pubkey1,pubkey2,..." "shard" "txStakeHash1, txStakeHash2,..." "txStakeRewardReceiver1, txStakeRewardReceiver2,..."]
	- ["stake" "pubkey1,pubkey2,..." "beacon" "txStakeHash1, txStakeHash2,..." "txStakeRewardReceiver1, txStakeRewardReceiver2,..."]

*/
func (blockchain *BlockChain) GetShardStateFromBlock(newBeaconHeight uint64, shardBlock *ShardToBeaconBlock, shardID byte) (map[byte]ShardState, [][]string, map[byte][][]string, [][]string, []string) {
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
	stakeShardRewardReceiver := []string{}
	stakeBeaconRewardReceiver := []string{}
	acceptedBlockRewardInfo := metadata.NewAcceptedBlockRewardInfo(shardID, shardBlock.Header.TotalTxsFee, shardBlock.Header.Height)
	acceptedRewardInstructions, err := acceptedBlockRewardInfo.GetStringFormat()
	if err != nil {
		// if err then ignore accepted reward instruction
		acceptedRewardInstructions = []string{}
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
				//- ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
				//- ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
				// only allow shard to swap committee for it self
				if l[3] == "shard" && len(l) != 5 && l[4] != strconv.Itoa(int(shardID)) {
					continue
				}
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
	for _, stakeInstruction := range stakeInstructionFromShardBlock {
		var tempStakePublicKey []string
		newBeaconCandidate, newShardCandidate := getStakeValidatorArrayString(stakeInstruction)
		assignShard := true
		if !reflect.DeepEqual(newBeaconCandidate, []string{}) {
			tempStakePublicKey = make([]string, len(newBeaconCandidate))
			copy(tempStakePublicKey, newBeaconCandidate[:])
			assignShard = false
		} else {
			tempStakePublicKey = make([]string, len(newShardCandidate))
			copy(tempStakePublicKey, newShardCandidate[:])
		}
		if len(tempStakePublicKey) != len(strings.Split(stakeInstruction[3], ",")) && len(strings.Split(stakeInstruction[3], ",")) != len(strings.Split(stakeInstruction[4], ",")) {
			continue
		}
		tempStakePublicKey = blockchain.BestState.Beacon.GetValidStakers(tempStakePublicKey)
		tempStakePublicKey = common.GetValidStaker(stakeShard, tempStakePublicKey)
		tempStakePublicKey = common.GetValidStaker(stakeBeacon, tempStakePublicKey)
		if len(tempStakePublicKey) > 0 {
			if assignShard {
				stakeShard = append(stakeShard, tempStakePublicKey...)
				for i, v := range strings.Split(stakeInstruction[1], ",") {
					if common.IndexOfStr(v, tempStakePublicKey) > -1 {
						stakeShardTx = append(stakeShardTx, strings.Split(stakeInstruction[3], ",")[i])
						stakeShardRewardReceiver = append(stakeShardRewardReceiver, strings.Split(stakeInstruction[4], ",")[i])
					}
				}
			} else {
				stakeBeacon = append(stakeBeacon, tempStakePublicKey...)
				for i, v := range strings.Split(stakeInstruction[1], ",") {
					if common.IndexOfStr(v, tempStakePublicKey) > -1 {
						stakeBeaconTx = append(stakeBeaconTx, strings.Split(stakeInstruction[3], ",")[i])
						stakeBeaconRewardReceiver = append(stakeBeaconRewardReceiver, strings.Split(stakeInstruction[4], ",")[i])
					}
				}
			}
		}
	}
	if len(stakeShard) > 0 {
		stakeInstructions = append(stakeInstructions, []string{StakeAction, strings.Join(stakeShard, ","), "shard", strings.Join(stakeShardTx, ","), strings.Join(stakeShardRewardReceiver, ",")})
	}
	if len(stakeBeacon) > 0 {
		stakeInstructions = append(stakeInstructions, []string{StakeAction, strings.Join(stakeBeacon, ","), "beacon", strings.Join(stakeBeaconTx, ","), strings.Join(stakeBeaconRewardReceiver, ",")})
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
	bridgeInstructionForBlock, err := blockchain.buildBridgeInstructions(
		shardID,
		shardBlock.Instructions,
		newBeaconHeight,
		//beaconBestState,
		blockchain.config.DataBase,
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
