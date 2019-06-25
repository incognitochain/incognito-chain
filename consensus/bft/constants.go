package bft

import "time"

const (
	PROPOSE  = "PROPOSE"
	LISTEN   = "LISTEN"
	PREPARE  = "PREPARE"
	NEWROUND = "NEWROUND"
)

//
const (
	TIMEOUT = 60 * time.Second
)
