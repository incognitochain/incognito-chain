package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/devframework"
	"github.com/incognitochain/incognito-chain/devframework/account"
	"github.com/incognitochain/incognito-chain/devframework/rpcclient"
)

func main() {
	sAcc, err := account.NewAccountFromPrivatekey("")
	if err != nil {
		panic(err)
	}
	b, err := json.MarshalIndent(sAcc, "", "")
	fmt.Println(string(b))
	seed := ""
	keyID := 0
	account1, _ := account.GenerateAccountByShard(0, keyID, seed)
	keyID++
	//account2, _ := devframework.GenerateAccountByShard(1, keyID, seed)
	//keyID++
	//account3, _ := devframework.GenerateAccountByShard(2, keyID, seed)
	//keyID++
	//account4, _ := devframework.GenerateAccountByShard(3, keyID, seed)
	//keyID++
	//account5, _ := devframework.GenerateAccountByShard(4, keyID, seed)
	//keyID++
	//account6, _ := devframework.GenerateAccountByShard(5, keyID, seed)
	//keyID++
	//account7, _ := devframework.GenerateAccountByShard(6, keyID, seed)
	//keyID++
	//account8, _ := devframework.GenerateAccountByShard(7, keyID, seed)
	//keyID++

	//fmt.Println(account1.SelfCommitteePubkey)
	//fmt.Println(account2.SelfCommitteePubkey)

	client := devframework.NewRPCClient("http://139.162.54.236:38934")

	res, err := client.API_GetBalance(*sAcc)
	fmt.Println(err)
	fmt.Println(res)

	stake1 := rpcclient.StakingTxParam{
		Name:        "stake_1",
		BurnAddr:    "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA",
		StakerPrk:   sAcc.PrivateKey,
		MinerPrk:    account1.PrivateKey,
		RewardAddr:  sAcc.PaymentAddress,
		StakeShard:  true,
		AutoRestake: true,
	}

	resStakingTx, err := client.API_CreateAndSendStakingTransaction(stake1)
	if err == nil {
		fmt.Println(resStakingTx)
	}

}
