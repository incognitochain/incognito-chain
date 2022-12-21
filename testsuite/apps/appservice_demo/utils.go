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
	validators map[string]*Validator,
	shardKeys, beaconKeys []string,
	app *devframework.AppService,
) {
	if beaconHeight == submitkeyHeight {
		//submitkey
		otaPrivateKey := "14yJXBcq3EZ8dGh2DbL3a78bUUhWHDN579fMFx6zGVBLhWGzr2V4ZfUgjGHXkPnbpcvpepdzqAJEKJ6m8Cfq4kYiqaeSRGu37ns87ss"
		log.Println("Start submitkey for ota privateKey:", shortKey(otaPrivateKey))
		app.SubmitKey(otaPrivateKey)
		for _, k := range shardKeys {
			v := validators[k]
			log.Println("Start submitkey for ota privateKey:", shortKey(v.OTAPrivateKey))
			app.SubmitKey(v.OTAPrivateKey)
		}
	} else if beaconHeight == convertTxHeight {
		//convert from token v1 to token v2
		privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
		log.Println("Start convert token v1 to v2 for privateKey:", shortKey(privateKey))
		app.ConvertTokenV1ToV2(privateKey)
		for _, k := range beaconKeys {
			v := validators[k]
			log.Println("Start submitkey for ota privateKey:", shortKey(v.OTAPrivateKey))
			app.SubmitKey(v.OTAPrivateKey)
		}
	} else if beaconHeight == sendFundsHeight {
		//Send funds to 30 nodes
		privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
		receivers := map[string]interface{}{}
		log.Println("Start send funds from privateKey:", shortKey(privateKey))

		for _, k := range beaconKeys {
			v := validators[k]
			receivers[v.PaymentAddress] = 90000000000000
		}
		for _, k := range shardKeys {
			v := validators[k]
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
