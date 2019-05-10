package rpcserver

import (
	"encoding/json"
	"fmt"
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
type CountResult struct {
	Success int
	Fail int
}
func (rpcServer RpcServer) handleGetAndSendTxsFromFile(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	Logger.log.Critical(arrayParams)
	shardIDParam := int(arrayParams[0].(float64))
	isPrivacy := arrayParams[1].(bool)
	isSent := arrayParams[2].(bool)
	interval := arrayParams[3].(int64)
	datadir := "./utility/"
	filename := ""
	success := 0
	fail := 0
	if isPrivacy {
		filename = "txs-shard" + fmt.Sprintf("%d",shardIDParam) + "-privacy-5000.json"
	} else {
		filename = "txs-shard" + fmt.Sprintf("%d",shardIDParam) + "-noprivacy-5000.json"
	}
	Logger.log.Critical("Getting Transactions from file: ", datadir+filename)
	file, err := ioutil.ReadFile(datadir+filename)
	if err != nil {
		Logger.log.Error("Fail to get Transactions from file: ", err)
	}
	data := txs{}
	count := 0
	_ = json.Unmarshal([]byte(file), &data)
	Logger.log.Criticalf("Get %+v Transactions from file \n", len(data.Txs))
	intervalDuration := time.Duration(interval)*time.Millisecond
	for index, txBase58Data := range data.Txs {
		<-time.Tick(intervalDuration)
		Logger.log.Critical("Number of Transaction: ", index)
		//<-time.Tick(50*time.Millisecond)
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58Data)
		if err != nil {
			fail++
			continue
		}
		var tx transaction.Tx
		err = json.Unmarshal(rawTxBytes, &tx)
		if err != nil {
			fail++
			continue
		}
		if !isSent {
			_, _, err = rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
			if err != nil {
				fail++
				continue
			} else {
				success++
				continue
			}
		} else {
			_, _, err = rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
			if err != nil {
				fail++
				continue
			}
			txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
			if err != nil {
				fail++
				continue
			}
			txMsg.(*wire.MessageTx).Transaction = &tx
			err = rpcServer.config.Server.PushMessageToAll(txMsg)
			if err != nil {
				fail++
				continue
			}
		}
		if err == nil {
			count++
			success++
		}
	}
	return CountResult{Success: success, Fail:fail}, nil
}

