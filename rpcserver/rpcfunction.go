package rpcserver

import (
	"fmt"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/pkg/errors"
	"log"
	"net"
	"os"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
	"github.com/constant-money/constant-chain/wallet"
)

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, *RPCError)

// Commands valid for normal user
var RpcHandler = map[string]commandHandler{
	// node
	GetNetworkInfo:           RpcServer.handleGetNetWorkInfo,
	GetConnectionCount:       RpcServer.handleGetConnectionCount,
	GetAllPeers:              RpcServer.handleGetAllPeers,
	GetRawMempool:            RpcServer.handleGetRawMempool,
	GetMempoolEntry:          RpcServer.handleMempoolEntry,
	EstimateFee:              RpcServer.handleEstimateFee,
	EstimateFeeWithEstimator: RpcServer.handleEstimateFeeWithEstimator,
	GetGenerate:              RpcServer.handleGetGenerate,
	GetMiningInfo:            RpcServer.handleGetMiningInfo,

	// block
	GetBestBlock:        RpcServer.handleGetBestBlock,
	GetBestBlockHash:    RpcServer.handleGetBestBlockHash,
	RetrieveBlock:       RpcServer.handleRetrieveBlock,
	RetrieveBeaconBlock: RpcServer.handleRetrieveBeaconBlock,
	GetBlocks:           RpcServer.handleGetBlocks,
	GetBlockChainInfo:   RpcServer.handleGetBlockChainInfo,
	GetBlockCount:       RpcServer.handleGetBlockCount,
	GetBlockHash:        RpcServer.handleGetBlockHash,
	CheckHashValue:      RpcServer.handleCheckHashValue, // get data in blockchain from hash value
	GetBlockHeader:      RpcServer.handleGetBlockHeader, // Current committee, next block committee and candidate is included in block header

	// transaction
	ListOutputCoins:                 RpcServer.handleListOutputCoins,
	CreateRawTransaction:            RpcServer.handleCreateRawTransaction,
	SendRawTransaction:              RpcServer.handleSendRawTransaction,
	CreateAndSendTransaction:        RpcServer.handleCreateAndSendTx,
	GetMempoolInfo:                  RpcServer.handleGetMempoolInfo,
	GetTransactionByHash:            RpcServer.handleGetTransactionByHash,
	CreateAndSendStakingTransaction: RpcServer.handleCreateAndSendStakingTx,
	RandomCommitments:               RpcServer.handleRandomCommitments,
	HasSerialNumbers:                RpcServer.handleHasSerialNumbers,
	HasSnDerivators:                 RpcServer.handleHasSnDerivators,

	// Beststate
	GetCandidateList:              RpcServer.handleGetCandidateList,
	GetCommitteeList:              RpcServer.handleGetCommitteeList,
	GetBlockProducerList:          RpcServer.handleGetBlockProducerList,
	GetShardBestState:             RpcServer.handleGetShardBestState,
	GetBeaconBestState:            RpcServer.handleGetBeaconBestState,
	GetBeaconPoolState:            RpcServer.handleGetBeaconPoolState,
	GetShardPoolState:             RpcServer.handleGetShardPoolState,
	GetShardPoolLatestValidHeight: RpcServer.handleGetShardPoolLatestValidHeight,
	GetShardToBeaconPoolState:     RpcServer.handleGetShardToBeaconPoolState,
	GetCrossShardPoolState:        RpcServer.handleGetCrossShardPoolState,
	CanPubkeyStake:                RpcServer.handleCanPubkeyStake,

	// custom token
	CreateRawCustomTokenTransaction:     RpcServer.handleCreateRawCustomTokenTransaction,
	SendRawCustomTokenTransaction:       RpcServer.handleSendRawCustomTokenTransaction,
	CreateAndSendCustomTokenTransaction: RpcServer.handleCreateAndSendCustomTokenTransaction,
	ListUnspentCustomToken:              RpcServer.handleListUnspentCustomToken,
	ListCustomToken:                     RpcServer.handleListCustomToken,
	CustomToken:                         RpcServer.handleCustomTokenDetail,
	GetListCustomTokenBalance:           RpcServer.handleGetListCustomTokenBalance,

	// custom token which support privacy
	CreateRawPrivacyCustomTokenTransaction:     RpcServer.handleCreateRawPrivacyCustomTokenTransaction,
	SendRawPrivacyCustomTokenTransaction:       RpcServer.handleSendRawPrivacyCustomTokenTransaction,
	CreateAndSendPrivacyCustomTokenTransaction: RpcServer.handleCreateAndSendPrivacyCustomTokenTransaction,
	ListPrivacyCustomToken:                     RpcServer.handleListPrivacyCustomToken,
	PrivacyCustomToken:                         RpcServer.handlePrivacyCustomTokenDetail,
	GetListPrivacyCustomTokenBalance:           RpcServer.handleGetListPrivacyCustomTokenBalance,

	// Loan tx
	GetLoanParams:             RpcServer.handleGetLoanParams,
	CreateAndSendLoanRequest:  RpcServer.handleCreateAndSendLoanRequest,
	CreateAndSendLoanResponse: RpcServer.handleCreateAndSendLoanResponse,
	CreateAndSendLoanWithdraw: RpcServer.handleCreateAndSendLoanWithdraw,
	CreateAndSendLoanPayment:  RpcServer.handleCreateAndSendLoanPayment,
	GetLoanResponseApproved:   RpcServer.handleGetLoanResponseApproved,
	GetLoanResponseRejected:   RpcServer.handleGetLoanResponseRejected,
	GetLoanPaymentInfo:        RpcServer.handleGetLoanPaymentInfo,
	GetBankFund:               RpcServer.handleGetBankFund,
	GetLoanRequestTxStatus:    RpcServer.handleGetLoanRequestTxStatus,

	// Crowdsale
	GetListOngoingCrowdsale:               RpcServer.handleGetListOngoingCrowdsale,
	CreateCrowdsaleRequestToken:           RpcServer.handleCreateCrowdsaleRequestToken,
	SendCrowdsaleRequestToken:             RpcServer.handleSendCrowdsaleRequestToken,
	CreateAndSendCrowdsaleRequestToken:    RpcServer.handleCreateAndSendCrowdsaleRequestToken,
	CreateCrowdsaleRequestConstant:        RpcServer.handleCreateCrowdsaleRequestConstant,
	SendCrowdsaleRequestConstant:          RpcServer.handleSendCrowdsaleRequestConstant,
	CreateAndSendCrowdsaleRequestConstant: RpcServer.handleCreateAndSendCrowdsaleRequestConstant,
	GetListDCBProposalBuyingAssets:        RpcServer.handleGetListDCBProposalBuyingAssets,
	GetListDCBProposalSellingAssets:       RpcServer.handleGetListDCBProposalSellingAssets,

	// Reserve
	CreateIssuingRequest:            RpcServer.handleCreateIssuingRequest,
	SendIssuingRequest:              RpcServer.handleSendIssuingRequest,
	CreateAndSendIssuingRequest:     RpcServer.handleCreateAndSendIssuingRequest,
	CreateAndSendContractingRequest: RpcServer.handleCreateAndSendContractingRequest,
	GetIssuingStatus:                RpcServer.handleGetIssuingStatus,
	GetContractingStatus:            RpcServer.handleGetContractingStatus,
	ConvertETHToDCBTokenAmount:      RpcServer.handleConvertETHToDCBTokenAmount,
	ConvertCSTToETHAmount:           RpcServer.handleConvertCSTToETHAmount,

	// multisig
	CreateSignatureOnCustomTokenTx:       RpcServer.handleCreateSignatureOnCustomTokenTx,
	GetListDCBBoard:                      RpcServer.handleGetListDCBBoard,
	GetListGOVBoard:                      RpcServer.handleGetListGOVBoard,
	AppendListDCBBoard:                   RpcServer.handleAppendListDCBBoard,
	AppendListGOVBoard:                   RpcServer.handleAppendListGOVBoard,
	CreateAndSendTxWithMultiSigsReg:      RpcServer.handleCreateAndSendTxWithMultiSigsReg,
	CreateAndSendTxWithMultiSigsSpending: RpcServer.handleCreateAndSendTxWithMultiSigsSpending,

	// vote board
	CreateAndSendVoteDCBBoardTransaction: RpcServer.handleCreateAndSendVoteDCBBoardTransaction,
	CreateRawVoteDCBBoardTx:              RpcServer.handleCreateRawVoteDCBBoardTransaction,
	CreateAndSendVoteGOVBoardTransaction: RpcServer.handleCreateAndSendVoteGOVBoardTransaction,
	CreateRawVoteGOVBoardTx:              RpcServer.handleCreateRawVoteGOVBoardTransaction,

	// vote proposal
	GetEncryptionFlag:                RpcServer.handleGetEncryptionFlag,
	SetEncryptionFlag:                RpcServer.handleSetEncryptionFlag,
	GetEncryptionLastBlockHeightFlag: RpcServer.handleGetEncryptionLastBlockHeightFlag,
	CreateAndSendVoteProposal:        RpcServer.handleCreateAndSendVoteProposalTransaction,

	// Submit Proposal:
	CreateAndSendSubmitDCBProposalTx: RpcServer.handleCreateAndSendSubmitDCBProposalTransaction,
	CreateRawSubmitDCBProposalTx:     RpcServer.handleCreateRawSubmitDCBProposalTransaction,
	CreateAndSendSubmitGOVProposalTx: RpcServer.handleCreateAndSendSubmitGOVProposalTransaction,
	CreateRawSubmitGOVProposalTx:     RpcServer.handleCreateRawSubmitGOVProposalTransaction,

	// dcb
	GetDCBParams:       RpcServer.handleGetDCBParams,
	GetDCBConstitution: RpcServer.handleGetDCBConstitution,
	GetDCBBoardIndex:   RpcServer.handleGetDCBBoardIndex,
	GetGOVBoardIndex:   RpcServer.handleGetGOVBoardIndex,
	// CreateAndSendTxWithIssuingRequest:     RpcServer.handleCreateAndSendTxWithIssuingRequest,
	// CreateAndSendTxWithContractingRequest: RpcServer.handleCreateAndSendTxWithContractingRequest,

	// gov
	GetBondTypes:                           RpcServer.handleGetBondTypes,
	GetCurrentSellingBondTypes:             RpcServer.handleGetCurrentSellingBondTypes,
	GetCurrentStabilityInfo:                RpcServer.handleGetCurrentStabilityInfo,
	GetGOVConstitution:                     RpcServer.handleGetGOVConstitution,
	GetGOVParams:                           RpcServer.handleGetGOVParams,
	CreateAndSendTxWithBuyBackRequest:      RpcServer.handleCreateAndSendTxWithBuyBackRequest,
	CreateAndSendTxWithBuySellRequest:      RpcServer.handleCreateAndSendTxWithBuySellRequest,
	CreateAndSendTxWithOracleFeed:          RpcServer.handleCreateAndSendTxWithOracleFeed,
	CreateAndSendTxWithUpdatingOracleBoard: RpcServer.handleCreateAndSendTxWithUpdatingOracleBoard,
	CreateAndSendTxWithSenderAddress:       RpcServer.handleCreateAndSendTxWithSenderAddress,
	CreateAndSendTxWithBuyGOVTokensRequest: RpcServer.handleCreateAndSendTxWithBuyGOVTokensRequest,
	GetCurrentSellingGOVTokens:             RpcServer.handleGetCurrentSellingGOVTokens,

	// cmb
	CreateAndSendTxWithCMBInitRequest:     RpcServer.handleCreateAndSendTxWithCMBInitRequest,
	CreateAndSendTxWithCMBInitResponse:    RpcServer.handleCreateAndSendTxWithCMBInitResponse,
	CreateAndSendTxWithCMBDepositContract: RpcServer.handleCreateAndSendTxWithCMBDepositContract,
	CreateAndSendTxWithCMBDepositSend:     RpcServer.handleCreateAndSendTxWithCMBDepositSend,
	CreateAndSendTxWithCMBWithdrawRequest: RpcServer.handleCreateAndSendTxWithCMBWithdrawRequest,

	// wallet
	GetPublicKeyFromPaymentAddress: RpcServer.handleGetPublicKeyFromPaymentAddress,
	DefragmentAccount:              RpcServer.handleDefragmentAccount,
}

