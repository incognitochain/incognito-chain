package rpcserver

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/wire"
	"golang.org/x/crypto/ed25519"
)

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, error)

// Commands valid for normal user
var RpcHandler = map[string]commandHandler{
	// node
	GetNetworkInfo:     RpcServer.handleGetNetWorkInfo,
	GetConnectionCount: RpcServer.handleGetConnectionCount,
	GetAllPeers:        RpcServer.handleGetAllPeers,
	GetRawMempool:      RpcServer.handleGetRawMempool,
	GetMempoolEntry:    RpcServer.handleMempoolEntry,
	EstimateFee:        RpcServer.handleEstimateFee,
	GetGenerate:        RpcServer.handleGetGenerate,
	GetMiningInfo:      RpcServer.handleGetMiningInfo,

	// block
	GetBestBlock:      RpcServer.handleGetBestBlock,
	GetBestBlockHash:  RpcServer.handleGetBestBlockHash,
	RetrieveBlock:     RpcServer.handleRetrieveBlock,
	GetBlocks:         RpcServer.handleGetBlocks,
	GetBlockChainInfo: RpcServer.handleGetBlockChainInfo,
	GetBlockCount:     RpcServer.handleGetBlockCount,
	GetBlockHash:      RpcServer.handleGetBlockHash,

	// transaction
	ListTransactions:                   RpcServer.handleListTransactions,
	CreateTransaction:                  RpcServer.handleCreateTransaction,
	SendTransaction:                    RpcServer.handleSendTransaction,
	SendMany:                           RpcServer.handleSendMany,
	GetNumberOfCoinsAndBonds:           RpcServer.handleGetNumberOfCoinsAndBonds,
	CreateActionParamsTransaction:      RpcServer.handleCreateActionParamsTransaction,
	SendRegistrationCandidateCommittee: RpcServer.handleSendRegistrationCandidateCommittee,
	CustomTokenTransaction:             RpcServer.handleCustomTokenTransaction,
	GetMempoolInfo:                     RpcServer.handleGetMempoolInfo,

	GetCommitteeCandidateList:  RpcServer.handleGetCommitteeCandidateList,
	RetrieveCommitteeCandidate: RpcServer.handleRetrieveCommiteeCandidate,
	GetBlockProducerList:       RpcServer.handleGetBlockProducerList,

	//POS
	GetHeader: RpcServer.handleGetHeader, // Current committee, next block committee and candidate is included in block header
}

// Commands that are available to a limited user
var RpcLimited = map[string]commandHandler{
	// local WALLET
	ListAccounts:          RpcServer.HandleListAccounts,
	GetAccount:            RpcServer.handleGetAccount,
	GetAddressesByAccount: RpcServer.handleGetAddressesByAccount,
	GetAccountAddress:     RpcServer.handleGetAccountAddress,
	DumpPrivkey:           RpcServer.handleDumpPrivkey,
	ImportAccount:         RpcServer.handleImportAccount,
	ListUnspent:           RpcServer.handleListUnspent,
	GetBalance:            RpcServer.handleGetBalance,
	GetReceivedByAccount:  RpcServer.handleGetReceivedByAccount,
	SetTxFee:              RpcServer.handleSetTxFee,
	CreateProducerKeyset:  RpcServer.handleCreateProducerKeySet,
}

