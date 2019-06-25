package btc

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type BTCClient struct {
	User     string
	Password string
	IP       string
	Port     string
}

func NewBTCClient(user string, password string, ip string, port string) *BTCClient {
	return &BTCClient{
		User:     user,
		Password: password,
		IP:       ip,
		Port:     port,
	}
}
func (btcClient *BTCClient) GetNonceByTimestamp(timestamp int64) (int, int64, int64, error) {
	var (
		chainHeight    int
		chainTimestamp int64
		nonce          int64
		err            error
	)
	chainHeight, chainTimestamp, nonce, err = btcClient.GetChainTimeStampAndNonce()
	if err != nil {
		return 0, 0, -1, err
	}
	blockHeight, err := estimateBlockHeight(btcClient, timestamp, chainHeight, chainTimestamp)
	if err != nil {
		return 0, 0, -1, err
	}
	_, blockTimestamp, err = btcClient.GetTimeStampAndNonceByBlockHeight(blockHeight)
	if err != nil {
		return 0, 0, -1, err
	}
	if blockTimestamp == MAX_TIMESTAMP {
		return 0, 0, -1, NewBTCAPIError(APIError, errors.New("Can't get result from API"))
	}
	if blockTimestamp > timestamp {
		for blockTimestamp > timestamp {
			blockHeight--
			_, blockTimestamp, err = btcClient.GetTimeStampAndNonceByBlockHeight(blockHeight)
			if err != nil {
				return 0, 0, -1, err
			}
			if blockTimestamp == MAX_TIMESTAMP {
				return 0, 0, -1, NewBTCAPIError(APIError, errors.New("Can't get result from API"))
			}
			if blockTimestamp <= timestamp {
				blockHeight++
				break
			}
		}
	} else {
		for blockTimestamp <= timestamp {
			blockHeight++
			if blockHeight > chainHeight {
				return 0, 0, -1, NewBTCAPIError(APIError, errors.New("Timestamp is greater than timestamp of highest block"))
			}
			_, blockTimestamp, err = btcClient.GetTimeStampAndNonceByBlockHeight(blockHeight)
			if err != nil {
				return 0, 0, -1, err
			}
			if blockTimestamp == MAX_TIMESTAMP {
				return 0, 0, -1, NewBTCAPIError(APIError, errors.New("Can't get result from API"))
			}
			if blockTimestamp > timestamp {
				break
			}
		}
	}
	timestamp, nonce, err = btcClient.GetTimeStampAndNonceByBlockHeight(blockHeight)
	if err != nil {
		return 0, 0, -1, err
	}
	return blockHeight, timestamp, nonce, nil
}
func (btcClient *BTCClient) VerifyNonceWithTimestamp(timestamp int64, nonce int64) (bool, error) {
	_, _, tempNonce, err := btcClient.GetNonceByTimestamp(timestamp)
	if err != nil {
		return false, err
	}
	return tempNonce == nonce, nil
}
func (btcClient *BTCClient) GetCurrentChainTimeStamp() (int64, error) {
	_, timestamp, _, err := btcClient.GetChainTimeStampAndNonce()
	if err != nil {
		return -1, err
	}
	return timestamp, nil
}

func (btcClient *BTCClient) GetBlockchainInfo() (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	var err error
	//body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"getblockchaininfo\",\"params\":[]}")
	//req, err := http.NewRequest("POST", "http://"+btcClient.IP+":"+btcClient.Port, body)
	//if err != nil {
	//	return nil, NewBTCAPIError(APIError, err)
	//}
	//req.SetBasicAuth(btcClient.User, btcClient.Password)
	//req.Header.Set("Content-Type", "text/plain;")
	//
	//resp, err := http.DefaultClient.Do(req)
	//if err != nil {
	//	return nil, NewBTCAPIError(APIError, err)
	//}
	//defer resp.Body.Close()
	//response, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return nil, NewBTCAPIError(APIError, err)
	//}
	//err = json.Unmarshal(response, &result)
	//if err != nil {
	//	return nil, NewBTCAPIError(APIError, err)
	//}
	result, err = btcClient.callRPC("getblockchaininfo", "")
	if err != nil {
		return result, err
	}
	return result, nil
}

/*

 */
func (btcClient *BTCClient) GetBestBlockHeight() (int, error) {
	var result = make(map[string]interface{})
	var err error
	//body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"getblockcount\",\"params\":[]}")
	//req, err := http.NewRequest("POST", "http://"+btcClient.IP+":"+btcClient.Port, body)
	//if err != nil {
	//	return -1, NewBTCAPIError(APIError, err)
	//}
	//req.SetBasicAuth(btcClient.User, btcClient.Password)
	//req.Header.Set("Content-Type", "text/plain;")
	//
	//resp, err := http.DefaultClient.Do(req)
	//if err != nil {
	//	return -1, NewBTCAPIError(APIError, err)
	//}
	//defer resp.Body.Close()
	//response, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return -1, NewBTCAPIError(APIError, err)
	//}
	//err = json.Unmarshal(response, &result)
	//if err != nil {
	//	return -1, NewBTCAPIError(APIError, err)
	//}
	result, err = btcClient.callRPC("getblockcount", "")
	if err != nil {
		return -1, err
	}
	blockHeight := int(result["result"].(float64))
	return blockHeight, nil
}

