package committeestate

import (
	"bytes"
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

	beaconStakingAmount map[string]uint64
	delegate            map[string]string

	bDelegateState *BeaconDelegateState
	bLockingState  *BeaconLockingState
	Performance    map[string]uint64
	Reputation     map[string]uint64
}

func NewBeaconCommitteeStateV4() *BeaconCommitteeStateV4 {
	return &BeaconCommitteeStateV4{
		BeaconCommitteeStateV3: NewBeaconCommitteeStateV3(),
		bDelegateState:         &BeaconDelegateState{},
		Performance:            map[string]uint64{},
		Reputation:             map[string]uint64{},
		bLockingState:          NewBeaconLockingState(),
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
	res.bLockingState = NewBeaconLockingState()
	res.InitReputationState()
	if err != nil {
		panic(err)
	}
	return res
}

func (b *BeaconCommitteeStateV4) Version() int {
	return STAKING_FLOW_V4
}

func (b *BeaconCommitteeStateV4) GetConsensusDB() *statedb.StateDB {
	return b.consensusDB
}

func (b *BeaconCommitteeStateV4) GetSlashedDB() *statedb.StateDB {
	return b.slashingDB
}

func (b *BeaconCommitteeStateV4) SetConsensusDB(sDB *statedb.StateDB) {
	b.consensusDB = sDB
}

func (b *BeaconCommitteeStateV4) SetSlashedDB(sDB *statedb.StateDB) {
	b.slashingDB = sDB
}

// shallowCopy maintain dst mutex value
func (b *BeaconCommitteeStateV4) shallowCopy(newB *BeaconCommitteeStateV4) {
	b.BeaconCommitteeStateV3.shallowCopy(newB.BeaconCommitteeStateV3)
	newB.beaconSubstitute = b.beaconSubstitute
	newB.beaconWaiting = b.beaconWaiting
	// newB.beaconLocking = b.beaconLocking
	newB.delegate = b.delegate
	newB.beaconStakingAmount = b.beaconStakingAmount
	newB.bDelegateState = b.bDelegateState
	newB.bLockingState = b.bLockingState
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
	newB.bLockingState = NewBeaconLockingState()
	newB.bLockingState.isChange = b.bLockingState.isChange
	for k, v := range b.bLockingState.Data {
		newB.bLockingState.Data[k] = v
	}
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
	BeaconCommitteeStateV4.consensusDB = env.ConsensusStateDB
	BeaconCommitteeStateV4.slashingDB = env.SlashStateDB
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
	b.consensusDB = env.ConsensusStateDB
	b.slashingDB = env.SlashStateDB
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
	err = b.updateDelegateInfo(env)
	if err != nil {
		Logger.log.Error("[committee-state] err:", err)
		return nil, nil, err
	}
	iInsts := instruction.ValidateAndImportConsensusInstructionFromListString(env.BeaconInstructions)

	for _, iInst := range iInsts {
		returnStakingInstruction, err = b.processInstruction(env, returnStakingInstruction, iInst)
		if err != nil {
			Logger.log.Error("[committee-state] err:", err)
			return nil, nil, err
		}
	}
	if !returnStakingInstruction.IsEmpty() {
		newIInsts = append(newIInsts, returnStakingInstruction)
	}
	newIInsts, err = b.processAtSpecialTime(env, newIInsts)
	if err != nil {
		Logger.log.Error("[committee-state] err:", err)
		return nil, nil, err
	}

	for _, iInst := range newIInsts {
		incurredInstructions = append(incurredInstructions, iInst.ToString())
	}
	hashes, err := b.Hash()
	if err != nil {
		Logger.log.Error("[committee-state] err:", err)
		return hashes, incurredInstructions, err
	}
	return hashes, incurredInstructions, nil
}

func (b *BeaconCommitteeStateV4) CommitOnBlock(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
	consensusDB *statedb.StateDB,
	slashedDB *statedb.StateDB,
) (
	*statedb.StateDB,
	*statedb.StateDB,
	error,
) {
	b.consensusDB = consensusDB
	b.slashingDB = slashedDB

	commitChanges := []func(
		bBlock *types.BeaconBlock,
		env *BeaconCommitteeStateEnvironment,
	) error{
		b.commitStaker,
		b.commitSyncing,
		b.commitSubstitute,
		b.commitCommittee,
		b.commitSlashing,
		b.commitCommitteeInfo,
	}

	for _, commitChange := range commitChanges {
		if err := commitChange(bBlock, env); err != nil {
			return nil, nil, err
		}
	}
	if err := b.Backup(env); err != nil {
		return nil, nil, err
	}

	return b.consensusDB, b.slashingDB, nil
}

func (b *BeaconCommitteeStateV4) commitStaker(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	cDB := b.consensusDB
	stopAutoStakerKeys := b.committeeChange.StopAutoStakeKeys()
	if len(stopAutoStakerKeys) != 0 {
		keys, _ := incognitokey.CommitteeKeyListToString(stopAutoStakerKeys)
		fmt.Printf("Got list stop auto stake %v\n", common.ShortPKList(keys))
		if err := statedb.SaveStopAutoShardStakerInfo(cDB, stopAutoStakerKeys, b.GetAutoStaking()); err != nil {
			return err
		}
		if err := statedb.SaveStopAutoStakeBeaconStaker(cDB, stopAutoStakerKeys, b.GetAutoStaking()); err != nil {
			return err
		}
	}
	if err := b.commitShardStaker(env); err != nil {
		return err
	}
	return b.commitBeaconStaker(env)
}

func (b *BeaconCommitteeStateV4) commitSyncing(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	return b.commitShardSyncing(env)
}

func (b *BeaconCommitteeStateV4) commitSubstitute(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	if err := b.commitShardPending(env); err != nil {
		return err
	}
	return b.commitBeaconPending(env)
}

func (b *BeaconCommitteeStateV4) commitCommittee(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	if err := b.commitShardCommittee(env); err != nil {
		return err
	}
	return b.commitBeaconCommittee(env)
}

func (b *BeaconCommitteeStateV4) commitSlashing(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	cDB := b.consensusDB
	committeeChange := b.committeeChange
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
	err := statedb.StoreSlashingCommittee(b.slashingDB, bBlock.Header.Epoch-1, committeeChange.SlashingCommittee)
	if err != nil {
		return err
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitCommitteeInfo(
	bBlock *types.BeaconBlock,
	env *BeaconCommitteeStateEnvironment,
) error {
	if err := b.commitRedelegate(env); err != nil {
		return err
	}
	if err := b.commitAddStaking(env); err != nil {
		return err
	}
	return b.commitAutoStaking(env)
}

func (b *BeaconCommitteeStateV4) commitShardStaker(
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
	sRemoveKeys := committeeChange.RemovedStakers()
	sRemoveKeysStr, _ := incognitokey.CommitteeKeyListToString(sRemoveKeys)
	bc := append(b.beaconCommittee, b.beaconSubstitute...)
	bc = append(bc, b.beaconWaiting...)
	shardOnlyKeys := common.ExceptString(sRemoveKeysStr, bc)
	for _, k := range shardOnlyKeys {
		pkStr := incognitokey.CommitteePublicKey{}
		if err := pkStr.FromString(k); err != nil {
			return err
		}
		b.deleteStakerInfo(pkStr, k)
	}
	if err := statedb.DeleteStakerInfo(cDB, committeeChange.RemovedStakers()); err != nil {
		return err
	}
	if err := statedb.StoreCurrentEpochShardCandidate(cDB, committeeChange.CurrentEpochShardCandidateAdded); err != nil {
		return err
	}
	if err := statedb.StoreNextEpochShardCandidate(cDB, committeeChange.NextEpochShardCandidateAdded, b.GetRewardReceiver(), b.GetAutoStaking(), b.GetStakingTx()); err != nil {
		return err
	}
	if err := statedb.DeleteCurrentEpochShardCandidate(cDB, committeeChange.CurrentEpochShardCandidateRemoved); err != nil {
		return err
	}
	return statedb.DeleteNextEpochShardCandidate(cDB, committeeChange.NextEpochShardCandidateRemoved)
}

func (b *BeaconCommitteeStateV4) commitShardSyncing(
	env *BeaconCommitteeStateEnvironment,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	if err := statedb.StoreSyncingValidators(bStateDB, committeeChange.SyncingPoolAdded); err != nil {
		return err
	}
	return statedb.DeleteSyncingValidators(bStateDB, committeeChange.SyncingPoolRemoved)
}

func (b *BeaconCommitteeStateV4) commitShardPending(
	env *BeaconCommitteeStateEnvironment,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	//storeAllShardSubstitutesValidatorV3
	allAddedValidators := committeeChange.ShardSubstituteAdded
	for shardID, addedValidators := range allAddedValidators {
		if len(addedValidators) == 0 {
			continue
		}
		substituteValidatorList := b.GetOneShardSubstitute(shardID)
		err := statedb.StoreOneShardSubstitutesValidatorV3(
			bStateDB,
			shardID,
			substituteValidatorList,
		)
		if err != nil {
			return err
		}
	}
	if err := statedb.DeleteAllShardSubstitutesValidator(bStateDB, committeeChange.ShardSubstituteRemoved); err != nil {
		return err
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitShardCommittee(
	env *BeaconCommitteeStateEnvironment,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	listEnableCredit := []incognitokey.CommitteePublicKey{}
	for _, addedCommittee := range committeeChange.ShardCommitteeAdded {
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
	if env.IsEndOfEpoch {
		for _, outKeys := range b.shardCommittee {
			processList := outKeys[config.Param().CommitteeSize.NumberOfFixedShardBlockValidator:]
			if _, err := b.countActiveTimes(
				bStateDB,
				env.MissingSignature,
				b.beaconWaiting,
				processList,
			); err != nil {
				return err
			}

		}
	}
	err := statedb.StoreAllShardCommittee(bStateDB, committeeChange.ShardCommitteeAdded)
	if err != nil {
		return err
	}
	err = statedb.ReplaceAllShardCommittee(bStateDB, committeeChange.ShardCommitteeReplaced)
	if err != nil {
		return err
	}
	err = statedb.DeleteAllShardCommittee(bStateDB, committeeChange.ShardCommitteeRemoved)
	if err != nil {
		return err
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitBeaconStaker(
	env *BeaconCommitteeStateEnvironment,
) error {
	cDB := b.consensusDB
	beaconStakerKeys := b.committeeChange.BeaconStakerKeys()
	cChange := b.committeeChange
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
	listRemoves := b.committeeChange.RemovedBeaconStakers()
	listRemovesStr, _ := incognitokey.CommitteeKeyListToString(listRemoves)
	for idx, pkStruct := range listRemoves {
		b.deleteStakerInfo(pkStruct, listRemovesStr[idx])
	}
	if err := statedb.DeleteBeaconStakerInfo(cDB, listRemoves); err != nil {
		return err
	}
	if err := statedb.StoreCurrentEpochBeaconCandidate(cDB, cChange.CurrentEpochBeaconCandidateAdded); err != nil {
		return err
	}
	if err := statedb.DeleteCurrentEpochBeaconCandidate(cDB, cChange.CurrentEpochBeaconCandidateRemoved); err != nil {
		return err
	}
	for _, newStakerBeacon := range beaconStakerKeys {
		stakerKey, err := newStakerBeacon.ToBase58()
		if err != nil {
			return err
		}
		b.bDelegateState.AddBeaconCandidate(stakerKey, b.beaconStakingAmount[stakerKey])
		b.Reputation[stakerKey] = b.beaconStakingAmount[stakerKey] / 2
		b.Performance[stakerKey] = 500
	}
	return statedb.StoreNextEpochBeaconCandidate(cDB, cChange.NextEpochBeaconCandidateAdded, b.GetRewardReceiver(), b.GetAutoStaking(), b.GetStakingTx())
}

func (b *BeaconCommitteeStateV4) commitBeaconPending(
	env *BeaconCommitteeStateEnvironment,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	//storeAllShardSubstitutesValidatorV3
	if err := statedb.StoreBeaconSubstituteValidator(bStateDB, committeeChange.BeaconSubstituteAdded); err != nil {
		return err
	}
	return statedb.DeleteBeaconSubstituteValidator(bStateDB, committeeChange.BeaconSubstituteRemoved)
}

func (b *BeaconCommitteeStateV4) commitBeaconCommittee(
	env *BeaconCommitteeStateEnvironment,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	err := statedb.StoreBeaconCommittee(bStateDB, committeeChange.BeaconCommitteeAdded)
	if err != nil {
		return err
	}
	err = statedb.ReplaceBeaconCommittee(bStateDB, committeeChange.BeaconCommitteeReplaced)
	if err != nil {
		return err
	}
	err = statedb.DeleteBeaconCommittee(bStateDB, committeeChange.BeaconCommitteeRemoved)
	if err != nil {
		return err
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitAddStaking(
	env *BeaconCommitteeStateEnvironment,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	if len(committeeChange.AddStakingInfo) > 0 {
		for publicKey, addStakeInfo := range committeeChange.AddStakingInfo {
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
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitAutoStaking(
	env *BeaconCommitteeStateEnvironment,
) error {
	bStateDB := b.consensusDB
	committeeChange := b.committeeChange
	stopAutoStakerKeys := committeeChange.StopAutoStakeKeys()
	if len(stopAutoStakerKeys) != 0 {
		err := statedb.SaveStopAutoShardStakerInfo(bStateDB, stopAutoStakerKeys, b.GetAutoStaking())
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BeaconCommitteeStateV4) commitRedelegate(
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

func (b *BeaconCommitteeStateV4) updateDelegateInfo(
	env *BeaconCommitteeStateEnvironment,
) error {
	if env.BeaconHeight > 2 {
		if err := b.UpdateBeaconPerformanceWithValidationData(env.PreviousBlockValidationData); err != nil {
			return err
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

func (b *BeaconCommitteeStateV4) Backup(env *BeaconCommitteeStateEnvironment) error {
	// return nil
	needBackup := false
	cData, err := statedb.GetCommitteeStateBackupData(b.consensusDB)
	if err != nil {
		cData = &statedb.CommitteeData{}
		needBackup = true
	}

	curDData := cData.BeaconDelegateData()
	newDData := b.bDelegateState.Backup()
	if !bytes.Equal(curDData, newDData) {
		needBackup = true
		cData.SetBeaconDelegateData(newDData)
	}
	curLData := cData.BeaconLockingData()
	newLData := b.bLockingState.Backup()
	if !bytes.Equal(curLData, newLData) {
		needBackup = true
		cData.SetBeaconLockingData(newLData)
	}
	if env.IsBeaconChangeTime {
		pData := b.BackupPerformance(env.BeaconHeight)
		cData.SetBeaconPerformanceData(pData)
		needBackup = true
	}
	if needBackup {
		return statedb.StoreCommitteeStateBackupData(b.consensusDB, cData)
	}
	return nil
}

func (b *BeaconCommitteeStateV4) Restore(beaconBlocks []types.BeaconBlock) error {
	curData, err := statedb.GetCommitteeStateBackupData(b.consensusDB)
	if err != nil {
		return err
	}
	if err := b.bDelegateState.Restore(b, b.consensusDB, curData.BeaconDelegateData()); err != nil {
		return err
	}
	b.RestorePerformance(curData.BeaconPerformanceData(), beaconBlocks)

	return b.bLockingState.Restore(curData.BeaconLockingData())
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
		shardNewCandidates: env.newAllShardRoles,
	}
	beaconWaiting := []string{}
	slashedShard := map[string]interface{}{}
	for _, slashedPKs := range committeeChange.SlashingCommittee {
		for _, slashedPK := range slashedPKs {
			slashedShard[slashedPK] = nil
		}
	}
	outPublicKeys := []string{}
	for _, outPublicKey := range env.beaconWaiting {
		if stakerInfo, has, err := statedb.GetBeaconStakerInfo(env.ConsensusStateDB, outPublicKey); (err != nil) || (!has) {
			err = errors.Errorf("Can not found staker info for pk %v at block %v - %v err %v", outPublicKey, env.BeaconHeight, env.BeaconHash, err)
			return nil, nil, err
		} else {
			_, isSlashed := slashedShard[outPublicKey]
			if (stakerInfo.AutoStaking()) && (!isSlashed) {
				beaconWaiting = append(beaconWaiting, outPublicKey)
			} else {
				outPublicKeys = append(outPublicKeys, outPublicKey)
			}
		}
	}
	acceptedCandidate, notAcceptedCandidate, shouldRemove := b.assignRule.ProcessBeacon(beaconWaiting, &envAssign)
	if len(acceptedCandidate) != 0 {
		acceptedCandidateStr, err := incognitokey.CommitteeBase58KeyListToStruct(acceptedCandidate)
		if err != nil {
			return nil, nil, err
		}
		committeeChange.BeaconSubstituteAdded = append(committeeChange.BeaconSubstituteAdded, acceptedCandidateStr...)
		committeeChange.CurrentEpochBeaconCandidateRemoved = append(committeeChange.CurrentEpochBeaconCandidateRemoved, acceptedCandidateStr...)
		committeeChange.AddStopAutoStakes(shouldRemove)
		b.beaconSubstitute = append(b.beaconSubstitute, acceptedCandidate...)
	}
	if len(shouldRemove) > 0 {
		committeeChange.OutOfCommitteeCycle[common.BeaconChainSyncID] = append(committeeChange.OutOfCommitteeCycle[common.BeaconChainSyncID], shouldRemove...)
	}
	if len(outPublicKeys) > 0 {
		committeeChange.OutOfCommitteeCycle[common.BeaconChainSyncID] = append(committeeChange.OutOfCommitteeCycle[common.BeaconChainSyncID], outPublicKeys...)
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
	beaconPending := map[string]interface{}{}
	swapInKeys := []string{}
	addedsubtitute := []string{}
	beaconCommittee := env.beaconCommittee
	beaconSubstitute := []string{}
	if len(env.beaconSubstitute) > 0 {
		fmt.Printf("Before process swap and slash: %+v\n", common.ShortPKList(env.beaconSubstitute))
		for _, outPublicKey := range env.beaconSubstitute {
			if stakerInfo, has, err := statedb.GetBeaconStakerInfo(bCStateDB, outPublicKey); (err != nil) || (!has) {
				err = errors.Errorf("Can not found staker info for pk %v at block %v - %v err %v", outPublicKey, env.BeaconHeight, env.BeaconHash, err)
				return nil, nil, err
			} else {
				fmt.Printf("%v-%v ", outPublicKey[len(outPublicKey)-5:], stakerInfo.AutoStaking())
				if stakerInfo.AutoStaking() {
					beaconSubstitute = append(beaconSubstitute, outPublicKey)
				} else {
					outPublicKeys = append(outPublicKeys, outPublicKey)
				}
			}
		}
		fmt.Printf("\nAfter process swap and slash: got keys out %+v, beaconSubtitute %+v\n", common.ShortPKList(outPublicKeys), common.ShortPKList(beaconSubstitute))
	}
	if len(outPublicKeys) > 0 {
		if outKeysStr, err := incognitokey.CommitteeKeyListToStruct(outPublicKeys); err == nil {
			committeeChange.BeaconCommitteeRemoved = append(committeeChange.BeaconCommitteeRemoved, outKeysStr...)
		}
	}
	for _, v := range beaconSubstitute {
		beaconPending[v] = nil
	}
	newCommittee, newSubtitute, swapOutKeys, slashedCommittee := b.swapRule.ProcessBeacon(
		beaconCommittee,
		beaconSubstitute,
		env.MinBeaconCommitteeSize,
		env.MaxBeaconCommitteeSize,
		env.NumberOfFixedBeaconBlockValidator,
		b.Reputation,
		b.Performance,
	)
	for _, outPublicKey := range swapOutKeys {
		if stakerInfo, has, err := statedb.GetBeaconStakerInfo(bCStateDB, outPublicKey); (err != nil) || (!has) {
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
		if _, ok := beaconPending[pk]; ok {
			swapInKeys = append(swapInKeys, pk)
		}
	}
	if len(swapInKeys) > 0 {
		if swapInKeysStr, err := incognitokey.CommitteeKeyListToStruct(swapInKeys); err == nil {
			committeeChange.BeaconCommitteeAdded = append(committeeChange.BeaconCommitteeAdded, swapInKeysStr...)
			committeeChange.BeaconSubstituteRemoved = append(committeeChange.BeaconSubstituteRemoved, swapInKeysStr...)
		}
	}
	if len(swapOutKeys) > 0 {
		if swapOutKeysStr, err := incognitokey.CommitteeKeyListToStruct(swapOutKeys); err == nil {
			committeeChange.BeaconCommitteeRemoved = append(committeeChange.BeaconCommitteeRemoved, swapOutKeysStr...)
		}
	}
	if len(addedsubtitute) > 0 {
		if addedSubstituteStr, err := incognitokey.CommitteeKeyListToStruct(addedsubtitute); err == nil {
			committeeChange.BeaconSubstituteAdded = append(committeeChange.BeaconSubstituteAdded, addedSubstituteStr...)
		}
	}
	if len(slashedCommittee) > 0 {
		committeeChange.SlashingCommittee[common.BeaconChainSyncID] = append(committeeChange.SlashingCommittee[common.BeaconChainSyncID], slashedCommittee...)
	}
	if len(outPublicKeys) > 0 {
		committeeChange.OutOfCommitteeCycle[common.BeaconChainSyncID] = append(committeeChange.OutOfCommitteeCycle[common.BeaconChainSyncID], outPublicKeys...)
	}
	b.beaconCommittee = newCommittee
	b.beaconSubstitute = newSubtitute
	b.committeeChange = committeeChange
	return outPublicKeys, slashedCommittee, nil
}

func (b BeaconCommitteeStateV4) GetBeaconWaiting() []incognitokey.CommitteePublicKey {
	bPKStructs, err := incognitokey.CommitteeKeyListToStruct(b.beaconWaiting)
	if err != nil {
		panic(err)
	}
	return bPKStructs
}

func (b BeaconCommitteeStateV4) GetBeaconLocking() []incognitokey.CommitteePublicKey {
	lockingKeysStr := b.bLockingState.GetAllLockingPK()
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

func (b *BeaconCommitteeStateV4) SetBeaconSubstitute(subs []incognitokey.CommitteePublicKey) {
	bPKString, err := incognitokey.CommitteeKeyListToString(subs)
	if err != nil {
		panic(err)
	}
	b.beaconSubstitute = bPKString
}

func (b *BeaconCommitteeStateV4) SetBeaconWaiting(waiting []incognitokey.CommitteePublicKey) {
	bPKString, err := incognitokey.CommitteeKeyListToString(waiting)
	if err != nil {
		panic(err)
	}
	b.beaconWaiting = bPKString
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
	processList := normalSwapOutCommittees
	processList = common.IntersectionString(env.beaconWaiting, processList)
	enoughActiveTimes := []string{}
	for _, outKey := range processList {
		if bInfor, has, err := statedb.GetBeaconStakerInfo(b.consensusDB, outKey); (!has) || (err != nil) {
			err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", outKey, has, err)
			Logger.log.Error(err)
			continue
		} else {
			activeTimes := bInfor.ActiveTimesInCommittee()
			if activeTimes >= uint(config.Param().ConsensusParam.RequiredActiveTimes) {
				enoughActiveTimes = append(enoughActiveTimes, outKey)
				sInfor, has, err := statedb.GetShardStakerInfo(b.consensusDB, outKey)
				if (!has) || (err != nil) {
					Logger.log.Error(err)
					continue
				}
				sInfor.SetAutoStaking(false)
				if err := statedb.StoreShardStakerInfoObject(b.consensusDB, outKey, sInfor); err != nil {
					return nil, err
				}
			}
		}
	}
	if len(enoughActiveTimes) > 0 {
		Logger.log.Infof("Process count actives times in committee done, force unstake list keys: %+v", common.ShortPKList(enoughActiveTimes))
	}
	//TODO refactor
	enoughActivesTimesM := map[string]interface{}{}
	for _, key := range enoughActiveTimes {
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

	return returnStakingInstruction, nil
}

func (b *BeaconCommitteeStateV4) processBeaconLocking(
	env *BeaconCommitteeStateEnvironment,
) (
	*instruction.ReturnBeaconStakeInstruction,
	error,
) {
	if len(b.committeeChange.SlashingCommittee[common.BeaconChainSyncID]) > 0 {
		slashedPKs := b.committeeChange.SlashingCommittee[common.BeaconChainSyncID]
		for _, slashedPK := range slashedPKs {
			totalEpoch := (100 * (2000 - b.Performance[slashedPK])) / 1000
			fmt.Printf("Locked infor: Lock candidate %v by slash, total epoch %v, to epoch %v, their performance %v\n", slashedPK[len(slashedPK)-5:], totalEpoch, env.Epoch+totalEpoch, b.Performance[slashedPK])
			totalEpoch = 5
			b.LockNewCandidate(slashedPK, env.Epoch+totalEpoch, common.RETURN_BY_SLASHED)
		}
	}
	if len(b.committeeChange.OutOfCommitteeCycle[common.BeaconChainSyncID]) > 0 {
		outPKs := b.committeeChange.OutOfCommitteeCycle[common.BeaconChainSyncID]
		for _, outPK := range outPKs {
			totalEpoch := (100 * (2000 - b.Performance[outPK])) / 1000
			fmt.Printf("Locked infor: Lock candidate %v by unstake, total epoch %v, to epoch %v, their performance %v\n", outPK[len(outPK)-5:], totalEpoch, env.Epoch+totalEpoch, b.Performance[outPK])
			totalEpoch = 5
			b.LockNewCandidate(outPK, env.Epoch+totalEpoch, common.RETURN_BY_UNSTAKE)
		}
	}
	returnLocking := b.GetReturnStakingInstruction(env.ConsensusStateDB, env.Epoch)
	return returnLocking, nil
}

func (b *BeaconCommitteeStateV4) countActiveTimes(
	bStateDB *statedb.StateDB,
	signatureCounter map[string]signaturecounter.MissingSignature,
	waiting []string,
	processList []string,
) (
	[]string,
	error,
) {
	enoughActivetimes := []string{}
	processList = common.IntersectionString(waiting, processList)
	for _, outKey := range processList {
		if bInfor, has, err := statedb.GetBeaconStakerInfo(bStateDB, outKey); (!has) || (err != nil) {
			err = errors.Errorf("Can not get staker infor, found %v in db: %v, got err %v", outKey, has, err)
			return nil, err
		} else {
			activeTimes := uint(0)
			if voteInfo, has := signatureCounter[outKey]; has {
				votePercent := 100 - voteInfo.Missing*100/voteInfo.ActualTotal
				if votePercent >= 90 {
					activeTimes = bInfor.ActiveTimesInCommittee() + 1
				}
			}
			if activeTimes >= uint(config.Param().ConsensusParam.RequiredActiveTimes) {
				enoughActivetimes = append(enoughActivetimes, outKey)
			}
			bInfor.SetActiveTimesInCommittee(activeTimes)
			if err := statedb.StoreBeaconStakerInfoObject(bStateDB, outKey, bInfor); err != nil {
				return nil, err
			}
		}
	}
	if len(enoughActivetimes) > 0 {
		Logger.log.Infof("Process count actives times in committee done, force unstake list keys: %+v", common.ShortPKList(enoughActivetimes))
	}
	return enoughActivetimes, nil
}

func (b *BeaconCommitteeStateV4) addDataToEnvironment(env *BeaconCommitteeStateEnvironment) {
	b.BeaconCommitteeStateV3.addDataToEnvironment(env)
	env.beaconCommittee = common.DeepCopyString(b.beaconCommittee)
	env.beaconSubstitute = common.DeepCopyString(b.beaconSubstitute)
	env.beaconWaiting = common.DeepCopyString(b.beaconWaiting)
	env.newAllRoles = append(env.newAllRoles, env.beaconWaiting...)
	env.newAllRoles = append(env.newAllRoles, env.beaconSubstitute...)
	env.newAllSubstituteCommittees = append(env.newAllSubstituteCommittees, env.beaconWaiting...)
	env.newAllSubstituteCommittees = append(env.newAllSubstituteCommittees, env.beaconSubstitute...)
}

func (b *BeaconCommitteeStateV4) GetAllCandidateSubstituteCommittee() []string {
	res := b.BeaconCommitteeStateV3.GetAllCandidateSubstituteCommittee()
	res = append(res, b.beaconSubstitute...)
	res = append(res, b.beaconWaiting...)
	return res
}
