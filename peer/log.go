package peer

import "github.com/ninjadotorg/cash-prototype/common"

type PeerLoger struct {
	log common.Logger
}

func (self *PeerLoger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = PeerLoger{}
