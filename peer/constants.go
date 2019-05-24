package peer

import "time"

const (
	// listen all interface
	LocalHost         = "0.0.0.0"
	MaxRetryConn      = 15
	RetryConnDuration = 10 * time.Second
	ProtocolId        = "/blockchain/1.0.0"
	DelimMessageByte  = '\n'
	DelimMessageStr   = "\n"

	MsgLiveTime         = 3 * time.Second  // in second
	MsgsCleanupInterval = MsgLiveTime * 10 //in second
)

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.
const (
	ConnPending ConnState = iota
	ConnFailing
	ConnCanceled
	ConnEstablished
	ConnDisconnected
)

const (
	MAX_RETRIES_CHECK_HASH_MESSAGE = 5
	MAX_TIMEOUT_CHECK_HASH_MESSAGE = time.Duration(10)
	HEAVY_MESSAGE_SIZE             = 5 * 1024 * 1024
	SPAM_MESSAGE_SIZE              = 50 * 1024 * 1024
)

const (
	MESSAGE_TO_ALL    = byte('a')
	MESSAGE_TO_SHARD  = byte('s')
	MESSAGE_TO_PEER   = byte('p')
	MESSAGE_TO_BEACON = byte('b')
)
