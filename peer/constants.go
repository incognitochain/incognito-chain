package peer

import "time"

const (
	// listen all interface
	localHost         = "0.0.0.0"
	maxRetryConn      = 15
	retryConnDuration = 10 * time.Second // in 10 second
	PrefixProtocolID  = "/incognito/"
	delimMessageByte  = '\n'
	delimMessageStr   = "\n"

	messageLiveTime        = 3 * time.Second      // in 3 second
	messageCleanupInterval = messageLiveTime * 10 //in second: messageLiveTime * 10
)

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.
const (
	connPending ConnState = iota
	connCanceled
	connEstablished
)

const (
	maxRetriesCheckHashMessage = 5
	maxTimeoutCheckHashMessage = time.Duration(10)
	heavyMessageSize           = 5 * 1024 * 1024  // 5 Mb
	spamMessageSize            = 50 * 1024 * 1024 // 50 Mb
)

const (
	MessageToAll    = byte('a')
	MessageToShard  = byte('s')
	MessageToPeer   = byte('p')
	MessageToBeacon = byte('b')
)
