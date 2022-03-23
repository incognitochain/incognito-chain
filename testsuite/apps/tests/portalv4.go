package main

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/btcsuite/btcd/chaincfg"
//	"github.com/incognitochain/incognito-chain/config"
//	"github.com/incognitochain/incognito-chain/portal"
//	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
//	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
//	devframework "github.com/incognitochain/incognito-chain/testsuite"
//	"github.com/incognitochain/incognito-chain/testsuite/account"
//	"log"
//	"os/exec"
//	"strconv"
//)
//
//func Test_PortalV4() {
//
//	//sed -i s"|return putGenesisBlockIntoChainParams(genesisHash, genesisBlock, .*|return putGenesisBlockIntoChainParams\(genesisHash, genesisBlock, chaincfg.RegressionNetParams\)|" relaying/btc/relayinggenesis.go
//	//
//	node := devframework.NewStandaloneSimulation("newsim", devframework.Config{
//		Network: devframework.ID_LOCAL,
//		ResetDB: true,
//		AppNode: true,
//	})
//	config.Config().LimitFee = 0
//	config.Param().ActiveShards = 2
//	config.Param().BCHeightBreakPointNewZKP = 1
//	config.Param().BeaconHeightBreakPointBurnAddr = 2
//	config.Param().ConsensusParam.StakingFlowV2Height = 1
//	config.Param().EpochParam.NumberOfBlockInEpoch = 20
//	config.Param().EpochParam.RandomTime = 10
//	config.Param().EnableFeatureFlags["PortalV4"] = 1
//	portal.GetPortalParams().PortalParamsV4[0].PortalTokens[portalcommonv4.PortalBTCIDStr] = portaltokensv4.PortalBTCTokenProcessor{
//		PortalToken: &portaltokensv4.PortalToken{
//			ChainID:             portal.TestnetBTCChainID,
//			MinTokenAmount:      10,
//			MultipleTokenAmount: 10,
//			ExternalInputSize:   130,
//			ExternalOutputSize:  43,
//			ExternalTxMaxSize:   2048,
//		},
//		ChainParam: &chaincfg.RegressionNetParams,
//	}
//
//	node.Init()
//	for i := 0; i < 5; i++ {
//		node.GenerateBlock().NextRound()
//	}
//
//	acc0, _ := account.NewAccountFromPrivatekey("112t8rnX6USJnBzswUeuuanesuEEUGsxE8Pj3kkxkqvGRedUUPyocmtsqETX2WMBSvfBCwwsmMpxonhfQm2N5wy3SrNk11eYx6pMsmsic4Vz")
//	node.RPC.API_SubmitKey(acc0.PrivateKey)
//	err := node.RPC.API_CreateConvertCoinVer1ToVer2Transaction(node.GenesisAccount.PrivateKey)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	node.GenerateBlock().NextRound()
//	node.RPC.SendPRV(node.GenesisAccount, acc0, 888*1e9)
//	node.GenerateBlock().NextRound()
//	node.RPC.ShowBalance(node.GenesisAccount)
//
//	//generate bitcoin addressm
//
//	masterPubkey := portal.GetPortalParams().PortalParamsV4[0].MasterPubKeys[portalcommonv4.PortalBTCIDStr]
//	_, masterAddress, _ := portal.GetPortalParams().PortalParamsV4[0].PortalTokens[portalcommonv4.PortalBTCIDStr].GenerateOTMultisigAddress(masterPubkey, 3, "")
//	_, depositAddress, err := portal.GetPortalParams().PortalParamsV4[0].PortalTokens[portalcommonv4.PortalBTCIDStr].GenerateOTMultisigAddress(masterPubkey, 3, acc0.PaymentAddress)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println("Deposit address: ", depositAddress)
//	proof, err := exec.Command("bash", "portal_v4_test.sh", "send", "5", depositAddress).Output()
//	if err != nil {
//		log.Fatal(err)
//	}
//	tokenID := "ef5947f70ead81a76a53c7c8b7317dd5245510c665d3a13921dc9a581188728b"
//	resTx, err := node.RPC.Client.CreateAndSendTXShieldingRequest(node.GenesisAccount.PrivateKey, acc0.PaymentAddress, tokenID, string(proof))
//	fmt.Println("=========", resTx.TxID, err)
//
//	for i := 0; i < 10; i++ {
//		node.GenerateBlock().NextRound()
//		resReq, _ := node.RPC.Client.GetPortalShieldingRequestStatus(resTx.TxID)
//		if resReq != nil {
//			fmt.Printf("Response request shield %v - Mint Amount: %+v - Error: %+v\n", resReq.TxReqID, resReq.MintingAmount, resReq.Error)
//			break
//		}
//	}
//	GetBalance(depositAddress)
//	node.Pause()
//	node.GenerateBlock().NextRound()
//	node.RPC.ShowBalance(acc0)
//
//	resUnshieldTx, err := node.RPC.Client.CreateAndSendTxWithPortalV4UnshieldRequest(acc0.PrivateKey, tokenID, "3000000", acc0.PaymentAddress, "mgdwpAgvYNuJ2MyUimiKdTYsu2vpDZNpAa")
//	fmt.Printf("%+v %+v\n", resUnshieldTx, err)
//	for i := 0; i < 30; i++ {
//		node.GenerateBlock().NextRound()
//	}
//
//	node.RPC.ShowBalance(acc0)
//	node.GenerateBlock().NextRound()
//
//	heightStr := strconv.Itoa(int(node.GetBlockchain().BeaconChain.GetBestViewHeight()))
//	portalState, err := node.RPC.Client.GetPortalV4State(heightStr)
//	b, _ := json.Marshal(portalState)
//	fmt.Println(string(b))
//	batchID := ""
//	//for _, v := range portalState.ProcessedUnshieldRequests {
//	//	for _, v1 := range v {
//	//		//fmt.Printf("%+v\n", v1)
//	//		//fmt.Println("batchid", v1.GetBatchID())
//	//		batchID = v1.GetBatchID()
//	//	}
//	//}
//
//	if batchID == "" {
//		panic("no batch id")
//	}
//	batchReq, err := node.RPC.Client.GetPortalSignedRawTransaction(batchID)
//	fmt.Printf("%+v %+v\n", batchReq, err)
//
//	sendRes, err := exec.Command("bash", "portal_v4_test.sh", "sendrawtransaction", batchReq.SignedTx).Output()
//	fmt.Println(string(sendRes))
//
//	GetBalance("mgdwpAgvYNuJ2MyUimiKdTYsu2vpDZNpAa")
//	GetBalance(masterAddress)
//}
//
//type UnspentResult struct {
//	Amount float64 `json:"amount"`
//}
//
//func GetBalance(address string) {
//	sendRes, _ := exec.Command("bash", "portal_v4_test.sh", "listunspent", address).Output()
//	listUnspentRes := []UnspentResult{}
//	json.Unmarshal(sendRes, &listUnspentRes)
//	sum := float64(0)
//	for _, v := range listUnspentRes {
//		sum += v.Amount
//	}
//	fmt.Println("Balance", address, sum, "BTC")
//}
