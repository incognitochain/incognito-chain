package committeestate

import (
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeStateV4 struct {
	*BeaconCommitteeStateV3
	bDelegateState *BeaconDelegateState
	Reputation     map[string]uint64
}

func NewBeaconCommitteeStateV4() *BeaconCommitteeStateV4 {
	return &BeaconCommitteeStateV4{
		BeaconCommitteeStateV3: NewBeaconCommitteeStateV3(),
		bDelegateState:         &BeaconDelegateState{},
		Reputation:             map[string]uint64{},
	}
}

func NewBeaconCommitteeStateV4WithValue(
	beaconCommittee []string,
	shardCommittee map[byte][]string,
	shardSubstitute map[byte][]string,
	shardCommonPool []string,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	delegateList map[string]string,
	syncPool map[byte][]string,
	swapRule SwapRuleProcessor,
	assignRule AssignRuleProcessor,
) *BeaconCommitteeStateV4 {
	res := &BeaconCommitteeStateV4{}
	res.BeaconCommitteeStateV3 = NewBeaconCommitteeStateV3WithValue(beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool, numberOfAssignedCandidates, autoStake, rewardReceiver, stakingTx, delegateList, syncPool, swapRule, assignRule)
	var err error
	res.Reputation = map[string]uint64{}
	res.bDelegateState, err = InitBeaconDelegateState(res.BeaconCommitteeStateV3)
	res.InitReputationState()
	if err != nil {
		panic(err)
	}
	return res
}

func (b *BeaconCommitteeStateV4) Version() int {
	return STAKING_FLOW_V4
}

// shallowCopy maintain dst mutex value
func (b *BeaconCommitteeStateV4) shallowCopy(newB *BeaconCommitteeStateV4) {
	b.BeaconCommitteeStateV3.shallowCopy(newB.BeaconCommitteeStateV3)
	newB.bDelegateState = b.bDelegateState
	newB.Reputation = b.Reputation
}

func (b *BeaconCommitteeStateV4) Clone() BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.clone()
}

func (b *BeaconCommitteeStateV4) clone() *BeaconCommitteeStateV4 {
	newB := NewBeaconCommitteeStateV4()
	newB.BeaconCommitteeStateV3 = b.BeaconCommitteeStateV3.clone()
	newB.bDelegateState = b.bDelegateState.Clone()
	for k, v := range b.Reputation {
		newB.Reputation[k] = v
	}
	return newB
}

func (b BeaconCommitteeStateV4) GetDelegateState() map[string]BeaconDelegatorInfo {
	return b.bDelegateState.GetDelegateState()
}

func (b BeaconCommitteeStateV4) Hash(committeeChange *CommitteeChange) (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	hashes, err := b.BeaconCommitteeStateV3.Hash(committeeChange)
	if err != nil {
		return nil, err
	}

	var tempDelegateStateHash common.Hash
	if !isNilOrDelegateStateHash(b.hashes) &&
		len(committeeChange.ShardCommitteeAdded) == 0 &&
		len(committeeChange.RemovedStaker) == 0 &&
		len(committeeChange.ReDelegate) == 0 {
		tempDelegateStateHash = b.hashes.DelegateStateHash
	} else {
		tempDelegateStateHash = b.bDelegateState.Hash()
	}

	hashes.DelegateStateHash = tempDelegateStateHash

	return hashes, nil
}
func (b BeaconCommitteeStateV4) isEmpty() bool {
	return reflect.DeepEqual(b, NewBeaconCommitteeStateV4())
}

func initGenesisBeaconCommitteeStateV4(env *BeaconCommitteeStateEnvironment) *BeaconCommitteeStateV4 {
	BeaconCommitteeStateV4 := NewBeaconCommitteeStateV4()
	BeaconCommitteeStateV4.initCommitteeState(env)
	BeaconCommitteeStateV4.InitReputationState()
	BeaconCommitteeStateV4.bDelegateState, _ = InitBeaconDelegateState(BeaconCommitteeStateV4)
	return BeaconCommitteeStateV4
}

func (b *BeaconCommitteeStateV4) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
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

	b.addDataToEnvironment(env)
	b.setBeaconCommitteeStateHashes(env.PreviousBlockHashes)

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

		case instruction.SWAP_SHARD_ACTION:
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processSwapShardInstruction(
				swapShardInstruction, env, committeeChange, returnStakingInstruction)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.FINISH_SYNC_ACTION:
			finishSyncInstruction, err := instruction.ValidateAndImportFinishSyncInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processFinishSyncInstruction(
				finishSyncInstruction, env, committeeChange)
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
		case instruction.RE_DELEGATE:
			redelegateInstruction, err := instruction.ValidateAndImportReDelegateInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
				continue
			}
			b.processReDelegateInstruction(redelegateInstruction, env, committeeChange)
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