func (self RpcServer) handleGetHeader(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	result := jsonresult.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	Logger.log.Info(arrayParams)
	getBy := arrayParams[0].(string)
	block := arrayParams[1].(string)
	chainID := arrayParams[2].(float64)
	switch getBy {
	case "blockhash":
		bhash := common.Hash{}
		err := bhash.Decode(&bhash, block)
		Logger.log.Info(bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Invalid blockhash format"))
		}
		block, err := self.config.BlockChain.GetBlockByBlockHash(&bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		result.Header = block.Header
		result.BlockNum = int(block.Header.Height) + 1
		result.ChainID = uint8(chainID)
		result.BlockHash = bhash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(block)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Invalid blocknum format"))
		}
		fmt.Println(chainID)
		if int32(bnum-1) > self.config.BlockChain.BestState[uint8(chainID)].Height || bnum <= 0 {
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		block, _ := self.config.BlockChain.GetBlockByBlockHeight(int32(bnum-1), uint8(chainID))
		result.Header = block.Header
		result.BlockNum = bnum
		result.ChainID = uint8(chainID)
		result.BlockHash = block.Hash().String()
	default:
		return nil, NewRPCError(ErrUnexpected, errors.New("Wrong request format"))
	}

	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (self RpcServer) handleGetNetWorkInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetNetworkInfoResult{}

	result.Version = RpcServerVersion
	result.SubVersion = ""
	result.ProtocolVersion = self.config.ProtocolVersion
	result.NetworkActive = len(self.config.ConnMgr.ListeningPeers) > 0
	result.LocalAddresses = []string{}
	for _, listener := range self.config.ConnMgr.ListeningPeers {
		result.Connections += len(listener.PeerConns)
		result.LocalAddresses = append(result.LocalAddresses, listener.RawAddress)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	networks := []map[string]interface{}{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
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
	result.Networks = networks
	result.IncrementalFee = self.config.Wallet.Config.IncrementalFee
	result.Warnings = ""

	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (self RpcServer) handleGetBestBlock(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetBestBlockResult{
		BestBlocks: make(map[string]jsonresult.GetBestBlockItem),
	}
	for chainID, best := range self.config.BlockChain.BestState {
		result.BestBlocks[strconv.Itoa(chainID)] = jsonresult.GetBestBlockItem{
			Height:   best.BestBlock.Header.Height,
			Hash:     best.BestBlockHash.String(),
			TotalTxs: best.TotalTxns,
		}
	}
	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (self RpcServer) handleGetBestBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetBestBlockHashResult{
		BestBlockHashes: make(map[string]string),
	}
	for chainID, best := range self.config.BlockChain.BestState {
		result.BestBlockHashes[strconv.Itoa(chainID)] = best.BestBlockHash.String()
	}
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (self RpcServer) handleRetrieveBlock(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString := paramsT[0].(string)
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			return nil, NewRPCError(ErrUnexpected, errH)
		}
		block, errD := self.config.BlockChain.GetBlockByBlockHash(hash)
		if errD != nil {
			return nil, NewRPCError(ErrUnexpected, errD)
		}
		result := jsonresult.GetBlockResult{}

		verbosity := paramsT[1].(string)

		chainId := block.Header.ChainID

		if verbosity == "0" {
			data, err := json.Marshal(block)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			result.Data = hex.EncodeToString(data)
		} else if verbosity == "1" {
			best := self.config.BlockChain.BestState[chainId]

			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			if blockHeight < best.Height {
				nextHash, err := self.config.BlockChain.GetBlockByBlockHeight(blockHeight+1, chainId)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				nextHashString = nextHash.Hash().String()
			}

			result.Hash = block.Hash().String()
			result.Confirmations = int64(1 + best.Height - blockHeight)
			result.Height = block.Header.Height
			result.Version = block.Header.Version
			result.MerkleRoot = block.Header.MerkleRoot.String()
			result.Time = block.Header.Timestamp
			result.ChainID = block.Header.ChainID
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.TxHashes = []string{}
			result.BlockProducerSign = block.BlockProducerSig
			for _, tx := range block.Transactions {
				result.TxHashes = append(result.TxHashes, tx.Hash().String())
			}
		} else if verbosity == "2" {
			best := self.config.BlockChain.BestState[chainId]

			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			if blockHeight < best.Height {
				nextHash, err := self.config.BlockChain.GetBlockByBlockHeight(blockHeight+1, chainId)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				nextHashString = nextHash.Hash().String()
			}

			result.Hash = block.Hash().String()
			result.Confirmations = int64(1 + best.Height - blockHeight)
			result.Height = block.Header.Height
			result.Version = block.Header.Version
			result.MerkleRoot = block.Header.MerkleRoot.String()
			result.Time = block.Header.Timestamp
			result.ChainID = block.Header.ChainID
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.Txs = make([]jsonresult.GetBlockTxResult, 0)
			for _, tx := range block.Transactions {
				transactionT := jsonresult.GetBlockTxResult{}

				transactionT.Hash = tx.Hash().String()
				if tx.GetType() == common.TxNormalType || tx.GetType() == common.TxSalaryType {
					txN := tx.(*transaction.Tx)
					data, err := json.Marshal(txN)
					if err != nil {
						return nil, err
					}
					transactionT.HexData = hex.EncodeToString(data)
					transactionT.Locktime = txN.LockTime
				} else if tx.GetType() == common.TxActionParamsType {
					txA := tx.(*transaction.ActionParamTx)
					data, err := json.Marshal(txA)
					if err != nil {
						return nil, NewRPCError(ErrUnexpected, err)
					}
					transactionT.HexData = hex.EncodeToString(data)
					transactionT.Locktime = txA.LockTime
				}
				result.Txs = append(result.Txs, transactionT)
			}
		}

		return result, nil
	}
	return nil, nil
}

// handleGetBlocks - get n top blocks from chain ID
func (self RpcServer) handleGetBlocks(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := make([]jsonresult.GetBlockResult, 0)
	arrayParams := common.InterfaceSlice(params)
	numBlock := int(arrayParams[0].(float64))
	chainID := int(arrayParams[1].(float64))
	bestBlock := self.config.BlockChain.BestState[chainID].BestBlock
	blockResult := jsonresult.GetBlockResult{}
	blockResult.Init(bestBlock)
	result = append(result, blockResult)
	for numBlock > 0 {
		numBlock--
		emptyHash := common.Hash{}
		if blockResult.PreviousBlockHash == emptyHash.String() {
			break
		}
		hash, errH := common.Hash{}.NewHashFromStr(blockResult.PreviousBlockHash)
		if errH != nil {
			return nil, errH
		}
		block, errD := self.config.BlockChain.GetBlockByBlockHash(hash)
		if errD != nil {
			return nil, errD
		}
		blockResult := jsonresult.GetBlockResult{}
		blockResult.Init(block)
		result = append(result, blockResult)
	}
	return result, nil
}

/*
getblockchaininfo RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetBlockChainInfoResult{
		ChainName:  self.config.ChainParams.Name,
		BestBlocks: make(map[string]jsonresult.GetBestBlockItem),
	}
	for chainID, bestState := range self.config.BlockChain.BestState {
		result.BestBlocks[strconv.Itoa(chainID)] = jsonresult.GetBestBlockItem{
			Height:           bestState.BestBlock.Header.Height,
			Hash:             bestState.BestBlockHash.String(),
			TotalTxs:         bestState.TotalTxns,
			SalaryFund:       bestState.BestBlock.Header.SalaryFund,
			SalaryPerTx:      bestState.BestBlock.Header.GovernanceParams.SalaryPerTx,
			BlockProducer:    bestState.BestBlock.BlockProducer,
			BlockProducerSig: bestState.BestBlock.BlockProducerSig,
		}
	}
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	chainId := byte(int(params.(float64)))
	if self.config.BlockChain.BestState != nil && self.config.BlockChain.BestState[chainId] != nil && self.config.BlockChain.BestState[chainId].BestBlock != nil {
		return self.config.BlockChain.BestState[chainId].BestBlock.Header.Height + 1, nil
	}
	return 0, nil
}

/*
getblockhash RPC return information fo blockchain node
*/
func (self RpcServer) handleGetBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	chainId := byte(int(arrayParams[0].(float64)))
	height := int32(arrayParams[1].(float64))
	hash, err := self.config.BlockChain.GetBlockByBlockHeight(height, chainId)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return hash.Hash().String(), nil
}

/*
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
	Logger.log.Info(params)
	result := jsonresult.ListUnspentResult{
		ListUnspentResultItems: make(map[string]map[byte][]jsonresult.ListUnspentResultItem),
	}

	// get params
	paramsArray := common.InterfaceSlice(params)
	assetType := paramsArray[0].(string)
	if ok, err := common.SliceExists(common.ListAsset, assetType); !ok || err != nil {
		return nil, NewRPCError(ErrUnexpected, errors.New(fmt.Sprintf("Asset is not in list: ", common.ListAsset)))
	}
	listKeyParams := common.InterfaceSlice(paramsArray[1])
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain readonly-key by deserializing
		readonlyKeyStr := keys["ReadonlyKey"].(string)
		readonlyKey, err := wallet.Base58CheckDeserialize(readonlyKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}

		// get keyset only contain pub-key by deserializing
		pubKeyStr := keys["PublicKey"].(string)
		pubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}

		// create a key set
		keySet := cashec.KeySet{
			ReadonlyKey: readonlyKey.KeySet.ReadonlyKey,
			PublicKey:   pubKey.KeySet.PublicKey,
		}

		txsMap, err := self.config.BlockChain.GetListTxByReadonlyKey(&keySet, assetType)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		listTxs := make([]jsonresult.ListUnspentResultItem, 0)
		for chainId, txs := range txsMap {
			for _, tx := range txs {
				item := jsonresult.ListUnspentResultItem{
					TxId:          tx.Hash().String(),
					JoinSplitDesc: make([]jsonresult.JoinSplitDesc, 0),
				}
				for _, desc := range tx.Descs {
					notes := desc.GetNote()
					amounts := make([]uint64, 0)
					for _, note := range notes {
						amounts = append(amounts, note.Value)
					}
					item.JoinSplitDesc = append(item.JoinSplitDesc, jsonresult.JoinSplitDesc{
						Anchors:     desc.Anchor,
						Commitments: desc.Commitments,
						Amounts:     amounts,
					})
				}
				listTxs = append(listTxs, item)
			}
			result.ListUnspentResultItems[readonlyKeyStr][chainId] = listTxs
		}
	}

	return result, nil
}

/*
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
	result := jsonresult.ListUnspentResult{
		ListUnspentResultItems: make(map[string]map[byte][]jsonresult.ListUnspentResultItem),
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
			return nil, NewRPCError(ErrUnexpected, err)
		}

		txsMap, err := self.config.BlockChain.GetListTxByPrivateKey(&readonlyKey.KeySet.PrivateKey, common.AssetTypeCoin, transaction.NoSort, false)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		listTxs := make([]jsonresult.ListUnspentResultItem, 0)
		for chainId, txs := range txsMap {
			for _, tx := range txs {
				item := jsonresult.ListUnspentResultItem{
					TxId:          tx.Hash().String(),
					JoinSplitDesc: make([]jsonresult.JoinSplitDesc, 0),
				}
				for _, desc := range tx.Descs {
					notes := desc.GetNote()
					amounts := make([]uint64, 0)
					for _, note := range notes {
						amounts = append(amounts, note.Value)
					}
					item.JoinSplitDesc = append(item.JoinSplitDesc, jsonresult.JoinSplitDesc{
						Anchors:     desc.Anchor,
						Commitments: desc.Commitments,
						Amounts:     amounts,
					})
				}
				listTxs = append(listTxs, item)
			}
			result.ListUnspentResultItems[priKeyStr][chainId] = listTxs
		}
	}
	return result, nil
}

func (self RpcServer) handleCreateRegistrationCandidateCommittee(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)

	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, nil
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	lastByte := senderKey.KeySet.PublicKey.Apk[len(senderKey.KeySet.PublicKey.Apk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if err != nil {
		return nil, nil
	}

	// param #2: list receiver
	totalAmmount := int64(0)
	receiversParam := int64(arrayParams[1].(float64))
	paymentInfos := []*client.PaymentInfo{
		&client.PaymentInfo{
			Amount: common.ConstantToMiliConstant(uint64(receiversParam)),
		},
	}
	totalAmmount = int64(common.ConstantToMiliConstant(uint64(receiversParam)))

	// param #3: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: estimation fee coin per kb
	numBlock := uint32(arrayParams[3].(float64))

	nodeAddr := arrayParams[4].(string)
	if valid := common.ValidateNodeAddress(nodeAddr); !valid {
		return nil, errors.New("node address is wrong")
	}

	// ONLY use CONSTANT COIN for registration candidate tx
	assetType := common.AssetTypeCoin

	// list unspent tx for estimation fee
	estimateTotalAmount := totalAmmount
	usableTxsMap, _ := self.config.BlockChain.GetListTxByPrivateKey(&senderKey.KeySet.PrivateKey, assetType, transaction.SortByAmount, false)
	candidateTxs := make([]*transaction.Tx, 0)
	candidateTxsMap := make(map[byte][]*transaction.Tx)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, desc := range temp.Descs {
				for _, note := range desc.GetNote() {
					amount := note.Value
					estimateTotalAmount -= int64(amount)
				}
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			candidateTxs = append(candidateTxs, &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// check real fee per Tx
	var realFee uint64
	if int64(estimateFeeCoinPerKb) == -1 {
		temp, _ := self.config.FeeEstimator[chainIdSender].EstimateFee(numBlock)
		estimateFeeCoinPerKb = int64(temp)
	}
	estimateFeeCoinPerKb += int64(self.config.Wallet.Config.IncrementalFee)
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateTxs, paymentInfos)
	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)

	// list unspent tx for create tx
	totalAmmount += int64(realFee)
	candidateTxsMap = make(map[byte][]*transaction.Tx, 0)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, desc := range temp.Descs {
				for _, note := range desc.GetNote() {
					amount := note.Value
					estimateTotalAmount -= int64(amount)
				}
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// get merkleroot commitments, nullifers db, commitments db for every chain
	nullifiersDb := make(map[byte]([][]byte))
	commitmentsDb := make(map[byte]([][]byte))
	merkleRootCommitments := make(map[byte]*common.Hash)
	for chainId, _ := range candidateTxsMap {
		merkleRootCommitments[chainId] = &self.config.BlockChain.BestState[chainId].BestBlock.Header.MerkleRootCommitments
		// get tx view point
		txViewPoint, _ := self.config.BlockChain.FetchTxViewPoint(common.AssetTypeCoin, chainId)
		nullifiersDb[chainId] = txViewPoint.ListNullifiers(common.AssetTypeCoin)
		commitmentsDb[chainId] = txViewPoint.ListCommitments(common.AssetTypeCoin)
	}

	fmt.Println("Create voting Tx ...")
	tx, err := transaction.CreateVotingTx(&senderKey.KeySet.PrivateKey, paymentInfos,
		merkleRootCommitments,
		candidateTxsMap,
		nullifiersDb,
		commitmentsDb,
		realFee,
		assetType,
		chainIdSender,
		nodeAddr)
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

func (self RpcServer) handleSendRawRegistrationCandidateCommittee(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	var tx transaction.TxVoting
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, err
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, err
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdRegisteration)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageRegistration).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

// handleSendRegistrationCandidateCommittee handle sendregistration committee candidate command.
func (self RpcServer) handleSendRegistrationCandidateCommittee(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	hexStrOfTx, err := self.handleCreateRegistrationCandidateCommittee(params, closeChan)
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	txId, err := self.handleSendRawRegistrationCandidateCommittee(newParam, closeChan)
	return txId, err
}

// handleCustomTokenTransaction handle create a custom token command.
func (self RpcServer) handleCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	lastByte := senderKey.KeySet.PublicKey.Apk[len(senderKey.KeySet.PublicKey.Apk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// param #2: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[1].(float64))

	// param #3: estimation fee coin per kb by numblock
	numBlock := uint32(arrayParams[2].(float64))

	// param #4: token params
	tokenParamsRaw := arrayParams[3].(map[string]interface{})

	tokenParams := &transaction.CustomTokenParamTx{
		PropertyName:    tokenParamsRaw["TokenName"].(string),
		PropertySymbol:  tokenParamsRaw["TokenSymbol"].(string),
		TxCustomTokenID: tokenParamsRaw["TokenHash"].(string),
		TokenTxType:     tokenParamsRaw["TokenTxType"].(float64),
		Amount:          tokenParamsRaw["TokenAmount"].(float64),
		Receivers:       transaction.CreateCustomTokenReceiverArray(tokenParamsRaw["TokenReceivers"]),
	}

	totalAmmount := estimateFeeCoinPerKb

	// list unspent tx for estimation fee
	estimateTotalAmount := totalAmmount
	usableTxsMap, _ := self.config.BlockChain.GetListTxByPrivateKey(&senderKey.KeySet.PrivateKey, common.AssetTypeCoin, transaction.SortByAmount, false)
	candidateTxs := make([]*transaction.Tx, 0)
	candidateTxsMap := make(map[byte][]*transaction.Tx)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, desc := range temp.Descs {
				for _, note := range desc.GetNote() {
					amount := note.Value
					estimateTotalAmount -= int64(amount)
				}
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			candidateTxs = append(candidateTxs, &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// check real fee per Tx
	var realFee uint64
	if int64(estimateFeeCoinPerKb) == -1 {
		temp, _ := self.config.FeeEstimator[chainIdSender].EstimateFee(numBlock)
		estimateFeeCoinPerKb = int64(temp)
	}
	estimateFeeCoinPerKb += int64(self.config.Wallet.Config.IncrementalFee)
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateTxs, nil)
	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)

	// list unspent tx for create tx
	totalAmmount += int64(realFee)
	candidateTxsMap = make(map[byte][]*transaction.Tx, 0)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, desc := range temp.Descs {
				for _, note := range desc.GetNote() {
					amount := note.Value
					estimateTotalAmount -= int64(amount)
				}
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// get merkleroot commitments, nullifers db, commitments db for every chain
	nullifiersDb := make(map[byte]([][]byte))
	commitmentsDb := make(map[byte]([][]byte))
	merkleRootCommitments := make(map[byte]*common.Hash)
	for chainId, _ := range candidateTxsMap {
		merkleRootCommitments[chainId] = &self.config.BlockChain.BestState[chainId].BestBlock.Header.MerkleRootCommitments
		// get tx view point
		txViewPoint, _ := self.config.BlockChain.FetchTxViewPoint(common.AssetTypeCoin, chainId)
		nullifiersDb[chainId] = txViewPoint.ListNullifiers(common.AssetTypeCoin)
		commitmentsDb[chainId] = txViewPoint.ListCommitments(common.AssetTypeCoin)
	}

	tx, err := transaction.CreateTxCustomToken(&senderKey.KeySet.PrivateKey, nil,
		merkleRootCommitments,
		candidateTxsMap,
		nullifiersDb,
		commitmentsDb,
		realFee,
		common.AssetTypeCoin,
		chainIdSender,
		tokenParams)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

/*
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
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	lastByte := senderKey.KeySet.PublicKey.Apk[len(senderKey.KeySet.PublicKey.Apk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// param #2: list receiver
	totalAmmount := int64(0)
	receiversParam := arrayParams[1].(map[string]interface{})
	paymentInfos := make([]*client.PaymentInfo, 0)
	for pubKeyStr, amount := range receiversParam {
		receiverPubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		paymentInfo := &client.PaymentInfo{
			Amount:         common.ConstantToMiliConstant(uint64(amount.(float64))),
			PaymentAddress: receiverPubKey.KeySet.PublicKey,
		}
		totalAmmount += int64(paymentInfo.Amount)
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// param #3: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: estimation fee coin per kb by numblock
	numBlock := uint32(arrayParams[3].(float64))

	nodeAddr := arrayParams[4].(string)
	if valid := common.ValidateNodeAddress(nodeAddr); !valid {
		return nil, errors.New("node address is wrong")
	}

	// list unspent tx for estimation fee
	estimateTotalAmount := totalAmmount
	usableTxsMap, _ := self.config.BlockChain.GetListTxByPrivateKey(&senderKey.KeySet.PrivateKey, common.AssetTypeCoin, transaction.SortByAmount, false)
	candidateTxs := make([]*transaction.Tx, 0)
	candidateTxsMap := make(map[byte][]*transaction.Tx)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, desc := range temp.Descs {
				for _, note := range desc.GetNote() {
					amount := note.Value
					estimateTotalAmount -= int64(amount)
				}
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			candidateTxs = append(candidateTxs, &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// check real fee per Tx
	var realFee uint64
	if int64(estimateFeeCoinPerKb) == -1 {
		temp, _ := self.config.FeeEstimator[chainIdSender].EstimateFee(numBlock)
		estimateFeeCoinPerKb = int64(temp)
	}
	estimateFeeCoinPerKb += int64(self.config.Wallet.Config.IncrementalFee)
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateTxs, paymentInfos)
	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)

	// list unspent tx for create tx
	totalAmmount += int64(realFee)
	candidateTxsMap = make(map[byte][]*transaction.Tx, 0)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, desc := range temp.Descs {
				for _, note := range desc.GetNote() {
					amount := note.Value
					estimateTotalAmount -= int64(amount)
				}
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// get merkleroot commitments, nullifers db, commitments db for every chain
	nullifiersDb := make(map[byte]([][]byte))
	commitmentsDb := make(map[byte]([][]byte))
	merkleRootCommitments := make(map[byte]*common.Hash)
	for chainId, _ := range candidateTxsMap {
		merkleRootCommitments[chainId] = &self.config.BlockChain.BestState[chainId].BestBlock.Header.MerkleRootCommitments
		// get tx view point
		txViewPoint, _ := self.config.BlockChain.FetchTxViewPoint(common.AssetTypeCoin, chainId)
		nullifiersDb[chainId] = txViewPoint.ListNullifiers(common.AssetTypeCoin)
		commitmentsDb[chainId] = txViewPoint.ListCommitments(common.AssetTypeCoin)
	}

	tx, err := transaction.CreateTx(&senderKey.KeySet.PrivateKey, paymentInfos,
		merkleRootCommitments,
		candidateTxsMap,
		nullifiersDb,
		commitmentsDb,
		realFee,
		common.AssetTypeCoin,
		chainIdSender)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrUnexpected, err)
	}
	hexData := hex.EncodeToString(byteArrays)
	result := jsonresult.CreateTransactionResult{
		HexData: hexData,
	}
	return result, nil
}

/*
// handleSendTransaction implements the sendtransaction command.
Parameter #1—a serialized transaction to broadcast
Parameter #2–whether to allow high fees
Result—a TXID or error Message
*/
func (self RpcServer) handleSendTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	var tx transaction.Tx
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast Message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	err = self.config.Server.PushMessageToAll(txMsg)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txID := tx.Hash().String()
	result := jsonresult.CreateTransactionResult{
		TxID: txID,
	}
	return result, nil
}

/*
handleSendMany - RPC creates transaction and send to network
*/
func (self RpcServer) handleSendMany(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	hexStrOfTx, err := self.handleCreateTransaction(params, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, hexStrOfTx)
	sendResult, err := self.handleSendTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

/*
 * handleGetNumberOfCoins handles getNumberOfCoins commands.
 */
func (self RpcServer) handleGetNumberOfCoinsAndBonds(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	result, err := self.config.BlockChain.GetAllUnitCoinSupplier()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
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

/*
// handleCreateRawTransaction handles createrawtransaction commands.
*/
func (self RpcServer) handleCreateActionParamsTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
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

	_, _, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// broadcast Message
	// self.config.Server.PushTxMessage(hash)

	return tx.Hash(), nil
}

/*
listaccount RPC lists accounts and their balances.

Parameter #1—the minimum number of confirmations a transaction must have
Parameter #2—whether to include watch-only addresses in results
Result—a list of accounts and their balances

*/
func (self RpcServer) HandleListAccounts(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.ListAccounts{
		Accounts:   make(map[string]uint64),
		WalletName: self.config.Wallet.Name,
	}
	accounts := self.config.Wallet.ListAccounts()
	for accountName, account := range accounts {
		txsMap, err := self.config.BlockChain.GetListTxByPrivateKey(&account.Key.KeySet.PrivateKey, common.AssetTypeCoin, transaction.NoSort, false)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		amount := uint64(0)
		for _, txs := range txsMap {
			for _, tx := range txs {
				for _, desc := range tx.Descs {
					notes := desc.GetNote()
					for _, note := range notes {
						amount += note.Value
					}
				}
			}
		}
		result.Accounts[accountName] = amount
	}

	return result, nil
}

/*
getaccount RPC returns the name of the account associated with the given address.
- Param #1: address
*/
func (self RpcServer) handleGetAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	for _, account := range self.config.Wallet.MasterAccount.Child {
		address := account.Key.Base58CheckSerialize(wallet.PubKeyType)
		if address == params.(string) {
			return account.Name, nil
		}
	}
	return nil, nil
}

/*
getaddressesbyaccount RPC returns a list of every address assigned to a particular account.

Parameter #1—the account name
Result—a list of addresses
*/
func (self RpcServer) handleGetAddressesByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetAddressesByAccount{}
	result.Addresses = self.config.Wallet.GetAddressesByAccount(params.(string))
	return result, nil
}