// Commands that are available to a limited user
var RpcLimited = map[string]commandHandler{
	// local WALLET
	ListAccounts:                       RpcServer.handleListAccounts,
	GetAccount:                         RpcServer.handleGetAccount,
	GetAddressesByAccount:              RpcServer.handleGetAddressesByAccount,
	GetAccountAddress:                  RpcServer.handleGetAccountAddress,
	DumpPrivkey:                        RpcServer.handleDumpPrivkey,
	ImportAccount:                      RpcServer.handleImportAccount,
	RemoveAccount:                      RpcServer.handleRemoveAccount,
	ListUnspentOutputCoins:             RpcServer.handleListUnspentOutputCoins,
	GetBalance:                         RpcServer.handleGetBalance,
	GetBalanceByPrivatekey:             RpcServer.handleGetBalanceByPrivatekey,
	GetBalanceByPaymentAddress:         RpcServer.handleGetBalanceByPaymentAddress,
	GetReceivedByAccount:               RpcServer.handleGetReceivedByAccount,
	SetTxFee:                           RpcServer.handleSetTxFee,
	GetRecentTransactionsByBlockNumber: RpcServer.handleGetRecentTransactionsByBlockNumber,
}

func (rpcServer RpcServer) handleGetNetWorkInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetNetworkInfoResult{}

	result.Commit = os.Getenv("commit")
	result.Version = RpcServerVersion
	result.SubVersion = ""
	result.ProtocolVersion = rpcServer.config.ProtocolVersion
	result.NetworkActive = rpcServer.config.ConnMgr.ListeningPeer != nil
	result.LocalAddresses = []string{}
	listener := rpcServer.config.ConnMgr.ListeningPeer
	result.Connections = len(listener.PeerConns)
	result.LocalAddresses = append(result.LocalAddresses, listener.RawAddress)

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
	if rpcServer.config.Wallet != nil && rpcServer.config.Wallet.GetConfig() != nil {
		result.IncrementalFee = rpcServer.config.Wallet.GetConfig().IncrementalFee
	}
	result.Warnings = ""

	return result, nil
}

