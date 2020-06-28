package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

type BeaconCommitteeStateEnvironment struct {
	beaconHeight                               uint64
	beaconHash                                 common.Hash
	paramEpoch                                 uint64
	beaconInstructions                         [][]string
	newBeaconHeight                            uint64
	epochLength                                uint64
	epochBreakPointSwapNewKey                  []uint64
	randomNumber                               uint64
	randomFlag                                 bool
	allCandidateSubstituteCommittee            []string
	selectBeaconNodeSerializedPubkeyV2         map[uint64][]string
	selectBeaconNodeSerializedPaymentAddressV2 map[uint64][]string
	preSelectBeaconNodeSerializedPubkey        []string
	selectShardNodeSerializedPubkeyV2          map[uint64][]string
	selectShardNodeSerializedPaymentAddressV2  map[uint64][]string
	preSelectShardNodeSerializedPubkey         []string
}

func NewBeaconCommitteeStateEnvironment(beaconInstructions [][]string, newBeaconHeight uint64, epochLength uint64, epochBreakPointSwapNewKey []uint64, randomNumber uint64, randomFlag bool) *BeaconCommitteeStateEnvironment {
	return &BeaconCommitteeStateEnvironment{beaconInstructions: beaconInstructions, newBeaconHeight: newBeaconHeight, epochLength: epochLength, epochBreakPointSwapNewKey: epochBreakPointSwapNewKey, randomNumber: randomNumber, randomFlag: randomFlag}
}

type BeaconCommitteeStateV1 struct {
	beaconHeight                uint64
	beaconHash                  common.Hash
	beaconCommittee             []incognitokey.CommitteePublicKey
	beaconSubstitute            []incognitokey.CommitteePublicKey
	currentEpochShardCandidate  []incognitokey.CommitteePublicKey
	currentEpochBeaconCandidate []incognitokey.CommitteePublicKey
	nextEpochShardCandidate     []incognitokey.CommitteePublicKey
	nextEpochBeaconCandidate    []incognitokey.CommitteePublicKey
	shardCommittee              map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
	autoStaking                 map[string]bool
	rewardReceiver              map[string]string
}

func NewBeaconCommitteeStateV1() *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{
		shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
		autoStaking:     make(map[string]bool),
		rewardReceiver:  make(map[string]string),
	}
}

func NewBeaconCommitteeStateV1WithValue(beaconCommittee []incognitokey.CommitteePublicKey, beaconSubstitute []incognitokey.CommitteePublicKey, candidateShardWaitingForCurrentRandom []incognitokey.CommitteePublicKey, candidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey, candidateShardWaitingForNextRandom []incognitokey.CommitteePublicKey, candidateBeaconWaitingForNextRandom []incognitokey.CommitteePublicKey, shardCommittee map[byte][]incognitokey.CommitteePublicKey, shardSubstitute map[byte][]incognitokey.CommitteePublicKey, autoStaking map[string]bool, rewardReceiver map[string]string) *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{beaconCommittee: beaconCommittee, beaconSubstitute: beaconSubstitute, currentEpochShardCandidate: candidateShardWaitingForCurrentRandom, currentEpochBeaconCandidate: candidateBeaconWaitingForCurrentRandom, nextEpochShardCandidate: candidateShardWaitingForNextRandom, nextEpochBeaconCandidate: candidateBeaconWaitingForNextRandom, shardCommittee: shardCommittee, shardSubstitute: shardSubstitute, autoStaking: autoStaking, rewardReceiver: rewardReceiver}
}

