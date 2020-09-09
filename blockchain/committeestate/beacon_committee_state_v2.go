package committeestate

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/blockchain/instructionsprocessor"

	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV2 struct {
	beaconCommittee []incognitokey.CommitteePublicKey
	shardCommittee  map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey
	shardCommonPool []incognitokey.CommitteePublicKey
	// TODO: @hung record already requested shard substitute/committee
	numberOfAssignedCandidates int

	autoStake      map[string]bool                   // committee public key => reward receiver payment address
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address
	numberOfRound  map[string]int                    // committee public key => number of round in epoch

	mu *sync.RWMutex
}

type BeaconCommitteeEngineV2 struct {
	beaconHeight                      uint64
	beaconHash                        common.Hash
	finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
	uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
	insProcessor                      *instructionsprocessor.BInsProcessor
}

func NewBeaconCommitteeEngineV2(beaconHeight uint64, beaconHash common.Hash, finalBeaconCommitteeStateV2 *BeaconCommitteeStateV2) *BeaconCommitteeEngineV2 {
	return &BeaconCommitteeEngineV2{beaconHeight: beaconHeight, beaconHash: beaconHash, finalBeaconCommitteeStateV2: finalBeaconCommitteeStateV2, uncommittedBeaconCommitteeStateV2: NewBeaconCommitteeStateV2()}
}

func NewBeaconCommitteeStateV2() *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
		autoStake:       make(map[string]bool),
		rewardReceiver:  make(map[string]privacy.PaymentAddress),
		stakingTx:       make(map[string]common.Hash),
		numberOfRound:   make(map[string]int),
		mu:              new(sync.RWMutex),
	}
}

func NewBeaconCommitteeStateV2WithValue(
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	shardCommonPool []incognitokey.CommitteePublicKey,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	numberOfRound map[string]int,
) *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		beaconCommittee:            beaconCommittee,
		shardCommittee:             shardCommittee,
		shardSubstitute:            shardSubstitute,
		shardCommonPool:            shardCommonPool,
		numberOfAssignedCandidates: numberOfAssignedCandidates,
		autoStake:                  autoStake,
		rewardReceiver:             rewardReceiver,
		stakingTx:                  stakingTx,
		numberOfRound:              numberOfRound,
		mu:                         new(sync.RWMutex),
	}
}

func (b BeaconCommitteeStateV2) clone(newB *BeaconCommitteeStateV2) {
	newB.reset()
	newB.beaconCommittee = b.beaconCommittee
	newB.shardCommonPool = b.shardCommonPool
	newB.numberOfAssignedCandidates = b.numberOfAssignedCandidates
	for k, v := range b.shardCommittee {
		newB.shardCommittee[k] = v
	}
	for k, v := range b.shardSubstitute {
		newB.shardSubstitute[k] = v
	}
	for k, v := range b.autoStake {
		newB.autoStake[k] = v
	}
	for k, v := range b.numberOfRound {
		newB.numberOfRound[k] = v
	}
	for k, v := range b.rewardReceiver {
		newB.rewardReceiver[k] = v
	}
	for k, v := range b.stakingTx {
		newB.stakingTx[k] = v
	}
}

func (b *BeaconCommitteeStateV2) reset() {
	b.beaconCommittee = []incognitokey.CommitteePublicKey{}
	b.shardCommonPool = []incognitokey.CommitteePublicKey{}
	b.shardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	b.shardSubstitute = make(map[byte][]incognitokey.CommitteePublicKey)
	b.autoStake = make(map[string]bool)
	b.numberOfRound = make(map[string]int)
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
	return engine.finalBeaconCommitteeStateV2.beaconCommittee
}

//GetBeaconSubstitute :
func (engine BeaconCommitteeEngineV2) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetCandidateShardWaitingForCurrentRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommonPool[:engine.finalBeaconCommitteeStateV2.numberOfAssignedCandidates]
}

//GetCandidateBeaconWaitingForCurrentRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetCandidateShardWaitingForNextRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommonPool[engine.finalBeaconCommitteeStateV2.numberOfAssignedCandidates:]
}

//GetCandidateBeaconWaitingForNextRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetOneShardCommittee :
func (engine BeaconCommitteeEngineV2) GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommittee[shardID]
}

