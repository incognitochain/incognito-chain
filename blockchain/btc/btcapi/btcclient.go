package btcapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type BTCClient struct {
	User string
	Password string
	IP string
	Port string
}

func NewBTCClient(user string, password string, ip string, port string) *BTCClient {
	return &BTCClient{
		User: user,
		Password: password,
		IP: ip,
		Port: port,
	}
}

func (btcClient *BTCClient) GetBlockchainInfo() (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	var err error
	body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"getblockchaininfo\",\"params\":[]}")
	req, err := http.NewRequest("POST", "http://" + btcClient.IP + ":" + btcClient.Port, body)
	if err != nil {
		return nil, NewBTCAPIError(APIError, err)
	}
	req.SetBasicAuth(btcClient.User, btcClient.Password)
	req.Header.Set("Content-Type", "text/plain;")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, NewBTCAPIError(APIError, err)
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NewBTCAPIError(APIError, err)
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, NewBTCAPIError(APIError, err)
	}
	return result ,nil
}
