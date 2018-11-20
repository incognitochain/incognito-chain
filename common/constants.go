package common

const (
	EmptyString         = ""
	MiliConstant        = 3 // 1 constant = 10^3 mili constant, we will use 1 miliconstant as minimum unit constant in tx
	IncMerkleTreeHeight = 29
)

const (
	TxSubmitDCBProposal  = "pd" // submit DCB proposal tx
	TxSubmitGOVProposal  = "pg" // submit GOV proposal tx
	TxVoteDCBProposal    = "vd" // submit DCB proposal voted tx
	TxVoteGOVProposal    = "vg" // submit GOV proposal voted tx
	TxVoteDCBBoard = "vbd" // vote DCB board tx
	TxVoteGOVBoard = "vbg" // vote DCB board tx

	TxAcceptDCBProposal  = "ad" // accept DCB proposal
	TxAcceptGOVProposal  = "ag" // accept GOV proposal
	TxNormalType         = "n"  // normal tx(send and receive coin)
	TxSalaryType         = "s"  // salary tx(gov pay salary for block producer)
	TxCustomTokenType    = "t"  // token  tx
	TxLoanRequest        = "lr"
	TxLoanResponse       = "ls"
	TxLoanPayment        = "lp"
	TxLoanWithdraw       = "lw"
	TxDividendPayout     = "td"
	TxCrowdsale          = "cs"
	TxBuyFromGOVRequest  = "bgr"
	TxBuySellDCBRequest  = "bsdr"
	TxBuySellDCBResponse = "bsdrs"
	TxBuyFromGOVResponse = "bgrs"
	TxBuyBackRequest     = "bbr"
	TxBuyBackResponse    = "bbrs"
)

// unit type use in tx
// coin or token or bond
const (
	AssetTypeCoin      = "c" // 'constant' coin
	AssetTypeBond      = "b" // bond
	AssetTypeGovToken  = "g" // government token
	AssetTypeBankToken = "d" // decentralized central bank token
)

var ListAsset = []string{AssetTypeCoin, AssetTypeBond, AssetTypeGovToken, AssetTypeBankToken}

// for mining consensus
const (
	MaxBlockSize          = 5000000 //byte 5MB
	MaxTxsInBlock         = 1000
	MinTxsInBlock         = 10                    // minium txs for block to get immediate process (meaning no wait time)
	MinBlockWaitTime      = 3                     // second
	MaxBlockWaitTime      = 20 - MinBlockWaitTime // second
	MaxSyncChainTime      = 5                     // second
	MaxBlockSigWaitTime   = 5                     // second
	MaxBlockPerTurn       = 100                   // maximum blocks that a validator can create per turn
	TotalValidators       = 20                    // = TOTAL CHAINS
	MinBlockSigs          = (TotalValidators / 2) + 1
	GetChainStateInterval = 10 //second
	MaxBlockTime          = 10 //second Maximum for a chain to grow
)

// board types
const (
	DCB = 1
	GOV = 2
)

// board addresses
var (
	DCBAddress = []byte{}
	GOVAddress = []byte{}
)
