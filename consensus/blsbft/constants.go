package blsbft

import "time"

const (
	PROPOSE  = "PROPOSE"
	LISTEN   = "LISTEN"
	AGREE    = "AGREE"
	NEWROUND = "NEWROUND"
)

//
const (
	TIMEOUT             = 5 * time.Second
	MaxNetworkDelayTime = 150 * time.Millisecond // in ms
)
