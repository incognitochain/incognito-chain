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

func updateRole(shardValidators, beaconValidators map[string]*Validator, cs *jsonresult.CommiteeState) error {
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
				if beaconValidators[k].Role != BeaconCommitteeRole {
					if beaconValidators[k].Role != BeaconPendingRole {
						return fmt.Errorf("beacon key %s is not BeaconPendingRole when switch to BeaconCommitteeRole", shortKey(beaconValidators[k].MiningPublicKey))
					}
					beaconValidators[k].Role = BeaconCommitteeRole
				}
			}
		} else {
			for _, c := range v {
				if k, found := bvs[c]; found {
					if beaconValidators[k].Role != ShardCommitteeRole && beaconValidators[k].Role != BeaconWaitingRole {
						if beaconValidators[k].Role != ShardPendingRole {
							return fmt.Errorf("beacon key %s is not ShardPendingRole when switch to ShardCommitteeRole", shortKey(beaconValidators[k].MiningPublicKey))
						}
						beaconValidators[k].Role = ShardCommitteeRole
					}
				}
				if k, found := svs[c]; found {
					if shardValidators[k].Role != ShardCommitteeRole {
						if shardValidators[k].Role != ShardPendingRole {
							return fmt.Errorf("shard key %s is not ShardPendingRole when switch to ShardCommitteeRole", shortKey(shardValidators[k].MiningPublicKey))
						}
						shardValidators[k].Role = ShardCommitteeRole
					}
				}
			}
		}
	}
	for i, v := range cs.Substitute {
		if i == -1 {
			for _, c := range v {
				k := bvs[c]
				if beaconValidators[k].Role != BeaconPendingRole {
					if beaconValidators[k].Role != BeaconWaitingRole {
						return fmt.Errorf("beacon key %s is not BeaconWaitingRole when switch to BeaconPendingRole", shortKey(beaconValidators[k].MiningPublicKey))
					}
					beaconValidators[k].Role = BeaconPendingRole
				}
			}
		} else {
			for _, c := range v {
				if k, found := bvs[c]; found {
					if beaconValidators[k].Role != ShardPendingRole {
						if beaconValidators[k].Role != ShardSyncingRole {
							return fmt.Errorf("beacon key %s is not ShardSyncingRole when switch to ShardPendingRole", shortKey(beaconValidators[k].MiningPublicKey))
						}
						beaconValidators[k].Role = ShardPendingRole
					}
				}
				if k, found := svs[c]; found {
					if shardValidators[k].Role != ShardPendingRole {
						if shardValidators[k].Role != ShardSyncingRole {
							return fmt.Errorf("shard key %s is not ShardSyncingRole when switch to ShardPendingRole", shortKey(shardValidators[k].MiningPublicKey))
						}
						shardValidators[k].Role = ShardPendingRole
					}
				}
			}
		}
	}
	for _, v := range cs.Syncing {
		for _, c := range v {
			if k, found := bvs[c]; found {
				if beaconValidators[k].Role != ShardSyncingRole {
					if beaconValidators[k].Role != ShardCandidateRole {
						return fmt.Errorf("beacon key %s is not ShardCandidateRole when switch to ShardSyncingRole", shortKey(beaconValidators[k].MiningPublicKey))
					}
					beaconValidators[k].Role = ShardSyncingRole
				}
			}
			if k, found := svs[c]; found {
				if shardValidators[k].Role != ShardSyncingRole {
					if shardValidators[k].Role != ShardCandidateRole {
						return fmt.Errorf("shard key %s is not ShardCandidateRole when switch to ShardSyncingRole", shortKey(shardValidators[k].MiningPublicKey))
					}
					shardValidators[k].Role = ShardSyncingRole
				}
			}
		}
	}
	for _, v := range cs.NextCandidate {
		if k, found := bvs[v]; found {
			if beaconValidators[k].Role != ShardCandidateRole {
				if beaconValidators[k].Role != NormalRole {
					return fmt.Errorf("beacon key %s is not NormalRole when switch to ShardCandidateRole", shortKey(beaconValidators[k].MiningPublicKey))
				}
				beaconValidators[k].Role = ShardCandidateRole
			}
		}
		if k, found := svs[v]; found {
			if shardValidators[k].Role != ShardCandidateRole {
				if shardValidators[k].Role != NormalRole {
					return fmt.Errorf("shard key %s is not NormalRole when switch to ShardCandidateRole", shortKey(shardValidators[k].MiningPublicKey))
				}
				shardValidators[k].Role = ShardCandidateRole
			}
		}
	}
	for _, v := range cs.CurrentCandidate {
		if k, found := bvs[v]; found {
			if beaconValidators[k].Role != ShardCandidateRole {
				if beaconValidators[k].Role != NormalRole {
					return fmt.Errorf("beacon key %s is not NormalRole when switch to ShardCandidateRole", shortKey(beaconValidators[k].MiningPublicKey))
				}
				beaconValidators[k].Role = ShardCandidateRole
			}
		}
		if k, found := svs[v]; found {
			if shardValidators[k].Role != ShardCandidateRole {
				if shardValidators[k].Role != NormalRole {
					return fmt.Errorf("shard key %s is not NormalRole when switch to ShardCandidateRole", shortKey(shardValidators[k].MiningPublicKey))
				}
				shardValidators[k].Role = ShardCandidateRole
			}
		}
	}
	for _, v := range cs.BeaconWaiting {
		if k, found := bvs[v]; found {
			if beaconValidators[k].Role != BeaconWaitingRole {
				if beaconValidators[k].Role != ShardCommitteeRole {
					return fmt.Errorf("beacon key %s is not ShardCommitteeRole when switch to BeaconWaitingRole", shortKey(beaconValidators[k].MiningPublicKey))
				}
				beaconValidators[k].Role = BeaconWaitingRole
			}
		}
	}
	return nil
}
