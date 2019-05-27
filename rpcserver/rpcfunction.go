package rpcserver

import (
	"log"
	"net"
	"os"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, *RPCError)

// Commands valid for normal user
var RpcHandler = map[string]commandHandler{

	StartProfiling: RpcServer.handleStartProfiling,
	StopProfiling:  RpcServer.handleStopProfiling,
	// node
	GetNetworkInfo:           RpcServer.handleGetNetWorkInfo,
	GetConnectionCount:       RpcServer.handleGetConnectionCount,
	GetAllPeers:              RpcServer.handleGetAllPeers,
	EstimateFee:              RpcServer.handleEstimateFee,
	EstimateFeeWithEstimator: RpcServer.handleEstimateFeeWithEstimator,
	GetActiveShards:          RpcServer.handleGetActiveShards,
	GetMaxShardsNumber:       RpcServer.handleGetMaxShardsNumber,

	//pool
	GetMiningInfo:               RpcServer.handleGetMiningInfo,
	GetRawMempool:               RpcServer.handleGetRawMempool,
	GetNumberOfTxsInMempool:     RpcServer.handleGetNumberOfTxsInMempool,
	GetMempoolEntry:             RpcServer.handleMempoolEntry,
	GetShardToBeaconPoolStateV2: RpcServer.handleGetShardToBeaconPoolStateV2,
	GetCrossShardPoolStateV2:    RpcServer.handleGetCrossShardPoolStateV2,
	GetShardPoolStateV2:         RpcServer.handleGetShardPoolStateV2,
	GetBeaconPoolStateV2:        RpcServer.handleGetBeaconPoolStateV2,
	GetShardToBeaconPoolState:   RpcServer.handleGetShardToBeaconPoolState,
	GetCrossShardPoolState:      RpcServer.handleGetCrossShardPoolState,
	GetNextCrossShard:           RpcServer.handleGetNextCrossShard,
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
	GetCrossShardBlock:  RpcServer.handleGetCrossShardBlock,

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

	//======Testing and Benchmark======
	GetAndSendTxsFromFile: RpcServer.handleGetAndSendTxsFromFile,
	//=================================

	//pool

	// Beststate
	GetCandidateList:              RpcServer.handleGetCandidateList,
	GetCommitteeList:              RpcServer.handleGetCommitteeList,
	GetBlockProducerList:          RpcServer.handleGetBlockProducerList,
	GetShardBestState:             RpcServer.handleGetShardBestState,
	GetBeaconBestState:            RpcServer.handleGetBeaconBestState,
	GetBeaconPoolState:            RpcServer.handleGetBeaconPoolState,
	GetShardPoolState:             RpcServer.handleGetShardPoolState,
	GetShardPoolLatestValidHeight: RpcServer.handleGetShardPoolLatestValidHeight,
	CanPubkeyStake:                RpcServer.handleCanPubkeyStake,
	GetTotalTransaction:           RpcServer.handleGetTotalTransaction,

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

	// wallet
	GetPublicKeyFromPaymentAddress: RpcServer.handleGetPublicKeyFromPaymentAddress,
	DefragmentAccount:              RpcServer.handleDefragmentAccount,

	GetStackingAmount: RpcServer.handleGetStakingAmount,

	HashToIdenticon: RpcServer.handleHashToIdenticon,
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
	Logger.log.Infof("handleListUnspentOutputCoins params: %+v", params)
	result := jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}

	// get component
	paramsArray := common.InterfaceSlice(params)
	var min int
	var max int
	if len(paramsArray) > 0 && paramsArray[0] != nil {
		min = int(paramsArray[0].(float64))
	}
	if len(paramsArray) > 1 && paramsArray[1] != nil {
		max = int(paramsArray[1].(float64))
	}
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
			if outCoin.CoinDetails.Value == 0 {
				continue
			}
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
	Logger.log.Infof("handleListUnspentOutputCoins result: %+v", result)
	return result, nil
}

