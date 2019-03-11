package peer

import "github.com/big0t/constant-chain/common"

type PeerLoger struct {
	log common.Logger
}

func (peerLogger *PeerLoger) Init(inst common.Logger) {
	peerLogger.log = inst
}

// Global instant to use
var Logger = PeerLoger{}
