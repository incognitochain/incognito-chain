package devframework
//This file is auto generated. Please do not change if you dont know what you are doing
import (
	"errors"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)
type LocalRPCClient struct {
	rpcServer *rpcserver.RpcServer
}
func (r *LocalRPCClient) RPC_getbalancebyprivatekey(privateKey string) (res uint64,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.LimitedHttpHandler["getbalancebyprivatekey"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(uint64),nil
}
func (r *LocalRPCClient) RPC_getlistprivacycustomtokenbalance(privateKey string) (res jsonresult.ListCustomTokenBalance,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getlistprivacycustomtokenbalance"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.ListCustomTokenBalance),nil
}
func (r *LocalRPCClient) RPC_getcommitteelist(empty string) (res jsonresult.CommitteeListsResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getcommitteelist"]
	resI, rpcERR := c(httpServer, []interface{}{empty}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CommitteeListsResult),nil
}
func (r *LocalRPCClient) RPC_getrewardamount(paymentAddress string) (res map[string]uint64,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getrewardamount"]
	resI, rpcERR := c(httpServer, []interface{}{paymentAddress}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(map[string]uint64),nil
}
func (r *LocalRPCClient) RPC_withdrawreward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["withdrawreward"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,amount,privacy,info}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_createandsendstakingtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendstakingtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,stakeInfo}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_createandsendstopautostakingtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendstopautostakingtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,stopStakeInfo}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_createandsendprivacycustomtokentransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendprivacycustomtokentransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,tokenInfo,p1,pPrivacy}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtxwithwithdrawalreqv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithwithdrawalreqv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtxwithpdefeewithdrawalreq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithpdefeewithdrawalreq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtxwithptokentradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokentradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtxwithptokencrosspooltradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokencrosspooltradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtxwithprvtradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvtradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtxwithprvcrosspooltradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvcrosspooltradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtxwithptokencontributionv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokencontributionv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult),nil
}
func (r *LocalRPCClient) RPC_createandsendtxwithprvcontributionv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvcontributionv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (r *LocalRPCClient) RPC_getpdestate(data map[string]interface{}) (res jsonresult.CurrentPDEState,err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getpdestate"]
	resI, rpcERR := c(httpServer, []interface{}{data}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CurrentPDEState),nil
}