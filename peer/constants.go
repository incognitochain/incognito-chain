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
)

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.
const (
	ConnPending      ConnState = iota
	ConnFailing
	ConnCanceled
	ConnEstablished
	ConnDisconnected
)
