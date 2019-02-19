package rpcserver

// rpc cmd method
const (
	GetNetworkInfo     = "getnetworkinfo"
	GetConnectionCount = "getconnectioncount"
	GetAllPeers        = "getallpeers"
	GetRawMempool      = "getrawmempool"
	GetMempoolEntry    = "getmempoolentry"
	EstimateFee        = "estimatefee"
	GetGenerate        = "getgenerate"
	GetMiningInfo      = "getmininginfo"

	GetBestBlock      = "getbestblock"
	GetBestBlockHash  = "getbestblockhash"
	GetBlocks         = "getblocks"
	RetrieveBlock     = "retrieveblock"
	GetBlockChainInfo = "getblockchaininfo"
	GetBlockCount     = "getblockcount"
	GetBlockHash      = "getblockhash"

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
	GetCommitteeCandidateList                  = "getcommitteecandidate"
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

	CreateAndSendStakingTransaction = "createandsendstakingtransaction"

	GetShardBestState  = "getshardbeststate"
	GetBeaconBestState = "getbeaconbeststate"

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
	GetListCBBoard                       = "getlistcbboard"
	AppendListDCBBoard                   = "testappendlistdcbboard"
	AppendListGOVBoard                   = "testappendlistgovboard"
	GetGOVParams                         = "getgovparams"
	GetDCBParams                         = "getdcbparams"
	GetGOVConstitution                   = "getgovconstitution"
	GetDCBConstitution                   = "getdcbconstitution"
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

	// crowdsale
	GetListOngoingCrowdsale               = "getlistongoingcrowdsale"
	CreateCrowdsaleRequestToken           = "createcrowdsalerequesttoken"
	SendCrowdsaleRequestToken             = "sendcrowdsalerequesttoken"
	CreateAndSendCrowdsaleRequestToken    = "createandsendcrowdsalerequesttoken"
	CreateCrowdsaleRequestConstant        = "createcrowdsalerequestconstant"
	SendCrowdsaleRequestConstant          = "sendcrowdsalerequestconstant"
	CreateAndSendCrowdsaleRequestConstant = "createandsendcrowdsalerequestconstant"
	TestStoreCrowdsale                    = "teststorecrowdsale"

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
	GetAmountVoteToken                   = "getamountvotetoken"
	SetAmountVoteToken                   = "testsetamountvotetoken"

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