//GetShardCommittee :
func (engine BeaconCommitteeEngineV2) GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalBeaconCommitteeStateV2.shardCommittee {
		shardCommittee[k] = v
	}
	return shardCommittee
}

//GetOneShardSubstitute :
func (engine BeaconCommitteeEngineV2) GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardSubstitute[shardID]
}

//GetShardSubstitute :
func (engine BeaconCommitteeEngineV2) GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	shardSubstitute := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalBeaconCommitteeStateV2.shardSubstitute {
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
			for index, candidate := range stakeInstruction.PublicKeyStructs {
				b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
				b.autoStake[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
				b.numberOfRound[stakeInstruction.PublicKeys[index]] = 1
				b.stakingTx[stakeInstruction.PublicKeys[index]] = stakeInstruction.TxStakeHashes[index]
			}
			if stakeInstruction.Chain == instruction.BEACON_INST {
				newBeaconCandidates = append(newBeaconCandidates, stakeInstruction.PublicKeyStructs...)
			} else {
				newShardCandidates = append(newShardCandidates, stakeInstruction.PublicKeyStructs...)
			}
			_ = statedb.StoreStakerInfoV2(
				env.ConsensusStateDB,
				stakeInstruction.PublicKeyStructs,
				b.rewardReceiver,
				b.autoStake,
				b.stakingTx,
				b.numberOfRound,
			)
		}
	}
	b.beaconCommittee = append(b.beaconCommittee, newBeaconCandidates...)
	for shardID := 0; shardID < env.ActiveShards; shardID++ {
		b.shardCommittee[byte(shardID)] = append(b.shardCommittee[byte(shardID)], newShardCandidates[shardID*env.MinShardCommitteeSize:(shardID+1)*env.MinShardCommitteeSize]...)
	}
}

