package rpcserver

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ninjadotorg/cash-prototype/wire"
	"log"
	"strconv"
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/rpcserver/jsonrpc"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"golang.org/x/crypto/ed25519"
	"github.com/ninjadotorg/cash-prototype/wallet"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/cashec"
)

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"getblockchaininfo":             RpcServer.handleGetBlockChainInfo,
	"getblockcount":                 RpcServer.handleGetBlockCount,
	"getblockhash":                  RpcServer.handleGetBlockHash,
	"getblocktemplate":              RpcServer.handleGetBlockTemplate,
	"listtransactions":              RpcServer.handleListTransactions,
	"createtransaction":             RpcServer.handleCreateTrasaction,
	"sendtransaction":               RpcServer.handleSendTransaction,
	"getnumberofcoinsandbonds":      RpcServer.handleGetNumberOfCoinsAndBonds,
	"createactionparamstransaction": RpcServer.handleCreateActionParamsTransaction,

	//POS
	"votecandidate": RpcServer.handleVoteCandidate,
	"getheader":     RpcServer.handleGetHeader, // Current committee, next block committee and candidate is included in block header

	//
	"getallpeers": RpcServer.handleGetAllPeers,
}

// Commands that are available to a limited user
var RpcLimited = map[string]commandHandler{
	// WALLET
	"listaccounts":          RpcServer.handleListAccounts,
	"getaccount":            RpcServer.handleGetAccount,
	"getaddressesbyaccount": RpcServer.handleGetAddressesByAccount,
	"getaccountaddress":     RpcServer.handleGetAccountAddress,
	"dumpprivkey":           RpcServer.handleDumpPrivkey,
	"importaccount":         RpcServer.handleImportAccount,
	"listunspent":           RpcServer.handleListUnspent,
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

/**
getblockcount RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if self.Config.BlockChain.BestState != nil && self.Config.BlockChain.BestState.BestBlock != nil {
		return self.Config.BlockChain.BestState.BestBlock.Height + 1, nil
	}
	return nil, errors.New("Wrong data")
}

/**
getblockhash RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	heights, ok := params.([]interface{})
	if ok && len(heights) >= 1 {
		height := int32(heights[0].(float64))
		hash, err := self.Config.BlockChain.GetBlockByBlockHeight(height)
		if err != nil {
			return nil, err
		}
		return hash.Hash().String(), nil
	}
	return nil, errors.New("Wrong request format")
}

/**
getblocktemplate RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockTemplate(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if self.Config.BlockChain.BestState != nil && self.Config.BlockChain.BestState.BestBlock != nil {
		block := self.Config.BlockChain.BestState.BestBlock
		result := map[string]interface{}{}
		result["capabilities"] = []string{"proposal"}
		result["version"] = block.Header.Version
		result["rules"] = []string{"csv", "segwit"}
		result["vbavailable"] = []string{}
		result["vbrequired"] = 0
		result["previousblockhash"] = block.Header.PrevBlockHash.String()

		transactions := []map[string]interface{}{}
		for _, tx := range block.Transactions {
			transactionT := map[string]interface{}{}

			transactionT["data"] = nil
			transactionT["txid"] = tx.Hash().String()
			transactionT["hash"] = tx.Hash().String()
			transactionT["depends"] = []string{}

			if tx.GetType() == common.TxNormalType {
				txN := tx.(*transaction.Tx)
				transactionT["fee"] = txN.Fee
				data, err := json.Marshal(txN)
				if err != nil {
					return nil, err
				}
				transactionT["data"] = hex.EncodeToString(data)

			} else if tx.GetType() == common.TxActionParamsType {
				txA := tx.(*transaction.ActionParamTx)
				transactionT["fee"] = 0
				data, err := json.Marshal(txA)
				if err != nil {
					return nil, err
				}
				transactionT["data"] = hex.EncodeToString(data)
			} else {
				transactionT["fee"] = 0
			}

			transactionT["sigops"] = 0
			transactionT["weight"] = 0

			transactions = append(transactions, transactionT)
		}
		result["transactions"] = transactions

		return result, nil
	}
	return nil, errors.New("Wrong data")
}

/**
// handleList returns a slice of objects representing the wallet
// transactions fitting the given criteria. The confirmations will be more than
// minconf, less than maxconf and if addresses is populated only the addresses
// contained within it will be considered.  If we know nothing about a
// transaction an empty array will be returned.
// params:
Parameter #1—the minimum number of confirmations an output must have
Parameter #2—the maximum number of confirmations an output may have
Parameter #3—the list readonly which be used to view utxo
*/
func (self RpcServer) handleListTransactions(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	result := jsonrpc.ListUnspentResult{
		ListUnspentResultItems: make(map[string][]jsonrpc.ListUnspentResultItem),
	}

	// get params
	paramsArray := common.InterfaceSlice(params)
	min := int(paramsArray[0].(float64))
	max := int(paramsArray[1].(float64))
	_ = min
	_ = max
	listKeyParams := common.InterfaceSlice(paramsArray[2])
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain readonly-key by deserializing
		readonlyKeyStr := keys["ReadonlyKey"].(string)
		readonlyKey, err := wallet.Base58CheckDeserialize(readonlyKeyStr)
		if err != nil {
			return nil, err
		}

		// get keyset only contain pub-key by deserializing
		pubKeyStr := keys["PublicKey"].(string)
		pubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			return nil, err
		}

		// create a key set
		keySet := cashec.KeySet{
			ReadonlyKey: readonlyKey.KeyPair.ReadonlyKey,
			PublicKey:   pubKey.KeyPair.PublicKey,
		}

		txs, err := self.Config.BlockChain.GetListTxByReadonlyKey(&keySet, common.TxOutCoinType)
		if err != nil {
			return nil, err
		}
		listTxs := make([]jsonrpc.ListUnspentResultItem, 0)
		for _, tx := range txs {
			item := jsonrpc.ListUnspentResultItem{
				TxId:          tx.Hash().String(),
				JoinSplitDesc: make([]jsonrpc.JoinSplitDesc, 0),
			}
			for _, desc := range tx.Descs {
				notes := desc.GetNote()
				amounts := make([]uint64, 0)
				for _, note := range notes {
					amounts = append(amounts, note.Value)
				}
				item.JoinSplitDesc = append(item.JoinSplitDesc, jsonrpc.JoinSplitDesc{
					Anchor:      desc.Anchor,
					Commitments: desc.Commitments,
					Amounts:     amounts,
				})
			}
			listTxs = append(listTxs, item)
		}
		result.ListUnspentResultItems[readonlyKeyStr] = listTxs
	}
	return result, nil
}

