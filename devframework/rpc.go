package devframework

import (
	"errors"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)
func (sim *SimulationEngine) rpc_createtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (res jsonresult.CreateTransactionResult,err error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey,receivers,fee,privacy}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult),nil
}
func (sim *SimulationEngine) rpc_getbalancebyprivatekey(privateKey string) (res uint64,err error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.LimitedHttpHandler["getbalancebyprivatekey"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey}, nil)
	if rpcERR != nil {
		return res,errors.New(rpcERR.Error())
	}
	return resI.(uint64),nil
}
func (sim *SimulationEngine) rpc_getrewardamount(paymentAddress string) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getrewardamount"]
	resI, rpcERR := c(httpServer, []interface{}{paymentAddress}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	_ = resI 
 return nil
}