/*
getaccountaddress RPC returns the current coin address for receiving payments to this account. If the account doesn’t exist, it creates both the account and a new address for receiving payment. Once a payment has been received to an address, future calls to this RPC for the same account will return a different address.
Parameter #1—an account name
Result—a constant address
*/
func (self RpcServer) handleGetAccountAddress(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.config.Wallet.GetAccountAddress(params.(string)), nil
}

/*
 dumpprivkey RPC returns the wallet-import-format (WIP) private key corresponding to an address. (But does not remove it from the wallet.)

Parameter #1—the address corresponding to the private key to get
Result—the private key
*/
func (self RpcServer) handleDumpPrivkey(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.config.Wallet.DumpPrivkey(params.(string)), nil
}

/*
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
	account, err := self.config.Wallet.ImportAccount(privateKey, accountName, passPhrase)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return wallet.KeySerializedData{
		PublicKey:   account.Key.Base58CheckSerialize(wallet.PubKeyType),
		ReadonlyKey: account.Key.Base58CheckSerialize(wallet.ReadonlyKeyType),
	}, nil
}

/*
handleGetAllPeers - return all peers which this node connected
*/
func (self RpcServer) handleGetAllPeers(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	result := jsonresult.GetAllPeersResult{}
	peersMap := []string{}
	peers := self.config.AddrMgr.AddressCache()
	for _, peer := range peers {
		for _, peerConn := range peer.PeerConns {
			peersMap = append(peersMap, peerConn.RemoteRawAddress)
		}
	}
	result.Peers = peersMap
	return result, nil
}

