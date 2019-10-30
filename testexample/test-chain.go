package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type user struct {
	privatekey, paymentaddress string
}

func createAndSendTransaction(RPC string, privatekey string, paymentaddress string, amount uint64) (interface{}, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtransaction",
		"id":      1,
		"params":  []interface{}{privatekey, map[string]uint64{paymentaddress: amount}, -1, 0},
	})

	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error parsing request body", requestBody, err)
		return requestBody, err
	}

	resp, err := http.Post(RPC, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error sending json rpc", resp, err)
		return resp, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error parsing response body", string(body), err)
	}

	// log.Println(string(body))
	var data map[string]interface{}
	errr := json.Unmarshal([]byte(body), &data)
	if errr != nil {
		panic(errr)
	}
	result := data["Result"].(map[string]interface{})
	return result["TxID"], err
}

func getBalanceByPrivateKey(RPC string, privatekey string) (uint64, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbalancebyprivatekey",
		"id":      1,
		"params":  []interface{}{privatekey},
	})

	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error parsing request body", requestBody, err)
		return 0, err
	}

	resp, err := http.Post(RPC, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error sending json rpc", resp, err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error parsing response body", string(body), err)
	}

	var data map[string]interface{}
	errr := json.Unmarshal([]byte(body), &data)
	if errr != nil {
		fmt.Println(string(body))
		panic(errr)
	}
	balance := data["Result"].(float64)
	return uint64(balance), errr
}

func getTransactionFeeByHash(RPC string, txID string) (uint64, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "gettransactionbyhash",
		"id":      1,
		"params":  []interface{}{txID},
	})

	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error parsing request body", requestBody, err)
		return 0, err
	}

	resp, err := http.Post(RPC, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error sending json rpc", resp, err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		fmt.Println("ERROR: error parsing response body", string(body), err)
	}

	var data map[string]interface{}
	errr := json.Unmarshal([]byte(body), &data)
	if errr != nil {
		panic(errr)
	}
	result := data["Result"].(map[string]interface{})
	fee := result["Fee"].(float64)
	return uint64(fee), err
}

func main() {
	/* TEST SETUP */
	user1 := user{
		privatekey:     "112t8rnX5E2Mkqywuid4r4Nb2XTeLu3NJda43cuUM1ck2brpHrufi4Vi42EGybFhzfmouNbej81YJVoWewJqbR4rPhq2H945BXCLS2aDLBTA",
		paymentaddress: "12RxERBySmquLtM1R1Dk2s7J4LyPxqHxcZ956kupQX3FPhVo2KtoUYJWKet2nWqWqSh3asWmgGTYsvz3jX73HqD8Jr2LwhjhJfpG756",
	}
	user2 := user{
		privatekey:     "112t8rnXWRThUTJQgoyH6evV8w19dFZfKWpCh8rZpfymW9JTgKPEVQS44nDRPpsooJiGStHxu81m3HA84t9DBVobz8hgBKRMcz2hddPWNX9N",
		paymentaddress: "12S2x6SHiah9GToSvwXzbDeBrJzhPkENLJosgozv7AQE55xrEkVQqD95fTyGf6xt69PD4oxZ6xZ5qaPcVQAWqFjEt5TQ4cgimBgW2j2",
	}
	var amount uint64 = 5000000000
	// var RPCServer string = "https://test-node.incognito.org"
	var RPCServer string = "http://localhost:9334"

	/* TEST CASE STARTED HERE */
	// Step 1: Check balance of address 1 and address 2 before sending PRV
	fmt.Println("// Step 1: Check balance of address 1 and address 2 before sending PRV")
	balance1Before, err := getBalanceByPrivateKey(RPCServer, user1.privatekey)
	if err != nil {
		fmt.Println("Error go here: ", err)
	} else {
		fmt.Println("Get Balance Address-1:", balance1Before)

	}

	balance2Before, err := getBalanceByPrivateKey(RPCServer, user2.privatekey)
	if err != nil {
		fmt.Println("Error go here: ", err)
	} else {
		fmt.Println("Get Balance Address-2:", balance2Before)

	}

	// Step 2: Sending PRV from address 1 to address 2
	fmt.Println("// Step 2: Sending PRV from address 1 to address 2")
	fmt.Println("-Sending ", amount)
	sendTX, err := createAndSendTransaction(RPCServer, user1.privatekey, user2.paymentaddress, amount)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("TxID: ", sendTX)
	}

	fmt.Println("-Sleep 10 seconds, wait for balance update")
	time.Sleep(10 * time.Second)

	// Step 3: Get Transaction Fee
	fmt.Println("// Step 3: Get Transaction Fee")
	txID := sendTX.(string)
	fee, err := getTransactionFeeByHash(RPCServer, txID)
	if err != nil {
		fmt.Println("Error go here: ", err)
	} else {
		fmt.Println("Transaction Fee: ", fee)
	}

	// Step 4: Check balance of address 1 and address 2 after sent PRV
	fmt.Println("// Step 3: Check balance of address 1 and address 2 after sent PRV")

	balance1After, err := getBalanceByPrivateKey(RPCServer, user1.privatekey)
	if err != nil {
		fmt.Println("Error go here: ", err)
	} else {
		fmt.Println("Get Balance Address-1:", balance1After)
	}

	balance2After, err := getBalanceByPrivateKey(RPCServer, user2.privatekey)
	if err != nil {
		fmt.Printf("Error go here: %v", err)
	} else {
		fmt.Println("Get Balance Address-2:", balance2After)

	}

	/* TEST RESULT */
	fmt.Println("// TEST RESULT:")
	if (balance1After+amount+fee != balance1Before) || (balance2After-amount != balance2Before) {
		fmt.Println("FAILED")
	} else {
		fmt.Println("PASSED")
	}

	/* TEST CLEAN UP*/
	fmt.Println("// TEST CLEAN UP")
	cleanupTX, err := createAndSendTransaction(RPCServer, user2.privatekey, user1.paymentaddress, balance2After-2) // default tx fee = 2
	if err != nil {
		fmt.Print("Clean up error: ", err)
	} else {
		fmt.Println("Clean up success ", cleanupTX)
	}

}
