package common

const (
	EmptyString         = ""
	NanoConstant        = 2 // 1 constant = 10^2 nano constant, we will use 1 miliconstant as minimum unit constant in tx
	IncMerkleTreeHeight = 29
	RefundPeriod        = 1000 // after 1000 blocks since a tx (small & no-privacy) happens, the network will refund an amount of constants to tx initiator automatically
)

const (
	TxNormalType             = "n" // normal tx(send and receive coin)
	TxSalaryType             = "s" // salary tx(gov pay salary for block producer)
	TxCustomTokenType        = "t" // token  tx with no supporting privacy
	TxCustomTokenPrivacyType = "t" // token  tx with supporting privacy

	TxSubmitDCBProposal = "pd"  // submit DCB proposal tx
	TxSubmitGOVProposal = "pg"  // submit GOV proposal tx
	TxVoteDCBProposal   = "vd"  // submit DCB proposal voted tx
	TxVoteGOVProposal   = "vg"  // submit GOV proposal voted tx
	TxVoteDCBBoard      = "vbd" // vote DCB board tx
	TxVoteGOVBoard      = "vbg" // vote DCB board tx

	TxAcceptDCBProposal = "ad" // accept DCB proposal
	TxAcceptGOVProposal = "ag" // accept GOV proposal

	TxLoanRequest        = "lr"
	TxLoanResponse       = "ls"
	TxLoanPayment        = "lp"
	TxLoanWithdraw       = "lw"
	TxDividendPayout     = "td"
	TxBuyFromGOVRequest  = "bgr"
	TxBuySellDCBRequest  = "bsdr"
	TxBuySellDCBResponse = "bsds"
	TxBuyFromGOVResponse = "bgrs"
	TxBuyBackRequest     = "bbr"
	TxBuyBackResponse    = "bbrs"
)

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
