package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

var (
	cfg *config
	rpc *RPC
	ai = 1
)

func main() {
	// Show version at startup.
	log.Printf("Version %s\n", "1")

	// load config
	tcfg, err := loadConfig()
	if err != nil {
		log.Println("Parse config error", err.Error())
		return
	}
	cfg = tcfg
	rpc = InitRPC(cfg.RPCAddress[0])

	if cfg.Strategy == 1 {
		strategy1()
	} else if cfg.Strategy == 2 {
		strategy2()
	}
}

/*
Strategy 1: send out n transactions per second by m transactions
*/
func strategy1() {
	totalSendOut := 0
	stepSendout := 100

	if stepSendout > cfg.TotalTxs {
		stepSendout = cfg.TotalTxs
	}

	for {
		if totalSendOut >= cfg.TotalTxs {
			log.Println("totalSendout", totalSendOut, "cfg.TotalTxs", cfg.TotalTxs, stepSendout)
			break
		}

		for i := 0; i < stepSendout; i++ {
			go func() {
				isSuccess, hash := sendRandomTransaction()
				if isSuccess {
					log.Printf("Send a transaction success: %s", hash)
				}
			}()
		}

		totalSendOut += stepSendout

		log.Printf("Send out %d transactions\n", totalSendOut)
		time.Sleep(1 * time.Second)
	}
}

/*
Strategy 2: send out n transactions once
*/
func strategy2() {
	totalSendOut := 0

	for i := 0; i < cfg.TotalTxs; i++ {
		isSuccess, hash := sendRandomTransaction()
		if isSuccess {
			log.Printf("Send a transaction success: %s", hash)
		}
		totalSendOut += 1
	}

	log.Printf("Send out %d transactions\n", totalSendOut)
}

func randomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func sendRandomTransaction() (bool, interface{}) {
	// todo create account
	err, wallet := rpc.GetAccountAddress(fmt.Sprintf("Account %d", ai))
	if err != nil {
		return false, nil
	}
	ai += 1
	// todo send many
	value := float64(randomInt(1, 10000000)) / math.Pow(10, 18)
	err, txId := rpc.SendMany(cfg.GenesisPrvKey, wallet.PublicKey, value)

	return true, txId
}