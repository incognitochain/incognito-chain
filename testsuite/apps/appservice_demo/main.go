package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver1"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var mapOTAString = make(map[string]map[int]map[uint64]string) //txtype -> sid -> index -> pubkey

const MONGODB = "mongodb://127.0.0.1:38118"
const INCOGNITO_FULLNODE = "http://127.0.0.1:9334/"

func processOTA(app *devframework.AppService, otaFD *os.File, txType string, sid int, index uint64, tx metadata.Transaction) {
	if _, ok := mapOTAString[txType]; !ok {
		mapOTAString[txType] = make(map[int]map[uint64]string)
	}
	if _, ok := mapOTAString[txType][sid]; !ok {
		mapOTAString[txType][sid] = make(map[uint64]string)
	}
	if _, ok := mapOTAString[txType][sid][index]; ok {
		return
	}
	pubkey := ""
	if txType == "tx" {
		pubkey = app.GetOTACoinByIndices(index, sid, "0000000000000000000000000000000000000000000000000000000000000004")
	} else {
		pubkey = app.GetOTACoinByIndices(index, sid, "0000000000000000000000000000000000000000000000000000000000000005")
	}
	if pubkey == "" {
		panic(1)
	}

	//fmt.Println("process ota", req.index, req.shardID, "0000000000000000000000000000000000000000000000000000000000000005", pubkey)
	mapOTAString[txType][sid][index] = pubkey
	otaFD.Write([]byte(fmt.Sprintf("%v %v %v %v\n", txType, sid, index, pubkey)))
}

