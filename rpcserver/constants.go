package rpcserver

// rpc cmd method
const (
	GetNetworkInfo           = "getnetworkinfo"
	GetConnectionCount       = "getconnectioncount"
	GetAllPeers              = "getallpeers"
	GetRawMempool            = "getrawmempool"
	GetMempoolEntry          = "getmempoolentry"
	EstimateFee              = "estimatefee"
	EstimateFeeWithEstimator = "estimatefeewithestimator"
	GetGenerate              = "getgenerate"
	GetMiningInfo            = "getmininginfo"

	GetBestBlock        = "getbestblock"
	GetBestBlockHash    = "getbestblockhash"
	GetBlocks           = "getblocks"
	RetrieveBlock       = "retrieveblock"
	RetrieveBeaconBlock = "retrievebeaconblock"
	GetBlockChainInfo   = "getblockchaininfo"
	GetBlockCount       = "getblockcount"
	GetBlockHash        = "getblockhash"

	ListOutputCoins                            = "listoutputcoins"
	CreateRawTransaction                       = "createtransaction"
	SendRawTransaction                         = "sendtransaction"
	CreateAndSendTransaction                   = "createandsendtransaction"
	CreateAndSendCustomTokenTransaction        = "createandsendcustomtokentransaction"
	SendRawCustomTokenTransaction              = "sendrawcustomtokentransaction"
	CreateRawCustomTokenTransaction            = "createrawcustomtokentransaction"
	CreateRawPrivacyCustomTokenTransaction     = "createrawprivacycustomtokentransaction"
	SendRawPrivacyCustomTokenTransaction       = "sendrawprivacycustomtokentransaction"
	CreateAndSendPrivacyCustomTokenTransaction = "createandsendprivacycustomtokentransaction"
	GetMempoolInfo                             = "getmempoolinfo"
	GetCandidateList                           = "getcandidatelist"
	GetCommitteeList                           = "getcommitteelist"
	CanPubkeyStake                             = "canpubkeystake"
	GetBlockProducerList                       = "getblockproducer"
	ListUnspentCustomToken                     = "listunspentcustomtoken"
	GetTransactionByHash                       = "gettransactionbyhash"
	ListCustomToken                            = "listcustomtoken"
	ListPrivacyCustomToken                     = "listprivacycustomtoken"
	CustomToken                                = "customtoken"
	PrivacyCustomToken                         = "privacycustomtoken"
	CheckHashValue                             = "checkhashvalue"
	GetListCustomTokenBalance                  = "getlistcustomtokenbalance"
	GetListPrivacyCustomTokenBalance           = "getlistprivacycustomtokenbalance"
	GetBlockHeader                             = "getheader"
	RandomCommitments                          = "randomcommitments"
	HasSerialNumbers                           = "hasserialnumbers"
	HasSnDerivators                            = "hassnderivators"

	CreateAndSendStakingTransaction = "createandsendstakingtransaction"

	GetShardBestState  = "getshardbeststate"
	GetBeaconBestState = "getbeaconbeststate"

	GetBeaconPoolState            = "getbeaconpoolstate"
	GetShardPoolState             = "getshardpoolstate"
	GetShardPoolLatestValidHeight = "getshardpoollatestvalidheight"

	GetShardToBeaconPoolState = "getshardtobeaconpoolstate"
	GetCrossShardPoolState    = "getcrossshardpoolstate"

	// Wallet rpc cmd
	ListAccounts                       = "listaccounts"
	GetAccount                         = "getaccount"
	GetAddressesByAccount              = "getaddressesbyaccount"
	GetAccountAddress                  = "getaccountaddress"
	DumpPrivkey                        = "dumpprivkey"
	ImportAccount                      = "importaccount"
	RemoveAccount                      = "removeaccount"
	ListUnspentOutputCoins             = "listunspentoutputcoins"
	GetBalance                         = "getbalance"
	GetBalanceByPrivatekey             = "getbalancebyprivatekey"
	GetBalanceByPaymentAddress         = "getbalancebypaymentaddress"
	GetReceivedByAccount               = "getreceivedbyaccount"
	SetTxFee                           = "settxfee"
	GetRecentTransactionsByBlockNumber = "getrecenttransactionsbyblocknumber"

	// multisig for board spending
	CreateSignatureOnCustomTokenTx       = "createsignatureoncustomtokentx"
	GetListDCBBoard                      = "getlistdcbboard"
	GetListGOVBoard                      = "getlistgovboard"
	GetListCMBBoard                      = "getlistcmbboard"
	AppendListDCBBoard                   = "testappendlistdcbboard"
	AppendListGOVBoard                   = "testappendlistgovboard"
	GetGOVParams                         = "getgovparams"
	GetDCBParams                         = "getdcbparams"
	GetGOVConstitution                   = "getgovconstitution"
	GetDCBConstitution                   = "getdcbconstitution"
	GetDCBBoardIndex                     = "getdcbboardindex"
	GetGOVBoardIndex                     = "getgovboardindex"
	CreateAndSendTxWithMultiSigsReg      = "createandsendtxwithmultisigsreg"
	CreateAndSendTxWithMultiSigsSpending = "createandsendtxwithmultisigsspending"

	// dcb loan
	CreateAndSendLoanRequest  = "createandsendloanrequest"
	CreateAndSendLoanResponse = "createandsendloanresponse"
	CreateAndSendLoanPayment  = "createandsendloanpayment"
	CreateAndSendLoanWithdraw = "createandsendloanwithdraw"
	GetLoanResponseApproved   = "getloanresponseapproved"
	GetLoanResponseRejected   = "getloanresponserejected"
	GetLoanParams             = "loanparams"
	GetLoanPaymentInfo        = "getloanpaymentinfo"
	GetBankFund               = "getbankfund"
	GetLoanRequestTxStatus    = "getloanrequesttxstatus"

	// crowdsale
	GetListOngoingCrowdsale               = "getlistongoingcrowdsale"
	CreateCrowdsaleRequestToken           = "createcrowdsalerequesttoken"
	SendCrowdsaleRequestToken             = "sendcrowdsalerequesttoken"
	CreateAndSendCrowdsaleRequestToken    = "createandsendcrowdsalerequesttoken"
	CreateCrowdsaleRequestConstant        = "createcrowdsalerequestconstant"
	SendCrowdsaleRequestConstant          = "sendcrowdsalerequestconstant"
	CreateAndSendCrowdsaleRequestConstant = "createandsendcrowdsalerequestconstant"
	GetListDCBProposalBuyingAssets        = "getlistdcbproposalbuyingassets"
	GetListDCBProposalSellingAssets       = "getlistdcbproposalsellingassets"

	// reserve
	CreateIssuingRequest            = "createissuingrequest"
	SendIssuingRequest              = "sendissuingrequest"
	CreateAndSendIssuingRequest     = "createandsendissuingrequest"
	CreateAndSendContractingRequest = "createandsendcontractingrequest"
	GetIssuingStatus                = "getissuingstatus"
	GetContractingStatus            = "getcontractingstatus"
	ConvertETHToDCBTokenAmount      = "convertethtodcbtokenamount"
	ConvertCSTToETHAmount           = "convertcsttoethamount"

	// vote
	SendRawVoteBoardDCBTx                = "sendrawvoteboarddcbtx"
	CreateRawVoteDCBBoardTx              = "createrawvotedcbboardtx"
	CreateAndSendVoteDCBBoardTransaction = "createandsendvotedcbboardtransaction"
	SendRawVoteBoardGOVTx                = "sendrawvoteboardgovtx"
	CreateRawVoteGOVBoardTx              = "createrawvotegovboardtx"
	CreateAndSendVoteGOVBoardTransaction = "createandsendvotegovboardtransaction"

	//vote propopsal
	GetEncryptionFlag                         = "getencryptionflag"
	SetEncryptionFlag                         = "testsetencryptionflag"
	GetEncryptionLastBlockHeightFlag          = "getencryptionlastblockheightflag"
	CreateAndSendSealLv3VoteProposal          = "createandsendseallv3voteproposal"
	CreateAndSendSealLv2VoteProposal          = "createandsendseallv2voteproposal"
	CreateAndSendSealLv1VoteProposal          = "createandsendseallv1voteproposal"
	CreateAndSendNormalVoteProposalFromOwner  = "createandsendnormalvoteproposalfromowner"
	CreateAndSendNormalVoteProposalFromSealer = "createandsendnormalvoteproposalfromsealer"

	// Submit Proposal
	CreateAndSendSubmitDCBProposalTx = "createandsendsubmitdcbproposaltx"
	CreateRawSubmitDCBProposalTx     = "createrawsubmitdcbproposaltx"
	SendRawSubmitDCBProposalTx       = "sendrawsubmitdcbproposaltx"
	CreateAndSendSubmitGOVProposalTx = "createandsendsubmitgovproposaltx"
	CreateRawSubmitGOVProposalTx     = "createrawsubmitgovproposaltx"
	SendRawSubmitGOVProposalTx       = "sendrawsubmitgovproposaltx"

	// dcb
	// CreateAndSendTxWithIssuingRequest     = "createandsendtxwithissuingrequest"
	// CreateAndSendTxWithContractingRequest = "createandsendtxwithcontractingrequest"

	// gov
	GetBondTypes                           = "getbondtypes"
	GetCurrentSellingBondTypes             = "getcurrentsellingbondtypes"
	GetCurrentStabilityInfo                = "getcurrentstabilityinfo"
	CreateAndSendTxWithBuyBackRequest      = "createandsendtxwithbuybackrequest"
	CreateAndSendTxWithBuySellRequest      = "createandsendtxwithbuysellrequest"
	CreateAndSendTxWithOracleFeed          = "createandsendtxwithoraclefeed"
	CreateAndSendTxWithUpdatingOracleBoard = "createandsendtxwithupdatingoracleboard"
	CreateAndSendTxWithSenderAddress       = "createandsendtxwithsenderaddress"
	CreateAndSendTxWithBuyGOVTokensRequest = "createandsendtxwithbuygovtokensrequest"
	GetCurrentSellingGOVTokens             = "getcurrentsellinggovtokens"

	// cmb
	CreateAndSendTxWithCMBInitRequest     = "createandsendtxwithcmbinitrequest"
	CreateAndSendTxWithCMBInitResponse    = "createandsendtxwithcmbinitresponse"
	CreateAndSendTxWithCMBDepositContract = "createandsendtxwithcmbdepositcontract"
	CreateAndSendTxWithCMBDepositSend     = "createandsendtxwithcmbdepositsend"
	CreateAndSendTxWithCMBWithdrawRequest = "createandsendtxwithcmbwithdrawrequest"

	// wallet
	GetPublicKeyFromPaymentAddress = "getpublickeyfrompaymentaddress"
	DefragmentAccount              = "defragmentaccount"
)

//Fee of specific transaction
const (
	FeeSubmitProposal = 100
	FeeVote           = 100
)
