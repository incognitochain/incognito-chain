package main

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/portal"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
	"github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"strconv"
)

func Test_PortalV4() {
	//sed -i s"|return putGenesisBlockIntoChainParams(genesisHash, genesisBlock, .*|return putGenesisBlockIntoChainParams\(genesisHash, genesisBlock, chaincfg.RegressionNetParams\)|" relaying/btc/relayinggenesis.go

	node := devframework.NewStandaloneSimulation("newsim", devframework.Config{
		Network: devframework.ID_TESTNET,
		ResetDB: true,
	})
	config.Config().LimitFee = 0
	config.Param().ActiveShards = 2
	config.Param().BCHeightBreakPointNewZKP = 1
	config.Param().BeaconHeightBreakPointBurnAddr = 2
	config.Param().ConsensusParam.StakingFlowV2Height = 1
	config.Param().EpochParam.NumberOfBlockInEpoch = 20
	config.Param().EpochParam.RandomTime = 10
	config.Param().EnableFeatureFlags[2] = 1
	portal.GetPortalParams().PortalParamsV4[0].PortalTokens[portalcommonv4.PortalBTCIDStr] = portaltokensv4.PortalBTCTokenProcessor{
		PortalToken: &portaltokensv4.PortalToken{
			ChainID:             portal.TestnetBTCChainID,
			MinTokenAmount:      10,
			MultipleTokenAmount: 10,
			ExternalInputSize:   130,
			ExternalOutputSize:  43,
			ExternalTxMaxSize:   2048,
		},
		ChainParam: &chaincfg.RegressionNetParams,
	}

	node.Init()
	for i := 0; i < 5; i++ {
		node.GenerateBlock().NextRound()
	}

	acc0, _ := account.NewAccountFromPrivatekey("112t8rnX6USJnBzswUeuuanesuEEUGsxE8Pj3kkxkqvGRedUUPyocmtsqETX2WMBSvfBCwwsmMpxonhfQm2N5wy3SrNk11eYx6pMsmsic4Vz")

	node.RPC.API_SubmitKey(node.GenesisAccount.PrivateKey)
	node.RPC.API_SubmitKey(acc0.PrivateKey)
	err := node.RPC.API_CreateConvertCoinVer1ToVer2Transaction(node.GenesisAccount.PrivateKey)
	if err != nil {
		fmt.Println(err)
	}

	node.GenerateBlock().NextRound()
	node.RPC.SendPRV(node.GenesisAccount, acc0, 888*1e9)
	node.GenerateBlock().NextRound()
	node.RPC.ShowBalance(node.GenesisAccount)

	proof := "eyJNZXJrbGVQcm9vZnMiOlt7IlByb29mSGFzaCI6WzMsMTk3LDExNiw2NCwxNTQsMjAzLDIyNiwyMTMsMTUsODQsODEsNjEsMTUwLDEwNSwxNjIsOCwxNjMsMjQ3LDYwLDM2LDE3NCwyMjcsNTgsNDgsODUsMzcsNzcsMTcsMjI3LDIxMSw5MCw0OV0sIklzTGVmdCI6dHJ1ZX1dLCJCVENUeCI6eyJWZXJzaW9uIjoyLCJUeEluIjpbeyJQcmV2aW91c091dFBvaW50Ijp7Ikhhc2giOlsyMywyMDMsNTAsMTczLDE5NywyMzMsMTgyLDM0LDMxLDEwNiwxNzQsMjE1LDQxLDE4MywzMCwyNTEsMjMxLDE4MywxODYsMTgsMTA2LDIyNywxOTQsNywxMDgsOTMsMjA2LDg3LDgsOTYsMjQyLDIwOF0sIkluZGV4IjoxfSwiU2lnbmF0dXJlU2NyaXB0IjoiU0RCRkFpRUF0QmcrWTlTMHdiMWM5Q0hWN1NyczF1K1A3QndRQUFXQklJRStmbENibmljQ0lIa3lOc3RCMDhVUUpMRXZNdFRtTlFuSUlLMkpvY1lEMnJCUFJxVk1PNXIvQVNNaEFuajNxZVhtWE5UUWNPTU16dXJNazkwbGRJZGFyamtWcHArdXlUcHdxdW11ckE9PSIsIldpdG5lc3MiOm51bGwsIlNlcXVlbmNlIjo0Mjk0OTY3Mjk1fV0sIlR4T3V0IjpbeyJWYWx1ZSI6MTU1NTUwMDAwLCJQa1NjcmlwdCI6IkFDQWlDM0dPZktOTXg2UTR4KzhNQktsWnVGUFltMGdWYklzaEFnbWZhdm56aHc9PSJ9LHsiVmFsdWUiOjg0Mzk1MDAwMCwiUGtTY3JpcHQiOiJxUlNKc0IwZ1lyNDRDTW1BMlcwcGswcGh5cVFEZkljPSJ9XSwiTG9ja1RpbWUiOjB9LCJCbG9ja0hhc2giOls4NCwxODgsMjI2LDIxOCw2NiwxNiwxNSwyMTgsOTEsMjQwLDQyLDE4MSwyMzAsMTIyLDcyLDMzLDAsMTQ2LDM2LDIwMSwxNjUsMjE2LDE1NiwyNDksMTEsMTc3LDE4LDk1LDg0LDI0MCwyMTYsODJdfQ=="
	tokenID := "ef5947f70ead81a76a53c7c8b7317dd5245510c665d3a13921dc9a581188728b"
	resTx, err := node.RPC.Client.CreateAndSendTXShieldingRequest(node.GenesisAccount.PrivateKey, acc0.PaymentAddress, tokenID, proof)
	fmt.Println("=========", resTx.TxID, err)

	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
		resReq, _ := node.RPC.Client.GetPortalShieldingRequestStatus(resTx.TxID)
		if resReq != nil {
			fmt.Printf("Response request shield %v - Mint Amount: %+v - Error: %+v\n", resReq.TxReqID, resReq.MintingAmount, resReq.Error)
			break
		}
	}
	node.GenerateBlock().NextRound()
	node.RPC.ShowBalance(acc0)

	resUnshieldTx, err := node.RPC.Client.CreateAndSendTxWithPortalV4UnshieldRequest(acc0.PrivateKey, tokenID, "2000000", acc0.PaymentAddress, "mgdwpAgvYNuJ2MyUimiKdTYsu2vpDZNpAa")
	fmt.Printf("%+v %+v\n", resUnshieldTx, err)
	for i := 0; i < 30; i++ {
		node.GenerateBlock().NextRound()
	}

	node.RPC.ShowBalance(acc0)
	node.GenerateBlock().NextRound()

	heightStr := strconv.Itoa(int(node.GetBlockchain().BeaconChain.GetBestViewHeight()))
	portalState, err := node.RPC.Client.GetPortalV4State(heightStr)
	b, _ := json.Marshal(portalState)
	fmt.Println(string(b))
	batchID := ""
	for _, v := range portalState.ProcessedUnshieldRequests {
		for _, v1 := range v {
			fmt.Printf("%+v\n", v1)
			fmt.Println("batchid", v1.GetBatchID())
			batchID = v1.GetBatchID()
		}

	}
	if batchID == "" {
		panic("no batch id")
	}
	batchReq, err := node.RPC.Client.GetPortalSignedRawTransaction(batchID)
	fmt.Printf("%+v %+v", batchReq, err)

}
