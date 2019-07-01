package rpcserver

type httpHandler func(*HttpServer, interface{}, <-chan struct{}) (interface{}, *RPCError)
type wsHandler func(*WsServer, interface{}, string, chan RpcSubResult, <-chan struct{})

// Commands valid for normal user
var HttpHandler = map[string]httpHandler{

	startProfiling: (*HttpServer).handleStartProfiling,
	stopProfiling:  (*HttpServer).handleStopProfiling,
	// node
	getNetworkInfo:           (*HttpServer).handleGetNetWorkInfo,
	getConnectionCount:       (*HttpServer).handleGetConnectionCount,
	getAllPeers:              (*HttpServer).handleGetAllPeers,
	estimateFee:              (*HttpServer).handleEstimateFee,
	estimateFeeWithEstimator: (*HttpServer).handleEstimateFeeWithEstimator,
	getActiveShards:          (*HttpServer).handleGetActiveShards,
	getMaxShardsNumber:       (*HttpServer).handleGetMaxShardsNumber,
	//pool
	getMiningInfo:               (*HttpServer).handleGetMiningInfo,
	getRawMempool:               (*HttpServer).handleGetRawMempool,
	getNumberOfTxsInMempool:     (*HttpServer).handleGetNumberOfTxsInMempool,
	getMempoolEntry:             (*HttpServer).handleMempoolEntry,
	getShardToBeaconPoolStateV2: (*HttpServer).handleGetShardToBeaconPoolStateV2,
	getCrossShardPoolStateV2:    (*HttpServer).handleGetCrossShardPoolStateV2,
	getShardPoolStateV2:         (*HttpServer).handleGetShardPoolStateV2,
	getBeaconPoolStateV2:        (*HttpServer).handleGetBeaconPoolStateV2,
	getShardToBeaconPoolState:   (*HttpServer).handleGetShardToBeaconPoolState,
	getCrossShardPoolState:      (*HttpServer).handleGetCrossShardPoolState,
	getNextCrossShard:           (*HttpServer).handleGetNextCrossShard,
	// block
	getBestBlock:        (*HttpServer).handleGetBestBlock,
	getBestBlockHash:    (*HttpServer).handleGetBestBlockHash,
	retrieveBlock:       (*HttpServer).handleRetrieveBlock,
	retrieveBeaconBlock: (*HttpServer).handleRetrieveBeaconBlock,
	getBlocks:           (*HttpServer).handleGetBlocks,
	getBlockChainInfo:   (*HttpServer).handleGetBlockChainInfo,
	getBlockCount:       (*HttpServer).handleGetBlockCount,
	getBlockHash:        (*HttpServer).handleGetBlockHash,
	checkHashValue:      (*HttpServer).handleCheckHashValue, // get data in blockchain from hash value
	getBlockHeader:      (*HttpServer).handleGetBlockHeader, // Current committee, next block committee and candidate is included in block header
	getCrossShardBlock:  (*HttpServer).handleGetCrossShardBlock,
	// transaction
	listOutputCoins:                 (*HttpServer).handleListOutputCoins,
	createRawTransaction:            (*HttpServer).handleCreateRawTransaction,
	sendRawTransaction:              (*HttpServer).handleSendRawTransaction,
	createAndSendTransaction:        (*HttpServer).handleCreateAndSendTx,
	getMempoolInfo:                  (*HttpServer).handleGetMempoolInfo,
	getTransactionByHash:            (*HttpServer).handleGetTransactionByHash,
	createAndSendStakingTransaction: (*HttpServer).handleCreateAndSendStakingTx,
	randomCommitments:               (*HttpServer).handleRandomCommitments,
	hasSerialNumbers:                (*HttpServer).handleHasSerialNumbers,
	hasSnDerivators:                 (*HttpServer).handleHasSnDerivators,
	//======Testing and Benchmark======
	getAndSendTxsFromFile:   (*HttpServer).handleGetAndSendTxsFromFile,
	getAndSendTxsFromFileV2: (*HttpServer).handleGetAndSendTxsFromFileV2,
	unlockMempool:           (*HttpServer).handleUnlockMempool,
	//=================================
	// Beststate
	getCandidateList:              (*HttpServer).handleGetCandidateList,
	getCommitteeList:              (*HttpServer).handleGetCommitteeList,
	getBlockProducerList:          (*HttpServer).handleGetBlockProducerList,
	getShardBestState:             (*HttpServer).handleGetShardBestState,
	getBeaconBestState:            (*HttpServer).handleGetBeaconBestState,
	getBeaconPoolState:            (*HttpServer).handleGetBeaconPoolState,
	getShardPoolState:             (*HttpServer).handleGetShardPoolState,
	getShardPoolLatestValidHeight: (*HttpServer).handleGetShardPoolLatestValidHeight,
	canPubkeyStake:                (*HttpServer).handleCanPubkeyStake,
	getTotalTransaction:           (*HttpServer).handleGetTotalTransaction,
	// custom token
	createRawCustomTokenTransaction:     (*HttpServer).handleCreateRawCustomTokenTransaction,
	sendRawCustomTokenTransaction:       (*HttpServer).handleSendRawCustomTokenTransaction,
	createAndSendCustomTokenTransaction: (*HttpServer).handleCreateAndSendCustomTokenTransaction,
	listUnspentCustomToken:              (*HttpServer).handleListUnspentCustomToken,
	listCustomToken:                     (*HttpServer).handleListCustomToken,
	customTokenTxs:                      (*HttpServer).handleCustomTokenDetail,
	listCustomTokenHolders:              (*HttpServer).handleGetListCustomTokenHolders,
	getListCustomTokenBalance:           (*HttpServer).handleGetListCustomTokenBalance,
	// custom token which support privacy
	createRawPrivacyCustomTokenTransaction:     (*HttpServer).handleCreateRawPrivacyCustomTokenTransaction,
	sendRawPrivacyCustomTokenTransaction:       (*HttpServer).handleSendRawPrivacyCustomTokenTransaction,
	createAndSendPrivacyCustomTokenTransaction: (*HttpServer).handleCreateAndSendPrivacyCustomTokenTransaction,
	listPrivacyCustomToken:                     (*HttpServer).handleListPrivacyCustomToken,
	privacyCustomTokenTxs:                      (*HttpServer).handlePrivacyCustomTokenDetail,
	getListPrivacyCustomTokenBalance:           (*HttpServer).handleGetListPrivacyCustomTokenBalance,
	// Bridge
	createIssuingRequest:            (*HttpServer).handleCreateIssuingRequest,
	sendIssuingRequest:              (*HttpServer).handleSendIssuingRequest,
	createAndSendIssuingRequest:     (*HttpServer).handleCreateAndSendIssuingRequest,
	createAndSendContractingRequest: (*HttpServer).handleCreateAndSendContractingRequest,
	getBridgeTokensAmounts:          (*HttpServer).handleGetBridgeTokensAmounts,
	// wallet
	getPublicKeyFromPaymentAddress:        (*HttpServer).handleGetPublicKeyFromPaymentAddress,
	defragmentAccount:                     (*HttpServer).handleDefragmentAccount,
	getStackingAmount:                     (*HttpServer).handleGetStakingAmount,
	hashToIdenticon:                       (*HttpServer).handleHashToIdenticon,
	createAndSendBurningRequest:           (*HttpServer).handleCreateAndSendBurningRequest,
	createAndSendTxWithETHHeadersRelaying: (*HttpServer).handleCreateAndSendTxWithETHHeadersRelaying,
	createAndSendTxWithIssuingETHReq:      (*HttpServer).handleCreateAndSendTxWithIssuingETHReq,
	getRelayedETHHeader:                   (*HttpServer).handleGetRelayedETHHeader,

	// Incognito -> Ethereum bridge
	getBeaconSwapProof: (*HttpServer).handleGetBeaconSwapProof,
	getBridgeSwapProof: (*HttpServer).handleGetBridgeSwapProof,
	getBurnProof:       (*HttpServer).handleGetBurnProof,

	//reward
	CreateRawWithDrawTransaction: (*HttpServer).handleCreateAndSendWithDrawTransaction,
	getRewardAmount:              (*HttpServer).handleGetRewardAmount,
	//revert
	revertbeaconchain: (*HttpServer).handleRevertBeacon,
	revertshardchain:  (*HttpServer).handleRevertShard,
}