func (b BeaconCommitteeStateV1) GenerateBeaconCommitteeInstruction(env *BeaconCommitteeStateEnvironment) {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GenerateCommitteeRootHashes() ([]common.Hash, error) {
	panic("implement me")
}

func (b *BeaconCommitteeStateV1) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (*incognitokey.CommitteeChange, error) {
	newB := b.cloneBeaconCommitteeStateV1()
	committeeChange := incognitokey.NewCommitteeChange()
	committeeStateInstruction := instruction.ImportCommitteeStateInstruction(env.beaconInstructions)
	newB.processStakeInstruction(committeeStateInstruction.StakeInstructions, env, committeeChange)
	err := newB.processSwapInstruction(committeeStateInstruction.SwapInstructions, env, committeeChange)
	if err != nil {
		return nil, err
	}
	newB.processStopAutoStakeInstruction(committeeStateInstruction.StopAutoStakeInstructions, env, committeeChange)
	return nil, nil
}

func (b *BeaconCommitteeStateV1) processStakeInstruction(
	stakeInstructions []*instruction.StakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *incognitokey.CommitteeChange,
) {
	for _, stakeInstruction := range stakeInstructions {
		for index, candidate := range stakeInstruction.PublicKeyStructs {
			b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceivers[index]
			b.autoStaking[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
		}
		if stakeInstruction.Chain == instruction.BEACON_INST {
			b.nextEpochBeaconCandidate = append(b.nextEpochBeaconCandidate, stakeInstruction.PublicKeyStructs...)
			committeeChange.NextEpochBeaconCandidateAdded = append(committeeChange.NextEpochBeaconCandidateAdded, stakeInstruction.PublicKeyStructs...)
		} else {
			b.nextEpochShardCandidate = append(b.nextEpochShardCandidate, stakeInstruction.PublicKeyStructs...)
			committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, stakeInstruction.PublicKeyStructs...)
		}
	}
}

func (b *BeaconCommitteeStateV1) processStopAutoStakeInstruction(
	stopAutoStakeInstructions []*instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *incognitokey.CommitteeChange,
) {
	for _, stopAutoStakeInstruction := range stopAutoStakeInstructions {
		for _, committeePublicKey := range stopAutoStakeInstruction.PublicKeys {
			if common.IndexOfStr(committeePublicKey, env.allCandidateSubstituteCommittee) == -1 {
				// if not found then delete auto staking data for this public key if present
				if _, ok := b.autoStaking[committeePublicKey]; ok {
					delete(b.autoStaking, committeePublicKey)
				}
			} else {
				// if found in committee list then turn off auto staking
				if _, ok := b.autoStaking[committeePublicKey]; ok {
					b.autoStaking[committeePublicKey] = false
					committeeChange.StopAutoStake = append(committeeChange.StopAutoStake, committeePublicKey)
				}
			}
		}
	}
}

func (b *BeaconCommitteeStateV1) processSwapInstruction(
	swapInstructions []*instruction.SwapInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *incognitokey.CommitteeChange,
) error {
	for _, swapInstruction := range swapInstructions {
		if common.IndexOfUint64(env.beaconHeight/env.paramEpoch, env.epochBreakPointSwapNewKey) > -1 || swapInstruction.IsReplace {
			err := b.processReplaceInstruction(swapInstruction, committeeChange)
			if err != nil {
				return err
			}
			continue
		} else {
			Logger.log.Debug("Swap Instruction In Public Keys", swapInstruction.InPublicKeys)
			Logger.log.Debug("Swap Instruction Out Public Keys", swapInstruction.OutPublicKeys)
			if swapInstruction.ChainID != instruction.BEACON_CHAIN_ID {
				shardID := byte(swapInstruction.ChainID)
				// delete in public key out of sharding pending validator list
				if len(swapInstruction.InPublicKeys) > 0 {
					shardSubstituteStr, err := incognitokey.CommitteeKeyListToString(b.shardSubstitute[shardID])
					if err != nil {
						return err
					}
					tempShardSubstitute, err := RemoveValidator(shardSubstituteStr, swapInstruction.InPublicKeys)
					if err != nil {
						return err
					}
					// update shard pending validator
					committeeChange.ShardSubstituteRemoved[shardID] = append(committeeChange.ShardSubstituteRemoved[shardID], swapInstruction.InPublicKeyStructs...)
					b.shardSubstitute[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardSubstitute)
					if err != nil {
						return err
					}
					// add new public key to committees
					committeeChange.ShardCommitteeAdded[shardID] = append(committeeChange.ShardCommitteeAdded[shardID], swapInstruction.InPublicKeyStructs...)
					b.shardCommittee[shardID] = append(b.shardCommittee[shardID], swapInstruction.InPublicKeyStructs...)
				}
				// delete out public key out of current committees
				if len(swapInstruction.OutPublicKeys) > 0 {
					//for _, value := range outPublickeyStructs {
					//	delete(beaconBestState.RewardReceiver, value.GetIncKeyBase58())
					//}
					shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(b.shardCommittee[shardID])
					if err != nil {
						return err
					}
					tempShardCommittees, err := RemoveValidator(shardCommitteeStr, swapInstruction.OutPublicKeys)
					if err != nil {
						return err
					}
					// remove old public key in shard committee update shard committee
					committeeChange.ShardCommitteeRemoved[shardID] = append(committeeChange.ShardCommitteeRemoved[shardID], swapInstruction.OutPublicKeyStructs...)
					b.shardCommittee[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardCommittees)
					if err != nil {
						return err
					}
					// Check auto stake in out public keys list
					// if auto staking not found or flag auto stake is false then do not re-stake for this out public key
					// if auto staking flag is true then system will automatically add this out public key to current candidate list
					for index, outPublicKey := range swapInstruction.OutPublicKeys {
						if isAutoStaking, ok := b.autoStaking[outPublicKey]; !ok {
							if _, ok := b.rewardReceiver[outPublicKey]; ok {
								delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
							}
							continue
						} else {
							if !isAutoStaking {
								// delete this flag for next time staking
								delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
								delete(b.autoStaking, outPublicKey)
							} else {
								shardCandidate, err := incognitokey.CommitteeBase58KeyListToStruct([]string{outPublicKey})
								if err != nil {
									return err
								}
								b.nextEpochShardCandidate = append(b.nextEpochShardCandidate, shardCandidate...)
								committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, shardCandidate...)
							}
						}
					}
				}
			} else {
				if len(swapInstruction.InPublicKeys) > 0 {
					beaconSubstituteStr, err := incognitokey.CommitteeKeyListToString(b.beaconSubstitute)
					if err != nil {
						return err
					}
					tempBeaconSubstitute, err := RemoveValidator(beaconSubstituteStr, swapInstruction.InPublicKeys)
					if err != nil {
						return err
					}
					// update beacon pending validator
					committeeChange.BeaconSubstituteRemoved = append(committeeChange.BeaconSubstituteRemoved, swapInstruction.InPublicKeyStructs...)
					b.beaconSubstitute, err = incognitokey.CommitteeBase58KeyListToStruct(tempBeaconSubstitute)
					if err != nil {
						return err
					}
					// add new public key to beacon committee
					committeeChange.BeaconCommitteeAdded = append(committeeChange.BeaconCommitteeAdded, swapInstruction.InPublicKeyStructs...)
					b.beaconCommittee = append(b.beaconCommittee, swapInstruction.InPublicKeyStructs...)
				}
				if len(swapInstruction.OutPublicKeys) > 0 {
					// delete reward receiver
					//for _, value := range swapInstruction.OutPublicKeyStructs {
					//	delete(beaconBestState.RewardReceiver, value.GetIncKeyBase58())
					//}
					beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(b.beaconCommittee)
					if err != nil {
						return err
					}
					tempBeaconCommittes, err := RemoveValidator(beaconCommitteeStr, swapInstruction.OutPublicKeys)
					if err != nil {
						return err
					}
					// remove old public key in beacon committee and update beacon best state
					committeeChange.BeaconCommitteeRemoved = append(committeeChange.BeaconCommitteeRemoved, swapInstruction.OutPublicKeyStructs...)
					b.beaconCommittee, err = incognitokey.CommitteeBase58KeyListToStruct(tempBeaconCommittes)
					if err != nil {
						return err
					}
					for index, outPublicKey := range swapInstruction.OutPublicKeys {
						if isAutoStaking, ok := b.autoStaking[outPublicKey]; !ok {
							if _, ok := b.rewardReceiver[outPublicKey]; ok {
								delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
							}
							continue
						} else {
							if !isAutoStaking {
								delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
								delete(b.autoStaking, outPublicKey)
							} else {
								beaconCandidate, err := incognitokey.CommitteeBase58KeyListToStruct([]string{outPublicKey})
								if err != nil {
									return err
								}
								b.nextEpochBeaconCandidate = append(b.nextEpochBeaconCandidate, beaconCandidate...)
								committeeChange.NextEpochBeaconCandidateAdded = append(committeeChange.NextEpochBeaconCandidateAdded, beaconCandidate...)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (b *BeaconCommitteeStateV1) processReplaceInstruction(
	swapInstruction *instruction.SwapInstruction,
	committeeChange *incognitokey.CommitteeChange,
) error {
	return nil
}

func (b BeaconCommitteeStateV1) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("implement me")
}

func (b BeaconCommitteeStateV1) GetBeaconHeight() uint64 {
	return b.beaconHeight
}
func (b BeaconCommitteeStateV1) GetBeaconHash() common.Hash {
	return b.beaconHash
}

func (b BeaconCommitteeStateV1) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return b.beaconCommittee
}

func (b BeaconCommitteeStateV1) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return b.beaconSubstitute
}

func (b BeaconCommitteeStateV1) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return b.currentEpochShardCandidate
}

func (b BeaconCommitteeStateV1) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return b.currentEpochBeaconCandidate
}

func (b BeaconCommitteeStateV1) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return b.nextEpochShardCandidate
}

func (b BeaconCommitteeStateV1) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return b.nextEpochBeaconCandidate
}

func (b BeaconCommitteeStateV1) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return b.shardCommittee[shardID]
}

