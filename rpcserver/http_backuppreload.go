package rpcserver

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"net"
)

func (httpServer *HttpServer) handleGetLatestBackup(params interface{}, closeChan <-chan struct{}) (blockchain.BootstrapProcess, *rpcservice.RPCError) {
	return httpServer.GetBlockchain().BootstrapManager.GetLastestBootstrap(), nil
}

func (httpServer *HttpServer) handleGetBootstrapStateDB(conn net.Conn, params interface{}) {
	paramArray, ok := params.([]interface{})
	if !ok || len(paramArray) != 5 {
		return
	}

	checkpoint, ok := paramArray[0].(string)
	chainID, ok := paramArray[1].(float64)
	dbType, ok := paramArray[2].(float64)
	offset, ok := paramArray[3].(float64)

	ff := httpServer.GetBlockchain().BootstrapManager.GetBackupReader(checkpoint, int(chainID), int(dbType))
	ch, _, _ := ff.ReadRecently(uint64(offset))

	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\n\r\n"))
	if err != nil {
		return
	}
	for {
		data := <-ch
		if len(data) == 0 {
			break
		}
		_, err = conn.Write(data)
		if err != nil {
			return
		}
	}

	return
}
