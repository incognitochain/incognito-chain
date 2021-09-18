package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type JsonRpcFormat struct {
	ID      string        `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type Response struct {
	Result interface{} `json:"Result"`
	Errror interface{} `json:"Error"`
}

func sendHttpRequest(url, method string, params []interface{}, isToConsole bool) (interface{}, error) {
	payload := &JsonRpcFormat{
		ID:      "1",
		JsonRpc: "jsonrpc",
		Method:  method,
		Params:  params,
	}
	payloadData, err := json.Marshal(payload)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadData))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if isToConsole {
		log.Println(string(body))
	}
	response := Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response.Result, nil
}
