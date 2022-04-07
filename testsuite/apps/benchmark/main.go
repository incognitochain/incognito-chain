package main

import (
	"fmt"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"log"
	"time"
)

var fullnodeRPC = devframework.RemoteRPCClient{"http://51.178.74.7:9333"}
var icoAccount, _ = account.NewAccountFromPrivatekey("112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or")

const ACCOUNT_SIZE = 30
const INPUT_COIN = 2

func CreateAccounts() []account.Account {
	//seed := time.Now().String()
	seed := "3"
	shard0 := [ACCOUNT_SIZE]account.Account{}
	semaphore := make(chan int, 100)
	for i := 0; i < len(shard0); i++ {
		semaphore <- 1
		go func(i int) {
			//fmt.Println("create accoutn ", i)
			acc, _ := account.GenerateAccountByShard(0, 0, fmt.Sprintf("%v%v", seed, i))
			shard0[i] = acc
			go fullnodeRPC.SubmitKey(acc.PrivateKey)
			<-semaphore
		}(i)
	}
	for {
		if len(semaphore) == 0 {
			break
		}
		time.Sleep(time.Second)
	}
	return shard0[:]
}

func main() {

	//submit  all key first, before convert
	//{
	//	res, err := fullnodeRPC.SubmitKey(icoAccount.PrivateKey)
	//	log.Println("Submit key ICO", res, err)
	//	err = fullnodeRPC.CreateConvertCoinVer1ToVer2Transaction(icoAccount.PrivateKey)
	//	log.Println("Convert coin", err)
	//	return
	//}

	//{
	//	res, err := fullnodeRPC.CreateAndSendTokenInitTransaction(rpcserver.TokenInitParam{icoAccount.PrivateKey, "random1", "random1", 10000000000000000})
	//	fmt.Println(res, err)
	//	return
	//}

	//{
	//	res, err := fullnodeRPC.AuthorizedSubmitKey(icoAccount.PrivateKey)
	//	log.Println("Submit key ICO", res, err)
	//}

	balanceICO, err := fullnodeRPC.GetBalanceByPrivateKey(icoAccount.PrivateKey)
	log.Println("Balance ICO", balanceICO, err)

	fmt.Println("creating accounts ...")
	shard0 := CreateAccounts()

	fmt.Println("send PRV to accounts ...")
	for coin := 0; coin < INPUT_COIN; coin++ {
		for i := 0; i < len(shard0); i += 30 {
			receiver := []interface{}{icoAccount.PrivateKey}
		RETRY:
			for j := i; j < i+30 && j < len(shard0); j++ {
				receiver = append(receiver, shard0[j].PaymentAddress)
				receiver = append(receiver, float64(10))
			}
			sendTx1, err := SendPRV(fullnodeRPC, receiver...)
			if err != nil {
				time.Sleep(5 * time.Second)
				goto RETRY
			}
			log.Println("Send Tx", sendTx1, err)
			time.Sleep(10 * time.Second)
		RECHECK:
			time.Sleep(500 * time.Millisecond)
			txRes, err := fullnodeRPC.GetTransactionByHash(sendTx1)
			if err != nil || !txRes.IsInBlock {
				fmt.Println("recheck ", sendTx1, err)
				goto RECHECK
			}
		}
	}

	//balance, err := fullnodeRPC.GetBalanceByPrivateKey(icoAccount.PrivateKey)
	//log.Println("ICO Balance", balance, err)
	//
	//receiver := []interface{}{icoAccount.PrivateKey}
	//for i := 0; i < 30; i++ {
	//	receiver = append(receiver, shard0[i].PaymentAddress)
	//	receiver = append(receiver, float64(10))
	//}
	//
	//sendTx1, err := SendPRV(fullnodeRPC, receiver...)
	//log.Println("Send Tx", sendTx1, err)
	//

	//txBuffer := []string{}
	for i, acc := range shard0 {
		balance, err := fullnodeRPC.GetBalanceByPrivateKey(acc.PrivateKey)
		log.Println("Balance", i, balance, err)
		//bytes, err := CreateRawTx(fullnodeRPC, acc.PrivateKey, acc.PaymentAddress, float64(10))
		//if err != nil {
		//	panic(err)
		//}
		//txBuffer = append(txBuffer, bytes)
	}

	//for _, tx := range txBuffer {
	//	go func() {
	//		data, err := fullnodeRPC.SendRawTransaction(tx)
	//		fmt.Println(data, err)
	//	}()
	//
	//}
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

func CreateRawTx(fullnodeRPC devframework.RemoteRPCClient, args ...interface{}) (string, error) {
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
	res, err := fullnodeRPC.CreateRawTransaction(sender, mapUintToInterface(receivers), 0, 1)
	if err != nil {
		return "", err
	}
	return res.Base58CheckData, nil
}

func mapUintToInterface(m map[string]uint64) map[string]interface{} {
	mfl := make(map[string]interface{})
	for a, v := range m {
		mfl[a] = float64(v)
	}
	return mfl
}