// Commands that are available to a limited user
var LimitedHttpHandler = map[string]httpHandler{
	// local WALLET
	listAccounts:               (*HttpServer).handleListAccounts,
	getAccount:                 (*HttpServer).handleGetAccount,
	getAddressesByAccount:      (*HttpServer).handleGetAddressesByAccount,
	getAccountAddress:          (*HttpServer).handleGetAccountAddress,
	dumpPrivkey:                (*HttpServer).handleDumpPrivkey,
	importAccount:              (*HttpServer).handleImportAccount,
	removeAccount:              (*HttpServer).handleRemoveAccount,
	listUnspentOutputCoins:     (*HttpServer).handleListUnspentOutputCoins,
	getBalance:                 (*HttpServer).handleGetBalance,
	getBalanceByPrivatekey:     (*HttpServer).handleGetBalanceByPrivatekey,
	getBalanceByPaymentAddress: (*HttpServer).handleGetBalanceByPaymentAddress,
	getReceivedByAccount:       (*HttpServer).handleGetReceivedByAccount,
	setTxFee:                   (*HttpServer).handleSetTxFee,
}

var WsHandler = map[string]wsHandler{
	testSubcrice:                (*WsServer).handleTestSubcribe,
	subcribeNewShardBlock:       (*WsServer).handleSubscribeNewShardBlock,
	subcribeNewBeaconBlock:      (*WsServer).handleSubscribeNewBeaconBlock,
	subcribePendingTransaction:  (*WsServer).handleSubscribePendingTransaction,
	subcribeMempoolInfo:         (*WsServer).handleSubcribeMempoolInfo,
	subcribeShardBestState:      (*WsServer).handleSubscribeShardBestState,
	subcribeBeaconBestState:     (*WsServer).handleSubscribeBeaconBestState,
	subcribeBeaconPoolBeststate: (*WsServer).handleSubscribeBeaconPoolBestState,
	subcribeShardPoolBeststate:  (*WsServer).handleSubscribeShardPoolBeststate,
}
