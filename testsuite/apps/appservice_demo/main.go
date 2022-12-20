package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

const (
	submitkeyArg           = "submitkey"
	stakingShardArg        = "staking-shard"
	stakingBeaconArg       = "staking-beacon"
	stopAutoStakeBeaconArg = "stop-auto-stake-beacon"
)

type Key struct {
	PrivateKey         string `json:"private_key"`
	PaymentAddress     string `json:"payment_address"`
	OTAPrivateKey      string `json:"ota_private_key"`
	MiningKey          string `json:"mining_key"`
	MiningPublicKey    string `json:"mining_public_key"`
	ValidatorPublicKey string `json:"validator_public_key"`
}

type Validator struct {
	Key
	HasStakedShard  bool `json:"has_staked_shard"`
	HasStakedBeacon bool `json:"has_staked_beacon"`
}

const (
	sKey0  = "12DuNECQJWcHM1CtK942EAHNLMFUR5aB5SSEEhsDZs7vxvq9aCj"
	sKey1  = "12G1n7hQHVeGZQxQuSVaBMj8kgzrgtbLZBtNDU8NPCnuzb98UbJ"
	sKey2  = "1HVbZV6og9wvd1KGSoQakVMNoeWLJzLrYDZ7spyoAR4tFwDXAL"
	sKey3  = "1dn4HYLrBu6Q9FZNM9WjEAkrWURPqB95BzSVzUcJpY2bBMk6rz"
	sKey4  = "1Kc6WCsxEa9BxwFbnbSiboADfW9FL2NUN7NWwM7PvY48B3ow9m"
	sKey5  = "12DdGt6ckrDmUcfhj6thfP3S5ajJLwjKjxDNsLXTVaJ1FL9hpZE"
	sKey6  = "12UkKRgNCPWR9FrSP2z92yXyyHF1AuL11RZDzfpnqnphC6ET8Pa"
	sKey7  = "1tkwFJt8FnTr1XEpnSmtF67xCEJWSZ24fNSsJpqUKbGDhGtLxE"
	sKey8  = "12C8AJzbBCo8Z2tjLaSEUv5G4KZEq8MEWubVh9LPg9KynY83X7u"
	sKey9  = "1FBpchyQkch8BojMUCtxNpBp3v3aYwFHjk41836m3ooeKVpf34"
	sKey10 = "12nLs1ftPJuqQZy29mExUUx7K1ZDWreTYxtCr7E44EoCK6rP4ET"
	sKey11 = "12crV4U6fJsh1zd8Kk24yiu7A85WhGdKZiZrMaLBLn99ohJgZba"
	bKey0  = "1ybFGPehhSWCMzyvHwVZBqrtFBTV3H9MgQfhHbNR5rZApH836s"
	bKey1  = "12F2sX5PptkKroF2tJztML3CU2MZRmoUXX5znAqLJWqmDL916eZ"
	bKey2  = "12TKYy4n44MXpUmKHQuTNhNFwGjkSxpJHaATL5vL4e78btDPGPw"
	bKey3  = "1EnCA7Jrb9NsvnXc6JxGDkds85VeE3zMRcoxHe4EnKfsUGVS3f"
	bKey4  = "1267P2JH7yU2Bw3fmz2dp51AgKwr7aqAZmPBg2bpLoXb3YwiCBq"
	bKey5  = "126ySw8rdeAbcfozNe3i9TRXNowrMmkoPsjWWKadSGfzuhzjP2n"
)

