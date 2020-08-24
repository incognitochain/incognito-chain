package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/instructionsprocessor"

	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV2 struct {
	blockHeight uint64
	blockHash   common.Hash

	beaconCommitteePool    []incognitokey.CommitteePublicKey
	beaconSubstitutePool   []incognitokey.CommitteePublicKey
	beaconCommonPool       []incognitokey.CommitteePublicKey
	assignBeaconCheckPoint int

	shardCommitteePool    map[byte][]incognitokey.CommitteePublicKey
	shardSubstitutePool   map[byte][]incognitokey.CommitteePublicKey
	shardCommonPool       []incognitokey.CommitteePublicKey
	assignShardCheckPoint int

	autoStake      map[string]bool                   // committee public key => reward receiver payment address
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address

	mu *sync.RWMutex
}

type BeaconCommitteeEngineV2 struct {
	beaconHeight                      uint64
	beaconHash                        common.Hash
	finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
	uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
	insProcessor                      *instructionsprocessor.BInsProcessor
}

func (b BeaconCommitteeStateV2) clone(newB *BeaconCommitteeStateV2) {
	newB.reset()
	newB.beaconCommitteePool = b.beaconCommitteePool
	newB.beaconSubstitutePool = b.beaconSubstitutePool
	newB.beaconCommonPool = b.beaconCommonPool
	newB.shardCommonPool = b.shardCommonPool
	for k, v := range b.shardCommitteePool {
		newB.shardCommitteePool[k] = v
	}
	for k, v := range b.shardSubstitutePool {
		newB.shardSubstitutePool[k] = v
	}
	for k, v := range b.autoStake {
		newB.autoStake[k] = v
	}
	for k, v := range b.rewardReceiver {
		newB.rewardReceiver[k] = v
	}
	for k, v := range b.stakingTx {
		newB.stakingTx[k] = v
	}
}

func (b *BeaconCommitteeStateV2) reset() {
	b.beaconCommitteePool = []incognitokey.CommitteePublicKey{}
	b.beaconSubstitutePool = []incognitokey.CommitteePublicKey{}
	b.beaconCommonPool = []incognitokey.CommitteePublicKey{}
	b.shardCommonPool = []incognitokey.CommitteePublicKey{}
	b.shardCommitteePool = make(map[byte][]incognitokey.CommitteePublicKey)
	b.shardSubstitutePool = make(map[byte][]incognitokey.CommitteePublicKey)
	b.autoStake = make(map[string]bool)
	b.rewardReceiver = make(map[string]privacy.PaymentAddress)
	b.stakingTx = make(map[string]common.Hash)
}

//GetBeaconHeight :
func (engine BeaconCommitteeEngineV2) GetBeaconHeight() uint64 {
	return engine.beaconHeight
}

//GetBeaconHash :
func (engine BeaconCommitteeEngineV2) GetBeaconHash() common.Hash {
	return engine.beaconHash
}

//GetBeaconCommittee :
func (engine BeaconCommitteeEngineV2) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.beaconCommitteePool
}

//GetBeaconSubstitute :
func (engine BeaconCommitteeEngineV2) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.beaconSubstitutePool
}

//GetCandidateShardWaitingForCurrentRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommonPool[:engine.finalBeaconCommitteeStateV2.assignShardCheckPoint]
}

//GetCandidateBeaconWaitingForCurrentRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.beaconCommonPool[:engine.finalBeaconCommitteeStateV2.assignBeaconCheckPoint]
}

//GetCandidateShardWaitingForNextRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommonPool[engine.finalBeaconCommitteeStateV2.assignShardCheckPoint:]
}

//GetCandidateBeaconWaitingForNextRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.beaconCommonPool[engine.finalBeaconCommitteeStateV2.assignBeaconCheckPoint:]
}

//GetOneShardCommittee :
func (engine BeaconCommitteeEngineV2) GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommitteePool[shardID]
}

//GetShardCommittee :
func (engine BeaconCommitteeEngineV2) GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalBeaconCommitteeStateV2.shardCommitteePool {
		shardCommittee[k] = v
	}
	return shardCommittee
}

//GetOneShardSubstitute :
func (engine BeaconCommitteeEngineV2) GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardSubstitutePool[shardID]
}

//GetShardSubstitute :
func (engine BeaconCommitteeEngineV2) GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	shardSubstitute := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalBeaconCommitteeStateV2.shardSubstitutePool {
		shardSubstitute[k] = v
	}
	return shardSubstitute
}

