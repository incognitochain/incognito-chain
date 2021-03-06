package committeestate

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type BeaconCommitteeStateV2 struct {
	beaconCommittee            []incognitokey.CommitteePublicKey
	shardCommittee             map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
	shardCommonPool            []incognitokey.CommitteePublicKey
	numberOfAssignedCandidates int

	autoStake      map[string]bool                   // committee public key => true or false
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address

	hashes *BeaconCommitteeStateHash

	mu *sync.RWMutex
}

func (b *BeaconCommitteeStateV2) setHashes(hashes *BeaconCommitteeStateHash) {
	b.hashes = hashes
}

type BeaconCommitteeEngineV2 struct {
	beaconHeight                      uint64
	beaconHash                        common.Hash
	finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
	uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
}

func NewBeaconCommitteeEngineV2(
	beaconHeight uint64,
	beaconHash common.Hash,
	finalBeaconCommitteeStateV2 *BeaconCommitteeStateV2) *BeaconCommitteeEngineV2 {
	return &BeaconCommitteeEngineV2{
		beaconHeight:                      beaconHeight,
		beaconHash:                        beaconHash,
		finalBeaconCommitteeStateV2:       finalBeaconCommitteeStateV2,
		uncommittedBeaconCommitteeStateV2: NewBeaconCommitteeStateV2(),
	}
}

func NewBeaconCommitteeStateV2() *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
		autoStake:       make(map[string]bool),
		rewardReceiver:  make(map[string]privacy.PaymentAddress),
		stakingTx:       make(map[string]common.Hash),
		hashes:          NewBeaconCommitteeStateHash(),
		mu:              new(sync.RWMutex),
	}
}
func NewBeaconCommitteeStateV2WithMu(mu *sync.RWMutex) *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
		autoStake:       make(map[string]bool),
		rewardReceiver:  make(map[string]privacy.PaymentAddress),
		stakingTx:       make(map[string]common.Hash),
		mu:              mu,
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
		hashes:                     NewBeaconCommitteeStateHash(),
		mu:                         new(sync.RWMutex),
	}
}

//shallowCopy maintain dst mutex value
func (b *BeaconCommitteeStateV2) shallowCopy(newB *BeaconCommitteeStateV2) {
	newB.beaconCommittee = b.beaconCommittee
	newB.shardCommittee = b.shardCommittee
	newB.shardSubstitute = b.shardSubstitute
	newB.shardCommonPool = b.shardCommonPool
	newB.numberOfAssignedCandidates = b.numberOfAssignedCandidates
	newB.autoStake = b.autoStake
	newB.rewardReceiver = b.rewardReceiver
	newB.stakingTx = b.stakingTx
	newB.hashes = b.hashes
}

