package common

const (
	EmptyString          = ""
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
	DurationOfDCBBoard = 6       //number of block one DCB board in charge
	DurationOfGOVBoard = 1000    //number of block one GOV board in charge
	MaxBlockSize       = 5000000 //byte 5MB
	MaxTxsInBlock      = 1000
	MinTxsInBlock      = 10                    // minium txs for block to get immediate process (meaning no wait time)
	MinBlockWaitTime   = 3                     // second
	MaxBlockWaitTime   = 10 - MinBlockWaitTime // second
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

//voting flag
const (
	Lv3EncryptionFlag = byte(iota)
	Lv2EncryptionFlag
	Lv1EncryptionFlag
	NormalEncryptionFlag
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
	USDAssetID = Hash{99, 99, 99, 99, 99, 99, 99, 99, 3}
)

// centralized website's pubkey
var (
	// PrivateKey: 112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV
	// PaymentAddress: 1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba
	CentralizedWebsitePubKey = []byte{3, 36, 133, 3, 185, 44, 62, 112, 196, 239, 49, 190, 100, 172, 50, 147, 196, 154, 105, 211, 203, 57, 242, 110, 34, 126, 100, 226, 74, 148, 128, 167, 0}
)

// board addresses
const (
	DCBAddress     = "1Uv4CtusMLW1GZpMS2HwqJ5fp654J6VUxUPEm8CpgkqDmKRbKuSH2dvBrrq4P75GNZdbCW3QGXmMQDRqpLnpyHrESqKvawLX7Ju29HjaD"
	GOVAddress     = "1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba"
	BurningAddress = "1NHp16Y29xjc1PoXb1qwr65BfVVoHZuCbtTkVyucRzbeydgQHs2wPu5PC1hD"
)

// CONSENSUS
const (
	EPOCH       = 10
	RANDOM_TIME = 5
	OFFSET      = 1

	NODEMODE_RELAY  = "relay"
	NODEMODE_SHARD  = "shard"
	NODEMODE_AUTO   = "auto"
	NODEMODE_BEACON = "beacon"

	BEACON_ROLE    = "beacon"
	SHARD_ROLE     = "shard"
	PROPOSER_ROLE  = "proposer"
	VALIDATOR_ROLE = "validator"
	PENDING_ROLE   = "pending"

	MAX_SHARD_NUMBER = 4
)

// Units converter
const (
	WeiToMilliEtherRatio = int64(1000000000000000)
)
