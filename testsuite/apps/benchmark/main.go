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

const ACCOUNT_SIZE = 300
const INPUT_COIN = 2

func CreateAccounts() []account.Account {
	seed := time.Now().String()
	shard0 := [ACCOUNT_SIZE]account.Account{}
	semaphore := make(chan int, 100)
	for i := 0; i < len(shard0); i++ {
		semaphore <- 1
		go func(i int) {
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
	// res, err := fullnodeRPC.SubmitKey(icoAccount.PrivateKey)
	// log.Println("Submit key ICO", res, err)
	// err := fullnodeRPC.CreateConvertCoinVer1ToVer2Transaction(icoAccount.PrivateKey)
	// log.Println("Convert coin", err)
	// res, err := fullnodeRPC.CreateAndSendTokenInitTransaction(rpcserver.TokenInitParam{icoAccount.PrivateKey, "random1", "random1", 10000000000000000})
	// fmt.Println(res, err)
	//}

	fmt.Println("creating accounts ...")
	shard0 := CreateAccounts()

	fmt.Println("send PRV to accounts ...")
	for i := 0; i < len(shard0); i += (30 / INPUT_COIN) {
	RETRY:
		receiver := []interface{}{icoAccount.PrivateKey}
		for j := i; j < len(shard0); j += (30 / INPUT_COIN) {
			for coin := 0; coin < INPUT_COIN; coin++ {
				receiver = append(receiver, shard0[j].PaymentAddress)
				receiver = append(receiver, float64(10))
			}
		}
		log.Println("try to send from ", i, "to", i+(30/INPUT_COIN))
		sendTx1, err := SendPRV(fullnodeRPC, receiver...)
		if err != nil {
			time.Sleep(5 * time.Second)
			goto RETRY
		}
		log.Println("Send Tx", sendTx1, err)
		time.Sleep(10 * time.Second)
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
	balance0, err := fullnodeRPC.GetBalanceByPrivateKey(shard0[0].PrivateKey)
	log.Println("Acc00 Balance", balance0, err)

	balance1, err := fullnodeRPC.GetBalanceByPrivateKey(shard0[200].PrivateKey)
	log.Println("Acc01 Balance", balance1, err)

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
