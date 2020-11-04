package devrpcclient

import "github.com/incognitochain/incognito-chain/rpcserver/jsonresult"

type ClientInterface interface {
	RPC_createtransaction(privatekey string, receivers map[string]interface{}, fee float64, privacy float64) (jsonresult.CreateTransactionResult, error)
	RPC_getbalancebyprivatekey(privatekey string) (uint64, error)
	RPC_getlistprivacycustomtokenbalance(privatekey string) (jsonresult.ListCustomTokenBalance, error)
	RPC_getrewardamount(paymentaddress string) (map[string]uint64, error)
	RPC_withdrawreward(privatekey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendstakingtransaction(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, stakeinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendstopautostakingtransaction(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, stopstakeinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_getcommitteelist(empty string) (jsonresult.CommitteeListsResult, error)
	RPC_createandsendprivacycustomtokentransaction(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, tokeninfo map[string]interface{}) (jsonresult.CreateTransactionTokenResult, error)
	RPC_createandsendtxwithwithdrawalreqv2(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendtxwithpdefeewithdrawalreq(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendtxwithptokentradereq(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}, p1 string, pprivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	RPC_createandsendtxwithptokencrosspooltradereq(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}, p1 string, pprivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	RPC_createandsendtxwithprvtradereq(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendtxwithprvcrosspooltradereq(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendtxwithptokencontributionv2(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}, p1 string, pprivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	RPC_createandsendtxwithprvcontributionv2(privatekey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_getpdestate(data map[string]interface{}) (interface{}, error)
}

type RPCWRAPPER struct {
	client ClientInterface
}

func NewRPCClient(client ClientInterface) *RPCWRAPPER {
	rpc := &RPCWRAPPER{
		client: client,
	}
	return rpc
}

func (r *RPCWRAPPER) API_CreateAndSendPrivacyCustomTokenTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := r.client.RPC_createandsendprivacycustomtokentransaction(privateKey, receivers, fee, privacy, tokenInfo)
	return &result, err
}
func (r *RPCWRAPPER) API_CreateAndSendTxWithWithdrawalReqV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithwithdrawalreqv2(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCWRAPPER) API_CreateAndSendTxWithPDEFeeWithdrawalReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithpdefeewithdrawalreq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCWRAPPER) API_CreateAndSendTxWithPTokenTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := r.client.RPC_createandsendtxwithptokentradereq(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (r *RPCWRAPPER) API_CreateAndSendTxWithPTokenCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := r.client.RPC_createandsendtxwithptokencrosspooltradereq(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (r *RPCWRAPPER) API_CreateAndSendTxWithPRVTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithprvtradereq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCWRAPPER) API_CreateAndSendTxWithPRVCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithprvcrosspooltradereq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCWRAPPER) API_CreateAndSendTxWithPTokenContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := r.client.RPC_createandsendtxwithptokencontributionv2(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (r *RPCWRAPPER) API_CreateAndSendTxWithPRVContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithprvcontributionv2(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCWRAPPER) API_GetPDEState(beaconHeight float64) (interface{}, error) {
	result, err := r.client.RPC_getpdestate(map[string]interface{}{"BeaconHeight": beaconHeight})
	return &result, err
}
