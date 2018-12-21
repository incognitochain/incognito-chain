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
	CreateAndSendLoanRequest                   = "createandsendloanrequest"
	CreateAndSendLoanResponse                  = "createandsendloanresponse"
	CreateAndSendLoanPayment                   = "createandsendloanpayment"
	CreateAndSendLoanWithdraw                  = "createandsendloanwithdraw"
	GetMempoolInfo                             = "getmempoolinfo"
	GetCommitteeCandidateList                  = "getcommitteecandidate"
	RetrieveCommitteeCandidate                 = "retrievecommitteecandidate"
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
	GetLoanParams                              = "loanparams"
	GetBlockHeader                             = "getheader"
	RandomCommitments                          = "randomcommitments"
	HasSerialNumbers                           = "hasserialnumbers"

	// Wallet rpc cmd
	ListAccounts               = "listaccounts"
	GetAccount                 = "getaccount"
	GetAddressesByAccount      = "getaddressesbyaccount"
	GetAccountAddress          = "getaccountaddress"
	DumpPrivkey                = "dumpprivkey"
	ImportAccount              = "importaccount"
	RemoveAccount              = "removeaccount"
	ListUnspentOutputCoins     = "listunspentoutputcoins"
	GetBalance                 = "getbalance"
	GetBalanceByPrivatekey     = "getbalancebyprivatekey"
	GetBalanceByPaymentAddress = "getbalancebypaymentaddress"
	GetReceivedByAccount       = "getreceivedbyaccount"
	SetTxFee                   = "settxfee"
	EncryptData                = "encryptdata"

	// multisig for board spending
	CreateSignatureOnCustomTokenTx = "createsignatureoncustomtokentx"
	GetListDCBBoard                = "getlistdcbboard"
	GetListCBBoard                 = "getlistcbboard"
	GetListGOVBoard                = "getlistgovboard"
	GetGOVParams                   = "getgovparams"
	GetDCBParams                   = "getdcbparams"
	GetGOVConstitution             = "getgovconstitution"
	GetDCBConstitution             = "getdcbconstitution"

	// vote
	SendRawVoteBoardDCBTx                = "sendrawvoteboarddcbtx"
	CreateRawVoteDCBBoardTx              = "createrawvotedcbboardtx"
	CreateAndSendVoteDCBBoardTransaction = "createandsendvotedcbboardtransaction"
	SendRawVoteBoardGOVTx                = "sendrawvoteboardgovtx"
	CreateRawVoteGOVBoardTx              = "createrawvotegovboardtx"
	CreateAndSendVoteGOVBoardTransaction = "createandsendvotegovboardtransaction"

	// Submit Proposal
	CreateAndSendSubmitDCBProposalTx = "createandsendsubmitdcbproposaltx"
	CreateRawSubmitDCBProposalTx     = "createrawsubmitdcbproposaltx"
	SendRawSubmitDCBProposalTx       = "sendrawsubmitdcbproposaltx"
	CreateAndSendSubmitGOVProposalTx = "createandsendsubmitgovproposaltx"
	CreateRawSubmitGOVProposalTx     = "createrawsubmitgovproposaltx"
	SendRawSubmitGOVProposalTx       = "sendrawsubmitgovproposaltx"

	// dcb
	CreateAndSendTxWithIssuingRequest     = "createandsendtxwithissuingrequest"
	CreateAndSendTxWithContractingRequest = "createandsendtxwithcontractingrequest"

	// gov
	GetBondTypes                      = "getbondtypes"
	CreateAndSendTxWithBuyBackRequest = "createandsendtxwithbuybackrequest"
	CreateAndSendTxWithBuySellRequest = "createandsendtxwithbuysellrequest"

	// cmb
	CreateAndSendTxWithCMBInitRequest = "createandsendtxwithcmbinitrequest"
)