var convertTokenCache = map[string]string{}  //tx -> convertedTokenID
var shieldRequestCache = map[string]string{} //tx -> tokenID
func GetDetailInfo(tx metadata.Transaction) DetailInfo {
	info := DetailInfo{}
	switch tx.(type) {
	case (*transaction.TxVersion2):
		info.TokenID = common.PRVCoinID.String()
	case (*transaction.TxTokenVersion2):
		info.TokenID = common.ConfidentialAssetID.String()
	}
	//TODO: update WithdrawLiquidityRequest -> AddLiquidityRequest
	switch tx.GetMetadataType() {
	case 1:
	case 244: //InitTokenResponse
		info.MetaName = "InitTokenRequestMeta"
		//md, _ := tx.GetMetadata().(*metadata.InitTokenRequest)
	case 245:
		info.MetaName = "InitTokenResponseMeta"
	case 24:
		info.MetaName = "IssuingRequestMeta"
		md, _ := tx.GetMetadata().(*metadata.IssuingRequest)
		info.TokenID = md.TokenID.String()
		info.Amount = md.DepositedAmount
	case 25: //ContractingRequest
		info.MetaName = "IssuingResponseMeta"
		md, _ := tx.GetMetadata().(*metadata.IssuingResponse)
		info.PreviousLinkTx = md.RequestedTxID.String()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	case 26: //ContractingRequestMeta
		info.MetaName = "ContractingRequestMeta"
		//md, _ := tokenTx.GetMetadata().(*metadata.ContractingRequest)
	case 44:
		info.MetaName = "WithDrawRewardRequestMeta"
	case 45:
		info.MetaName = "WithDrawRewardResponseMeta"
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	/* PDEX */
	case 93:
		info.MetaName = "PDEWithdrawalRequestMeta"
	case 94:
		info.MetaName = "PDEWithdrawalResponseMeta"
	case 207:
		info.MetaName = "PDEFeeWithdrawalRequestMeta"
	case 208:
		info.MetaName = "PDEFeeWithdrawalResponseMeta"
	case 281: //AddLiquidityRequest
		info.MetaName = "AddLiquidityRequest"
		md, _ := tx.GetMetadata().(*pdexv3.AddLiquidityRequest)
		info.TokenID = md.TokenID()
		info.Amount = md.TokenAmount()
	case 282: //AddLiquidityResponse
		info.MetaName = "AddLiquidityResponse"
		md, _ := tx.GetMetadata().(*pdexv3.AddLiquidityResponse)
		info.PreviousLinkTx = md.TxReqID()
	case 283: //AddLiquidityResponse
		info.MetaName = "WithdrawLiquidityRequest"
		//md, _ := tx.GetMetadata().(*pdexv3.WithdrawLiquidityRequest)
	case 284: //WithdrawLiquidityResponse
		info.MetaName = "WithdrawLiquidityResponse"
		md, _ := tx.GetMetadata().(*pdexv3.WithdrawLiquidityResponse)
		info.PreviousLinkTx = md.TxReqID()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	case 285: //TradeRequest
		info.MetaName = "TradeRequest"
		md, _ := tx.GetMetadata().(*pdexv3.TradeRequest)
		info.Amount = md.SellAmount
		info.TokenID = md.TokenToSell.String()
	case 286: //TradeResponse
		info.MetaName = "TradeResponse"
		md, _ := tx.GetMetadata().(*pdexv3.TradeResponse)
		info.PreviousLinkTx = md.RequestTxID.String()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	case 287: //AddOrderRequest
		info.MetaName = "AddOrderRequest"
		md, _ := tx.GetMetadata().(*pdexv3.AddOrderRequest)
		info.TokenID = md.TokenToSell.String()
		info.Amount = md.SellAmount
	case 288: //AddOrderRequest
		info.MetaName = "AddOrderResponse"
		md, _ := tx.GetMetadata().(*pdexv3.AddOrderResponse)
		info.PreviousLinkTx = md.RequestTxID.String()

	case 289: //WithdrawOrderRequest
		info.MetaName = "WithdrawOrderRequest"
		md, _ := tx.GetMetadata().(*pdexv3.WithdrawOrderRequest)
		info.PreviousLinkTx = md.OrderID
	case 290: //WithdrawOrderResponse
		info.MetaName = "WithdrawOrderResponse"
		md, _ := tx.GetMetadata().(*pdexv3.WithdrawOrderResponse)
		info.PreviousLinkTx = md.RequestTxID.String()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	case 291: //UserMintNftRequest
		info.MetaName = "UserMintNftRequest"
		//md, _ := tokenTx.GetMetadata().(*pdexv3.WithdrawOrderResponse)
	case 292: //MintNftResponse
		info.MetaName = "UserMintNftResponse"
		md, _ := tx.GetMetadata().(*pdexv3.UserMintNftResponse)
		info.PreviousLinkTx = md.TxReqID()
	case 294: //MintNftResponse
		info.MetaName = "MintNftResponse"
	case 299: //WithdrawalLPFeeRequest
		info.MetaName = "WithdrawalLPFeeRequest"
		//md, _ := tx.GetMetadata().(*pdexv3.WithdrawalLPFeeRequest)
	case 300: //WithdrawalLPFeeResponse
		info.MetaName = "WithdrawalLPFeeResponses"
		md, _ := tx.GetMetadata().(*pdexv3.WithdrawalLPFeeResponse)
		info.PreviousLinkTx = md.ReqTxID.String()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	/* Unify */
	case 341: //ConvertTokenToUnifiedTokenRequest
		info.MetaName = "ConvertTokenToUnifiedTokenRequest"
		md, _ := tx.GetMetadata().(*bridge.ConvertTokenToUnifiedTokenRequest)
		info.TokenID = md.TokenID.String()
		info.Amount = md.Amount

	case 342: //ConvertTokenToUnifiedTokenResponse
		info.MetaName = "ConvertTokenToUnifiedTokenResponse"
		md, _ := tx.GetMetadata().(*bridge.ConvertTokenToUnifiedTokenResponse)
		info.PreviousLinkTx = md.TxReqID.String()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount

	/* Portal BTC */
	case 260: //PortalShieldingRequest
		info.MetaName = "PortalShieldingRequest"
		//md, _ := tx.GetMetadata().(*metadata.PortalShieldingRequest)
	case 261: //PotalShieldingResponse
		md, _ := tx.GetMetadata().(*metadata.PortalShieldingResponse)
		info.MetaName = "PotalShieldingResponse"
		info.PreviousLinkTx = md.ReqTxID.String()
		info.TokenID = md.IncTokenID
		info.Amount = md.MintingAmount
	case 262:
		md, _ := tx.GetMetadata().(*metadata.PortalUnshieldRequest)
		info.MetaName = "PortalV4UnshieldingRequestMeta"
		info.TokenID = md.TokenID
		info.Amount = md.UnshieldAmount
	case 263:
		md, _ := tx.GetMetadata().(*metadata.PortalUnshieldResponse)
		info.MetaName = "PortalV4UnshieldingResponseMeta"
		info.Amount = md.UnshieldAmount
		info.TokenID = md.IncTokenID
		info.PreviousLinkTx = md.ReqTxID.String()
	case 265:
		info.MetaName = "PortalV4FeeReplacementRequestMeta"
	//* Shield */
	case 335:
		info.MetaName = "IssuingNearRequestMeta"
	case 80, 250, 272, 270, 327, 331, 351, 354:
		info.MetaName = "IssuingETHRequestMeta"
	case 81, 251, 271, 273, 328, 332, 352, 355, 336:
		info.MetaName = "IssuingEVMResponse"
		md, _ := tx.GetMetadata().(*bridge.IssuingEVMResponse)
		info.PreviousLinkTx = md.RequestedTxID.String()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	case 343:
		info.MetaName = "ShieldRequest"
		//md, _ := tx.GetMetadata().(*bridge.ShieldRequest)
	case 344: //ShieldResponse
		info.MetaName = "ShieldResponse"
		md, _ := tx.GetMetadata().(*bridge.ShieldResponse)
		info.Amount = md.ShieldAmount
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
		info.PreviousLinkTx = md.RequestedTxID.String()
	/* UnShield */
	case 240, 242, 252, 274, 275, 326, 329, 330, 333, 334, 356, 337, 353, 357, 358: //token unshield
		info.MetaName = "BurningRequest"
		md, _ := tx.GetMetadata().(*bridge.BurningRequest)
		info.TokenID = md.TokenID.String()
		info.Amount = md.BurningAmount
	case 345: // unify unshield
		info.MetaName = "UnshieldRequest"
		md, _ := tx.GetMetadata().(*bridge.UnshieldRequest)
		info.TokenID = md.UnifiedTokenID.String()
		for _, req := range md.Data {
			info.Amount += req.BurningAmount
		}
	case 346: // unify unshield
		info.MetaName = "BurningUnifiedTokenResponseMeta"
		md, _ := tx.GetMetadata().(*bridge.UnshieldResponse)
		info.PreviousLinkTx = md.RequestedTxID.String()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	case 350: // re-shield
		md, _ := tx.GetMetadata().(*bridge.IssuingReshieldResponse)
		info.MetaName = "IssuingReshieldResponse"
		info.PreviousLinkTx = md.RequestedTxID.String()
		_, _, amount, tokenID := tx.GetTransferData()
		info.TokenID = tokenID.String()
		info.Amount = amount
	case 348:
		info.MetaName = "BurnForCallRequestMeta"
		md, _ := tx.GetMetadata().(*bridge.BurnForCallRequest)
		info.TokenID = md.BurnTokenID.String()
		for _, req := range md.Data {
			info.Amount += req.BurningAmount
		}
	case 266:
		info.MetaName = "PortalV4SubmitConfirmedTxMeta"
	case 63:
		info.MetaName = "StakeShard"
	case 210:
		info.MetaName = "UnStakingShard"
	case 41:
		info.MetaName = "ReturnStakingMeta"
		//TODO: which committee public key
	case 201:
		info.MetaName = "RelayingBTCHeaderMeta"
	case 200:
		info.MetaName = "RelayingBNBHeaderMeta"
	case 280:
		info.MetaName = "Pdexv3ModifyParamsMeta"
	case 127:
		info.MetaName = "StopAutoStakingMeta"
	default:
		fmt.Println("cannot find", tx.GetMetadataType())
	}
	return info
}