func (b BeaconCommitteeStateV1) GetShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	return b.shardSubstitute[shardID]
}

func (b BeaconCommitteeStateV1) GetAutoStaking() map[string]bool {
	return b.autoStaking
}

func (b BeaconCommitteeStateV1) GetRewardReceiver() map[string]string {
	return b.rewardReceiver
}

func (b BeaconCommitteeStateV1) cloneBeaconCommitteeStateV1() *BeaconCommitteeStateV1 {
	newB := NewBeaconCommitteeStateV1()
	newB.beaconHeight = b.beaconHeight
	newB.beaconHash = b.beaconHash
	newB.beaconCommittee = b.beaconCommittee
	newB.beaconSubstitute = b.beaconSubstitute
	newB.currentEpochShardCandidate = b.currentEpochShardCandidate
	newB.currentEpochBeaconCandidate = b.currentEpochBeaconCandidate
	newB.nextEpochShardCandidate = b.nextEpochShardCandidate
	newB.nextEpochBeaconCandidate = b.nextEpochBeaconCandidate
	for k, v := range b.shardCommittee {
		newB.shardCommittee[k] = v
	}
	for k, v := range b.shardSubstitute {
		newB.shardSubstitute[k] = v
	}
	for k, v := range b.autoStaking {
		newB.autoStaking[k] = v
	}
	for k, v := range b.rewardReceiver {
		newB.rewardReceiver[k] = v
	}
	return newB
}
