package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

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
	//app.OnBeaconBlock(8664, func(blk types.BeaconBlock) {
	//	for sid, states := range blk.Body.ShardState {
	//		fmt.Println("Shard ", sid)
	//		for _, s := range states {
	//			fmt.Println(s.Height, s.Hash.String())
	//			fmt.Println(s.ValidationData, s.PreviousValidationData)
	//		}
	//	}
	//})

	privateKey := "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
	receivers := map[string]interface{}{}

	for _, v := range keys {
		receivers[v.PaymentAddress] = 2750000001000
	}

	app.PreparePRVForTest(privateKey, receivers)

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