//GetAutoStaking :
func (engine BeaconCommitteeEngineV2) GetAutoStaking() map[string]bool {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	autoStake := make(map[string]bool)
	for k, v := range engine.finalBeaconCommitteeStateV2.autoStake {
		autoStake[k] = v
	}
	return autoStake
}

func (engine BeaconCommitteeEngineV2) GetRewardReceiver() map[string]privacy.PaymentAddress {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	rewardReceiver := make(map[string]privacy.PaymentAddress)
	for k, v := range engine.finalBeaconCommitteeStateV2.rewardReceiver {
		rewardReceiver[k] = v
	}
	return rewardReceiver
}

func (engine BeaconCommitteeEngineV2) GetStakingTx() map[string]common.Hash {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	stakingTx := make(map[string]common.Hash)
	for k, v := range engine.finalBeaconCommitteeStateV2.stakingTx {
		stakingTx[k] = v
	}
	return stakingTx
}

func (engine *BeaconCommitteeEngineV2) GetAllCandidateSubstituteCommittee() []string {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	return engine.finalBeaconCommitteeStateV2.getAllCandidateSubstituteCommittee()
}

func (b *BeaconCommitteeStateV2) getAllCandidateSubstituteCommittee() []string {
	res := []string{}
	for _, committee := range b.shardCommitteePool {
		shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(committee)
		if err != nil {
			panic(err)
		}
		res = append(res, shardCommitteeStr...)
	}
	for _, substitute := range b.shardSubstitutePool {
		beaconSubstituteStr, err := incognitokey.CommitteeKeyListToString(substitute)
		if err != nil {
			panic(err)
		}
		res = append(res, beaconSubstituteStr...)
	}
	beaconCommittee := b.beaconCommitteePool
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconCommittee)
	if err != nil {
		panic(err)
	}
	res = append(res, beaconCommitteeStr...)
	beaconSubstitute := b.beaconSubstitutePool
	beaconSubstituteStr, err := incognitokey.CommitteeKeyListToString(beaconSubstitute)
	if err != nil {
		panic(err)
	}
	res = append(res, beaconSubstituteStr...)
	beaconCandidates := b.beaconCommonPool
	beaconCandidatesStr, err := incognitokey.CommitteeKeyListToString(beaconCandidates)
	if err != nil {
		panic(err)
	}
	res = append(res, beaconCandidatesStr...)
	shardCandidates := b.shardCommonPool
	shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(shardCandidates)
	if err != nil {
		panic(err)
	}
	res = append(res, shardCandidatesStr...)
	return res
}

func (engine *BeaconCommitteeEngineV2) Commit(hashes *BeaconCommitteeStateHash) error {
	if reflect.DeepEqual(engine.uncommittedBeaconCommitteeStateV2, NewBeaconCommitteeStateV1()) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("%+v", engine.uncommittedBeaconCommitteeStateV2))
	}
	engine.uncommittedBeaconCommitteeStateV2.mu.Lock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.Unlock()
	engine.finalBeaconCommitteeStateV2.mu.Lock()
	defer engine.finalBeaconCommitteeStateV2.mu.Unlock()
	comparedHashes, err := engine.generateUncommittedCommitteeHashes()
	if err != nil {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, err)
	}
	if !comparedHashes.BeaconCommitteeAndValidatorHash.IsEqual(&hashes.BeaconCommitteeAndValidatorHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted BeaconCommitteeAndValidatorHash want value %+v but have %+v", comparedHashes.BeaconCommitteeAndValidatorHash, hashes.BeaconCommitteeAndValidatorHash))
	}
	if !comparedHashes.BeaconCandidateHash.IsEqual(&hashes.BeaconCandidateHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted BeaconCandidateHash want value %+v but have %+v", comparedHashes.BeaconCandidateHash, hashes.BeaconCandidateHash))
	}
	if !comparedHashes.ShardCandidateHash.IsEqual(&hashes.ShardCandidateHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted ShardCandidateHash want value %+v but have %+v", comparedHashes.ShardCandidateHash, hashes.ShardCandidateHash))
	}
	if !comparedHashes.ShardCommitteeAndValidatorHash.IsEqual(&hashes.ShardCommitteeAndValidatorHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted ShardCommitteeAndValidatorHash want value %+v but have %+v", comparedHashes.ShardCommitteeAndValidatorHash, hashes.ShardCommitteeAndValidatorHash))
	}
	if !comparedHashes.AutoStakeHash.IsEqual(&hashes.AutoStakeHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted AutoStakingHash want value %+v but have %+v", comparedHashes.AutoStakeHash, hashes.AutoStakeHash))
	}
	engine.uncommittedBeaconCommitteeStateV2.clone(engine.finalBeaconCommitteeStateV2)
	engine.uncommittedBeaconCommitteeStateV2.reset()
	return nil
}

