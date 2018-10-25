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

	ListTransactions                  = "listtransactions"
	CreateTransaction                 = "createtransaction"
	SendTransaction                   = "sendtransaction"
	SendMany                          = "sendmany"
	GetNumberOfCoinsAndBonds          = "getnumberofcoinsandbonds"
	CreateActionParamsTransaction     = "createactionparamstransaction"
	SendRegistrationCandidateCommitee = "sendregistration"
	GetMempoolInfo                    = "getmempoolinfo"
	GetCndList                        = "getcndlist"

	GetHeader = "getheader"

	// Wallet rpc cmd
	ListAccounts          = "listaccounts"
	GetAccount            = "getaccount"
	GetAddressesByAccount = "getaddressesbyaccount"
	GetAccountAddress     = "getaccountaddress"
	DumpPrivkey           = "dumpprivkey"
	ImportAccount         = "importaccount"
	ListUnspent           = "listunspent"
	GetBalance            = "getbalance"
	GetReceivedByAccount  = "getreceivedbyaccount"
	SetTxFee              = "settxfee"
	CreateSealerKeyset    = "createsealerkeyset"
)
