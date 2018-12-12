package rpcserver

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
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
	CheckHashValue:    RpcServer.handleCheckHashValue, // get data in blockchain from hash value
	GetBlockHeader:    RpcServer.handleGetBlockHeader, // Current committee, next block committee and candidate is included in block header

	// transaction
	ListTransactions:         RpcServer.handleListTransactions,
	CreateRawTransaction:     RpcServer.handleCreateRawTransaction,
	SendRawTransaction:       RpcServer.handleSendRawTransaction,
	CreateAndSendTransaction: RpcServer.handlCreateAndSendTx,
	GetMempoolInfo:           RpcServer.handleGetMempoolInfo,
	GetTransactionByHash:     RpcServer.handleGetTransactionByHash,

	GetCommitteeCandidateList:  RpcServer.handleGetCommitteeCandidateList,
	RetrieveCommitteeCandidate: RpcServer.handleRetrieveCommiteeCandidate,
	GetBlockProducerList:       RpcServer.handleGetBlockProducerList,

	// custom token
	CreateRawCustomTokenTransaction:     RpcServer.handleCreateRawCustomTokenTransaction,
	SendRawCustomTokenTransaction:       RpcServer.handleSendRawCustomTokenTransaction,
	CreateAndSendCustomTokenTransaction: RpcServer.handleCreateAndSendCustomTokenTransaction,
	ListUnspentCustomToken:              RpcServer.handleListUnspentCustomTokenTransaction,
	ListCustomToken:                     RpcServer.handleListCustomToken,
	CustomToken:                         RpcServer.handleCustomTokenDetail,
	GetListCustomTokenBalance:           RpcServer.handleGetListCustomTokenBalance,

	// Loan tx
	GetLoanParams:             RpcServer.handleGetLoanParams,
	CreateAndSendLoanRequest:  RpcServer.handleCreateAndSendLoanRequest,
	CreateAndSendLoanResponse: RpcServer.handleCreateAndSendLoanResponse,
	CreateAndSendLoanWithdraw: RpcServer.handleCreateAndSendLoanWithdraw,
	CreateAndSendLoanPayment:  RpcServer.handleCreateAndSendLoanPayment,

	// multisig
	CreateSignatureOnCustomTokenTx: RpcServer.handleCreateSignatureOnCustomTokenTx,
	GetListDCBBoard:                RpcServer.handleGetListDCBBoard,
	GetListCBBoard:                 RpcServer.handleGetListCBBoard,
	GetListGOVBoard:                RpcServer.handleGetListGOVBoard,

	// vote
	CreateAndSendVoteDCBBoardTransaction: RpcServer.handleCreateAndSendVoteDCBBoardTransaction,
	CreateRawVoteDCBBoardTx:              RpcServer.handleCreateRawVoteDCBBoardTransaction,
	SendRawVoteBoardDCBTx:                RpcServer.handleSendRawVoteBoardDCBTransaction,
	CreateAndSendVoteGOVBoardTransaction: RpcServer.handleCreateAndSendVoteGOVBoardTransaction,
	CreateRawVoteGOVBoardTx:              RpcServer.handleCreateRawVoteDCBBoardTransaction,
	SendRawVoteBoardGOVTx:                RpcServer.handleSendRawVoteBoardDCBTransaction,

	// Submit Proposal:
	CreateAndSendSubmitDCBProposalTx: RpcServer.handleCreateAndSendSubmitDCBProposalTransaction,
	CreateRawSubmitDCBProposalTx:     RpcServer.handleCreateRawSubmitDCBProposalTransaction,
	SendRawSubmitDCBProposalTx:       RpcServer.handleSendRawSubmitDCBProposalTransaction,
	CreateAndSendSubmitGOVProposalTx: RpcServer.handleCreateAndSendSubmitGOVProposalTransaction,
	CreateRawSubmitGOVProposalTx:     RpcServer.handleCreateRawSubmitGOVProposalTransaction,
	SendRawSubmitGOVProposalTx:       RpcServer.handleSendRawSubmitGOVProposalTransaction,

	// dcb
	GetDCBParams:       RpcServer.handleGetDCBParams,
	GetDCBConstitution: RpcServer.handleGetDCBConstitution,

	// gov
	GetBondTypes:       RpcServer.handleGetBondTypes,
	GetGOVConstitution: RpcServer.handleGetGOVConstitution,
	GetGOVParams:       RpcServer.handleGetGOVParams,
}