func (b BeaconCommitteeStateV2) clone(newB *BeaconCommitteeStateV2) {
	newB.reset()
	newB.beaconCommittee = make([]incognitokey.CommitteePublicKey, len(b.beaconCommittee))
	copy(newB.beaconCommittee, b.beaconCommittee)
	newB.numberOfAssignedCandidates = b.numberOfAssignedCandidates
	newB.shardCommonPool = make([]incognitokey.CommitteePublicKey, len(b.shardCommonPool))
	copy(newB.shardCommonPool, b.shardCommonPool)

	for i, v := range b.shardCommittee {
		newB.shardCommittee[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(newB.shardCommittee[i], v)
	}

	for i, v := range b.shardSubstitute {
		newB.shardSubstitute[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(newB.shardSubstitute[i], v)
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
	b.beaconCommittee = []incognitokey.CommitteePublicKey{}
	b.shardCommonPool = []incognitokey.CommitteePublicKey{}
	b.shardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	b.shardSubstitute = make(map[byte][]incognitokey.CommitteePublicKey)
	b.autoStake = make(map[string]bool)
	b.rewardReceiver = make(map[string]privacy.PaymentAddress)
	b.stakingTx = make(map[string]common.Hash)
	b.hashes = NewBeaconCommitteeStateHash()
}

//Clone :
func (engine *BeaconCommitteeEngineV2) Clone() BeaconCommitteeEngine {

	finalCommitteeState := NewBeaconCommitteeStateV2()
	engine.finalBeaconCommitteeStateV2.clone(finalCommitteeState)
	engine.uncommittedBeaconCommitteeStateV2 = NewBeaconCommitteeStateV2()

	res := NewBeaconCommitteeEngineV2(
		engine.beaconHeight,
		engine.beaconHash,
		finalCommitteeState,
	)

	return res
}

//Version :
func (engine BeaconCommitteeEngineV2) Version() uint {
	return SLASHING_VERSION
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

//GetUncommittedCommittee :
func (engine BeaconCommitteeEngineV2) GetUncommittedCommittee() map[byte][]incognitokey.CommitteePublicKey {
	engine.uncommittedBeaconCommitteeStateV2.mu.RLock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.RUnlock()
	shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.uncommittedBeaconCommitteeStateV2.shardCommittee {
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

func (engine *BeaconCommitteeEngineV2) compareHashes(hash1, hash2 *BeaconCommitteeStateHash) error {
	if !hash1.BeaconCommitteeAndValidatorHash.IsEqual(&hash2.BeaconCommitteeAndValidatorHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState,
			fmt.Errorf("Uncommitted BeaconCommitteeAndValidatorHash want value %+v but have %+v",
				hash1.BeaconCommitteeAndValidatorHash, hash2.BeaconCommitteeAndValidatorHash))
	}
	if !hash1.BeaconCandidateHash.IsEqual(&hash2.BeaconCandidateHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState,
			fmt.Errorf("Uncommitted BeaconCandidateHash want value %+v but have %+v",
				hash1.BeaconCandidateHash, hash2.BeaconCandidateHash))
	}
	if !hash1.ShardCandidateHash.IsEqual(&hash2.ShardCandidateHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState,
			fmt.Errorf("Uncommitted ShardCandidateHash want value %+v but have %+v",
				hash1.ShardCandidateHash, hash2.ShardCandidateHash))
	}
	if !hash1.ShardCommitteeAndValidatorHash.IsEqual(&hash2.ShardCommitteeAndValidatorHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState,
			fmt.Errorf("Uncommitted ShardCommitteeAndValidatorHash want value %+v but have %+v",
				hash1.ShardCommitteeAndValidatorHash, hash2.ShardCommitteeAndValidatorHash))
	}
	if !hash1.AutoStakeHash.IsEqual(&hash2.AutoStakeHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState,
			fmt.Errorf("Uncommitted AutoStakingHash want value %+v but have %+v",
				hash1.AutoStakeHash, hash2.AutoStakeHash))
	}
	return nil
}

//Commit is deprecate
func (engine *BeaconCommitteeEngineV2) Commit(hashes *BeaconCommitteeStateHash, committeeChange *CommitteeChange) error {
	engine.uncommittedBeaconCommitteeStateV2.mu.Lock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.Unlock()
	engine.finalBeaconCommitteeStateV2.mu.Lock()
	defer engine.finalBeaconCommitteeStateV2.mu.Unlock()
	engine.uncommittedBeaconCommitteeStateV2.shallowCopy(engine.finalBeaconCommitteeStateV2)
	engine.uncommittedBeaconCommitteeStateV2 = NewBeaconCommitteeStateV2WithMu(engine.uncommittedBeaconCommitteeStateV2.mu)
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
				panic(err)
			}
		}
	}
	b.beaconCommittee = append(b.beaconCommittee, newBeaconCandidates...)
	for shardID := 0; shardID < env.ActiveShards; shardID++ {
		b.shardCommittee[byte(shardID)] = append(b.shardCommittee[byte(shardID)], newShardCandidates[shardID*env.MinShardCommitteeSize:(shardID+1)*env.MinShardCommitteeSize]...)
	}
}

