package devframework //This file is auto generated. Please do not change if you dont know what you are doing
import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type RemoteRPCClient struct {
	endpoint string
}

type ErrMsg struct {
	Code       int
	Message    string
	StackTrace string
}

func (r *RemoteRPCClient) sendRequest(requestBody []byte) ([]byte, error) {
	resp, err := http.Post(r.endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}


func (r *RemoteRPCClient) GetBalanceByPrivateKey(privateKey string) (res uint64,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbalancebyprivatekey",
		"params":   []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  uint64
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) GetListPrivacyCustomTokenBalance(privateKey string) (res jsonresult.ListCustomTokenBalance,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getlistprivacycustomtokenbalance",
		"params":   []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.ListCustomTokenBalance
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) GetCommitteeList(empty string) (res jsonresult.CommitteeListsResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getcommitteelist",
		"params":   []interface{}{empty},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CommitteeListsResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) GetRewardAmount(paymentAddress string) (res map[string]uint64,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getrewardamount",
		"params":   []interface{}{paymentAddress},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  map[string]uint64
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) WithdrawReward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "withdrawreward",
		"params":   []interface{}{privateKey,receivers,amount,privacy,info},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendstakingtransaction",
		"params":   []interface{}{privateKey,receivers,fee,privacy,stakeInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendStopAutoStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendstopautostakingtransaction",
		"params":   []interface{}{privateKey,receivers,fee,privacy,stopStakeInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtransaction",
		"params":   []interface{}{privateKey,receivers,fee,privacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendPrivacyCustomTokenTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendprivacycustomtokentransaction",
		"params":   []interface{}{privateKey,receivers,fee,privacy,tokenInfo,p1,pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionTokenResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTxWithWithdrawalReqV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithwithdrawalreqv2",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTxWithPDEFeeWithdrawalReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithpdefeewithdrawalreq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTxWithPTokenTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokentradereq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionTokenResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTxWithPTokenCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokencrosspooltradereq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionTokenResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTxWithPRVTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvtradereq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTxWithPRVCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvcrosspooltradereq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTxWithPTokenContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokencontributionv2",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionTokenResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) CreateAndSendTxWithPRVContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvcontributionv2",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) GetPDEState(data map[string]interface{}) (res jsonresult.CurrentPDEState,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getpdestate",
		"params":   []interface{}{data},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CurrentPDEState
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) GetBeaconBestState() (res jsonresult.GetBeaconBestState,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbeaconbeststate",
		"params":   []interface{}{},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.GetBeaconBestState
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) GetShardBestState(sid int) (res jsonresult.GetShardBestState,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getshardbeststate",
		"params":   []interface{}{sid},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.GetShardBestState
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) GetTransactionByHash(transactionHash string) (res *jsonresult.TransactionDetail,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "gettransactionbyhash",
		"params":   []interface{}{transactionHash},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  *jsonresult.TransactionDetail
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (r *RemoteRPCClient) GetPrivacyCustomToken(tokenStr string) (res *jsonresult.GetCustomToken,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getprivacycustomtoken",
		"params":   []interface{}{tokenStr},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  *jsonresult.GetCustomToken
		Error *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}