/*
handleGetBalance - RPC gets the balances in decimal
*/
func (self RpcServer) handleGetBalance(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	balance := uint64(0)

	if self.config.Wallet == nil {
		return balance, NewRPCError(ErrUnexpected, errors.New("Wallet is not existed"))
	}
	if len(self.config.Wallet.MasterAccount.Child) == 0 {
		return balance, NewRPCError(ErrUnexpected, errors.New("No account is existed"))
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

	if passPhrase != self.config.Wallet.PassPhrase {
		return balance, NewRPCError(ErrUnexpected, errors.New("Password phrase is wrong for local wallet"))
	}

	if accountName == "*" {
		// get balance for all accounts in wallet
		for _, account := range self.config.Wallet.MasterAccount.Child {
			txsMap, err := self.config.BlockChain.GetListTxByPrivateKey(&account.Key.KeySet.PrivateKey, common.AssetTypeCoin, transaction.NoSort, false)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, txs := range txsMap {
				for _, tx := range txs {
					for _, desc := range tx.Descs {
						notes := desc.GetNote()
						for _, note := range notes {
							balance += note.Value
						}
					}
				}
			}
		}
	} else {
		for _, account := range self.config.Wallet.MasterAccount.Child {
			if account.Name == accountName {
				// get balance for accountName in wallet
				txsMap, err := self.config.BlockChain.GetListTxByPrivateKey(&account.Key.KeySet.PrivateKey, common.AssetTypeCoin, transaction.NoSort, false)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				for _, txs := range txsMap {
					for _, tx := range txs {
						for _, desc := range tx.Descs {
							notes := desc.GetNote()
							for _, note := range notes {
								balance += note.Value
							}
						}
					}
				}
				break
			}
		}
	}

	return balance, nil
}

/*
handleGetReceivedByAccount -  RPC returns the total amount received by addresses in a particular account from transactions with the specified number of confirmations. It does not count salary transactions.
*/
func (self RpcServer) handleGetReceivedByAccount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	balance := uint64(0)

	if self.config.Wallet == nil {
		return balance, NewRPCError(ErrUnexpected, errors.New("Wallet is not existed"))
	}
	if len(self.config.Wallet.MasterAccount.Child) == 0 {
		return balance, NewRPCError(ErrUnexpected, errors.New("No account is existed"))
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

	if passPhrase != self.config.Wallet.PassPhrase {
		return balance, NewRPCError(ErrUnexpected, errors.New("Password phrase is wrong for local wallet"))
	}

	for _, account := range self.config.Wallet.MasterAccount.Child {
		if account.Name == accountName {
			// get balance for accountName in wallet
			txsMap, err := self.config.BlockChain.GetListTxByPrivateKey(&account.Key.KeySet.PrivateKey, common.AssetTypeCoin, transaction.NoSort, false)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, txs := range txsMap {
				for _, tx := range txs {
					if blockchain.IsSalaryTx(&tx) {
						continue
					}
					for _, desc := range tx.Descs {
						notes := desc.GetNote()
						for _, note := range notes {
							balance += note.Value
						}
					}
				}
			}
			break
		}
	}
	return balance, nil
}

/*
handleGetConnectionCount - RPC returns the number of connections to other nodes.
*/
func (self RpcServer) handleGetConnectionCount(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if self.config.ConnMgr == nil || len(self.config.ConnMgr.ListeningPeers) == 0 {
		return 0, nil
	}
	result := 0
	for _, listeningPeer := range self.config.ConnMgr.ListeningPeers {
		result += len(listeningPeer.PeerConns)
	}
	return result, nil
}

/*
handleGetGenerate - RPC returns true if the node is set to generate blocks using its CPU
*/
func (self RpcServer) handleGetGenerate(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.config.IsGenerateNode, nil
}

/*
handleGetMempoolInfo - RPC returns information about the node's current txs memory pool
*/
func (self RpcServer) handleGetMempoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetMempoolInfo{}
	result.Size = self.config.TxMemPool.Count()
	result.Bytes = self.config.TxMemPool.Size()
	result.MempoolMaxFee = self.config.TxMemPool.MaxFee()
	return result, nil
}

