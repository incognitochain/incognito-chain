package rpcserver

import "github.com/incognitochain/incognito-chain/common"

// rpc cmd method
const (
	// test rpc server
	testHttpServer = "testrpcserver"
	startProfiling = "startprofiling"
	stopProfiling  = "stopprofiling"
	exportMetrics  = "exportmetrics"

	getNetworkInfo       = "getnetworkinfo"
	getConnectionCount   = "getconnectioncount"
	getAllConnectedPeers = "getallconnectedpeers"
	getAllPeers          = "getallpeers"
	getNodeRole          = "getnoderole"
	getInOutMessages     = "getinoutmessages"
	getInOutMessageCount = "getinoutmessagecount"

	estimateFee              = "estimatefee"
	estimateFeeV2            = "estimatefeev2"
	estimateFeeWithEstimator = "estimatefeewithestimator"

	getActiveShards    = "getactiveshards"
	getMaxShardsNumber = "getmaxshardsnumber"

	getMiningInfo                 = "getmininginfo"
	getRawMempool                 = "getrawmempool"
	getSyncPoolValidator          = "getsyncpoolvalidator"
	getSyncPoolValidatorDetail    = "getsyncpoolvalidatordetail"
	getMempoolInfoDetails         = "getmempoolinfodetails"
	getNumberOfTxsInMempool       = "getnumberoftxsinmempool"
	getMempoolEntry               = "getmempoolentry"
	removeTxInMempool             = "removetxinmempool"
	getBeaconPoolState            = "getbeaconpoolstate"
	getShardPoolState             = "getshardpoolstate"
	getShardPoolLatestValidHeight = "getshardpoollatestvalidheight"
	// getShardToBeaconPoolState     = "getshardtobeaconpoolstate"
	// getCrossShardPoolState        = "getcrossshardpoolstate"
	getNextCrossShard           = "getnextcrossshard"
	getShardToBeaconPoolStateV2 = "getshardtobeaconpoolstatev2"
	getCrossShardPoolStateV2    = "getcrossshardpoolstatev2"
	getShardPoolStateV2         = "getshardpoolstatev2"
	getBeaconPoolStateV2        = "getbeaconpoolstatev2"
	// getFeeEstimator             = "getfeeestimator"
	getLatestBackup             = "getlatestbackup"
	getBootstrapStateDB         = "getbootstrapstatedb"
	getBestBlock                = "getbestblock"
	getBestBlockHash            = "getbestblockhash"
	getBlocks                   = "getblocks"
	retrieveBlock               = "retrieveblock"
	retrieveBlockByHeight       = "retrieveblockbyheight"
	retrieveBeaconBlock         = "retrievebeaconblock"
	retrieveBeaconBlockByHeight = "retrievebeaconblockbyheight"
	getBlockChainInfo           = "getblockchaininfo"
	getBlockCount               = "getblockcount"
	getBlockHash                = "getblockhash"

	listOutputCoins                            = "listoutputcoins"
	listOutputCoinsFromCache                   = "listoutputcoinsfromcache"
	listOutputTokens                           = "listoutputtokens"
	createRawTransaction                       = "createtransaction"
	sendRawTransaction                         = "sendtransaction"
	createAndSendTransaction                   = "createandsendtransaction"
	createConvertCoinVer1ToVer2Transaction     = "createconvertcoinver1tover2transaction"
	createAndSendCustomTokenTransaction        = "createandsendcustomtokentransaction"
	sendRawCustomTokenTransaction              = "sendrawcustomtokentransaction"
	createRawCustomTokenTransaction            = "createrawcustomtokentransaction"
	createConvertCoinVer1ToVer2TxToken         = "createconvertcoinver1tover2txtoken"
	createRawPrivacyCustomTokenTransaction     = "createrawprivacycustomtokentransaction"
	sendRawPrivacyCustomTokenTransaction       = "sendrawprivacycustomtokentransaction"
	createAndSendPrivacyCustomTokenTransaction = "createandsendprivacycustomtokentransaction"
	getMempoolInfo                             = "getmempoolinfo"
	getPendingTxsInBlockgen                    = "getpendingtxsinblockgen"
	getCandidateList                           = "getcandidatelist"
	getCommitteeList                           = "getcommitteelist"
	canPubkeyStake                             = "canpubkeystake"
	getTotalTransaction                        = "gettotaltransaction"
	listUnspentCustomToken                     = "listunspentcustomtoken"
	getBalanceCustomToken                      = "getbalancecustomtoken"
	getTransactionByHash                       = "gettransactionbyhash"
	getEncodedTransactionsByHashes             = "getencodedtransactionsbyhashes"
	gettransactionhashbyreceiver               = "gettransactionhashbyreceiver"
	gettransactionhashbyreceiverv2             = "gettransactionhashbyreceiverv2"
	gettransactionbyreceiver                   = "gettransactionbyreceiver"
	gettransactionbyreceiverv2                 = "gettransactionbyreceiverv2"
	gettransactionbyserialnumber               = "gettransactionbyserialnumber"
	gettransactionbypublickey                  = "gettransactionbypublickey"
	listUnspentOutputTokens                    = "listunspentoutputtokens"
	getOTACoinLength                           = "getotacoinlength"
	getOTACoinsByIndices                       = "getotacoinsbyindices"
	randomCommitments                          = "randomcommitments"
	hasSerialNumbers                           = "hasserialnumbers"
	hasSnDerivators                            = "hassnderivators"
	listSnDerivators                           = "listsnderivators"
	listSerialNumbers                          = "listserialnumbers"
	listCommitments                            = "listcommitments"
	listCommitmentIndices                      = "listcommitmentindices"
	createAndSendStakingTransaction            = "createandsendstakingtransaction"
	createAndSendAddStakingTransaction         = "createandsendaddstakingtransaction"
	createAndSendStopAutoStakingTransaction    = "createandsendstopautostakingtransaction"
	createAndSendTokenInitTransaction          = "createandsendtokeninittransaction"
	decryptoutputcoinbykeyoftransaction        = "decryptoutputcoinbykeyoftransaction"
	randomCommitmentsAndPublicKeys             = "randomcommitmentsandpublickeys"

	createAndSendTransactionV2                   = "createandsendtransactionv2"
	createAndSendPrivacyCustomTokenTransactionV2 = "createandsendprivacycustomtokentransactionv2"
	listCustomToken                              = "listcustomtoken"
	listPrivacyCustomToken                       = "listprivacycustomtoken"
	listPrivacyCustomTokenIDs                    = "listprivacycustomtokenids"
	getPrivacyCustomToken                        = "getprivacycustomtoken"
	listPrivacyCustomTokenByShard                = "listprivacycustomtokenbyshard"
	getBalancePrivacyCustomToken                 = "getbalanceprivacycustomtoken"
	customTokenTxs                               = "customtoken"
	listCustomTokenHolders                       = "customtokenholder"
	privacyCustomTokenTxs                        = "privacycustomtoken"
	checkHashValue                               = "checkhashvalue"
	getListCustomTokenBalance                    = "getlistcustomtokenbalance"
	getListPrivacyCustomTokenBalance             = "getlistprivacycustomtokenbalance"
	getBlockHeader                               = "getheader"
	getCrossShardBlock                           = "getcrossshardblock"
	getBlocksFromHeight                          = "getblocksfromheight"
	hasSerialNumbersInMempool                    = "hasserialnumbersinmempool"
	createAndSendStakingTransactionV2            = "createandsendstakingtransactionv2"
	createAndSendStopAutoStakingTransactionV2    = "createandsendstopautostakingtransactionv2"

	// ===========For Testing and Benchmark==============
	getAndSendTxsFromFile              = "getandsendtxsfromfile"
	getAndSendTxsFromFileV2            = "getandsendtxsfromfilev2"
	unlockMempool                      = "unlockmempool"
	handleGetConsensusInfoV3           = "getconsensusinfov3"
	getAutoStakingByHeight             = "getautostakingbyheight"
	sendFinishSync                     = "sendfinishsync"
	setAutoEnableFeatureConfig         = "setautoenablefeatureconfig"
	getAutoEnableFeatureConfig         = "getautoenablefeatureconfig"
	getCommitteeState                  = "getcommitteestate"
	convertPaymentAddress              = "convertpaymentaddress"
	getTotalBlockInEpoch               = "gettotalblockinepoch"
	getDetailBlocksOfEpoch             = "getdetailblocksofepoch"
	getCommitteeStateByShard           = "getcommitteestatebyshard"
	getShardCommitteeStateByBeaconHash = "getshardcommitteestatebybeaconhash"
	getSlashingCommittee               = "getslashingcommittee"
	getSlashingCommitteeDetail         = "getslashingcommitteedetail"
	getFinalityProof                   = "getfinalityproof"
	setConsensusRule                   = "setconsensusrule"
	getConsensusRule                   = "getconsensusrule"
	getByzantineDetectorInfo           = "getbyzantinedetectorinfo"
	getByzantineBlackList              = "getbyzantineblacklist"
	removeByzantineDetector            = "removebyzantinedetector"
	getConsensusData                   = "getconsensusdata"
	getProposerIndex                   = "getproposerindex"
	resetCache                         = "resetcache"
	testValidate                       = "testvalidate"
	stallInsert                        = "stallInsert"
	// ==================================================

	getShardBestState        = "getshardbeststate"
	getShardBestStateDetail  = "getshardbeststatedetail"
	getBeaconViewByHash      = "getbeaconviewbyhash"
	getBeaconBestState       = "getbeaconbeststate"
	getBeaconBestStateDetail = "getbeaconbeststatedetail"

	// Wallet rpc cmd
	listAccounts                    = "listaccounts"
	getAccount                      = "getaccount"
	getAddressesByAccount           = "getaddressesbyaccount"
	getAccountAddress               = "getaccountaddress"
	dumpPrivkey                     = "dumpprivkey"
	importAccount                   = "importaccount"
	removeAccount                   = "removeaccount"
	listUnspentOutputCoins          = "listunspentoutputcoins"
	listUnspentOutputCoinsFromCache = "listunspentoutputcoinsfromcache"
	getBalance                      = "getbalance"
	getBalanceByPrivatekey          = "getbalancebyprivatekey"
	getBalanceByPaymentAddress      = "getbalancebypaymentaddress"
	getReceivedByAccount            = "getreceivedbyaccount"
	setTxFee                        = "settxfee"
	submitKey                       = "submitkey"
	authorizedSubmitKey             = "authorizedsubmitkey"
	getKeySubmissionInfo            = "getkeysubmissioninfo"

	// walletsta
	getPublicKeyFromPaymentAddress = "getpublickeyfrompaymentaddress"
	defragmentAccount              = "defragmentaccount"
	defragmentAccountV2            = "defragmentaccountv2"
	defragmentAccountToken         = "defragmentaccounttoken"
	defragmentAccountTokenV2       = "defragmentaccounttokenv2"

	getStackingAmount = "getstackingamount"

	// utils
	hashToIdenticon = "hashtoidenticon"
	generateTokenID = "generatetokenid"

	createIssuingRequest                  = "createissuingrequest"
	sendIssuingRequest                    = "sendissuingrequest"
	createAndSendIssuingRequest           = "createandsendissuingrequest"
	createAndSendIssuingRequestV2         = "createandsendissuingrequestv2"
	createAndSendContractingRequest       = "createandsendcontractingrequest"
	createAndSendContractingRequestV2     = "createandsendcontractingrequestv2"
	createAndSendBurningRequest           = "createandsendburningrequest"
	createAndSendBurningRequestV2         = "createandsendburningrequestv2"
	createAndSendTxWithIssuingETHReq      = "createandsendtxwithissuingethreq"
	createAndSendTxWithIssuingETHReqV2    = "createandsendtxwithissuingethreqv2"
	checkETHHashIssued                    = "checkethhashissued"
	getAllBridgeTokens                    = "getallbridgetokens"
	getAllBridgeTokensByHeight            = "getallbridgetokensbyheight"
	getETHHeaderByHash                    = "getethheaderbyhash"
	getBridgeReqWithStatus                = "getbridgereqwithstatus"
	createAndSendTxWithIssuingBSCReq      = "createandsendtxwithissuingbscreq"
	createAndSendTxWithIssuingPRVERC20Req = "createandsendtxwithissuingprverc20req"
	createAndSendTxWithIssuingPRVBEP20Req = "createandsendtxwithissuingprvbep20req"
	createAndSendBurningBSCRequest        = "createandsendburningbscrequest"
	checkBSCHashIssued                    = "checkbschashissued"
	checkPRVPeggingHashIssued             = "checkpeggedprvhashissued"
	createAndSendBurningPRVERC20Request   = "createandsendburningprverc20request"
	createAndSendBurningPRVBEP20Request   = "createandsendburningprvbep20request"
	createAndSendBurningPLGRequest        = "createandsendburningplgrequest"
	createAndSendTxWithIssuingPLGReq      = "createandsendtxwithissuingplgreq"
	createAndSendTxWithIssuingFTMReq      = "createandsendtxwithissuingftmreq"
	createAndSendBurningFTMRequest        = "createandsendburningftmrequest"
	createAndSendTxWithIssuingAURORAReq   = "createandsendtxwithissuingaurorareq"
	createAndSendTxWithIssuingAVAXReq     = "createandsendtxwithissuingavaxreq"
	createAndSendBurningAURORARequest     = "createandsendburningaurorarequest"
	createAndSendBurningAVAXRequest       = "createandsendburningavaxrequest"
	createAndSendTxWithIssuingNearReq     = "createandsendtxwithissuingnearreq"
	createAndSendBurningNearRequest       = "createandsendburningnearrequest"
	createAndSendBurningPRVRequest        = "createandsendburningprvrequest"

	// Incognito -> Ethereum bridge
	getBeaconSwapProof       = "getbeaconswapproof"
	getLatestBeaconSwapProof = "getlatestbeaconswapproof"
	getBridgeSwapProof       = "getbridgeswapproof"
	getLatestBridgeSwapProof = "getlatestbridgeswapproof"
	getBurnProof             = "getburnproof"
	getBSCBurnProof          = "getbscburnproof"
	getPRVERC20BurnProof     = "getprverc20burnproof"
	getPRVBEP20BurnProof     = "getprvbep20burnproof"
	getPLGBurnProof          = "getplgburnproof"
	getFTMBurnProof          = "getftmburnproof"
	getAURORABurnProof       = "getauroraburnproof"
	getAVAXBurnProof         = "getavaxburnproof"
	getNearBurnProof         = "getnearburnproof"
	getPrvBurnProof          = "getprvburnproof"

	// reward
	CreateRawWithDrawTransaction = "withdrawreward"
	getRewardAmount              = "getrewardamount"
	getRewardAmountByPublicKey   = "getrewardamountbypublickey"
	listRewardAmount             = "listrewardamount"

	revertbeaconchain = "revertbeaconchain"
	revertshardchain  = "revertshardchain"

	enableMining                = "enablemining"
	getChainMiningStatus        = "getchainminingstatus"
	getPublickeyMining          = "getpublickeymining"
	getPublicKeyRole            = "getpublickeyrole"
	getRoleByValidatorKey       = "getrolebyvalidatorkey"
	getIncognitoPublicKeyRole   = "getincognitopublickeyrole"
	getMinerRewardFromMiningKey = "getminerrewardfromminingkey"

	// pde
	getPDEState                                = "getpdestate"
	createAndSendTxWithWithdrawalReq           = "createandsendtxwithwithdrawalreq"
	createAndSendTxWithWithdrawalReqV2         = "createandsendtxwithwithdrawalreqv2"
	createAndSendTxWithPDEFeeWithdrawalReq     = "createandsendtxwithpdefeewithdrawalreq"
	createAndSendTxWithPTokenTradeReq          = "createandsendtxwithptokentradereq"
	createAndSendTxWithPTokenCrossPoolTradeReq = "createandsendtxwithptokencrosspooltradereq"
	createAndSendTxWithPRVTradeReq             = "createandsendtxwithprvtradereq"
	createAndSendTxWithPRVCrossPoolTradeReq    = "createandsendtxwithprvcrosspooltradereq"
	createAndSendTxWithPTokenContribution      = "createandsendtxwithptokencontribution"
	createAndSendTxWithPRVContribution         = "createandsendtxwithprvcontribution"
	createAndSendTxWithPTokenContributionV2    = "createandsendtxwithptokencontributionv2"
	createAndSendTxWithPRVContributionV2       = "createandsendtxwithprvcontributionv2"
	convertNativeTokenToPrivacyToken           = "convertnativetokentoprivacytoken"
	convertPrivacyTokenToNativeToken           = "convertprivacytokentonativetoken"
	getPDEContributionStatus                   = "getpdecontributionstatus"
	getPDEContributionStatusV2                 = "getpdecontributionstatusv2"
	getPDETradeStatus                          = "getpdetradestatus"
	getPDEWithdrawalStatus                     = "getpdewithdrawalstatus"
	getPDEFeeWithdrawalStatus                  = "getpdefeewithdrawalstatus"
	convertPDEPrices                           = "convertpdeprices"
	extractPDEInstsFromBeaconBlock             = "extractpdeinstsfrombeaconblock"
	//

	// pDex v3
	pdexv3MintNft                         = "pdexv3_txMintNft"
	getPdexv3State                        = "pdexv3_getState"
	createAndSendTxWithPdexv3ModifyParams = "pdexv3_txModifyParams"
	getPdexv3ParamsModifyingStatus        = "pdexv3_getParamsModifyingStatus"
	pdexv3AddLiquidityV3                  = "pdexv3_txAddLiquidity"
	pdexv3WithdrawLiquidityV3             = "pdexv3_txWithdrawLiquidity"
	getPdexv3ContributionStatus           = "pdexv3_getContributionStatus"

	getPdexv3WithdrawLiquidityStatus               = "pdexv3_getWithdrawLiquidityStatus"
	getPdexv3MintNftStatus                         = "pdexv3_getMintNftStatus"
	pdexv3TxTrade                                  = "pdexv3_txTrade"
	pdexv3TxAddOrder                               = "pdexv3_txAddOrder"
	pdexv3TxWithdrawOrder                          = "pdexv3_txWithdrawOrder"
	pdexv3GetTradeStatus                           = "pdexv3_getTradeStatus"
	pdexv3GetAddOrderStatus                        = "pdexv3_getAddOrderStatus"
	pdexv3GetWithdrawOrderStatus                   = "pdexv3_getWithdrawOrderStatus"
	pdexv3Staking                                  = "pdexv3_txStake"
	pdexv3Unstaking                                = "pdexv3_txUnstake"
	pdexv3GetStakingStatus                         = "pdexv3_getStakingStatus"
	pdexv3GetUnstakingStatus                       = "pdexv3_getUnstakingStatus"
	getPdexv3EstimatedLPValue                      = "pdexv3_getEstimatedLPValue"
	getPdexv3EstimatedLPPoolReward                 = "pdexv3_getEstimatedLPPoolReward"
	createAndSendTxWithPdexv3WithdrawLPFee         = "pdexv3_txWithdrawLPFee"
	getPdexv3WithdrawalLPFeeStatus                 = "pdexv3_getWithdrawalLPFeeStatus"
	createAndSendTxWithPdexv3WithdrawProtocolFee   = "pdexv3_txWithdrawProtocolFee"
	getPdexv3WithdrawalProtocolFeeStatus           = "pdexv3_getWithdrawalProtocolFeeStatus"
	getPdexv3EstimatedStakingReward                = "pdexv3_getEstimatedStakingReward"
	getPdexv3EstimatedStakingPoolReward            = "pdexv3_getEstimatedStakingPoolReward"
	createAndSendTxWithPdexv3WithdrawStakingReward = "pdexv3_txWithdrawStakingReward"
	getPdexv3WithdrawalStakingRewardStatus         = "pdexv3_getWithdrawalStakingRewardStatus"

	// bridgeagg method
	bridgeaggState                       = "bridgeaggGetState"
	bridgeaggModifyParam                 = "bridgeaggModifyParam"
	bridgeaggStatusModifyParam           = "bridgeaggStatusModifyParam"
	bridgeaggConvert                     = "bridgeaggConvert"
	bridgeaggStatusConvert               = "bridgeaggGetStatusConvert"
	bridgeaggShield                      = "bridgeaggShield"
	bridgeaggStatusShield                = "bridgeaggGetStatusShield"
	bridgeaggUnshield                    = "bridgeaggUnshield"
	bridgeaggBurnForCall                 = "bridgeaggBurnForCall"
	bridgeaggStatusUnshield              = "bridgeaggGetStatusUnshield"
	bridgeaggEstimateFeeByExpectedAmount = "bridgeaggEstimateFeeByExpectedAmount"
	bridgeaggEstimateFeeByBurntAmount    = "bridgeaggEstimateFeeByBurntAmount"
	bridgeaggEstimateReward              = "bridgeaggEstimateReward"
	bridgeaggGetBurnProof                = "bridgeaggGetBurnProof"

	// get burning address
	getBurningAddress = "getburningaddress"

	// portal
	createAndSendTxWithCustodianDeposit           = "createandsendtxwithcustodiandeposit"
	createAndSendTxWithReqPToken                  = "createandsendtxwithreqptoken"
	getPortalState                                = "getportalstate"
	getPortalCustodianDepositStatus               = "getportalcustodiandepositstatus"
	createAndSendRegisterPortingPublicTokens      = "createandsendregisterportingpublictokens"
	createAndSendPortalExchangeRates              = "createandsendportalexchangerates"
	getPortalFinalExchangeRates                   = "getportalfinalexchangerates"
	getPortalPortingRequestByKey                  = "getportalportingrequestbykey"
	getPortalPortingRequestByPortingId            = "getportalportingrequestbyportingid"
	convertExchangeRates                          = "convertexchangerates"
	getPortalReqPTokenStatus                      = "getportalreqptokenstatus"
	getPortingRequestFees                         = "getportingrequestfees"
	createAndSendTxWithRedeemReq                  = "createandsendtxwithredeemreq"
	createAndSendTxWithReqUnlockCollateral        = "createandsendtxwithrequnlockcollateral"
	getPortalReqUnlockCollateralStatus            = "getportalrequnlockcollateralstatus"
	getPortalReqRedeemStatus                      = "getportalreqredeemstatus"
	createAndSendCustodianWithdrawRequest         = "createandsendcustodianwithdrawrequest"
	getCustodianWithdrawByTxId                    = "getcustodianwithdrawbytxid"
	getCustodianLiquidationStatus                 = "getcustodianliquidationstatus"
	createAndSendTxWithReqWithdrawRewardPortal    = "createandsendtxwithreqwithdrawrewardportal"
	createAndSendTxRedeemFromLiquidationPoolV3    = "createandsendtxredeemfromliquidationpoolv3"
	createAndSendCustodianTopup                   = "createandsendcustodiantopup"
	createAndSendTopUpWaitingPorting              = "createandsendtopupwaitingporting"
	createAndSendCustodianTopupV3                 = "createandsendcustodiantopupv3"
	createAndSendTopUpWaitingPortingV3            = "createandsendtopupwaitingportingv3"
	getTopupAmountForCustodian                    = "gettopupamountforcustodian"
	getLiquidationExchangeRatesPool               = "getliquidationtpexchangeratespool"
	getPortalReward                               = "getportalreward"
	getRequestWithdrawPortalRewardStatus          = "getrequestwithdrawportalrewardstatus"
	createAndSendTxWithReqMatchingRedeem          = "createandsendtxwithreqmatchingredeem"
	getReqMatchingRedeemStatus                    = "getreqmatchingredeemstatus"
	getPortalCustodianTopupStatus                 = "getcustodiantopupstatus"
	getPortalCustodianTopupStatusV3               = "getcustodiantopupstatusv3"
	getPortalCustodianTopupWaitingPortingStatus   = "getcustodiantopupwaitingportingstatus"
	getPortalCustodianTopupWaitingPortingStatusV3 = "getcustodiantopupwaitingportingstatusv3"
	getAmountTopUpWaitingPorting                  = "getamounttopupwaitingporting"
	getPortalReqRedeemByTxIDStatus                = "getreqredeemstatusbytxid"
	getReqRedeemFromLiquidationPoolByTxIDStatus   = "getreqredeemfromliquidationpoolbytxidstatus"
	getReqRedeemFromLiquidationPoolByTxIDStatusV3 = "getreqredeemfromliquidationpoolbytxidstatusv3"
	createAndSendTxWithCustodianDepositV3         = "createandsendtxwithcustodiandepositv3"
	getPortalCustodianDepositStatusV3             = "getportalcustodiandepositstatusv3"
	checkPortalExternalHashSubmitted              = "checkportalexternalhashsubmitted"
	createAndSendTxWithCustodianWithdrawRequestV3 = "createandsendtxwithcustodianwithdrawrequestv3"
	getCustodianWithdrawRequestStatusV3ByTxId     = "getcustodianwithdrawrequeststatusv3"
	getPortalWithdrawCollateralProof              = "getportalwithdrawcollateralproof"
	createAndSendUnlockOverRateCollaterals        = "createandsendtxwithunlockoverratecollaterals"
	getPortalUnlockOverRateCollateralsStatus      = "getportalunlockoverratecollateralsbytxidstatus"

	// relaying
	createAndSendTxWithRelayingBNBHeader = "createandsendtxwithrelayingbnbheader"
	createAndSendTxWithRelayingBTCHeader = "createandsendtxwithrelayingbtcheader"
	getRelayingBNBHeaderState            = "getrelayingbnbheaderstate"
	getRelayingBNBHeaderByBlockHeight    = "getrelayingbnbheaderbyblockheight"
	getBTCRelayingBestState              = "getbtcrelayingbeststate"
	getBTCBlockByHash                    = "getbtcblockbyhash"
	getLatestBNBHeaderBlockHeight        = "getlatestbnbheaderblockheight"

	// incognito mode for sc
	getBurnProofForDepositToSC                      = "getburnprooffordeposittosc"
	getBurnPBSCProofForDepositToSC                  = "getburnpbscprooffordeposittosc"
	createAndSendBurningForDepositToSCRequest       = "createandsendburningfordeposittoscrequest"
	createAndSendBurningForDepositToSCRequestV2     = "createandsendburningfordeposittoscrequestv2"
	createAndSendBurningPBSCForDepositToSCRequest   = "createandsendburningpbscfordeposittoscrequest"
	getBurnPLGProofForDepositToSC                   = "getburnplgprooffordeposittosc"
	createAndSendBurningPLGForDepositToSCRequest    = "createandsendburningplgfordeposittoscrequest"
	createAndSendBurningFTMForDepositToSCRequest    = "createandsendburningftmfordeposittoscrequest"
	getBurnFTMProofForDepositToSC                   = "getburnftmprooffordeposittosc"
	createAndSendBurningAURORAForDepositToSCRequest = "createandsendburningaurorafordeposittoscrequest"
	getBurnAURORAProofForDepositToSC                = "getburnauroraprooffordeposittosc"
	createAndSendBurningAVAXForDepositToSCRequest   = "createandsendburningavaxfordeposittoscrequest"
	getBurnAVAXProofForDepositToSC                  = "getburnavaxprooffordeposittosc"

	getFeatureStats       = "getfeaturestats"
	getSyncStats          = "getsyncstats"
	getBeaconPoolInfo     = "getbeaconpoolinfo"
	getShardPoolInfo      = "getshardpoolinfo"
	getCrossShardPoolInfo = "getcrossshardpoolinfo"
	getAllView            = "getallview"
	getAllViewDetail      = "getallviewdetail"
	isInstantFinality     = "isinstantfinality"

	// feature rewards
	getRewardFeature = "getrewardfeature"

	getTotalStaker = "gettotalstaker"

	// validator state
	getValKeyState = "getvalkeystate"

	// portal v4
	getPortalV4State                           = "getportalv4state"
	getPortalV4Params                          = "getportalv4params"
	createAndSendTxWithShieldingRequest        = "createandsendtxshieldingrequest"
	getPortalShieldingRequestStatus            = "getportalshieldingrequeststatus"
	createAndSendTxWithPortalV4UnshieldRequest = "createandsendtxwithportalv4unshieldrequest"
	getPortalUnshieldingRequestStatus          = "getportalunshieldrequeststatus"
	getPortalBatchUnshieldingRequestStatus     = "getportalbatchunshieldrequeststatus"
	getSignedRawTransactionByBatchID           = "getportalsignedrawtransaction"
	createAndSendTxWithPortalReplacementFee    = "createandsendtxwithportalreplacebyfee"
	getPortalReplacementFeeStatus              = "getportalreplacebyfeestatus"
	createAndSendTxWithPortalSubmitConfirmedTx = "createandsendtxwithportalsubmitconfirmedtx"
	getPortalSubmitConfirmedTx                 = "getportalsubmitconfirmedtxstatus"
	getSignedRawReplaceFeeTransaction          = "getportalsignedrawreplacebyfeetransaction"
	createAndSendTxPortalConvertVaultRequest   = "createandsendtxportalconvertvault"
	getPortalConvertVaultTxStatus              = "getportalconvertvaultstatus"
	generatePortalShieldMultisigAddress        = "generateportalshieldmultisigaddress"

	// stake
	unstake = "createunstaketransaction"

	connectionStatus        = "getconnectionstatus"
	getBeaconStaker         = "getbeaconstaker"
	getShardStaker          = "getshardstaker"
	getBeaconCommitteeState = "getbeaconcommitteestate"
	// prune
	prune          = "pruneState"
	getPruneState  = "getPruneState"
	checkPruneData = "checkprunedata"
)