// Commands that are available to a limited user
var RpcLimited = map[string]commandHandler{
	// local WALLET
	ListAccounts:               RpcServer.handleListAccounts,
	GetAccount:                 RpcServer.handleGetAccount,
	GetAddressesByAccount:      RpcServer.handleGetAddressesByAccount,
	GetAccountAddress:          RpcServer.handleGetAccountAddress,
	DumpPrivkey:                RpcServer.handleDumpPrivkey,
	ImportAccount:              RpcServer.handleImportAccount,
	RemoveAccount:              RpcServer.handleRemoveAccount,
	ListUnspent:                RpcServer.handleListUnspent,
	GetBalance:                 RpcServer.handleGetBalance,
	GetBalanceByPrivatekey:     RpcServer.handleGetBalanceByPrivatekey,
	GetBalanceByPaymentAddress: RpcServer.handleGetBalanceByPaymentAddress,
	GetReceivedByAccount:       RpcServer.handleGetReceivedByAccount,
	SetTxFee:                   RpcServer.handleSetTxFee,
	EncryptData:                RpcServer.handleEncryptDataByPaymentAddress,
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
		txsMap, err := self.config.BlockChain.GetListUnspentTxByKeyset(&readonlyKey.KeySet, transaction.NoSort, false)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		listTxs := make([]jsonresult.ListUnspentResultItem, 0)
		for chainId, txs := range txsMap {
			for _, tx := range txs {
				item := jsonresult.ListUnspentResultItem{
					TxId:     tx.Hash().String(),
					OutCoins: make([]jsonresult.OutCoin, 0),
				}
				for _, outCoin := range tx.Proof.OutputCoins {
					item.OutCoins = append(item.OutCoins, jsonresult.OutCoin{
						SerialNumber:   base58.Base58Check{}.Encode(outCoin.CoinDetails.SerialNumber.Compress(), byte(0x00)),
						PublicKey:      base58.Base58Check{}.Encode(outCoin.CoinDetails.PublicKey.Compress(), byte(0x00)),
						Value:          outCoin.CoinDetails.Value,
						Info:           base58.Base58Check{}.Encode(outCoin.CoinDetails.Info[:], byte(0x00)),
						CoinCommitment: base58.Base58Check{}.Encode(outCoin.CoinDetails.CoinCommitment.Compress(), byte(0x00)),
						Randomness:     *outCoin.CoinDetails.Randomness,
						SNDerivator:    *outCoin.CoinDetails.SNDerivator,
					})
				}
				listTxs = append(listTxs, item)
			}
			//fmt.Println("listTxs in handleListUnspent", listTxs)

			if result.ListUnspentResultItems[priKeyStr] == nil {
				result.ListUnspentResultItems[priKeyStr] = map[byte][]jsonresult.ListUnspentResultItem{}
			}
			if result.ListUnspentResultItems[priKeyStr][chainId] == nil {
				result.ListUnspentResultItems[priKeyStr][chainId] = []jsonresult.ListUnspentResultItem{}
			}
			result.ListUnspentResultItems[priKeyStr][chainId] = listTxs
		}
	}
	return result, nil
}

func (self RpcServer) handleCheckHashValue(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	var (
		isTransaction bool
		isBlock       bool
	)
	arrayParams := common.InterfaceSlice(params)
	// param #1: transaction Hash
	Logger.log.Infof("Check hash value  input Param %+v", arrayParams[0].(string))
	hash, _ := common.Hash{}.NewHashFromStr(arrayParams[0].(string))

	// Check block
	_, err := self.config.BlockChain.GetBlockByBlockHash(hash)
	if err != nil {
		isBlock = false
	} else {
		isBlock = true
		result := jsonresult.HashValueDetail{
			IsBlock:       isBlock,
			IsTransaction: false,
		}
		return result, nil
	}
	_, _, _, _, err1 := self.config.BlockChain.GetTransactionByHash(hash)
	if err1 != nil {
		isTransaction = false
	} else {
		isTransaction = true
		result := jsonresult.HashValueDetail{
			IsBlock:       false,
			IsTransaction: isTransaction,
		}
		return result, nil
	}
	return jsonresult.HashValueDetail{
		IsBlock:       isBlock,
		IsTransaction: isTransaction,
	}, nil
}