/*
handleGetMiningInfo - RPC returns various mining-related info
*/
func (self RpcServer) handleGetMiningInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if !self.config.IsGenerateNode {
		return nil, NewRPCError(ErrUnexpected, errors.New("Not mining"))
	}
	chainId := byte(int(params.(float64)))
	result := jsonresult.GetMiningInfoResult{}
	result.Blocks = uint64(self.config.BlockChain.BestState[chainId].BestBlock.Header.Height + 1)
	result.PoolSize = self.config.TxMemPool.Count()
	result.Chain = self.config.ChainParams.Name
	result.CurrentBlockTx = len(self.config.BlockChain.BestState[chainId].BestBlock.Transactions)
	return result, nil
}

/*
handleGetRawMempool - RPC returns all transaction ids in memory pool as a json array of string transaction ids
Hint: use getmempoolentry to fetch a specific transaction from the mempool.
*/
func (self RpcServer) handleGetRawMempool(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonresult.GetRawMempoolResult{
		TxHashes: self.config.TxMemPool.ListTxs(),
	}
	return result, nil
}

/*
handleMempoolEntry - RPC fetch a specific transaction from the mempool
*/
func (self RpcServer) handleMempoolEntry(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Param #1: hash string of tx(tx id)
	txID, err := common.Hash{}.NewHashFromStr(params.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	result := jsonresult.GetMempoolEntryResult{}
	result.Tx, err = self.config.TxMemPool.GetTx(txID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return result, nil
}

/*
handleEstimateFee - RPC estimates the transaction fee per kilobyte that needs to be paid for a transaction to be included within a certain number of blocks.
*/
func (self RpcServer) handleEstimateFee(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Param #1: —how many blocks the transaction may wait before being included
	arrayParams := common.InterfaceSlice(params)
	numBlock := uint32(arrayParams[0].(float64))
	result := jsonresult.EstimateFeeResult{
		FeeRate: make(map[string]uint64),
	}
	for chainID, feeEstimator := range self.config.FeeEstimator {
		feeRate, err := feeEstimator.EstimateFee(numBlock)
		result.FeeRate[strconv.Itoa(int(chainID))] = uint64(feeRate)
		if err != nil {
			return -1, NewRPCError(ErrUnexpected, err)
		}
	}
	return result, nil
}

/*
handleSetTxFee - RPC sets the transaction fee per kilobyte paid more by transactions created by this wallet. default is 1 coin per 1 kb
*/
func (self RpcServer) handleSetTxFee(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	self.config.Wallet.Config.IncrementalFee = uint64(params.(float64))
	err := self.config.Wallet.Save(self.config.Wallet.PassPhrase)
	return err == nil, NewRPCError(ErrUnexpected, err)
}

func (self RpcServer) handleCreateProducerKeySet(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// param #1: private key of sender
	senderKey, err := wallet.Base58CheckDeserialize(params.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	producerKeySet, err := senderKey.KeySet.CreateProducerKeySet()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := make(map[string]string)
	result["ProducerKeySet"] = producerKeySet.EncodeToString()
	result["ProducerPublicKey"] = base58.Base58Check{}.Encode(producerKeySet.SpublicKey, byte(0x00))
	return result, nil
}

func (self RpcServer) handleGetCommitteeCandidateList(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// param #1: private key of sender
	cndList := self.config.BlockChain.GetCommitteeCandidateList()
	return cndList, nil
}

func (self RpcServer) handleRetrieveCommiteeCandidate(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	candidateInfo := self.config.BlockChain.GetCommitteCandidate(params.(string))
	if candidateInfo == nil {
		return nil, nil
	}
	result := jsonresult.RetrieveCommitteecCandidateResult{}
	result.Init(candidateInfo)
	return result, nil
}

func (self RpcServer) handleGetBlockProducerList(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := make(map[string]string)
	for chainID, bestState := range self.config.BlockChain.BestState {
		if bestState.BestBlock.BlockProducer != "" {
			result[strconv.Itoa(chainID)] = bestState.BestBlock.BlockProducer
		} else {
			result[strconv.Itoa(chainID)] = self.config.ChainParams.GenesisBlock.Header.Committee[chainID]
		}
	}
	return result, nil
}
