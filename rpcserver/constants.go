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

	ListTransactions                = "listtransactions"
	CreateTransaction               = "createtransaction"
	SendTransaction                 = "sendtransaction"
	CreateAndSendTransaction        = "sendmany" // TODO rename
	CreateActionParamsTransaction   = "createactionparamstransaction"
	SendCustomTokenTransaction      = "sendcustomtokentransaction"
	SendRawCustomTokenTransaction   = "sendrawcustomtokentransaction"
	CreateRawCustomTokenTransaction = "createrawcustomtokentransaction"
	GetMempoolInfo                  = "getmempoolinfo"
	GetCommitteeCandidateList       = "getcommitteecandidate"
	RetrieveCommitteeCandidate      = "retrievecommitteecandidate"
	GetBlockProducerList            = "getblockproducer"
	ListUnspentCustomToken          = "listunspentcustomtoken"
	GetTransactionByHash            = "gettransactionbyhash"
	ListCustomToken                 = "listcustomtoken"
	CustomToken                     = "customtoken"
	CheckHashValue                  = "checkhashvalue"
	GetListCustomTokenBalance       = "getlistcustomtokenbalance"

	GetHeader = "getheader"

	// Wallet rpc cmd
	ListAccounts           = "listaccounts"
	GetAccount             = "getaccount"
	GetAddressesByAccount  = "getaddressesbyaccount"
	GetAccountAddress      = "getaccountaddress"
	DumpPrivkey            = "dumpprivkey"
	ImportAccount          = "importaccount"
	RemoveAccount          = "removeaccount"
	ListUnspent            = "listunspent"
	GetBalance             = "getbalance"
	GetBalanceByPrivatekey = "getbalancebyprivatekey"
	GetReceivedByAccount   = "getreceivedbyaccount"
	SetTxFee               = "settxfee"
	EncryptData            = "encryptdata"

	// multisig for board spending
	CreateSignatureOnCustomTokenTx = "createsignatureoncustomtokentx"
	GetListDCBBoard                = "getlistdcbboard"
	GetListCBBoard                 = "getlistcbboard"
	GetListGOVBoard                = "getlistgovboard"
)
