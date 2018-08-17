package rpcserver

import (
	"log"
	"encoding/json"
	"github.com/internet-cash/prototype/rpcserver/jsonrpc"
	"bytes"
	"strings"
)

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"dosomething":       RpcServer.handleDoSomething,
	"getblockchaininfo": RpcServer.handleGetBlockChainInfo,
	"createtransaction": RpcServer.handleCreateTransaction,
	"listunspent":       RpcServer.handleListUnSpent,
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
	min := string(params.([]json.RawMessage)[0])
	max := string(params.([]json.RawMessage)[1])
	listAddresses := string(params.([]json.RawMessage)[2][1:len(params.([]json.RawMessage)[2])-1])
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
