package rpcserver

import (
	"encoding/hex"
	"net"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
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
	ListOutputCoins:          RpcServer.handleListOutputCoins,
	CreateRawTransaction:     RpcServer.handleCreateRawTransaction,
	SendRawTransaction:       RpcServer.handleSendRawTransaction,
	CreateAndSendTransaction: RpcServer.handlCreateAndSendTx,
	GetMempoolInfo:           RpcServer.handleGetMempoolInfo,
	GetTransactionByHash:     RpcServer.handleGetTransactionByHash,

	GetCommitteeCandidateList:  RpcServer.handleGetCommitteeCandidateList,
	RetrieveCommitteeCandidate: RpcServer.handleRetrieveCommiteeCandidate,
	GetBlockProducerList:       RpcServer.handleGetBlockProducerList,

	RandomCommitments: RpcServer.handleRandomCommitments,
	HasSerialNumbers:  RpcServer.handleRandomCommitments,

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
	GetDCBParams:                          RpcServer.handleGetDCBParams,
	GetDCBConstitution:                    RpcServer.handleGetDCBConstitution,
	CreateAndSendTxWithIssuingRequest:     RpcServer.handleCreateAndSendTxWithIssuingRequest,
	CreateAndSendTxWithContractingRequest: RpcServer.handleCreateAndSendTxWithContractingRequest,

	// gov
	GetBondTypes:                      RpcServer.handleGetBondTypes,
	GetGOVConstitution:                RpcServer.handleGetGOVConstitution,
	GetGOVParams:                      RpcServer.handleGetGOVParams,
	CreateAndSendTxWithBuyBackRequest: RpcServer.handleCreateAndSendTxWithBuyBackRequest,
	CreateAndSendTxWithBuySellRequest: RpcServer.handleCreateAndSendTxWithBuySellRequest,
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
	ListUnspentOutputCoins:     RpcServer.handleListUnspentOutputCoins,
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

//handleListUnspentTx - use private key to get all tx which contains output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
//params:
//Parameter #1—the minimum number of confirmations an output must have
//Parameter #2—the maximum number of confirmations an output may have
//Parameter #3—the list readonly which be used to view utxo
//
func (self RpcServer) handleListUnspentOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	result := jsonresult.ListUnspentResult{
		ListUnspentResultItems: make(map[string][]jsonresult.ListUnspentResultItem),
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
		key, err := wallet.Base58CheckDeserialize(priKeyStr)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		outCoins, err := self.config.BlockChain.GetListOutputCoinsByKeyset(&key.KeySet, 14)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		listTxs := make([]jsonresult.ListUnspentResultItem, 0)
		item := jsonresult.ListUnspentResultItem{
			OutCoins: make([]jsonresult.OutCoin, 0),
		}
		for _, outCoin := range outCoins {
			item.OutCoins = append(item.OutCoins, jsonresult.OutCoin{
				SerialNumber:   base58.Base58Check{}.Encode(outCoin.CoinDetails.SerialNumber.Compress(), byte(0x00)),
				PublicKey:      base58.Base58Check{}.Encode(outCoin.CoinDetails.PublicKey.Compress(), byte(0x00)),
				Value:          outCoin.CoinDetails.Value,
				Info:           base58.Base58Check{}.Encode(outCoin.CoinDetails.Info[:], byte(0x00)),
				CoinCommitment: base58.Base58Check{}.Encode(outCoin.CoinDetails.CoinCommitment.Compress(), byte(0x00)),
				Randomness:     *outCoin.CoinDetails.Randomness,
				SNDerivator:    *outCoin.CoinDetails.SNDerivator,
			})
			listTxs = append(listTxs, item)

			if result.ListUnspentResultItems[priKeyStr] == nil {
				result.ListUnspentResultItems[priKeyStr] = []jsonresult.ListUnspentResultItem{}
			}
			if result.ListUnspentResultItems[priKeyStr] == nil {
				result.ListUnspentResultItems[priKeyStr] = []jsonresult.ListUnspentResultItem{}
			}
			result.ListUnspentResultItems[priKeyStr] = listTxs
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
	// return self.config.IsGenerateNode, nil
	return false, nil
}

/*
handleGetMiningInfo - RPC returns various mining-related info
*/
func (self RpcServer) handleGetMiningInfo(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// TODO update code to new consensus
	// if !self.config.IsGenerateNode {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Not mining"))
	// }
	// chainId := byte(int(params.(float64)))
	// result := jsonresult.GetMiningInfoResult{}
	// result.Blocks = uint64(self.config.BlockChain.BestState[chainId].BestBlock.Header.Height + 1)
	// result.PoolSize = self.config.TxMemPool.Count()
	// result.Chain = self.config.ChainParams.Name
	// result.CurrentBlockTx = len(self.config.BlockChain.BestState[chainId].BestBlock.Transactions)
	return jsonresult.GetMiningInfoResult{}, nil
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
	numBlock := uint64(arrayParams[0].(float64))
	result := jsonresult.EstimateFeeResult{
		FeeRate: make(map[string]uint64),
	}
	for chainID, feeEstimator := range self.config.FeeEstimator {
		var feeRate uint64
		var err error
		temp, err := feeEstimator.EstimateFee(numBlock)
		if err != nil {
			feeRate = uint64(temp)
		}
		if feeRate == 0 {
			feeRate = self.config.BlockChain.GetFeePerKbTx()
		}
		result.FeeRate[strconv.Itoa(int(chainID))] = feeRate
		if err != nil {
			return -1, NewRPCError(ErrUnexpected, err)
		}
	}
	return result, nil
}

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