// UpdateCommitteeState New flow
// Store information from instructions into temp stateDB in env
// When all thing done and no problems, in commit function, we read data in statedb and update
func (engine *BeaconCommitteeEngineV2) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var err error
	incurredInstructions := [][]string{}
	returnStakingInstruction := instruction.NewReturnStakeIns()
	committeeChange := NewCommitteeChange()
	oldState := engine.finalBeaconCommitteeStateV2

	oldState.mu.RLock()
	defer oldState.mu.RUnlock()

	oldState.clone(engine.uncommittedBeaconCommitteeStateV2)
	newState := engine.uncommittedBeaconCommitteeStateV2

	newState.mu.Lock()
	defer newState.mu.Unlock()

	newState.setHashes(env.PreviousBlockHashes)
	// snapshot shard common pool in beacon random time
	if env.IsBeaconRandomTime {
		newState.numberOfAssignedCandidates = SnapshotShardCommonPoolV2(
			oldState.shardCommonPool,
			oldState.shardCommittee,
			oldState.shardSubstitute,
			env.NumberOfFixedShardBlockValidator,
			env.MinShardCommitteeSize,
		)
		Logger.log.Infof("Block %+v, Number of Snapshot to Assign Candidate %+v", env.BeaconHeight, newState.numberOfAssignedCandidates)
	}

	env.newUnassignedCommonPool, err = newState.unassignedCommonPool()
	if err != nil {
		return nil, nil, nil, err
	}
	env.newAllSubstituteCommittees, err = newState.getAllSubstituteCommittees()
	if err != nil {
		return nil, nil, nil, err
	}
	env.newAllCandidateSubstituteCommittee = append(env.newUnassignedCommonPool, env.newAllSubstituteCommittees...)

	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		switch inst[0] {
		case instruction.STAKE_ACTION:
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, err = newState.processStakeInstruction(stakeInstruction, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		case instruction.RANDOM_ACTION:
			randomInstruction, err := instruction.ValidateAndImportRandomInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = newState.processAssignWithRandomInstruction(
				randomInstruction.BtcNonce, env.ActiveShards, committeeChange, oldState)
		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = newState.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange, oldState)
		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = newState.processUnstakeInstruction(
				unstakeInstruction, env, committeeChange, returnStakingInstruction, oldState)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		case instruction.SWAP_SHARD_ACTION:
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = newState.processSwapShardInstruction(
				swapShardInstruction, env, committeeChange, returnStakingInstruction, oldState)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		}
	}

	hashes, err := engine.generateCommitteeHashes(newState, committeeChange)
	if !returnStakingInstruction.IsEmpty() {
		incurredInstructions = append(incurredInstructions, returnStakingInstruction.ToString())
	}

	return hashes, committeeChange, incurredInstructions, nil
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
		substitutes := engine.finalBeaconCommitteeStateV2.shardSubstitute[shardID]
		tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
		tempSubstitutes, _ := incognitokey.CommitteeKeyListToString(substitutes)

		swapShardInstruction, _, _, _, _ := createSwapShardInstructionV3(
			shardID,
			tempSubstitutes,
			tempCommittees,
			env.MinShardCommitteeSize,
			env.MaxShardCommitteeSize,
			instruction.SWAP_BY_END_EPOCH,
			env.NumberOfFixedShardBlockValidator,
			env.MissingSignaturePenalty,
		)
		if !swapShardInstruction.IsEmpty() {
			swapShardInstructions = append(swapShardInstructions, swapShardInstruction)
		} else {
			Logger.log.Infof("[swap-instructions] Generate empty instructions beacon hash: %s & height: %v \n", engine.beaconHash, engine.beaconHash)
		}
	}
	return swapShardInstructions, nil
}

func (engine *BeaconCommitteeEngineV2) GenerateAssignInstruction(rand int64, assignOffset int, activeShards int) ([]*instruction.AssignInstruction, []string, map[byte][]string) {
	return []*instruction.AssignInstruction{}, []string{}, make(map[byte][]string)
}

func SnapshotShardCommonPoolV2(
	shardCommonPool []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	numberOfFixedValidator int,
	minCommitteeSize int,
) (numberOfAssignedCandidates int) {
	for k, v := range shardCommittee {
		assignPerShard := getAssignOffset(
			len(shardSubstitute[k]),
			len(v),
			numberOfFixedValidator,
			minCommitteeSize,
		)

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
) (*CommitteeChange, error) {
	var err error
	// var key string
	for index, candidate := range stakeInstruction.PublicKeyStructs {
		committeePublicKey := stakeInstruction.PublicKeys[index]
		b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
		b.autoStake[committeePublicKey] = stakeInstruction.AutoStakingFlag[index]
		b.stakingTx[committeePublicKey] = stakeInstruction.TxStakeHashes[index]
	}
	committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, stakeInstruction.PublicKeyStructs...)
	b.shardCommonPool = append(b.shardCommonPool, stakeInstruction.PublicKeyStructs...)

	return committeeChange, err
}

