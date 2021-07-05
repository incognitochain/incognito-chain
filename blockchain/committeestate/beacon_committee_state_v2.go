package committeestate

import (
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
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

	assignRule AssignRuleProcessor

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
		uncommittedBeaconCommitteeStateV2: NewBeaconCommitteeStateV2(finalBeaconCommitteeStateV2.assignRule),
	}
}

func NewBeaconCommitteeStateV2(assignRule AssignRuleProcessor) *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
		autoStake:       make(map[string]bool),
		rewardReceiver:  make(map[string]privacy.PaymentAddress),
		stakingTx:       make(map[string]common.Hash),
		hashes:          NewBeaconCommitteeStateHash(),
		assignRule:      assignRule,
		mu:              new(sync.RWMutex),
	}
}
func NewBeaconCommitteeStateV2WithMu(mu *sync.RWMutex, assignRule AssignRuleProcessor) *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
		autoStake:       make(map[string]bool),
		rewardReceiver:  make(map[string]privacy.PaymentAddress),
		stakingTx:       make(map[string]common.Hash),
		assignRule:      assignRule,
		mu:              mu,
	}
}

func NewBeaconCommitteeStateV2WithValue(
	beaconCommittee []string,
	shardCommittee map[byte][]string,
	shardSubstitute map[byte][]string,
	shardCommonPool []string,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	assignRule AssignRuleProcessor,
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
		assignRule:                 assignRule,
		mu:                         new(sync.RWMutex),
	}
}

//shallowCopy maintain dst mutex value
func (b *BeaconCommitteeStateV2) shallowCopy(newB *BeaconCommitteeStateV2) {
	newB.beaconCommitteeStateSlashingBase = b.beaconCommitteeStateSlashingBase
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
	newB.assignRule = b.assignRule
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

	finalCommitteeState := NewBeaconCommitteeStateV2(engine.finalBeaconCommitteeStateV2.assignRule)
	engine.finalBeaconCommitteeStateV2.clone(finalCommitteeState)
	engine.uncommittedBeaconCommitteeStateV2 = NewBeaconCommitteeStateV2(engine.finalBeaconCommitteeStateV2.assignRule)

	res := NewBeaconCommitteeEngineV2(
		engine.beaconHeight,
		engine.beaconHash,
		finalCommitteeState,
	)

	return res
}

//Version :
func (b *BeaconCommitteeStateV2) Version() int {
	return SLASHING_VERSION
}

//Version :
func (engine BeaconCommitteeEngineV2) AssignRuleVersion() uint {
	_, ok := engine.finalBeaconCommitteeStateV2.assignRule.(*AssignRuleV2)
	if ok {
		return ASSIGN_RULE_V2
	}

	_, ok = engine.finalBeaconCommitteeStateV2.assignRule.(*AssignRuleV3)
	if ok {
		return ASSIGN_RULE_V3
	}

	panic("unknown version")
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

//ActiveShards ...
func (engine *BeaconCommitteeEngineV2) UpgradeAssignRuleV3() {
	engine.finalBeaconCommitteeStateV2.mu.Lock()
	defer engine.finalBeaconCommitteeStateV2.mu.Unlock()
	engine.finalBeaconCommitteeStateV2.assignRule = AssignRuleV3{}
	engine.uncommittedBeaconCommitteeStateV2.assignRule = AssignRuleV3{}
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
	engine.uncommittedBeaconCommitteeStateV2 = NewBeaconCommitteeStateV2WithMu(
		engine.uncommittedBeaconCommitteeStateV2.mu,
		engine.uncommittedBeaconCommitteeStateV2.assignRule)
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
func (b *BeaconCommitteeStateV2) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var err error
	incurredInstructions := [][]string{}
	returnStakingInstruction := instruction.NewReturnStakeIns()
	committeeChange := NewCommitteeChange()
	b.mu.Lock()
	defer b.mu.Unlock()
	// snapshot shard common pool in beacon random time
	if env.IsBeaconRandomTime {
		b.numberOfAssignedCandidates = SnapshotShardCommonPoolV2(
			b.shardCommonPool,
			b.shardCommittee,
			b.shardSubstitute,
			env.NumberOfFixedShardBlockValidator,
			env.MinShardCommitteeSize,
			b.swapRule,
		)

		Logger.log.Infof("Block %+v, Number of Snapshot to Assign Candidate %+v", env.BeaconHeight, b.numberOfAssignedCandidates)
	}

	b.addData(env)
	b.setHashes(env.PreviousBlockHashes)

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
			committeeChange, err = b.processStakeInstruction(stakeInstruction, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.RANDOM_ACTION:
			randomInstruction, err := instruction.ValidateAndImportRandomInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processAssignWithRandomInstruction(
				randomInstruction.RandomNumber(), env.numberOfValidator, committeeChange)

		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange)

		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processUnstakeInstruction(
				unstakeInstruction, env, committeeChange, returnStakingInstruction)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.SWAP_SHARD_ACTION:
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processSwapShardInstruction(
				swapShardInstruction, env.numberOfValidator, env, committeeChange, returnStakingInstruction)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		}
	}

	hashes, err := b.Hash(committeeChange)
	if err != nil {
		return hashes, committeeChange, incurredInstructions, err
	}
	if !returnStakingInstruction.IsEmpty() {
		incurredInstructions = append(incurredInstructions, returnStakingInstruction.ToString())
	}

	return hashes, committeeChange, incurredInstructions, nil
}

