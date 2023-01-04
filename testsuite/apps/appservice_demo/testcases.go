package main

import (
	"fmt"
	"log"

	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func shortKey(src string) string {
	return src[len(src)-5:]
}

func (v *Validator) watch(beaconHeight, epochBlockTime uint64) error {
	if beaconHeight%epochBlockTime == 1 {

	}
	return nil
}

func (v *Validator) shouldInBeaconWaiting(cs *jsonresult.CommiteeState) {
	k := shortKey(v.MiningPublicKey)
	for _, c := range cs.BeaconWaiting {
		if c == k {
			return
		}
	}
	log.Printf("key %s is not in beacon waiting list\n", k)
	panic("Stop")
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
	return nil
}