func main() {
	fmt.Println("start...")
	//str := "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"
	//rawTxBytes, _, err := base58.Base58Check{}.Decode(str)
	//if err != nil {
	//	fmt.Println("Send Raw Transaction Error: %+v", err)
	//}
	//
	//fmt.Println(string(rawTxBytes))
	//select {}
	statDB, err := NewStatDB(MONGODB, "netmonitor", "stat")
	coinDB, err := NewCoinDB(MONGODB, "netmonitor", "coin")

	config.LoadConfig()
	config.LoadParam()
	otaFD, _ := os.OpenFile("ota", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	scanner := bufio.NewScanner(otaFD)
	const maxCapacity int = 10 * 1024 * 1024 * 1024 // your required line length
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	totalPk := 0
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		res := strings.Split(scanner.Text(), " ")
		txType := res[0]
		sid, _ := strconv.Atoi(res[1])
		index, _ := strconv.Atoi(res[2])
		if _, ok := mapOTAString[txType]; !ok {
			mapOTAString[txType] = make(map[int]map[uint64]string)
		}
		if _, ok := mapOTAString[txType][sid]; !ok {
			mapOTAString[txType][sid] = make(map[uint64]string)
		}
		mapOTAString[txType][sid][uint64(index)] = res[3]
		totalPk++
	}
	fmt.Println("all coin:", totalPk)

	if err != nil {
		panic(err)
	}
	app := devframework.NewAppService(INCOGNITO_FULLNODE, true)

	latestCheckpoint := map[int]uint64{
		0: 1698080,
		1: 1698080,
		2: 1698080,
		3: 1698080,
		4: 1698080,
		5: 1698080,
		6: 1698080,
		7: 1698080,
	}

	for i := 0; i < config.Param().ActiveShards; i++ {
		fromBlock := statDB.lastBlock(i)
		if fromBlock == 1 {
			fromBlock = latestCheckpoint[i]
		}
		//i := 7
		//fromBlock := latestCheckpoint[i]
		app.OnShardBlock(i, uint64(int64(fromBlock)), func(shardBlk types.ShardBlock) {
			if shardBlk.GetHeight()%10000 == 0 {
				fmt.Println(shardBlk.GetShardID(), time.Unix(shardBlk.GetProposeTime(), 0), shardBlk.GetHeight(), shardBlk.Hash().String(), len(shardBlk.Body.Transactions))
			}
			if len(shardBlk.Body.Transactions) == 0 {
				return
			}
			//fmt.Println(shardBlk.GetShardID(), time.Unix(shardBlk.GetProposeTime(), 0), shardBlk.GetHeight(), shardBlk.Hash().String(), len(shardBlk.Body.Transactions))

			shardID := shardBlk.GetShardID()
			blockHash := shardBlk.Hash().String()
			blkHeight := shardBlk.GetHeight()
			blockTime := time.Unix(shardBlk.GetProposeTime(), 0)
			epoch := shardBlk.GetCurrentEpoch()
			for _, tx := range shardBlk.Body.Transactions {
				outCoins := []coin.Coin{}
				inCoins := [][]string{}
				txSig2 := new(transaction.TxSigPubKeyVer2)
				txtype := "tx"
				metadataType := tx.GetMetadataType()
				metadataStr := ""
				if metadataType != 1 {
					bytes, _ := json.MarshalIndent(tx.GetMetadata(), "", "\t")
					metadataStr = string(bytes)
				}
				//fmt.Println(metadataType, metadataStr)
				whiteListMap := []int{25, 81, 94, 261, 263, 282, 288, 284, 292, 245, 45, 41, 251, 290, 286, 300, 294, 342, 251, 271, 273, 328, 332, 352, 355, 336, 344, 346, 350}
				switch tx.(type) {
				case (*transaction.TxVersion2):
					for _, coin := range tx.GetProof().GetOutputCoins() {
						outCoins = append(outCoins, coin)
					}
					if err := txSig2.SetBytes(tx.(*transaction.TxVersion2).SigPubKey); err != nil {
						if common.IndexOfInt(metadataType, whiteListMap) != -1 {
							if len(tx.(*transaction.TxVersion2).SigPubKey) != 0 && len(tx.(*transaction.TxVersion2).SigPubKey) != 32 { // ma len != 0
								fmt.Println(err, len(tx.(*transaction.TxVersion2).SigPubKey), tx.GetMetadataType(), tx.Hash().String())
								panic(1)
							}
						} else {
							if len(tx.(*transaction.TxVersion2).SigPubKey) == 0 {
								fmt.Println(tx.GetMetadataType(), tx.Hash().String())
								panic(1)
							} else if len(tx.(*transaction.TxVersion2).SigPubKey) != 32 {
								fmt.Println(tx.GetMetadataType(), tx.Hash().String())
								panic(1)
							}
						}
					}
				case (*transaction.TxTokenVersion2):
					txtype = "txToken"
					for _, coin := range tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).GetProof().GetOutputCoins() {
						outCoins = append(outCoins, coin)
					}
					if err := txSig2.SetBytes(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey); err != nil {
						if common.IndexOfInt(metadataType, whiteListMap) != -1 { //neu trong list
							if len(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey) != 0 && len(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey) != 32 { // ma len != 0
								fmt.Println(tx.GetMetadataType(), tx.Hash().String())
								panic(1)
							}
						} else {
							if len(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey) == 0 {
								fmt.Println(tx.GetMetadataType(), tx.Hash().String())
								panic(1)
							} else if len(tx.(*transaction.TxTokenVersion2).GetTxNormal().(*transaction.TxVersion2).SigPubKey) != 32 { // ma len != 0
								fmt.Println(tx.GetMetadataType(), tx.Hash().String())
								panic(1)
							}
						}
					}

				case (*tx_ver1.Tx):
				case (*tx_ver1.TxToken):
				default:
					panic(reflect.TypeOf(tx))
				}

				for _, i := range txSig2.Indexes {
					for _, j := range i {
						processOTA(app, otaFD, txtype, shardID, j.Uint64(), tx)
					}
				}

				for _, i := range txSig2.Indexes {
					incoin := []string{}
					for _, j := range i {
						incoin = append(incoin, mapOTAString[txtype][shardID][j.Uint64()])
						if mapOTAString[txtype][shardID][j.Uint64()] == "1y4gnYS1Ns2K7BjQTjgfZ5nTR8JZMkMJ3CTGMj2Pk7CQkSTFgA" {
							continue
						}
						coinDB.UpdateCoinLink(mapOTAString[txtype][shardID][j.Uint64()], tx.Hash().String())
					}
					inCoins = append(inCoins, incoin)

				}

				details := GetDetailInfo(tx)
				info := StatInfo{
					BlockHeight:  blkHeight,
					BlockTime:    blockTime,
					BlockHash:    blockHash,
					Incoin:       inCoins,
					Outcoin:      outCoins,
					MetadataType: metadataType,
					Metadata:     metadataStr,
					Tx:           tx.Hash().String(),
					ShardID:      shardID,
					Detail:       GetDetailInfo(tx),
					Epoch:        epoch,
				}
				statDB.set(info)
				for _, outcoin := range outCoins {
					publicKey := base58.Base58Check{}.Encode(outcoin.(*privacy.CoinV2).GetPublicKey().ToBytesS(), common.ZeroByte)
					if publicKey == "1y4gnYS1Ns2K7BjQTjgfZ5nTR8JZMkMJ3CTGMj2Pk7CQkSTFgA" {
						continue
					}
					coinDB.set(CoinInfo{
						Pubkey:     publicKey,
						CreateTime: blockTime,
						TokenID:    details.TokenID,
						TokenName:  details.MetaName,
						CreatedTx:  tx.Hash().String(),
						Amount:     outcoin.GetValue(),
					})
				}
			}
		})
	}
	select {}
}
