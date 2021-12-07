package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func main() {

	ticker := time.NewTicker(10 * time.Second)
	for _ = range ticker.C {
		url := "http://localhost:9355"
		method := "POST"

		payload := strings.NewReader(`{
	"jsonrpc": "1.0",
    "method": "getconsensusdata",
    "params": [0],
    "id": 1
}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("x-api-key", "MwUBtRBWcH8kDr9m40Y027Rt6GyqjOpC73iioXTf")
		req.Header.Add("X-Amz-Content-Sha256", "beaead3198f7da1e70d03ab969765e0821b24fc913697e929e726aeaebf0eba3")
		req.Header.Add("X-Amz-Date", "20211119T104927Z")
		req.Header.Add("Authorization", "AWS4-HMAC-SHA256 Credential=AKIAIFNQPBGPLTLC2AKA/20211119/us-east-2c/execute-api/aws4_request, SignedHeaders=content-type;host;x-amz-content-sha256;x-amz-date;x-api-key, Signature=97f3d7f3c6f31b31512c5815dbc906a0baa2fed499a638efd9d57e94fb00bf6c")

		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		tempRes := make(map[string]interface{})
		json.Unmarshal(body, &tempRes)
		result := tempRes["Result"].(map[string]interface{})
		jsonResult, _ := json.Marshal(result["voteHistory"])
		fmt.Println(string(jsonResult))
		fmt.Println("---------------------------------")
	}
}
