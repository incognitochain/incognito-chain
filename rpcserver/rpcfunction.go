package rpcserver

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ninjadotorg/cash-prototype/wire"

	"net"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/rpcserver/jsonrpc"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/wallet"
	"golang.org/x/crypto/ed25519"
)

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"getnetworkinfo":                RpcServer.handleGetNetWorkInfo,
	"getbestblock":                  RpcServer.handleGetBestBlock,
	"getbestblockhash":              RpcServer.handleGetBestBlockHash,
	"getblock":                      RpcServer.handleGetBlock,
	"getblockchaininfo":             RpcServer.handleGetBlockChainInfo,
	"getblockcount":                 RpcServer.handleGetBlockCount,
	"getblockhash":                  RpcServer.handleGetBlockHash,
	"getblocktemplate":              RpcServer.handleGetBlockTemplate,
	"listtransactions":              RpcServer.handleListTransactions,
	"createtransaction":             RpcServer.handleCreateTransaction,
	"sendtransaction":               RpcServer.handleSendTransaction,
	"sendmany":                      RpcServer.handleSendMany,
	"getnumberofcoinsandbonds":      RpcServer.handleGetNumberOfCoinsAndBonds,
	"createactionparamstransaction": RpcServer.handleCreateActionParamsTransaction,
	"getconnectioncount":            RpcServer.handleGetConnectionCount,
	"getgenerate":                   RpcServer.handleGetGenerate,
	"getmempoolinfo":                RpcServer.handleGetMempoolInfo,
	"getmininginfo":                 RpcServer.handleGetMiningInfo,
	"getrawmempool":                 RpcServer.handleGetRawMempool,
	"getmempoolentry":               RpcServer.handleMempoolEntry,
	"estimatefee":                   RpcServer.handleEstimateFee,

	//POS
	"votecandidate": RpcServer.handleVoteCandidate,
	"getheader":     RpcServer.handleGetHeader, // Current committee, next block committee and candidate is included in block header

	//
	//"getallpeers": RpcServer.handleGetAllPeers,
}

// Commands that are available to a limited user
var RpcLimited = map[string]commandHandler{
	"addnode":          RpcServer.handleAddNode,
	"getaddednodeinfo": RpcServer.handleGetAddedNodeInfo,
	// WALLET
	"listaccounts":          RpcServer.handleListAccounts,
	"getaccount":            RpcServer.handleGetAccount,
	"getaddressesbyaccount": RpcServer.handleGetAddressesByAccount,
	"getaccountaddress":     RpcServer.handleGetAccountAddress,
	"dumpprivkey":           RpcServer.handleDumpPrivkey,
	"importaccount":         RpcServer.handleImportAccount,
	"listunspent":           RpcServer.handleListUnspent,
	"getbalance":            RpcServer.handleGetBalance,
	"getreceivedbyaccount":  RpcServer.handleGetReceivedByAccount,
}

