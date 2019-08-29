package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/incognitochain/incognito-chain/rpcserver"
)

var (
	flags = make(map[string]*string)
)

type Client struct {
	host string
	port string
	ws   string
}

func newClient() *Client {
	return &Client{}
}
func newClientWithHost(host, port string) *Client {
	return &Client{
		host: host,
		port: port,
	}
}
func newClientWithFullInform(host, port, ws string) *Client {
	return &Client{
		host: host,
		port: port,
		ws:   ws,
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
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	resp, err := http.Post(client.host+":"+client.port, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	body := resp.Body
	defer body.Close()
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	response := rpcserver.JsonResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	return &response, nil
}

func makeRPCRequestJson(client *Client, method string, params ...interface{}) (interface{}, *rpcserver.RPCError) {
	request := rpcserver.JsonRequest{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
		Id:      "1",
	}
	requestBytes, err := json.Marshal(&request)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	resp, err := http.Post("http://"+client.host+":"+client.port, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	body := resp.Body
	defer body.Close()
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	response := rpcserver.JsonResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	result := parseResult(response.Result)
	if result == nil {
		return result, rpcserver.NewRPCError(rpcserver.NetworkError, ParseFailedError)
	}
	return result, response.Error
}

func makeWsRequest(client *Client, method string, timeout time.Duration, params ...interface{}) (interface{}, *rpcserver.RPCError) {
	var done = make(chan struct{})
	var wsError error
	request := rpcserver.JsonRequest{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
		Id:      "1",
	}
	subcription := rpcserver.SubcriptionRequest{
		JsonRequest: request,
		Subcription: "0",
		Type:        0,
	}
	subcriptionBytes, err := json.Marshal(&subcription)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	var addr string
	if flag.Lookup("address:"+client.host+client.ws) != nil {
		addr = flag.Lookup("address:" + client.host + client.ws).Value.(flag.Getter).Get().(string)
	} else {
		addr = *flag.String("address:"+client.host+client.ws, client.host+":"+client.ws, "http service address")
	}
	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	defer conn.Close()
	err = conn.WriteMessage(websocket.BinaryMessage, subcriptionBytes)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	responseBytes := []byte{}
	go func() {
		defer close(done)
		for {
			_, reader, err := conn.NextReader()
			wsError = err
			if err != nil {
				return
			}
			responseChunk, err := ioutil.ReadAll(reader)
			responseBytes = append(responseBytes, responseChunk...)
			return

		}
	}()
	ticker := time.NewTicker(timeout)
loop:
	for {
		select {
		case <-ticker.C:
			{
				break loop
			}
		case <-done:
			{
				break loop
			}
		}
	}
	if wsError != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, wsError)
	}
	response := rpcserver.JsonResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	subResult := rpcserver.SubcriptionResult{}
	err = json.Unmarshal(response.Result, &subResult)
	if err != nil {
		return nil, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	result := parseResult(subResult.Result)
	if result == nil {
		return result, rpcserver.NewRPCError(rpcserver.NetworkError, ParseFailedError)
	}
	return result, response.Error
}