//handleListUnspentOutputCoins - use private key to get all tx which contains output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
//component:
//Parameter #1—the minimum number of confirmations an output must have
//Parameter #2—the maximum number of confirmations an output may have
//Parameter #3—the list priv-key which be used to view utxo
//
func (rpcServer RpcServer) handleListUnspentOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	result := jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}

	// get component
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
		keyWallet, err := wallet.Base58CheckDeserialize(priKeyStr)
		if err != nil {
			log.Println("Check Deserialize err", err)
			continue
		}
		if keyWallet.KeySet.PrivateKey == nil {
			log.Println("Private key empty")
			continue
		}

		keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)
		shardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		tokenID := &common.Hash{}
		tokenID.SetBytes(common.ConstantID[:])
		outCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&keyWallet.KeySet, shardID, tokenID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		item := make([]jsonresult.OutCoin, 0)
		for _, outCoin := range outCoins {
			item = append(item, jsonresult.OutCoin{
				SerialNumber:   base58.Base58Check{}.Encode(outCoin.CoinDetails.SerialNumber.Compress(), common.ZeroByte),
				PublicKey:      base58.Base58Check{}.Encode(outCoin.CoinDetails.PublicKey.Compress(), common.ZeroByte),
				Value:          outCoin.CoinDetails.Value,
				Info:           base58.Base58Check{}.Encode(outCoin.CoinDetails.Info[:], common.ZeroByte),
				CoinCommitment: base58.Base58Check{}.Encode(outCoin.CoinDetails.CoinCommitment.Compress(), common.ZeroByte),
				Randomness:     base58.Base58Check{}.Encode(outCoin.CoinDetails.Randomness.Bytes(), common.ZeroByte),
				SNDerivator:    base58.Base58Check{}.Encode(outCoin.CoinDetails.SNDerivator.Bytes(), common.ZeroByte),
			})
		}
		result.Outputs[priKeyStr] = item
	}
	return result, nil
}

