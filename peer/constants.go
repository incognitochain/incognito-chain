package peer

import "time"

const (
	// listen all interface
	localHost         = "0.0.0.0"
	maxRetryConn      = 15
	retryConnDuration = 10 * time.Second
	protocolID        = "/incognito/0.6.1-beta"
	delimMessageByte  = '\n'
	delimMessageStr   = "\n"

	messageLiveTime        = 3 * time.Second      // in second
	messageCleanupInterval = messageLiveTime * 10 //in second
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
	heavyMessageSize           = 5 * 1024 * 1024
	spamMessageSize            = 50 * 1024 * 1024
)

const (
	MessageToAll    = byte('a')
	MessageToShard  = byte('s')
	MessageToPeer   = byte('p')
	MessageToBeacon = byte('b')
)

var (
	RelayNode = []string{
		"16Hn1SNtGTsYS7zYBcct4b5Jn5xCzzC8S846Er1kFVUfGRxs1Ht",
	}
)
