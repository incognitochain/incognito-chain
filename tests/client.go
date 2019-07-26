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

func (client *Client) getBlockChainInfo() (*jsonresult.GetBlockChainInfoResult, *rpcserver.RPCError) {
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

type ExampleReponse struct {
	F1 string
	F2 int
}
// example response => json result
func (client *Client) getExampleRpc(p1 string, p2 int) (result *ExampleReponse, err *rpcserver.RPCError) {
	res, rpcError := makeRPCRequest(client.Host, client.Port, getMethodName(), p1, p2)
	err = handleResponse(res.Result, rpcError, &result)
	return result, err
}

func handleResponse(resResult json.RawMessage, rpcError *rpcserver.RPCError, resultObj interface{}) *rpcserver.RPCError {
	if rpcError != nil {
		return rpcError
	}
	errUnMarshal := json.Unmarshal(resResult, resultObj)
	if errUnMarshal != nil {
		//TODO: unmarshal error
		return rpcserver.NewRPCError(rpcserver.ErrNetwork, errUnMarshal)
	}
	return nil
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
