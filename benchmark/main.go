package main

import (
	"github.com/ninjadotorg/cash-prototype/benchmark/api"
	"log"
	"math/rand"
	"time"
)


var (
	cfg *config
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

	if cfg.Strategy == 1 {
		strategy1()
	} else if cfg.Strategy == 2 {
		strategy2()
	} else if cfg.Strategy == 3 {
		strategy3()
	}
}

/**
Strategy 1: send out 1k transactions per second by n transactions
 */
func strategy1() {
	totalSendOut := 0
	stepSendout := 1000
	for {
		if totalSendOut >= cfg.TotalTxs {
			break
		}

		for i := 0; i < stepSendout; i++ {
			isSuccess, hash := sendRandomTransaction(-1)
			if isSuccess {
				log.Printf("Send a transaction success: %s", hash)
			}
		}

		totalSendOut += stepSendout

		log.Printf("Send out %dk transactions\n", totalSendOut)
		time.Sleep(1 * time.Second)
	}
}

/**
Strategy 2: send out n transactions
 */
func strategy2() {
	totalSendOut := 0

	for i := 0; i < cfg.TotalTxs; i++ {
		isSuccess, hash := sendRandomTransaction(-1)
		if isSuccess {
			log.Printf("Send a transaction success: %s", hash)
		}
		totalSendOut += 1
	}

	log.Printf("Send out %d transactions\n", totalSendOut)
}

/**
Strategy 3: send out n transactions to 1 node only
 */
func strategy3() {
	totalSendOut := 0

	for i := 0; i < cfg.TotalTxs; i++ {
		isSuccess, hash := sendRandomTransaction(0)
		if isSuccess {
			log.Printf("Send a transaction success: %s", hash)
		}
		totalSendOut += 1
	}

	log.Printf("Send out %d transactions\n", totalSendOut)
}

func randomInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max - min) + min
}

func sendRandomTransaction(ah int) (bool, string) {
	id := randomInt(0, 10)
	addressHash := ah // random number 0-255
	if addressHash == -1 {
		addressHash = randomInt(0, 255)
	}
	txId := "9f224940a04bb371a0f63d69b41e5adb32e34b74446458bb73d088f02ef71bea" // random hash
	pkScript := "mgnUx4Ah4VBvtaL7U1VXkmRjKUk3h8pbst" // random hash
	value := randomInt(1, 10000000) // random number 1 - 10000000

	params := buildCreateParams(addressHash, txId, pkScript, value, id)
	err, tx := api.Get(cfg.RPCAddress, params)

	if err != nil {
		log.Printf("send transaction error: %s", err.Error())
		return false, ""
	}

	sendParams := buildSendParams(tx["result"].(string), id)

	err, response := api.Get(cfg.RPCAddress, sendParams)

	if err != nil {
		return false, ""
	}

	return true, response["result"].(string)
}

func buildCreateParams(addressHash int, txId string, pkScript string, value int, id int) map[string]interface{} {
	data := map[string]interface{}{}

	data["jsonrpc"] = "1.0"
	data["method"] = "createrawtransaction"

	params := []interface{}{}
	params = append(params, map[string]interface{}{"addressHash": addressHash})
	inTxs := []map[string]interface{}{}
	inTxs = append(inTxs, map[string]interface{}{
		"txid": txId,
		"vout": 0,
	})
	params = append(params, inTxs)
	outTxs := []map[string]interface{}{}
	outTxs = append(outTxs, map[string]interface{}{
		"pkScript": pkScript,
		"value": value,
		"txOutType": "TXOUT_COIN",
	})
	params = append(params, outTxs)

	data["params"] = params
	data["id"] = id
	return data
}

func buildSendParams(params string, id int) map[string]interface{} {
	data := map[string]interface{}{}

	data["jsonrpc"] = "1.0"
	data["method"] = "sendrawtransaction"
	data["params"] = []string{params}
	data["id"] = id

	return data
}