// buildRawCustomTokenTransaction ...
func (self RpcServer) buildRawCustomTokenTransaction(
	params interface{},
) (*transaction.TxCustomToken, error) {
	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	lastByte := senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1]
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
		PropertyID:     tokenParamsRaw["TokenID"].(string),
		PropertyName:   tokenParamsRaw["TokenName"].(string),
		PropertySymbol: tokenParamsRaw["TokenSymbol"].(string),
		TokenTxType:    int(tokenParamsRaw["TokenTxType"].(float64)),
		Amount:         uint64(tokenParamsRaw["TokenAmount"].(float64)),
		Receiver:       transaction.CreateCustomTokenReceiverArray(tokenParamsRaw["TokenReceivers"]),
	}
	switch tokenParams.TokenTxType {
	case transaction.CustomTokenTransfer:
		{
			tokenID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			unspentTxTokenOuts, err := self.config.BlockChain.GetUnspentTxCustomTokenVout(senderKey.KeySet, tokenID)
			fmt.Println("buildRawCustomTokenTransaction ", unspentTxTokenOuts)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			if len(unspentTxTokenOuts) == 0 {
				return nil, NewRPCError(ErrUnexpected, errors.New("Balance of token is zero"))
			}
			txTokenIns := []transaction.TxTokenVin{}
			txTokenInsAmount := uint64(0)
			for _, out := range unspentTxTokenOuts {
				item := transaction.TxTokenVin{
					PaymentAddress:  out.PaymentAddress,
					TxCustomTokenID: out.GetTxCustomTokenID(),
					VoutIndex:       out.GetIndex(),
				}
				// create signature by keyset -> base58check.encode of txtokenout double hash
				signature, err := senderKey.KeySet.Sign(out.Hash()[:])
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				// add signature to TxTokenVin to use token utxo
				item.Signature = base58.Base58Check{}.Encode(signature, 0)
				txTokenIns = append(txTokenIns, item)
				txTokenInsAmount += out.Value
			}
			tokenParams.SetVins(txTokenIns)
			tokenParams.SetVinsAmount(txTokenInsAmount)
		}
	case transaction.CustomTokenInit:
		{
			if tokenParams.Receiver[0].Value != tokenParams.Amount { // Init with wrong max amount of custom token
				return nil, NewRPCError(ErrUnexpected, errors.New("Init with wrong max amount of property"))
			}
		}
	}

	totalAmmount := estimateFeeCoinPerKb

	// list unspent tx for estimation fee
	estimateTotalAmount := totalAmmount
	usableTxsMap, _ := self.config.BlockChain.GetListUnspentTxByKeyset(&senderKey.KeySet, transaction.SortByAmount, false)
	candidateTxs := make([]*transaction.Tx, 0)
	candidateTxsMap := make(map[byte][]*transaction.Tx)
	for chainId, usableTxs := range usableTxsMap {
		for _, temp := range usableTxs {
			for _, note := range temp.Proof.OutputCoins {
				amount := note.CoinDetails.Value
				estimateTotalAmount -= int64(amount)
			}
			txData := temp
			candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
			candidateTxs = append(candidateTxs, &txData)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// check real fee per TxNormal
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
	if totalAmmount > 0 {
		candidateTxsMap = make(map[byte][]*transaction.Tx, 0)
		for chainId, usableTxs := range usableTxsMap {
			for _, temp := range usableTxs {
				for _, note := range temp.Proof.OutputCoins {
					amount := note.CoinDetails.Value
					estimateTotalAmount -= int64(amount)
				}
				txData := temp
				candidateTxsMap[chainId] = append(candidateTxsMap[chainId], &txData)
				if estimateTotalAmount <= 0 {
					break
				}
			}
		}
	}

	// get list custom token
	listCustomTokens, err := self.config.BlockChain.ListCustomToken()

	tx, err := transaction.CreateTxCustomToken(
		&senderKey.KeySet.PrivateKey,
		nil,
		candidateTxsMap[chainIdSender],
		realFee,
		tokenParams,
		listCustomTokens,
	)

	return tx, err
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

// payment address -> balance of all custom token

// handleEncryptDataByPaymentAddress - get payment address and make an encrypted data
func (self RpcServer) handleEncryptDataByPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddress := arrayParams[0].(string)
	plainData := arrayParams[1].(string)
	keySet, err := wallet.Base58CheckDeserialize(paymentAddress)
	if err != nil {
		return nil, err
	}
	encryptData, err := keySet.KeySet.Encrypt([]byte(plainData))
	if err != nil {
		return nil, err
	}
	_ = encryptData
	return hex.EncodeToString([]byte{}), nil
}
