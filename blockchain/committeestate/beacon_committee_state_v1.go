package committeestate

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type BeaconCommitteeStateV1 struct {
	beaconCommitteeStateBase
	currentEpochShardCandidate []incognitokey.CommitteePublicKey
	nextEpochShardCandidate    []incognitokey.CommitteePublicKey
}

func NewBeaconCommitteeStateEnvironment() *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{}
}

func NewBeaconCommitteeStateV1() *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{
		beaconCommitteeStateBase: *NewBeaconCommitteeStateBase(),
	}
}

func NewBeaconCommitteeStateV1WithValue(
	beaconCurrentValidator []incognitokey.CommitteePublicKey,
	beaconSubstituteValidator []incognitokey.CommitteePublicKey,
	nextEpochShardCandidate []incognitokey.CommitteePublicKey,
	currentEpochShardCandidate []incognitokey.CommitteePublicKey,
	nextEpochBeaconCandidate []incognitokey.CommitteePublicKey,
	currentEpochBeaconCandidate []incognitokey.CommitteePublicKey,
	shardCurrentValidator map[byte][]incognitokey.CommitteePublicKey,
	shardSubstituteValidator map[byte][]incognitokey.CommitteePublicKey,
	autoStaking map[string]bool,
	rewardReceivers map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
) *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{
		beaconCommitteeStateBase: *NewBeaconCommitteeStateBaseWithValue(
			beaconCurrentValidator, shardCurrentValidator, shardSubstituteValidator,
			autoStaking, rewardReceivers, stakingTx,
		),
		nextEpochShardCandidate:    nextEpochShardCandidate,
		currentEpochShardCandidate: currentEpochShardCandidate,
	}
}

func (b *BeaconCommitteeStateV1) Version() int {
	return SELF_SWAP_SHARD_VERSION
}

func (b *BeaconCommitteeStateV1) Reset() {
	b.reset()
}

func (b *BeaconCommitteeStateV1) reset() {
	b.beaconCommitteeStateBase.reset()
	b.currentEpochShardCandidate = []incognitokey.CommitteePublicKey{}
	b.nextEpochShardCandidate = []incognitokey.CommitteePublicKey{}
}

func (b *BeaconCommitteeStateV1) cloneFrom(fromB BeaconCommitteeStateV1) {
	b.reset()
	b.beaconCommitteeStateBase.cloneFrom(fromB.beaconCommitteeStateBase)
	b.currentEpochShardCandidate = make([]incognitokey.CommitteePublicKey, len(fromB.currentEpochShardCandidate))
	copy(b.currentEpochShardCandidate, fromB.currentEpochShardCandidate)
	b.nextEpochShardCandidate = make([]incognitokey.CommitteePublicKey, len(fromB.nextEpochShardCandidate))
	copy(b.nextEpochShardCandidate, fromB.nextEpochShardCandidate)
}

func (b *BeaconCommitteeStateV1) clone() *BeaconCommitteeStateV1 {
	newB := NewBeaconCommitteeStateV1()
	newB.beaconCommitteeStateBase = *b.beaconCommitteeStateBase.clone()
	newB.currentEpochShardCandidate = make([]incognitokey.CommitteePublicKey, len(b.currentEpochShardCandidate))
	copy(newB.currentEpochShardCandidate, b.currentEpochShardCandidate)
	newB.nextEpochShardCandidate = make([]incognitokey.CommitteePublicKey, len(b.nextEpochShardCandidate))
	copy(newB.nextEpochShardCandidate, b.nextEpochShardCandidate)
	return newB
}

func (b *BeaconCommitteeStateV1) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	env *BeaconCommitteeStateEnvironment,
) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	committeeChange := NewCommitteeChange()
	committeeChange, err := b.beaconCommitteeStateBase.processStakeInstruction(stakeInstruction, committeeChange)
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, err
	}
	newShardCandidates := make([]incognitokey.CommitteePublicKey, len(committeeChange.NextEpochShardCandidateAdded))
	copy(newShardCandidates, committeeChange.NextEpochShardCandidateAdded)

	err = statedb.StoreStakerInfo(
		env.ConsensusStateDB,
		stakeInstruction.PublicKeyStructs,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
	)
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, newShardCandidates, err
	}

	return []incognitokey.CommitteePublicKey{}, newShardCandidates, nil
}