func (b *BeaconCommitteeStateV2) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	oldState *BeaconCommitteeStateV2,
) *CommitteeChange {
	// b == newstate -> only write
	// oldstate -> only read

	//careful with this variable
	// validators := oldState.getAllCandidateSubstituteCommittee()
	validators := env.newAllCandidateSubstituteCommittee

	for _, committeePublicKey := range stopAutoStakeInstruction.CommitteePublicKeys {
		if common.IndexOfStr(committeePublicKey, validators) == -1 {
			// if not found then delete auto staking data for this public key if present
			if _, ok := oldState.autoStake[committeePublicKey]; ok {
				delete(b.autoStake, committeePublicKey)
			}
		} else {
			// if found in committee list then turn off auto staking
			if _, ok := oldState.autoStake[committeePublicKey]; ok {
				committeeChange = b.stopAutoStake(committeePublicKey, committeeChange)
			}
		}
	}
	return committeeChange
}

func (b *BeaconCommitteeStateV2) processAssignWithRandomInstruction(
	rand int64,
	activeShards int,
	committeeChange *CommitteeChange,
	oldState *BeaconCommitteeStateV2,
) *CommitteeChange {
	// b == newstate -> only write
	// oldstate -> only read
	newCommitteeChange := committeeChange
	candidateStructs := oldState.shardCommonPool[:b.numberOfAssignedCandidates]
	candidates, _ := incognitokey.CommitteeKeyListToString(candidateStructs)
	newCommitteeChange = b.assign(candidates, rand, activeShards, newCommitteeChange, oldState)
	newCommitteeChange.NextEpochShardCandidateRemoved = append(newCommitteeChange.NextEpochShardCandidateRemoved, candidateStructs...)
	b.shardCommonPool = b.shardCommonPool[b.numberOfAssignedCandidates:]
	b.numberOfAssignedCandidates = 0

	return newCommitteeChange
}

func (b *BeaconCommitteeStateV2) assign(
	candidates []string, rand int64, activeShards int, committeeChange *CommitteeChange,
	oldState *BeaconCommitteeStateV2,
) *CommitteeChange {
	numberOfValidator := make([]int, activeShards)
	for i := 0; i < activeShards; i++ {
		numberOfValidator[byte(i)] += len(oldState.shardSubstitute[byte(i)])
		numberOfValidator[byte(i)] += len(oldState.shardCommittee[byte(i)])
	}

	assignedCandidates := assignShardCandidateV2(candidates, numberOfValidator, rand)
	for shardID, tempCandidates := range assignedCandidates {
		tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(tempCandidates)
		committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
		b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], tempCandidateStructs...)
	}
	return committeeChange
}

//processSwapShardInstruction update committees state by swap shard instruction
// Process single swap shard instruction for and update committee state
func (b *BeaconCommitteeStateV2) processSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState *BeaconCommitteeStateV2,
) (
	*CommitteeChange, *instruction.ReturnStakeInstruction, error) {

	var err error
	shardID := byte(swapShardInstruction.ChainID)
	committees := oldState.shardCommittee[shardID]
	substitutes := oldState.shardSubstitute[shardID]
	tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
	tempSubstitutes, _ := incognitokey.CommitteeKeyListToString(substitutes)

	comparedShardSwapInstruction, newCommittees, _,
		slashingCommittees, normalSwapOutCommittees := createSwapShardInstructionV3(
		shardID,
		tempSubstitutes,
		tempCommittees,
		env.MinShardCommitteeSize,
		env.MaxShardCommitteeSize,
		instruction.SWAP_BY_END_EPOCH,
		env.NumberOfFixedShardBlockValidator,
		env.MissingSignaturePenalty,
	)

	if len(slashingCommittees) > 0 {
		Logger.log.Infof("SHARD %+v, Epoch %+v, Slashing Committees %+v", shardID, env.Epoch, slashingCommittees)
	} else {
		Logger.log.Infof("SHARD %+v, Epoch %+v, NO Slashing Committees", shardID, env.Epoch)
	}

	if !reflect.DeepEqual(comparedShardSwapInstruction.InPublicKeys, swapShardInstruction.InPublicKeys) {
		return nil, returnStakingInstruction, fmt.Errorf("expect swap in keys %+v, got %+v",
			comparedShardSwapInstruction.InPublicKeys, swapShardInstruction.InPublicKeys)
	}
	if !reflect.DeepEqual(comparedShardSwapInstruction.OutPublicKeys, swapShardInstruction.OutPublicKeys) {
		return nil, returnStakingInstruction, fmt.Errorf("expect swap out keys %+v, got %+v",
			comparedShardSwapInstruction.OutPublicKeys, swapShardInstruction.OutPublicKeys)
	}

	b.shardCommittee[shardID], _ = incognitokey.CommitteeBase58KeyListToStruct(newCommittees)
	b.shardSubstitute[shardID] = b.shardSubstitute[shardID][len(swapShardInstruction.InPublicKeys):]

	committeeChange.ShardCommitteeRemoved[shardID] = append(committeeChange.ShardCommitteeRemoved[shardID],
		incognitokey.DeepCopy(swapShardInstruction.OutPublicKeyStructs)...)
	committeeChange.ShardSubstituteRemoved[shardID] = append(committeeChange.ShardSubstituteRemoved[shardID],
		incognitokey.DeepCopy(swapShardInstruction.InPublicKeyStructs)...)
	committeeChange.ShardCommitteeAdded[shardID] = append(committeeChange.ShardCommitteeAdded[shardID],
		incognitokey.DeepCopy(swapShardInstruction.InPublicKeyStructs)...)

	// process after swap for assign old committees to current shard pool
	committeeChange, returnStakingInstruction, err = b.processAfterNormalSwap(env,
		normalSwapOutCommittees,
		committeeChange,
		returnStakingInstruction,
		oldState,
	)

	if err != nil {
		return nil, returnStakingInstruction, err
	}

	returnStakingInstruction, committeeChange, err = b.processSlashing(
		env,
		slashingCommittees,
		returnStakingInstruction,
		committeeChange,
	)

	if err != nil {
		return nil, returnStakingInstruction, err
	}

	committeeChange.SlashingCommittee[shardID] = append(committeeChange.SlashingCommittee[shardID], slashingCommittees...)

	return committeeChange, returnStakingInstruction, nil
}