func (engine *BeaconCommitteeEngineV2) AbortUncommittedBeaconState() {
	engine.uncommittedBeaconCommitteeStateV2.mu.Lock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.Unlock()
	engine.uncommittedBeaconCommitteeStateV2.reset()
}

func (engine *BeaconCommitteeEngineV2) InitCommitteeState(env *BeaconCommitteeStateEnvironment) {
	engine.finalBeaconCommitteeStateV2.mu.Lock()
	defer engine.finalBeaconCommitteeStateV2.mu.Unlock()
	b := engine.finalBeaconCommitteeStateV2
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		if inst[0] == instruction.STAKE_ACTION {
			stakeInstruction := instruction.ImportInitStakeInstructionFromString(inst)
			tempNewBeaconCandidates, tempNewShardCandidates, _ := b.processStakeInstruction(stakeInstruction, env)
			newBeaconCandidates = append(newBeaconCandidates, tempNewBeaconCandidates...)
			newShardCandidates = append(newShardCandidates, tempNewShardCandidates...)
		}
	}
	b.beaconCommitteePool = append(b.beaconCommitteePool, newBeaconCandidates...)
	for shardID := 0; shardID < env.ActiveShards; shardID++ {
		b.shardCommitteePool[byte(shardID)] = append(b.shardCommitteePool[byte(shardID)], newShardCandidates[shardID*env.MinShardCommitteeSize:(shardID+1)*env.MinShardCommitteeSize]...)
	}
}

// New flow
// Store information from instructions into temp stateDB in env
// When all thing done and no problems, in commit function, we read data in statedb and update
// BeaconCommitteeStateV2
func (engine *BeaconCommitteeEngineV2) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (*BeaconCommitteeStateHash, *CommitteeChange, error) {
	engine.uncommittedBeaconCommitteeStateV2.mu.Lock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.Unlock()
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	engine.finalBeaconCommitteeStateV2.clone(engine.uncommittedBeaconCommitteeStateV2)
	env.allCandidateSubstituteCommittee = engine.finalBeaconCommitteeStateV2.getAllCandidateSubstituteCommittee()
	engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	newB := engine.uncommittedBeaconCommitteeStateV2
	committeeChange := NewCommitteeChange()
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	if env.IsBeaconRandomTime {
		engine.uncommittedBeaconCommitteeStateV2.assignShardCheckPoint = len(engine.uncommittedBeaconCommitteeStateV2.shardCommonPool)
		engine.uncommittedBeaconCommitteeStateV2.assignBeaconCheckPoint = len(engine.uncommittedBeaconCommitteeStateV2.beaconCommonPool)
	}
	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		tempNewBeaconCandidates, tempNewShardCandidates := []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		switch inst[0] {
		case instruction.STAKE_ACTION:
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
			tempNewBeaconCandidates, tempNewShardCandidates, err = newB.processStakeInstruction(stakeInstruction, env)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
		case instruction.SWAP_ACTION:
			swapInstruction, err := instruction.ValidateAndImportSwapInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP swap instruction %+v, error %+v", inst, err)
				continue
			}
			tempNewBeaconCandidates, tempNewShardCandidates, err = newB.processSwapInstruction(swapInstruction, env, committeeChange)
			if err != nil {
				return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
			}
			newB.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange)
		}
		if len(tempNewBeaconCandidates) > 0 {
			newBeaconCandidates = append(newBeaconCandidates, tempNewBeaconCandidates...)
		}
		if len(tempNewShardCandidates) > 0 {
			newShardCandidates = append(newShardCandidates, tempNewShardCandidates...)
		}
	}

	if env.IsFoundRandomNumber {
		//Already update committee when process assign instruction
	}
	err := newB.processAutoStakingChange(committeeChange, env)
	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	hashes, err := engine.generateUncommittedCommitteeHashes()
	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	return hashes, committeeChange, nil
	// panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GenerateAssignInstruction(candidates []string, numberOfPendingValidator map[byte]int, rand int64, assignOffset int, activeShards int) ([]string, map[byte][]string) {
	panic("implement me")
}

