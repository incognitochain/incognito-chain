package common

const (
	EmptyString          = ""
	RefundPeriod         = 1000 // after 1000 blocks since a tx (small & no-privacy) happens, the network will refund an amount of constants to tx initiator automatically
	PubKeyLength         = 33
	PaymentAddressLength = 66
	ZeroByte             = byte(0x00)
	DateOutputFormat     = "2006-01-02T15:04:05.999999"
)

const (
	TxNormalType             = "n"  // normal tx(send and receive coin)
	TxSalaryType             = "s"  // salary tx(gov pay salary for block producer)
	TxCustomTokenType        = "t"  // token  tx with no supporting privacy
	TxCustomTokenPrivacyType = "tp" // token  tx with supporting privacy
)

// for mining consensus
const (
	DurationOfDCBBoard    = 6       //number of block one DCB board in charge
	DurationOfGOVBoard    = 1000    //number of block one GOV board in charge
	MaxBlockSize          = 5000000 //byte 5MB
	MaxTxsInBlock         = 1000
	MinTxsInBlock         = 10                    // minium txs for block to get immediate process (meaning no wait time)
	MinBlockWaitTime      = 3                     // second
	MaxBlockWaitTime      = 10 - MinBlockWaitTime // second
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
	NumberOfDCBGovernors              = 3
	NumberOfGOVGovernors              = 3
	EncryptionOnePhraseDuration       = 5
	RewardProposalSubmitter           = 500
	BasePercentage                    = 10000
	PercentageBoardSalary             = 5
)

//Fee of specific transaction
const (
	FeeSubmitProposal = 100
	FeeVoteProposal   = 100
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
	DCBBoard = byte(1)
	GOVBoard = byte(2)
)

// special token ids (aka. PropertyID in custom token)
var (
	BondTokenID      = Hash{0, 0, 0, 0, 0, 0, 0, 0} // first 8 bytes must be 0
	DCBTokenID       = Hash{1}
	GOVTokenID       = Hash{2}
	CMBTokenID       = Hash{3}
	ConstantID       = Hash{4} // To send Constant in custom token
	DCBVotingTokenID = Hash{5}
	GOVVotingTokenID = Hash{6}
)

// asset IDs for oracle feed (must prefix with 99)
var (
	BTCAssetID = Hash{99, 99, 99, 99, 99, 99, 99, 99, 1}
	ETHAssetID = Hash{99, 99, 99, 99, 99, 99, 99, 99, 2}
)

// board addresses
const (
	DCBAddress     = "1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba"
	GOVAddress     = "1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba"
	BurningAddress = "1NHp16Y29xjc1PoXb1qwr65BfVVoHZuCbtTkVyucRzbeydgQHs2wPu5PC1hD"
)

// CONSENSUS
const (
	// SHARD_NUMBER = 4
	SHARD_NUMBER = 1 //single-node mode
	EPOCH        = 10
	RANDOM_TIME  = 5
	COMMITEES    = 1
	OFFSET       = 1

	SHARD_ROLE            = "shard"
	SHARD_PROPOSER_ROLE   = "shard-proposer"
	SHARD_VALIDATOR_ROLE  = "shard-validator"
	SHARD_PENDING_ROLE    = "shard-pending"
	BEACON_ROLE           = "beacon"
	BEACON_PROPOSER_ROLE  = "beacon-proposer"
	BEACON_VALIDATOR_ROLE = "beacon-validator"
	BEACON_PENDING_ROLE   = "beacon-pending"
	AUTO_ROLE             = "auto"
)

//PBFT
const (
	PBFT_LISTEN  = "listen"
	PBFT_PROPOSE = "propose"
	PBFT_PREPARE = "prepare"
	PBFT_COMMIT  = "commit"
)
