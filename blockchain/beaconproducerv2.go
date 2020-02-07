package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"sort"
	"strings"
	"time"
)

/*
	I. Block Produce:
	1. Clone Current Best State
	2. Build Essential Header Data:
		- Version: Get Proper version value
		- Height: Previous block height + 1
		- Epoch: Increase Epoch if next height mod epoch is 1 (begin of new epoch), otherwise use current epoch value
		- Round: Get Round Value from consensus
		- Previous Block Hash: Get Current Best Block Hash
		- Producer: Get producer value from round and current beacon committee
		- Consensus type: get from beaacon best state
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
func (blockGenerator *BlockGenerator) NewBlockBeaconV2(round int, shardsToBeaconLimit map[byte]uint64) (*BeaconBlock, error) {
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
	err = beaconBestState.cloneBeaconBestStateFrom(blockGenerator.chain.BestState.Beacon)
	if err != nil {
		return nil, err
	}
	//======Build Header Essential Data=======
	beaconBlock.Header.Version = BEACON_BLOCK_VERSION
	beaconBlock.Header.Height = beaconBestState.BeaconHeight + 1
	if (beaconBestState.BeaconHeight+1)%blockGenerator.chain.config.ChainParams.Epoch == 1 {
		epoch = beaconBestState.Epoch + 1
	} else {
		epoch = beaconBestState.Epoch
	}
	beaconBlock.Header.Epoch = epoch
	beaconBlock.Header.Round = round
	beaconBlock.Header.PreviousBlockHash = beaconBestState.BestBlockHash
	committee := blockGenerator.chain.BestState.Beacon.GetBeaconCommittee()
	producerPosition := (blockGenerator.chain.BestState.Beacon.BeaconProposerIndex + round) % len(beaconBestState.BeaconCommittee)
	beaconBlock.Header.Producer, err = committee[producerPosition].ToBase58() // .GetMiningKeyBase58(common.BridgeConsensus)
	if err != nil {
		return nil, err
	}
	beaconBlock.Header.ConsensusType = beaconBestState.ConsensusAlgorithm
	beaconBlock.Header.ProducerPubKeyStr, err = committee[producerPosition].ToBase58()
	if err != nil {
		Logger.log.Error(err)
		return nil, NewBlockChainError(ConvertCommitteePubKeyToBase58Error, err)
	}
	BLogger.log.Infof("Producing block: %d (epoch %d)", beaconBlock.Header.Height, beaconBlock.Header.Epoch)
	//=====END Build Header Essential Data=====
	//============Build body===================
	if (beaconBestState.BeaconHeight+1)%blockGenerator.chain.config.ChainParams.Epoch == 1 {
		rewardByEpochInstruction, err = blockGenerator.chain.BuildRewardInstructionByEpochV2(beaconBlock.Header.Height, beaconBestState.Epoch)
		if err != nil {
			return nil, NewBlockChainError(BuildRewardInstructionError, err)
		}
	}
	tempShardState, stakeInstructions, swapInstructions, bridgeInstructions, acceptedRewardInstructions, stopAutoStakingInstructions := blockGenerator.GetShardStateV2(beaconBestState, shardsToBeaconLimit)
	Logger.log.Infof("In NewBlockBeacon tempShardState: %+v", tempShardState)
	tempInstruction, err := beaconBestState.GenerateInstructionV2(
		beaconBlock.Header.Height, stakeInstructions, swapInstructions, stopAutoStakingInstructions,
		beaconBestState.CandidateShardWaitingForCurrentRandom, bridgeInstructions, acceptedRewardInstructions, blockGenerator.chain.config.ChainParams.Epoch,
		blockGenerator.chain.config.ChainParams.RandomTime, blockGenerator.chain,
	)
	if err != nil {
		return nil, err
	}
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
	//============Update Beacon Best State================
	// Process new block with beststate
	err = beaconBestState.updateBeaconBestStateV2(beaconBlock, blockGenerator.chain.config.ChainParams.Epoch, blockGenerator.chain.config.ChainParams.AssignOffset, blockGenerator.chain.config.ChainParams.RandomTime, newCommitteeChange())
	if err != nil {
		return nil, err
	}
	//============Build Header Hash=============
	// calculate hash
	// BeaconValidator root: beacon committee + beacon pending committee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	validatorArr := append([]string{}, beaconCommitteeStr...)

	beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	validatorArr = append(validatorArr, beaconPendingValidatorStr...)
	tempBeaconCommitteeAndValidatorRoot, err := generateHashFromStringArray(validatorArr)
	if err != nil {
		return nil, NewBlockChainError(GenerateBeaconCommitteeAndValidatorRootError, err)
	}
	// BeaconCandidate root: beacon current candidate + beacon next candidate
	beaconCandidateArr := append(beaconBestState.CandidateBeaconWaitingForCurrentRandom, beaconBestState.CandidateBeaconWaitingForNextRandom...)
	beaconCandidateArrStr, err := incognitokey.CommitteeKeyListToString(beaconCandidateArr)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	tempBeaconCandidateRoot, err := generateHashFromStringArray(beaconCandidateArrStr)
	if err != nil {
		return nil, NewBlockChainError(GenerateBeaconCandidateRootError, err)
	}
	// Shard candidate root: shard current candidate + shard next candidate
	shardCandidateArr := append(beaconBestState.CandidateShardWaitingForCurrentRandom, beaconBestState.CandidateShardWaitingForNextRandom...)

	shardCandidateArrStr, err := incognitokey.CommitteeKeyListToString(shardCandidateArr)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	tempShardCandidateRoot, err := generateHashFromStringArray(shardCandidateArrStr)
	if err != nil {
		return nil, NewBlockChainError(GenerateShardCandidateRootError, err)
	}
	// Shard Validator root
	shardPendingValidator := make(map[byte][]string)
	for shardID, keys := range beaconBestState.ShardPendingValidator {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, NewBlockChainError(UnExpectedError, err)
		}
		shardPendingValidator[shardID] = keysStr
	}
	shardCommittee := make(map[byte][]string)
	for shardID, keys := range beaconBestState.ShardCommittee {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, NewBlockChainError(UnExpectedError, err)
		}
		shardCommittee[shardID] = keysStr
	}
	tempShardCommitteeAndValidatorRoot, err := generateHashFromMapByteString(shardPendingValidator, shardCommittee)
	if err != nil {
		return nil, NewBlockChainError(GenerateShardCommitteeAndValidatorRootError, err)
	}

	tempAutoStakingRoot, err := generateHashFromMapStringBool(beaconBestState.AutoStaking)
	if err != nil {
		return nil, NewBlockChainError(AutoStakingRootHashError, err)
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
	beaconBlock.Header.AutoStakingRoot = tempAutoStakingRoot
	copy(beaconBlock.Header.InstructionMerkleRoot[:], GetKeccak256MerkleRoot(flattenInsts))
	beaconBlock.Header.Timestamp = time.Now().Unix()
	//============END Build Header Hash=========
	return beaconBlock, nil
}

func (blockGenerator *BlockGenerator) GetShardStateV2(beaconBestState *BeaconBestState, shardsToBeacon map[byte]uint64) (map[byte][]ShardState, [][]string, map[byte][][]string, [][]string, [][]string, [][]string) {
	shardStates := make(map[byte][]ShardState)
	validStakeInstructions := [][]string{}
	validStakePublicKeys := []string{}
	validStopAutoStakingInstructions := [][]string{}
	validSwapInstructions := make(map[byte][][]string)
	//Get shard to beacon block from pool
	Logger.log.Infof("In GetShardState shardsToBeacon limit: %+v", shardsToBeacon)
	allShardBlocks := blockGenerator.shardToBeaconPool.GetValidBlock(shardsToBeacon)
	Logger.log.Infof("In GetShardState allShardBlocks: %+v", allShardBlocks)
	//Shard block is a map ShardId -> array of shard block
	bridgeInstructions := [][]string{}
	acceptedRewardInstructions := [][]string{}
	statefulActionsByShardID := map[byte][][]string{}
	var keys []int
	for k := range allShardBlocks {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	Logger.log.Infof("In GetShardState keys: %+v", keys)
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
			if index == MAX_S2B_BLOCK-1 {
				break
			}
			err := blockGenerator.chain.config.ConsensusEngine.ValidateBlockCommitteSig(shardBlock, currentCommittee, beaconBestState.ShardConsensusAlgorithm[shardID])
			Logger.log.Infof("Beacon Producer/ Validate Agg Signature for shard %+v, block height %+v, err %+v", shardID, shardBlock.Header.Height, err == nil)
			if err != nil {
				break
			}
			totalBlock = index
			if totalBlock > MAX_S2B_BLOCK {
				totalBlock = MAX_S2B_BLOCK
				break
			}
		}
		Logger.log.Infof("Beacon Producer/ AFTER FILTER, Shard %+v ONLY GET %+v block", shardID, totalBlock+1)
		for _, shardBlock := range shardBlocks[:totalBlock+1] {
			shardState, validStakeInstruction, tempValidStakePublicKeys, validSwapInstruction, bridgeInstruction, acceptedRewardInstruction, stopAutoStakingInstruction, statefulActions := blockGenerator.chain.GetShardStateFromBlock(beaconBestState.BeaconHeight+1, shardBlock, shardID, true, validStakePublicKeys)
			shardStates[shardID] = append(shardStates[shardID], shardState[shardID])
			validStakeInstructions = append(validStakeInstructions, validStakeInstruction...)
			validSwapInstructions[shardID] = append(validSwapInstructions[shardID], validSwapInstruction[shardID]...)
			bridgeInstructions = append(bridgeInstructions, bridgeInstruction...)
			acceptedRewardInstructions = append(acceptedRewardInstructions, acceptedRewardInstruction)
			validStopAutoStakingInstructions = append(validStopAutoStakingInstructions, stopAutoStakingInstruction...)
			validStakePublicKeys = append(validStakePublicKeys, tempValidStakePublicKeys...)

			// group stateful actions by shardID
			_, found := statefulActionsByShardID[shardID]
			if !found {
				statefulActionsByShardID[shardID] = statefulActions
			} else {
				statefulActionsByShardID[shardID] = append(statefulActionsByShardID[shardID], statefulActions...)
			}
		}
	}
	// build stateful instructions
	statefulInsts := blockGenerator.chain.buildStatefulInstructionsV2(beaconBestState.featureStateDB, statefulActionsByShardID, beaconBestState.BeaconHeight+1)
	bridgeInstructions = append(bridgeInstructions, statefulInsts...)
	return shardStates, validStakeInstructions, validSwapInstructions, bridgeInstructions, acceptedRewardInstructions, validStopAutoStakingInstructions
}

/*
	- swap instruction
	format
	+ ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
	+ ["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
	- random instruction
	- stake instruction
	+ ["stake", "pubkey1,pubkey2,..." "shard" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,...", "flag1,flag2..."]
	+ ["stake", "pubkey1,pubkey2,..." "beacon" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,...", "flag1,flag2..."]
	- assign instruction
	+ ["assign" "shardCandidate1,shardCandidate2,..." "shard" "{shardID}"]
*/
func (beaconBestState *BeaconBestState) GenerateInstructionV2(
	newBeaconHeight uint64,
	stakeInstructions [][]string,
	swapInstructions map[byte][][]string,
	stopAutoStakingInstructions [][]string,
	shardCandidates []incognitokey.CommitteePublicKey,
	bridgeInstructions [][]string,
	acceptedRewardInstructions [][]string,
	chainParamEpoch uint64,
	randomTime uint64,
	blockchain *BlockChain,
) ([][]string, error) {
	instructions := [][]string{}
	instructions = append(instructions, bridgeInstructions...)
	instructions = append(instructions, acceptedRewardInstructions...)
	//=======Swap
	// Shard Swap: both abnormal or normal swap
	var keys []int
	for k := range swapInstructions {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		instructions = append(instructions, swapInstructions[byte(shardID)]...)
	}
	// Beacon normal swap
	if newBeaconHeight%uint64(chainParamEpoch) == 0 {
		swapBeaconInstructions := []string{}
		beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator)
		if err != nil {
			return [][]string{}, err
		}
		beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)
		if err != nil {
			return [][]string{}, err
		}
		rootHash, err := blockchain.GetBeaconSlashStateRootHash(blockchain.GetDatabase(), newBeaconHeight-1)
		if err != nil {
			return [][]string{}, err
		}
		slashStateDB, err := statedb.NewWithPrefixTrie(rootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
		producersBlackList, err := blockchain.getUpdatedProducersBlackListV2(slashStateDB, true, -1, beaconCommitteeStr, newBeaconHeight-1)
		if err != nil {
			Logger.log.Error(err)
		}
		badProducersWithPunishment := blockchain.buildBadProducersWithPunishment(true, -1, beaconCommitteeStr)
		badProducersWithPunishmentBytes, err := json.Marshal(badProducersWithPunishment)
		if err != nil {
			Logger.log.Error(err)
		}
		_, currentValidators, swappedValidator, beaconNextCommittee, err := SwapValidator(beaconPendingValidatorStr, beaconCommitteeStr, beaconBestState.MaxBeaconCommitteeSize, beaconBestState.MinBeaconCommitteeSize, blockchain.config.ChainParams.Offset, producersBlackList, blockchain.config.ChainParams.SwapOffset)
		if len(swappedValidator) > 0 || len(beaconNextCommittee) > 0 && err == nil {
			swapBeaconInstructions = append(swapBeaconInstructions, "swap")
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(beaconNextCommittee, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, strings.Join(swappedValidator, ","))
			swapBeaconInstructions = append(swapBeaconInstructions, "beacon")
			swapBeaconInstructions = append(swapBeaconInstructions, string(badProducersWithPunishmentBytes))
			instructions = append(instructions, swapBeaconInstructions)
			// Generate instruction storing validators pubkey and send to bridge
			beaconRootInst, _ := buildBeaconSwapConfirmInstruction(currentValidators, newBeaconHeight)
			instructions = append(instructions, beaconRootInst)
		}
	}
	// Stake
	instructions = append(instructions, stakeInstructions...)
	// Stop Auto Staking
	instructions = append(instructions, stopAutoStakingInstructions...)
	// Random number for Assign Instruction
	if newBeaconHeight%chainParamEpoch > randomTime && !beaconBestState.IsGetRandomNumber {
		var err error
		var chainTimeStamp int64
		if !TestRandom {
			if newBeaconHeight%chainParamEpoch == chainParamEpoch-1 {
				startTime := time.Now()
				for {
					Logger.log.Criticalf("Block %+v, Enter final block of epoch but still no random number", newBeaconHeight)
					chainTimeStamp, err = blockchain.config.RandomClient.GetCurrentChainTimeStamp()
					if err != nil {
						Logger.log.Error(err)
					} else {
						if chainTimeStamp < beaconBestState.CurrentRandomTimeStamp {
							Logger.log.Infof("Final Block %+v in Epoch but still haven't found new random number", newBeaconHeight)
						} else {
							break
						}
					}
					if time.Since(startTime).Seconds() > beaconBestState.BlockMaxCreateTime.Seconds() {
						return [][]string{}, NewBlockChainError(GenerateInstructionError, fmt.Errorf("Get Current Chain Timestamp for New Block Height %+v Timeout", newBeaconHeight))
					}
					time.Sleep(100 * time.Millisecond)
				}
			} else {
				Logger.log.Criticalf("Block %+v, finding random number", newBeaconHeight)
				chainTimeStamp, err = blockchain.config.RandomClient.GetCurrentChainTimeStamp()
				if err != nil {
					Logger.log.Error(err)
				}
			}
		} else {
			chainTimeStamp = beaconBestState.CurrentRandomTimeStamp + 1
		}
		//==================================
		if err == nil && chainTimeStamp > beaconBestState.CurrentRandomTimeStamp {
			numberOfPendingValidator := make(map[byte]int)
			for i := 0; i < beaconBestState.ActiveShards; i++ {
				if pendingValidators, ok := beaconBestState.ShardPendingValidator[byte(i)]; ok {
					numberOfPendingValidator[byte(i)] = len(pendingValidators)
				} else {
					numberOfPendingValidator[byte(i)] = 0
				}
			}
			randomInstruction, rand, err := beaconBestState.generateRandomInstruction(beaconBestState.CurrentRandomTimeStamp, blockchain.config.RandomClient)
			if err != nil {
				return [][]string{}, err
			}
			instructions = append(instructions, randomInstruction)
			Logger.log.Infof("Beacon Producer found Random Instruction at Block Height %+v, %+v", randomInstruction, newBeaconHeight)
			shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(shardCandidates)
			if err != nil {
				panic(err)
			}
			_, assignedCandidates := assignShardCandidate(shardCandidatesStr, numberOfPendingValidator, rand, blockchain.config.ChainParams.AssignOffset, beaconBestState.ActiveShards)
			var keys []int
			for k := range assignedCandidates {
				keys = append(keys, int(k))
			}
			sort.Ints(keys)
			for _, key := range keys {
				shardID := byte(key)
				candidates := assignedCandidates[shardID]
				Logger.log.Infof("Assign Candidate at Shard %+v: %+v", shardID, candidates)
				shardAssingInstruction := []string{AssignAction}
				shardAssingInstruction = append(shardAssingInstruction, strings.Join(candidates, ","))
				shardAssingInstruction = append(shardAssingInstruction, "shard")
				shardAssingInstruction = append(shardAssingInstruction, fmt.Sprintf("%v", shardID))
				instructions = append(instructions, shardAssingInstruction)
			}
		}
	}
	return instructions, nil
}