func (self RpcServer) handleGetHeader(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	result := jsonrpc.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	log.Println(arrayParams)
	getBy := arrayParams[0].(string)
	block := arrayParams[1].(string)
	chainID := arrayParams[2].(float64)
	switch getBy {
	case "blockhash":
		bhash := common.Hash{}
		err := bhash.Decode(&bhash, block)
		log.Println(bhash)
		if err != nil {
			return nil, errors.New("Invalid blockhash format")
		}
		block, err := self.Config.BlockChain.GetBlockByBlockHash(&bhash)
		if err != nil {
			return nil, errors.New("Block not exist")
		}
		result.Header = block.Header
		result.BlockNum = int(block.Height) + 1
		result.ChainID = uint8(chainID)
		result.BlockHash = bhash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(block)
		if err != nil {
			return nil, errors.New("Invalid blocknum format")
		}
		fmt.Println(chainID)
		if int32(bnum-1) > self.Config.BlockChain.BestState[uint8(chainID)].Height || bnum <= 0 {
			return nil, errors.New("Block not exist")
		}
		block, _ := self.Config.BlockChain.GetBlockByBlockHeight(int32(bnum-1), uint8(chainID))
		result.Header = block.Header
		result.BlockNum = bnum
		result.ChainID = uint8(chainID)
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
getblockcount RPC return information fo blockchain node
*/
func (self RpcServer) handleGetNetWorkInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := map[string]interface{}{}

	result["version"] = 1
	result["subversion"] = ""
	result["protocolversion"] = 0
	result["localservices"] = ""
	result["localrelay"] = true
	result["timeoffset"] = 0
	result["networkactive"] = true
	result["connections"] = 0

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	networks := []map[string]interface{}{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			network := map[string]interface{}{}

			network["name"] = "ipv4"
			network["limited"] = false
			network["reachable"] = true
			network["proxy"] = ""
			network["proxy_randomize_credentials"] = false

			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To16() != nil {
					network["name"] = "ipv6"
				}
			}

			networks = append(networks, network)
		}
	}

	result["networks"] = networks

	result["localaddresses"] = []string{}

	result["relayfee"] = 0
	result["incrementalfee"] = 0
	result["warnings"] = ""

	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (self RpcServer) handleGetBestBlock(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// All other "get block" commands give either the height, the
	// hash, or both but require the block SHA.  This gets both for
	// the best block.
	best := self.Config.BlockChain.BestState
	result := map[string]interface{}{
		"hash":   best.BestBlockHash.String(),
		"height": best.Height,
	}
	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (self RpcServer) handleGetBestBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// All other "get block" commands give either the height, the
	// hash, or both but require the block SHA.  This gets both for
	// the best block.
	best := self.Config.BlockChain.BestState
	return best.BestBlockHash.String(), nil
}