// processAfterNormalSwap process swapped out committee public key
// - auto stake is false then remove completely out of any committee, candidate, substitute list
// - auto stake is true then using assignment rule v2 to assign this committee public key
func (b *BeaconCommitteeStateV2) processAfterNormalSwap(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState *BeaconCommitteeStateV2,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	candidates := []string{}
	outPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
	for index, outPublicKey := range outPublicKeys {
		stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, outPublicKey)
		if err != nil {
			return committeeChange, returnStakingInstruction, err
		}
		if !has {
			return committeeChange, returnStakingInstruction, errors.Errorf("Can not found info of this public key %v", outPublicKey)
		}
		if stakerInfo.AutoStaking() {
			candidates = append(candidates, outPublicKey)
		} else {
			returnStakingInstruction, committeeChange, err = b.buildReturnStakingInstructionAndDeleteStakerInfo(
				returnStakingInstruction,
				outPublicKeyStructs[index],
				outPublicKey,
				stakerInfo,
				committeeChange,
			)

			if err != nil {
				return committeeChange, returnStakingInstruction, err
			}
		}
	}

	committeeChange = b.assign(candidates, env.RandomNumber, env.ActiveShards, committeeChange, oldState)
	return committeeChange, returnStakingInstruction, nil
}

// processAfterNormalSwap process swapped out committee public key
// if number of round is less than MAX_NUMBER_OF_ROUND go back to THAT shard pool, and increase number of round
// if number of round is equal to or greater than MAX_NUMBER_OF_ROUND
// - auto stake is false then remove completely out of any committee, candidate, substitute list
// - auto stake is true then using assignment rule v2 to assign this committee public key
func (b *BeaconCommitteeStateV2) processSlashing(
	env *BeaconCommitteeStateEnvironment,
	slashingPublicKeys []string,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	committeeChange *CommitteeChange,
) (*instruction.ReturnStakeInstruction, *CommitteeChange, error) {
	slashingPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(slashingPublicKeys)
	for index, outPublicKey := range slashingPublicKeys {
		stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, outPublicKey)
		if err != nil {
			return returnStakingInstruction, committeeChange, err
		}
		if !has {
			return returnStakingInstruction, committeeChange, fmt.Errorf("Can not found info of this public key %v", outPublicKey)
		}
		returnStakingInstruction, committeeChange, err = b.buildReturnStakingInstructionAndDeleteStakerInfo(
			returnStakingInstruction,
			slashingPublicKeyStructs[index],
			outPublicKey,
			stakerInfo,
			committeeChange,
		)
		if err != nil {
			return returnStakingInstruction, committeeChange, err
		}
	}

	return returnStakingInstruction, committeeChange, nil
}

