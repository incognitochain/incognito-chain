package common

const (
	EmptyString         = ""
	MiliConstant        = 3 // 1 constant = 10^3 mili constant, we will use 1 miliconstant as minimum unit constant in tx
	IncMerkleTreeHeight = 29
	RefundPeriod        = 1000 // after 1000 blocks since a tx (small & no-privacy) happens, the network will refund an amount of constants to tx initiator automatically
)

const (
	TxSubmitDCBProposal = "pd"  // submit DCB proposal tx
	TxSubmitGOVProposal = "pg"  // submit GOV proposal tx
	TxVoteDCBProposal   = "vd"  // submit DCB proposal voted tx
	TxVoteGOVProposal   = "vg"  // submit GOV proposal voted tx
	TxVoteDCBBoard      = "vbd" // vote DCB board tx
	TxVoteGOVBoard      = "vbg" // vote DCB board tx

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

// ico amounts
const (
	InitialDCBAmt = 10000
	InitialGOVAmt = 10000
)

// board addresses
var (
	DCBAddress = []byte{}
	GOVAddress = []byte{}
	ICOAddress = []byte{}
)

// special token ids (aka. PropertyID in custom token)
var (
	GOVTokenID  = [HashSize]byte{82, 253, 252, 7, 33, 130, 101, 79, 22, 63, 95, 15, 154, 98, 29, 114, 149, 102, 199, 77, 16, 3, 124, 77, 123, 187, 4, 7, 209, 226, 198, 73}
	DCBTokenID  = [HashSize]byte{83, 140, 127, 150, 177, 100, 191, 27, 151, 187, 159, 75, 180, 114, 232, 159, 91, 20, 132, 242, 82, 9, 201, 217, 52, 62, 146, 186, 9, 221, 157, 82}
	BondTokenID = [HashSize]byte{210, 186, 142, 112, 7, 41, 131, 56, 66, 3, 196, 56, 212, 233, 75, 243, 153, 203, 216, 139, 188, 175, 184, 43, 97, 204, 150, 237, 18, 84, 23, 7}
)
