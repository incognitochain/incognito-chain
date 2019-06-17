package rpcserver


type httpHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, *RPCError)
type wsHandler func(RpcServer, interface{}, <-chan struct{}) (chan interface{}, *RPCError)
// Commands valid for normal user
var HttpHandler = map[string]httpHandler{
	
	startProfiling: RpcServer.handleStartProfiling,
	stopProfiling:  RpcServer.handleStopProfiling,
	// node
	getNetworkInfo:           RpcServer.handleGetNetWorkInfo,
	getConnectionCount:       RpcServer.handleGetConnectionCount,
	getAllPeers:              RpcServer.handleGetAllPeers,
	estimateFee:              RpcServer.handleEstimateFee,
	estimateFeeWithEstimator: RpcServer.handleEstimateFeeWithEstimator,
	getActiveShards:          RpcServer.handleGetActiveShards,
	getMaxShardsNumber:       RpcServer.handleGetMaxShardsNumber,
	
	//pool
	getMiningInfo:               RpcServer.handleGetMiningInfo,
	getRawMempool:               RpcServer.handleGetRawMempool,
	getNumberOfTxsInMempool:     RpcServer.handleGetNumberOfTxsInMempool,
	getMempoolEntry:             RpcServer.handleMempoolEntry,
	getShardToBeaconPoolStateV2: RpcServer.handleGetShardToBeaconPoolStateV2,
	getCrossShardPoolStateV2:    RpcServer.handleGetCrossShardPoolStateV2,
	getShardPoolStateV2:         RpcServer.handleGetShardPoolStateV2,
	getBeaconPoolStateV2:        RpcServer.handleGetBeaconPoolStateV2,
	getShardToBeaconPoolState:   RpcServer.handleGetShardToBeaconPoolState,
	getCrossShardPoolState:      RpcServer.handleGetCrossShardPoolState,
	getNextCrossShard:           RpcServer.handleGetNextCrossShard,
	// block
	getBestBlock:        RpcServer.handleGetBestBlock,
	getBestBlockHash:    RpcServer.handleGetBestBlockHash,
	retrieveBlock:       RpcServer.handleRetrieveBlock,
	retrieveBeaconBlock: RpcServer.handleRetrieveBeaconBlock,
	getBlocks:           RpcServer.handleGetBlocks,
	getBlockChainInfo:   RpcServer.handleGetBlockChainInfo,
	getBlockCount:       RpcServer.handleGetBlockCount,
	getBlockHash:        RpcServer.handleGetBlockHash,
	checkHashValue:      RpcServer.handleCheckHashValue, // get data in blockchain from hash value
	getBlockHeader:      RpcServer.handleGetBlockHeader, // Current committee, next block committee and candidate is included in block header
	getCrossShardBlock:  RpcServer.handleGetCrossShardBlock,
	
	// transaction
	listOutputCoins:                 RpcServer.handleListOutputCoins,
	createRawTransaction:            RpcServer.handleCreateRawTransaction,
	sendRawTransaction:              RpcServer.handleSendRawTransaction,
	createAndSendTransaction:        RpcServer.handleCreateAndSendTx,
	getMempoolInfo:                  RpcServer.handleGetMempoolInfo,
	getTransactionByHash:            RpcServer.handleGetTransactionByHash,
	createAndSendStakingTransaction: RpcServer.handleCreateAndSendStakingTx,
	randomCommitments:               RpcServer.handleRandomCommitments,
	hasSerialNumbers:                RpcServer.handleHasSerialNumbers,
	hasSnDerivators:                 RpcServer.handleHasSnDerivators,
	
	//======Testing and Benchmark======
	getAndSendTxsFromFile: RpcServer.handleGetAndSendTxsFromFile,
	getAndSendTxsFromFileV2: RpcServer.handleGetAndSendTxsFromFileV2,
	unlockMempool:         RpcServer.handleUnlockMempool,
	//=================================
	
	//pool
	
	// Beststate
	getCandidateList:              RpcServer.handleGetCandidateList,
	getCommitteeList:              RpcServer.handleGetCommitteeList,
	getBlockProducerList:          RpcServer.handleGetBlockProducerList,
	getShardBestState:             RpcServer.handleGetShardBestState,
	getBeaconBestState:            RpcServer.handleGetBeaconBestState,
	getBeaconPoolState:            RpcServer.handleGetBeaconPoolState,
	getShardPoolState:             RpcServer.handleGetShardPoolState,
	getShardPoolLatestValidHeight: RpcServer.handleGetShardPoolLatestValidHeight,
	canPubkeyStake:                RpcServer.handleCanPubkeyStake,
	getTotalTransaction:           RpcServer.handleGetTotalTransaction,
	
	// custom token
	createRawCustomTokenTransaction:     RpcServer.handleCreateRawCustomTokenTransaction,
	sendRawCustomTokenTransaction:       RpcServer.handleSendRawCustomTokenTransaction,
	createAndSendCustomTokenTransaction: RpcServer.handleCreateAndSendCustomTokenTransaction,
	listUnspentCustomToken:              RpcServer.handleListUnspentCustomToken,
	listCustomToken:                     RpcServer.handleListCustomToken,
	customTokenTxs:                      RpcServer.handleCustomTokenDetail,
	listCustomTokenHolders:              RpcServer.handleGetListCustomTokenHolders,
	getListCustomTokenBalance:           RpcServer.handleGetListCustomTokenBalance,
	
	// custom token which support privacy
	createRawPrivacyCustomTokenTransaction:     RpcServer.handleCreateRawPrivacyCustomTokenTransaction,
	sendRawPrivacyCustomTokenTransaction:       RpcServer.handleSendRawPrivacyCustomTokenTransaction,
	createAndSendPrivacyCustomTokenTransaction: RpcServer.handleCreateAndSendPrivacyCustomTokenTransaction,
	listPrivacyCustomToken:                     RpcServer.handleListPrivacyCustomToken,
	privacyCustomTokenTxs:                      RpcServer.handlePrivacyCustomTokenDetail,
	getListPrivacyCustomTokenBalance:           RpcServer.handleGetListPrivacyCustomTokenBalance,
	
	// Bridge
	createIssuingRequest:            RpcServer.handleCreateIssuingRequest,
	sendIssuingRequest:              RpcServer.handleSendIssuingRequest,
	createAndSendIssuingRequest:     RpcServer.handleCreateAndSendIssuingRequest,
	createAndSendContractingRequest: RpcServer.handleCreateAndSendContractingRequest,
	getBridgeTokensAmounts:          RpcServer.handleGetBridgeTokensAmounts,
	
	// wallet
	getPublicKeyFromPaymentAddress: RpcServer.handleGetPublicKeyFromPaymentAddress,
	defragmentAccount:              RpcServer.handleDefragmentAccount,
	
	getStackingAmount: RpcServer.handleGetStakingAmount,
	
	hashToIdenticon: RpcServer.handleHashToIdenticon,
	
	//reward
	CreateRawWithDrawTransaction: RpcServer.handleCreateAndSendWithDrawTransaction,
	getRewardAmount:              RpcServer.handleGetRewardAmount,
	
	//revert
	revertbeaconchain: RpcServer.handleRevertBeacon,
	revertshardchain:  RpcServer.handleRevertShard,
}