func main() {
	shouldSubmitKey := false
	shouldStakeShard := false
	shouldStakeBeacon := false
	shouldStopAutoStakeBeacon := false

	shardKeys := []string{sKey0, sKey1, sKey2, sKey3, sKey4, sKey5, sKey6, sKey7, sKey8, sKey9, sKey10, sKey11}
	beaconKeys := []string{bKey0, bKey1, bKey2, bKey3, bKey4, bKey5}

	validators := map[string]*Validator{
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
		bKey0:  {},
		bKey1:  {},
		bKey2:  {},
		bKey3:  {},
		bKey4:  {},
		bKey5:  {},
	}

	args := os.Args

	if len(args) > 1 {
		t := args[1:]
		for _, v := range t {
			if v == submitkeyArg {
				shouldSubmitKey = true
			} else if v == stakingShardArg {
				shouldStakeShard = true
			} else if v == stakingBeaconArg {
				shouldStakeBeacon = true
			} else if v == stopAutoStakeBeaconArg {
				shouldStopAutoStakeBeacon = true
			}
		}
	} else {
		shouldSubmitKey = true
		shouldStakeShard = true
		shouldStakeBeacon = true
	}

	var keys []Key
	lastCs := &jsonresult.CommiteeState{}

	data, err := ioutil.ReadFile("accounts.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &keys); err != nil {
		panic(err)
	}

	for _, k := range keys {
		if _, found := validators[k.MiningKey]; found {
			validators[k.MiningKey] = &Validator{Key: k, HasStakedShard: false, HasStakedBeacon: false}
		}
	}

	fullnode := flag.String("h", "http://localhost:8334/", "Fullnode Endpoint")
	flag.Parse()

	app := devframework.NewAppService(*fullnode, true)

	bState, err := app.GetBeaconBestState()
	if err != nil {
		panic(err)
	}
	bHeight := bState.BeaconHeight + 5
	if bHeight < 15 {
		bHeight = 15
	}
	epochBlockTime := uint64(10)
	submitkeyHeight := bHeight
	convertTxHeight := bHeight + 5
	sendFundsHeight := bHeight + 15

	log.Println("Will be listening to beacon height:", bHeight)
	var startStakingHeight uint64
	if shouldSubmitKey {
		startStakingHeight = bHeight + 40
	} else {
		startStakingHeight = bHeight
	}
	startStakingBeaconHeight := startStakingHeight + epochBlockTime + 5
	log.Println("Will be starting shard staking on beacon height:", startStakingHeight)
	log.Println("Will be starting beacon staking on beacon height:", startStakingBeaconHeight)

	app.OnBeaconBlock(bHeight, func(blk types.BeaconBlock) {
		if shouldSubmitKey {
			submitkeys(
				blk.GetBeaconHeight(),
				submitkeyHeight, convertTxHeight, sendFundsHeight,
				validators, shardKeys, beaconKeys, app,
			)
		}
		if shouldStakeShard {
			if blk.GetBeaconHeight() == startStakingHeight {
				//Stake each nodes
				for _, v := range validators {
					v.ShardStaking(app)
				}
			}
		}
		if shouldStakeBeacon {
			if blk.GetBeaconHeight() >= startStakingBeaconHeight {
				cs, err := getCSByHeight(blk.GetBeaconHeight(), app)
				if err != nil {
					panic(err)
				}
				//Stake beacon nodes
				for _, k := range beaconKeys {
					v := validators[k]
					if !v.HasStakedBeacon {
						var shouldStake bool
						for _, committee := range cs.Committee {
							for _, c := range committee {
								miningPublicKey := v.MiningPublicKey[len(v.MiningPublicKey)-5:]
								if c == miningPublicKey {
									shouldStake = true
								}
							}
						}
						if shouldStake {
							v.BeaconStaking(app)
						}
					}
				}
			}
		}
		if shouldStopAutoStakeBeacon {
			v := validators[bKey0]
			resp, err := app.StopAutoStaking(v.PrivateKey, v.PaymentAddress, v.MiningKey)
			if err != nil {
				panic(err)
			}
			fmt.Println(resp)
		}
		cs, err := getCSByHeight(blk.GetBeaconHeight(), app)
		if err != nil {
			panic(err)
		}
		if cs.IsDiffFrom(lastCs) {
			lastCs = new(jsonresult.CommiteeState)
			*lastCs = *cs
			cs.Print()
		}
	})

	select {}
}

