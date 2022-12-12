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
	consensusDB *statedb.StateDB
	slashingDB  *statedb.StateDB
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

	committeeChange *CommitteeChange
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

func (b BeaconCommitteeStateV4) Hash() (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	hashes, err := b.BeaconCommitteeStateV3.Hash()
	if err != nil {
		return nil, err
	}
	committeeChange := b.committeeChange

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

func (b *BeaconCommitteeStateV4) UpdateCommitteeState(
	env *BeaconCommitteeStateEnvironment,
) (
	*BeaconCommitteeStateHash,
	[][]string,
	error,
) {
	var err error
	incurredInstructions := [][]string{}
	newIInsts := []instruction.Instruction{}
	returnStakingInstruction := instruction.NewReturnStakeIns()
	b.committeeChange = NewCommitteeChange()
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
	iInsts := instruction.ValidateAndImportConsensusInstructionFromListString(env.BeaconInstructions)

	for _, iInst := range iInsts {
		returnStakingInstruction, err = b.processInstruction(env, returnStakingInstruction, iInst)
		if err != nil {
			return nil, nil, err
		}
	}
	newIInsts = append(newIInsts, returnStakingInstruction)
	newIInsts, err = b.processAtSpecialTime(env, newIInsts)
	if err != nil {
		return nil, nil, err
	}

	for _, iInst := range newIInsts {
		incurredInstructions = append(incurredInstructions, iInst.ToString())
	}
	hashes, err := b.Hash()
	if err != nil {
		return hashes, incurredInstructions, err
	}
	return hashes, incurredInstructions, nil
}

func (b *BeaconCommitteeStateV4) CommitOnBlock(
	bBlock *types.BeaconBlock,
	newCommitteeChange *CommitteeChange,
	inst [][]string,
) error {
	env := &BeaconCommitteeStateEnvironment{
		BeaconHeight: bBlock.GetBeaconHeight(),
	}
	commitChanges := []func(bBlock *types.BeaconBlock, env *BeaconCommitteeStateEnvironment) error{
		b.commitUpdateStaker,
		b.commitRedelegate,
	}
	for _, commitChange := range commitChanges {
		if err := commitChange(bBlock, env); err != nil {
			return err
		}
	}
	// sDB := b.slashingDB
	return nil
}

func (b *BeaconCommitteeStateV4) commitUpdateStaker(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	cDB := b.consensusDB
	stopAutoStakerKeys := b.committeeChange.StopAutoStakeKeys()
	if len(stopAutoStakerKeys) != 0 {
		if err := statedb.SaveStopAutoStakerInfo(cDB, stopAutoStakerKeys, b.GetAutoStaking()); err != nil {
			return err
		}
	}
	if err := b.commitUpdateShardStaker(env); err != nil {
		return err
	}
	if err := b.commitUpdateBeaconStaker(bBlock, env); err != nil {
		return err
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitUpdateShardStaker(
	env *BeaconCommitteeStateEnvironment,
) error {
	cDB := b.consensusDB
	committeeChange := b.committeeChange
	shardStakerKeys := committeeChange.ShardStakerKeys()
	if len(shardStakerKeys) != 0 {
		err := statedb.StoreShardStakerInfo(
			cDB,
			shardStakerKeys,
			b.GetRewardReceiver(),
			b.GetAutoStaking(),
			b.GetStakingTx(),
			env.BeaconHeight,
			b.GetDelegate(),
			map[string]interface{}{},
		)
		if err != nil {
			return err
		}
	}
	for _, removedCommittee := range committeeChange.SlashingCommittee {
		for _, stakerPubkeyStr := range removedCommittee {
			oldD := ""
			newD := ""
			stakerInfo, exist, err := statedb.GetShardStakerInfo(cDB, stakerPubkeyStr)
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
	return statedb.StoreSyncingValidators(cDB, committeeChange.SyncingPoolAdded)
}

func (b *BeaconCommitteeStateV4) commitUpdateBeaconStaker(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	cDB := b.consensusDB
	beaconStakerKeys := b.committeeChange.BeaconStakerKeys()
	if len(beaconStakerKeys) != 0 {
		err := statedb.StoreBeaconStakersInfo(
			cDB,
			beaconStakerKeys,
			b.GetRewardReceiver(),
			b.committeeChange.FunderAddress,
			b.GetAutoStaking(),
			b.GetStakingTx(),
			env.BeaconHeight,
			b.GetBCStakingAmount(),
		)
		if err != nil {
			return err
		}
	}
	if env.BeaconHeight > 2 {
		if err := b.UpdateBeaconReputationWithBlock(bBlock); err != nil {
			return err
		}
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitRedelegate(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	if len(committeeChange.ReDelegate) != 0 {
		for stakerPubkeyStr, newDelegate := range committeeChange.ReDelegate {
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
		if err := statedb.SaveStakerReDelegateInfo(bStateDB, committeeChange.ReDelegate); err != nil {
			return err
		}
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitRedelegate2(
	env *BeaconCommitteeStateEnvironment,
	// committeeChange *CommitteeChange,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	if len(committeeChange.ReDelegate) != 0 {
		for stakerPubkeyStr, newDelegate := range committeeChange.ReDelegate {
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
		if err := statedb.SaveStakerReDelegateInfo(bStateDB, committeeChange.ReDelegate); err != nil {
			return err
		}
	}
	return nil
}

func (b *BeaconCommitteeStateV4) Backup() error {
	return nil
}

func (b *BeaconCommitteeStateV4) Restore() error {
	return nil
}

func (b *BeaconCommitteeStateV4) processInstruction(
	env *BeaconCommitteeStateEnvironment,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	iInst instruction.Instruction,
) (
	*instruction.ReturnStakeInstruction,
	error,
) {
	var err error = nil
	switch iInst.GetType() {
	case instruction.STAKE_ACTION:
		stakeInstruction := iInst.(*instruction.StakeInstruction)
		err = b.processStakeInstruction(stakeInstruction)
	case instruction.RANDOM_ACTION:
		randomInstruction := iInst.(*instruction.RandomInstruction)
		b.processAssignWithRandomInstruction(randomInstruction.RandomNumber(), env.numberOfValidator)
	case instruction.STOP_AUTO_STAKE_ACTION:
		stopAutoStakeInstruction := iInst.(*instruction.StopAutoStakeInstruction)
		b.processStopAutoStakeInstruction(stopAutoStakeInstruction, env)
	case instruction.SWAP_SHARD_ACTION:
		swapShardInstruction := iInst.(*instruction.SwapShardInstruction)
		returnStakingInstruction, err = b.processSwapShardInstruction(swapShardInstruction, env, returnStakingInstruction)
	case instruction.FINISH_SYNC_ACTION:
		finishSyncInstruction := iInst.(*instruction.FinishSyncInstruction)
		b.processFinishSyncInstruction(finishSyncInstruction, env)
	case instruction.UNSTAKE_ACTION:
		unstakeInstruction := iInst.(*instruction.UnstakeInstruction)
		returnStakingInstruction, err = b.processUnstakeInstruction(unstakeInstruction, env, returnStakingInstruction)
	case instruction.RE_DELEGATE:
		redelegateInstruction := iInst.(*instruction.ReDelegateInstruction)
		b.processReDelegateInstruction(redelegateInstruction, env)
	case instruction.ADD_STAKING_ACTION:
		addStakingInstruction := iInst.(*instruction.AddStakingInstruction)
		b.processAddStakingInstruction(addStakingInstruction, env)
	}
	return returnStakingInstruction, err
}

func (b *BeaconCommitteeStateV4) processAtSpecialTime(
	env *BeaconCommitteeStateEnvironment,
	iInsts []instruction.Instruction,
) (
	[]instruction.Instruction,
	error,
) {
	if env.IsBeaconChangeTime {
		var (
			newSubs, newWaiting []string
			err                 error
		)
		newSubs, newWaiting, err = b.processAssignBeacon(env)
		if err != nil {
			Logger.log.Errorf("Process assign beacon got err %v", err)
			return nil, err
		} else {
			if len(newSubs) == len(env.beaconSubstitute) {
				Logger.log.Debugf("Process assign beacon done, list pending and waiting is not changed")
			} else {
				Logger.log.Debugf("Process assign beacon done, list pending %+v and waiting %+v", newSubs, newWaiting)
			}
			env.beaconSubstitute = b.beaconSubstitute
			env.beaconWaiting = b.beaconWaiting
		}
		_, _, err = b.processSwapAndSlashBeacon(env)
		if err != nil {
			Logger.log.Errorf("Process swap and slash beacon got err %v", err)
			return nil, err
		}
		returnBeaconInst, err := b.processBeaconLocking(env)
		if err != nil {
			Logger.log.Errorf("Process return stakng beacon from beacon locking got err %v", err)
			return nil, err
		}
		if returnBeaconInst != nil {
			iInsts = append(iInsts, returnBeaconInst)
		}
	}
	return iInsts, nil
}

func (b *BeaconCommitteeStateV4) processReDelegateInstruction(
	redelegateInstruction *instruction.ReDelegateInstruction,
	env *BeaconCommitteeStateEnvironment,
) {
	changeMap := map[string]string{}
	for index, committeePublicKey := range redelegateInstruction.CommitteePublicKeys {
		b.delegate[committeePublicKey] = redelegateInstruction.DelegateList[index]
		changeMap[committeePublicKey] = redelegateInstruction.DelegateList[index]
	}
	b.committeeChange.AddReDelegateInfo(changeMap)
}

func (b *BeaconCommitteeStateV4) processAddStakingInstruction(
	addStakingInstruction *instruction.AddStakingInstruction,
	env *BeaconCommitteeStateEnvironment,
	// committeeChange *CommitteeChange,
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
	b.committeeChange.AddAddStakingInfo(changeMap)
}

func (b *BeaconCommitteeStateV4) processAssignBeacon(
	env *BeaconCommitteeStateEnvironment,
) (
	newSubtitute []string,
	newWaiting []string,
	err error,
) {
	committeeChange := b.committeeChange
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
			return nil, nil, err
		}
		committeeChange.BeaconSubstituteAdded = append(committeeChange.BeaconSubstituteAdded, acceptedCandidateStr...)
		committeeChange.CurrentEpochBeaconCandidateRemoved = append(committeeChange.CurrentEpochBeaconCandidateRemoved, acceptedCandidateStr...)
		b.beaconSubstitute = append(b.beaconSubstitute, acceptedCandidate...)
		b.waitingStatus = waitingStatus
	}
	b.beaconWaiting = notAcceptedCandidate
	b.committeeChange = committeeChange
	return b.beaconSubstitute, b.beaconWaiting, nil
}

func (b *BeaconCommitteeStateV4) processSwapAndSlashBeacon(
	env *BeaconCommitteeStateEnvironment,
) (
	outPublicKeys []string,
	slashedCommittee []string,
	err error,
) {
	bCStateDB := env.ConsensusStateDB
	committeeChange := b.committeeChange
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
			return nil, nil, err
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
	b.committeeChange = committeeChange
	return outPublicKeys, slashedCommittee, nil
}

// TODO refactor this code
func (b *BeaconCommitteeStateV4) ProcessStoreCommitteeStateInfo(
	bBlock *types.BeaconBlock,
	signatureCounter map[string]signaturecounter.MissingSignature,
	x *statedb.StateDB,
	isEndOfEpoch bool,
) error {
	bStateDB := b.consensusDB
	cChange := b.committeeChange
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
	//-
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
			b.countActiveTimes(
				bStateDB,
				signatureCounter,
				b.beaconWaiting,
				processList,
			)
			// for _, outKey := range processList {
			// 	if sInfor, has, err := statedb.GetBeaconStakerInfo(bStateDB, outKey); (!has) || (err != nil) {
			// 		err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", outKey, has, err)
			// 		return err
			// 	} else {
			// 		activeTimes := 0
			// 		if votePercent, has := signatureCounter[outKey]; has {
			// 			if votePercent.VotePercent >= 90 {
			// 				activeTimes = sInfor.ActiveTimesInCommittee() + 1
			// 			}
			// 		}
			// 		sInfor.SetActiveTimesInCommittee(activeTimes)
			// 		if err := statedb.StoreShardStakerInfoObject(bStateDB, outKey, sInfor); err != nil {
			// 			return err
			// 		}
			// 	}
			// }
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
	// for _, outKeys := range cChange.SwapoutAndBackToPending {
	// 	for _, outKey := range outKeys {
	// 		if sInfor, has, err := statedb.GetShardStakerInfo(bStateDB, outKey); (!has) || (err != nil) {
	// 			err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", outKey, has, err)
	// 			return err
	// 		} else {
	// 			activeTimes := 0
	// 			if votePercent, has := signatureCounter[outKey]; has {
	// 				if votePercent.VotePercent >= 90 {
	// 					activeTimes = sInfor.ActiveTimesInCommittee() + 1
	// 				}
	// 			}
	// 			sInfor.SetActiveTimesInCommittee(activeTimes)
	// 			if err := statedb.StoreShardStakerInfoObject(bStateDB, outKey, sInfor); err != nil {
	// 				return err
	// 			}
	// 		}
	// 	}
	// }
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

func (b BeaconCommitteeStateV4) GetBeaconLocking() []incognitokey.CommitteePublicKey {
	lockingKeysStr := []string{}
	for _, lockingInfo := range b.beaconLocking {
		lockingKeysStr = append(lockingKeysStr, lockingInfo.PublicKey)
	}
	bPKStructs, err := incognitokey.CommitteeKeyListToStruct(lockingKeysStr)
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
) error {
	err := b.BeaconCommitteeStateV3.processStakeInstruction(stakeInstruction)
	if err != nil {
		return err
	}
	if stakeInstruction.Chain == instruction.BEACON_INST {
		b.committeeChange.CurrentEpochBeaconCandidateAdded = append(b.committeeChange.CurrentEpochBeaconCandidateAdded, stakeInstruction.PublicKeyStructs...)
	}
	for index, committeePublicKey := range stakeInstruction.PublicKeys {
		if stakeInstruction.Chain == instruction.SHARD_INST {
			b.delegate[committeePublicKey] = stakeInstruction.DelegateList[index]
		} else {
			b.beaconStakingAmount[committeePublicKey] = stakeInstruction.StakingAmount[index]
		}
	}
	newWaitingStr, err := incognitokey.CommitteeKeyListToString(b.committeeChange.CurrentEpochBeaconCandidateAdded)
	if err != nil {
		return err
	}
	newWaitingStatus := []byte{}
	for range newWaitingStr {
		newWaitingStatus = append(newWaitingStatus, 0)
	}
	b.waitingStatus = append(b.waitingStatus, newWaitingStatus...)
	b.beaconWaiting = append(b.beaconWaiting, newWaitingStr...)
	return err
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
	env *BeaconCommitteeStateEnvironment,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) (
	*instruction.ReturnStakeInstruction,
	error,
) {
	shardID := byte(swapShardInstruction.ChainID)
	env.ShardID = shardID
	// process normal swap out
	_, normalSwapOutCommittees, slashingCommittees, err := b.processSwap(swapShardInstruction, env)
	if err != nil {
		return nil, err
	}
	enoughActivesTimes, err := b.countActiveTimes(
		env.ConsensusStateDB,
		env.MissingSignature,
		env.beaconWaiting,
		normalSwapOutCommittees,
	)
	if err != nil {
		return nil, err
	}
	//TODO refactor
	enoughActivesTimesM := map[string]interface{}{}
	for _, key := range enoughActivesTimes {
		enoughActivesTimesM[key] = nil
	}

	// process after swap for assign old committees to current shard pool
	_, returnStakingInstruction, err = b.processAfterNormalSwap(env,
		normalSwapOutCommittees,
		returnStakingInstruction,
	)
	if err != nil {
		return returnStakingInstruction, err
	}

	for i, returnStakingReason := range returnStakingInstruction.Reasons {
		if _, ok := enoughActivesTimesM[returnStakingInstruction.PublicKeys[i]]; ok {
			returnStakingInstruction.Reasons[i] = common.RETURN_BY_PROMOTE
		} else if returnStakingReason == common.RETURN_BY_UNKNOWN {
			returnStakingInstruction.Reasons[i] = common.RETURN_BY_UNSTAKE
		}

	}
	// process slashing after normal swap out
	returnStakingInstruction, err = b.processSlashing(
		shardID,
		env,
		slashingCommittees,
		returnStakingInstruction,
	)
	for i, returnStakingReason := range returnStakingInstruction.Reasons {
		if returnStakingReason == common.RETURN_BY_UNKNOWN {
			returnStakingInstruction.Reasons[i] = common.RETURN_BY_SLASHED
		}
	}

	if err != nil {
		return returnStakingInstruction, err
	}
	// tmp := map[string]interface{}{}
	// if (len(candidateFromCommittee) > 0) && (len(normalSwapOutCommittees) > 0) {
	// 	for _, v := range normalSwapOutCommittees {
	// 		tmp[v] = nil
	// 	}
	// 	fmt.Printf("Overlap: ")
	// 	for _, v := range candidateFromCommittee {
	// 		if _, ok := tmp[v]; ok {
	// 			fmt.Printf("[%v] ", v[len(v)-5:])
	// 		} else {
	// 			fmt.Printf("{%v} ", v[len(v)-5:])
	// 		}
	// 	}
	// 	fmt.Println()
	// }
	// newCommitteeChange.SwapoutAndBackToPending[shardID] = candidateFromCommittee

	return returnStakingInstruction, nil
}

func (b *BeaconCommitteeStateV4) processBeaconLocking(
	env *BeaconCommitteeStateEnvironment,
) (
	*instruction.ReturnBeaconStakeInstruction,
	error,
) {
	if len(b.committeeChange.SlashingCommittee[common.BeaconChainSyncID]) > 0 {
		slashedPK := b.committeeChange.SlashingCommittee[common.BeaconChainSyncID]
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
	bStateDB *statedb.StateDB,
	signatureCounter map[string]signaturecounter.MissingSignature,
	waiting []string,
	// env *BeaconCommitteeStateEnvironment,
	processList []string,
) (
	[]string,
	error,
) {
	if (len(processList) > 0) && (len(waiting) > 0) {
		fmt.Printf("Start count active times for waiting %+v processList %+v\n", common.ShortPKList(waiting), common.ShortPKList(processList))
	} else {
		if len(processList) > 0 {
			fmt.Printf("Start count active times for waiting %+v processList %+v\n", common.ShortPKList(waiting), common.ShortPKList(processList))
		}
	}
	enoughActivetimes := []string{}
	processList = common.IntersectionString(waiting, processList)
	for _, outKey := range processList {
		if sInfor, has, err := statedb.GetBeaconStakerInfo(bStateDB, outKey); (!has) || (err != nil) {
			err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", outKey, has, err)
			return nil, err
		} else {
			activeTimes := uint(0)
			if voteInfo, has := signatureCounter[outKey]; has {
				votePercent := 100 - voteInfo.Missing*100/voteInfo.ActualTotal
				if votePercent >= 90 {
					activeTimes = sInfor.ActiveTimesInCommittee() + 1
				}
			}
			//TODO remove hardcode
			if activeTimes >= 10 {
				enoughActivetimes = append(enoughActivetimes, outKey)
				sInfor.SetAutoStaking(false)
			}
			sInfor.SetActiveTimesInCommittee(activeTimes)
			if err := statedb.StoreBeaconStakerInfoObject(bStateDB, outKey, sInfor); err != nil {
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