func (b *BeaconCommitteeStateV4) processReDelegateInstruction(
	redelegateInstruction *instruction.ReDelegateInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) {
	changeMap := map[string]string{}
	for index, committeePublicKey := range redelegateInstruction.CommitteePublicKeys {
		b.delegate[committeePublicKey] = redelegateInstruction.DelegateList[index]
		changeMap[committeePublicKey] = redelegateInstruction.DelegateList[index]
	}
	committeeChange.AddReDelegateInfo(changeMap)
}

// processAssignWithRandomInstruction assign candidates to syncPool
// update beacon state and committeechange
// func (b *BeaconCommitteeStateV4) processAssignWithRandomInstruction(
// 	rand int64,
// 	numberOfValidator []int,
// 	committeeChange *CommitteeChange,
// ) *CommitteeChange {
// 	newCommitteeChange, candidates := b.getCandidatesForRandomAssignment(committeeChange)
// 	assignedCandidates := b.processRandomAssignment(candidates, rand, numberOfValidator)
// 	for shardID, candidates := range assignedCandidates {
// 		newCommitteeChange = b.assignToSyncPool(shardID, candidates, newCommitteeChange)
// 	}
// 	return newCommitteeChange
// }

// func (b *BeaconCommitteeStateV4) processSwapShardInstruction(
// 	swapShardInstruction *instruction.SwapShardInstruction,
// 	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
// 	returnStakingInstruction *instruction.ReturnStakeInstruction,
// ) (
// 	*CommitteeChange,
// 	*instruction.ReturnStakeInstruction,
// 	error,
// ) {
// 	shardID := byte(swapShardInstruction.ChainID)
// 	env.ShardID = shardID
// 	newCommitteeChange := &CommitteeChange{}
// 	var err error
// 	newCommitteeChange, returnStakingInstruction, err = b.BeaconCommitteeStateV3.processSwapShardInstruction(swapShardInstruction, env, committeeChange, returnStakingInstruction)

// 	return newCommitteeChange, returnStakingInstruction, err
// }

func (b *BeaconCommitteeStateV4) ProcessStoreCommitteeStateInfo(
	bBlock *types.BeaconBlock,
	// mView multiview.MultiView,
	cChange *CommitteeChange,
	bStateDB *statedb.StateDB,
	isEndOfEpoch bool,
) error {
	if len(cChange.ReDelegate) != 0 {
		for stakerPubkeyStr, newDelegate := range cChange.ReDelegate {
			if stakerInfo, exist, err := statedb.GetStakerInfo(bStateDB, stakerPubkeyStr); err != nil || !exist {
				return err
			} else {
				if stakerInfo.HasCredit() {
					oldD := ""
					newD := newDelegate
					if delegateChange, exist := b.bDelegateState.NextEpochDelegate[stakerPubkeyStr]; exist {
						oldD = delegateChange.Old
					} else {
						oldD = stakerInfo.Delegate()
					}
					b.bDelegateState.NextEpochDelegate[stakerPubkeyStr] = struct {
						Old string
						New string
					}{
						Old: oldD,
						New: newD,
					}
				}
			}
		}
		if err := statedb.SaveStakerReDelegateInfo(bStateDB, cChange.ReDelegate); err != nil {
			return err
		}
	}
	listEnableCredit := []incognitokey.CommitteePublicKey{}
	for _, addedCommittee := range cChange.ShardCommitteeAdded {
		for _, stakerPubkey := range addedCommittee {
			if stakerPubkeyStr, err := stakerPubkey.ToBase58(); err != nil {
				return err
			} else {
				stakerInfo, exist, err := statedb.GetStakerInfo(bStateDB, stakerPubkeyStr)
				if err != nil || !exist {
					return err
				}
				if !stakerInfo.HasCredit() {
					listEnableCredit = append(listEnableCredit, stakerPubkey)
				}
				b.bDelegateState.AddReDelegate(stakerPubkeyStr, "", stakerInfo.Delegate())
			}
		}
		if err := statedb.EnableStakerCredit(bStateDB, listEnableCredit); err != nil {
			return err
		}
	}
	for _, removedCommittee := range cChange.ShardCommitteeRemoved {
		for _, stakerPubkey := range removedCommittee {
			if stakerPubkeyStr, err := stakerPubkey.ToBase58(); err != nil {
				return err
			} else {
				oldD := ""
				newD := ""
				stakerInfo, exist, err := statedb.GetStakerInfo(bStateDB, stakerPubkeyStr)
				if err != nil || !exist {
					return err
				}
				if delegateChange, exist := b.bDelegateState.NextEpochDelegate[stakerPubkeyStr]; exist {
					oldD = delegateChange.Old
				} else {
					oldD = stakerInfo.Delegate()
				}
				b.bDelegateState.NextEpochDelegate[stakerPubkeyStr] = struct {
					Old string
					New string
				}{
					Old: oldD,
					New: newD,
				}
			}
		}
	}
	if bBlock.Header.Height > 2 {
		if err := b.UpdateBeaconReputationWithBlock(bBlock); err != nil {
			return err
		}
	}
	if isEndOfEpoch {
		if err := b.bDelegateState.AcceptNextEpochChange(); err != nil {
			return err
		}
	}
	return nil
}
