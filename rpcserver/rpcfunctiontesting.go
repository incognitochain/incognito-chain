package rpcserver

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wire"
	"io/ioutil"
	"time"
)

type txs struct {
	Txs []string `json:"Txs"`
}

/*
For testing and benchmark only
*/
func (rpcServer RpcServer) handleGetAndSendTxsFromFile(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Critical("Getting Transactions from file")
	file, err := ioutil.ReadFile("./utility/txs-shard0-5000.json")
	arrayParams := common.InterfaceSlice(params)
	isSent := arrayParams[0].(bool)
	if err != nil {
		Logger.log.Error("Fail to get Transactions from file: ", err)
	}
	data := txs{}
	count := 0
	_ = json.Unmarshal([]byte(file), &data)
	Logger.log.Criticalf("Get %+v Transactions from file \n", len(data.Txs))
	for index, txBase58Data := range data.Txs {
		<-time.Tick(50*time.Millisecond)
		if index % 500 == 0 {
			<-time.Tick(2 * time.Second)
		}
		Logger.log.Critical("Number of Transaction: ", index)
		//<-time.Tick(50*time.Millisecond)
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58Data)
		if err != nil {
			return nil, NewRPCError(ErrSendTxData, err)
		}
		var tx transaction.Tx
		err = json.Unmarshal(rawTxBytes, &tx)
		if err != nil {
			return nil, NewRPCError(ErrSendTxData, err)
		}
		if !isSent {
			_, _, err = rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
		} else {
			_, _, err = rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
			txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
			if err != nil {
				return nil, NewRPCError(ErrSendTxData, err)
			}
			
			txMsg.(*wire.MessageTx).Transaction = &tx
			err = rpcServer.config.Server.PushMessageToAll(txMsg)
		}
		if err == nil {
			count++
			// broadcast Message
		}
	}
	return count, nil
}