const (
	testSubcrice                                = "testsubcribe"
	subcribeNewShardBlock                       = "subcribenewshardblock"
	subcribeNewBeaconBlock                      = "subcribenewbeaconblock"
	subcribePendingTransaction                  = "subcribependingtransaction"
	subcribeShardCandidateByPublickey           = "subcribeshardcandidatebypublickey"
	subcribeShardPendingValidatorByPublickey    = "subcribeshardpendingvalidatorbypublickey"
	subcribeShardCommitteeByPublickey           = "subcribeshardcommitteebypublickey"
	subcribeBeaconCandidateByPublickey          = "subcribebeaconcandidatebypublickey"
	subcribeBeaconPendingValidatorByPublickey   = "subcribebeaconpendingvalidatorbypublickey"
	subcribeBeaconCommitteeByPublickey          = "subcribebeaconcommitteebypublickey"
	subcribeCrossOutputCoinByPrivateKey         = "subcribecrossoutputcoinbyprivatekey"
	subcribeCrossCustomTokenByPrivateKey        = "subcribecrosscustomtokenbyprivatekey"
	subcribeCrossCustomTokenPrivacyByPrivateKey = "subcribecrosscustomtokenprivacybyprivatekey"
	subcribeMempoolInfo                         = "subcribemempoolinfo"
	subcribeShardBestState                      = "subcribeshardbeststate"
	subcribeBeaconBestState                     = "subcribebeaconbeststate"
	subcribeBeaconBestStateFromMem              = "subcribebeaconbeststatefrommem"
	subcribeBeaconPoolBeststate                 = "subcribebeaconpoolbeststate"
	subcribeShardPoolBeststate                  = "subcribeshardpoolbeststate"
)

