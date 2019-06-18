package rpcserver

// rpc cmd method
const (
	startProfiling = "startprofiling"
	stopProfiling  = "stopprofiling"

	getNetworkInfo     = "getnetworkinfo"
	getConnectionCount = "getconnectioncount"
	getAllPeers        = "getallpeers"

	estimateFee              = "estimatefee"
	estimateFeeWithEstimator = "estimatefeewithestimator"

	getActiveShards    = "getactiveshards"
	getMaxShardsNumber = "getmaxshardsnumber"

	getMiningInfo                 = "getmininginfo"
	getRawMempool                 = "getrawmempool"
	getNumberOfTxsInMempool       = "getnumberoftxsinmempool"
	getMempoolEntry               = "getmempoolentry"
	getBeaconPoolState            = "getbeaconpoolstate"
	getShardPoolState             = "getshardpoolstate"
	getShardPoolLatestValidHeight = "getshardpoollatestvalidheight"
	getShardToBeaconPoolState     = "getshardtobeaconpoolstate"
	getCrossShardPoolState        = "getcrossshardpoolstate"
	getNextCrossShard             = "getnextcrossshard"
	getShardToBeaconPoolStateV2   = "getshardtobeaconpoolstatev2"
	getCrossShardPoolStateV2      = "getcrossshardpoolstatev2"
	getShardPoolStateV2           = "getshardpoolstatev2"
	getBeaconPoolStateV2          = "getbeaconpoolstatev2"

	getBestBlock        = "getbestblock"
	getBestBlockHash    = "getbestblockhash"
	getBlocks           = "getblocks"
	retrieveBlock       = "retrieveblock"
	retrieveBeaconBlock = "retrievebeaconblock"
	getBlockChainInfo   = "getblockchaininfo"
	getBlockCount       = "getblockcount"
	getBlockHash        = "getblockhash"

	listOutputCoins                            = "listoutputcoins"
	createRawTransaction                       = "createtransaction"
	sendRawTransaction                         = "sendtransaction"
	createAndSendTransaction                   = "createandsendtransaction"
	createAndSendCustomTokenTransaction        = "createandsendcustomtokentransaction"
	sendRawCustomTokenTransaction              = "sendrawcustomtokentransaction"
	createRawCustomTokenTransaction            = "createrawcustomtokentransaction"
	createRawPrivacyCustomTokenTransaction     = "createrawprivacycustomtokentransaction"
	sendRawPrivacyCustomTokenTransaction       = "sendrawprivacycustomtokentransaction"
	createAndSendPrivacyCustomTokenTransaction = "createandsendprivacycustomtokentransaction"
	getMempoolInfo                             = "getmempoolinfo"
	getCandidateList                           = "getcandidatelist"
	getCommitteeList                           = "getcommitteelist"
	canPubkeyStake                             = "canpubkeystake"
	getTotalTransaction                        = "gettotaltransaction"
	getBlockProducerList                       = "getblockproducer"
	listUnspentCustomToken                     = "listunspentcustomtoken"
	getTransactionByHash                       = "gettransactionbyhash"
	listCustomToken                            = "listcustomtoken"
	listPrivacyCustomToken                     = "listprivacycustomtoken"
	customTokenTxs                             = "customtoken"
	listCustomTokenHolders                     = "customtokenholder"
	privacyCustomTokenTxs                      = "privacycustomtoken"
	checkHashValue                             = "checkhashvalue"
	getListCustomTokenBalance                  = "getlistcustomtokenbalance"
	getListPrivacyCustomTokenBalance           = "getlistprivacycustomtokenbalance"
	getBlockHeader                             = "getheader"
	getCrossShardBlock                         = "getcrossshardblock"
	randomCommitments                          = "randomcommitments"
	hasSerialNumbers                           = "hasserialnumbers"
	hasSnDerivators                            = "hassnderivators"

	createAndSendStakingTransaction = "createandsendstakingtransaction"

	//===========For Testing and Benchmark==============
	getAndSendTxsFromFile   = "getandsendtxsfromfile"
	getAndSendTxsFromFileV2 = "getandsendtxsfromfilev2"
	unlockMempool           = "unlockmempool"
	//==================================================

	getShardBestState  = "getshardbeststate"
	getBeaconBestState = "getbeaconbeststate"

	// Wallet rpc cmd
	listAccounts               = "listaccounts"
	getAccount                 = "getaccount"
	getAddressesByAccount      = "getaddressesbyaccount"
	getAccountAddress          = "getaccountaddress"
	dumpPrivkey                = "dumpprivkey"
	importAccount              = "importaccount"
	removeAccount              = "removeaccount"
	listUnspentOutputCoins     = "listunspentoutputcoins"
	getBalance                 = "getbalance"
	getBalanceByPrivatekey     = "getbalancebyprivatekey"
	getBalanceByPaymentAddress = "getbalancebypaymentaddress"
	getReceivedByAccount       = "getreceivedbyaccount"
	setTxFee                   = "settxfee"

	// walletsta
	getPublicKeyFromPaymentAddress = "getpublickeyfrompaymentaddress"
	defragmentAccount              = "defragmentaccount"

	getStackingAmount = "getstackingamount"

	// utils
	hashToIdenticon = "hashtoidenticon"

	createIssuingRequest            = "createissuingrequest"
	sendIssuingRequest              = "sendissuingrequest"
	createAndSendIssuingRequest     = "createandsendissuingrequest"
	createAndSendContractingRequest = "createandsendcontractingrequest"
	getBridgeTokensAmounts          = "getbridgetokensamounts"

	// Incognito -> Ethereum bridge
	getBeaconSwapProof = "getbeaconswapproof"

	// reward
	CreateRawWithDrawTransaction = "withdrawreward"
	getRewardAmount              = "getrewardamount"

	revertbeaconchain = "revertbeaconchain"
	revertshardchain  = "revertshardchain"
)
