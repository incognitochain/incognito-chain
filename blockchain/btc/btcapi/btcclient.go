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
/*

*/
func (btcClient *BTCClient) GetBestBlockHeight() (int, error) {
	var result = make(map[string]interface{})
	var err error
	body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"getblockcount\",\"params\":[]}")
	req, err := http.NewRequest("POST", "http://" + btcClient.IP + ":" + btcClient.Port, body)
	if err != nil {
		return -1, NewBTCAPIError(APIError, err)
	}
	req.SetBasicAuth(btcClient.User, btcClient.Password)
	req.Header.Set("Content-Type", "text/plain;")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, NewBTCAPIError(APIError, err)
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, NewBTCAPIError(APIError, err)
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return -1, NewBTCAPIError(APIError, err)
	}
	blockHeight := int(result["result"].(float64))
	return blockHeight,nil
}
/*
curl --user admin:autonomous --data-binary '{"jsonrpc":"1.0","id":"curltext","method":"getblock","params":["000000000000000000210a7be63100bf18ccd43ea8c3e536c476d8d339baa1d9"]}' -H 'content-type:text/plain;' http://159.65.142.153:8332
*/
func (btcClient *BTCClient) GetChainTimeStamp() (int64, error) {
	var err error
	var result = make(map[string]interface{})
	res, err := btcClient.GetBlockchainInfo()
	if err != nil {
		return -1, err
	}
	bestBlockHash := res["result"].(map[string]interface{})["bestblockhash"].(string)
	body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"getblock\",\"params\":[\""+bestBlockHash+"\"]}")
	req, err := http.NewRequest("POST", "http://" + btcClient.IP + ":" + btcClient.Port, body)
	if err != nil {
		return -1, NewBTCAPIError(APIError, err)
	}
	req.SetBasicAuth(btcClient.User, btcClient.Password)
	req.Header.Set("Content-Type", "text/plain;")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, NewBTCAPIError(APIError, err)
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, NewBTCAPIError(APIError, err)
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return -1, NewBTCAPIError(APIError, err)
	}
	timeStamp := result["result"].(map[string]interface{})["time"].(float64)
	return int64(timeStamp), nil
}