/**
getblockcount RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlock(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 1 {
		hashString := paramsT[0].(string)
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			return nil, errH
		}
		block, errD := self.Config.BlockChain.GetBlockByBlockHash(hash)
		if errD != nil {
			return nil, errD
		}
		result := map[string]interface{}{}

		verbosity := "1"
		if len(paramsT) >= 2 {
			verbosity = paramsT[1].(string)
		}

		if verbosity == "0" {
			data, err := json.Marshal(block)
			if err != nil {
				return nil, err
			}
			result["data"] = hex.EncodeToString(data)
		} else if verbosity == "1" {
			best := self.Config.BlockChain.BestState

			blockHeight := block.Height
			// Get next block hash unless there are none.
			var nextHashString string
			if blockHeight < best.Height {
				nextHash, err := self.Config.BlockChain.GetBlockByBlockHeight(blockHeight + 1)
				if err != nil {
					return nil, err
				}
				nextHashString = nextHash.Hash().String()
			}

			result["hash"] = block.Hash().String()
			result["confirmations"] = int64(1 + best.Height - blockHeight)
			result["size"] = -1
			result["strippedsize"] = -1
			result["weight"] = -1
			result["height"] = block.Height
			result["version"] = block.Header.Version
			result["versionHex"] = fmt.Sprintf("%x", block.Header.Version)
			result["merkleroot"] = block.Header.MerkleRoot.String()
			result["time"] = block.Header.Timestamp
			result["mediantime"] = 0
			result["nonce"] = block.Header.Nonce
			result["bits"] = ""
			result["difficulty"] = block.Header.Difficulty
			result["chainwork"] = block.Header.ChainID
			result["previousblockhash"] = block.Header.PrevBlockHash.String()
			result["nextblockhash"] = nextHashString
			result["tx"] = []string{}
			for _, tx := range block.Transactions {
				result["tx"] = append(result["tx"].([]string), tx.Hash().String())
			}
		} else if verbosity == "2" {
			best := self.Config.BlockChain.BestState

			blockHeight := block.Height
			// Get next block hash unless there are none.
			var nextHashString string
			if blockHeight < best.Height {
				nextHash, err := self.Config.BlockChain.GetBlockByBlockHeight(blockHeight + 1)
				if err != nil {
					return nil, err
				}
				nextHashString = nextHash.Hash().String()
			}

			result["hash"] = block.Hash().String()
			result["confirmations"] = int64(1 + best.Height - blockHeight)
			result["size"] = -1
			result["strippedsize"] = -1
			result["weight"] = -1
			result["height"] = block.Height
			result["version"] = block.Header.Version
			result["versionHex"] = fmt.Sprintf("%x", block.Header.Version)
			result["merkleroot"] = block.Header.MerkleRoot.String()
			result["time"] = block.Header.Timestamp
			result["mediantime"] = 0
			result["nonce"] = block.Header.Nonce
			result["bits"] = ""
			result["difficulty"] = block.Header.Difficulty
			result["chainwork"] = block.Header.ChainID
			result["previousblockhash"] = block.Header.PrevBlockHash.String()
			result["nextblockhash"] = nextHashString
			result["tx"] = []map[string]interface{}{}
			for _, tx := range block.Transactions {
				transactionT := map[string]interface{}{}

				transactionT["version"] = block.Header.Version
				transactionT["size"] = -1
				transactionT["vsize"] = -1
				transactionT["hex"] = nil
				transactionT["txid"] = tx.Hash().String()
				transactionT["hash"] = tx.Hash().String()

				if tx.GetType() == common.TxNormalType {
					txN := tx.(*transaction.Tx)
					data, err := json.Marshal(txN)
					if err != nil {
						return nil, err
					}
					transactionT["hex"] = hex.EncodeToString(data)
					transactionT["locktime"] = txN.LockTime
				} else if tx.GetType() == common.TxActionParamsType {
					txA := tx.(*transaction.ActionParamTx)
					data, err := json.Marshal(txA)
					if err != nil {
						return nil, err
					}
					transactionT["hex"] = hex.EncodeToString(data)
					transactionT["locktime"] = txA.LockTime
				}

				transactionT["blockhash"] = block.Hash().String()
				transactionT["confirmations"] = 0
				transactionT["time"] = block.Header.Timestamp
				transactionT["blocktime"] = block.Header.Timestamp

				result["tx"] = append(result["tx"].([]map[string]interface{}), transactionT)
			}
		}

		return result, nil
	}
	return nil, errors.New("Wrong request format")
}

/**
getblockchaininfo RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	allHashBlocks, _ := self.Config.BlockChain.GetAllHashBlocks()
	result := jsonrpc.GetBlockChainInfoResult{
		Chain:  self.Config.ChainParams.Name,
		Blocks: len(allHashBlocks),
	}

	for _, bestState := range self.Config.BlockChain.BestState {
		result.BestBlockHash = append(result.BestBlockHash, bestState.BestBlockHash.String())
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
getaddednodeinfo RPC return information fo blockchain node
*/
func (self RpcServer) handleGetAddedNodeInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// get params
	paramsArray := common.InterfaceSlice(params)
	nodes := []map[string]interface{}{}
	for _, nodeAddrI := range paramsArray {
		if nodeAddr, ok := nodeAddrI.(string); ok {
			for _, listen := range self.Config.ConnMgr.ListeningPeers {
				peerIDstr := self.Config.ConnMgr.GetPeerIDStr(nodeAddr)

				peerConn, existed := listen.PeerConns[peerIDstr]
				if existed {
					node := map[string]interface{}{}

					node["addednode"] = peerConn.Peer.RawAddress
					node["connected"] = true
					connected := "inbound"
					if peerConn.IsOutbound {
						connected = "outbound"
					}
					node["addresses"] = []map[string]interface{}{
						map[string]interface{}{
							"address":   peerConn.Peer.RawAddress,
							"connected": connected,
						},
					}

					nodes = append(nodes, node)
				} else {
					peer, existed := listen.PendingPeers[peerIDstr]
					if existed {
						node := map[string]interface{}{}

						node["addednode"] = peer.RawAddress
						node["connected"] = false
						node["addresses"] = []map[string]interface{}{
							map[string]interface{}{
								"address":   peer.RawAddress,
								"connected": "outbound",
							},
						}

						nodes = append(nodes, node)
					}
				}
			}
		}
	}

	return nodes, nil
}