//Upgrade check interface method for des
func (b *BeaconCommitteeStateV2) Upgrade(env *BeaconCommitteeStateEnvironment) BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule := b.getDataForUpgrading(env)

	committeeStateV3 := NewBeaconCommitteeStateV3WithValue(
		beaconCommittee,
		shardCommittee,
		shardSubstitute,
		shardCommonPool,
		numberOfAssignedCandidates,
		autoStake,
		rewardReceiver,
		stakingTx,
		map[byte][]string{},
		swapRule,
	)
	return committeeStateV3
}

func (b *BeaconCommitteeStateV2) getDataForUpgrading(env *BeaconCommitteeStateEnvironment) (
	[]string,
	map[byte][]string,
	map[byte][]string,
	[]string,
	int,
	map[string]bool,
	map[string]privacy.PaymentAddress,
	map[string]common.Hash,
	SwapRuleProcessor,
) {
	shardCommittee := make(map[byte][]string)
	shardSubstitute := make(map[byte][]string)
	numberOfAssignedCandidates := b.numberOfAssignedCandidates
	autoStake := make(map[string]bool)
	rewardReceiver := make(map[string]privacy.PaymentAddress)
	stakingTx := make(map[string]common.Hash)
	swapRule := b.swapRule

	beaconCommittee := common.DeepCopyString(b.beaconCommittee)

	for shardID, oneShardCommittee := range b.shardCommittee {
		shardCommittee[shardID] = common.DeepCopyString(oneShardCommittee)
	}
	for shardID, oneShardSubsitute := range b.shardSubstitute {
		shardSubstitute[shardID] = common.DeepCopyString(oneShardSubsitute)
	}
	nextEpochShardCandidate := b.shardCommonPool[numberOfAssignedCandidates:]
	currentEpochShardCandidate := b.shardCommonPool[:numberOfAssignedCandidates]
	shardCandidates := append(currentEpochShardCandidate, nextEpochShardCandidate...)

	shardCommonPool := common.DeepCopyString(shardCandidates)
	for k, v := range b.autoStake {
		autoStake[k] = v
	}
	for k, v := range b.rewardReceiver {
		rewardReceiver[k] = v
	}
	for k, v := range b.stakingTx {
		stakingTx[k] = v
	}

	assignedCandidates := b.assignRule.Process(candidates, numberOfValidator, rand)
	for shardID, tempCandidates := range assignedCandidates {
		tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(tempCandidates)
		committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
		b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], tempCandidateStructs...)
	}
	return committeeChange
}

