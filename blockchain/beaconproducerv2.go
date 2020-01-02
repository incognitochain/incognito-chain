package blockchain

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"sort"
	"time"
)

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

	beaconBlock.Header.Producer, err = committee[producerPosition].ToBase58() // .GetMiningKeyBase58(common.BridgeConsensus)
	if err != nil {
		return nil, err
	}
	beaconBlock.Header.ProducerPubKeyStr, err = committee[producerPosition].ToBase58()
	if err != nil {
		Logger.log.Error(err)
		return nil, NewBlockChainError(ConvertCommitteePubKeyToBase58Error, err)
	}
	beaconBlock.Header.Version = BEACON_BLOCK_VERSION
	beaconBlock.Header.Height = beaconBestState.BeaconHeight + 1
	beaconBlock.Header.Epoch = epoch
	beaconBlock.Header.Round = round
	beaconBlock.Header.PreviousBlockHash = beaconBestState.BestBlockHash
	BLogger.log.Infof("Producing block: %d (epoch %d)", beaconBlock.Header.Height, beaconBlock.Header.Epoch)
	//=====END Build Header Essential Data=====
	//============Build body===================
	if (beaconBestState.BeaconHeight+1)%blockGenerator.chain.config.ChainParams.Epoch == 1 {
		rewardByEpochInstruction, err = blockGenerator.chain.BuildRewardInstructionByEpoch(beaconBlock.Header.Height, beaconBestState.Epoch)
		if err != nil {
			return nil, NewBlockChainError(BuildRewardInstructionError, err)
		}
	}
	tempShardState, stakeInstructions, swapInstructions, bridgeInstructions, acceptedRewardInstructions, stopAutoStakingInstructions := blockGenerator.GetShardStateV2(beaconBestState, shardsToBeaconLimit)
	Logger.log.Infof("In NewBlockBeacon tempShardState: %+v", tempShardState)
	tempInstruction, err := beaconBestState.GenerateInstruction(
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
	err = beaconBestState.updateBeaconBestState(beaconBlock, blockGenerator.chain.config.ChainParams.Epoch, blockGenerator.chain.config.ChainParams.AssignOffset, blockGenerator.chain.config.ChainParams.RandomTime)
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