/**
// handleList returns a slice of objects representing the unspent wallet
// transactions fitting the given criteria. The confirmations will be more than
// minconf, less than maxconf and if addresses is populated only the addresses
// contained within it will be considered.  If we know nothing about a
// transaction an empty array will be returned.
// params:
Parameter #1—the minimum number of confirmations an output must have
Parameter #2—the maximum number of confirmations an output may have
Parameter #3—the list readonly which be used to view utxo
*/
func (self RpcServer) handleListUnspent(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	result := jsonrpc.ListUnspentResult{
		ListUnspentResultItems: make(map[string][]jsonrpc.ListUnspentResultItem),
	}

	// get params
	paramsArray := common.InterfaceSlice(params)
	min := int(paramsArray[0].(float64))
	max := int(paramsArray[1].(float64))
	_ = min
	_ = max
	listKeyParams := common.InterfaceSlice(paramsArray[2])
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain pri-key by deserializing
		priKeyStr := keys["PrivateKey"].(string)
		readonlyKey, err := wallet.Base58CheckDeserialize(priKeyStr)
		if err != nil {
			return nil, err
		}

		txs, err := self.Config.BlockChain.GetListTxByPrivateKey(&readonlyKey.KeyPair.PrivateKey, common.TxOutCoinType, transaction.NoSort, false)
		if err != nil {
			return nil, err
		}
		listTxs := make([]jsonrpc.ListUnspentResultItem, 0)
		for _, tx := range txs {
			item := jsonrpc.ListUnspentResultItem{
				TxId:          tx.Hash().String(),
				JoinSplitDesc: make([]jsonrpc.JoinSplitDesc, 0),
			}
			for _, desc := range tx.Descs {
				notes := desc.GetNote()
				amounts := make([]uint64, 0)
				for _, note := range notes {
					amounts = append(amounts, note.Value)
				}
				item.JoinSplitDesc = append(item.JoinSplitDesc, jsonrpc.JoinSplitDesc{
					Anchor:      desc.Anchor,
					Commitments: desc.Commitments,
					Amounts:     amounts,
				})
			}
			listTxs = append(listTxs, item)
		}
		result.ListUnspentResultItems[priKeyStr] = listTxs
	}
	return result, nil
}

