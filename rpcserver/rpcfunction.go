package rpcserver

import (
	"log"
	"encoding/json"
	"bytes"
	"strings"
	"encoding/hex"
	"fmt"
	"time"
	"errors"
	"strconv"

	"github.com/ninjadotorg/cash-prototype/rpcserver/jsonrpc"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/common"
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
	"getNumberOfCoinsAndBonds":     RpcServer.handleGetNumberOfCoinsAndBonds,
	"createActionParamsTrasaction": RpcServer.handleCreateActionParamsTrasaction,

	//POS
	"votecandidate": RpcServer.handleVoteCandidate,
	"getheader":     RpcServer.handleGetHeader, // Current committee, next block committee and candidate is included in block header

	// WALLET
	"getaccount":            RpcServer.handleGetAccount,
	"listaccounts":          RpcServer.handleListAccounts,
	"getaddressesbyaccount": RpcServer.handleGetAddressesByAccount,
	"getaccountaddress":     RpcServer.handleGetAccountAddress,
	"dumpprivkey":           RpcServer.handleDumpPrivkey,
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

func (self RpcServer) handleGetHeader(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	result := jsonrpc.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	log.Println(arrayParams)
	getBy := arrayParams[0].(string)
	query := arrayParams[1].(string)
	log.Println(getBy, query)
	switch getBy {
	case "blockhash":
		bhash := common.Hash{}
		err := bhash.Decode(&bhash, query)
		log.Println(bhash)
		if err != nil {
			return nil, errors.New("Invalid blockhash format")
		}
		bnum, err := self.Config.BlockChain.GetBlockHeightByBlockHash(&bhash)
		block, err := self.Config.BlockChain.GetBlockByBlockHash(&bhash)
		if err != nil {
			return nil, errors.New("Block not exist")
		}
		result.Header = block.Header
		result.BlockNum = int(bnum) + 1
		result.BlockHash = bhash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(query)
		if err != nil {
			return nil, errors.New("Invalid blocknum format")
		}
		allHashBlocks, _ := self.Config.BlockChain.GetAllHashBlocks()
		if len(allHashBlocks) < bnum || bnum <= 0 {
			return nil, errors.New("Block not exist")
		}
		block, _ := self.Config.BlockChain.GetBlockByBlockHeight(int32(bnum - 1))
		result.Header = block.Header
		result.BlockNum = bnum
		result.BlockHash = block.Hash().String()
	default:
		return nil, errors.New("Wrong request format")
	}

	return result, nil
}

func (self RpcServer) handleVoteCandidate(params interface{}, closeChan <-chan struct{}) (interface{}, error) {

	return "", nil
}

/**
getblockchaininfo RPC return information fo blockchain node
 */
func (self RpcServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	allHashBlocks, _ := self.Config.BlockChain.GetAllHashBlocks()
	result := jsonrpc.GetBlockChainInfoResult{
		Chain:         self.Config.ChainParams.Name,
		Blocks:        len(allHashBlocks),
		BestBlockHash: self.Config.BlockChain.BestState.BestBlockHash.String(),
		Difficulty:    self.Config.BlockChain.BestState.Difficulty,
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
	blocks, _ := self.Config.BlockChain.GetAllBlocks()
	result := make([]jsonrpc.ListUnspentResult, 0)
	for _, block := range blocks {
		if (len(block.Transactions) > 0) {
			for _, tx := range block.Transactions {
				if tx.GetType() == common.TxActionParamsType {
					continue
				}
				normalTx := tx.(*transaction.Tx)
				if (len(normalTx.TxOut) > 0) {
					for index, txOut := range normalTx.TxOut {
						if (bytes.Compare(txOut.PkScript, []byte(addresses[0])) == 0) {
							result = append(result, jsonrpc.ListUnspentResult{
								Vout:      index,
								TxID:      normalTx.Hash().String(),
								Address:   string(txOut.PkScript),
								Amount:    txOut.Value,
								TxOutType: txOut.TxOutType,
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
		Type:    common.TxNormalType,
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
				Vout: uint32(temp["vout"].(float64)),
			},
		}
		tx.AddTxIn(item)
	}
	// txOut := arrayParams[1].(map[string]interface{})
	txOuts := common.InterfaceSlice(arrayParams[1])
	for _, txOut := range txOuts {
		temp := txOut.(map[string]interface{})
		tx.AddTxOut(transaction.TxOut{
			PkScript:  []byte(temp["pkScript"].(string)),
			Value:     temp["value"].(float64),
			TxOutType: temp["txOutType"].(string),
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

func isExisted(item int, arr []int) bool {
	for _, i := range arr {
		if item == i {
			return true
		}
	}
	return false
}

/**
 * handleGetNumberOfCoins handles getNumberOfCoins commands.
 */
func (self RpcServer) handleGetNumberOfCoinsAndBonds(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// return 1000, nil
	log.Println(params)
	blocks, _ := self.Config.BlockChain.GetAllBlocks()
	txInsMap := map[string][]int{}
	txOuts := []jsonrpc.ListUnspentResult{}
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			if tx.GetType() == common.TxActionParamsType {
				continue
			}
			normalTx := tx.(*transaction.Tx)
			for _, txIn := range normalTx.TxIn {
				txInKey := txIn.PreviousOutPoint.Hash.String()
				idx := txIn.PreviousOutPoint.Vout
				txInsMap[txInKey] = append(txInsMap[txInKey], int(idx))
			}

			for index, txOut := range normalTx.TxOut {
				txOuts = append(txOuts, jsonrpc.ListUnspentResult{
					Vout:      index,
					TxID:      normalTx.Hash().String(),
					Address:   string(txOut.PkScript),
					Amount:    txOut.Value,
					TxOutType: txOut.TxOutType,
				})
			}
		}
	}

	result := map[string]float64{
		common.TxOutCoinType: 0,
		common.TxOutBondType: 0,
	}
	for _, txOut := range txOuts {
		idxs, ok := txInsMap[txOut.TxID]
		if !ok { // not existing -> not used yet
			result[txOut.TxOutType] += txOut.Amount
			continue
		}
		// existing in txIns -> check Vout index
		if !isExisted(txOut.Vout, idxs) {
			result[txOut.TxOutType] += txOut.Amount
		}
	}
	return result, nil
}

func assertEligibleAgentIDs(eligibleAgentIDs interface{}) ([]string) {
	assertedEligibleAgentIDs := eligibleAgentIDs.([]interface{})
	results := []string{}
	for _, item := range assertedEligibleAgentIDs {
		results = append(results, item.(string))
	}
	return results
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
		Version:  1,
		Type:     common.TxActionParamsType,
		LockTime: time.Now().Unix(),
	}

	param := arrayParams[0].(map[string]interface{})
	tx.Param = &transaction.Param{
		AgentID:          param["agentId"].(string),
		AgentSig:         param["agentSig"].(string),
		NumOfCoins:       param["numOfCoins"].(float64),
		NumOfBonds:       param["numOfBonds"].(float64),
		Tax:              param["tax"].(float64),
		EligibleAgentIDs: assertEligibleAgentIDs(param["eligibleAgentIDs"]),
	}

	_, _, err := self.Config.TxMemPool.CanAcceptTransaction(&tx)
	if err != nil {
		return nil, err
	}

	// broadcast message
	// self.Config.Server.PushTxMessage(hash)

	return tx.Hash(), nil
}

/**
getaccount RPC returns the name of the account associated with the given address.
- Param #1: address
 */
func (self RpcServer) handleGetAccount(params interface{}, closeChan <-chan struct{}, ) (interface{}, error) {
	for _, account := range self.Config.Wallet.MasterAccount.Child {
		address := account.Key.ToAddress(false)
		if address == params.(string) {
			return account.Name, nil
		}
	}
	return "", nil
}

/**
listaccount RPC lists accounts and their balances.

Parameter #1—the minimum number of confirmations a transaction must have
Parameter #2—whether to include watch-only addresses in results
Result—a list of accounts and their balances

 */
func (self RpcServer) handleListAccounts(params interface{}, closeChan <-chan struct{}, ) (interface{}, error) {
	result := jsonrpc.ListAccounts{}
	result.Accounts = self.Config.Wallet.ListAccounts()
	return result, nil
}

/**
getaddressesbyaccount RPC returns a list of every address assigned to a particular account.

Parameter #1—the account name
Result—a list of addresses
 */
func (self RpcServer) handleGetAddressesByAccount(params interface{}, closeChan <-chan struct{}, ) (interface{}, error) {
	result := jsonrpc.GetAddressesByAccount{}
	var err error
	result.Addresses, err = self.Config.Wallet.GetAddressesByAccount(params.(string))
	return result, err
}

/**
getaccountaddress RPC returns the current Bitcoin address for receiving payments to this account. If the account doesn’t exist, it creates both the account and a new address for receiving payment. Once a payment has been received to an address, future calls to this RPC for the same account will return a different address.
Parameter #1—an account name
Result—a bitcoin address
 */
func (self RpcServer) handleGetAccountAddress(params interface{}, closeChan <-chan struct{}, ) (interface{}, error) {
	return self.Config.Wallet.GetAccountAddress(params.(string))
}

/**
 dumpprivkey RPC returns the wallet-import-format (WIP) private key corresponding to an address. (But does not remove it from the wallet.)

Parameter #1—the address corresponding to the private key to get
Result—the private key
 */
func (self RpcServer) handleDumpPrivkey(params interface{}, closeChan <-chan struct{}, ) (interface{}, error) {
	return self.Config.Wallet.DumpPrivkey(params.(string))
}
