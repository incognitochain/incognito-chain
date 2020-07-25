package syncker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/incdb"
	"io/ioutil"
	"net/http"
)

//JsonRequest ...
type JsonRequest struct {
	Jsonrpc string      `json:"Jsonrpc"`
	Method  string      `json:"Method"`
	Params  interface{} `json:"Params"`
	Id      interface{} `json:"Id"`
}

type RPCError struct {
	Code       int    `json:"Code,omitempty"`
	Message    string `json:"Message,omitempty"`
	StackTrace string `json:"StackTrace"`

	err error `json:"Err"`
}

type JsonResponse struct {
	Id      *interface{}    `json:"Id"`
	Result  json.RawMessage `json:"Result"`
	Error   *RPCError       `json:"Error"`
	Params  interface{}     `json:"Params"`
	Method  string          `json:"Method"`
	Jsonrpc string          `json:"Jsonrpc"`
}

func makeRPCRequest(address string, method string, params ...interface{}) (*JsonResponse, error) {
	request := JsonRequest{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
		Id:      "1",
	}
	requestBytes, err := json.Marshal(&request)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(address, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	body := resp.Body
	defer body.Close()
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	response := JsonResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

//preloadDatabase call to backuped database node ...
func preloadDatabase(chainID int, currentEpoch int, url string, db incdb.Database) error {
	chainName := "beacon"
	if chainID > -1 {
		chainName = fmt.Sprintf("shard%v", chainID)
	}
	response, err := makeRPCRequest(url, "getlatestbackup", chainName)
	if err != nil {
		return err
	}
	type LatestEpochResult struct {
		LatestEpoch int
	}
	result := LatestEpochResult{}
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return err
	}

	if currentEpoch < result.LatestEpoch-10 {
		backupFile := ""

		db.Close()
		defer db.ReOpen()

		err := db.PreloadBackup(backupFile)
		if err != nil {
			return err
		}

	}
	return nil
}
