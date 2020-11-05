package devframework

import (
	"errors"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)
func (sim *LocalRPCClient) rpc_getbalancebyprivatekey(privateKey string) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.LimitedHttpHandler["getbalancebyprivatekey"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_getlistprivacycustomtokenbalance(privateKey string) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getlistprivacycustomtokenbalance"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_getcommitteelist(empty string) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getcommitteelist"]
	resI, rpcERR := c(httpServer, []interface{}{empty}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_getrewardamount(paymentAddress string) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getrewardamount"]
	resI, rpcERR := c(httpServer, []interface{}{paymentAddress}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_withdrawreward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["withdrawreward"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,amount,privacy,info}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendstakingtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendstakingtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,stakeInfo}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendstopautostakingtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendstopautostakingtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,stopStakeInfo}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendprivacycustomtokentransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendprivacycustomtokentransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,tokenInfo}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendtxwithwithdrawalreqv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithwithdrawalreqv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendtxwithpdefeewithdrawalreq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithpdefeewithdrawalreq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendtxwithptokentradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokentradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendtxwithptokencrosspooltradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokencrosspooltradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendtxwithprvtradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvtradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendtxwithprvcrosspooltradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvcrosspooltradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendtxwithptokencontributionv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokencontributionv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo,p1,pPrivacy}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_createandsendtxwithprvcontributionv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvcontributionv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy,reqInfo}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}
func (sim *LocalRPCClient) rpc_getpdestate(data map[string]interface{}) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getpdestate"]
	resI, rpcERR := c(httpServer, []interface{}{data}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}