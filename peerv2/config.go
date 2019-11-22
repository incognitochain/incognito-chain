package peerv2

import "github.com/libp2p/go-libp2p-core/protocol"

func (s Host) GetProxyStreamProtocolID() protocol.ID {
	return protocol.ID("proxy/" + s.Version)
}

func (s Host) GetDirectProtocolID() protocol.ID {
	return protocol.ID("direct/" + s.Version)
}
