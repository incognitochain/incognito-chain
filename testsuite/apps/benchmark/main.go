package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common/base58"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	//submitKey()
	//time.Sleep(time.Second * 10)
	//sendSeed()
	//time.Sleep(time.Second * 10)
	//distributePRVCoin()
	//time.Sleep(time.Second * 20)
	createBenchmark()
	time.Sleep(time.Second * 10)
	sendTx()

}

var ts = 4

var fullnodeRPC = devframework.RemoteRPCClient{"http://127.0.0.1:30001"}
var fullnode1RPC = devframework.RemoteRPCClient{"http://127.0.0.1:30001"}
var shard0RPC = devframework.RemoteRPCClient{"http://127.0.0.1:20004"}
var icoAccount, _ = account.NewAccountFromPrivatekey("112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or")

const ACCOUNT_SIZE = 600
const INPUT_COIN = 3
const BATCH = 20

func CreateAccounts(rpc *devframework.RemoteRPCClient, seed string, size int) []account.Account {
	shard0 := make([]account.Account, size)
	semaphore := make(chan int, 50)
	for i := 0; i < size; i++ {
		semaphore <- 1
		go func(i int) {
			fmt.Println("create account ", i)
			acc, _ := account.GenerateAccountByShard(0, 0, fmt.Sprintf("%v%v", seed, i))
			shard0[i] = acc
			if rpc != nil {
				rpc.SubmitKey(acc.PrivateKey)
				time.Sleep(time.Millisecond * 10)
			}
			//if i == 8368 {
			//	//res, err := shard0RPC.AuthorizedSubmitKey(shard0[i].PrivateKey)
			//	//log.Println("Submit key ICO", res, err)
			//	//os.Exit(-1)
			//	balance, err := shard0RPC.GetBalanceByPrivateKey(shard0[i].PrivateKey)
			//	log.Println("Balance", balance, err)
			//	os.Exit(-1)
			//
			//}
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

func submitKey() {
	seedAccounts := CreateAccounts(&shard0RPC, "seed", BATCH)
	//submit  all key first, before convert
	{
		CreateAccounts(&shard0RPC, "seed", BATCH)
		res, err := fullnodeRPC.SubmitKey(icoAccount.PrivateKey)
		log.Println("Submit key ICO", res, err)
		res, err = fullnode1RPC.SubmitKey(icoAccount.PrivateKey)
		log.Println("Submit key ICO", res, err)
		res, err = shard0RPC.SubmitKey(icoAccount.PrivateKey)
		log.Println("Submit key ICO", res, err)
		err = fullnodeRPC.CreateConvertCoinVer1ToVer2Transaction(icoAccount.PrivateKey)
		log.Println("Convert coin", err)
	}

	{
		for i := range seedAccounts {
			res, err := fullnode1RPC.AuthorizedSubmitKey(seedAccounts[i].PrivateKey)
			log.Println("Submit key ICO", res, err)
		}
		return
	}

}

func sendSeed() {
	seedAccounts := CreateAccounts(&shard0RPC, "seed", BATCH)

	//split account
	{
		balanceICO, _ := shard0RPC.GetBalanceByPrivateKey(icoAccount.PrivateKey)
		fmt.Println("ICO Balance", balanceICO)
		startIndex := 0
		for {
			//balanceICO, _ := shard0RPC.GetBalanceByPrivateKey(icoAccount.PrivateKey)
			//fmt.Println("ICO Balance", balanceICO)
			fmt.Println("send", float64(balanceICO/BATCH/2)-100)
			endIndex := startIndex + 30
			if endIndex >= len(seedAccounts) {
				endIndex = len(seedAccounts)
			}
			receiver := []interface{}{icoAccount.PrivateKey}
			for _, sacc := range seedAccounts[startIndex:endIndex] {
				receiver = append(receiver, sacc.PaymentAddress)
				receiver = append(receiver, float64(balanceICO/BATCH/2)-100)
			}
		RESEND:
			sendSeed, err := SendPRV(shard0RPC, receiver...)
			if err != nil {
				log.Println(err.Error())
				balanceICO, _ := shard0RPC.GetBalanceByPrivateKey(icoAccount.PrivateKey)
				fmt.Println("ICO Balance", balanceICO)
				goto RESEND
			}
			fmt.Println(sendSeed, err, float64(balanceICO/BATCH/2)-100)
			time.Sleep(time.Duration(ts) * time.Second)
		RECHECK:
			time.Sleep(time.Duration(ts) * time.Second / 2)
			txRes, err := shard0RPC.GetTransactionByHash(sendSeed)
			if err != nil || !txRes.IsInBlock {
				fmt.Println("recheck ", sendSeed, txRes.IsInMempool)
				goto RECHECK
			}
			log.Println("cont to send...")
			startIndex += 30
			if startIndex >= len(seedAccounts) {
				break
			}
		}

		for i := range seedAccounts {
			balanceSeed, err := shard0RPC.GetBalanceByPrivateKey(seedAccounts[i].PrivateKey)
			log.Println("Balance Seed", i, balanceSeed, err)
		}
		return
	}
}

func distributePRVCoin() {
	seedAccounts := CreateAccounts(nil, "seed", BATCH)
	shard0 := CreateAccounts(&fullnode1RPC, "5", ACCOUNT_SIZE)

	fmt.Println("send PRV to accounts ...")
	wg := sync.WaitGroup{}
	for sid, sacc := range seedAccounts {
		wg.Add(1)
		batchsize := len(shard0) / BATCH
		receiveAccounts := shard0[sid*batchsize : (sid+1)*batchsize]
		go func(sacc account.Account, receiveAccounts []account.Account) {
			log.Println("Run account", sacc.PrivateKey, len(receiveAccounts))
			for coin := 0; coin < INPUT_COIN; coin++ {
				for i := 0; i < len(receiveAccounts); i += 30 {
					receiver := []interface{}{sacc.PrivateKey}
				RETRY:
					for j := i; j < i+30 && j < len(receiveAccounts); j++ {
						receiver = append(receiver, receiveAccounts[j].PaymentAddress)
						receiver = append(receiver, float64(1e10))
					}
					sendTx1, err := SendPRV(shard0RPC, receiver...)
					if err != nil {
						fmt.Println(err)
						time.Sleep(5 * time.Second)
						goto RETRY
					}
					log.Println("Send Tx", sendTx1, err, i, coin)
					time.Sleep(time.Duration(ts) * time.Second)
				RECHECK:
					time.Sleep(time.Second)
					txRes, err := shard0RPC.GetTransactionByHash(sendTx1)
					if err != nil || !txRes.IsInBlock {
						fmt.Println("recheck ", sendTx1, err)
						goto RECHECK
					}
				}
			}
			wg.Done()
		}(sacc, receiveAccounts)
	}
	wg.Wait()
	return
}

func createBenchmark() {
	shard0 := CreateAccounts(&fullnode1RPC, "5", ACCOUNT_SIZE)
	{
		semaphore := make(chan int, 50)
		fd, _ := os.OpenFile("input3", os.O_CREATE|os.O_RDWR, 0666)
		for i, acc := range shard0 {
			balance, err := fullnodeRPC.GetBalanceByPrivateKey(acc.PrivateKey)
			log.Println("Balance", i, balance, acc.PrivateKey, err)
			semaphore <- 1
			go func(i int, acc account.Account) {
				bytes, err := CreateRawTx(fullnodeRPC, acc.PrivateKey, acc.PaymentAddress, float64(30))
				if err != nil {
					panic(err)
				}
				_, _, err = base58.Base58Check{}.Decode(string(bytes))
				if err != nil {
					panic(err)
				}
				//fmt.Println(bytes, len(bytes))
				//panic(1)
				//log.Println("Balance", i, acc.PrivateKey, len(bytes))
				if err != nil {
					balance, err := fullnodeRPC.GetBalanceByPrivateKey(acc.PrivateKey)
					log.Println("Balance", i, balance, acc.PrivateKey, err)
					panic(err)
				}
				fd.WriteString(bytes + "\n")
				<-semaphore
			}(i, acc)
		}
		for {
			if len(semaphore) == 0 {
				break
			}
			time.Sleep(time.Second)
		}
		fd.Close()
		fmt.Println("finish")
	}
}

func sendTx() {
	fd, err := os.OpenFile("input3", os.O_RDONLY, 0666)
	fmt.Println("start...")
	if err != nil {
		panic(err)
	}
	semaphore := make(chan int, 50)

	s := bufio.NewReader(fd)
	i := 0
	for {
		var buffer bytes.Buffer
		var l []byte
		var isPrefix bool
		for {
			l, isPrefix, err = s.ReadLine()
			buffer.Write(l)

			// If we've reached the end of the line, stop reading.
			if !isPrefix {
				break
			}

			// If we're just at the EOF, break
			if err != nil {
				break
			}
		}
		if err == io.EOF {
			break
		}
		line := string(buffer.Bytes())

		if i%1000 == 0 {
		CHECKMEM:
			memInfo, _ := shard0RPC.GetMempoolInfo()
			if memInfo.Size > 2000 {
				time.Sleep(time.Second)
				fmt.Println("mempool", memInfo.Size)
				goto CHECKMEM
			}
		}

		//go func(i int) {
		if len(line) == 0 {
			continue
		}
		i++
		semaphore <- 1
		go func(i int) {
			defer func() {
				<-semaphore
			}()
			data, err := fullnodeRPC.SendRawTransaction(string(line))
			fmt.Println(i, data.TxID, err)
			if err != nil {
				fmt.Println(string(line))
				panic(1)
			}
		}(i)
		time.Sleep(time.Millisecond * 20)
	}
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