// New flow
// Store information from instructions into temp stateDB in env
// When all thing done and no problems, in commit function, we read data in statedb and update
// BeaconCommitteeStateV2
func (engine *BeaconCommitteeEngineV2) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var err error
	incurredInstructions := [][]string{}

	engine.uncommittedBeaconCommitteeStateV2.mu.Lock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.Unlock()

	engine.finalBeaconCommitteeStateV2.mu.RLock()

	env.unassignedCommonPool, err = engine.finalBeaconCommitteeStateV2.unassignedCommonPool()
	if err != nil {
		return nil, nil, nil, err
	}
	env.allSubstituteCommittees, err = engine.finalBeaconCommitteeStateV2.getAllSubstituteCommittees()
	if err != nil {
		return nil, nil, nil, err
	}
	env.allCandidateSubstituteCommittee = append(env.unassignedCommonPool, env.allSubstituteCommittees...)

	engine.finalBeaconCommitteeStateV2.clone(engine.uncommittedBeaconCommitteeStateV2)
	env.allCandidateSubstituteCommittee = engine.finalBeaconCommitteeStateV2.getAllCandidateSubstituteCommittee()
	engine.finalBeaconCommitteeStateV2.mu.RUnlock()

	newB := engine.uncommittedBeaconCommitteeStateV2
	committeeChange := NewCommitteeChange()
	incuredInstructions := [][]string{}
	// snapshot shard common pool in beacon random time
	if env.IsBeaconRandomTime {
		newB.numberOfAssignedCandidates = SnapshotShardCommonPoolV2(
			newB.shardCommonPool,
			newB.shardCommittee,
			newB.shardSubstitute,
			env.MaxCommitteeSize,
		)
		Logger.log.Infof("Block %+v, Number of Snapshot to Assign Candidate %+v", env.BeaconHeight, newB.numberOfAssignedCandidates)
	}

	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		switch inst[0] {
		case instruction.STAKE_ACTION:
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
			committeeChange, err = newB.processStakeInstruction(stakeInstruction, committeeChange, env)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
		case instruction.RANDOM_ACTION:
			randomInstruction, err := instruction.ValidateAndImportRandomInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
			}
			committeeChange = newB.processAssignWithRandomInstruction(
				randomInstruction.BtcNonce, env.ActiveShards, committeeChange)
			Logger.log.Infof("Block %+v, Committee Change %+v", env.BeaconHeight, committeeChange.ShardSubstituteAdded)
		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
			}
			committeeChange = newB.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange)
		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP unstake instruction %+v, error %+v", inst, err)
				continue
			}
			tempIncurredIns := [][]string{}
			committeeChange, tempIncurredIns, err =
				newB.processUnstakeInstruction(unstakeInstruction, env, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			if tempIncurredIns != nil {
				incurredInstructions = append(incurredInstructions, tempIncurredIns...)
			}
		case instruction.SWAP_SHARD_ACTION:
			// Process swap shard action here
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP Swap Shard Committees instruction %+v, error %+v", inst, err)
				continue
			}
			committeeChange, err = newB.processSwapShardInstruction(swapShardInstruction, env, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		}
	}
	err = newB.processAutoStakingChange(committeeChange, env)
	if err != nil {
		return nil, nil, incuredInstructions, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	hashes, err := engine.generateUncommittedCommitteeHashes()
	if err != nil {
		return nil, nil, incuredInstructions, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	return hashes, committeeChange, incuredInstructions, nil
}

// GenerateAllSwapShardInstructions generate swap shard instructions for all shard
// it also assigned swapped out committee back to substitute list if auto stake is true
// generate all swap shard instructions by only swap by the end of epoch (normally execution)
func (engine *BeaconCommitteeEngineV2) GenerateAllSwapShardInstructions(
	env *BeaconCommitteeStateEnvironment) (
	[]*instruction.SwapShardInstruction, error) {
	swapShardInstructions := []*instruction.SwapShardInstruction{}
	for i := 0; i < env.ActiveShards; i++ {
		shardID := byte(i)
		committees := engine.finalBeaconCommitteeStateV2.shardCommittee[shardID]
		committees = committees[env.NumberOfFixedBlockValidator:]
		substitutes := engine.finalBeaconCommitteeStateV2.shardSubstitute[shardID]
		if len(substitutes) == 0 {
			continue
		}
		tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
		tempSubstitutes, _ := incognitokey.CommitteeKeyListToString(substitutes)
		//TODO: @hung rewrite swapV2
		// 1. don't create swap: len(substitutes) == 0 and len(slashedCommittee) == 0
		// 2. create swap: len(substitutes) == 0 and len(slashedCommittee) > 0
		// 3. create swap: len(substitutes) > 0 and len(slashedCommittee) == 0

		latestHeight := getLatestHeightByShardsState(env.LatestShardsState[shardID])
		if latestHeight == 0 {
			continue
		}

		swapShardInstruction, _, err := createSwapShardInstructionV2(
			shardID,
			tempSubstitutes,
			tempCommittees,
			env.MaxCommitteeSize,
			engine.finalBeaconCommitteeStateV2.numberOfRound,
			instruction.SWAP_BY_END_EPOCH,
			latestHeight+1,
		)
		if err != nil {
			return swapShardInstructions, err
		}

		swapShardInstructions = append(swapShardInstructions, swapShardInstruction)
	}
	return swapShardInstructions, nil
}

func (engine *BeaconCommitteeEngineV2) GenerateAssignInstruction(rand int64, assignOffset int, activeShards int) ([]*instruction.AssignInstruction, []string, map[byte][]string) {
	return []*instruction.AssignInstruction{}, []string{}, make(map[byte][]string)
}

func (engine *BeaconCommitteeEngineV2) BuildIncurredInstructions(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	panic("implement me")
}

func SnapshotShardCommonPoolV2(
	shardCommonPool []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	maxAssignPerShard int,
) (numberOfAssignedCandidates int) {
	for k, v := range shardCommittee {
		shardCommitteeSize := len(v)
		shardCommitteeSize += len(shardSubstitute[k])
		assignPerShard := shardCommitteeSize / MAX_SWAP_OR_ASSIGN_PERCENT
		Logger.log.Infof("SnapshotShardCommonPoolV2 | Shard %+v, shardCommitteeSize %+v", k, shardCommitteeSize)
		Logger.log.Infof("SnapshotShardCommonPoolV2 | Shard %+v, assignPerShard %+v", k, assignPerShard)
		if assignPerShard > maxAssignPerShard {
			assignPerShard = maxAssignPerShard
		}
		Logger.log.Infof("SnapshotShardCommonPoolV2 | Shard %+v, maxAssignPerShard %+v", k, maxAssignPerShard)
		numberOfAssignedCandidates += assignPerShard
		Logger.log.Infof("SnapshotShardCommonPoolV2 | Shard %+v, numberOfAssignedCandidates %+v", k, numberOfAssignedCandidates)
	}
	Logger.log.Infof("SnapshotShardCommonPoolV2 | Shard Common Pool Size %+v", len(shardCommonPool))
	if numberOfAssignedCandidates > len(shardCommonPool) {
		numberOfAssignedCandidates = len(shardCommonPool)
	}
	return numberOfAssignedCandidates
}

func (b *BeaconCommitteeStateV2) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	committeeChange *CommitteeChange,
	env *BeaconCommitteeStateEnvironment,
) (*CommitteeChange, error) {
	for index, candidate := range stakeInstruction.PublicKeyStructs {
		b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
		b.autoStake[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
		b.numberOfRound[stakeInstruction.PublicKeys[index]] = 0
		b.stakingTx[stakeInstruction.PublicKeys[index]] = stakeInstruction.TxStakeHashes[index]
	}
	committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, stakeInstruction.PublicKeyStructs...)
	b.shardCommonPool = append(b.shardCommonPool, stakeInstruction.PublicKeyStructs...)
	err := statedb.StoreStakerInfoV2(
		env.ConsensusStateDB,
		stakeInstruction.PublicKeyStructs,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
		b.numberOfRound,
	)
	if err != nil {
		return committeeChange, err
	}
	return committeeChange, nil
}

