package main

import (
	"log"

	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

func (v *Validator) ShardStaking(app *devframework.AppService) {
	log.Printf("Start shard staking from privateKey %s for candidatePaymentAddress %s with privateSeed %s rewardReceiver %s",
		shortKey(v.PrivateKey), shortKey(v.PaymentAddress), shortKey(v.MiningKey), shortKey(v.PaymentAddress))
	app.ShardStaking(v.PrivateKey, v.PaymentAddress, v.MiningKey, v.PaymentAddress, "", true)
	v.HasStakedShard = true
}

func (v *Validator) BeaconStaking(app *devframework.AppService) {
	log.Printf("Start beacon staking from privateKey %s for candidatePaymentAddress %s with privateSeed %s rewardReceiver %s",
		shortKey(v.PrivateKey), shortKey(v.PaymentAddress), shortKey(v.MiningKey), shortKey(v.PaymentAddress))
	app.BeaconStaking(v.PrivateKey, v.PaymentAddress, v.MiningKey, v.PaymentAddress, "", true)
	v.HasStakedBeacon = true
}

func submitkeys(
	beaconHeight, submitkeyHeight, convertTxHeight, sendFundsHeight uint64,
	shardValidators, beaconValidators map[string]*Validator,
	app *devframework.AppService,
) {
	if beaconHeight == submitkeyHeight {
		//submitkey
		otaPrivateKey := "14yJXBcq3EZ8dGh2DbL3a78bUUhWHDN579fMFx6zGVBLhWGzr2V4ZfUgjGHXkPnbpcvpepdzqAJEKJ6m8Cfq4kYiqaeSRGu37ns87ss"
		log.Println("Start submitkey for ota privateKey:", shortKey(otaPrivateKey))
		app.SubmitKey(otaPrivateKey)
		for _, v := range shardValidators {
			log.Println("Start submitkey for ota privateKey:", shortKey(v.OTAPrivateKey))
			app.SubmitKey(v.OTAPrivateKey)
		}
	} else if beaconHeight == convertTxHeight {
		//convert from token v1 to token v2
		privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
		log.Println("Start convert token v1 to v2 for privateKey:", shortKey(privateKey))
		app.ConvertTokenV1ToV2(privateKey)
		for _, v := range beaconValidators {
			log.Println("Start submitkey for ota privateKey:", shortKey(v.OTAPrivateKey))
			app.SubmitKey(v.OTAPrivateKey)
		}
	} else if beaconHeight == sendFundsHeight {
		//Send funds to 30 nodes
		privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
		receivers := map[string]interface{}{}
		log.Println("Start send funds from privateKey:", shortKey(privateKey))

		for _, v := range beaconValidators {
			receivers[v.PaymentAddress] = 90000000000000
		}
		for _, v := range shardValidators {
			receivers[v.PaymentAddress] = 1760000000000
		}
		app.PreparePRVForTest(privateKey, receivers)
	}
}

func getCSByHeight(beaconHeight uint64, app *devframework.AppService) (*jsonresult.CommiteeState, error) {
	log.Println("get committee state at beacon height:", beaconHeight)
	cs, err := app.GetCommitteeState(0, "")
	if err != nil {
		return nil, err
	}
	cs.Filter(fixedCommiteesNodes, fixedRewardReceivers)
	return cs, nil
}

func updateRole(shardValidators, beaconValidators map[string]*Validator, cs *jsonresult.CommiteeState) {
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
				beaconValidators[k].Role = BeaconCommitteeRole
			}
		} else {
			for _, c := range v {
				if k, found := bvs[c]; found {
					beaconValidators[k].Role = ShardCommitteeRole
				}
				if k, found := svs[c]; found {
					shardValidators[k].Role = ShardCommitteeRole
				}
			}
		}
	}
	for i, v := range cs.Substitute {
		if i == -1 {
			for _, c := range v {
				k := bvs[c]
				beaconValidators[k].Role = BeaconPendingRole
			}
		} else {
			for _, c := range v {
				if k, found := bvs[c]; found {
					beaconValidators[k].Role = ShardPendingRole
				}
				if k, found := svs[c]; found {
					shardValidators[k].Role = ShardPendingRole
				}
			}
		}
	}
	for _, v := range cs.Syncing {
		for _, c := range v {
			if k, found := bvs[c]; found {
				beaconValidators[k].Role = ShardSyncingRole
			}
			if k, found := svs[c]; found {
				shardValidators[k].Role = ShardSyncingRole
			}
		}
	}
	for _, v := range cs.NextCandidate {
		if k, found := bvs[v]; found {
			beaconValidators[k].Role = ShardCandidateRole
		}
		if k, found := svs[v]; found {
			shardValidators[k].Role = ShardCandidateRole
		}
	}
	for _, v := range cs.CurrentCandidate {
		if k, found := bvs[v]; found {
			beaconValidators[k].Role = ShardCandidateRole
		}
		if k, found := svs[v]; found {
			shardValidators[k].Role = ShardCandidateRole
		}
	}
	for _, v := range cs.BeaconWaiting {
		if k, found := bvs[v]; found {
			beaconValidators[k].Role = BeaconWaitingRole
		}
	}
}
