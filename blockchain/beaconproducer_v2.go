package blockchain

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"time"
)

func (blockchain *BlockChain) NewBlockBeacon_V2(curView *BeaconBestState, version int, proposer string, round int, shardsToBeaconLimit map[byte]uint64) (newBlock *BeaconBlock, err error) {
	processState := &BeaconProcessState{
		curView:            curView,
		newView:            nil,
		blockchain:         blockchain,
		version:            version,
		proposer:           proposer,
		round:              round,
		newBlock:           NewBeaconBlock(),
		shardToBeaconBlock: make(map[byte][]*ShardToBeaconBlock),
	}

	if err := processState.PreProduceProcess(); err != nil {
		return nil, err
	}

	if err := processState.BuildBody(); err != nil {
		return nil, err
	}

	processState.newView, err = processState.curView.updateBeaconBestState(processState.newBlock, blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.AssignOffset, blockchain.config.ChainParams.RandomTime, newCommitteeChange())
	if err != nil {
		return nil, err
	}

	if err := processState.BuildHeader(); err != nil {
		return nil, err
	}

	return processState.newBlock, nil
}

type BeaconProcessState struct {
	//init state
	curView    *BeaconBestState
	newView    *BeaconBestState
	blockchain *BlockChain
	version    int
	proposer   string
	round      int

	//pre process state
	newBlock           *BeaconBlock
	shardToBeaconBlock map[byte][]*ShardToBeaconBlock
}

func (s *BeaconProcessState) PreProduceProcess() error {
	//get s2b blocks
	for sid, v := range s.blockchain.config.Syncker.GetS2BBlocksForBeaconProducer(s.curView.GetBestShardHash()) {
		for _, b := range v {
			s.shardToBeaconBlock[sid] = append(s.shardToBeaconBlock[sid], b.(*ShardToBeaconBlock))
		}
	}
	return nil
}

func (s *BeaconProcessState) BuildBody() (err error) {
	curView := s.curView

	//Reward by Epoch
	rewardByEpochInstruction := [][]string{}
	if (curView.BeaconHeight+1)%s.blockchain.config.ChainParams.Epoch == 1 {
		rewardByEpochInstruction, err = s.blockchain.buildRewardInstructionByEpoch(curView, curView.GetHeight()+1, curView.Epoch, curView.GetCopiedRewardStateDB())
		if err != nil {
			return NewBlockChainError(BuildRewardInstructionError, err)
		}
	}

	//Collect shard instruction
	tempShardState, stakeInstructions, swapInstructions, bridgeInstructions, acceptedRewardInstructions, stopAutoStakingInstructions := s.blockchain.GetShardState(curView, nil)
	Logger.log.Infof("In NewBlockBeacon tempShardState: %+v", tempShardState)

	//Generate beacon instruction
	tempInstruction, err := curView.GenerateInstruction(
		curView.GetHeight()+1, stakeInstructions, swapInstructions, stopAutoStakingInstructions,
		curView.CandidateShardWaitingForCurrentRandom, bridgeInstructions, acceptedRewardInstructions, s.blockchain.config.ChainParams.Epoch,
		s.blockchain.config.ChainParams.RandomTime, s.blockchain)
	if err != nil {
		return err
	}

	tempInstruction = append(tempInstruction, rewardByEpochInstruction...)

	//assign content to body
	s.newBlock.Body.Instructions = tempInstruction
	s.newBlock.Body.ShardState = tempShardState
	return
}

