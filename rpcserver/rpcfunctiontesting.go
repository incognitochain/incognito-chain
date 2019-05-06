package rpcserver

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/transaction"
	"io/ioutil"
)

type txs struct {
	Txs []string `json:"Txs"`
}

/*
For testing and benchmark only
*/
func (rpcServer RpcServer) handleGetAndSendTxsFromFile(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	file, _ := ioutil.ReadFile("./utility/txs-shard0-5000.json")
	data := txs{}
	count := 0
	_ = json.Unmarshal([]byte(file), &data)
	for index, txBase58Data := range data.Txs {
		//if index <= 200 {
		//	continue
		//}
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
		_, _, err = rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
		if err == nil {
			count++
		}
	}
	return count, nil
}