func (b *BeaconCommitteeStateV1) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) {
	for _, committeePublicKey := range stopAutoStakeInstruction.CommitteePublicKeys {
		if common.IndexOfStr(committeePublicKey, b.getAllCandidateSubstituteCommittee()) == -1 {
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

func (b *BeaconCommitteeStateV1) processSwapInstruction(
	swapInstruction *instruction.SwapInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	if common.IndexOfUint64(env.Epoch, env.EpochBreakPointSwapNewKey) > -1 || swapInstruction.IsReplace {
		err := b.processReplaceInstruction(swapInstruction, committeeChange, env)
		if err != nil {
			return newBeaconCandidates, newShardCandidates, err
		}
	} else {
		Logger.log.Debug("Swap Instruction In Public Keys", swapInstruction.InPublicKeys)
		Logger.log.Debug("Swap Instruction Out Public Keys", swapInstruction.OutPublicKeys)
		if swapInstruction.ChainID != instruction.BEACON_CHAIN_ID {
			shardID := byte(swapInstruction.ChainID)
			// delete in public key out of sharding pending validator list
			if len(swapInstruction.InPublicKeys) > 0 {
				shardSubstituteStr, err := incognitokey.CommitteeKeyListToString(b.shardSubstitute[shardID])
				if err != nil {
					return newBeaconCandidates, newShardCandidates, err
				}
				tempShardSubstitute, err := removeValidatorV1(shardSubstituteStr, swapInstruction.InPublicKeys)
				if err != nil {
					return newBeaconCandidates, newShardCandidates, err
				}
				// update shard pending validator
				committeeChange.ShardSubstituteRemoved[shardID] = append(committeeChange.ShardSubstituteRemoved[shardID], swapInstruction.InPublicKeyStructs...)
				b.shardSubstitute[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardSubstitute)
				if err != nil {
					return newBeaconCandidates, newShardCandidates, err
				}
				// add new public key to committees
				committeeChange.ShardCommitteeAdded[shardID] = append(committeeChange.ShardCommitteeAdded[shardID], swapInstruction.InPublicKeyStructs...)
				b.shardCommittee[shardID] = append(b.shardCommittee[shardID], swapInstruction.InPublicKeyStructs...)
			}
			// delete out public key out of current committees
			if len(swapInstruction.OutPublicKeys) > 0 {
				//for _, value := range outPublickeyStructs {
				//	delete(b,cue.GetIncKeyBase58(
				//}
				shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(b.shardCommittee[shardID])
				if err != nil {
					return newBeaconCandidates, newShardCandidates, err
				}
				tempShardCommittees, err := removeValidatorV1(shardCommitteeStr, swapInstruction.OutPublicKeys)
				if err != nil {
					return newBeaconCandidates, newShardCandidates, err
				}
				// remove old public key in shard committee update shard committee
				committeeChange.ShardCommitteeRemoved[shardID] = append(committeeChange.ShardCommitteeRemoved[shardID], swapInstruction.OutPublicKeyStructs...)
				b.shardCommittee[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardCommittees)
				if err != nil {
					return newBeaconCandidates, newShardCandidates, err
				}
				// Check auto stake in out public keys list
				// if auto staking not found or flag auto stake is false then do not re-stake for this out public key
				// if auto staking flag is true then system will automatically add this out public key to current candidate list
				for index, outPublicKey := range swapInstruction.OutPublicKeys {
					stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, outPublicKey)
					if err != nil {
						panic(err)
					}
					if !has {
						panic(errors.Errorf("Can not found info of this public key %v", outPublicKey))
					}
					if stakerInfo.AutoStaking() {
						newShardCandidates = append(newShardCandidates, swapInstruction.OutPublicKeyStructs[index])
					} else {
						delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
						delete(b.autoStake, outPublicKey)
						delete(b.stakingTx, outPublicKey)
					}
				}
			}
		}
	}
	return newBeaconCandidates, newShardCandidates, nil
}

func (b *BeaconCommitteeStateV1) processReplaceInstruction(
	swapInstruction *instruction.SwapInstruction,
	committeeChange *CommitteeChange,
	env *BeaconCommitteeStateEnvironment,
) error {
	removedCommittee := len(swapInstruction.InPublicKeys)
	if swapInstruction.ChainID == instruction.BEACON_CHAIN_ID {
		committeeChange.BeaconCommitteeReplaced[common.REPLACE_OUT] = append(committeeChange.BeaconCommitteeReplaced[common.REPLACE_OUT], swapInstruction.OutPublicKeyStructs...)
		// add new public key to committees
		committeeChange.BeaconCommitteeReplaced[common.REPLACE_IN] = append(committeeChange.BeaconCommitteeReplaced[common.REPLACE_IN], swapInstruction.InPublicKeyStructs...)
		remainedBeaconCommittees := b.beaconCommittee[removedCommittee:]
		b.beaconCommittee = append(swapInstruction.InPublicKeyStructs, remainedBeaconCommittees...)
	} else {
		shardID := byte(swapInstruction.ChainID)
		committeeReplace := [2][]incognitokey.CommitteePublicKey{}
		// update shard COMMITTEE
		committeeReplace[common.REPLACE_OUT] = append(committeeReplace[common.REPLACE_OUT], swapInstruction.OutPublicKeyStructs...)
		// add new public key to committees
		committeeReplace[common.REPLACE_IN] = append(committeeReplace[common.REPLACE_IN], swapInstruction.InPublicKeyStructs...)
		committeeChange.ShardCommitteeReplaced[shardID] = committeeReplace
		remainedShardCommittees := b.shardCommittee[shardID][removedCommittee:]
		b.shardCommittee[shardID] = append(swapInstruction.InPublicKeyStructs, remainedShardCommittees...)
	}
	for index := 0; index < removedCommittee; index++ {
		delete(b.autoStake, swapInstruction.OutPublicKeys[index])
		delete(b.stakingTx, swapInstruction.OutPublicKeys[index])
		delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
		b.autoStake[swapInstruction.InPublicKeys[index]] = false
		b.rewardReceiver[swapInstruction.InPublicKeyStructs[index].GetIncKeyBase58()] = swapInstruction.NewRewardReceiverStructs[index]
		b.stakingTx[swapInstruction.InPublicKeys[index]] = common.HashH([]byte{0})
	}
	err := statedb.StoreStakerInfo(
		env.ConsensusStateDB,
		swapInstruction.InPublicKeyStructs,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
	)
	return err
}

func (b BeaconCommitteeStateV1) AllCandidateSubstituteCommittees() []string {
	return b.getAllCandidateSubstituteCommittee()
}

func (b *BeaconCommitteeStateV1) getAllCandidateSubstituteCommittee() []string {
	res := []string{}
	for _, committee := range b.shardCommittee {
		committeeStr, err := incognitokey.CommitteeKeyListToString(committee)
		if err != nil {
			panic(err)
		}
		res = append(res, committeeStr...)
	}
	for _, substitute := range b.shardSubstitute {
		substituteStr, err := incognitokey.CommitteeKeyListToString(substitute)
		if err != nil {
			panic(err)
		}
		res = append(res, substituteStr...)
	}
	beaconCommittee := b.beaconCommittee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconCommittee)
	if err != nil {
		panic(err)
	}
	res = append(res, beaconCommitteeStr...)
	candidateShardWaitingForCurrentRandom := b.currentEpochShardCandidate
	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(candidateShardWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	res = append(res, candidateShardWaitingForCurrentRandomStr...)
	candidateShardWaitingForNextRandom := b.nextEpochShardCandidate
	candidateShardWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(candidateShardWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	res = append(res, candidateShardWaitingForNextRandomStr...)
	return res
}

func (b *BeaconCommitteeStateV1) processAutoStakingChange(committeeChange *CommitteeChange, env *BeaconCommitteeStateEnvironment) error {
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

func (b *BeaconCommitteeStateV1) Hash() (*BeaconCommitteeStateHash, error) {
	res, err := b.beaconCommitteeStateBase.Hash()
	if err != nil {
		return res, err
	}
	// Shard candidate root: shard current candidate + shard next candidate
	shardCandidateArr := append(b.currentEpochShardCandidate, b.nextEpochShardCandidate...)
	shardCandidateArrStr, err := incognitokey.CommitteeKeyListToString(shardCandidateArr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	tempShardCandidateHash, err := common.GenerateHashFromStringArray(shardCandidateArrStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	res.ShardCandidateHash = tempShardCandidateHash
	return res, nil
}

func (b BeaconCommitteeStateV1) CandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return b.nextEpochShardCandidate
}

func (b BeaconCommitteeStateV1) CandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return b.currentEpochShardCandidate
}