// add method names when add new feature flags
var FeatureFlagWithMethodNames = map[string][]string{
	common.PortalRelayingFlag: {
		createAndSendTxWithRelayingBNBHeader,
		createAndSendTxWithRelayingBTCHeader,
		getRelayingBNBHeaderState,
		getRelayingBNBHeaderByBlockHeight,
		getBTCRelayingBestState,
		getBTCBlockByHash,
		getLatestBNBHeaderBlockHeight,
	},
	common.PortalV3Flag: {
		createAndSendTxWithCustodianDeposit,
		createAndSendTxWithReqPToken,
		getPortalState,
		getPortalCustodianDepositStatus,
		createAndSendRegisterPortingPublicTokens,
		createAndSendPortalExchangeRates,
		getPortalFinalExchangeRates,
		getPortalPortingRequestByKey,
		getPortalPortingRequestByPortingId,
		convertExchangeRates,
		getPortalReqPTokenStatus,
		getPortingRequestFees,
		createAndSendTxWithRedeemReq,
		createAndSendTxWithReqUnlockCollateral,
		getPortalReqUnlockCollateralStatus,
		getPortalReqRedeemStatus,
		createAndSendCustodianWithdrawRequest,
		getCustodianWithdrawByTxId,
		getCustodianLiquidationStatus,
		createAndSendTxWithReqWithdrawRewardPortal,
		createAndSendTxRedeemFromLiquidationPoolV3,
		createAndSendCustodianTopup,
		createAndSendTopUpWaitingPorting,
		createAndSendCustodianTopupV3,
		createAndSendTopUpWaitingPortingV3,
		getTopupAmountForCustodian,
		getLiquidationExchangeRatesPool,
		getPortalReward,
		getRequestWithdrawPortalRewardStatus,
		createAndSendTxWithReqMatchingRedeem,
		getReqMatchingRedeemStatus,
		getPortalCustodianTopupStatus,
		getPortalCustodianTopupStatusV3,
		getPortalCustodianTopupWaitingPortingStatus,
		getPortalCustodianTopupWaitingPortingStatusV3,
		getAmountTopUpWaitingPorting,
		getPortalReqRedeemByTxIDStatus,
		getReqRedeemFromLiquidationPoolByTxIDStatus,
		getReqRedeemFromLiquidationPoolByTxIDStatusV3,
		getPortalCustodianDepositStatusV3,
		checkPortalExternalHashSubmitted,
		createAndSendTxWithCustodianWithdrawRequestV3,
		getCustodianWithdrawRequestStatusV3ByTxId,
		getPortalWithdrawCollateralProof,
		createAndSendUnlockOverRateCollaterals,
		getPortalUnlockOverRateCollateralsStatus,
		getRewardFeature,
	},
	common.PortalV4Flag: {
		getPortalV4State,
		createAndSendTxWithShieldingRequest,
		getPortalShieldingRequestStatus,
		createAndSendTxWithPortalV4UnshieldRequest,
		getPortalUnshieldingRequestStatus,
		getPortalBatchUnshieldingRequestStatus,
		getSignedRawTransactionByBatchID,
		createAndSendTxWithPortalReplacementFee,
		getPortalReplacementFeeStatus,
		createAndSendTxWithPortalSubmitConfirmedTx,
		getPortalSubmitConfirmedTx,
		getSignedRawReplaceFeeTransaction,
		createAndSendTxPortalConvertVaultRequest,
		getPortalConvertVaultTxStatus,
		getPortalV4Params,
		generatePortalShieldMultisigAddress,
	},
}
