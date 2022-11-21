package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/transaction"
	"os"
	"strconv"
	"strings"
)

func main() {
	fullnode := flag.String("h", "http://51.222.153.212:9334/", "Fullnode Endpoint")
	flag.Parse()
	config.LoadConfig()
	config.LoadParam()
	app := devframework.NewAppService(*fullnode, true)

	lastProcess := map[int]uint64{
		0: 2163378,
		1: 2167083,
		2: 2164835,
		3: 2442777,
		4: 2161694,
		5: 2158484,
		6: 2162553,
		7: 2161572,
	}
	otaFD, _ := os.OpenFile("ota", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	//read otaFD
	scanner := bufio.NewScanner(otaFD)
	const maxCapacity int = 10 * 1024 * 1024 * 1024 // your required line length
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	mapOTAString := map[int]map[uint64]string{}              // shardid -> index >- string
	mapOTATx := map[string]map[string]metadata.Transaction{} // pubkey string -> tx hash string ->  tx
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		res := strings.Split(scanner.Text(), " ")
		sid, _ := strconv.Atoi(res[0])
		index, _ := strconv.Atoi(res[1])
		if mapOTAString[sid] == nil {
			mapOTAString[sid] = make(map[uint64]string)
		}
		mapOTAString[sid][uint64(index)] = res[2]
	}

	processOTA := func(sid int, index uint64, tx metadata.Transaction) {
		if _, ok := mapOTAString[sid]; !ok {
			mapOTAString[sid] = make(map[uint64]string)
		}
		if pubkey, ok := mapOTAString[sid][index]; ok {
			if mapOTATx[pubkey] == nil {
				mapOTATx[pubkey] = make(map[string]metadata.Transaction)
			}
			mapOTATx[pubkey][tx.String()] = tx
		}

		pubkey := app.GetOTACoinByIndices(index, sid, "0000000000000000000000000000000000000000000000000000000000000005")
		if pubkey == "" {
			panic(1)
		}
		//fmt.Println("process ota", req.index, req.shardID, "0000000000000000000000000000000000000000000000000000000000000005", pubkey)
		mapOTAString[sid][index] = pubkey
		if mapOTATx[pubkey] == nil {
			mapOTATx[pubkey] = make(map[string]metadata.Transaction)
		}
		mapOTATx[pubkey][tx.String()] = tx
	}
	processShard := func(sid int) {
		app.OnShardBlock(sid, lastProcess[sid], func(block types.ShardBlock) {
			fmt.Println("Block:", block.GetHeight(), block.Hash().String())
			for _, tx := range block.Body.Transactions {
				if tokenTx, ok := tx.(*transaction.TxTokenVersion2); ok {
					fmt.Println("Tx:", tx.Hash().String(), " - Type", tx.GetMetadataType())
					outCoins := []
					for _, coin := range tx.(*transaction.TxTokenVersion2).GetTxNormal().GetProof().GetOutputCoins() {
						publicKey := base58.Base58Check{}.Encode(coin.(*privacy.CoinV2).GetPublicKey().ToBytesS(), common.ZeroByte)
					}

					txSig2 := new(transaction.TxSigPubKeyVer2)
					if err := txSig2.SetBytes(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey); err != nil {
						fmt.Println("set byte error", tx.GetType(), tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey)
						continue
					}
					for _, i := range txSig2.Indexes {
						for _, j := range i {
							//mapCoin[blk.GetShardID()][j.Uint64()] = append(mapCoin[blk.GetShardID()][j.Uint64()], tx)
							processOTACh <- processOTA(sid, j.Uint64(), tx)
						}
					}

					if tokenTx.GetMetadata() != nil {
						switch tokenTx.GetMetadata().GetType() {
						case 244, 245: //InitTokenResponse
						case 25: //ContractingRequest
							//md, _ := tokenTx.GetMetadata().(*metadata.IssuingResponse)
						case 26: //ContractingRequest
							//md, _ := tokenTx.GetMetadata().(*metadata.ContractingRequest)
						/* PDEX */
						case 281: //AddLiquidityRequest
							//md, _ := tokenTx.GetMetadata().(*pdexv3.AddLiquidityRequest)
						case 282: //AddLiquidityResponse
						//md, _ := tokenTx.GetMetadata().(*pdexv3.AddLiquidityResponse)
						case 283: //AddLiquidityResponse
						//md, _ := tokenTx.GetMetadata().(*pdexv3.WithdrawLiquidityRequest)
						case 284: //WithdrawLiquidityResponse
							//md, _ := tokenTx.GetMetadata().(*pdexv3.WithdrawLiquidityResponse)
						case 285: //TradeRequest
							//md, _ := tokenTx.GetMetadata().(*pdexv3.TradeRequest)
						case 286: //TradeResponse
							//md, _ := tokenTx.GetMetadata().(*pdexv3.TradeResponse)
						case 287: //AddOrderRequest
							//md ,_ := tokenTx.GetMetadata().(*pdexv3.AddOrderRequest)
						case 288: //AddOrderRequest
							//md ,_ := tokenTx.GetMetadata().(*pdexv3.AddOrderResponse)
						case 289: //WithdrawOrderRequest
							//md ,_ := tokenTx.GetMetadata().(*pdexv3.WithdrawOrderRequest)
						case 290: //WithdrawOrderResponse
						//md, _ := tokenTx.GetMetadata().(*pdexv3.WithdrawOrderResponse)
						case 292: //MintNftResponse
							//md, _ := tokenTx.GetMetadata().(*pdexv3.UserMintNftResponse)
						case 294: //MintNftResponse
							//md, _ := tokenTx.GetMetadata().(*pdexv3.MintNftResponse)
						case 299: //WithdrawalLPFeeRequest
							//md, _ := tokenTx.GetMetadata().(*pdexv3.WithdrawalLPFeeRequest)
						case 300: //WithdrawalLPFeeResponse
							//md, _ := tokenTx.GetMetadata().(*pdexv3.WithdrawalLPFeeResponse)

						/* Unify */
						case 341: //ConvertTokenToUnifiedTokenRequest
							//md, _ := tokenTx.GetMetadata().(*bridge.ConvertTokenToUnifiedTokenRequest)
						case 342: //ConvertTokenToUnifiedTokenResponse
							//md, _ := tokenTx.GetMetadata().(*bridge.ConvertTokenToUnifiedTokenResponse)
						/* Portal BTC */
						case 260: //PortalShieldingRequest
							//md, _ := tokenTx.GetMetadata().(*metadata.PortalShieldingRequest)
						case 261: //PortalShieldingResponse
							//md, _ := tokenTx.GetMetadata().(*metadata.PortalShieldingResponse)
						case 262:
							//md, _ := tokenTx.GetMetadata().(*metadata.PortalUnshieldRequest)
						//case 263:
						//	//md, _ := tokenTx.GetMetadata().(*metadata.PortalUnshieldResponse)
						//* Shield */
						case 80, 24, 250, 270, 272, 327, 331, 343, 351, 354, 335:
							//md, _ := tokenTx.GetMetadata().(*bridge.ShieldRequest)
						case 344: //IssuingUnifiedTokenResponseMeta
							//md, _ := tokenTx.GetMetadata().(*bridge.ShieldResponse)
						case 81, 251, 271, 273, 328, 332, 352, 355, 336:
							//md, _ := tokenTx.GetMetadata().(*bridge.IssuingEVMResponse)
						/* UnShield */
						case 240, 242, 252, 274, 275, 326, 329, 330, 333, 334, 356, 337, 353, 357, 358: //token unshield
						//md, _ := tokenTx.GetMetadata().(*bridge.BurningRequest)
						case 345: // unify unshield
						//md, _ := tokenTx.GetMetadata().(*bridge.UnshieldRequest)
						case 350: // re-shield
						//md, _ := tokenTx.GetMetadata().(*bridge.IssuingReshieldResponse)
						case 348:
							//md, _ := tokenTx.GetMetadata().(*bridge.BurnForCallRequest)

						default:
							fmt.Println("cannot find", tokenTx.GetMetadata().GetType())
						}
					} else {
						//TODO: this is transfer token
					}
				}
			}
		})
	}
	processShard(3)

	//type processOTA struct {
	//	shardID int
	//	index   uint64
	//	tx      metadata.Transaction
	//}
	//processOTACh := make(chan processOTA, 1000)
	//
	//txStringMap := map[string]bool{
	//	"d2496118c3f73bed6ee33cf38cea570203abee9116c299909c3881e3ebb2aeb6": true,
	//}
	//otaFD, _ := os.OpenFile("ota", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//blockFD, _ := os.OpenFile("block", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//
	////read otaFD
	//scanner := bufio.NewScanner(otaFD)
	//const maxCapacity int = 10 * 1024 * 1024 * 1024 // your required line length
	//buf := make([]byte, maxCapacity)
	//scanner.Buffer(buf, maxCapacity)
	//wg := sync.WaitGroup{}
	////mapCoin := map[int]map[uint64][]metadata.Transaction{} //shardid -> index -> tx
	//mapOTAString := map[int]map[uint64]string{}              // shardid -> index >- string
	//mapOTATx := map[string]map[string]metadata.Transaction{} // pubkey string -> tx hash string ->  tx
	//for scanner.Scan() {
	//	if scanner.Text() == "" {
	//		continue
	//	}
	//	res := strings.Split(scanner.Text(), " ")
	//	sid, _ := strconv.Atoi(res[0])
	//	index, _ := strconv.Atoi(res[1])
	//	if mapOTAString[sid] == nil {
	//		mapOTAString[sid] = make(map[uint64]string)
	//	}
	//	mapOTAString[sid][uint64(index)] = res[2]
	//}
	//
	//var initOTAList = []string{""}
	//var requestWithdraw = map[string]string{}
	//var doneProcess = map[int]bool{}
	//persist := false
	//var processBlk = func(blk types.ShardBlock) {
	//
	//	if persist {
	//		raw, _ := json.Marshal(blk)
	//		blockFD.WriteString(string(raw) + "\n")
	//	}
	//	if blk.GetHeight()%10000 == 0 {
	//		fmt.Println("process ", blk.GetShardID(), blk.GetHeight())
	//	}
	//
	//	if blk.GetBeaconHeight() >= 2431258 {
	//		if doneProcess[blk.GetShardID()] {
	//			return
	//		}
	//		doneProcess[blk.GetShardID()] = true
	//		wg.Done()
	//		return
	//	}
	//
	//	for _, tx := range blk.Body.Transactions {
	//		//init ota
	//		if tx.GetMetadataType() == metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta {
	//			md := tx.GetMetadata()
	//			req, ok := md.(*metadataBridge.ConvertTokenToUnifiedTokenResponse)
	//			if !ok {
	//				panic(100)
	//			}
	//			if _, ok := txStringMap[req.TxReqID.String()]; ok {
	//				if _, ok := tx.(*transaction.TxTokenVersion2); ok {
	//					if len(tx.(*transaction.TxTokenVersion2).GetTxNormal().GetProof().GetOutputCoins()) == 0 {
	//						continue
	//					}
	//					for _, coin := range tx.(*transaction.TxTokenVersion2).GetTxNormal().GetProof().GetOutputCoins() {
	//						publicKey := base58.Base58Check{}.Encode(coin.(*privacy.CoinV2).GetPublicKey().ToBytesS(), common.ZeroByte)
	//						fmt.Printf("init 2 tx %v outcoin %v from convert-unified %v amount %v", tx.Hash().String(), publicKey, req.TxReqID.String(), req.ConvertAmount)
	//						initOTAList = append(initOTAList, publicKey)
	//					}
	//				}
	//			}
	//		}
	//
	//		if tx.GetMetadataType() == metadataCommon.Pdexv3WithdrawOrderRequestMeta {
	//			md := tx.GetMetadata()
	//			req, ok := md.(*metadataPdexv3.WithdrawOrderRequest)
	//			if !ok {
	//				panic(100)
	//			}
	//			if _, ok := txStringMap[req.OrderID]; ok {
	//				requestWithdraw[tx.Hash().String()] = req.OrderID
	//			}
	//		}
	//
	//		if tx.GetMetadataType() == metadataCommon.Pdexv3WithdrawOrderResponseMeta {
	//			md := tx.GetMetadata()
	//			req, ok := md.(*metadataPdexv3.WithdrawOrderResponse)
	//			if !ok {
	//				panic(100)
	//			}
	//			_, ok2 := txStringMap[req.RequestTxID.String()]
	//			if _, ok := requestWithdraw[req.RequestTxID.String()]; ok || ok2 {
	//				if _, ok := tx.(*transaction.TxTokenVersion2); ok {
	//					if len(tx.(*transaction.TxTokenVersion2).GetTxNormal().GetProof().GetOutputCoins()) == 0 {
	//						continue
	//					}
	//					for _, coin := range tx.(*transaction.TxTokenVersion2).GetTxNormal().GetProof().GetOutputCoins() {
	//						publicKey := base58.Base58Check{}.Encode(coin.(*privacy.CoinV2).GetPublicKey().ToBytesS(), common.ZeroByte)
	//						fmt.Printf("init 1 tx %v outcoin %v from requestWithdraw %v addOrderID %v", tx.Hash().String(), publicKey, req.RequestTxID.String(), requestWithdraw[req.RequestTxID.String()])
	//						initOTAList = append(initOTAList, publicKey)
	//					}
	//				}
	//			}
	//		}
	//
	//		//build map
	//		txSig2 := new(transaction.TxSigPubKeyVer2)
	//		if _, ok := tx.(*transaction.TxTokenVersion2); ok {
	//			if len(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey) == 0 {
	//				continue
	//			}
	//			if err := txSig2.SetBytes(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey); err != nil {
	//				fmt.Println("set byte error", tx.GetType(), tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey)
	//				continue
	//			}
	//			for _, i := range txSig2.Indexes {
	//				for _, j := range i {
	//					//mapCoin[blk.GetShardID()][j.Uint64()] = append(mapCoin[blk.GetShardID()][j.Uint64()], tx)
	//					processOTACh <- processOTA{blk.GetShardID(), j.Uint64(), tx}
	//				}
	//			}
	//		}
	//	}
	//}
	//
	//go func() {
	//	for {
	//		req := <-processOTACh
	//		if _, ok := mapOTAString[req.shardID]; !ok {
	//			mapOTAString[req.shardID] = make(map[uint64]string)
	//		}
	//		if pubkey, ok := mapOTAString[req.shardID][req.index]; ok {
	//			if mapOTATx[pubkey] == nil {
	//				mapOTATx[pubkey] = make(map[string]metadata.Transaction)
	//			}
	//			mapOTATx[pubkey][req.tx.String()] = req.tx
	//			continue
	//		}
	//
	//		pubkey := app.GetOTACoinByIndices(req.index, req.shardID, "0000000000000000000000000000000000000000000000000000000000000005")
	//		if pubkey == "" {
	//			panic(1)
	//		}
	//		//fmt.Println("process ota", req.index, req.shardID, "0000000000000000000000000000000000000000000000000000000000000005", pubkey)
	//		mapOTAString[req.shardID][req.index] = pubkey
	//		if mapOTATx[pubkey] == nil {
	//			mapOTATx[pubkey] = make(map[string]metadata.Transaction)
	//		}
	//		mapOTATx[pubkey][req.tx.String()] = req.tx
	//		otaFD.Write([]byte(fmt.Sprintf("%v %v %v\n", req.shardID, req.index, pubkey)))
	//	}
	//}()
	//
	////read block
	//lastProcess := map[int]uint64{
	//	0: 2163378,
	//	1: 2167083,
	//	2: 2164835,
	//	3: 2161584,
	//	4: 2161694,
	//	5: 2158484,
	//	6: 2162553,
	//	7: 2161572,
	//}
	//scanner = bufio.NewScanner(blockFD)
	//buf = make([]byte, maxCapacity)
	//scanner.Buffer(buf, maxCapacity)
	//wg.Add(8)
	//
	//sem := semaphore.NewWeighted(1)
	//mu := sync.Mutex{}
	//for scanner.Scan() {
	//	if scanner.Text() == "" {
	//		continue
	//	}
	//
	//	if len(doneProcess) == 8 {
	//		fmt.Println("finish scan")
	//		break
	//	}
	//	sem.Acquire(context.Background(), 1)
	//	go func(data string) {
	//		blk := types.ShardBlock{}
	//		err := json.Unmarshal([]byte(data), &blk)
	//		if err != nil {
	//			fmt.Println(data)
	//			panic(err)
	//		}
	//		sid := blk.GetShardID()
	//		height := blk.GetHeight()
	//		mu.Lock()
	//		lastProcess[sid] = height
	//		mu.Unlock()
	//		processBlk(blk)
	//		sem.Release(1)
	//	}(scanner.Text())
	//
	//}
	//
	//fmt.Println(lastProcess, len(doneProcess))
	//persist = true
	////from beacon hash: f1f11862e3a63a6afa08fc85a829d9d277bab6dfd726b6338e6b30782b5adb31
	//app.OnShardBlock(0, lastProcess[0], processBlk)
	//app.OnShardBlock(1, lastProcess[1], processBlk)
	//app.OnShardBlock(2, lastProcess[2], processBlk)
	//app.OnShardBlock(3, lastProcess[3], processBlk)
	//app.OnShardBlock(4, lastProcess[4], processBlk)
	//app.OnShardBlock(5, lastProcess[5], processBlk)
	//app.OnShardBlock(6, lastProcess[6], processBlk)
	//app.OnShardBlock(7, lastProcess[7], processBlk)
	//
	//wg.Wait()
	//fmt.Println()
	//for len(processOTACh) != 0 {
	//	fmt.Println("still process processOTACh...")
	//	time.Sleep(time.Second * 10)
	//}
	////process
	////mapCoin [sharid][index][]tx
	////mapOTA [sharid][index][string]
	////mapOTAString [string][]tx
	//
	//currentOTAList := initOTAList
	//fmt.Println("process", initOTAList)
	//txStringMapConvert := make(map[string]bool)
	//for level := 0; len(currentOTAList) > 0; level++ {
	//	fmt.Println("\nLevel ", level)
	//	nextOTAList := []string{}
	//
	//	for _, ota := range currentOTAList {
	//		txs := mapOTATx[ota]
	//		if len(txs) == 0 {
	//			fmt.Println("  ", ota, " -> none yet")
	//		}
	//		for _, temp := range txs {
	//			tx, ok := temp.(*transaction.TxTokenVersion2)
	//			if !ok {
	//				fmt.Println("  skip prv tx", temp.Hash().String())
	//				continue
	//			}
	//			proof, ok := tx.GetTxNormal().GetProof().(*privacy.ProofV2)
	//			if !ok {
	//				fmt.Println("  skip unknown proof type in tx", temp.Hash().String())
	//				continue
	//			}
	//
	//			for _, c := range proof.GetOutputCoins() {
	//				publicKey := base58.Base58Check{}.Encode(c.(*privacy.CoinV2).GetPublicKey().ToBytesS(), common.ZeroByte)
	//				fmt.Println("  ", ota, " -> ", publicKey, " ", tx.Hash().String())
	//				if md := tx.GetMetadata(); md != nil {
	//					if md.GetType() == metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta {
	//						txStringMapConvert[tx.Hash().String()] = true
	//					}
	//				}
	//				nextOTAList = append(nextOTAList, publicKey)
	//			}
	//		}
	//	}
	//	currentOTAList = nextOTAList
	//}
	//fmt.Println(txStringMapConvert)
	select {}
}
