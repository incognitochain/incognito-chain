package rpcserver

import (
	"log"
	"encoding/json"
	"github.com/internet-cash/prototype/rpcserver/jsonrpc"
	"bytes"
	"strings"
	"reflect"
	"github.com/internet-cash/prototype/transaction"
	"github.com/internet-cash/prototype/common"
	"encoding/hex"
)

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"dosomething":          RpcServer.handleDoSomething,
	"getblockchaininfo":    RpcServer.handleGetBlockChainInfo,
	"createtransaction":    RpcServer.handleCreateTransaction,
	"listunspent":          RpcServer.handleListUnSpent,
	"createrawtransaction": RpcServer.handleCreateRawTrasaction,
}

// Commands that are available to a limited user
var RpcLimited = map[string]struct{}{

}

func (self RpcServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonrpc.GetBlockChainInfoResult{
		Chain:  self.Config.ChainParams.Name,
		Blocks: len(self.Config.Chain.Blocks),
	}
	return result, nil
}

func (self RpcServer) handleDoSomething(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	result := make(map[string]string)
	result["param"] = string(params.([]json.RawMessage)[0])
	return result, nil
}

func (self RpcServer) handleCreateTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	return nil, nil
}

/**
// ListUnspent returns a slice of objects representing the unspent wallet
// transactions fitting the given criteria. The confirmations will be more than
// minconf, less than maxconf and if addresses is populated only the addresses
// contained within it will be considered.  If we know nothing about a
// transaction an empty array will be returned.
// params:
 Parameter #1—the minimum number of confirmations an output must have
Parameter #2—the maximum number of confirmations an output may have
Parameter #3—the addresses an output must pay
 */
func (self RpcServer) handleListUnSpent(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	paramsArray := InterfaceSlice(params)
	min := int(paramsArray[0].(float64))
	max := int(paramsArray[1].(float64))
	listAddresses := paramsArray[2].(string)
	_ = min
	_ = max
	var addresses []string
	addresses = strings.Fields(listAddresses)
	blocks := self.Config.Chain.Blocks
	result := make([]jsonrpc.ListUnspentResult, 0)
	for _, block := range blocks {
		if (len(block.Transactions) > 0) {
			for _, tx := range block.Transactions {
				if (len(tx.TxOut) > 0) {
					for index, txOut := range tx.TxOut {
						if (bytes.Compare(txOut.PkScript, []byte(addresses[0])) == 0) {
							result = append(result, jsonrpc.ListUnspentResult{
								Vout:    index,
								TxID:    tx.Hash().String(),
								Address: string(txOut.PkScript),
								Amount:  txOut.Value,
							})
						}
					}
				}
			}
		}
	}
	return result, nil
}

/**
// handleCreateRawTransaction handles createrawtransaction commands.
 */
func (self RpcServer) handleCreateRawTrasaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	arrayParams := InterfaceSlice(params)
	tx := transaction.Tx{
		Version: 1,
	}
	txIns := InterfaceSlice(arrayParams[0])
	for _, txIn := range txIns {
		temp := txIn.(map[string]interface{})
		txId := temp["txid"].(string)
		hashTxId, err := common.Hash{}.NewHashFromStr(txId)
		if err != nil {
			return nil, err
		}
		item := transaction.TxIn{
			PreviousOutPoint: transaction.OutPoint{
				Hash: *hashTxId,
				Vout: int(temp["vout"].(float64)),
			},
		}
		tx.AddTxIn(item)
	}
	txOut := arrayParams[1].(map[string]interface{})
	for key, val := range txOut {
		tx.AddTxOut(transaction.TxOut{
			PkScript: []byte(key),
			Value:    val.(float64),
		})
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return hex.EncodeToString(byteArrays), nil
}

func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}
