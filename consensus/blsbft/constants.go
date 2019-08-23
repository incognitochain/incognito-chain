package blsbft

import "time"

const (
	PROPOSE  = "PROPOSE"
	LISTEN   = "LISTEN"
	VOTE     = "VOTE"
	NEWROUND = "NEWROUND"
	BLS      = "bls"
	BRI      = "dsa"
)

//
const (
	TIMEOUT             = 5 * time.Second
	MaxNetworkDelayTime = 150 * time.Millisecond // in ms
)