/*
curl --user admin:autonomous --data-binary '{"jsonrpc":"1.0","id":"curltext","method":"getblock","params":["000000000000000000210a7be63100bf18ccd43ea8c3e536c476d8d339baa1d9"]}' -H 'content-type:text/plain;' http://159.65.142.153:8332
#return param1: chain height
#return param2: timestamp
#return param3: nonce
*/
func (btcClient *BTCClient) GetChainTimeStampAndNonce() (int, int64, int64, error) {
	res, err := btcClient.GetBlockchainInfo()
	if err != nil {
		return -1, -1, -1, err
	}
	bestBlockHash := res["result"].(map[string]interface{})["bestblockhash"].(string)
	bestBlockHeight := res["result"].(map[string]interface{})["blocks"].(float64)
	timestamp, nonce, err := btcClient.GetTimeStampAndNonceByBlockHash(bestBlockHash)
	return int(bestBlockHeight), timestamp, nonce, err

}
func (btcClient *BTCClient) GetTimeStampAndNonceByBlockHash(blockHash string) (int64, int64, error) {
	var err error
	var result = make(map[string]interface{})
	//body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"getblockheader\",\"params\":[\"" + blockHash + "\"]}")
	//req, err := http.NewRequest("POST", "http://"+btcClient.IP+":"+btcClient.Port, body)
	//if err != nil {
	//	return -1, -1, NewBTCAPIError(APIError, err)
	//}
	//req.SetBasicAuth(btcClient.User, btcClient.Password)
	//req.Header.Set("Content-Type", "text/plain;")
	//
	//resp, err := http.DefaultClient.Do(req)
	//if err != nil {
	//	return -1, -1, NewBTCAPIError(APIError, err)
	//}
	//defer resp.Body.Close()
	//response, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return -1, -1, NewBTCAPIError(APIError, err)
	//}
	//err = json.Unmarshal(response, &result)
	//if err != nil {
	//	return -1, -1, NewBTCAPIError(APIError, err)
	//}
	result, err = btcClient.callRPC("getblockheader", blockHash)
	if err != nil {
		return -1, -1, err
	}
	timeStamp := result["result"].(map[string]interface{})["time"].(float64)
	nonce := result["result"].(map[string]interface{})["nonce"].(float64)
	return int64(timeStamp), int64(nonce), nil
}
func (btcClient *BTCClient) GetTimeStampAndNonceByBlockHeight(blockHeight int) (int64, int64, error) {
	blockHash, err := btcClient.GetBlockHashByHeight(blockHeight)
	if err != nil {
		return -1, -1, err
	}
	return btcClient.GetTimeStampAndNonceByBlockHash(blockHash)
}

func (btcClient *BTCClient) GetBlockHashByHeight(blockHeight int) (string, error) {
	var err error
	var result = make(map[string]interface{})
	//body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"getblockhash\",\"params\":[" + strconv.Itoa(blockHeight) + "]}")
	//req, err := http.NewRequest("POST", "http://"+btcClient.IP+":"+btcClient.Port, body)
	//if err != nil {
	//	return common.EmptyString, NewBTCAPIError(APIError, err)
	//}
	//req.SetBasicAuth(btcClient.User, btcClient.Password)
	//req.Header.Set("Content-Type", "text/plain;")
	//
	//resp, err := http.DefaultClient.Do(req)
	//if err != nil {
	//	return common.EmptyString, NewBTCAPIError(APIError, err)
	//}
	//defer resp.Body.Close()
	//response, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return common.EmptyString, NewBTCAPIError(APIError, err)
	//}
	//err = json.Unmarshal(response, &result)
	//if err != nil {
	//	return common.EmptyString, NewBTCAPIError(APIError, err)
	//}
	result, err = btcClient.callRPC("getblockhash", strconv.Itoa(blockHeight))
	if err != nil {
		return common.EmptyString, err
	}
	blockHash := result["result"].(string)
	return blockHash, nil
}

func (btcClient *BTCClient) callRPC(method string, params string) (map[string]interface{}, error) {
	var err error
	var result = make(map[string]interface{})
	body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"" + method + "\",\"params\":[" + params + "]}")
	req, err := http.NewRequest("POST", "http://"+btcClient.IP+":"+btcClient.Port, body)
	if err != nil {
		return result, NewBTCAPIError(APIError, err)
	}
	req.SetBasicAuth(btcClient.User, btcClient.Password)
	req.Header.Set("Content-Type", "text/plain;")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, NewBTCAPIError(APIError, err)
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, NewBTCAPIError(APIError, err)
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return result, NewBTCAPIError(APIError, err)
	}
	return result, nil
}