func (rpcServer RpcServer) handleCheckHashValue(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var (
		isTransaction bool
		isBlock       bool
	)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Expected array component"))
	}
	hashParams, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Expected hash string value"))
	}
	// param #1: transaction Hash
	// Logger.log.Infof("Check hash value  input Param %+v", arrayParams[0].(string))
	log.Printf("Check hash value  input Param %+v", hashParams)
	hash, _ := common.Hash{}.NewHashFromStr(hashParams)

	// Check block
	// _, err := rpcServer.config.BlockChain.GetBlockByHash(hash)
	_, err, _ := rpcServer.config.BlockChain.GetShardBlockByHash(hash)
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
	_, _, _, _, err1 := rpcServer.config.BlockChain.GetTransactionByHash(hash)
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
func (rpcServer RpcServer) handleGetConnectionCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	if rpcServer.config.ConnMgr == nil || rpcServer.config.ConnMgr.ListeningPeer == nil {
		return 0, nil
	}
	result := 0
	listeningPeer := rpcServer.config.ConnMgr.ListeningPeer
	result += len(listeningPeer.PeerConns)
	return result, nil
}

/*
handleGetGenerate - RPC returns true if the node is set to generate blocks using its CPU
*/
func (rpcServer RpcServer) handleGetGenerate(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// return rpcServer.config.IsGenerateNode, nil
	return false, nil
}