func (v *Validator) ShardStaking(app *devframework.AppService) {
	log.Printf("Start shard staking from privateKey %s for candidatePaymentAddress %s with privateSeed %s rewardReceiver %s",
		v.PrivateKey[len(v.PrivateKey)-5:], v.PaymentAddress[len(v.PaymentAddress)-5:], v.MiningKey[len(v.MiningKey)-5:], v.PaymentAddress[len(v.PaymentAddress)-5:])
	app.ShardStaking(v.PrivateKey, v.PaymentAddress, v.MiningKey, v.PaymentAddress, "", true)
	v.HasStakedShard = true
}

func (v *Validator) BeaconStaking(app *devframework.AppService) {
	log.Printf("Start beacon staking from privateKey %s for candidatePaymentAddress %s with privateSeed %s rewardReceiver %s",
		v.PrivateKey[len(v.PrivateKey)-5:], v.PaymentAddress[len(v.PaymentAddress)-5:], v.MiningKey[len(v.MiningKey)-5:], v.PaymentAddress[len(v.PaymentAddress)-5:])
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
		log.Println("Start submitkey for ota privateKey:", otaPrivateKey[len(otaPrivateKey)-5:])
		app.SubmitKey(otaPrivateKey)
		for _, k := range shardKeys {
			v := validators[k]
			log.Println("Start submitkey for ota privateKey:", v.OTAPrivateKey[len(v.OTAPrivateKey)-5:])
			app.SubmitKey(v.OTAPrivateKey)
		}
	} else if beaconHeight == convertTxHeight {
		//convert from token v1 to token v2
		privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
		log.Println("Start convert token v1 to v2 for privateKey:", privateKey[len(privateKey)-5:])
		app.ConvertTokenV1ToV2(privateKey)
		for _, k := range beaconKeys {
			v := validators[k]
			log.Println("Start submitkey for ota privateKey:", v.OTAPrivateKey[len(v.OTAPrivateKey)-5:])
			app.SubmitKey(v.OTAPrivateKey)
		}
	} else if beaconHeight == sendFundsHeight {
		//Send funds to 30 nodes
		privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
		receivers := map[string]interface{}{}
		log.Println("Start send funds from privateKey:", privateKey[len(privateKey)-5:])

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

func StopAutoStake() {

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

var fixedCommiteesNodes = map[int][]string{
	-1: {
		"121VhftSAygpEJZ6i9jGkEKLMQTKTiiHzeUfeuhpQCcLZtys8FazpWwytpHebkAwgCxvqgUUF13fcSMtp5dgV1YkbRMj3z42TW2EebzAaiGg2DkGPodckN2UsbqhVDibpMgJUHVkLXardemfLdgUqWGtymdxaaRyPM38BAZcLpo2pAjxKv5vG5Uh9zHMkn7ZHtdNHmBmhG8B46UeiGBXYTwhyMe9KGS83jCMPAoUwHhTEXj5qQh6586dHjVxwEkRzp7SKn9iG1FFWdJ97xEkP2ezAapNQ46quVrMggcHFvoZofs1xdd4o5vAmPKnPTZtGTKunFiTWGnpSG9L6r5QpcmapqvRrK5SiuFhNM5DqgzUeHBb7fTfoiWd2N29jkbTGSq8CPUSjx3zdLR9sZguvPdnAA8g25cFPGSZt8aEnFJoPRzM",
		"121VhftSAygpEJZ6i9jGkEqPGAXcmKffwMbzpwxnEfzJxen4oZKPukWAUBbqvV5xPnowZ2eQmAj2mEebG2oexebQPh1MPFC6vEZAk6i7AiRPrZmfaRrRVrBp4WXnVJmL3xK4wzTfkR2rZkhUmSZm112TTyhDNkDQSaBGJkexrPbryqUygazCA2eyo6LnK5qs7jz2RhhsWqUTQ3sQJUuFcYdf2pSnYwhqZqphDCSRizDHeysaua5L7LwS8fY7KZHhPgTuFjvUWWnWSRTmV8u1dTY5kcmMdDZsPiyN9WfqjgVoTFNALjFG8U4GMvzV3kKwVVjuPMsM2XqyPDVpdNQUgLnv2bJS8Tr22A9NgF1FQfWyAny1DYyY3N5H3tfCggsybzZXzrbYPPgokvEynac91y8hPkRdgKW1e7FHzuBnEisPuKzy",
		"121VhftSAygpEJZ6i9jGkGLcYhJBeaJTGY5aFjqQA2WwyxU69Utrviuy9AJ3ATkeEyigVGScQUZw22cD1HeFKiyASYAs82WEamujt3nefYA9FPhURBpRTn6jDmGKUdb4QNbs7HVCJkRRaL9aktg1yaQaZE8TJFg2UeE9tBqUdmvD8fy36aDCYM5W86jaTVCXeEJQWPxUunP2EEL3e283PJ8zqPeBkpoFvkvhB28Hk3oRDeCCTC7QhbaV18ayKeToYqAxoUMBBihanfA33ixeX1daeKpajLCgDZ6jrfphwdYwQbf7dMcZ2NVvQ1a5JUCTJUZypwgKRt8tnTAKCowt2L1KNGP4NJJZm61cfHAGbKRyG9QxCJgK2SdMKsKPVefZSc9LbVaB7VeBby5LHxvMoCD7bN7g1HYRp4BX9n1fZJUeEkVa",
		"121VhftSAygpEJZ6i9jGkDjJj7e2cfgQvrLsPsmLhGMmGD9U9Knffa1MZAw79EijnpueVfTStN2VYt5jRqEr2DTjVqzUinwHVKWH4Tg4szHUntiBdWeqzNC4E8iiwC9Y2KtcRr3hBkpfqvyuBvchigatrigRvFVWu8H2RQqjvopLL51DQ4LFD87L9Zgj9HhasMeyr6f37yirs47JgtGs4BM7EhhpM5zD3TCsFabPphtwDKnfuLMaGzoAw5fM8zEXvdLMuohk96oayjdYothncdtZom17DxB1Mmw535eEjxBwz9ELoZRKk3LYiheSd4xGN9QsxrT2WnZCTd8B5QktARte5S91QYvRMixKC8UEuovQhXt8jMZNkq7CmMeXoybfYdmNaAHuqbY1QeUT2AgaqPho4ay3z5eeKRhnB28H18RGWQ1L",
	},
	0: {
		"121VhftSAygpEJZ6i9jGkKZixz3MxWLn8H4bq6g9i1dxGkbJYni7TgHu4rrmzdnrzNeudRBVepuX66NTNzBP1vxjtVuFAYT7BkCkzKceR9CpQnYs2zJpZNhLeufPBFSpdYz397T2LQHnFAQit44f961H8z1LBbA5t8SEe9rs6GSf57iSkaDer752nZfKyn6arudjiDnzgMgi3uPh62wewSnJtZ71RXukKiYVFjAy9MCeAwHTfAzY28Fc7o4VnA185H9QDZSdY8b4pHDbk6Bx8dsMcFsVwDHv7kzWbprajj8KBqKm24SmjUCLm8HJ3BNuLUwtcd3vPZb1s5BavkK2LGbogWJW9PkrUwyBuffb9U11ziqEpBR92EfdfWJ9GwpsLqgx1KvirvAVZkzEswSf3vuBRe9QLWAXzP3Kq8BgpNJXVBMd",
		"121VhftSAygpEJZ6i9jGkBdCTvErp9PhxXoZonqQxZE6N7qFGh9D7rJw6Q6sCHGmcpeBP4eTiipMVkJfmNmK2Vft1aDkS9tH4c97rZ5eHt9CSiaC7bpNUbDjiMujadQZyY1Hs3UrHwzhF1d8LSV2gUtYk5cVtrw1Lns2gN95u9z84XqHpvh3v8ZhuiXvJP3hqaMAd8vzWV8hd495GUifLjqmDjmypH6FQT8wSzm99ssPASFxeTRuT9pDNu8n5SKyfcbdNt8K3iczYbvkYvsDqd8NnqcKUtEHKxNNh9ArPePcxi953HdPPFFFQGFzqJpPRiKrTDbG1vSgKRWh6YUrUseP6HYZzpTBnLw2fb6Kn31Hw6eLmgR2Bb3N65vHRrsc8P5VTerj7nH8fpdKpJbp2ZCy8te4ejXqtCXA5zQMtc8TTupj",
		"121VhftSAygpEJZ6i9jGkG3MhETkzDfyLzSZZD1vAt8xyPjmeN4s5NKcn2cjePBSpRNeH3ar67pgKAaxjTrpGRqc8g3uuE34D6ZXiRDFQhYBpEFQnZC3nm1HncqJy2riucweTJaTCxNpwTtak915qAp5NSjzFY9qMiSYAgz5QC8iWXqtPLdtJazAoekqTNxFRaBnsBfuLL8TEjjF4Y7awZmyu1G3qke8VsmNdVPqTv7RW42n9fxVBSVHv4QH9h6AQYoZxzDjA8Jd2P9Z61gEQiLZ6E5bbRDojQA5EDfSw4nrALqE5LRGMQCWUDKJQbwXsKEqvkYfuwYqQKW5n8dgpopduJ5XjAZyhuN6AT6Vb3x7tSsj4EzqFohyDSEhkKLz7BhSW6fQ4nYktUAiuCFPm8KRmqLTcPqCz8o6CA3UE2CPCQKK",
		"121VhftSAygpEJZ6i9jGkMtdVkuSDwxaYnUsHEnjbJBckPp1XrTBEyVHnx66bGCZiXesZMcZBxQD2fqaeWPZuTLw4Av7wrTeSnNLg8ErTbhFfhJD5nrTSCdCbnLbmybQiVYtUGcgMtRmnxAriaVL5dEBNkvNuUoxVzKXSSitnRAFQfA4BpPX1S8vR7zWtJP77CsYo47tnvcu8jSCEtjwEjGNeuNZPSzfnqBRyztYP1sMDgBvvJxUvMm8nTAxMm6YYVabVEdPBkJE89ZN5ZB7NE3SLxB4exqVKrcEoXVwGgJkWwdLaiQxFrDVgP2gi4RjGKpyrhvNCjjyU63Kp5aBFRcb8epkrByBERwDib6yeHmTZ22qFDQmsrdpVPGeafBvhuhNghKv9impjsbpBusu6BiEcSC7H5CEz9XzrGBWgDxvK9VC",
	},
	1: {
		"121VhftSAygpEJZ6i9jGk4LgFbiKrTAEiaittCRCifnQppfZxVHNnhkhTUf2b3HyZhMHp2yargJBz78pWDch3ou2phSTp18VYqbuJLrgywDXRdeERkv3U7Bnj3fiRfWNziuzZ7cS9LquUP3QeiLmKPdBUUp5iFq19VSy9TAhsSTeCB3tFNNiiuBiXVQEEmw8ypaHQ2LxG6YQAJ2iu416SuU598UktYhNuFpi99hzt88hNFVivEu1KDsUApvujLFeGXc1hZNq3oN8aDAfHmuxALqJUVjL9ts8f1pMi66m1tDK2yWETxg1DYeEcP2wCHACAFuGez3fZHwU5J9msynPtC1JgggE4iArJtN9vaK359bpkteiLcJvP5Q7TFxGTJbndy8UoGEgvoq1e1FqZLMk5tD2ov785HKkED2fMntMoJ9wiSh3",
		"121VhftSAygpEJZ6i9jGk5pf6nehBKH6X8NPU5hcz1Zg5G4chjcH7JH1stAUA83SbKJDbwnjVJ6Hm4RoaekTb5ruTpxGL6VcF5Zq6TLAq5hrQAvaGfCwzwta63srWEe5b8BtrM1Z8iMWH2TjbhBdKp8NnMHXUdP6cvuxyRJMrVkiuabM6tZYLc2xQPx2RQKAXApt1kBFTWfVRZdCzLzAFP1ZnKRhYFe2AFdgzeR21t4mFfeiqLEEp8MfnhDtkXRVqAWXTisFqNUeZzTD1XFk7pEnvu4uKijNfoRChapKkh6VJzXBAFr2HFdzLH27iJsLnp2uTtrcdNXQbssDBZNc7CaRXsyN7PuwWmiFQXbm1dSbuRLgmzp6oxWxXunPmvMugNBfW4U6Pj9wWkF2TMLmPbV4JgSqB4hq8jGFw2aSQP6jmTwL",
		"121VhftSAygpEJZ6i9jGkKHnnScYhTyYxWukbukeuQgTsQ5AiVvcKUnFH6u8rPZvUro9vMYdrayMcPoau5KxY2pkd9z9uMExzqYYFEmRTBG9dGNubWW2cmg7sPaJtFeo7ZeM2hnMyjnQwvZqfhYZGvLQ4UoAxTXpU9oUxtED7LSE3fx3219P4ZjyKKi63AHRbiK3BAc9iVnTHx58fXk9furAqA5PvxiVPC8w5XhCsWN6oG8EMS1czg1u67qeZsZA4zKd6MxcxwYPc9kuNPP5fcEtzdgVJUh5B62mQ5ytWiP7MecNkwtSsmJayhKLeJHzxijYaKN6vbaxh5J9t4xcEVzaFZDL1oA9zzCcZ1vTgWFgMNXpSZT6PXk8EL3n1VZRbiW4HEimnqgmpM43NNCjuw5xCyHkLJzcYdeYRWWDauYjKJpX",
		"121VhftSAygpEJZ6i9jGkMZbL9tf99A48srDuCvihAowAeiqRZ5YRXm7QJPJ1vfqYKyvgWzx4B9TrCv1MnBwMZuyL8XBQxM5ve8k1tTsbyWxwMkuyzvmDQw4sFcZQ4sZQSbnemKBmZmATWzAFKBvfZQxsfTLjFE3KPJ2jtK9dyuVNdXxw24JwDWqivg3sZfJ4yxTpoj2wikxmrCMjmavivPWptdVuPHCW5Wo5NE1isAZXDndpiYpDrFdctBWesReYPHBiEygGbmXiP5idcnTFt6LjMU5nEpvZrBEgKpHh2F4EMY9Jjq3yqygLmpMnA8bQBtNsPhXDrwLbM3ncJLKAZ1QBkVb4zfttYYStEK2mJjpBJKpDDLsURauuBSr6DgB3K1FjZiawhuLnnYXRi9TueWG4aAud1MLkPJMG55Z46XMvZ6J",
	},
}

var fixedRewardReceivers = []string{
	"126NQDuq7gLwugfit6h8J5S9K59ci4cpX4nKGp3mRM1PrR1Tuyu",
	"126g4qkfqpuzprAVzCBqKj1WGRAEAzjzxXKTrJBnhi1WKWbm9h9",
	"12iZ44KsLqaeis6opt2CpKpuFw9WY5FLydVStvWm2fiTenuTCzy",
	"12vciM6T9CtscVWwkfPfBYp2UATVm7X6dstZNxgRGQeWTkZZ1E5",
	"1F1y93YbnLvYRzsoEftnEjGA1GE1ULSRatEZ8HfohDnQe1MiPe",
	"1SdRA4PQj8x62SCJD1XP8QzH2pV1dcU4CGFdUkBTnXdaaNiRxX",
	"1WBVehvXnEvoYQz4DdD33vcWE35orkgB3N1eDLCGKnc7KBbHGV",
	"1a9jQQFtunBB2zJ86w31kR7DSFE7XuraVjUYXQdAKFY15q5Du7",
	"1g9PoWTHfLprcy9bCvYNxaUiR1hDcBpujxSfr3QetxZrt1q5Ee",
	"1i2qrYwZzgNrjvFpmDmCVG72Vnaye4FPzLYPk5cpjgy8QCo4oa",
	"1qVyf5hje3crq4V15NFkjQNyY2MDKjQaatPpLveCewYuKfUdqU",
	"1ryW3G6NYsFDrAnwWh3Ck6uEMWjxDpjdS1ES4p6UtV1eGbjEX3",
}

//notes
// map keys:
///ZUkjYnY9 -> 45z8B1VFD