// Commands that are available to a limited user
var LimitedHttpHandler = map[string]httpHandler{
	// local WALLET
	listAccounts:                       RpcServer.handleListAccounts,
	getAccount:                         RpcServer.handleGetAccount,
	getAddressesByAccount:              RpcServer.handleGetAddressesByAccount,
	getAccountAddress:                  RpcServer.handleGetAccountAddress,
	dumpPrivkey:                        RpcServer.handleDumpPrivkey,
	importAccount:                      RpcServer.handleImportAccount,
	removeAccount:                      RpcServer.handleRemoveAccount,
	listUnspentOutputCoins:             RpcServer.handleListUnspentOutputCoins,
	getBalance:                         RpcServer.handleGetBalance,
	getBalanceByPrivatekey:             RpcServer.handleGetBalanceByPrivatekey,
	getBalanceByPaymentAddress:         RpcServer.handleGetBalanceByPaymentAddress,
	getReceivedByAccount:               RpcServer.handleGetReceivedByAccount,
	setTxFee:                           RpcServer.handleSetTxFee,
	getRecentTransactionsByBlockNumber: RpcServer.handleGetRecentTransactionsByBlockNumber,
}

var WsHandler = map[string]wsHandler{
	subcribeNewBlock:                           RpcServer.handleSubcribeNewBlock,
}
