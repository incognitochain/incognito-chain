package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

type Key struct {
	PrivateKey         string `json:"private_key"`
	PaymentAddress     string `json:"payment_address"`
	OTAPrivateKey      string `json:"ota_private_key"`
	MiningKey          string `json:"mining_key"`
	MiningPublicKey    string `json:"mining_public_key"`
	ValidatorPublicKey string `json:"validator_public_key"`
}

func main() {

	args := os.Args
	isSkipSubmitKey := false
	isOnlySubmitKey := false
	isWatchingOnly := false
	if len(args) > 1 {
		t, err := strconv.Atoi(args[1])
		if err != nil {
			panic(err)
		}
		if t == 1 {
			isSkipSubmitKey = true
		} else if t == 0 {
			isOnlySubmitKey = true
		} else if t == 2 {
			isWatchingOnly = true
		}
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

	log.Println("Will be listening to beacon height:", bHeight)
	var startStakingHeight uint64
	if isSkipSubmitKey {
		startStakingHeight = bHeight
	} else {
		startStakingHeight = bHeight + 30
	}
	log.Println("Will be start shard staking on beacon height:", startStakingHeight)

	app.OnBeaconBlock(bHeight, func(blk types.BeaconBlock) {
		if !isSkipSubmitKey {
			if blk.GetBeaconHeight() == bHeight {
				//submitkey
				otaPrivateKey := "14yJXBcq3EZ8dGh2DbL3a78bUUhWHDN579fMFx6zGVBLhWGzr2V4ZfUgjGHXkPnbpcvpepdzqAJEKJ6m8Cfq4kYiqaeSRGu37ns87ss"
				log.Println("Start submitkey for ota privateKey:", otaPrivateKey[len(otaPrivateKey)-5:])
				app.SubmitKey(otaPrivateKey)
				k := keys[0]
				log.Println("Start submitkey for ota privateKey:", k.OTAPrivateKey[len(k.OTAPrivateKey)-5:])
				app.SubmitKey(k.OTAPrivateKey)
			} else if blk.GetBeaconHeight() == bHeight+5 {
				//convert from token v1 to token v2
				privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
				log.Println("Start convert token v1 to v2 for privateKey:", privateKey[len(privateKey)-5:])
				app.ConvertTokenV1ToV2(privateKey)
			} else if blk.GetBeaconHeight() == bHeight+15 {
				//Send funds to 30 nodes
				privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
				receivers := map[string]interface{}{}
				log.Println("Start send funds from privateKey:", privateKey[len(privateKey)-5:])
				for _, v := range keys {
					receivers[v.PaymentAddress] = 2750000001000
				}
				app.PreparePRVForTest(privateKey, receivers)
			}
		}
		if isOnlySubmitKey {
			return
		}
		if blk.GetBeaconHeight() == startStakingHeight && !isWatchingOnly {
			//Stake one node
			k := keys[0]
			privateSeedBytes := common.HashB(common.HashB([]byte(k.PrivateKey)))
			privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
			log.Printf("Start staking from privateKey %s for candidatePaymentAddress %s with privateSeed %s rewardReceiver %s",
				k.PrivateKey[len(k.PrivateKey)-5:], k.PaymentAddress[len(k.PaymentAddress)-5:], privateSeed[len(privateSeed)-5:], k.PaymentAddress[len(k.PaymentAddress)-5:])
			app.ShardStaking(k.PrivateKey, k.PaymentAddress, privateSeed, k.PaymentAddress, "", true)
		} else if blk.GetBeaconHeight() >= startStakingHeight+2 {
			log.Println("get committee state at beacon height:", blk.GetBeaconHeight())
			cs, err := app.GetCommitteeState(0, "")
			if err != nil {
				panic(err)
			}
			if cs.IsDiffFrom(lastCs) {
				lastCs = new(jsonresult.CommiteeState)
				*lastCs = *cs
				cs.Print()
			}
		}

		/*} else if blk.GetBeaconHeight() == startStakingHeight+5 {*/
		/*//Unstake one node*/
		/*privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"*/
		/*k := keys[0]*/
		/*privateSeedBytes := common.HashB(common.HashB([]byte(k.PrivateKey)))*/
		/*privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)*/
		/*log.Printf("Start unstake from privateKey %s for candidatePaymentAddress %s with privateSeed %s",*/
		/*privateKey[len(privateKey)-5:], k.PaymentAddress[len(k.PaymentAddress)-5:], privateSeed[len(privateSeed)-5:])*/

		/*app.ShardUnstaking(privateKey, k.PaymentAddress, privateSeed)*/
		/*}*/

	})

	//app.OnBeaconBlock(8664, func(blk types.BeaconBlock) {
	//	for sid, states := range blk.Body.ShardState {
	//		fmt.Println("Shard ", sid)
	//		for _, s := range states {
	//			fmt.Println(s.Height, s.Hash.String())
	//			fmt.Println(s.ValidationData, s.PreviousValidationData)
	//		}
	//	}
	//})

	/*app.OnShardBlock(0, 8650, func(blk types.ShardBlock) {*/
	/*shardID := blk.GetShardID()*/
	/*fmt.Println("blk", blk.GetHeight(), shardID, blk.GetVersion())*/
	/*})*/

	//for j := 0; j < 8; j++ {
	//	app.OnShardBlock(j, 2, func(blk types.ShardBlock) {
	//		shardID := blk.GetShardID()
	//		fmt.Println("blk", shardID, blk.GetHeight())
	//	})
	//}

	select {}
}