//processUnstakeInstruction : process unstake instruction from beacon block
func (b *BeaconCommitteeStateV2) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState *BeaconCommitteeStateV2,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	// b == newstate -> only write
	// oldstate -> only read
	shardCommonPoolStr, _ := incognitokey.CommitteeKeyListToString(b.shardCommonPool)
	for index, publicKey := range unstakeInstruction.CommitteePublicKeys {
		if common.IndexOfStr(publicKey, env.newUnassignedCommonPool) == -1 {
			if common.IndexOfStr(publicKey, env.newAllSubstituteCommittees) != -1 {
				// if found in committee list then turn off auto staking
				if _, ok := oldState.autoStake[publicKey]; ok {
					committeeChange = b.stopAutoStake(publicKey, committeeChange)
				}
			}
		} else {
			indexCandidate := common.IndexOfStr(publicKey, shardCommonPoolStr)
			if indexCandidate == -1 {
				return committeeChange, returnStakingInstruction, errors.Errorf("Committee public key: %s is not valid for any committee sets", publicKey)
			}
			shardCommonPoolStr = append(shardCommonPoolStr[:indexCandidate], shardCommonPoolStr[indexCandidate+1:]...)
			b.shardCommonPool = append(b.shardCommonPool[:indexCandidate], b.shardCommonPool[indexCandidate+1:]...)
			stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, publicKey)
			if err != nil {
				return committeeChange, returnStakingInstruction, err
			}
			if !has {
				return committeeChange, returnStakingInstruction, errors.New("Can't find staker info")
			}

			committeeChange.NextEpochShardCandidateRemoved =
				append(committeeChange.NextEpochShardCandidateRemoved, unstakeInstruction.CommitteePublicKeysStruct[index])

			returnStakingInstruction, committeeChange, err = b.buildReturnStakingInstructionAndDeleteStakerInfo(
				returnStakingInstruction,
				unstakeInstruction.CommitteePublicKeysStruct[index],
				publicKey,
				stakerInfo,
				committeeChange,
			)

			if err != nil {
				return committeeChange, returnStakingInstruction, errors.New("Can't find staker info")
			}
		}
	}

	return committeeChange, returnStakingInstruction, nil
}