func (s *BeaconProcessState) BuildHeader() (err error) {

	//======Build Header Essential Data=======
	curView := s.curView
	s.newBlock.Header.Version = s.version
	s.newBlock.Header.Height = s.curView.BeaconHeight + 1
	epoch := s.curView.Epoch
	if (s.curView.BeaconHeight+1)%s.blockchain.config.ChainParams.Epoch == 1 {
		epoch = s.curView.Epoch + 1
	}
	s.newBlock.Header.ConsensusType = s.curView.ConsensusAlgorithm

	if s.version == 1 {
		committee := curView.GetBeaconCommittee()
		producerPosition := (curView.BeaconProposerIndex + s.round) % len(curView.BeaconCommittee)
		s.newBlock.Header.Producer, err = committee[producerPosition].ToBase58() // .GetMiningKeyBase58(common.BridgeConsensus)
		if err != nil {
			return err
		}
		s.newBlock.Header.ProducerPubKeyStr, err = committee[producerPosition].ToBase58()
		if err != nil {
			Logger.log.Error(err)
			return NewBlockChainError(ConvertCommitteePubKeyToBase58Error, err)
		}
	} else {
		s.newBlock.Header.Producer = s.proposer
		s.newBlock.Header.ProducerPubKeyStr = s.proposer
	}

	s.newBlock.Header.Height = curView.BeaconHeight + 1
	s.newBlock.Header.Epoch = epoch
	s.newBlock.Header.Round = s.round
	s.newBlock.Header.PreviousBlockHash = curView.BestBlockHash

	//============Build Header Hash=============
	newView := s.newView
	// calculate hash
	// BeaconValidator root: beacon committee + beacon pending committee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(newView.BeaconCommittee)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	validatorArr := append([]string{}, beaconCommitteeStr...)

	beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(newView.BeaconPendingValidator)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	validatorArr = append(validatorArr, beaconPendingValidatorStr...)
	tempBeaconCommitteeAndValidatorRoot, err := generateHashFromStringArray(validatorArr)
	if err != nil {
		return NewBlockChainError(GenerateBeaconCommitteeAndValidatorRootError, err)
	}
	// BeaconCandidate root: beacon current candidate + beacon next candidate
	beaconCandidateArr := append(newView.CandidateBeaconWaitingForCurrentRandom, newView.CandidateBeaconWaitingForNextRandom...)

	beaconCandidateArrStr, err := incognitokey.CommitteeKeyListToString(beaconCandidateArr)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	tempBeaconCandidateRoot, err := generateHashFromStringArray(beaconCandidateArrStr)
	if err != nil {
		return NewBlockChainError(GenerateBeaconCandidateRootError, err)
	}
	// Shard candidate root: shard current candidate + shard next candidate
	shardCandidateArr := append(newView.CandidateShardWaitingForCurrentRandom, newView.CandidateShardWaitingForNextRandom...)

	shardCandidateArrStr, err := incognitokey.CommitteeKeyListToString(shardCandidateArr)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	tempShardCandidateRoot, err := generateHashFromStringArray(shardCandidateArrStr)
	if err != nil {
		return NewBlockChainError(GenerateShardCandidateRootError, err)
	}
	// Shard Validator root
	shardPendingValidator := make(map[byte][]string)
	for shardID, keys := range newView.ShardPendingValidator {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}
		shardPendingValidator[shardID] = keysStr
	}
	shardCommittee := make(map[byte][]string)
	for shardID, keys := range newView.ShardCommittee {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}
		shardCommittee[shardID] = keysStr
	}
	tempShardCommitteeAndValidatorRoot, err := generateHashFromMapByteString(shardPendingValidator, shardCommittee)
	if err != nil {
		return NewBlockChainError(GenerateShardCommitteeAndValidatorRootError, err)
	}

	tempAutoStakingRoot, err := generateHashFromMapStringBool(newView.AutoStaking)
	if err != nil {
		return NewBlockChainError(AutoStakingRootHashError, err)
	}
	// Shard state hash
	tempShardStateHash, err := generateHashFromShardState(s.newBlock.Body.ShardState)
	if err != nil {
		Logger.log.Error(err)
		return NewBlockChainError(GenerateShardStateError, err)
	}
	// Instruction Hash
	tempInstructionArr := []string{}
	for _, strs := range s.newBlock.Body.Instructions {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	tempInstructionHash, err := generateHashFromStringArray(tempInstructionArr)
	if err != nil {
		Logger.log.Error(err)
		return NewBlockChainError(GenerateInstructionHashError, err)
	}
	// Instruction merkle root
	flattenInsts, err := FlattenAndConvertStringInst(s.newBlock.Body.Instructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, err)
	}
	// add hash to header
	s.newBlock.Header.BeaconCommitteeAndValidatorRoot = tempBeaconCommitteeAndValidatorRoot
	s.newBlock.Header.BeaconCandidateRoot = tempBeaconCandidateRoot
	s.newBlock.Header.ShardCandidateRoot = tempShardCandidateRoot
	s.newBlock.Header.ShardCommitteeAndValidatorRoot = tempShardCommitteeAndValidatorRoot
	s.newBlock.Header.ShardStateHash = tempShardStateHash
	s.newBlock.Header.InstructionHash = tempInstructionHash
	s.newBlock.Header.AutoStakingRoot = tempAutoStakingRoot
	copy(s.newBlock.Header.InstructionMerkleRoot[:], GetKeccak256MerkleRoot(flattenInsts))
	s.newBlock.Header.Timestamp = time.Now().Unix()
	return
}