func (rpcServer RpcServer) handleCheckHashValue(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCheckHashValue params: %+v", params)
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
	Logger.log.Infof("Check hash value  input Param %+v", arrayParams[0].(string))
	log.Printf("Check hash value  input Param %+v", hashParams)
	hash, _ := common.Hash{}.NewHashFromStr(hashParams)

	// Check block
	_, _, err := rpcServer.config.BlockChain.GetShardBlockByHash(hash)
	if err != nil {
		isBlock = false
	} else {
		isBlock = true
		result := jsonresult.HashValueDetail{
			IsBlock:       isBlock,
			IsTransaction: false,
		}
		Logger.log.Infof("handleCheckHashValue result: %+v", result)
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
		Logger.log.Infof("handleCheckHashValue result: %+v", result)
		return result, nil
	}
	result := jsonresult.HashValueDetail{
		IsBlock:       isBlock,
		IsTransaction: isTransaction,
	}
	Logger.log.Infof("handleCheckHashValue result: %+v", result)
	return result, nil
}

/*
handleGetConnectionCount - RPC returns the number of connections to other nodes.
*/
func (rpcServer RpcServer) handleGetConnectionCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetConnectionCount params: %+v", params)
	if rpcServer.config.ConnMgr == nil || rpcServer.config.ConnMgr.ListeningPeer == nil {
		return 0, nil
	}
	result := 0
	listeningPeer := rpcServer.config.ConnMgr.ListeningPeer
	result += len(listeningPeer.PeerConns)
	Logger.log.Infof("handleGetConnectionCount result: %+v", result)
	return result, nil
}

/*
handleGetMiningInfo - RPC returns various mining-related info
*/
func (rpcServer RpcServer) handleGetMiningInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetMiningInfo params: %+v", params)
	if !rpcServer.config.IsMiningNode || rpcServer.config.MiningPubKeyB58 == "" {
		return jsonresult.GetMiningInfoResult{
			IsCommittee: false,
		}, nil
	}

	result := jsonresult.GetMiningInfoResult{}
	result.IsCommittee = true
	result.PoolSize = rpcServer.config.TxMemPool.Count()
	result.Chain = rpcServer.config.ChainParams.Name

	result.BeaconHeight = rpcServer.config.BlockChain.BestState.Beacon.BeaconHeight

	role, shardID := rpcServer.config.BlockChain.BestState.Beacon.GetPubkeyRole(rpcServer.config.MiningPubKeyB58, 0)
	result.Role = role
	if role == common.SHARD_ROLE {
		result.ShardHeight = rpcServer.config.BlockChain.BestState.Shard[shardID].ShardHeight
		result.CurrentShardBlockTx = len(rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock.Body.Transactions)
		result.ShardID = int(shardID)
	} else if role == common.VALIDATOR_ROLE || role == common.PROPOSER_ROLE || role == common.PENDING_ROLE {
		result.ShardID = -1
	}
	Logger.log.Infof("handleGetMiningInfo result: %+v", result)
	return result, nil
}

/*
handleGetRawMempool - RPC returns all transaction ids in memory pool as a json array of string transaction ids
Hint: use getmempoolentry to fetch a specific transaction from the mempool.
*/
func (rpcServer RpcServer) handleGetRawMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetRawMempool params: %+v", params)
	result := jsonresult.GetRawMempoolResult{
		TxHashes: rpcServer.config.TxMemPool.ListTxs(),
	}
	Logger.log.Infof("handleGetRawMempool result: %+v", result)
	return result, nil
}

func (rpcServer RpcServer) handleGetNumberOfTxsInMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetNumberOfTxsInMempool params: %+v", params)
	result := len(rpcServer.config.TxMemPool.ListTxs())
	Logger.log.Infof("handleGetNumberOfTxsInMempool result: %+v", result)
	return result, nil
}

