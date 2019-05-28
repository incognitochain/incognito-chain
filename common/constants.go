package common

import "time"

// for common
const (
	EmptyString          = ""
	PaymentAddressLength = 66
	ZeroByte             = byte(0x00)
	DateOutputFormat     = "2006-01-02T15:04:05.999999"
	DateInputFormat      = "2006-01-02T15:04:05.999999"
	NextForceUpdate      = "2019-06-15T23:59:00.000000"
)

// for exit code
const (
	ExitCodeUnknow = iota
	ExitByOs
	ExitByLogging
	ExitCodeForceUpdate
)

// For all Transaction information
const (
	TxNormalType             = "n"  // normal tx(send and receive coin)
	TxRewardType             = "s"  // reward tx
	TxReturnStakingType      = "rs" //
	TxCustomTokenType        = "t"  // token  tx with no supporting privacy
	TxCustomTokenPrivacyType = "tp" // token  tx with supporting privacy
	MaxTxSize                = 100  // unit KB = 100KB
)

// for mining consensus
const (
	MaxBlockSize         = 2000 //unit kilobytes = 2 Megabyte
	MaxTxsInBlock        = 1000
	MinTxsInBlock        = 10                   // minium txs for block to get immediate process (meaning no wait time)
	MinBlockWaitTime     = 2                    // second
	MaxBlockWaitTime     = 4 - MinBlockWaitTime // second
	MinBeaconBlkInterval = 3 * time.Second      //second
	MinShardBlkInterval  = 5 * time.Second      //second
)

// special token ids (aka. PropertyID in custom token)
var (
	ConstantID = Hash{4} // To send Constant in custom token
)

// centralized website's pubkey
var (
	CentralizedWebsitePubKey = []byte{2, 194, 130, 176, 102, 36, 183, 114, 109, 135, 49, 114, 177, 92, 214, 31, 25, 4, 72, 103, 196, 161, 36, 69, 121, 102, 159, 24, 31, 131, 101, 20, 0}
	// CentralizedWebsitePubKey = []byte{3, 159, 2, 42, 22, 163, 195, 221, 129, 31, 217, 133, 149, 16, 68, 108, 42, 192, 58, 95, 39, 204, 63, 68, 203, 132, 221, 48, 181, 131, 40, 189, 0}
)

// board addresses
const (
	// DCBAddress     = "1NHpWKZYCLQeGKSSsJewsA8p3nsPoAZbmEmtsuBqd6yU7KJnzJZVt39b7AgP"
	// GOVAddress     = "1NHoFQ3Nr8fQm3ZLk2ACSgZXjVH6JobpuV65RD3QAEEGe76KknMQhGbc4g8P"
	BurningAddress = "1NHp2EKw7ALdXUzBfoRJvKrBBM9nkejyDcHVPvUjDcWRyG22dHHyiBKQGL1c"
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

	MAX_SHARD_NUMBER = 2
)

// Units converter
const (
	WeiToMilliEtherRatio = int64(1000000000000000)
	WeiToEtherRatio      = int64(1000000000000000000)
)