/*
handleGetMiningInfo - RPC returns various mining-related info
*/
func (rpcServer RpcServer) handleGetMiningInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// TODO update code to new consensus
	// if !rpcServer.config.IsGenerateNode {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Not mining"))
	// }
	// shardID := byte(int(component.(float64)))
	// result := jsonresult.GetMiningInfoResult{}
	// result.Blocks = uint64(rpcServer.config.BlockChain.BestState[shardID].BestBlock.Header.Height + 1)
	// result.PoolSize = rpcServer.config.TxMemPool.Count()
	// result.Chain = rpcServer.config.ChainParams.Name
	// result.CurrentBlockTx = len(rpcServer.config.BlockChain.BestState[shardID].BestBlock.Transactions)
	return jsonresult.GetMiningInfoResult{}, nil
}

/*
handleGetRawMempool - RPC returns all transaction ids in memory pool as a json array of string transaction ids
Hint: use getmempoolentry to fetch a specific transaction from the mempool.
*/
func (rpcServer RpcServer) handleGetRawMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetRawMempoolResult{
		TxHashes: rpcServer.config.TxMemPool.ListTxs(),
	}
	return result, nil
}

/*
handleMempoolEntry - RPC fetch a specific transaction from the mempool
*/
func (rpcServer RpcServer) handleMempoolEntry(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// Param #1: hash string of tx(tx id)
	if params == nil {
		params = ""
	}
	txID, err := common.Hash{}.NewHashFromStr(params.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	result := jsonresult.GetMempoolEntryResult{}
	result.Tx, err = rpcServer.config.TxMemPool.GetTx(txID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return result, nil
}

/*
handleEstimateFee - RPC estimates the transaction fee per kilobyte that needs to be paid for a transaction to be included within a certain number of blocks.
*/
func (rpcServer RpcServer) handleEstimateFee(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	/******* START Fetch all component to ******/
	// all component
	arrayParams := common.InterfaceSlice(params)
	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	// param #3: estimation fee coin per kb
	defaultFeeCoinPerKb := int64(arrayParams[2].(float64))
	// param #4: hasPrivacy flag for constant
	hasPrivacy := int(arrayParams[3].(float64)) > 0

	senderKeySet, err := rpcServer.GetKeySetFromPrivateKeyParams(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderPrivateKey, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	fmt.Printf("Done param #1: keyset: %+v\n", senderKeySet)

	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	outCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(senderKeySet, shardIDSender, constantTokenID)
	if err != nil {
		return nil, NewRPCError(ErrGetOutputCoin, err)
	}
	// remove out coin in mem pool
	outCoins, err = rpcServer.filterMemPoolOutCoinsToSpent(outCoins)
	if err != nil {
		return nil, NewRPCError(ErrGetOutputCoin, err)
	}

	govFeePerKbTx := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams.FeePerKbTx
	estimateFeeCoinPerKb := uint64(0)
	estimateTxSizeInKb := uint64(0)
	if len(outCoins) > 0 {
		// param #2: list receiver
		receiversPaymentAddressStrParam := make(map[string]interface{})
		if arrayParams[1] != nil {
			receiversPaymentAddressStrParam = arrayParams[1].(map[string]interface{})
		}
		paymentInfos := make([]*privacy.PaymentInfo, 0)
		for paymentAddressStr, amount := range receiversPaymentAddressStrParam {
			keyWalletReceiver, err := wallet.Base58CheckDeserialize(paymentAddressStr)
			if err != nil {
				return nil, NewRPCError(ErrInvalidReceiverPaymentAddress, err)
			}
			paymentInfo := &privacy.PaymentInfo{
				Amount:         uint64(amount.(float64)),
				PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
			}
			paymentInfos = append(paymentInfos, paymentInfo)
		}

		// Check custom token param
		var customTokenParams *transaction.CustomTokenParamTx
		var customPrivacyTokenParam *transaction.CustomTokenPrivacyParamTx
		if len(arrayParams) > 4 {
			// param #5: token params
			tokenParamsRaw := arrayParams[4].(map[string]interface{})
			privacy := tokenParamsRaw["Privacy"].(bool)
			if !privacy {
				// Check normal custom token param
				customTokenParams, _, err = rpcServer.buildCustomTokenParam(tokenParamsRaw, senderKeySet)
				if err.(*RPCError) != nil {
					return nil, err.(*RPCError)
				}
			} else {
				// Check privacy custom token param
				customPrivacyTokenParam, _, err = rpcServer.buildPrivacyCustomTokenParam(tokenParamsRaw, senderKeySet, shardIDSender)
				if err.(*RPCError) != nil {
					return nil, err.(*RPCError)
				}
			}
		}

		// check real fee(nano constant) per tx
		_, estimateFeeCoinPerKb, estimateTxSizeInKb = rpcServer.estimateFee(defaultFeeCoinPerKb, outCoins, paymentInfos, shardIDSender, 8, hasPrivacy, nil, customTokenParams, customPrivacyTokenParam)
	}
	result := jsonresult.EstimateFeeResult{
		EstimateFeeCoinPerKb: estimateFeeCoinPerKb,
		EstimateTxSizeInKb:   estimateTxSizeInKb,
		GOVFeePerKbTx:        govFeePerKbTx,
	}
	return result, nil
}

// handleEstimateFeeWithEstimator -- get fee from estomator
func (rpcServer RpcServer) handleEstimateFeeWithEstimator(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: estimation fee coin per kb from client
	defaultFeeCoinPerKb := int64(arrayParams[0].(float64))

	// param #2: payment address
	senderKeyParam := arrayParams[1]
	senderKeySet, err := rpcServer.GetKeySetFromKeyParams(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderPrivateKey, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)

	// param #2: numblocl
	estimateFeeCoinPerKb := rpcServer.estimateFeeWithEstimator(defaultFeeCoinPerKb, shardIDSender, 8)
	govFeePerKbTx := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams.FeePerKbTx

	result := jsonresult.EstimateFeeResult{
		EstimateFeeCoinPerKb: estimateFeeCoinPerKb,
		GOVFeePerKbTx:        govFeePerKbTx,
	}
	return result, nil
}

// handleGetActiveShards - return active shard num
func (rpcServer RpcServer) handleGetActiveShards(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	if rpcServer.config.BlockChain.IsReady(false, 0) {
		activeShards := rpcServer.config.BlockChain.BestState.Beacon.ActiveShards
		return activeShards, nil
	}
	return -1, nil
}
