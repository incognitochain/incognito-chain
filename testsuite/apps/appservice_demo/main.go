package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
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
	var keys []Key

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

	app.OnBeaconBlock(bHeight, func(blk types.BeaconBlock) {
		if blk.GetBeaconHeight() == bHeight {
			//submitkey
			otaPrivateKey := "14yJXBcq3EZ8dGh2DbL3a78bUUhWHDN579fMFx6zGVBLhWGzr2V4ZfUgjGHXkPnbpcvpepdzqAJEKJ6m8Cfq4kYiqaeSRGu37ns87ss"
			app.AuthorizedSubmitKey(otaPrivateKey)
		} else if blk.GetBeaconHeight() == bHeight+5 {
			//convert from token v1 to token v2
			privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
			app.ConvertTokenV1ToV2(privateKey)
		} else if blk.GetBeaconHeight() == bHeight+10 {
			//submitkey to make sure
			otaPrivateKey := "14yJXBcq3EZ8dGh2DbL3a78bUUhWHDN579fMFx6zGVBLhWGzr2V4ZfUgjGHXkPnbpcvpepdzqAJEKJ6m8Cfq4kYiqaeSRGu37ns87ss"
			app.AuthorizedSubmitKey(otaPrivateKey)

		} else if blk.GetBeaconHeight() == bHeight+15 {
			//Send funds to 30 nodes
			privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
			receivers := map[string]interface{}{}

			for _, v := range keys {
				receivers[v.PaymentAddress] = 2750000001000
			}

			app.PreparePRVForTest(privateKey, receivers)
		} else if blk.GetBeaconHeight() == bHeight+20 {
			//Stake one node
			k := keys[0]
			privateSeedBytes := common.HashB(common.HashB([]byte(k.PrivateKey)))
			privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
			privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
			app.ShardStaking(privateKey, k.PaymentAddress, privateSeed, k.PaymentAddress, true)
		} else if blk.GetBeaconHeight() == bHeight+22 {
			list, err := app.GetCommitteeList()
			if err != nil {
				panic(err)
			}
			log.Println(list)

		} else if blk.GetBeaconHeight() == bHeight+25 {
			//Unstake one node
			privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
			k := keys[0]
			privateSeedBytes := common.HashB(common.HashB([]byte(k.PrivateKey)))
			privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
			app.ShardUnstaking(privateKey, k.PaymentAddress, privateSeed)
		}

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
