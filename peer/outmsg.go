package peer

import (
	"github.com/incognitochain/incognito-chain/wire"
	net "github.com/libp2p/go-libp2p-net"
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

type NewPeerMsg struct {
	Peer  *Peer
	CConn chan *PeerConn
}

type NewStreamMsg struct {
	Stream net.Stream
	CConn  chan *PeerConn
}
