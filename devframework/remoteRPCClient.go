package devframework
import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func (sim *RemoteRPCClient) rpc_getbalancebyprivatekey(privateKey string) (res uint64,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbalancebyprivatekey",
		"params":   []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  uint64
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_getlistprivacycustomtokenbalance(privateKey string) (res jsonresult.ListCustomTokenBalance,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getlistprivacycustomtokenbalance",
		"params":   []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.ListCustomTokenBalance
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_getcommitteelist(empty string) (res jsonresult.CommitteeListsResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getcommitteelist",
		"params":   []interface{}{empty},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CommitteeListsResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_getrewardamount(paymentAddress string) (res map[string]uint64,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getrewardamount",
		"params":   []interface{}{paymentAddress},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  map[string]uint64
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_withdrawreward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "withdrawreward",
		"params":   []interface{}{privateKey,receivers,amount,privacy,info},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendstakingtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendstakingtransaction",
		"params":   []interface{}{privateKey,receivers,fee,privacy,stakeInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendstopautostakingtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendstopautostakingtransaction",
		"params":   []interface{}{privateKey,receivers,fee,privacy,stopStakeInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createtransaction",
		"params":   []interface{}{privateKey,receivers,fee,privacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendprivacycustomtokentransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}) (res jsonresult.CreateTransactionTokenResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendprivacycustomtokentransaction",
		"params":   []interface{}{privateKey,receivers,fee,privacy,tokenInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionTokenResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendtxwithwithdrawalreqv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithwithdrawalreqv2",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendtxwithpdefeewithdrawalreq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithpdefeewithdrawalreq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendtxwithptokentradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokentradereq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionTokenResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendtxwithptokencrosspooltradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokencrosspooltradereq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionTokenResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendtxwithprvtradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvtradereq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendtxwithprvcrosspooltradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvcrosspooltradereq",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendtxwithptokencontributionv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokencontributionv2",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionTokenResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_createandsendtxwithprvcontributionv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvcontributionv2",
		"params":   []interface{}{privateKey,receivers,fee,privacy,reqInfo},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}


func (sim *RemoteRPCClient) rpc_getpdestate(data map[string]interface{}) (res interface{},err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getpdestate",
		"params":   []interface{}{data},
		"id":      1,
	})
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	resp := struct {
		Result  interface{}
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resp.Result,err
}
