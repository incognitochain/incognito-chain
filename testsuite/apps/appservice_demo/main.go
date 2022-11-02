package main

import (
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/transaction"
	"log"
	"sync"
)

func main() {
	fullnode := flag.String("h", "http://51.222.153.212:9334/", "Fullnode Endpoint")
	flag.Parse()
	config.LoadConfig()
	config.LoadParam()
	app := devframework.NewAppService(*fullnode, true)

	txStringMap := map[string]bool{
		"aed63c9b01073bbfb77d5f4257d95f8a54ada2556fc263d10151495c886f3dfe": true,
		"fc77ad70daa4b1efa9775acdced6cbc1cd2f891181ea9a4386acae372aafb4d0": true,
		"8485e2eea87cba03285cb43d3eeebcd1bdd14dba52d22fd8264cd931a0a9585e": true,
		"387b87bea174f5ac483af3ca31667ec4c7ec41409425ceb655310675777e0ae4": true,
		"cc13c7776b6675278f3600380d11c06a5c384f24e4edcf61aa000a9cfd1d6160": true,
		"b1b19af9e794f082f2a266409e41eb3ca6926536d5f069189db091ff9118f1e0": true,
		"2441ee207abfe6c3a12609f0d5bb247254996e97ba5ef3b99c76ad7bc5c28c35": true,
		"2c3e9bbc96d82ca7a3add34055996278b09e4d9c6f24deef74e261d8fd8e94c3": true,
		"cd9346bfad1ab1eb3fc20e0d663e234e2d306e00d0e589596a651899f79075d2": true,
		"876392c508e022030ae8964f3a1bc7e91ebc26d62bf69244a75ca5f2c0a2beec": true,
		"7fdf41355069f7a68fe19b7a8c0ee10ea06dda5edad7391e491237418830d64a": true,
		"10fabc1076aae0f5973b688e11fd1b290feca2a29dab3aaddab0d65278d83951": true,
	}

	//app.OnBeaconBlock(8664, func(blk types.BeaconBlock) {
	//	for sid, states := range blk.Body.ShardState {
	//		fmt.Println("Shard ", sid)
	//		for _, s := range states {
	//			fmt.Println(s.Height, s.Hash.String())
	//			fmt.Println(s.ValidationData, s.PreviousValidationData)
	//		}
	//	}
	//})

	//app.OnShardBlock(4, 1699680, func(blk types.ShardBlock) {
	//
	//})

	//for j := 0; j < 8; j++ {
	//	app.OnShardBlock(j, 1699680, func(blk types.ShardBlock) {
	//		for _, tx := range blk.Body.Transactions {
	//			if tx.GetMetadataType() == metadataCommon.Pdexv3AddOrderRequestMeta {
	//				md := tx.GetMetadata()
	//				req, ok := md.(*metadataPdexv3.AddOrderRequest)
	//				if !ok {
	//					panic(100)
	//				}
	//				tokenToSell := req.TokenToSell.String()
	//				pair := req.PoolPairID
	//				strs := strings.Split(pair, "-")
	//				if tokenToSell != strs[0] && tokenToSell != strs[1] {
	//					fmt.Printf("ShardID:%v tx: %+v , token: %v, pair: %v-%v \n", j, tx.Hash().String(), tokenToSell, strs[0], strs[1])
	//				}
	//			}
	//		}
	//	})
	//}
	wg := sync.WaitGroup{}
	mapCoin := map[int]map[uint64][]metadata.Transaction{} //shardid -> index -> tx
	mapOTAString := map[int]map[uint64]string{}            // shardid -> index >- string
	mapOTATx := map[string][]metadata.Transaction{}        // string -> tx

	type processOTA struct {
		shardID int
		index   uint64
		tx      metadata.Transaction
	}
	processOTACh := make(chan processOTA, 1000)

	wg.Add(8)
	go func() {
		for {
			req := <-processOTACh
			pubkey := app.GetOTACoinByIndices(req.index, req.shardID, "0000000000000000000000000000000000000000000000000000000000000005")
			if pubkey == "" {
				panic(1)
			}
			fmt.Println("process ota", req.index, req.shardID, "0000000000000000000000000000000000000000000000000000000000000005", pubkey)
			if _, ok := mapOTAString[req.shardID]; !ok {
				mapOTAString[req.shardID] = make(map[uint64]string)
			}
			mapOTAString[req.shardID][req.index] = pubkey
			mapOTATx[pubkey] = append(mapOTATx[pubkey], req.tx)
		}
	}()

	var initOTAList []string
	var processBlk = func(blk types.ShardBlock) {
		if blk.GetBeaconHeight() == 2426573 {
			wg.Done()
			return
		}
		if _, ok := mapCoin[blk.GetShardID()]; !ok {
			mapCoin[blk.GetShardID()] = make(map[uint64][]metadata.Transaction)
		}
		for _, tx := range blk.Body.Transactions {
			//init ota
			if tx.GetMetadataType() == metadataCommon.Pdexv3WithdrawOrderResponseMeta {
				md := tx.GetMetadata()
				req, ok := md.(*metadataPdexv3.WithdrawOrderResponse)
				if !ok {
					panic(100)
				}
				if _, ok := txStringMap[req.RequestTxID.String()]; ok {
					if _, ok := tx.(*transaction.TxTokenVersion2); ok {
						if len(tx.(*transaction.TxTokenVersion2).GetTxNormal().GetProof().GetOutputCoins()) == 0 {
							continue
						}
						for _, coin := range tx.(*transaction.TxTokenVersion2).GetTxNormal().GetProof().GetOutputCoins() {
							publicKey := base58.Base58Check{}.Encode(coin.(*privacy.CoinV2).GetPublicKey().ToBytesS(), common.ZeroByte)
							log.Printf("init tx %v outcoin %v", tx.Hash().String(), publicKey)
							initOTAList = append(initOTAList, publicKey)
						}
					}
				}
			}

			//build map
			txSig2 := new(transaction.TxSigPubKeyVer2)
			if _, ok := tx.(*transaction.TxTokenVersion2); ok {
				if len(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey) == 0 {
					continue
				}
				if err := txSig2.SetBytes(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey); err != nil {
					panic(err)
				}
				for _, i := range txSig2.Indexes {
					for _, j := range i {
						mapCoin[blk.GetShardID()][j.Uint64()] = append(mapCoin[blk.GetShardID()][j.Uint64()], tx)
						processOTACh <- processOTA{blk.GetShardID(), j.Uint64(), tx}
					}
				}
			}
		}
	}

	//from beacon hash: f1f11862e3a63a6afa08fc85a829d9d277bab6dfd726b6338e6b30782b5adb31
	app.OnShardBlock(0, 2163378, processBlk)
	app.OnShardBlock(1, 2167083, processBlk)
	app.OnShardBlock(2, 2164835, processBlk)
	app.OnShardBlock(3, 2161584, processBlk)
	app.OnShardBlock(4, 2161694, processBlk)
	app.OnShardBlock(5, 2158484, processBlk)
	app.OnShardBlock(6, 2162553, processBlk)
	app.OnShardBlock(7, 2161572, processBlk)

	wg.Wait()

	//process
	//mapCoin [sharid][index][]tx
	//mapOTA [sharid][index][string]
	//mapOTAString [string][]tx

	currentOTAList := initOTAList
	fmt.Println("process", initOTAList)
	for level := 0; len(currentOTAList) > 0; level++ {
		fmt.Println("\nLevel ", level)
		nextOTAList := []string{}

		for _, ota := range currentOTAList {
			txs := mapOTATx[ota]
			if len(txs) == 0 {
				fmt.Println("  ", ota, " -> none yet")
			}
			for _, temp := range txs {
				tx, ok := temp.(*transaction.TxTokenVersion2)
				if !ok {
					fmt.Println("  skip prv tx", temp.Hash().String())
					continue
				}
				proof, ok := tx.GetTxNormal().GetProof().(*privacy.ProofV2)
				if !ok {
					fmt.Println("  skip unknown proof type in tx", temp.Hash().String())
					continue
				}

				for _, c := range proof.GetOutputCoins() {
					publicKey := base58.Base58Check{}.Encode(c.(*privacy.CoinV2).GetPublicKey().ToBytesS(), common.ZeroByte)
					fmt.Println("  ", ota, " -> ", publicKey)
					nextOTAList = append(nextOTAList, publicKey)
				}
			}
		}
		currentOTAList = nextOTAList
	}
	select {}
}