func (b *BeaconCommitteeStateV2) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) *CommitteeChange {
	for _, committeePublicKey := range stopAutoStakeInstruction.CommitteePublicKeys {
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
	return committeeChange
}

func (b *BeaconCommitteeStateV2) processAssignWithRandomInstruction(
	rand int64, activeShards int, committeeChange *CommitteeChange) *CommitteeChange {
	candidates, _ := incognitokey.CommitteeKeyListToString(b.shardCommonPool[:b.numberOfAssignedCandidates])
	committeeChange = b.assign(candidates, rand, activeShards, committeeChange)
	committeeChange.NextEpochShardCandidateRemoved = append(committeeChange.NextEpochShardCandidateRemoved, b.shardCommonPool[:b.numberOfAssignedCandidates]...)
	b.shardCommonPool = b.shardCommonPool[b.numberOfAssignedCandidates:]
	b.numberOfAssignedCandidates = 0
	return committeeChange
}

func (b *BeaconCommitteeStateV2) assign(
	candidates []string, rand int64, activeShards int, committeeChange *CommitteeChange) *CommitteeChange {
	numberOfValidator := make([]int, activeShards)
	for i := 0; i < activeShards; i++ {
		numberOfValidator[byte(i)] += len(b.shardSubstitute[byte(i)])
		numberOfValidator[byte(i)] += len(b.shardCommittee[byte(i)])
	}
	assignedCandidates := assignShardCandidateV2(candidates, numberOfValidator, rand)
	for shardID, tempCandidates := range assignedCandidates {
		candidates, _ := incognitokey.CommitteeBase58KeyListToStruct(tempCandidates)
		b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], candidates...)
		committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], candidates...)
		for _, tempCandidate := range tempCandidates {
			b.numberOfRound[tempCandidate]++
		}
	}
	return committeeChange
}

func (b *BeaconCommitteeStateV2) processSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange) (*CommitteeChange, error) {

	newCommitteeChange := committeeChange

	//TODO: @tin add random number to env variable
	// b.processAfterSwap(env, swapShardInstruction.OutPublicKeys)

	return newCommitteeChange, nil
}