func (engine *BeaconCommitteeEngineV2) generateUncommittedCommitteeHashes() (*BeaconCommitteeStateHash, error) {
	if reflect.DeepEqual(engine.uncommittedBeaconCommitteeStateV2, NewBeaconCommitteeStateV1()) {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	newB := engine.uncommittedBeaconCommitteeStateV2
	// beacon committee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(newB.beaconCommitteePool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	validatorArr := append([]string{}, beaconCommitteeStr...)

	beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(newB.beaconSubstitutePool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	validatorArr = append(validatorArr, beaconPendingValidatorStr...)
	tempBeaconCommitteeAndValidatorHash, err := common.GenerateHashFromStringArray(validatorArr)
	// beacon candidate: current candidate + next candidate
	// BeaconCandidate root: beacon current candidate + beacon next candidate
	beaconCandidateArrStr, err := incognitokey.CommitteeKeyListToString(newB.beaconCommonPool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	tempBeaconCandidateHash, err := common.GenerateHashFromStringArray(beaconCandidateArrStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	// Shard candidate root: shard current candidate + shard next candidate
	shardCandidateArrStr, err := incognitokey.CommitteeKeyListToString(newB.shardCommonPool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	tempShardCandidateHash, err := common.GenerateHashFromStringArray(shardCandidateArrStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	// Shard Validator root
	shardPendingValidator := make(map[byte][]string)
	for shardID, keys := range newB.shardSubstitutePool {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		shardPendingValidator[shardID] = keysStr
	}
	shardCommittee := make(map[byte][]string)
	for shardID, keys := range newB.shardCommitteePool {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		shardCommittee[shardID] = keysStr
	}
	tempShardCommitteeAndValidatorHash, err := common.GenerateHashFromMapByteString(shardPendingValidator, shardCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	tempAutoStakingHash, err := common.GenerateHashFromMapStringBool(newB.autoStake)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	hashes := &BeaconCommitteeStateHash{
		BeaconCommitteeAndValidatorHash: tempBeaconCommitteeAndValidatorHash,
		BeaconCandidateHash:             tempBeaconCandidateHash,
		ShardCandidateHash:              tempShardCandidateHash,
		ShardCommitteeAndValidatorHash:  tempShardCommitteeAndValidatorHash,
		AutoStakeHash:                   tempAutoStakingHash,
	}
	return hashes, nil
}

func (b *BeaconCommitteeStateV2) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	env *BeaconCommitteeStateEnvironment,
) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	for index, candidate := range stakeInstruction.PublicKeyStructs {
		b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
		b.autoStake[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
		b.stakingTx[stakeInstruction.PublicKeys[index]] = stakeInstruction.TxStakeHashes[index]
	}
	if stakeInstruction.Chain == instruction.BEACON_INST {
		newBeaconCandidates = append(newBeaconCandidates, stakeInstruction.PublicKeyStructs...)
	} else {
		newShardCandidates = append(newShardCandidates, stakeInstruction.PublicKeyStructs...)
	}
	err := statedb.StoreStakerInfo(
		env.ConsensusStateDB,
		stakeInstruction.PublicKeyStructs,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
	)
	if err != nil {
		return newBeaconCandidates, newShardCandidates, err
	}
	return newBeaconCandidates, newShardCandidates, nil
}

func (b *BeaconCommitteeStateV2) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) {
	for _, committeePublicKey := range stopAutoStakeInstruction.PublicKeys {
		if common.IndexOfStr(committeePublicKey, env.allCandidateSubstituteCommittee) == -1 {
			// if not found then delete auto staking data for this public key if present
			if _, ok := b.autoStake[committeePublicKey]; ok {
				delete(b.autoStake, committeePublicKey)
			}
		} else {
			// if found in committee list then turn off auto staking
			if _, ok := b.autoStake[committeePublicKey]; ok {
				b.autoStake[committeePublicKey] = false
				committeeChange.StopAutoStake = append(committeeChange.StopAutoStake, committeePublicKey)
			}
		}
	}
}

func (b *BeaconCommitteeStateV2) processSwapInstruction(
	swapInstruction *instruction.SwapInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	//TODO
	return nil, nil, nil
}

func (b *BeaconCommitteeStateV2) processAutoStakingChange(committeeChange *CommitteeChange, env *BeaconCommitteeStateEnvironment) error {
	stopAutoStakingIncognitoKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeeChange.StopAutoStake)
	if err != nil {
		return err
	}
	err = statedb.StoreStakerInfo(
		env.ConsensusStateDB,
		stopAutoStakingIncognitoKey,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
	)
	return nil
}
