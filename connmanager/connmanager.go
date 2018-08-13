package connmanager

import (
	"github.com/internet-cash/prototype/peer"
)

type ConnManager struct {
	Config Config
}

type Config struct {
	// OnAccept is a callback that is fired when an inbound connection is accepted
	OnAccept func(*peer.Peer)

	//OnConnectiontion is a callback that is fired when an outbound connection is established
	OnConnection func(*peer.Peer)

	//OnDisconnectiontion is a callback that is fired when an outbound connection is disconnected
	OnDisconnection func(peer *peer.Peer)
}
