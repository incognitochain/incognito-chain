package main

import (
	"flag"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"github.com/incognitochain/incognito-chain/wallet"
	"time"
)

func StakeShard(client *devframework.RemoteRPCClient, privateKey string, miningPrivateKey string, stakeAmount uint64) (*jsonresult.CreateTransactionResult, error) {
	burnAddr := "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
	wl, _ := wallet.Base58CheckDeserialize(privateKey)
	RewardAddr := wl.Base58CheckSerialize(wallet.PaymentAddressType)
	funderPayment := wl.Base58CheckSerialize(wallet.PaymentAddressType)

	miningWl, _ := wallet.Base58CheckDeserialize(miningPrivateKey)
	privateSeedBytes := common.HashB(common.HashB(miningWl.KeySet.PrivateKey))
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)

	txResp, err := client.CreateAndSendStakingTransaction(privateKey, map[string]interface{}{burnAddr: float64(stakeAmount)}, 1, 0, map[string]interface{}{
		"StakingType":                  float64(63),
		"CandidatePaymentAddress":      funderPayment,
		"PrivateSeed":                  privateSeed,
		"RewardReceiverPaymentAddress": RewardAddr,
		"AutoReStaking":                true,
	})

	return &txResp, err
}
func SendPRV(fullnodeRPC devframework.RemoteRPCClient, args ...interface{}) (string, error) {
	var sender string
	var receivers = make(map[string]uint64)
	for i, arg := range args {
		if i == 0 {
			sender = arg.(string)
		} else {
			switch arg.(type) {
			default:
				if i%2 == 1 {
					amount, ok := args[i+1].(float64)
					if !ok {
						amountF64 := args[i+1].(float64)
						amount = amountF64
					}
					receivers[arg.(string)] = uint64(amount)
				}
			}
		}
	}
	res, err := fullnodeRPC.CreateAndSendTransaction(sender, mapUintToInterface(receivers), -1, 1)
	if err != nil {
		return "", err
	}
	return res.TxID, nil
}
func mapUintToInterface(m map[string]uint64) map[string]interface{} {
	mfl := make(map[string]interface{})
	for a, v := range m {
		mfl[a] = float64(v)
	}
	return mfl
}

func main() {
	flag.Parse()
	fullnodeRPC := devframework.RemoteRPCClient{"http://127.0.0.1:30001"}
	shard0RPC := devframework.RemoteRPCClient{"http://127.0.0.1:20004"}
	var icoAccount, _ = account.NewAccountFromPrivatekey("112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or")
	account1, _ := account.GenerateAccountByShard(0, 0, "1")
	account2, _ := account.GenerateAccountByShard(0, 1, "1")
	account3, _ := account.GenerateAccountByShard(0, 2, "1")

	submitKey := func() {
		fullnodeRPC.SubmitKey(icoAccount.PrivateKey)
		shard0RPC.SubmitKey(icoAccount.PrivateKey)
		fullnodeRPC.SubmitKey(account1.PrivateKey)
		shard0RPC.SubmitKey(account1.PrivateKey)
		fullnodeRPC.SubmitKey(account2.PrivateKey)
		shard0RPC.SubmitKey(account2.PrivateKey)
		fullnodeRPC.SubmitKey(account3.PrivateKey)
		shard0RPC.SubmitKey(account3.PrivateKey)
		fullnodeRPC.CreateConvertCoinVer1ToVer2Transaction(icoAccount.PrivateKey)
	}
	send_prv := func() {
		receiver := []interface{}{icoAccount.PrivateKey}
		receiver = append(receiver, account1.PaymentAddress)
		receiver = append(receiver, float64(100000*1e10))
		receiver = append(receiver, account2.PaymentAddress)
		receiver = append(receiver, float64(100000*1e10))
		receiver = append(receiver, account3.PaymentAddress)
		receiver = append(receiver, float64(100000*1e10))
		tx, err := SendPRV(shard0RPC, receiver...)
		fmt.Println(tx, err)
	}
	stake_shard := func() {
		res, err := shard0RPC.GetBalanceByPrivateKey(account1.PrivateKey)
		fmt.Println(res, err)
		stakeRes, err := StakeShard(&shard0RPC, account1.PrivateKey, account1.PrivateKey, 1750*1e9)
		fmt.Println(stakeRes, err)

		res, err = shard0RPC.GetBalanceByPrivateKey(account2.PrivateKey)
		fmt.Println(res, err)
		stakeRes, err = StakeShard(&shard0RPC, account2.PrivateKey, account2.PrivateKey, 1750*1e9)
		fmt.Println(stakeRes, err)

		res, err = shard0RPC.GetBalanceByPrivateKey(account3.PrivateKey)
		fmt.Println(res, err)
		stakeRes, err = StakeShard(&shard0RPC, icoAccount.PrivateKey, account3.PrivateKey, 1750*1e9)
		fmt.Println(stakeRes, err)
	}
	submitKey()
	time.Sleep(time.Second * 10)
	send_prv()
	time.Sleep(time.Second * 10)
	stake_shard()
	return
	stake_shard()
	submitKey()
	send_prv()

}
