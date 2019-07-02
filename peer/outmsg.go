package peer

import (
	"github.com/incognitochain/incognito-chain/wire"
)

// outMsg is used to house a message to be sent along with a channel to signal
// when the message has been sent (or won't be sent due to things such as
// shutdown)
type outMsg struct {
	forwardType  byte // a all, s shard, p  peer, b beacon
	forwardValue *byte
	rawBytes     *[]byte
	message      wire.Message
	doneChan     chan<- struct{}
}
