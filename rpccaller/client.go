package rpccaller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type RPCClient struct {
	*http.Client
}

// NewHttpClient to get http client instance
func NewRPCClient() *RPCClient {
	httpClient := &http.Client{
		Timeout: time.Second * 60,
	}
	return &RPCClient{
		httpClient,
	}
}

func buildRPCServerAddress(protocol string, host string, port int) string {
	return fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

func (client *RPCClient) RPCCall(
	rpcProtocol string,
	rpcHost string,
	rpcPortStr string,
	method string,
	params interface{},
	rpcResponse interface{},
) (err error) {
	rpcPort, _ := strconv.Atoi(GetENV("RPC_PORT", rpcPortStr))
	rpcEndpoint := buildRPCServerAddress(rpcProtocol, rpcHost, rpcPort)

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}
	payloadInBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := client.Post(rpcEndpoint, "application/json", bytes.NewBuffer(payloadInBytes))
	if err != nil {
		return err
	}
	respBody := resp.Body
	defer respBody.Close()

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, rpcResponse)
	if err != nil {
		return err
	}
	return nil
}
