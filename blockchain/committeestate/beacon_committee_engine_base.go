package committeestate

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type beaconCommitteeEngineBase struct {
	beaconHeight     uint64
	beaconHash       common.Hash
	finalState       BeaconCommitteeState
	uncommittedState BeaconCommitteeState
}

func NewBeaconCommitteeEngineBaseWithValue(
	beaconHeight uint64,
	beaconHash common.Hash,
	finalState BeaconCommitteeState) *beaconCommitteeEngineBase {
	Logger.log.Debugf("Init Beacon Committee Engine Base With Height %+v And Beacon Committee State Version %+v", beaconHeight, finalState.Version())
	return &beaconCommitteeEngineBase{
		beaconHeight: beaconHeight,
		beaconHash:   beaconHash,
		finalState:   finalState,
	}
}

func (engine *beaconCommitteeEngineBase) Clone() BeaconCommitteeEngine {
	finalState := cloneBeaconCommitteeStateFrom(engine.finalState)
	res := NewBeaconCommitteeEngineBaseWithValue(
		engine.beaconHeight,
		engine.beaconHash,
		finalState,
	)

	return res
}

//Version :
func (engine beaconCommitteeEngineBase) Version() uint {
	panic("Implement this function")
}

//GetBeaconHeight :
func (engine beaconCommitteeEngineBase) GetBeaconHeight() uint64 {
	return engine.beaconHeight
}

//GetBeaconHash :
func (engine beaconCommitteeEngineBase) GetBeaconHash() common.Hash {
	return engine.beaconHash
}

//GetBeaconCommittee :
func (engine beaconCommitteeEngineBase) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return engine.finalState.BeaconCommittee()
}

//GetBeaconSubstitute :
func (engine beaconCommitteeEngineBase) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetCandidateShardWaitingForCurrentRandom :
func (engine beaconCommitteeEngineBase) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetCandidateBeaconWaitingForCurrentRandom :
func (engine beaconCommitteeEngineBase) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetCandidateShardWaitingForNextRandom :
func (engine beaconCommitteeEngineBase) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return engine.finalState.ShardCommonPool()[engine.finalState.NumberOfAssignedCandidates():]
}

//GetCandidateBeaconWaitingForNextRandom :
func (engine beaconCommitteeEngineBase) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetOneShardCommittee :
func (engine beaconCommitteeEngineBase) GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.finalState.ShardCommittee()[shardID]
}

//GetShardCommittee :
func (engine beaconCommitteeEngineBase) GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalState.Mu().RLock()
	defer engine.finalState.Mu().RUnlock()
	shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalState.ShardCommittee() {
		shardCommittee[k] = v
	}
	return shardCommittee
}

//GetUncommittedCommittee :
func (engine beaconCommitteeEngineBase) GetUncommittedCommittee() map[byte][]incognitokey.CommitteePublicKey {
	engine.uncommittedState.Mu().RLock()
	defer engine.uncommittedState.Mu().RUnlock()
	shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.uncommittedState.ShardCommittee() {
		shardCommittee[k] = v
	}
	return shardCommittee
}

//GetOneShardSubstitute :
func (engine beaconCommitteeEngineBase) GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.finalState.ShardSubstitute()[shardID]
}

//GetShardSubstitute :
func (engine beaconCommitteeEngineBase) GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalState.Mu().RLock()
	defer engine.finalState.Mu().RUnlock()
	shardSubstitute := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalState.ShardSubstitute() {
		shardSubstitute[k] = v
	}
	return shardSubstitute
}

//GetAutoStaking :
func (engine beaconCommitteeEngineBase) GetAutoStaking() map[string]bool {
	engine.finalState.Mu().RLock()
	defer engine.finalState.Mu().RUnlock()
	autoStake := make(map[string]bool)
	for k, v := range engine.finalState.AutoStake() {
		autoStake[k] = v
	}
	return autoStake
}

func (engine beaconCommitteeEngineBase) GetRewardReceiver() map[string]privacy.PaymentAddress {
	engine.finalState.Mu().RLock()
	defer engine.finalState.Mu().RUnlock()
	rewardReceiver := make(map[string]privacy.PaymentAddress)
	for k, v := range engine.finalState.RewardReceiver() {
		rewardReceiver[k] = v
	}
	return rewardReceiver
}

func (engine beaconCommitteeEngineBase) GetStakingTx() map[string]common.Hash {
	engine.finalState.Mu().RLock()
	defer engine.finalState.Mu().RUnlock()
	stakingTx := make(map[string]common.Hash)
	for k, v := range engine.finalState.StakingTx() {
		stakingTx[k] = v
	}
	return stakingTx
}