//TODO: @tin Use this function to build new swap rule
// processAfterSwap process swapped out committee public key
// if number of round is less than MAX_NUMBER_OF_ROUND go back to THAT shard pool, and increase number of round
// if number of round is equal to or greater than MAX_NUMBER_OF_ROUND
// - auto stake is false then remove completely out of any committee, candidate, substitute list
// - auto stake is true then using assignment rule v2 to assign this committee public key
func (b *BeaconCommitteeStateV2) processAfterSwap(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	outPublicKeyStructs []incognitokey.CommitteePublicKey,
	shardID byte,
	committeeChange *CommitteeChange,
) *CommitteeChange {
	backToSubstitutesIndex := []int{}
	swappedOutSubstitutesIndex := []int{}
	candidates := []string{}

	for index, outPublicKey := range outPublicKeys {
		if b.numberOfRound[outPublicKey] >= MAX_NUMBER_OF_ROUND {
			swappedOutSubstitutesIndex = append(swappedOutSubstitutesIndex, index)
		} else {
			backToSubstitutesIndex = append(backToSubstitutesIndex, index)
		}
	}

	for _, index := range backToSubstitutesIndex {
		b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], outPublicKeyStructs[index])
		committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], outPublicKeyStructs[index])
		b.numberOfRound[outPublicKeys[index]]++
	}

	for _, index := range swappedOutSubstitutesIndex {
		// @NOTICE: these lines of code is for debug purpose
		stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, outPublicKeys[index])
		if err != nil {
			panic(err)
		}
		if !has {
			panic(errors.Errorf("Can not found info of this public key %v", outPublicKeys[index]))
		}
		if stakerInfo.AutoStaking() {
			b.numberOfRound[outPublicKeys[index]] = 0
			candidates = append(candidates, outPublicKeys[index])
		} else {
			delete(b.rewardReceiver, outPublicKeyStructs[index].GetIncKeyBase58())
			delete(b.autoStake, outPublicKeys[index])
			delete(b.numberOfRound, outPublicKeys[index])
			delete(b.stakingTx, outPublicKeys[index])
		}
	}

	committeeChange = b.assign(candidates, env.RandomNumber, env.ActiveShards, committeeChange)

	return committeeChange
}

func (engine *BeaconCommitteeEngineV2) generateUncommittedCommitteeHashes() (*BeaconCommitteeStateHash, error) {
	if reflect.DeepEqual(engine.uncommittedBeaconCommitteeStateV2, NewBeaconCommitteeStateV1()) {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	newB := engine.uncommittedBeaconCommitteeStateV2
	// beacon committee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(newB.beaconCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	validatorArr := append([]string{}, beaconCommitteeStr...)

	tempBeaconCommitteeAndValidatorHash, err := common.GenerateHashFromStringArray(validatorArr)
	// Shard candidate root: shard current candidate + shard next candidate
	shardNextEpochCandidateStr, err := incognitokey.CommitteeKeyListToString(newB.shardCommonPool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	tempShardCandidateHash, err := common.GenerateHashFromStringArray(shardNextEpochCandidateStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	// Shard Validator root
	shardPendingValidator := make(map[byte][]string)
	for shardID, keys := range newB.shardSubstitute {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		shardPendingValidator[shardID] = keysStr
	}
	shardCommittee := make(map[byte][]string)
	for shardID, keys := range newB.shardCommittee {
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
		ShardCandidateHash:              tempShardCandidateHash,
		ShardCommitteeAndValidatorHash:  tempShardCommitteeAndValidatorHash,
		AutoStakeHash:                   tempAutoStakingHash,
	}
	return hashes, nil
}

func (b *BeaconCommitteeStateV2) getAllCandidateSubstituteCommittee() []string {
	res := []string{}
	for _, committee := range b.shardCommittee {
		shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(committee)
		if err != nil {
			panic(err)
		}
		res = append(res, shardCommitteeStr...)
	}
	for _, substitute := range b.shardSubstitute {
		beaconSubstituteStr, err := incognitokey.CommitteeKeyListToString(substitute)
		if err != nil {
			panic(err)
		}
		res = append(res, beaconSubstituteStr...)
	}
	beaconCommittee := b.beaconCommittee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconCommittee)
	if err != nil {
		panic(err)
	}
	res = append(res, beaconCommitteeStr...)
	shardCandidates := b.shardCommonPool
	shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(shardCandidates)
	if err != nil {
		panic(err)
	}
	res = append(res, shardCandidatesStr...)
	return res
}

func (b *BeaconCommitteeStateV2) processAutoStakingChange(committeeChange *CommitteeChange, env *BeaconCommitteeStateEnvironment) error {
	stopAutoStakingIncognitoKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeeChange.StopAutoStake)
	if err != nil {
		return err
	}
	err = statedb.StoreStakerInfoV2(
		env.ConsensusStateDB,
		stopAutoStakingIncognitoKey,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
		b.numberOfRound,
	)
	return nil
}
