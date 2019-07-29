package main

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Client struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

func newClient() *Client {
	return &Client{}
}
func newClientWithHost(host, port string) *Client {
	return &Client{
		Host: host,
		Port: port,
	}
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
func makeRPCRequest(client *Client, method string, params ...interface{}) (*rpcserver.JsonResponse, *rpcserver.RPCError) {
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
	resp, err := http.Post(client.Host+":"+client.Port, "application/json", bytes.NewBuffer(requestBytes))
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

func makeRPCRequestV2(client *Client, method string, params ...interface{}) (map[string]interface{}, *rpcserver.RPCError) {
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
	resp, err := http.Post(client.Host+":"+client.Port, "application/json", bytes.NewBuffer(requestBytes))
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
	result := make(map[string]interface{})
	rpcError := json.Unmarshal(response.Result, &result)
	if rpcError != nil {
		return result, rpcserver.NewRPCError(rpcserver.ErrNetwork, rpcError)
	}
	return result, response.Error
}

func makeWsRequest(client *Client, method string, timeout time.Duration, params ...interface{}) (map[string]interface{}, *rpcserver.RPCError) {
	var cQuit = make(chan struct{})
	request := rpcserver.JsonRequest{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
		Id:      "1",
	}
	subcription := rpcserver.SubcriptionRequest{
		JsonRequest: request,
		Subcription: "0",
		Type: 0,
	}
	subcriptionBytes, err := json.Marshal(&subcription)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	conn, err := net.Dial("tcp", client.Host+":"+client.Port)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	defer conn.Close()
	_, err = conn.Write(subcriptionBytes)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	responseBytes := []byte{}
	go func(cQuit chan struct{}) {
		<-time.Tick(timeout)
		close(cQuit)
	}(cQuit)
	loop:
	for {
		select {
			case <-cQuit:
				break loop
			default:
			_, err = conn.Read(responseBytes)
		}
	}
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	response := rpcserver.JsonResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	result := make(map[string]interface{})
	rpcError := json.Unmarshal(response.Result, &result)
	if rpcError != nil {
		return result, rpcserver.NewRPCError(rpcserver.ErrNetwork, rpcError)
	}
	return result, response.Error
}

func (client *Client) getBlockChainInfo(params ...interface{}) (interface{}, *rpcserver.RPCError) {
	//result := &jsonresult.GetBlockChainInfoResult{}
	result := make(map[string]interface{})
	res, rpcError := makeRPCRequest(client, getBlockChainInfo, []string{})
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, res.Error
}

func (client *Client) createAndSendTransaction(params ...interface{}) (interface{}, *rpcserver.RPCError) {
	//result := &jsonresult.CreateTransactionResult{}
	result := make(map[string]interface{})
	res, rpcError := makeRPCRequest(client, createAndSendTransaction, params)
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, nil
}

func (client *Client) getBalanceByPrivatekey(params ...interface{}) (interface{}, *rpcserver.RPCError) {
	result := make(map[string]interface{})
	res, rpcError := makeRPCRequest(client, getBalanceByPrivatekey, params)
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, nil
}
func (client *Client) getTransactionByHash(params ...interface{}) (interface{}, *rpcserver.RPCError) {
	result := make(map[string]interface{})
	res, rpcError := makeRPCRequest(client, getTransactionByHash, params)
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, nil
}