//SplitReward ...
func (b *BeaconCommitteeStateV2) SplitReward(
	env *SplitRewardEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64,
	map[common.Hash]uint64, map[common.Hash]uint64, error,
) {
	// @NOTICE: No use split rule reward v2
	/*
		devPercent := uint64(env.DAOPercent)
		allCoinTotalReward := env.TotalReward
		rewardForBeacon := map[common.Hash]uint64{}
		rewardForShard := map[common.Hash]uint64{}
		rewardForIncDAO := map[common.Hash]uint64{}
		rewardForCustodian := map[common.Hash]uint64{}
		lenBeaconCommittees := uint64(len(b.getBeaconCommittee()))
		lenShardCommittees := uint64(len(b.getShardCommittee()[env.ShardID]))

		if len(allCoinTotalReward) == 0 {
			Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
			return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
		}

		for key, totalReward := range allCoinTotalReward {
			totalRewardForDAOAndCustodians := devPercent * totalReward / 100
			totalRewardForShardAndBeaconValidators := totalReward - totalRewardForDAOAndCustodians
			shardWeight := float64(lenShardCommittees)
			beaconWeight := 2 * float64(lenBeaconCommittees) / float64(len(b.shardCommittee))
			totalValidatorWeight := shardWeight + beaconWeight

			rewardForShard[key] = uint64(shardWeight * float64(totalRewardForShardAndBeaconValidators) / totalValidatorWeight)
			Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n",
				key.String(), totalRewardForDAOAndCustodians)

			if env.IsSplitRewardForCustodian {
				rewardForCustodian[key] += env.PercentCustodianReward * totalRewardForDAOAndCustodians / 100
				rewardForIncDAO[key] += totalRewardForDAOAndCustodians - rewardForCustodian[key]
			} else {
				rewardForIncDAO[key] += totalRewardForDAOAndCustodians
			}
			rewardForBeacon[key] += totalReward - (rewardForShard[key] + totalRewardForDAOAndCustodians)
		}

		return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
	**/

func (engine *BeaconCommitteeEngineV2) generateCommitteeHashes(state *BeaconCommitteeStateV2, committeeChange *CommitteeChange) (*BeaconCommitteeStateHash, error) {
	if reflect.DeepEqual(state, NewBeaconCommitteeStateV2(engine.finalBeaconCommitteeStateV2.assignRule)) {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	newB := state
	var tempShardCandidateHash common.Hash
	var tempBeaconCandidateHash common.Hash
	var tempShardCommitteeAndValidatorHash common.Hash
	var tempAutoStakingHash common.Hash
	var tempBeaconCommitteeAndValidatorHash common.Hash
	var err error

	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
	}

	for key, totalReward := range allCoinTotalReward {
		rewardForBeacon[key] += 2 * ((100 - devPercent) * totalReward) / ((uint64(env.ActiveShards) + 2) * 100)
		totalRewardForDAOAndCustodians := uint64(devPercent) * totalReward / uint64(100)

		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n",
			key.String(), totalRewardForDAOAndCustodians)

		if env.IsSplitRewardForCustodian {
			rewardForCustodian[key] += uint64(env.PercentCustodianReward) * totalRewardForDAOAndCustodians / uint64(100)
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians - rewardForCustodian[key]
		} else {
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians
		}

		rewardForShard[key] = totalReward - (rewardForBeacon[key] + totalRewardForDAOAndCustodians)
	}

	return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
}

func (b *beaconCommitteeStateSlashingBase) addData(env *BeaconCommitteeStateEnvironment) {
	env.newUnassignedCommonPool = common.DeepCopyString(b.shardCommonPool[b.numberOfAssignedCandidates:])
	env.newAllSubstituteCommittees, _ = b.getAllSubstituteCommittees()
	env.newAllRoles = append(env.newUnassignedCommonPool, env.newAllSubstituteCommittees...)
	env.shardCommittee = make(map[byte][]string)
	for shardID, committees := range b.shardCommittee {
		env.shardCommittee[shardID] = common.DeepCopyString(committees)
	}
	env.shardSubstitute = make(map[byte][]string)
	for shardID, substitutes := range b.shardSubstitute {
		env.shardSubstitute[shardID] = common.DeepCopyString(substitutes)
	}
	env.numberOfValidator = make([]int, env.ActiveShards)
	for i := 0; i < env.ActiveShards; i++ {
		env.numberOfValidator[i] += len(b.shardCommittee[byte(i)])
		env.numberOfValidator[i] += len(b.shardSubstitute[byte(i)])
	}
}

func (b *BeaconCommitteeStateV2) Clone() BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.clone()
}

func (b *BeaconCommitteeStateV2) clone() *BeaconCommitteeStateV2 {
	res := NewBeaconCommitteeStateV2()
	res.beaconCommitteeStateSlashingBase = *b.beaconCommitteeStateSlashingBase.clone()
	return res
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

//VersionByBeaconHeight get version of committee engine by beaconHeight and config of blockchain
func SFV2VersionAssignRule(beaconHeight, assignRuleV2, assignRuleV3 uint64) AssignRuleProcessor {

	if beaconHeight >= assignRuleV3 {
		Logger.log.Infof("Beacon Height %+v, using Assign Rule V3", beaconHeight)
		return NewAssignRuleV3()

	}

	Logger.log.Infof("Beacon Height %+v, using Assign Rule V2", beaconHeight)

	if beaconHeight >= assignRuleV2 {

		return NewAssignRuleV2()
	}

	return NewAssignRuleV2()
}
