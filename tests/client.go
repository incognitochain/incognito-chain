package tests

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Ip string
	Port string
}

func makeRPCRequest(ip, port, method string, params interface{}) (*rpcserver.JsonResponse, error) {
	request := rpcserver.JsonRequest{
		Jsonrpc: "1.0",
		Method: method,
		Params: params,
		Id: "1",
	}
	requestBytes, err := json.Marshal(&request)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(ip+":"+port, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	body := resp.Body
	defer body.Close()
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	response := rpcserver.JsonResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
