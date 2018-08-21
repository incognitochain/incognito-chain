package rpcserver

import (
	"log"
	"encoding/json"
	"github.com/internet-cash/prototype/rpcserver/jsonrpc"
	"bytes"
	"strings"
	"github.com/internet-cash/prototype/transaction"
	"github.com/internet-cash/prototype/common"
	"encoding/hex"
	"fmt"
)

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"dosomething":                  RpcServer.handleDoSomething,
	"getblockchaininfo":            RpcServer.handleGetBlockChainInfo,
	"createtransaction":            RpcServer.handleCreateTransaction,
	"listunspent":                  RpcServer.handleListUnSpent,
	"createrawtransaction":         RpcServer.handleCreateRawTrasaction,
	"signrawtransaction":           RpcServer.handleSignRawTransaction,
	"sendrawtransaction":           RpcServer.handleSendRawTransaction,
	"getNumberOfCoins":             RpcServer.handleGetNumberOfCoins,
	"getNumberOfBonds":             RpcServer.handleGetNumberOfBonds,
	"createActionParamsTrasaction": RpcServer.handleCreateActionParamsTrasaction,
}

// Commands that are available to a limited user
var RpcLimited = map[string]struct{}{

}

func (self RpcServer) handleDoSomething(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	result := make(map[string]string)
	result["param"] = string(params.([]json.RawMessage)[0])
	return result, nil
}

func (self RpcServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonrpc.GetBlockChainInfoResult{
		Chain:  self.Config.ChainParams.Name,
		Blocks: len(self.Config.Chain.Blocks),
	}
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
	paramsArray := common.InterfaceSlice(params)
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
	arrayParams := common.InterfaceSlice(params)
	tx := transaction.Tx{
		Version: 1,
		Type:    "NORMAL",
	}
	txIns := common.InterfaceSlice(arrayParams[0])
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

/**
// SignTransaction uses secrets of the wallet, as well as additional secrets
// passed in by the caller, to create and add input signatures to a transaction.
//
// Transaction input script validation is used to confirm that all signatures
// are valid.  For any invalid input, a SignatureError is added to the returns.
// The final error return is reserved for unexpected or fatal errors, such as
// being unable to determine a previous output script to redeem.
//
// The transaction pointed to by tx is modified by this function.

Parameter #1—the transaction to sign
Parameter #2—unspent transaction output details
Parameter #3—private keys for signing
Parameter #4—signature hash type
Result—the transaction with any signatures made
*/
func (self RpcServer) handleSignRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	var tx transaction.Tx
	log.Println(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(tx.TxIn); i++ {
		log.Println(WALLET_OWNER_PUBKEY_ADDRESS)
		tx.TxIn[i].SignatureScript = []byte(WALLET_OWNER_PUBKEY_ADDRESS)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	log.Println(string(byteArrays))
	return hex.EncodeToString(byteArrays), nil
}

/**
// handleSendRawTransaction implements the sendrawtransaction command.
Parameter #1—a serialized transaction to broadcast
Parameter #2–whether to allow high fees
Result—a TXID or error message
*/
func (self RpcServer) handleSendRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	var tx transaction.Tx
	log.Println(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, err
	}

	hash, txDesc, err := self.Config.TxMemPool.CanAcceptTransaction(&tx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("there is hash of transaction: %s", hash.String())
	fmt.Println()
	fmt.Printf("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	self.Config.Server.PushTxMessage(hash)

	return tx.Hash(), nil
}

/**
 * handleGetNumberOfCoins handles getNumberOfCoins commands.
 */
func (self RpcServer) handleGetNumberOfCoins(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return 1000, nil
}

/**
 * handleGetNumberOfBonds handles getNumberOfBonds commands.
 */
func (self RpcServer) handleGetNumberOfBonds(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return 5, nil
}

/**
// handleCreateRawTransaction handles createrawtransaction commands.
 */
func (self RpcServer) handleCreateActionParamsTrasaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, error) {
	log.Println(params)
	arrayParams := common.InterfaceSlice(params)
	tx := transaction.ActionParamTx{
		Version: 1,
		Type:    "ACTION_PARAMS",
	}

	param := arrayParams[0].(map[string]interface{})
	tx.Param = &transaction.Param{
		AgentID:           param["agentId"].(string),
		NumOfIssuingCoins: int(param["numOfIssuingCoins"].(float64)),
		NumOfIssuingBonds: int(param["numOfIssuingBonds"].(float64)),
		Tax:               param["tax"].(float64),
	}

	_, _, err := self.Config.TxMemPool.CanAcceptTransaction(&tx)
	if err != nil {
		return nil, err
	}

	// broadcast message
	// self.Config.Server.PushTxMessage(hash)

	return tx.Hash(), nil
}
