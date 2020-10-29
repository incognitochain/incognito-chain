package devframework

import (
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)
func (sim *SimulationEngine) rpc_createtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (*jsonresult.CreateTransactionResult,error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createtransaction"]
	resI, err := c(httpServer, []interface{}{privateKey,receivers,fee,privacy}, nil)
	if err != nil {
		return nil,err
	}
	return resI.(*jsonresult.CreateTransactionResult),err
}
func (sim *SimulationEngine) rpc_getrewardamount(paymentAddress string) (error) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getrewardamount"]
	resI, err := c(httpServer, []interface{}{paymentAddress}, nil)
	if err != nil {
		return err
	}
	_ = resI 
 return err
}