package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

func (v *Validator) AddStaking(app *devframework.AppService, amount uint64) error {
	log.Printf("Start add staking from privateKey %s for candidatePaymentAddress %s with privateSeed %s rewardReceiver %s",
		shortKey(v.PrivateKey), shortKey(v.PaymentAddress), shortKey(v.MiningKey), shortKey(v.PaymentAddress))
	_, err := app.AddStaking(v.PrivateKey, v.MiningKey, v.PaymentAddress, amount)
	if err != nil {
		return err
	}
	return nil
}

func (v *Validator) ShardStaking(app *devframework.AppService) error {
	log.Printf("Start shard staking from privateKey %s for candidatePaymentAddress %s with privateSeed %s rewardReceiver %s",
		shortKey(v.PrivateKey), shortKey(v.PaymentAddress), shortKey(v.MiningKey), shortKey(v.PaymentAddress))
	if err := app.ShardStaking(v.PrivateKey, v.PaymentAddress, v.MiningKey, v.PaymentAddress, "", true); err != nil {
		return err
	}
	v.HasStakedShard = true
	return nil
}

func (v *Validator) BeaconStaking(app *devframework.AppService) error {
	log.Printf("Start beacon staking from privateKey %s for candidatePaymentAddress %s with privateSeed %s rewardReceiver %s",
		shortKey(v.PrivateKey), shortKey(v.PaymentAddress), shortKey(v.MiningKey), shortKey(v.PaymentAddress))
	if err := app.BeaconStaking(v.PrivateKey, v.PaymentAddress, v.MiningKey, v.PaymentAddress, "", true); err != nil {
		return err
	}
	v.HasStakedBeacon = true
	return nil
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
			receivers[v.PaymentAddress] = 270000000000000
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

func writeState(shardValidators, beaconValidators map[string]*Validator) error {
	svs, err := json.Marshal(shardValidators)
	if err != nil {
		return err
	}
	bvs, err := json.Marshal(beaconValidators)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile("shard-state.json", svs, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile("beacon-state.json", bvs, 0644); err != nil {
		return err
	}
	return nil
}

func readState(app *devframework.AppService) {
	svs, err0 := ioutil.ReadFile("shard-state.json")
	if err0 == nil {
		if err := json.Unmarshal(svs, &shardValidators); err != nil {
			panic(err)
		}
	}
	bvs, err1 := ioutil.ReadFile("beacon-state.json")
	if err1 == nil {
		if err := json.Unmarshal(bvs, &beaconValidators); err != nil {
			panic(err)
		}
	}
	if err0 == nil && err1 == nil {
		if lastCs == nil {
			lastCs = new(jsonresult.CommiteeState)
		}
		updateRole(shardValidators, beaconValidators, lastCs, true)
	} else {
		cs, err := getCSByHeight(0, app)
		if err != nil {
			panic(err)
		}
		if err = updateRole(shardValidators, beaconValidators, cs, true); err != nil {
			panic(err)
		}
		if lastCs == nil {
			lastCs = new(jsonresult.CommiteeState)
		}
		*lastCs = *cs
		cs.Print()
		if err = writeState(shardValidators, beaconValidators); err != nil {
			panic(err)
		}
	}
}

func readData(app *devframework.AppService) {
	shardValidators = map[string]*Validator{
		sKey0:  {},
		sKey1:  {},
		sKey2:  {},
		sKey3:  {},
		sKey4:  {},
		sKey5:  {},
		sKey6:  {},
		sKey7:  {},
		sKey8:  {},
		sKey9:  {},
		sKey10: {},
		sKey11: {},
	}

	beaconValidators = map[string]*Validator{
		bKey0: {},
		bKey1: {},
		bKey2: {},
		bKey3: {},
		bKey4: {},
		bKey5: {},
	}

	args := os.Args

	if len(args) > 1 {
		t := args[1:]
		for i, v := range t {
			if v == submitkeyArg {
				shouldSubmitKey = true
				shouldStop = true
			} else if v == stakingShardArg {
				shouldStakeShard = true
				shouldStop = true
				var err error
				watchBeaconIndex, err = strconv.Atoi(t[i+1])
				if err != nil {
					panic(err)
				}
			} else if v == stakingBeaconArg {
				shouldStakeBeacon = true
				shouldStop = true
				var err error
				watchBeaconIndex, err = strconv.Atoi(t[i+1])
				if err != nil {
					panic(err)
				}
			} else if v == unstakingBeaconArg {
				shouldUnstakeBeacon = true
				shouldStop = true
				var err error
				watchBeaconIndex, err = strconv.Atoi(t[i+1])
				if err != nil {
					panic(err)
				}
			} else if v == addStakingBeaconArg {
				shouldAddStakingBeacon = true
				shouldStop = true
				var err error
				watchBeaconIndex, err = strconv.Atoi(t[i+1])
				if err != nil {
					panic(err)
				}
			} else if v == watchValidatorArg {
				var err error
				shouldWatchValidator = true
				watchBeaconIndex, err = strconv.Atoi(t[i+1])
				if err != nil {
					panic(err)
				}
			} else if v == shouldWatchOnlyArg {
				shouldWatchOnly = true
			}
		}
	} else {
		shouldSubmitKey = true
		shouldStakeShard = true
		shouldStakeBeacon = true
		//shouldAddStakingBeacon = true
		shouldWatchValidator = false
	}

	if shouldWatchOnly {
		shouldSubmitKey = false
		shouldStakeShard = false
		shouldStakeBeacon = false
		shouldWatchValidator = false
	}

	data, err := ioutil.ReadFile("accounts.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &keys); err != nil {
		panic(err)
	}

	for _, k := range keys {
		if _, found := shardValidators[k.MiningKey]; found {
			shardValidators[k.MiningKey] = &Validator{Key: k, HasStakedShard: false, HasStakedBeacon: false, ActionsIndex: map[string]Action{}}
		}
		if _, found := beaconValidators[k.MiningKey]; found {
			beaconValidators[k.MiningKey] = &Validator{Key: k, HasStakedShard: false, HasStakedBeacon: false, ActionsIndex: map[string]Action{}}
		}
	}

	readState(app)
}
