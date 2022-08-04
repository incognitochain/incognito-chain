package rpcserver

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"net"
)

func (httpServer *HttpServer) handleGetLatestBackup(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	latest := httpServer.GetBlockchain().BackupManager.GetLastestBootstrap()
	b, _ := json.Marshal(latest)
	fmt.Println("GetLastestBootstrap", b)
	return latest, nil
}

func (httpServer *HttpServer) handleGetBootstrapStateDB(conn net.Conn, params interface{}) {
	paramArray, ok := params.([]interface{})
	if !ok || len(paramArray) != 4 {
		return
	}

	checkpoint, ok := paramArray[0].(string)
	chainID, ok := paramArray[1].(float64)
	dbType, ok := paramArray[2].(float64)
	offset, ok := paramArray[3].(float64)

	ff := httpServer.GetBlockchain().BackupManager.GetBackupReader(checkpoint, int(chainID), int(dbType))
	ch, _, _ := ff.ReadFromIndex(uint64(offset))

	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\n\r\n"))
	if err != nil {
		return
	}
	for {
		data := <-ch
		if len(data) == 0 {
			break
		}

		var dataLen = make([]byte, 8)
		binary.LittleEndian.PutUint64(dataLen, uint64(len(data)))

		_, err = conn.Write(dataLen)
		if err != nil {
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			return
		}
	}

	return
}