func (engine *BeaconCommitteeEngineV2) generateCommitteeHashes(state *BeaconCommitteeStateV2, committeeChange *CommitteeChange) (*BeaconCommitteeStateHash, error) {
	if reflect.DeepEqual(state, NewBeaconCommitteeStateV2()) {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	newB := state
	var tempShardCandidateHash common.Hash
	var tempBeaconCandidateHash common.Hash
	var tempShardCommitteeAndValidatorHash common.Hash
	var tempAutoStakingHash common.Hash
	var tempBeaconCommitteeAndValidatorHash common.Hash
	var err error

	if !isNilOrBeaconCommitteeAndValidatorHash(newB.hashes) &&
		len(committeeChange.BeaconCommitteeReplaced[0]) == 0 && len(committeeChange.BeaconCommitteeReplaced[1]) == 0 {
		tempBeaconCommitteeAndValidatorHash = newB.hashes.BeaconCommitteeAndValidatorHash
	} else {
		beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(newB.beaconCommittee)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		validatorArr := append([]string{}, beaconCommitteeStr...)
		tempBeaconCommitteeAndValidatorHash, err = common.GenerateHashFromStringArray(validatorArr)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
	}

	if !isNilOrShardCandidateHash(newB.hashes) &&
		len(committeeChange.NextEpochShardCandidateRemoved) == 0 && len(committeeChange.NextEpochShardCandidateAdded) == 0 {
		tempShardCandidateHash = newB.hashes.ShardCandidateHash
	} else {
		// Shard candidate root: shard current candidate + shard next candidate
		shardNextEpochCandidateStr, err := incognitokey.CommitteeKeyListToString(newB.shardCommonPool)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		tempShardCandidateHash, err = common.GenerateHashFromStringArray(shardNextEpochCandidateStr)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
	}

	if !isNilOrBeaconCandidateHash(newB.hashes) &&
		len(committeeChange.NextEpochBeaconCandidateRemoved) == 0 && len(committeeChange.NextEpochBeaconCandidateAdded) == 0 {
		tempBeaconCandidateHash = newB.hashes.BeaconCandidateHash
	} else {
		tempBeaconCandidateHash, _ = common.GenerateHashFromStringArray([]string{})
	}

	if !isNilOrShardCommitteeAndValidatorHash(newB.hashes) &&
		len(committeeChange.ShardSubstituteAdded) == 0 && len(committeeChange.ShardSubstituteRemoved) == 0 &&
		len(committeeChange.ShardCommitteeAdded) == 0 && len(committeeChange.ShardCommitteeRemoved) == 0 &&
		len(committeeChange.ShardCommitteeReplaced) == 0 {
		tempShardCommitteeAndValidatorHash = newB.hashes.ShardCommitteeAndValidatorHash
	} else {
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
		tempShardCommitteeAndValidatorHash, err = common.GenerateHashFromMapByteString(shardPendingValidator, shardCommittee)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
	}

	if !isNilOrAutoStakeHash(newB.hashes) &&
		len(committeeChange.StopAutoStake) == 0 {
		tempAutoStakingHash = newB.hashes.AutoStakeHash
	} else {
		tempAutoStakingHash, err = common.GenerateHashFromMapStringBool(newB.autoStake)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
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

func (b *BeaconCommitteeStateV2) stopAutoStake(publicKey string, committeeChange *CommitteeChange) *CommitteeChange {
	b.autoStake[publicKey] = false
	committeeChange.StopAutoStake = append(committeeChange.StopAutoStake, publicKey)
	return committeeChange
}

func (b *BeaconCommitteeStateV2) unassignedCommonPool() ([]string, error) {
	commonPoolValidators := []string{}
	candidateShardWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(b.shardCommonPool[b.numberOfAssignedCandidates:])
	if err != nil {
		return nil, err
	}
	commonPoolValidators = append(commonPoolValidators, candidateShardWaitingForNextRandomStr...)
	return commonPoolValidators, nil
}

func (b *BeaconCommitteeStateV2) getAllSubstituteCommittees() ([]string, error) {
	validators := []string{}

	for _, committee := range b.shardCommittee {
		committeeStr, err := incognitokey.CommitteeKeyListToString(committee)
		if err != nil {
			return nil, err
		}
		validators = append(validators, committeeStr...)
	}
	for _, substitute := range b.shardSubstitute {
		substituteStr, err := incognitokey.CommitteeKeyListToString(substitute)
		if err != nil {
			return nil, err
		}
		validators = append(validators, substituteStr...)
	}

	beaconCommittee := b.beaconCommittee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconCommittee)
	if err != nil {
		return nil, err
	}
	validators = append(validators, beaconCommitteeStr...)
	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(b.shardCommonPool[:b.numberOfAssignedCandidates])
	if err != nil {
		return nil, err
	}
	validators = append(validators, candidateShardWaitingForCurrentRandomStr...)

	return validators, nil
}

//ActiveShards ...
func (engine *BeaconCommitteeEngineV2) ActiveShards() int {
	return len(engine.finalBeaconCommitteeStateV2.shardCommittee)
}

func (b *BeaconCommitteeStateV2) buildReturnStakingInstructionAndDeleteStakerInfo(
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	committeePublicKeyStruct incognitokey.CommitteePublicKey,
	publicKey string,
	stakerInfo *statedb.StakerInfo,
	committeeChange *CommitteeChange,
) (*instruction.ReturnStakeInstruction, *CommitteeChange, error) {
	returnStakingInstruction = buildReturnStakingInstruction(
		returnStakingInstruction,
		publicKey,
		stakerInfo.TxStakingID().String(),
	)
	committeeChange, err := b.deleteStakerInfo(committeePublicKeyStruct, publicKey, committeeChange)
	if err != nil {
		return returnStakingInstruction, committeeChange, err
	}
	return returnStakingInstruction, committeeChange, nil
}

func buildReturnStakingInstruction(
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	publicKey string,
	txStake string,
) *instruction.ReturnStakeInstruction {
	returnStakingInstruction.AddNewRequest(publicKey, txStake)
	return returnStakingInstruction
}

func (b *BeaconCommitteeStateV2) deleteStakerInfo(
	committeePublicKeyStruct incognitokey.CommitteePublicKey,
	committeePublicKey string,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	committeeChange.RemovedStaker = append(committeeChange.RemovedStaker, committeePublicKey)
	delete(b.rewardReceiver, committeePublicKeyStruct.GetIncKeyBase58())
	delete(b.autoStake, committeePublicKey)
	delete(b.stakingTx, committeePublicKey)
	return committeeChange, nil
}
