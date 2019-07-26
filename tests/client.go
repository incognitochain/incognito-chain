package main

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Host string
	Port string
}

func makeRPCRequest(ip, port, method string, params interface{}) (*rpcserver.JsonResponse, *rpcserver.RPCError) {
	request := rpcserver.JsonRequest{
		Jsonrpc: "1.0",
		Method: method,
		Params: params,
		Id: "1",
	}
	requestBytes, err := json.Marshal(&request)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	resp, err := http.Post(ip+":"+port, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	body := resp.Body
	defer body.Close()
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	response := rpcserver.JsonResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return &response, nil
}

func (client *Client) getBlockChainInfo() (map[string]interface{}, *rpcserver.RPCError) {
	res, rpcError := makeRPCRequest(client.Host, client.Port, "getblockchaininfo", []string{})
	if rpcError != nil {
		return nil, rpcError
	}
	result := make(map[string]interface{})
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, res.Error
}