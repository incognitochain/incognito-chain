package committeestate

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

//processUnstakeInstruction : process unstake instruction from beacon block

func (b *BeaconCommitteeStateV2) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (*CommitteeChange, [][]string, error) {
	newCommitteeChange := committeeChange
	incurredInstructions := [][]string{}
	returnStakerInfoPublicKeys := make(map[byte][]string)
	stakingTxs := make(map[byte][]string)
	percentReturns := make(map[byte][]uint)
	indexNextEpochShardCandidate := make(map[string]int)
	for i, v := range b.shardCommonPool {
		key, err := v.ToBase58()
		if err != nil {
			return newCommitteeChange, nil, err
		}
		indexNextEpochShardCandidate[key] = i
	}

	for index, committeePublicKey := range unstakeInstruction.CommitteePublicKeys {
		if common.IndexOfStr(committeePublicKey, env.unassignedCommonPool) == -1 {
			if common.IndexOfStr(committeePublicKey, env.allSubstituteCommittees) != -1 {
				// if found in committee list then turn off auto staking
				if _, ok := b.autoStake[committeePublicKey]; ok {
					b.autoStake[committeePublicKey] = false
					newCommitteeChange.Unstake = append(newCommitteeChange.Unstake, committeePublicKey)
				}
			}
		} else {
			delete(b.autoStake, committeePublicKey)
			delete(b.stakingTx, committeePublicKey)
			delete(b.numberOfRound, committeePublicKey)
			delete(b.rewardReceiver, unstakeInstruction.CommitteePublicKeysStruct[index].GetIncKeyBase58())
			indexCandidate := indexNextEpochShardCandidate[committeePublicKey]
			b.shardCommonPool = append(b.shardCommonPool[:indexCandidate], b.shardCommonPool[indexCandidate+1:]...)
			stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, committeePublicKey)
			if err != nil {
				return newCommitteeChange, nil, err
			}
			if !has {
				return newCommitteeChange, nil, errors.New("Can't find staker info")
			}

			newCommitteeChange.NextEpochShardCandidateRemoved =
				append(newCommitteeChange.NextEpochShardCandidateRemoved, unstakeInstruction.CommitteePublicKeysStruct[index])

			returnStakerInfoPublicKeys[stakerInfo.ShardID()] =
				append(returnStakerInfoPublicKeys[stakerInfo.ShardID()], committeePublicKey)
			percentReturns[stakerInfo.ShardID()] =
				append(percentReturns[stakerInfo.ShardID()], 100)
			stakingTxs[stakerInfo.ShardID()] =
				append(stakingTxs[stakerInfo.ShardID()], stakerInfo.TxStakingID().String())
		}
	}

	for i, v := range returnStakerInfoPublicKeys {
		if v != nil {
			returnStakingIns := instruction.NewReturnStakeInsWithValue(
				v,
				i,
				stakingTxs[i],
				percentReturns[i],
			)
			incurredInstructions = append(incurredInstructions, returnStakingIns.ToString())
		}
	}
	return newCommitteeChange, incurredInstructions, nil
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

func (b *BeaconCommitteeStateV2) processUnstakeChange(committeeChange *CommitteeChange, env *BeaconCommitteeStateEnvironment) (*CommitteeChange, error) {

	newCommitteeChange := committeeChange

	unstakingIncognitoKey, err := incognitokey.CommitteeBase58KeyListToStruct(newCommitteeChange.Unstake)
	if err != nil {
		return newCommitteeChange, err
	}
	err = statedb.StoreStakerInfoV1(
		env.ConsensusStateDB,
		unstakingIncognitoKey,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
	)

	return newCommitteeChange, err
}
