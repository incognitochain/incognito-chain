package main

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/ninjadotorg/cash-prototype/benchmark/api"
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
	stepSendout := 10

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
				isSuccess, hash := sendRandomTransaction(-1)
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
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func sendRandomTransaction(ah int) (bool, interface{}) {
	id := randomInt(0, 10)
	addressHash := ah // random number 0-255
	if addressHash == -1 {
		addressHash = randomInt(0, 255)
	}

	txId := strconv.Itoa(randomInt(10000000000, 20000000000))     // random hash
	pkScript := strconv.Itoa(randomInt(20000000000, 30000000000)) // random hash
	value := randomInt(1, 10000000)                               // random number 1 - 10000000

	params := buildCreateParams(addressHash, txId, pkScript, value, id)
	//jsonValue, _ := json.Marshal(params)
	//log.Println("Create transaction params", string(jsonValue))

	var endpoint string

	if len(cfg.RPCAddress) == 1 {
		endpoint = cfg.RPCAddress[0]
	} else {
		randEndpointIdx := randomInt(0, len(cfg.RPCAddress)-1)
		endpoint = cfg.RPCAddress[randEndpointIdx]
	}

	err, tx := api.Get(endpoint, params)

	//log.Println("CREATE transaction params response", tx)

	if err != nil {
		log.Printf("send transaction error: %s", err.Error())
		return false, ""
	}

	sendParams := buildSendParams(tx["result"].(string), id)
	//jsonValue2, _ := json.Marshal(sendParams)
	//log.Println("Send transaction params", string(jsonValue2))

	err, response := api.Get(endpoint, sendParams)

	//log.Println("Send transaction params response", response)

	if err != nil {
		return false, ""
	}

	return true, response["result"]
}

func buildCreateParams(addressHash int, txId string, pkScript string, value int, id int) map[string]interface{} {
	data := map[string]interface{}{}

	data["jsonrpc"] = "1.0"
	data["method"] = "createrawtransaction"

	params := []interface{}{}
	params = append(params, map[string]interface{}{"address_hash": addressHash})
	inTxs := []map[string]interface{}{}
	inTxs = append(inTxs, map[string]interface{}{
		"txid": txId,
		"vout": 0,
	})
	params = append(params, inTxs)
	outTxs := []map[string]interface{}{}
	outTxs = append(outTxs, map[string]interface{}{
		"pkScript":  pkScript,
		"value":     value,
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
