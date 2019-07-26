package tests

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"strings"
)

type Client struct {
	Host string
	Port string
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

type ExampleReponse struct {
	F1 string
	F2 int
}

func (client *Client) getExampleRpc(p1 string, p2 int) (result *ExampleReponse, err *rpcserver.RPCError) {
	res, rpcError := makeRPCRequest(client.Host, client.Port, getMethodName(), p1, p2)
	if rpcError != nil {
		return nil, rpcError
	}
	errUnMarshal := json.Unmarshal(res.Result, &result)
	if errUnMarshal != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, res.Error
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