func (engine beaconCommitteeEngineBase) GetAllCandidateSubstituteCommittee() []string {
	engine.finalState.Mu().RLock()
	defer engine.finalState.Mu().RUnlock()
	return engine.finalState.AllCandidateSubstituteCommittees()
}

func (engine beaconCommitteeEngineBase) SyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	return map[byte][]incognitokey.CommitteePublicKey{}
}

func (engine beaconCommitteeEngineBase) compareHashes(hash1, hash2 *BeaconCommitteeStateHash) error {
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

func (engine *beaconCommitteeEngineBase) Commit(hashes *BeaconCommitteeStateHash) error {
	if engine.uncommittedState.IsEmpty() {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("%+v", engine.uncommittedState))
	}

	engine.uncommittedState.Mu().Lock()
	defer engine.uncommittedState.Mu().Unlock()
	engine.finalState.Mu().Lock()
	defer engine.finalState.Mu().Unlock()
	comparedHashes, err := engine.uncommittedState.Hash()
	if err != nil {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, err)
	}
	err = engine.compareHashes(comparedHashes, hashes)
	if err != nil {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, err)
	}
	engine.finalState = cloneBeaconCommitteeStateFrom(engine.uncommittedState)
	engine.uncommittedState.Reset()
	return nil
}

func (engine *beaconCommitteeEngineBase) AbortUncommittedBeaconState() {
	engine.uncommittedState.Mu().Lock()
	defer engine.uncommittedState.Mu().Unlock()
	engine.uncommittedState.Reset()
}

func (engine *beaconCommitteeEngineBase) InitCommitteeState(env *BeaconCommitteeStateEnvironment) {
	engine.finalState.Mu().Lock()
	defer engine.finalState.Mu().Unlock()
	b := engine.finalState
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		if inst[0] == instruction.STAKE_ACTION {
			stakeInstruction := instruction.ImportInitStakeInstructionFromString(inst)
			for index, candidate := range stakeInstruction.PublicKeyStructs {
				b.RewardReceiver()[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
				b.AutoStake()[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
				b.StakingTx()[stakeInstruction.PublicKeys[index]] = stakeInstruction.TxStakeHashes[index]
			}
			if stakeInstruction.Chain == instruction.BEACON_INST {
				newBeaconCandidates = append(newBeaconCandidates, stakeInstruction.PublicKeyStructs...)
			} else {
				newShardCandidates = append(newShardCandidates, stakeInstruction.PublicKeyStructs...)
			}
			err := statedb.StoreStakerInfo(
				env.ConsensusStateDB,
				stakeInstruction.PublicKeyStructs,
				b.RewardReceiver(),
				b.AutoStake(),
				b.StakingTx(),
			)
			if err != nil {
				panic(err)
			}
		}
	}
	b.SetBeaconCommittees(append(b.BeaconCommittee(), newBeaconCandidates...))
	for shardID := 0; shardID < env.ActiveShards; shardID++ {
		b.ShardCommittee()[byte(shardID)] = append(b.ShardCommittee()[byte(shardID)], newShardCandidates[shardID*env.MinShardCommitteeSize:(shardID+1)*env.MinShardCommitteeSize]...)
	}
}

func (engine *beaconCommitteeEngineBase) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	return nil, nil, [][]string{}, nil
}

// GenerateAllSwapShardInstructions generate swap shard instructions for all shard
// it also assigned swapped out committee back to substitute list if auto stake is true
// generate all swap shard instructions by only swap by the end of epoch (normally execution)
func (engine *beaconCommitteeEngineBase) GenerateAllSwapShardInstructions(
	env *BeaconCommitteeStateEnvironment) (
	[]*instruction.SwapShardInstruction, error) {
	return []*instruction.SwapShardInstruction{}, nil
}

func (engine *beaconCommitteeEngineBase) AssignInstructions(env *BeaconCommitteeStateEnvironment) []*instruction.AssignInstruction {
	return []*instruction.AssignInstruction{}
}

func (engine *beaconCommitteeEngineBase) ActiveShards() int {
	return len(engine.finalState.ShardCommittee())
}

//SplitReward ...
func (engine *beaconCommitteeEngineBase) SplitReward(
	env *BeaconCommitteeStateEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64,
	map[common.Hash]uint64, map[common.Hash]uint64, error,
) {
	panic("Implement this function")
}

func (engine beaconCommitteeEngineBase) NumberOfAssignedCandidates() int {
	panic("Implement this function")
}

func (engine beaconCommitteeEngineBase) AddFinishedSyncValidators([]string) error {
	panic("Implement this function")
}

func (engine beaconCommitteeEngineBase) GenerateFinishSyncInstructions() ([]*instruction.FinishSyncInstruction, error) {
	return []*instruction.FinishSyncInstruction{}, nil
}
