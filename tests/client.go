package main

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"strings"
)

type Client struct {
	Host string `json:"host"`
	Port string `json:"port"`
}
func getMethodName(depthList ...int) string {
	var depth int
	if depthList == nil {
		depth = 1
	} else {
		depth = depthList[0]
	}
	function, _, _, _ := runtime.Caller(depth)
	r, _ := regexp.Compile("\\.(.*)")
	return strings.ToLower(r.FindStringSubmatch(runtime.FuncForPC(function).Name())[1])
}
func makeRPCRequest(ip, port, method string, params ...interface{}) (*rpcserver.JsonResponse, *rpcserver.RPCError) {
	request := rpcserver.JsonRequest{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
		Id:      "1",
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

func (client *Client) getBlockChainInfo(params ...interface{}) (interface{}, *rpcserver.RPCError) {
	result := &jsonresult.GetBlockChainInfoResult{}
	res, rpcError := makeRPCRequest(client.Host, client.Port, getMethodName(), []string{})
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, res.Error
}

func (client *Client) createAndSendTransaction(params ...interface{}) (interface{}, *rpcserver.RPCError) {
	result := &jsonresult.CreateTransactionResult{}
	res, rpcError := makeRPCRequest(client.Host, client.Port, getMethodName(), params)
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, nil
}