/**
// handleCreateTransaction handles createtransaction commands.
*/
func (self RpcServer) handleCreateTrasaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)

	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, nil
	}

	// param #2: list receiver
	totalAmmount := uint64(0)
	receiversParam := arrayParams[1].(map[string]interface{})
	paymentInfos := make([]*client.PaymentInfo, 0)
	for pubKeyStr, amount := range receiversParam {
		receiverPubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			return nil, nil
		}
		paymentInfo := &client.PaymentInfo{
			Amount:         uint64(amount.(float64)),
			PaymentAddress: receiverPubKey.KeyPair.PublicKey,
		}
		totalAmmount += paymentInfo.Amount
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// list unspent tx
	usableTxs, _ := self.Config.BlockChain.GetListTxByPrivateKey(&senderKey.KeyPair.PrivateKey, common.TxOutCoinType, transaction.SortByAmount, false)
	candidateTxs := make([]*transaction.Tx, 0)
	for _, temp := range usableTxs {
		for _, desc := range temp.Descs {
			for _, note := range desc.GetNote() {
				amount := note.Value
				totalAmmount -= amount
			}
		}
		txData := temp
		candidateTxs = append(candidateTxs, &txData)
		if totalAmmount <= 0 {
			break
		}
	}

	// get tx view point
	txViewPoint, err := self.Config.BlockChain.FetchTxViewPoint(common.TxOutCoinType)
	for _, c := range txViewPoint.ListCommitments(common.TxOutCoinType) {
		println(hex.EncodeToString(c))
	}
	// create a new tx
	tx, err := transaction.CreateTx(&senderKey.KeyPair.PrivateKey, paymentInfos, &self.Config.BlockChain.BestState.BestBlock.Header.MerkleRootCommitments, candidateTxs, txViewPoint.ListNullifiers(common.TxOutCoinType), txViewPoint.ListCommitments(common.TxOutCoinType))
	if err != nil {
		return nil, err
	}
	byteArrays, err := json.Marshal(tx)
	if err == nil {
		// return hex for a new tx
		return hex.EncodeToString(byteArrays), nil
	}
	return nil, err
}

/**
// handleSendTransaction implements the sendtransaction command.
Parameter #1—a serialized transaction to broadcast
Parameter #2–whether to allow high fees
Result—a TXID or error message
*/
func (self RpcServer) handleSendTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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

	hash, txDesc, err := self.Config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, err
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.Config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

/**
 * handleGetNumberOfCoins handles getNumberOfCoins commands.
 */
func (self RpcServer) handleGetNumberOfCoinsAndBonds(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	result, err := self.Config.BlockChain.GetAllUnitCoinSupplier()
	return result, err
}

