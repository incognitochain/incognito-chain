package committeestate

import (
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type BeaconCommitteeStateV4 struct {
	*BeaconCommitteeStateV3
	beaconSubstitute []string
	beaconWaiting    []string
	beaconLocking    []*LockingInfo
	//TODO refactor
	waitingStatus []byte

	beaconStakingAmount map[string]uint64
	delegate            map[string]string

	bDelegateState *BeaconDelegateState
	Performance    map[string]uint64
	Reputation     map[string]uint64
}

func NewBeaconCommitteeStateV4() *BeaconCommitteeStateV4 {
	return &BeaconCommitteeStateV4{
		BeaconCommitteeStateV3: NewBeaconCommitteeStateV3(),
		bDelegateState:         &BeaconDelegateState{},
		Performance:            map[string]uint64{},
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
	res.BeaconCommitteeStateV3 = NewBeaconCommitteeStateV3WithValue(beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool, numberOfAssignedCandidates, autoStake, rewardReceiver, stakingTx, syncPool, swapRule, assignRule)
	var err error
	res.Performance = map[string]uint64{}
	res.Reputation = map[string]uint64{}
	res.delegate = delegateList
	res.beaconStakingAmount = map[string]uint64{}
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
	newB.beaconSubstitute = b.beaconSubstitute
	newB.beaconWaiting = b.beaconWaiting
	newB.waitingStatus = b.waitingStatus
	newB.beaconLocking = b.beaconLocking
	newB.delegate = b.delegate
	newB.beaconStakingAmount = b.beaconStakingAmount
	newB.bDelegateState = b.bDelegateState
	newB.Performance = b.Performance
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
	for k, v := range b.Performance {
		newB.Performance[k] = v
	}
	newB.beaconSubstitute = []string{}
	newB.beaconSubstitute = append(newB.beaconSubstitute, b.beaconSubstitute...)
	newB.beaconWaiting = []string{}
	newB.beaconWaiting = append(newB.beaconWaiting, b.beaconWaiting...)
	newB.waitingStatus = []byte{}
	newB.waitingStatus = append(newB.waitingStatus, b.waitingStatus...)
	newB.beaconLocking = []*LockingInfo{}
	newB.beaconLocking = append(newB.beaconLocking, b.beaconLocking...)
	newB.delegate = map[string]string{}
	for k, v := range b.delegate {
		newB.delegate[k] = v
	}
	newB.beaconStakingAmount = map[string]uint64{}
	for k, v := range b.beaconStakingAmount {
		newB.beaconStakingAmount[k] = v
	}
	return newB
}

func (b BeaconCommitteeStateV4) GetDelegateState() map[string]BeaconDelegatorInfo {
	return b.bDelegateState.GetDelegateState()
}

func (b BeaconCommitteeStateV4) GetBCStakingAmount() map[string]uint64 {
	return b.beaconStakingAmount
}

func (b *BeaconCommitteeStateV4) GetReputation() map[string]uint64 {
	res := map[string]uint64{}
	for k, v := range b.Reputation {
		res[k] = v
	}
	return res
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
		case instruction.ADD_STAKING_ACTION:
			addStakingInstruction, err := instruction.ValidateAndImportAddStakingInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
				continue
			}
			b.processAddStakingInstruction(addStakingInstruction, env, committeeChange)
		}

	}

	if env.IsBeaconChangeTime {
		var y []string
		_, y, committeeChange, err = b.processAssignBeacon(committeeChange, env)
		_ = y
		if err != nil {
			return nil, nil, nil, err
		} else {
			env.beaconSubstitute = b.beaconSubstitute
			env.beaconWaiting = b.beaconWaiting
		}
		committeeChange, _, _, err = b.processSwapAndSlashBeacon(env, committeeChange)
		if err != nil {
			return nil, nil, nil, err
		}
		returnBeaconInst, err := b.processBeaconLocking(env, committeeChange)
		if err != nil {
			return nil, nil, nil, err
		}
		if returnBeaconInst != nil {
			incurredInstructions = append(incurredInstructions, returnBeaconInst.ToString())
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

func (b *BeaconCommitteeStateV4) processAddStakingInstruction(
	addStakingInstruction *instruction.AddStakingInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) {
	changeMap := map[string]struct {
		TxID          []string
		StakingAmount uint64
	}{}
	for index, committeePublicKey := range addStakingInstruction.CommitteePublicKeys {
		newInfo := struct {
			TxID          []string
			StakingAmount uint64
		}{
			TxID:          []string{},
			StakingAmount: 0,
		}
		if info, ok := changeMap[committeePublicKey]; ok {
			newInfo = info
		}
		newInfo.TxID = append(newInfo.TxID, addStakingInstruction.StakingTxIDs[index])
		newInfo.StakingAmount += addStakingInstruction.StakingAmount[index]
		changeMap[committeePublicKey] = newInfo
	}
	committeeChange.AddAddStakingInfo(changeMap)
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

func (b *BeaconCommitteeStateV4) processAssignBeacon(
	committeeChange *CommitteeChange,
	env *BeaconCommitteeStateEnvironment,
) (
	newSubtitute []string,
	newWaiting []string,
	newCommitteeChange *CommitteeChange,
	err error,
) {
	envAssign := AssignEnvironment{
		ConsensusStateDB:   env.ConsensusStateDB,
		delegateState:      b.bDelegateState.DelegateInfo,
		shardCommittee:     env.shardCommittee,
		shardSubstitute:    env.shardSubstitute,
		shardNewCandidates: env.newAllRoles,
	}
	acceptedCandidate, notAcceptedCandidate, waitingStatus := b.assignRule.ProcessBeacon(env.beaconWaiting, b.waitingStatus, &envAssign)
	if len(acceptedCandidate) != 0 {
		acceptedCandidateStr, err := incognitokey.CommitteeBase58KeyListToStruct(acceptedCandidate)
		if err != nil {
			return nil, nil, nil, err
		}
		committeeChange.BeaconSubstituteAdded = append(committeeChange.BeaconSubstituteAdded, acceptedCandidateStr...)
		committeeChange.CurrentEpochBeaconCandidateRemoved = append(committeeChange.CurrentEpochBeaconCandidateRemoved, acceptedCandidateStr...)
		b.beaconSubstitute = append(b.beaconSubstitute, acceptedCandidate...)
		b.waitingStatus = waitingStatus
	}
	b.beaconWaiting = notAcceptedCandidate
	return b.beaconSubstitute, b.beaconWaiting, committeeChange, nil
}

func (b *BeaconCommitteeStateV4) processSwapAndSlashBeacon(
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (
	newCommitteeChange *CommitteeChange,
	outPublicKeys []string,
	slashedCommittee []string,
	err error,
) {
	bCStateDB := env.ConsensusStateDB
	beaconStatus := map[string]string{}
	swapInKeys := []string{}
	addedsubtitute := []string{}
	for _, v := range env.beaconCommittee {
		beaconStatus[v] = "committee"
	}
	for _, v := range env.beaconSubstitute {
		beaconStatus[v] = "pending"
	}
	newCommittee, newSubtitute, swapOutKeys, slashedCommittee := b.swapRule.ProcessBeacon(
		env.beaconCommittee,
		env.beaconSubstitute,
		env.MinBeaconCommitteeSize,
		env.MaxBeaconCommitteeSize,
		env.NumberOfFixedBeaconBlockValidator,
		b.Reputation,
		b.Performance,
	)
	for _, outPublicKey := range swapOutKeys {
		if stakerInfo, has, err := statedb.GetShardStakerInfo(bCStateDB, outPublicKey); (err != nil) || (!has) {
			err = errors.Errorf("Can not found staker info for pk %v at block %v - %v err %v", outPublicKey, env.BeaconHeight, env.BeaconHash, err)
			return nil, nil, nil, err
		} else {
			if stakerInfo.AutoStaking() {
				newSubtitute = append(newSubtitute, outPublicKey)
				addedsubtitute = append(addedsubtitute, outPublicKey)
			} else {
				outPublicKeys = append(outPublicKeys, outPublicKey)
			}
		}
	}
	for _, pk := range newCommittee {
		if beaconStatus[pk] == "pending" {
			swapInKeys = append(swapInKeys, pk)
		}
	}
	if swapInKeysStr, err := incognitokey.CommitteeKeyListToStruct(swapInKeys); err != nil {
		committeeChange.BeaconCommitteeAdded = append(committeeChange.BeaconCommitteeAdded, swapInKeysStr...)
		committeeChange.BeaconSubstituteRemoved = append(committeeChange.BeaconSubstituteRemoved, swapInKeysStr...)
	}
	if swapOutKeysStr, err := incognitokey.CommitteeKeyListToStruct(swapOutKeys); err != nil {
		committeeChange.BeaconCommitteeRemoved = append(committeeChange.BeaconCommitteeRemoved, swapOutKeysStr...)
	}
	if addedSubstituteStr, err := incognitokey.CommitteeKeyListToStruct(addedsubtitute); err != nil {
		committeeChange.BeaconSubstituteAdded = append(committeeChange.BeaconSubstituteAdded, addedSubstituteStr...)
	}
	if len(slashedCommittee) > 0 {
		committeeChange.SlashingCommittee[common.BeaconChainSyncID] = append(committeeChange.SlashingCommittee[common.BeaconChainSyncID], slashedCommittee...)
	}
	b.beaconCommittee = newCommittee
	b.beaconSubstitute = newSubtitute
	return committeeChange, outPublicKeys, slashedCommittee, nil
}

// TODO refactor this code
func (b *BeaconCommitteeStateV4) ProcessStoreCommitteeStateInfo(
	bBlock *types.BeaconBlock,
	signatureCounter map[string]signaturecounter.MissingSignature,
	cChange *CommitteeChange,
	bStateDB *statedb.StateDB,
	isEndOfEpoch bool,
) error {
	if len(cChange.ReDelegate) != 0 {
		for stakerPubkeyStr, newDelegate := range cChange.ReDelegate {
			if stakerInfo, exist, err := statedb.GetShardStakerInfo(bStateDB, stakerPubkeyStr); err != nil || !exist {
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
				stakerInfo, exist, err := statedb.GetShardStakerInfo(bStateDB, stakerPubkeyStr)
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
	for _, removedCommittee := range cChange.SlashingCommittee {
		for _, stakerPubkeyStr := range removedCommittee {
			oldD := ""
			newD := ""
			stakerInfo, exist, err := statedb.GetShardStakerInfo(bStateDB, stakerPubkeyStr)
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
	if bBlock.Header.Height > 2 {
		if err := b.UpdateBeaconReputationWithBlock(bBlock); err != nil {
			return err
		}
	}
	if isEndOfEpoch {
		if err := b.bDelegateState.AcceptNextEpochChange(); err != nil {
			return err
		}
		for sID, outKeys := range b.shardCommittee {
			processList := outKeys[config.Param().CommitteeSize.NumberOfFixedShardBlockValidator:]
			processList = processList[:len(processList)-len(cChange.ShardCommitteeAdded[sID])]
			for _, outKey := range processList {
				if sInfor, has, err := statedb.GetShardStakerInfo(bStateDB, outKey); (!has) || (err != nil) {
					err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", outKey, has, err)
					return err
				} else {
					activeTimes := 0
					if votePercent, has := signatureCounter[outKey]; has {
						if votePercent.VotePercent >= 90 {
							activeTimes = sInfor.ActiveTimesInCommittee() + 1
						}
					}
					sInfor.SetActiveTimesInCommittee(activeTimes)
					if err := statedb.StoreShardStakerInfoObject(bStateDB, outKey, sInfor); err != nil {
						return err
					}
				}
			}
		}
	}
	for _, newStakerBeacon := range cChange.BeaconStakerKeys() {
		stakerKey, err := newStakerBeacon.ToBase58()
		if err != nil {
			return err
		}
		b.bDelegateState.AddBeaconCandidate(stakerKey, b.beaconStakingAmount[stakerKey])
		b.Reputation[stakerKey] = b.beaconStakingAmount[stakerKey] / 2
		b.Performance[stakerKey] = 500
	}
	for _, outKeys := range cChange.SwapoutAndBackToPending {
		for _, outKey := range outKeys {
			if sInfor, has, err := statedb.GetShardStakerInfo(bStateDB, outKey); (!has) || (err != nil) {
				err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", outKey, has, err)
				return err
			} else {
				activeTimes := 0
				if votePercent, has := signatureCounter[outKey]; has {
					if votePercent.VotePercent >= 90 {
						activeTimes = sInfor.ActiveTimesInCommittee() + 1
					}
				}
				sInfor.SetActiveTimesInCommittee(activeTimes)
				if err := statedb.StoreShardStakerInfoObject(bStateDB, outKey, sInfor); err != nil {
					return err
				}
			}
		}
	}
	for publicKey, addStakeInfo := range cChange.AddStakingInfo {
		if bInfor, has, err := statedb.GetBeaconStakerInfo(bStateDB, publicKey); (!has) || (err != nil) {
			err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", publicKey, has, err)
			return err
		} else {
			b.bDelegateState.AddStakingAmount(publicKey, addStakeInfo.StakingAmount)
			txAddStakeNew := []common.Hash{}
			for _, v := range addStakeInfo.TxID {
				if txHash, err := (common.Hash{}).NewHashFromStr(v); err != nil {
					return err
				} else {
					txAddStakeNew = append(txAddStakeNew, *txHash)
				}
			}
			txHashes := append(bInfor.TxStakingIDs(), txAddStakeNew...)
			newAmount := bInfor.StakingAmount() + addStakeInfo.StakingAmount
			bInfor.SetTxStakingIDs(txHashes)
			bInfor.SetStakingAmount(newAmount)

			if err := statedb.StoreBeaconStakerInfoObject(bStateDB, publicKey, bInfor); err != nil {
				return err
			}
		}
	}
	for pk, _ := range b.bDelegateState.DelegateInfo {
		perf := uint64(0)
		if v, ok := b.Performance[pk]; ok {
			perf = v
		}
		vpow := b.bDelegateState.GetBeaconCandidatePower(pk)
		b.Reputation[pk] = perf * vpow / 1000
	}
	return nil
}

func (b BeaconCommitteeStateV4) GetBeaconWaiting() []incognitokey.CommitteePublicKey {
	bPKStructs, err := incognitokey.CommitteeKeyListToStruct(b.beaconWaiting)
	if err != nil {
		panic(err)
	}
	return bPKStructs
}

func (b BeaconCommitteeStateV4) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	bPKStructs, err := incognitokey.CommitteeKeyListToStruct(b.beaconSubstitute)
	if err != nil {
		panic(err)
	}
	return bPKStructs
}

func (b *BeaconCommitteeStateV4) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	newCommitteeChange, err := b.BeaconCommitteeStateV3.processStakeInstruction(stakeInstruction, committeeChange)
	if err != nil {
		return nil, err
	}
	if stakeInstruction.Chain == instruction.BEACON_INST {
		committeeChange.CurrentEpochBeaconCandidateAdded = append(committeeChange.CurrentEpochBeaconCandidateAdded, stakeInstruction.PublicKeyStructs...)
	}
	for index, committeePublicKey := range stakeInstruction.PublicKeys {
		if stakeInstruction.Chain == instruction.SHARD_INST {
			b.delegate[committeePublicKey] = stakeInstruction.DelegateList[index]
		} else {
			b.beaconStakingAmount[committeePublicKey] = stakeInstruction.StakingAmount[index]
		}
	}
	newWaitingStr, err := incognitokey.CommitteeKeyListToString(newCommitteeChange.CurrentEpochBeaconCandidateAdded)
	if err != nil {
		return nil, err
	}
	newWaitingStatus := []byte{}
	for range newWaitingStr {
		newWaitingStatus = append(newWaitingStatus, 0)
	}
	b.waitingStatus = append(b.waitingStatus, newWaitingStatus...)
	b.beaconWaiting = append(b.beaconWaiting, newWaitingStr...)
	return newCommitteeChange, err
}

func (b BeaconCommitteeStateV4) GetDelegate() map[string]string {
	res := map[string]string{}
	for k, v := range b.delegate {
		res[k] = v
	}
	return res
}

func (b *BeaconCommitteeStateV4) processSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) (
	*CommitteeChange,
	*instruction.ReturnStakeInstruction,
	error,
) {
	shardID := byte(swapShardInstruction.ChainID)
	env.ShardID = shardID
	candidateFromCommittee := []string{}
	// process normal swap out
	newCommitteeChange, _, normalSwapOutCommittees, slashingCommittees, err := b.processSwap(swapShardInstruction, env, committeeChange)
	if err != nil {
		return nil, nil, err
	}
	enoughActivesTimes, err := b.countActiveTimes(env, normalSwapOutCommittees)
	if err != nil {
		return nil, nil, err
	}
	//TODO refactor
	enoughActivesTimesM := map[string]interface{}{}
	for _, key := range enoughActivesTimes {
		enoughActivesTimesM[key] = nil
	}

	// process after swap for assign old committees to current shard pool
	candidateFromCommittee, newCommitteeChange, returnStakingInstruction, err = b.processAfterNormalSwap(env,
		normalSwapOutCommittees,
		newCommitteeChange,
		returnStakingInstruction,
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}

	for i, returnStakingReason := range returnStakingInstruction.Reasons {
		if _, ok := enoughActivesTimesM[returnStakingInstruction.PublicKeys[i]]; ok {
			returnStakingInstruction.Reasons[i] = common.RETURN_BY_PROMOTE
		} else if returnStakingReason == common.RETURN_BY_UNKNOWN {
			returnStakingInstruction.Reasons[i] = common.RETURN_BY_UNSTAKE
		}

	}
	// process slashing after normal swap out
	returnStakingInstruction, newCommitteeChange, err = b.processSlashing(
		shardID,
		env,
		slashingCommittees,
		returnStakingInstruction,
		newCommitteeChange,
	)
	for i, returnStakingReason := range returnStakingInstruction.Reasons {
		if returnStakingReason == common.RETURN_BY_UNKNOWN {
			returnStakingInstruction.Reasons[i] = common.RETURN_BY_SLASHED
		}
	}

	if err != nil {
		return nil, returnStakingInstruction, err
	}
	newCommitteeChange.SwapoutAndBackToPending[shardID] = candidateFromCommittee

	return newCommitteeChange, returnStakingInstruction, nil
}

func (b *BeaconCommitteeStateV4) processBeaconLocking(
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (
	*instruction.ReturnBeaconStakeInstruction,
	error,
) {
	if len(committeeChange.SlashingCommittee[common.BeaconChainSyncID]) > 0 {
		slashedPK := committeeChange.SlashingCommittee[common.BeaconChainSyncID]
		for _, staking := range slashedPK {
			totalEpoch := (100 * (2000 - b.Performance[staking])) / 1000
			fmt.Printf("Locked infor: Lock candidate %v, total epoch %v, to epoch %v, their performance %v\n", staking[len(staking)-5:], totalEpoch, env.Epoch+totalEpoch, b.Performance[staking])
			totalEpoch = 5
			b.LockNewCandidate(staking, env.Epoch+totalEpoch, common.RETURN_BY_SLASHED)
		}
	}
	returnLocking := b.GetReturnStakingInstruction(env.ConsensusStateDB, env.Epoch)
	return returnLocking, nil
}

func (b *BeaconCommitteeStateV4) countActiveTimes(
	env *BeaconCommitteeStateEnvironment,
	processList []string,
) (
	[]string,
	error,
) {
	if (len(processList) > 0) && (len(env.beaconWaiting) > 0) {
		fmt.Printf("Start count active times for waiting %+v processList %+v\n", common.ShortPKList(env.beaconWaiting), common.ShortPKList(processList))
	} else {
		if len(processList) > 0 {
			fmt.Printf("Start count active times for waiting %+v processList %+v\n", common.ShortPKList(env.beaconWaiting), common.ShortPKList(processList))
		}
	}

	bStateDB := env.ConsensusStateDB
	signatureCounter := env.MissingSignature
	enoughActivetimes := []string{}
	processList = common.IntersectionString(env.beaconWaiting, processList)
	for _, outKey := range processList {
		if sInfor, has, err := statedb.GetShardStakerInfo(bStateDB, outKey); (!has) || (err != nil) {
			err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", outKey, has, err)
			return nil, err
		} else {
			activeTimes := 0
			if voteInfo, has := signatureCounter[outKey]; has {
				votePercent := 100 - voteInfo.Missing*100/voteInfo.ActualTotal
				if votePercent >= 90 {
					activeTimes = sInfor.ActiveTimesInCommittee() + 1
				}
			}
			fmt.Printf("Infor from public key %v is %+v %v \n", outKey, signatureCounter[outKey], activeTimes)
			//TODO remove hardcode
			if activeTimes >= 10 {
				enoughActivetimes = append(enoughActivetimes, outKey)
				sInfor.SetAutoStaking(false)
			}
			sInfor.SetActiveTimesInCommittee(activeTimes)
			if err := statedb.StoreShardStakerInfoObject(bStateDB, outKey, sInfor); err != nil {
				return nil, err
			}
		}
	}
	if len(enoughActivetimes) > 0 {
		Logger.log.Infof("Process count actives times in committee done, force unstake list keys: %+v", enoughActivetimes)
	}
	return enoughActivetimes, nil
}

func (b *BeaconCommitteeStateV4) addDataToEnvironment(env *BeaconCommitteeStateEnvironment) {
	b.BeaconCommitteeStateV3.addDataToEnvironment(env)
	env.beaconCommittee = common.DeepCopyString(b.beaconCommittee)
	env.beaconSubstitute = common.DeepCopyString(b.beaconSubstitute)
	env.beaconWaiting = common.DeepCopyString(b.beaconWaiting)
	env.waitingStatus = make([]byte, len(b.waitingStatus))
	copy(env.waitingStatus, b.waitingStatus)
}

func (b *BeaconCommitteeStateV4) processExitShardPublicKeys(exitPKs map[byte][]string) {

}
