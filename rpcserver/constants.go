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

	ListTransactions                   = "listtransactions"
	CreateTransaction                  = "createtransaction"
	SendTransaction                    = "sendtransaction"
	SendMany                           = "sendmany"
	CreateActionParamsTransaction      = "createactionparamstransaction"
	SendRegistrationCandidateCommittee = "sendregistration"
	SendCustomTokenTransaction         = "sendcustomtokentransaction"
	GetMempoolInfo                     = "getmempoolinfo"
	GetCommitteeCandidateList          = "getcommitteecandidate"
	RetrieveCommitteeCandidate         = "retrievecommitteecandidate"
	GetBlockProducerList               = "getblockproducer"
	ListUnspentCustomToken             = "listunspentcustomtoken"
	GetTransactionByHash               = "gettransactionbyhash"
	ListCustomToken                    = "listcustomtoken"
	CustomToken                        = "customtoken"
	CheckHashValue                     = "checkhashvalue"

	GetHeader = "getheader"

	// Wallet rpc cmd
	ListAccounts          = "listaccounts"
	GetAccount            = "getaccount"
	GetAddressesByAccount = "getaddressesbyaccount"
	GetAccountAddress     = "getaccountaddress"
	DumpPrivkey           = "dumpprivkey"
	ImportAccount         = "importaccount"
	RemoveAccount         = "removeaccount"
	ListUnspent           = "listunspent"
	GetBalance            = "getbalance"
	GetReceivedByAccount  = "getreceivedbyaccount"
	SetTxFee              = "settxfee"
	CreateProducerKeyset  = "createproducerkeyset"

	// multisig for board spending
	BuildCustomTokenTransaction = "buildcustomtokentransaction"
	GetCustomTokenSignature     = "getcustomtokensignature"
	GetListDCBBoard             = "getlistdcbboard"
	GetListCBBoard              = "getlistcbboard"
	GetListGOVBoard             = "getlistgovboard"
)