func assertEligibleAgentIDs(eligibleAgentIDs interface{}) []string {
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
func (self RpcServer) handleCreateActionParamsTransaction(
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

	// check signed tx
	message := map[string]interface{}{
		"agentId":          tx.Param.AgentID,
		"numOfCoins":       tx.Param.NumOfCoins,
		"numOfBonds":       tx.Param.NumOfBonds,
		"tax":              tx.Param.Tax,
		"eligibleAgentIDs": tx.Param.EligibleAgentIDs,
	}
	pubKeyInBytes, _ := base64.StdEncoding.DecodeString(tx.Param.AgentID)
	sigInBytes, _ := base64.StdEncoding.DecodeString(tx.Param.AgentSig)
	messageInBytes, _ := json.Marshal(message)

	isValid := ed25519.Verify(pubKeyInBytes, messageInBytes, sigInBytes)
	fmt.Println("isValid: ", isValid)

	_, _, err := self.Config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, err
	}

	// broadcast message
	// self.Config.Server.PushTxMessage(hash)

	return tx.Hash(), nil
}

/**
listaccount RPC lists accounts and their balances.

Parameter #1—the minimum number of confirmations a transaction must have
Parameter #2—whether to include watch-only addresses in results
Result—a list of accounts and their balances

*/
func (self RpcServer) handleListAccounts(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonrpc.ListAccounts{
		Accounts: make(map[string]uint64),
	}
	accounts := self.Config.Wallet.ListAccounts()
	for accountName, account := range accounts {
		txs, err := self.Config.BlockChain.GetListTxByPrivateKey(&account.Key.KeyPair.PrivateKey, common.TxOutCoinType, transaction.NoSort, false)
		if err != nil {
			return nil, err
		}
		amount := uint64(0)
		for _, tx := range txs {
			for _, desc := range tx.Descs {
				notes := desc.GetNote()
				for _, note := range notes {
					amount += note.Value
				}
			}
		}
		result.Accounts[accountName] = amount
	}

	return result, nil
}

/**
getaccount RPC returns the name of the account associated with the given address.
- Param #1: address
*/
func (self RpcServer) handleGetAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	for _, account := range self.Config.Wallet.MasterAccount.Child {
		address := account.Key.Base58CheckSerialize(wallet.PubKeyType)
		if address == params.(string) {
			return account.Name, nil
		}
	}
	return "", nil
}

/**
getaddressesbyaccount RPC returns a list of every address assigned to a particular account.

Parameter #1—the account name
Result—a list of addresses
*/
func (self RpcServer) handleGetAddressesByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonrpc.GetAddressesByAccount{}
	var err error
	result.Addresses, err = self.Config.Wallet.GetAddressesByAccount(params.(string))
	return result, err
}

/**
getaccountaddress RPC returns the current coin address for receiving payments to this account. If the account doesn’t exist, it creates both the account and a new address for receiving payment. Once a payment has been received to an address, future calls to this RPC for the same account will return a different address.
Parameter #1—an account name
Result—a bitcoin address
*/
func (self RpcServer) handleGetAccountAddress(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.Config.Wallet.GetAccountAddress(params.(string))
}

/**
 dumpprivkey RPC returns the wallet-import-format (WIP) private key corresponding to an address. (But does not remove it from the wallet.)

Parameter #1—the address corresponding to the private key to get
Result—the private key
*/
func (self RpcServer) handleDumpPrivkey(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.Config.Wallet.DumpPrivkey(params.(string))
}

/**
handleImportAccount - import a new account by private-key
- Param #1: private-key string
- Param #2: account name
- Param #3: passPhrase of wallet
 */
func (self RpcServer) handleImportAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	privateKey := arrayParams[0].(string)
	accountName := arrayParams[1].(string)
	passPhrase := arrayParams[2].(string)
	account, err := self.Config.Wallet.ImportAccount(privateKey, accountName, passPhrase)
	if err != nil {
		return "", err
	}
	return wallet.KeySerializedData{
		PublicKey:   account.Key.Base58CheckSerialize(wallet.PubKeyType),
		ReadonlyKey: account.Key.Base58CheckSerialize(wallet.ReadonlyKeyType),
	}, err
}

/**
handleGetAllPeers - return all peers which this node connected
 */
func (self RpcServer) handleGetAllPeers(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	result := make(map[string]interface{})

	peersMap := []string{}

	peers := self.Config.AddrMgr.AddressCache()
	for _, peer := range peers {
		peersMap = append(peersMap, peer.RawAddress)
	}

	result["peers"] = peersMap

	return result, nil
}
