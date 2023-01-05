package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

func shortKey(src string) string {
	return src[len(src)-5:]
}

func (v *Validator) watch(beaconHeight uint64, app *devframework.AppService) error {
	if v.Role != NormalRole {
		// stake shard must not in shard or beacon validator
		if err := v.ShardStaking(app); err == nil {
			panic("Expect error with shard staking")
		}
	}
	if v.Role == ShardPendingRole || v.Role == ShardCommitteeRole {
		// Validator must not in pending or shard committee
		if err := v.BeaconStaking(app); err == nil {
			panic("Expect error with beacon staking")
		}
	}
	if v.Role == BeaconCommitteeRole || v.Role == BeaconPendingRole || v.Role == BeaconWaitingRole || v.Role == BeaconLockingRole {
		// Validator must not stake beacon
		// Validator must not is in beacon locking
		if err := v.BeaconStaking(app); err == nil {
			panic("Expect error with beacon staking")
		}
		// Validator must add stake amount % 1750 = 0
		if err := v.AddStaking(app, 175000*1e9+100); err == nil {
			panic("Expect error with add staking")
		}
	}
	for key, value := range v.ActionsIndex {
		if key == addStakingBeaconArg {
			if beaconHeight > value.Height && beaconHeight-value.Height > 3 {
				stakerInfo, err := app.GetBeaconStakerInfo(beaconHeight, v.MiningPublicKey)
				if err != nil {
					panic(err)
				}

			}
		}
	}

	return nil
}