/*
handleMempoolEntry - RPC fetch a specific transaction from the mempool
*/
func (rpcServer RpcServer) handleMempoolEntry(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleMempoolEntry params: %+v", params)
	// Param #1: hash string of tx(tx id)
	if params == nil {
		params = ""
	}
	txID, err := common.Hash{}.NewHashFromStr(params.(string))
	if err != nil {
		Logger.log.Infof("handleMempoolEntry result: nil %+v", err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	result := jsonresult.GetMempoolEntryResult{}
	result.Tx, err = rpcServer.config.TxMemPool.GetTx(txID)
	if err != nil {
		Logger.log.Infof("handleMempoolEntry result: nil %+v", err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	Logger.log.Infof("handleMempoolEntry result: %+v", result)
	return result, nil
}

/*
handleEstimateFee - RPC estimates the transaction fee per kilobyte that needs to be paid for a transaction to be included within a certain number of blocks.
*/
func (rpcServer RpcServer) handleEstimateFee(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleEstimateFee params: %+v", params)
	/******* START Fetch all component to ******/
	// all component
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Not enough params"))
	}
	// param #1: private key of sender
	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Sender private key is invalid"))
	}
	// param #3: estimation fee coin per kb
	defaultFeeCoinPerKbtemp, ok := arrayParams[2].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Default FeeCoinPerKbtemp is invalid"))
	}
	defaultFeeCoinPerKb := int64(defaultFeeCoinPerKbtemp)
	// param #4: hasPrivacy flag for constant
	hashPrivacyTemp, ok := arrayParams[3].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("hasPrivacy is invalid"))
	}
	hasPrivacy := int(hashPrivacyTemp) > 0

	senderKeySet, err := rpcServer.GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		return nil, NewRPCError(ErrInvalidSenderPrivateKey, err)
	}
	lastByte := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	//fmt.Printf("Done param #1: keyset: %+v\n", senderKeySet)

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
				customPrivacyTokenParam, _, _, err = rpcServer.buildPrivacyCustomTokenParam(tokenParamsRaw, senderKeySet, shardIDSender)
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
	}
	Logger.log.Infof("handleEstimateFee result: %+v", result)
	return result, nil
}

// handleEstimateFeeWithEstimator -- get fee from estomator
func (rpcServer RpcServer) handleEstimateFeeWithEstimator(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleEstimateFeeWithEstimator params: %+v", params)
	// all params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 2 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Not enough params"))
	}
	// param #1: estimation fee coin per kb from client
	defaultFeeCoinPerKbTemp, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("defaultFeeCoinPerKbTemp is invalid"))
	}
	defaultFeeCoinPerKb := int64(defaultFeeCoinPerKbTemp)

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

	result := jsonresult.EstimateFeeResult{
		EstimateFeeCoinPerKb: estimateFeeCoinPerKb,
	}
	Logger.log.Infof("handleEstimateFeeWithEstimator result: %+v", result)
	return result, nil
}

// handleGetActiveShards - return active shard num
func (rpcServer RpcServer) handleGetActiveShards(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetActiveShards params: %+v", params)
	activeShards := rpcServer.config.BlockChain.BestState.Beacon.ActiveShards
	Logger.log.Infof("handleGetActiveShards result: %+v", activeShards)
	return activeShards, nil
}

func (rpcServer RpcServer) handleGetMaxShardsNumber(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetMaxShardsNumber params: %+v", params)
	result := common.MAX_SHARD_NUMBER
	Logger.log.Infof("handleGetMaxShardsNumber result: %+v", result)
	return result, nil
}

func (rpcServer RpcServer) handleGetStakingAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetStakingAmount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) <= 0 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("ErrRPCInvalidParams"))
	}
	stackingType := int64(arrayParams[0].(float64))
	amount := uint64(0)
	if stackingType == 1 {
		amount = metadata.GetBeaconStakeAmount()
	}
	if stackingType == 0 {
		amount = metadata.GetShardStateAmount()
	}
	Logger.log.Infof("handleGetStakingAmount result: %+v", amount)
	return amount, nil
}

func (rpcServer RpcServer) handleHashToIdenticon(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	result := make([]string, 0)
	for _, hash := range arrayParams {
		temp, err := common.Hash{}.NewHashFromStr(hash.(string))
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("Hash string is invalid"))
		}
		result = append(result, common.Render(temp.GetBytes()))
	}
	return result, nil
}