/**
addnode RPC return information fo blockchain node
*/
func (self RpcServer) handleAddNode(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	paramsArray := common.InterfaceSlice(params)
	for _, nodeAddrI := range paramsArray {
		if nodeAddr, ok := nodeAddrI.(string); ok {
			for _, listen := range self.Config.ConnMgr.ListeningPeers {
				peerIDstr := self.Config.ConnMgr.GetPeerIDStr(nodeAddr)
				_, existed := listen.PeerConns[peerIDstr]
				if existed {
				} else {
					_, existed := listen.PendingPeers[peerIDstr]
					if existed {
					} else {
						// TODO
					}
				}
			}
		}
	}

	return nil, nil
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
func (self RpcServer) handleCreateTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
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
	totalAmmount := int64(0)
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
		totalAmmount += int64(paymentInfo.Amount)
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// list unspent tx
	usableTxs, _ := self.Config.BlockChain.GetListTxByPrivateKey(&senderKey.KeyPair.PrivateKey, common.TxOutCoinType, transaction.SortByAmount, false)
	candidateTxs := make([]*transaction.Tx, 0)
	for _, temp := range usableTxs {
		for _, desc := range temp.Descs {
			for _, note := range desc.GetNote() {
				amount := note.Value
				totalAmmount -= int64(amount)
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
	// for _, c := range txViewPoint.ListCommitments(common.TxOutCoinType) {
	// 	println(hex.EncodeToString(c))
	// }
	// create a new tx
	fmt.Printf("[handleCreateTransaction] MerkleRootCommitments: %x\n", self.Config.BlockChain.BestState.BestBlock.Header.MerkleRootCommitments[:])
	var fee uint64 // TODO(@sirrush): provide correct value
	tx, err := transaction.CreateTx(&senderKey.KeyPair.PrivateKey, paymentInfos, &self.Config.BlockChain.BestState.BestBlock.Header.MerkleRootCommitments, candidateTxs, txViewPoint.ListNullifiers(common.TxOutCoinType), txViewPoint.ListCommitments(common.TxOutCoinType), fee)
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
	// log.Println(string(rawTxBytes))
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
handleSendMany - RPC creates transaction and send to network
*/
func (self RpcServer) handleSendMany(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	hexStrOfTx, err := self.handleCreateTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendTransaction(newParam, closeChan)
	return txId, err
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

///**
//handleGetAllPeers - return all peers which this node connected
// */
//func (self RpcServer) handleGetAllPeers(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
//	log.Println(params)
//	result := make(map[string]interface{})
//
//	peersMap := []string{}
//
//	peers := self.Config.AddrMgr.AddressCache()
//	for _, peer := range peers {
//		peersMap = append(peersMap, peer.RawAddress)
//	}
//
//	result["peers"] = peersMap
//
//	return result, nil
//}

/**
handleGetBalance - RPC gets the balances in decimal
*/
func (self RpcServer) handleGetBalance(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	balance := uint64(0)

	if self.Config.Wallet == nil {
		return balance, errors.New("Wallet is not existed")
	}
	if len(self.Config.Wallet.MasterAccount.Child) == 0 {
		return balance, errors.New("No account is existed")
	}

	// convert params to array
	arrayParams := common.InterfaceSlice(params)

	// Param #1: account "*" for all or a particular account
	accountName := arrayParams[0].(string)

	// Param #2: the minimum number of confirmations an output must have
	min := int(arrayParams[1].(float64))
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase := arrayParams[2].(string)

	if passPhrase != self.Config.Wallet.PassPhrase {
		return balance, errors.New("Password phrase is wrong for local wallet")
	}

	if accountName == "*" {
		// get balance for all accounts in wallet
		for _, account := range self.Config.Wallet.MasterAccount.Child {
			txs, err := self.Config.BlockChain.GetListTxByPrivateKey(&account.Key.KeyPair.PrivateKey, common.TxOutCoinType, transaction.NoSort, false)
			if err != nil {
				return nil, err
			}
			for _, tx := range txs {
				for _, desc := range tx.Descs {
					notes := desc.GetNote()
					for _, note := range notes {
						balance += note.Value
					}
				}
			}
		}
	} else {
		for _, account := range self.Config.Wallet.MasterAccount.Child {
			if account.Name == accountName {
				// get balance for accountName in wallet
				txs, err := self.Config.BlockChain.GetListTxByPrivateKey(&account.Key.KeyPair.PrivateKey, common.TxOutCoinType, transaction.NoSort, false)
				if err != nil {
					return nil, err
				}
				for _, tx := range txs {
					for _, desc := range tx.Descs {
						notes := desc.GetNote()
						for _, note := range notes {
							balance += note.Value
						}
					}
				}
				break
			}
		}
	}

	return balance, nil
}

/**
handleGetReceivedByAccount -  RPC returns the total amount received by addresses in a particular account from transactions with the specified number of confirmations. It does not count coinbase transactions.
*/
func (self RpcServer) handleGetReceivedByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	balance := uint64(0)

	if self.Config.Wallet == nil {
		return balance, errors.New("Wallet is not existed")
	}
	if len(self.Config.Wallet.MasterAccount.Child) == 0 {
		return balance, errors.New("No account is existed")
	}

	// convert params to array
	arrayParams := common.InterfaceSlice(params)

	// Param #1: account "*" for all or a particular account
	accountName := arrayParams[0].(string)

	// Param #2: the minimum number of confirmations an output must have
	min := int(arrayParams[1].(float64))
	_ = min

	// Param #3: passphrase to access local wallet of node
	passPhrase := arrayParams[2].(string)

	if passPhrase != self.Config.Wallet.PassPhrase {
		return balance, errors.New("Password phrase is wrong for local wallet")
	}

	for _, account := range self.Config.Wallet.MasterAccount.Child {
		if account.Name == accountName {
			// get balance for accountName in wallet
			txs, err := self.Config.BlockChain.GetListTxByPrivateKey(&account.Key.KeyPair.PrivateKey, common.TxOutCoinType, transaction.NoSort, false)
			if err != nil {
				return nil, err
			}
			for _, tx := range txs {
				if blockchain.IsCoinBaseTx(&tx) {
					continue
				}
				for _, desc := range tx.Descs {
					notes := desc.GetNote()
					for _, note := range notes {
						balance += note.Value
					}
				}
			}
			break
		}
	}
	return balance, nil
}

/**
handleGetConnectionCount - RPC returns the number of connections to other nodes.
*/
func (self RpcServer) handleGetConnectionCount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if self.Config.ConnMgr == nil || len(self.Config.ConnMgr.ListeningPeers) == 0 {
		return 0, nil
	}
	result := 0
	for _, listeningPeer := range self.Config.ConnMgr.ListeningPeers {
		result += len(listeningPeer.PeerConns)
	}
	return result, nil
}

/**
handleGetGenerate - RPC returns true if the node is set to generate blocks using its CPU
*/
func (self RpcServer) handleGetGenerate(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.Config.IsGenerateNode, nil
}

/**
handleGetMempoolInfo - RPC returns information about the node's current txs memory pool
*/
func (self RpcServer) handleGetMempoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonrpc.GetMempoolInfo{}
	result.Size = self.Config.TxMemPool.Count()
	result.Bytes = self.Config.TxMemPool.Size()
	result.MempoolMaxFee = self.Config.TxMemPool.MaxFee()
	return result, nil
}

/**
handleGetMiningInfo - RPC returns various mining-related info
*/
func (self RpcServer) handleGetMiningInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if !self.Config.IsGenerateNode {
		return nil, errors.New("Not mining")
	}
	result := jsonrpc.GetMiningInfoResult{}
	result.Blocks = uint64(self.Config.BlockChain.BestState.BestBlock.Height + 1)
	result.PoolSize = self.Config.TxMemPool.Count()
	result.Chain = self.Config.ChainParams.Name
	result.CurrentBlockTx = len(self.Config.BlockChain.BestState.BestBlock.Transactions)
	return result, nil
}

/**
handleGetRawMempool - RPC returns all transaction ids in memory pool as a json array of string transaction ids
Hint: use getmempoolentry to fetch a specific transaction from the mempool.
*/
func (self RpcServer) handleGetRawMempool(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	txIds := self.Config.TxMemPool.ListTxs()
	return txIds, nil
}

/**
handleMempoolEntry - RPC fetch a specific transaction from the mempool
*/
func (self RpcServer) handleMempoolEntry(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Param #1: hash string of tx(tx id)
	txId, err := common.Hash{}.NewHashFromStr(params.(string))
	if err != nil {
		return nil, err
	}

	tx, err := self.Config.TxMemPool.GetTx(txId)
	return tx, err
}

/**
handleEstimateFee - RPC estimates the transaction fee per kilobyte that needs to be paid for a transaction to be included within a certain number of blocks.
*/
func (self RpcServer) handleEstimateFee(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Param #1: —how many blocks the transaction may wait before being included
	feeRate, err := self.Config.FeeEstimator.EstimateFee(uint32(params.(float64)))
	if err != nil {
		return -1, err
	}
	return uint64(feeRate), nil
}