func updateRole(shardValidators, beaconValidators map[string]*Validator, cs *jsonresult.CommiteeState, isInit bool) error {
	bvs := map[string]string{}
	for _, v := range beaconValidators {
		bvs[shortKey(v.MiningPublicKey)] = v.MiningKey
	}
	svs := map[string]string{}
	for _, v := range shardValidators {
		svs[shortKey(v.MiningPublicKey)] = v.MiningKey
	}
	for i, v := range cs.Committee {
		if i == -1 {
			for _, c := range v {
				k := bvs[c]
				role := beaconValidators[k].Role
				if role != BeaconCommitteeRole && role != BeaconPendingRole && !isInit {
					return fmt.Errorf("beacon key %s is %v role when switch to BeaconCommitteeRole", shortKey(beaconValidators[k].MiningPublicKey), role)
				}
				beaconValidators[k].Role = BeaconCommitteeRole
			}
		} else {
			for _, c := range v {
				if k, found := bvs[c]; found {
					role := beaconValidators[k].Role
					if role != ShardCommitteeRole && role != BeaconWaitingRole && role != ShardPendingRole && !isInit {
						return fmt.Errorf("beacon key %s is %v role when switch to ShardCommitteeRole", shortKey(beaconValidators[k].MiningPublicKey), role)
					}
					if role != BeaconWaitingRole {
						beaconValidators[k].Role = ShardCommitteeRole
					}
				}
				if k, found := svs[c]; found {
					role := shardValidators[k].Role
					if role != ShardCommitteeRole && role != ShardPendingRole && !isInit {
						return fmt.Errorf("shard key %s is %v role when switch to ShardCommitteeRole", shortKey(shardValidators[k].MiningPublicKey), role)
					}
					shardValidators[k].Role = ShardCommitteeRole
				}
			}
		}
	}
	for i, v := range cs.Substitute {
		if i == -1 {
			for _, c := range v {
				k := bvs[c]
				role := beaconValidators[k].Role
				if role != BeaconPendingRole && role != BeaconWaitingRole && !isInit {
					return fmt.Errorf("beacon key %s is %v role when switch to BeaconPendingRole", shortKey(beaconValidators[k].MiningPublicKey), role)
				}
				beaconValidators[k].Role = BeaconPendingRole
			}
		} else {
			for _, c := range v {
				if k, found := bvs[c]; found {
					role := beaconValidators[k].Role
					if role != ShardPendingRole && role != ShardSyncingRole && role != ShardCommitteeRole && !isInit && role != BeaconWaitingRole {
						return fmt.Errorf("beacon key %s is role %v when switch to ShardPendingRole", shortKey(beaconValidators[k].MiningPublicKey), role)
					}
					if role != BeaconWaitingRole {
						beaconValidators[k].Role = ShardPendingRole
					}
				}
				if k, found := svs[c]; found {
					role := shardValidators[k].Role
					if role != ShardPendingRole && role != ShardSyncingRole && role != ShardCommitteeRole && !isInit {
						return fmt.Errorf("shard key %s is role %v when switch to ShardPendingRole", shortKey(shardValidators[k].MiningPublicKey), role)
					}
					shardValidators[k].Role = ShardPendingRole
				}
			}
		}
	}
	for _, v := range cs.Syncing {
		for _, c := range v {
			if k, found := bvs[c]; found {
				role := beaconValidators[k].Role
				if role != ShardSyncingRole && role != ShardCandidateRole && !isInit {
					return fmt.Errorf("beacon key %s is role %v when switch to ShardSyncingRole", shortKey(beaconValidators[k].MiningPublicKey), role)
				}
				beaconValidators[k].Role = ShardSyncingRole
			}
			if k, found := svs[c]; found {
				role := shardValidators[k].Role
				if role != ShardSyncingRole && role != ShardCandidateRole && !isInit {
					return fmt.Errorf("shard key %s is role %v when switch to ShardSyncingRole", shortKey(shardValidators[k].MiningPublicKey), role)
				}
				shardValidators[k].Role = ShardSyncingRole
			}
		}
	}
	for _, v := range cs.NextCandidate {
		if k, found := bvs[v]; found {
			role := beaconValidators[k].Role
			if role != ShardCandidateRole && role != NormalRole && !isInit {
				return fmt.Errorf("beacon key %s is role %v when switch to ShardCandidateRole", shortKey(beaconValidators[k].MiningPublicKey), role)
			}
			beaconValidators[k].Role = ShardCandidateRole
		}
		if k, found := svs[v]; found {
			role := shardValidators[k].Role
			if role != ShardCandidateRole && role != NormalRole && !isInit {
				return fmt.Errorf("shard key %s is role %v when switch to ShardCandidateRole", shortKey(shardValidators[k].MiningPublicKey), role)
			}
			shardValidators[k].Role = ShardCandidateRole
		}
	}
	for _, v := range cs.CurrentCandidate {
		if k, found := bvs[v]; found {
			role := beaconValidators[k].Role
			if role != ShardCandidateRole && role != NormalRole && !isInit {
				return fmt.Errorf("beacon key %s is role %v when switch to ShardCandidateRole", shortKey(beaconValidators[k].MiningPublicKey), role)
			}
			beaconValidators[k].Role = ShardCandidateRole
		}
		if k, found := svs[v]; found {
			role := shardValidators[k].Role
			if role != ShardCandidateRole && role != NormalRole && !isInit {
				return fmt.Errorf("shard key %s is role %v when switch to ShardCandidateRole", shortKey(shardValidators[k].MiningPublicKey), role)
			}
			shardValidators[k].Role = ShardCandidateRole
		}
	}
	for _, v := range cs.BeaconWaiting {
		if k, found := bvs[v]; found {
			role := beaconValidators[k].Role
			if role != BeaconWaitingRole && role != ShardCommitteeRole && !isInit {
				return fmt.Errorf("beacon key %s is role %v when switch to BeaconWaitingRole", shortKey(beaconValidators[k].MiningPublicKey), role)
			}
			beaconValidators[k].Role = BeaconWaitingRole
		}
	}
	for _, v := range cs.BeaconLocking {
		if k, found := bvs[v]; found {
			role := beaconValidators[k].Role
			if role != BeaconLockingRole && role != ShardCommitteeRole && !isInit {
				return fmt.Errorf("beacon key %s is role %v when switch to BeaconLockingRole", shortKey(beaconValidators[k].MiningPublicKey), role)
			}
			beaconValidators[k].Role = BeaconLockingRole
		}
	}

	return nil
}
