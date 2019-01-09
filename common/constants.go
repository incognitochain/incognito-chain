package common

const (
	EmptyString  = ""
	TrueValue    = true
	FalseValue   = false
	NanoConstant = 2    // 1 constant = 10^2 nano constant, we will use 1 miliconstant as minimum unit constant in tx
	RefundPeriod = 1000 // after 1000 blocks since a tx (small & no-privacy) happens, the network will refund an amount of constants to tx initiator automatically
	PubKeyLength = 33
	ZeroByte     = byte(0x00)
)

const (
	TxNormalType             = "n"  // normal tx(send and receive coin)
	TxSalaryType             = "s"  // salary tx(gov pay salary for block producer)
	TxCustomTokenType        = "t"  // token  tx with no supporting privacy
	TxCustomTokenPrivacyType = "tp" // token  tx with supporting privacy
)

// for mining consensus
const (
	DurationOfTermDCB     = 1000    //number of block one DCB board in charge
	DurationOfTermGOV     = 1000    //number of block one GOV board in charge
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

// for voting parameter
const (
	SumOfVoteDCBToken                 = 100000000
	SumOfVoteGOVToken                 = 100000000
	MinimumBlockOfProposalDuration    = 50
	MaximumBlockOfProposalDuration    = 200
	MaximumProposalExplainationLength = 1000
	NumberOfDCBGovernors              = 50
	NumberOfGOVGovernors              = 50
	EncryptionOnePhraseDuration       = uint32(5)
	RewardProposalSubmitter           = 500
	BasePercentage                    = 10000
	PercentageBoardSalary             = 5
)

//voting flag
const (
	Lv3EncryptionFlag = iota
	Lv2EncryptionFlag
	Lv1EncryptionFlag
	NormalEncryptionFlag
)

// board types
const (
	DCB = 1
	GOV = 2
)

// special token ids (aka. PropertyID in custom token)
var (
	BondTokenID      = [HashSize]byte{0, 0, 0, 0, 0, 0, 0, 0}
	DCBTokenID       = [HashSize]byte{1}
	GOVTokenID       = [HashSize]byte{2}
	CMBTokenID       = [HashSize]byte{3}
	ConstantID       = [HashSize]byte{4} // To send Constant in custom token
	DCBVotingTokenID = [HashSize]byte{5}
	GOVVotingTokenID = [HashSize]byte{6}
	OffchainAssetID  = [HashSize]byte{7, 7, 7, 7, 7, 7, 7, 7} // First 8 bytes of offchain asset
)

// asset IDs for oracle feed
var (
	ETHAssetID = [HashSize]byte{99}
	BTCAssetID = [HashSize]byte{88}
)

// board addresses
const (
	DCBAddress     = "1Uv3jP4ixNx3BkEtmUUxKXA1TXUduix3KMCWXHvLqVyA9CFfoLRZ949zTBNqDUPSzaPCZPrQKSfiEHguFazK6VeDmEk1RMLfX1kQiSqJ6"
	GOVAddress     = "1Uv3jP4ixNx3BkEtmUUxKXA1TXUduix3KMCWXHvLqVyA9CFfoLRZ949zTBNqDUPSzaPCZPrQKSfiEHguFazK6VeDmEk1RMLfX1kQiSqJ6"
	BurningAddress = "1NHp16Y29xjc1PoXb1qwr65BfVVoHZuCbtTkVyucRzbeydgQHs2wPu5PC1hD"
)
